package fix

import (
	"fmt"
	"math"
	"sort"

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
	if a < 0 && b < 0 && a < MinS17-b {
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
	if f <= Float[T](MinS17) {
		return MinS17
	}
	if f >= Float[T](MaxS17) {
		return MaxS17
	}
	return S17(math.Round(float64(f * T(1<<7))))
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

// U17 is an unsigned fixed point with 1 integer bit and 6 fractional bits, capable of
// representing 0 to just under two.
type U17 uint8

const (
	// MaxU17 is the highest U17: 1.9375.
	MaxU17 U17 = 0xFF
	// MinU17 is the lowest U17: 0.
	MinU17 U17 = 0
)

func U17ToFloat[T constraints.Float](u U17) T {
	var scale = 1.0 / T(1<<7)
	return T(u) * scale
}

func U17FromFloat[T constraints.Float](f T) U17 {
	if f <= 0 {
		return MinU17
	}
	if f >= U17ToFloat[T](MaxU17) {
		return MaxU17
	}
	return U17(f * T(1<<7))
}

func (u U17) SMul(v U17) U17 {
	panic("not implemented")
}

func (u U17) SAdd(v U17) U17 {
	panic("not implemented")
}

func (u U17) String() string {
	return fmt.Sprintf("%.7f", U17ToFloat[float32](u))
}

// SMulU17S17S17 performs a saturating multiply between an S17 and a U17, returning
// the result in an S17.
func SMulU17S17S17(u U17, s S17) S17 {
	rounder := int16(0x80)
	if s < 0 {
		rounder = -0x80
	}
	v := S17((int16(uint16(u)*uint16(s)) + rounder) >> 7)
	if s > 0 && v < 0 {
		return MaxS17
	}
	if s < 0 && v > 0 {
		return MinS17
	}
	return v
}

// Rat44 is a rational with 4 bits each of numerator and denominator,
// representing 1/16 (0.0625) to 16. It is intended for passing around
// multipliers (as it has the same number of bits to represent 1-16 as 0-1), so
// the only methods are for converting to and from floats.
// Note that it can not represent zero: Rat44ToFloat(Rat44(0)) = 1.
type Rat44 uint8

const (
	MaxRat44 Rat44 = 0xF0 // 16/1
	MinRat44 Rat44 = 0x0f // 1/16
)

func InterpretAsRat44(s S17) float32 {
	return Rat44ToFloat[float32](Rat44(s))
}

func Rat44ToFloat[T constraints.Float](r Rat44) T {
	num := (r >> 4) + 1  // high four bytes are the numerator
	den := (r & 0xf) + 1 // low four bytes are the denominator
	return T(num) / T(den)
}

func FloatToRat44[T constraints.Float](f T) Rat44 {
	g := float64(f)
	i := sort.Search(len(rathouse), func(j int) bool {
		return rathouse[j].f >= g
	})
	if i < len(rathouse) {
		if i > 0 && math.Abs(g-rathouse[i].f) > math.Abs(g-rathouse[i-1].f) {
			i--
		}
		return rathouse[i].r
	}
	return MaxRat44
}

type rat struct {
	r Rat44
	f float64
}

// converting floats to rationals is pretty tricky, but we only have 256 values
// (which aren't even all unique because how many ways there are to write 1.0),
// so just build a table.
var rathouse = func() []rat {
	rats := make([]rat, 256)
	for i := range rats {
		rats[i] = rat{
			r: Rat44(i),
			f: Rat44ToFloat[float64](Rat44(i)),
		}
	}
	sort.Slice(rats, func(i, j int) bool {
		if rats[i].f != rats[j].f {
			return rats[i].f < rats[j].f
		}
		return rats[i].r < rats[j].r
	})
	return rats
}()
