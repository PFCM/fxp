// package midi handles midi.
package midi

import (
	"context"
	"sync"
)

type ChannelMask byte

const AllChannels ChannelMask = 0xF

// Listener is function that blocks forever, calling a provided callback
// with UMP midi messages.
type Listener func(context.Context, func([]uint32)) error

type sub struct {
	f filter
	c chan Message
}

// Dispatcher routes MIDI messages to a set of channels.
type Dispatcher struct {
	mu   sync.Mutex
	subs []sub
}

// Listen starts listening for MIDI messages in the background with the provided
// Listener. It returns a Dispatcher whose Subscribe message can be used to get
// a channel on which to receive Messages.
func Listen(ctx context.Context, l Listener) *Dispatcher {
	var d Dispatcher

	go func() {
		if err := l(ctx, func(raw []uint32) {
			msgs, err := ParseMessages(raw)
			if err != nil {
				panic(err)
			}
			for _, m := range msgs {
				d.dispatch(m)
			}
		}); err != nil {
			panic(err)
		}
	}()

	return &d
}

func (d *Dispatcher) dispatch(msg Message) {
	d.mu.Lock()
	defer d.mu.Unlock()
	for _, s := range d.subs {
		if !s.f.match(msg) {
			continue
		}
		s.c <- msg
	}
}

func (d *Dispatcher) close() {
	d.mu.Lock()
	defer d.mu.Unlock()
	for _, s := range d.subs {
		close(s.c)
	}
	d.subs = d.subs[:0]
}

func (d *Dispatcher) Subscribe(opts ...SubscriptionFilter) <-chan Message {
	f := defaultFilter()
	for _, o := range opts {
		o(&f)
	}

	c := make(chan Message, 100)
	d.mu.Lock()
	d.subs = append(d.subs, sub{f: f, c: c})
	d.mu.Unlock()
	return c
}

type filter struct {
	channels ChannelMask
	cv1Types [7]bool
	// TODO: more cv1 things (cc index?)
	// TODO: more new UMP/MIDI 2.0 things?
}

func defaultFilter() filter {
	f := filter{
		channels: AllChannels,
	}
	for i := range f.cv1Types {
		f.cv1Types[i] = true
	}
	return f
}

func (f *filter) match(msg Message) bool {
	// TODO: this is wrong, what about channel 0?
	if msg.Channel&byte(f.channels) == 0 {
		return false
	}
	return f.cv1Types[int(msg.CV1Type&0x7)]
}

type SubscriptionFilter func(f *filter)

func WithChannelMask(cm ChannelMask) SubscriptionFilter {
	return func(f *filter) { f.channels = cm }
}

func WithoutCV1Type(t CV1MessageType) SubscriptionFilter {
	return func(f *filter) {
		f.cv1Types[int(t&0x7)] = false
	}
}
