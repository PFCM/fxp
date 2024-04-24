package filter

import (
	"math"

	"golang.org/x/exp/constraints"
)

// uf8 is an unsigned 8-bit floating point with 5 significand bits and
// 3 exponent bits. The layout uses the highest 3 bits for the
// exponent and the lowest 5 for the significand.
type uf8 uint8

func uf8ToFloat[T constraints.Float](u uf8) T {
	s := T(u & 0x1f)
	s /= T(1 << 5)
	e := T(math.Pow(2, float64(u>>5)-1))
	return s * e
}
