// package fxp does low bit depth audio.
package fxp

import (
	"fmt"
	"math"
	"strings"

	"github.com/pfcm/fxp/fix"
)

// Ticker is something that processes audio.
// TODO: allow more sample types?
type Ticker interface {
	// Inputs returns the number of expected input channels.
	Inputs() int
	// Outputs returns the number of expected output channels.
	Outputs() int
	// Tick processes a chunk of audio. The first dimension of the input
	// slice is always InChannels, and the first dimension of the output
	// slice is always OutChannels. Each individual element of both slices
	// is always the same length. Tickers may overwrite the input buffer.
	Tick(input, output [][]fix.S17)

	fmt.Stringer
}

// Splitter is a Ticker that just copies its single input to all of its outputs.
type Splitter struct {
	outs int
}

var _ Ticker = Splitter{}

func (s Splitter) Inputs() int    { return 1 }
func (s Splitter) Outputs() int   { return s.outs }
func (s Splitter) String() string { return fmt.Sprintf("Splitter%d", s.outs) }

func (s Splitter) Tick(input, output [][]fix.S17) {
	for _, o := range output {
		copy(o, input[0])
	}
}

// Const is a Ticker that always fills its single output with a given value.
type Const struct {
	Val fix.S17
}

var _ Ticker = Const{}

func (c Const) Inputs() int    { return 0 }
func (c Const) Outputs() int   { return 1 }
func (c Const) String() string { return fmt.Sprintf("Const(%v)", c.Val) }

func (c Const) Tick(_, output [][]fix.S17) {
	for i := range output[0] {
		output[0][i] = c.Val
	}
}

// Scale is a Ticker that multiplies it input by a constant and shifts it by a
// constant.
type Scale struct {
	Mul   fix.S17
	Shift fix.S17
}

var _ Ticker = Scale{}

func (s Scale) Inputs() int    { return 1 }
func (s Scale) Outputs() int   { return 1 }
func (s Scale) String() string { return fmt.Sprintf("Scale(%v, %v)", s.Mul, s.Shift) }

func (s Scale) Tick(input, output [][]fix.S17) {
	for i, c := range input[0] {
		output[0][i] = c.SMul(s.Mul).SAdd(s.Shift)
	}
}

// Chain is a ticker that applies a sequence of Tickers. The inputs and outputs all
// need to line up.
type Chain struct {
	ts              []Ticker
	inputs, outputs int
	b1, b2          [][]fix.S17
}

var _ Ticker = Chain{}

func Serially(ts ...Ticker) Chain {
	if len(ts) == 0 {
		panic(fmt.Errorf("empty chain"))
	}
	maxChans := ts[0].Inputs()
	for i := 1; i < len(ts); i++ {
		if ts[i-1].Outputs() != ts[i].Inputs() {
			panic(fmt.Errorf(
				"outputs/inputs mismatch:\n%v (%d outputs)\n->\n%v (%d inputs)",
				ts[i-1], ts[i-1].Outputs(), ts[i], ts[i].Inputs()))
		}
		maxChans = max(ts[i-1].Outputs(), maxChans)
		maxChans = max(ts[i].Inputs(), maxChans)
	}
	maxChans = max(ts[len(ts)-1].Outputs(), maxChans)
	b1 := make([][]fix.S17, maxChans)
	for i := range b1 {
		b1[i] = make([]fix.S17, 4096)
	}
	b2 := make([][]fix.S17, maxChans)
	for i := range b2 {
		b2[i] = make([]fix.S17, 4096)
	}
	return Chain{
		ts:      ts,
		inputs:  ts[0].Inputs(),
		outputs: ts[len(ts)-1].Outputs(),
		b1:      b1,
		b2:      b2,
	}
}

func (c Chain) Inputs() int    { return c.inputs }
func (c Chain) Outputs() int   { return c.outputs }
func (c Chain) String() string { return fmt.Sprintf("Chain(%v)", c.ts) }

