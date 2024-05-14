package fxp

import (
	"fmt"

	"github.com/pfcm/fxp/fix"
)

// LFSR is a ticker that uses an 8 bit linear-feedback shift register to make
// noise.
type LFSR struct {
	state uint16
	taps  uint16
}

const defaultTaps uint16 = 0xd008

func Noise() *LFSR {
	return &LFSR{
		state: 0xffff,
		taps:  defaultTaps,
	}
}

func (*LFSR) Inputs() int      { return 0 }
func (*LFSR) Outputs() int     { return 1 }
func (l *LFSR) String() string { return fmt.Sprintf("LFSR(%2x)", l.taps) }

func (l *LFSR) Tick(in, out [][]fix.S17) {
	for i := range out[0] {
		fb := l.state & 1
		l.state >>= 1
		if fb == 1 {
			l.state ^= l.taps
		}
		out[0][i] = fix.S17(l.state)
	}
}
