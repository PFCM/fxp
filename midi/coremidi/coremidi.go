// package coremidi uses Core MIDI to provide midi IO on MacOS.
package coremidi

/*
#cgo LDFLAGS: -framework CoreMIDI
#cgo LDFLAGS: -framework CoreFoundation
#include <CoreMIDI/CoreMIDI.h>
#include <CoreFoundation/CoreFoundation.h>
#include <stdlib.h>

typedef const MIDIEventList elist;
extern void receive(elist *evtlist, void *srcConnRefCon);

static inline OSStatus input_port_create(MIDIClientRef client, CFStringRef name, MIDIPortRef *out) {
	return MIDIInputPortCreateWithProtocol(
        	client,
                name,
                kMIDIProtocol_1_0,
                out,
                ^(const MIDIEventList *evtlist, void *srcConnRefCon) {
                	receive(evtlist, srcConnRefCon);
                });
}
*/
import "C"
import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"unsafe"
)

// ReceiveAll registers to receive every MIDI message from every source, calling
// the provided callback with every raw message. It blocks until the the
// provided context is cancelled. The callback must be thread-safe, and will be
// passed a batch of simultaneous messages in MIDI1U format.
// For now, it is an error to call ReceiveAll more than once, even
// after a previous call has returned.
// TODO: fix this, it shouldn't be.
func ReceiveAll(ctx context.Context, f func([]uint32)) error {
	ip, err := newInputPort("receive_all")
	if err != nil {
		return err
	}
	srcs, err := listSources()
	if err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "Listening to:\n")
	for _, src := range srcs {
		if err := ip.Connect(src); err != nil {
			return fmt.Errorf("connecting to %q: %w", src.name, err)
		}
		fmt.Fprintf(os.Stderr, "\t%q\n", src.name)
	}
	for {
		select {
		case msgl := <-msgchan:
			go func() {
				f(msgl.words)
				msglistPool.Put(msgl)
			}()
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

type client struct {
	ref C.MIDIClientRef
}

var getClient = sync.OnceValues(func() (*client, error) {
	cfn, free := cfstr("coremidi-go")
	defer free()

	var ref C.MIDIClientRef
	// TODO: add a callback to find out about system changes?
	status := C.MIDIClientCreate(cfn, nil, nil, &ref)
	if status != C.noErr {
		return nil, fmt.Errorf("MIDIClientCreate: %v", int(status))
	}
	return &client{ref: ref}, nil
})

func (c *client) Close() error {
	status := C.MIDIClientDispose(c.ref)
	if status != C.noErr {
		return fmt.Errorf("MIDIClientDispose: %v", int(status))
	}
	return nil
}

type inputPort struct {
	ref C.MIDIPortRef
}

func newInputPort(name string) (*inputPort, error) {
	cl, err := getClient()
	if err != nil {
		return nil, err
	}
	cfname, free := cfstr(name)
	defer free()

	var ref C.MIDIPortRef
	status := C.input_port_create(
		cl.ref,
		cfname,
		&ref,
	)
	if status != C.noErr {
		return nil, fmt.Errorf("MIDIInputPortCreateWithProtocol: %v", status)
	}
	return &inputPort{ref: ref}, nil
}

func (i *inputPort) Connect(src *source) error {
	status := C.MIDIPortConnectSource(i.ref, src.ref, nil)
	if status != C.noErr {
		return fmt.Errorf("connect source %v to port %v: %v", src, i, status)
	}
	return nil
}

//export receive
func receive(events *C.elist, _ unsafe.Pointer) {
	for i := 0; i < int(events.numPackets); i++ {
		pkt := events.packet[i]
		msgs := msglistPool.Get().(*msglist)
		msgs.timestamp = uint64(pkt.timeStamp)
		msgs.words = msgs.words[:0]
		for j := 0; i < int(pkt.wordCount); i++ {
			msgs.words = append(msgs.words, uint32(pkt.words[j]))
		}
		select {
		case msgchan <- msgs:
		default:
			panic("can't keep up with midi receiver")
		}
	}
}

var msgchan = make(chan *msglist, 100)

type msglist struct {
	timestamp uint64
	words     []uint32
}

var msglistPool = sync.Pool{
	New: func() any {
		return &msglist{words: make([]uint32, 0, 64)}
	},
}

type source struct {
	ref  C.MIDIEndpointRef
	name string
}

func listSources() ([]*source, error) {
	num := C.MIDIGetNumberOfSources()
	if num == 0 {
		return nil, errors.New("possibly no sources?")
	}
	var sources []*source
	for i := 0; i < int(num); i++ {
		srcRef := C.MIDIGetSource(C.ItemCount(i))
		if srcRef == 0 {
			return nil, fmt.Errorf("error getting source %d", i)
		}
		var (
			cfs  C.CFStringRef
			name string
		)
		if C.MIDIObjectGetStringProperty(
			srcRef,
			C.kMIDIPropertyDisplayName,
			&cfs,
		) == C.noErr {
			// do something with it
			// buf := C.malloc(C.ulong(C.CFStringGetLength(cfs) + 1))
			strp := C.CFStringGetCStringPtr(cfs, C.kCFStringEncodingUTF8)
			name = C.GoString(strp)
			// C.free(buf)
		}

		sources = append(sources, &source{ref: srcRef, name: name})
	}
	return sources, nil
}

func cfstr(str string) (C.CFStringRef, func()) {
	c := C.CString(str)
	defer C.free(unsafe.Pointer(c))
	cf := C.CFStringCreateWithCString(
		C.kCFAllocatorDefault,
		c,
		C.kCFStringEncodingUTF8,
	)
	return cf, func() { C.CFRelease(C.CFTypeRef(cf)) }
}
