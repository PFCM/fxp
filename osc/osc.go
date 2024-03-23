// package osc provides oscillators.
package osc

import (
	"math"

	"github.com/pfcm/fxp"
	"github.com/pfcm/fxp/fix"
	"github.com/pfcm/fxp/interp"
)

// Table is a wavetable oscillator. It receives a single input, which is the
// note to play, and has one output, an appropriate block of samples.
// The note wouldn't make sense in a fix.S17; it is reintepreted as a
// fix.U62 encoding a (fractional) midi note, offset by the Lowest field.
// TODO: we may want more fractional components.
type Table struct {
	tab        []fix.S17
	phase      float32
	samplerate float32
	Lowest     int
}

var _ fxp.Ticker = &Table{}

func (t *Table) Inputs() int    { return 1 }
func (t *Table) Outputs() int   { return 1 }
func (t *Table) String() string { return "osc.Table" }

func (t *Table) Tick(in, out [][]fix.S17) {
	for i, step := range in[0] {
		j, k := int(t.phase), int(t.phase+1)%len(t.tab)
		c := t.phase - float32(j)
		out[0][i] = interp.L(t.tab[j], t.tab[k], fix.FromFloat(c))
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
