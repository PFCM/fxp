// Code generated by by github.com/pfcm/fxp/fix/gen DO NOT EDIT.

package fix

import (
	"fmt"

	"golang.org/x/exp/constraints"
)

// U08 is an unsigned 8 bit fixed point number with 8
// fractional bits, representing numbers between 0 and 0.99609375, in steps
// of 0.0038909912109375.
// Especially useful for scaling and interpolation.
type U08 uint8

const (
	// MinU08 is the smallest U08: 0.
	MinU08 U08 = 0x00
	// MaxU08 is the largest U08: 0.99609375.
	MaxU08 U08 = 0xff
)

// U08ToFloat converts a U08 to a float value.
func U08ToFloat[T constraints.Float](u U08) T {
	return T(u) * 0.00390625
}

// U08FromFloat returns the nearest U08 to the provided
// float value.
func U08FromFloat[T constraints.Float](f T) U08 {
	if f < 0 {
		return 0
	}
	if f > 0.99609375 {
		return 0xff
	}
	// TODO: rounding? Then we would have to do it in SMul etc.
	return U08((f /*+0.00194549560546875*/) * 256)
}

// String returns a string representation of a U08.
func (u U08) String() string {
	return fmt.Sprintf("%.8f", U08ToFloat[float64](u))
}

// SAdd is a saturating addition betwenn two U08.
func (u U08) SAdd(v U08) U08 {
	return U08(usadd(uint8(u), uint8(v)))
}

// SSub is a saturating subtraction between two U08,
// subtracting v from u.
func (u U08) SSub(v U08) U08 {
	return U08(ussub(uint8(u), uint8(v)))
}

// SMul is a saturating multiply between two U08.
func (u U08) SMul(v U08) U08 {
	return U08(usmul(uint8(u), uint8(v), 8))
}
