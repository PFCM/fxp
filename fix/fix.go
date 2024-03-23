package fix

import (
	"fmt"

	"golang.org/x/exp/constraints"
)

// S17 is a signed (two's complement) 8 bit number with 1 integer bit and 7 franctional
// bits capable of representing (roughly) the range -1 to 1.
type S17 int8

const (
	// MaxS44 is the highest positive S17: 0.9921875.
	MaxS17 S17 = 0x7F
	// MinS17 is the lowest negative S17: -1.
	MinS17 S17 = -0x80
)

func (s S17) String() string {
	return fmt.Sprintf("%.7f", Float[float64](s))
}

// SAdd is a saturating +, clipping to the minimum or maximum value.
func (a S17) SAdd(b S17) S17 {
	// TODO: this can definitely be more efficient
	if a > 0 && b > 0 && a > MaxS17-b {
		return MaxS17
	}
	if a < 0 && b < 0 && a < MinS17+b {
		return MinS17
	}
	return a + b
}

// SMul multiplies an S17 with another, saturating at the maximum or minimum
// if it overflows.
func (a S17) SMul(b S17) S17 {
	return S17((int16(a) * int16(b)) >> 7)
}

func Float[T constraints.Float](s S17) T {
	// ideally this would be const, but apparently it can't be.
	var scale = 1.0 / T(1<<7)
	return T(s) * scale
}

// FromFloat converts a float into an S17, clamping to the maximum or minimum values.
func FromFloat[T constraints.Float](f T) S17 {
	if f < Float[T](MinS17) {
		return MinS17
	}
	if f > Float[T](MaxS17) {
		return MaxS17
	}
	return S17(f * T(1<<7))
}

// U62 is an unsigned fixed point number with 6 integer bits and 2 fractional bits,
// capable of representing 0 to 63.75.
type U62 uint8

const (
	// MaxU62 is the highest U62: 63.75
	MaxU62 U62 = 0xFF
	// MinU62 is the lowest U62: 0.
	MinU62 U62 = 0x00
)

func (s U62) String() string {
	return fmt.Sprintf("%.2f", U62ToFloat[float64](s))
}

func U62ToFloat[T constraints.Float](s U62) T {
	// ideally this would be const, but apparently it can't be.
	var scale = 1.0 / T(1<<2)
	return T(s) * scale
}

// U62FromFloat converts a float into an U62, clamping to the maximum or minimum values.
func U62FromFloat[T constraints.Float](f T) U62 {
	if f < U62ToFloat[T](MinU62) {
		return MinU62
	}
	if f > U62ToFloat[T](MaxU62) {
		return MaxU62
	}
	return U62(f * T(1<<2))
}
