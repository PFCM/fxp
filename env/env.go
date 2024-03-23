// package env provides envelope generators.
package env

import (
	"fmt"
	"time"

	"github.com/pfcm/fxp/fix"
)

type envState byte

const (
	idle envState = iota
	attack
	decay
	sustain
	release
)

// AD is a simple attack-decay envelope. It is triggered by any non-zero value,
// which causes it to ramp to 1 over the specified duration then back down. It
// resets as soon as it sees another non-zero value in the input regardless of
// where it was in its cycle.
type AD struct {
	nAttack int // in samples
	nDecay  int
	state   envState
	counter int
}

func AttackDecay(attack, decay time.Duration, samplerate float32) *AD {
	return &AD{
		nAttack: int(attack.Seconds() * float64(samplerate)),
		nDecay:  int(decay.Seconds() * float64(samplerate)),
	}
}

func (*AD) Inputs() int      { return 1 }
func (*AD) Outputs() int     { return 1 }
func (a *AD) String() string { return fmt.Sprintf("AD(%d,%d)", a.nAttack, a.nDecay) }

func (a *AD) Tick(in, out [][]fix.S17) {
	for i, s := range in[0] {
		if s != 0 {
			a.enter(attack)
		}
		switch a.state {
		case attack:
			out[0][i] = pos(0, a.counter, a.nAttack)
			a.counter++
			if a.counter >= a.nAttack {
				a.enter(decay)
			}
		case decay:
			out[0][i] = pos(0, a.nDecay-a.counter-1, a.nDecay)
			a.counter++
			if a.counter >= a.nDecay {
				a.enter(idle)
			}
		default:
			out[0][i] = 0
		}
	}
}

func (a *AD) enter(state envState) {
	a.state = state
	a.counter = 0
}

// pos returns a coefficient between 0 and 1 depending on where pos is between
// start and end.
func pos(start, pos, end int) fix.S17 {
	f := float32(pos-start) / float32(end-start)
	return fix.FromFloat(f)
}