func (c Chain) Tick(input, output [][]fix.S17) {
	// TODO: we could certainly skip some copies, but also that gets messy.
	in, out := c.b1, c.b2
	for i := range input {
		copy(in[i], input[i])
		in[i] = in[i][:len(input[i])]
	}
	for i := range output {
		out[i] = out[i][:len(output[i])]
		for j := range out[i] {
			out[i][j] = 0
		}
	}
	in = in[:len(input)]
	for _, t := range c.ts {
		out = out[:t.Outputs()]
		t.Tick(in, out)
		in, out = out, in
	}
	for i := range out {
		copy(output[i], in[i])
	}
}

// Mixer mixes together a number of inputs, first applying the provided gains.
type Mixer struct {
	Gains []fix.S17
}

var _ Ticker = Mixer{}

// Sum returns a ticker that sums the given number of inputs down to one,
// reducing their gains to try and keep a roughly constant power.
func Sum(n int) Mixer {
	var (
		g  = fix.FromFloat(math.Sqrt(float64(n)))
		gs = make([]fix.S17, n)
	)
	for i := range gs {
		gs[i] = g
	}
	return Mixer{Gains: gs}
}

func (m Mixer) Inputs() int    { return len(m.Gains) }
func (m Mixer) Outputs() int   { return 1 }
func (m Mixer) String() string { return "Mixer" }

func (m Mixer) Tick(input, output [][]fix.S17) {
	for i := range input[0] {
		for j, g := range m.Gains {
			a := input[j][i].SMul(g)
			output[0][i] = output[0][i].SAdd(a)
		}
	}
}

// Concurrent is a Ticker that joins a group of tickers and runs them at the
// same time.
type Concurrent struct {
	ts              []Ticker
	inputs, outputs int
}

func Concurrently(ts ...Ticker) Concurrent {
	ins, outs := 0, 0
	for _, t := range ts {
		ins += t.Inputs()
		outs += t.Outputs()
	}
	return Concurrent{
		ts:      ts,
		inputs:  ins,
		outputs: outs,
	}
}

var _ Ticker = Concurrent{}

func (c Concurrent) Inputs() int  { return c.inputs }
func (c Concurrent) Outputs() int { return c.outputs }

func (c Concurrent) String() string {
	s := make([]string, len(c.ts))
	for i, t := range c.ts {
		s[i] = t.String()
	}
	return fmt.Sprintf("(%s)", strings.Join(s, ","))
}

func (c Concurrent) Tick(inputs, outputs [][]fix.S17) {
	in, out := 0, 0
	for _, t := range c.ts {
		ni, no := in+t.Inputs(), out+t.Outputs()
		t.Tick(inputs[in:ni], outputs[out:no])
		in, out = ni, no
	}
}

// Mult copies a single input to the provided number of outputs.
type Mult struct {
	N int
}

var _ Ticker = Mult{}

func (Mult) Inputs() int      { return 1 }
func (m Mult) Outputs() int   { return m.N }
func (m Mult) String() string { return fmt.Sprintf("Mult(%d)", m.N) }

func (m Mult) Tick(inputs, outputs [][]fix.S17) {
	for _, o := range outputs {
		copy(o, inputs[0])
	}
}

// Amp is a Ticker that just multiplies its two inputs.
type Amp struct{}

func (Amp) Inputs() int    { return 2 }
func (Amp) Outputs() int   { return 1 }
func (Amp) String() string { return "Amp" }

func (Amp) Tick(inputs, outputs [][]fix.S17) {
	for i := range outputs[0] {
		outputs[0][i] = inputs[0][i].SMul(inputs[1][i])
	}
}

// Noop is a Ticker that just copies its inputs to its outputs.
type Noop struct {
	N int
}

func (n Noop) Inputs() int    { return n.N }
func (n Noop) Outputs() int   { return n.N }
func (n Noop) String() string { return fmt.Sprintf("Noop(%d)", n) }

func (n Noop) Tick(inputs, outputs [][]fix.S17) {
	for i := range inputs {
		copy(outputs[i], inputs[i])
	}
}
