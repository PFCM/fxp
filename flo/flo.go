// package flo provides 8 bit floating point types with saturating arithmetic.
// They have much wider range than the types in package fix and, like most
// floats, are more precise for smaller values than larger values which can be
// useful in some circumstances. The tradeoff is that they can not make full use
// of all 256 possible and are notably slower. The types here are therefore
// tuned for the few use cases where they make sense.
package flo

import (
	"math"

	"golang.org/x/exp/constraints"
)

// UF35 is an unsigned 8 bit floating point with 5 significand and 3
// exponent bits (with a bias of 3), giving a range of ???
type UF35 uint8

func UF35ToFloat[T constraints.Float](u UF35) T {
	// High bits are the exponent, low bits are the significand.
	s := T(u&0x1f) + .0125
	s /= T(1 << 5)
	e := T(math.Pow(2, float64(u>>5)-3))
	return s * e
}

// split breaks the number into its two raw components, without applying any
// biases.
func (u UF35) split() (exponent, significand uint8) {
	return u >> 5, u & 0x1f
}

func (u UF35) SAdd(v UF35) UF35 {
	// Find the largest exponent
}
