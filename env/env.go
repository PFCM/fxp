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

func (e envState) String() string {
	return []string{
		idle:    "x",
		attack:  "A",
		decay:   "D",
		sustain: "S",
		release: "R",
	}[e]
}

// ADSR is an attack-decay-sustain-release envelope. Any non-zero value in the
// input triggers the attack, which ramps to 1 then decays down to the sustain
// factor. When the input returns to zero it will begin the decay back down to
// zero.
type ADSR struct {
	nAttack  int // samples
	nDecay   int
	sus      fix.S17 // TODO: U17
	nRelease int
	state    envState
	counter  int

	last fix.S17
}

func NewADSR(attack, decay time.Duration,
	sustain fix.S17,
	release time.Duration,
	samplerate float32) *ADSR {
	return &ADSR{
		nAttack:  int(attack.Seconds() * float64(samplerate)),
		nDecay:   int(decay.Seconds() * float64(samplerate)),
		sus:      sustain,
		nRelease: int(release.Seconds() * float64(samplerate)),
	}
}

func (*ADSR) Inputs() int  { return 1 }
func (*ADSR) Outputs() int { return 1 }
func (a *ADSR) String() string {
	return fmt.Sprintf("ADSR(%v,%v,%v,%v)", a.nAttack, a.nDecay, a.sus, a.nRelease)
}

func (a *ADSR) Tick(in, out [][]fix.S17) {
	for i, s := range in[0] {
		if s == 0 && a.last != 0 {
			a.enter(release)
		}
		if a.last == 0 && s != 0 {
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
			d := fix.MaxS17.SAdd(-a.sus)
			out[0][i] = a.sus.SAdd(d.SMul(pos(0, a.nDecay-a.counter-1, a.nDecay)))
			a.counter++
			if a.counter >= a.nDecay {
				a.enter(sustain)
			}
		case sustain:
			out[0][i] = a.sus
		case release:
			out[0][i] = a.sus.SMul(pos(0, a.nRelease-a.counter-1, a.nRelease))
			a.counter++
			if a.counter >= a.nRelease {
				a.enter(idle)
			}
		}
		a.last = s
	}
}

func (a *ADSR) enter(state envState) {
	a.state = state
	a.counter = 0
}

// AD is a simple attack-decay envelope. It is triggered by any non-zero value,
// which causes it to ramp to 1 over the specified duration then back down. It
// resets as soon as it sees a zero value followed by a non-zero in the input
// regardless of where it was in its cycle.
type AD struct {
	nAttack int // in samples
	nDecay  int
	state   envState
	counter int

	last fix.S17
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
		if a.last == 0 && s != 0 {
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
		a.last = s
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
	return fix.S17FromFloat(f)
}
