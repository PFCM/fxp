package fxp

import (
	"fmt"
	"time"

	"github.com/pfcm/fxp/fix"
)

// tickFn is a generic ticker that helps us avoid some boilerplate.
type tickFn struct {
	name            string
	inputs, outputs int
	tick            func([][]fix.S17, [][]fix.S17)
}

func (t tickFn) Inputs() int           { return t.inputs }
func (t tickFn) Outputs() int          { return t.outputs }
func (t tickFn) String() string        { return t.name }
func (t tickFn) Tick(i, o [][]fix.S17) { t.tick(i, o) }

// Tick creates a Ticker that just outputs a value at regular intervals.
func Tick(val fix.S17, interval int) Ticker {
	samples := 0
	return tickFn{
		name:    fmt.Sprintf("Tick(%v,%d)", val, interval),
		inputs:  0,
		outputs: 1,
		tick: func(_, outputs [][]fix.S17) {
			l := len(outputs[0])
			if n := interval - samples; n < l {
				outputs[0][n] = val
				samples = -n // this is cursed
			}
			samples += l
		},
	}
}

// Every creates a Ticker that outputs a value every dur.
func Every(val fix.S17, dur time.Duration, samplerate float32) Ticker {
	period := int(float64(samplerate) * dur.Seconds())
	return Tick(val, period)
}

// Once creates a ticker that outputs a value once, in the first sample it
// produces and produces entirely zeros after that.
func Once(val fix.S17) Ticker {
	done := false
	return tickFn{
		name:    "Once",
		inputs:  0,
		outputs: 1,
		tick: func(_, outputs [][]fix.S17) {
			for i := range outputs[0] {
				outputs[0][i] = 0
			}
			if !done {
				outputs[0][1] = val
				done = true
			}
		},
	}
}
