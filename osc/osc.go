// package osc provides oscillators.
package osc

import (
	"math"

	"github.com/pfcm/fxp"
	"github.com/pfcm/fxp/fix"
	"github.com/pfcm/fxp/interp"
)

// InputMapping determines how an oscillator maps its input into frequency. This
// controls both the interpretation of the inputs as well as the number of
// channels expected.
type InputMapping interface {
	step(...fix.S17) float32
	inputs() int
}

// InputMappingMIDI yields oscillators with two inputs:
//   - 0: a standard MIDI note from 0 to 127, as a fix.U71
//   - 1: fix.S17 of pitch bend. The range of the pitch bend
//     is also set in MIDI notes but must be set at construction.
type InputMappingMIDI struct {
	samplerate float32
	bend       byte
}

// NewInputMappingMidi creates a new MIDI input mapping. The only argument apart
// from the global sample rate is half the range of the pitch bend, in MIDI
// notes (semitones). The actual pitch bend will be (almost) +/- the provided
// range.
func NewInputMappingMIDI(samplerate float32, bend byte) InputMappingMIDI {
	return InputMappingMIDI{samplerate: samplerate, bend: bend}
}

func (imm InputMappingMIDI) step(fs ...fix.S17) float32 {
	if len(fs) != 2 {
		panic("wrong number of inputs for InputMappingMIDI")
	}
	note := fix.U71(fs[0])
	_ = note
	panic("unimplemented")
}

func (imm InputMappingMIDI) inputs() int { return 2 }

// InputMappingLFO is specialised for more precision in lower frequencies
// TODO what does it look like? maybe it's a float type
type InputMappingLFO struct{}

func (iml InputMappingLFO) step(...fix.S17) float32 { panic("no") }
func (iml InputMappingLFO) inputs() int             { panic("idk") }

// InputMappingConst yields oscillators with zero inputs that always play at a
// constant frequency.
type InputMappingConst struct {
	s float32
}

func (imc InputMappingConst) step(...fix.S17) float32 { return imc.s }
func (InputMappingConst) inputs() int                 { return 0 }

// WaveTable is a wavetable oscillator. It receives a single input, which is the
// note to play, and has one output, an appropriate block of samples. The note
// wouldn't make sense in a fix.S17; it is reintepreted as a fix.U71 encoding a
// midi note, offset by the Lowest field (which may be negative).
// TODO: we may want more fractional components.
type Table struct {
	// TODO: should just be a buffer.Ring?
	tab        []fix.S17
	phase      float32
	samplerate float32
	Lowest     float32
	nn         bool
}

var _ fxp.Ticker = &Table{}

func (t *Table) Inputs() int    { return 1 }
func (t *Table) Outputs() int   { return 1 }
func (t *Table) String() string { return "osc.Table" }

func (t *Table) Tick(in, out [][]fix.S17) {
	for i, step := range in[0] {
		j := int(t.phase)
		if t.nn {
			out[0][i] = t.tab[j]
		} else {
			k := int(t.phase+1) % len(t.tab)
			c := t.phase - float32(j)
			out[0][i] = interp.L(t.tab[j], t.tab[k], fix.S17FromFloat(c))
		}
		t.phase += t.step(fix.U71(step))
		for t.phase >= float32(len(t.tab)) {
			t.phase -= float32(len(t.tab))
		}
	}
}

// Sine returns a Table initialised with a sensible sine wave.
func Sine(samplerate, lowest float32) *Table {
	const n = 128
	table := make([]fix.S17, n)
	for i := range table {
		f := math.Sin(math.Pi / float64(n/2) * float64(i))
		table[i] = fix.S17FromFloat(f)
	}
	return &Table{
		tab:        table,
		samplerate: samplerate,
		Lowest:     lowest,
	}
}

// RatSine returns a table initialised with an exponentiated sine wave intended
// to be interpreted as Rat44s. The result ranges between exp and 1/exp, give or
// take precision.
func RatSine(samplerate, lowest, exp float32) *Table {
	const n = 128
	table := make([]fix.S17, n)
	for i := range table {
		f := math.Sin(math.Pi / float64(n/2) * float64(i))
		f = math.Pow(float64(exp), f)
		table[i] = fix.S17(fix.Rat44FromFloat(f))
	}
	return &Table{
		tab:        table,
		samplerate: samplerate,
		Lowest:     lowest,
		nn:         true,
	}
}

func Square(samplerate float32, high, low fix.S17, lowestNote float32) *Table {
	// lol table
	table := []fix.S17{high, low}
	return &Table{
		tab:        table,
		samplerate: samplerate,
		Lowest:     lowestNote,
		nn:         true,
	}
}

func Saw(samplerate, lowestNote float32) *Table {
	table := make([]fix.S17, 0, 256)
	for i := -128; i < 128; i++ {
		table = append(table, fix.S17(i))
	}
	return &Table{
		tab:        table,
		samplerate: samplerate,
		Lowest:     lowestNote,
		nn:         true,
	}
}

// step calculates a step value to achieve the provided midi note value as closely as
// possible.
func (t Table) step(note fix.U71) float32 {
	// first turn the note into a frequency.
	// TODO: lookup table?
	n := float64(t.Lowest) + fix.U71ToFloat[float64](note)
	freq := math.Pow(2.0, (n-69)/12) * 440
	// freq is in tables per second, calculate how many samples from the
	// table we need per second.
	tableSamplesPerSecond := float64(len(t.tab)) * freq
	// figure out the seconds per output sample
	secondsPerOutputSample := 1.0 / t.samplerate
	// table samples per output sample is therefore the rate we need to advance
	// through the table.
	return float32(tableSamplesPerSecond) * secondsPerOutputSample
}
