package fix

import (
	"math"
	"slices"
	"sort"

	"golang.org/x/exp/constraints"
)

// Rat44 is a rational with 4 bits each of numerator and denominator,
// representing 1/16 (0.0625) to 16. It is handy sometimes for frequency/rate
// multipliers. It is similar to a float in that it isn't particularly efficient
// with its representation (for example there are a lot of ways to represent 1),
// but still has uses in situations where the linear spread of regular fixed
// points can be a waste.
// TODO: we may be able to just replace this with a float type.
type Rat44 uint8

const (
	MaxRat44 Rat44 = 0xf0 // 16/1
	MinRat44 Rat44 = 0x0f // 1/16
)

func Rat44ToFloat[T constraints.Float](r Rat44) T {
	num := (r >> 4) + 1
	den := (r & 0xf) + 1
	return T(num) / T(den)
}

func Rat44FromFloat[T constraints.Float](f T) Rat44 {
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

// rathouse is a table to convert floats to Rat44s: the conversion is tricky and
// there's less than 256 unique values so the easiest way is to just look up the
// answer.
var rathouse = func() []rat {
	rats := make([]rat, 256)
	for i := range rats {
		rats[i] = rat{
			r: Rat44(i),
			f: Rat44ToFloat[float64](Rat44(i)),
		}
	}
	sort.Slice(rats, func(i, j int) bool {
		ri, rj := rats[i], rats[j]
		if ri.f != rj.f {
			return ri.f < rj.f
		}
		return ri.r < rj.r
	})
	return slices.CompactFunc(rats, func(a, b rat) bool {
		return a.f == b.f
	})
}()
