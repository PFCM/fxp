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
}

// InputMappingMIDI yields oscillators with two inputs:
//   - 0: a standard MIDI note from 0 to 127
//   - 1: fix.S17 of pitch bend. The range of the pitch bend
//     is also set in MIDI notes but must be set at construction.
type InputMappingMIDI struct {
	pbRange byte
}

func (imm InputMappingMIDI) step(fs ...fix.S17) float32 {
	panic("unimplemented")
}

// InputMappingLFO is specialised for more precision in lower frequencies
// TODO what does it look like? maybe it's a float type
type InputMappingLFO struct{}

func (iml InputMappingLFO) step(...fix.S17) float32 { panic("no") }

// InputMappingConst yields oscillators with zero inputs that always play at a
// constant frequency.
type InputMappingConst struct {
	s float32
}

func (imc InputMappingConst) step(...fix.S17) float32 { return imc.s }

// WaveTable is a wavetable oscillator. It receives a single input, which is the
// note to play, and has one output, an appropriate block of samples. The note
// wouldn't make sense in a fix.S17; it is reintepreted as a fix.U62 encoding a
// (fractional) midi note, offset by the Lowest field (which may be negative).
// TODO: we may want more fractional components.
type Table struct {
	// TODO: should just be a buffer.Ring?
	tab        []fix.S17
	phase      float32
	samplerate float32
	Lowest     int
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
			out[0][i] = interp.L(t.tab[j], t.tab[k], fix.FromFloat(c))
		}
		t.phase += t.step(fix.U62(step))
		for t.phase >= float32(len(t.tab)) {
			t.phase -= float32(len(t.tab))
		}
	}
}

// Sine returns a Table initialised with a sensible sine wave.
func Sine(samplerate float32, lowest int) *Table {
	const n = 128
	table := make([]fix.S17, n)
	for i := range table {
		f := math.Sin(math.Pi / float64(n/2) * float64(i))
		table[i] = fix.FromFloat(f)
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
func RatSine(samplerate float32, lowest int, exp float32) *Table {
	const n = 128
	table := make([]fix.S17, n)
	for i := range table {
		f := math.Sin(math.Pi / float64(n/2) * float64(i))
		f = math.Pow(float64(exp), f)
		table[i] = fix.S17(fix.FloatToRat44(f))
	}
	return &Table{
		tab:        table,
		samplerate: samplerate,
		Lowest:     lowest,
		nn:         true,
	}
}

func Square(samplerate float32, high, low fix.S17, lowestNote int) *Table {
	// lol table
	table := []fix.S17{high, low}
	return &Table{
		tab:        table,
		samplerate: samplerate,
		Lowest:     lowestNote,
		nn:         true,
	}
}

// step calculates a step value to achieve the provided midi note value as closely as
// possible.
func (t Table) step(note fix.U62) float32 {
	// first turn the note into a frequency.
	// TODO: lookup table?
	n := float64(t.Lowest) + fix.U62ToFloat[float64](note)
	freq := math.Pow(2.0, (n-69)/12) * 440
	// freq is essentially in tables per second, calculate how many samples
	// from the table we need per second.
	tableSamplesPerSecond := float64(len(t.tab)) * freq
	// figure out the seconds per output sample
	secondsPerOutputSample := 1.0 / t.samplerate
	// table samples per output sample is therefore the rate we need to advance
	// through the table.
	return float32(tableSamplesPerSecond) * secondsPerOutputSample
}
