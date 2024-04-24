// package hid handles human interface devices. Or IO that uses the same protocols,
// like MIDI.
package hid

import (
	"context"
	"fmt"
	"sync"

	"github.com/pfcm/fxp/fix"
	"github.com/pfcm/fxp/midi"
	"github.com/pfcm/fxp/midi/coremidi"
)

var dispatcher = midi.Listen(context.TODO(), coremidi.ReceiveAll)

// MidiNotes polyphonically tracks MIDI note on and off messages, providing two
// outputs per voice: one containing the MIDI note number and the other the
// velocity. Both are zero when the voice has no note, and multiple voices are
// interleave.
type MidiNotes struct {
	voices int

	mu     sync.Mutex
	notes  []byte
	velos  []byte
	when   []uint64
	events uint64
}

func NewMidiNotes(voices int) *MidiNotes {
	md := &MidiNotes{
		voices: voices,
		notes:  make([]byte, voices),
		velos:  make([]byte, voices),
		when:   make([]uint64, voices),
	}

	c := dispatcher.Subscribe(
		midi.WithoutCV1Type(midi.CV1PolyPressure),
		midi.WithoutCV1Type(midi.CV1ControlChange),
		midi.WithoutCV1Type(midi.CV1ProgramChange),
		midi.WithoutCV1Type(midi.CV1ChannelPressure),
		midi.WithoutCV1Type(midi.CV1PitchBend),
	)
	go func() {
		for msg := range c {
			// TODO: not this
			msg.Note <<= 2
			switch msg.CV1Type {
			case midi.CV1NoteOn:
				md.noteOn(msg.Note, msg.Velocity)
			case midi.CV1NoteOff:
				md.noteOff(msg.Note)
			default:
				panic(msg.CV1Type)
			}
		}
	}()

	return md
}

func (m *MidiNotes) noteOn(n, v byte) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.events++
	i, w := -1, m.events
	for j, x := range m.when {
		// always pick an idle voice
		if m.velos[j] == 0 {
			i = j
			break
		}
		// otherwise, the oldest
		if x < w {
			w = x
			i = j
		}
	}
	m.notes[i] = n
	m.velos[i] = v
	m.when[i] = m.events
}

func (m *MidiNotes) noteOff(n byte) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for i, o := range m.notes {
		if o == n {
			// Only set velocity to 0, keep outputting the same
			// note.
			m.velos[i] = 0
			break
		}
	}
}

func (m *MidiNotes) Inputs() int    { return 0 }
func (m *MidiNotes) Outputs() int   { return m.voices * 2 }
func (m *MidiNotes) String() string { return fmt.Sprintf("MidiNotes(%d)", m.voices) }

func (m *MidiNotes) Tick(_, out [][]fix.S17) {
	for i := 0; i < m.voices; i++ {
		m.mu.Lock()
		n, v := m.notes[i], m.velos[i]
		m.mu.Unlock()
		if n == 0 {
			continue
		}
		for j := range out[i*2] {
			out[i*2][j] = fix.S17(n)
			out[i*2+1][j] = fix.S17(v)
		}
	}
}
