// package filter provides filters.
package filter

import (
	"fmt"
	"math"

	"github.com/pfcm/fxp/fix"
)

// SVF is a state-variable filter, with the derivation mostly due to
// https://www.dafx14.fau.de/papers/dafx14_aaron_wishnick_time_varying_filters_for_.pdf
type SVF struct {
	s      [2]float32
	cutoff fix.U08
	i      int
}

func (SVF) Inputs() int    { return 1 }
func (SVF) Outputs() int   { return 1 }
func (SVF) String() string { return "SVF" }

func (s *SVF) Tick(in, out [][]fix.S17) {
	// TODO: use more fixes, reorganise to get bandpass and highpass
	var (
		r  = fix.U26FromFloat(0.1)
		rf = fix.U26ToFloat[float32](r)
		g  = fix.U08ToFloat[float32](s.cutoff)

		hf  = 1.0 / (g*g + 2*rf*g + 1)
		h00 = hf
		h01 = -g * hf
		h10 = g * hf
		h11 = (2*rf*g + 1) * hf

		x0, x1 float32
	)
	for i, u := range in[0] {
		uf := fix.S17ToFloat[float32](u)

		hs0 := s.s[0]*h00 + s.s[1]*h01
		hs1 := s.s[0]*h10 + s.s[1]*h11

		hu0 := h00 * uf
		hu1 := h10 * uf

		x0 = g*hu0 + hs0
		x1 = g*hu1 + hs1

		ax0 := -2*rf*x0 - x1
		ax1 := x0
		s.s[0] = s.s[0] + (2 * g * ax0)
		s.s[1] = s.s[1] + (2 * g * ax1)

		out[0][i] = fix.S17FromFloat(s.s[0])
	}
	fmt.Printf(`h:
[[%.7f, %.7f]
 [%.7f, %.7f]]
S17h:
[[%v, %v]
 [%v, %v]]
`, h00, h01, h10, h11, fix.S17FromFloat(h00), fix.S17FromFloat(h01), fix.S17FromFloat(h10), fix.S17FromFloat(h11))
	s.i++
	if s.i%10 == 0 {
		s.cutoff++
	}
}

// SVF2 is a state-variable filter, with the derivation mostly due to
// https://www.dafx14.fau.de/papers/dafx14_aaron_wishnick_time_varying_filters_for_.pdf
type SVF2 struct {
	s      [2]fix.S26
	cutoff fix.U08
	i      int
}

func (SVF2) Inputs() int    { return 1 }
func (SVF2) Outputs() int   { return 1 }
func (SVF2) String() string { return "SVF2" }

func (s *SVF2) Tick(in, out [][]fix.S17) {
	// TODO: use more fixes, reorganise to get bandpass and highpass
	var (
		r  = fix.U26FromFloat(0.15)
		rf = fix.U26ToFloat[float32](r)
		// todo lookup table
		g = fix.U08ToFloat[float32](s.cutoff)

		hf  = 1.0 / (g*g + 2*rf*g + 1)
		h00 = fix.S26FromFloat(hf)
		h01 = fix.S26FromFloat(-g * hf)
		h10 = fix.S26FromFloat(g * hf)
		h11 = fix.S26FromFloat((2*rf*g + 1) * hf)

		x0, x1 fix.S26
	)
	for i, u := range in[0] {
		// uf := fix.S17ToFloat[float32](u)

		// s0 := fix.S17ToFloat[float32](s.s[0])
		// s1 := fix.S17ToFloat[float32](s.s[1])

		hs0 := s.s[0].SMul(h00).SAdd(s.s[1].SMul(h01))
		hs1 := s.s[0].SMul(h10).SAdd(s.s[1].SMul(h11))

		hu0 := h00.SMulS17(u)
		hu1 := h10.SMulS17(u)

		x0 = hu0.SMulU08(s.cutoff).SAdd(hs0)
		x1 = hu1.SMulU08(s.cutoff).SAdd(hs1)

		// r is big, so this is trouble
		// ax0 := -2*rf*x0 - x1
		ax0 := fix.S26FromFloat(-2 * rf * fix.S26ToFloat[float32](x0)).SSub(x1)
		ax1 := x0
		c2 := fix.U17FromFloat(2.0).SMulU08(s.cutoff)
		s.s[0] = s.s[0].SAdd(ax0.SMulU17(c2))
		s.s[1] = s.s[1].SAdd(ax1.SMulU17(c2))

		out[0][i] = fix.S26ToS17(s.s[0])
	}
	if s.i == 0 {
		s.i = 1
		s.cutoff = 17
	}
	// s.cutoff = fix.U08(int(s.cutoff) + s.i)
	// if s.cutoff == 50 || s.cutoff == 15 {
	// 	s.i = -s.i
	// }
	s.i++
	if s.i%10 == 0 {
		s.cutoff++
	}
}

// Ladder provides a Moog-style 4 pole ladder filter.
// TODO: not just lowpass?
type Ladder struct {
	states [4]fix.S17
	cutoff fix.U08
	i      int
}

func (Ladder) Inputs() int    { return 1 } // audio, cutoff and Q
func (Ladder) Outputs() int   { return 1 }
func (Ladder) String() string { return "Ladder" }

func (l *Ladder) Tick(in, out [][]fix.S17) {
	// TODO: look up table for this, can also pre-warping etc
	// g = cutoff*sample_period/2
	G := fix.U08ToU26(l.cutoff).SMulU08(l.cutoff).SMulU08(l.cutoff).SMulU08(l.cutoff)
	k := fix.U26FromFloat(4.0)
	gk := G.SMul(k)
	cf := fix.U08ToFloat[float32](l.cutoff)
	lpg := fix.U17FromFloat(cf / (1 + cf))
	_ = lpg
	for i := range in[0] {
		// TODO: rearranging
		S := l.states[3]
		S = S.SAdd(l.states[2].SMulU08(l.cutoff))
		S = S.SAdd(l.states[1].SMulU08(l.cutoff.SMul(l.cutoff)))
		S = S.SAdd(l.states[0].SMulU08(l.cutoff.SMul(l.cutoff).SMul(l.cutoff)))

		// u := in[0][i].SSubU26(gk)
		// This is where we need to divide :(
		u := fix.S17FromFloat(
			(fix.S17ToFloat[float32](in[0][i]) -
				fix.U26ToFloat[float32](gk)) /
				(1 + fix.U26ToFloat[float32](gk)))
		// u is the input to the integrators, which are all chained.
		v := u.SSub(l.states[0]).SMulU17(lpg)
		y := v.SAdd(l.states[0])
		l.states[0] = y.SAdd(v)
		v = y.SSub(l.states[1]).SMulU17(lpg)
		y = v.SAdd(l.states[1])
		l.states[1] = y.SAdd(v)
		v = y.SSub(l.states[2]).SMulU17(lpg)
		y = v.SAdd(l.states[2])
		l.states[2] = y.SAdd(v)
		v = y.SSub(l.states[3]).SMulU17(lpg)
		y = v.SAdd(l.states[3])
		l.states[3] = y.SAdd(v)

		out[0][i] = y
	}
	//	fmt.Println(l.states)
	l.cutoff += 2
}

type FloatLadder struct {
	states [4]float32
	zi     float32
	// cutoff is tan(pi*hz/samplerate), noting that with a samplerate of
	// 44100, a value of 1 is roughly 11000 Hz, so the filter is never
	// totally wide open. This does however mean tan is operating in a
	// fairly linear range, which helps us get away with using a fixed point
	// type for this and skipping the tan.
	cutoff fix.U08
	i      int
}

func (FloatLadder) Inputs() int    { return 1 } // audio, cutoff and Q
func (FloatLadder) Outputs() int   { return 1 }
func (FloatLadder) String() string { return "FloatLadder" }

func (l *FloatLadder) Tick(in, out [][]fix.S17) {
	// from https://www.kvraudio.com/forum/viewtopic.php?f=33&t=349859,
	// with credit to Mystran
	// resonance should range from 0~4.5, for now 0-4 will have to do
	r := fix.U26FromFloat(4.0)
	rf := fix.U26ToFloat[float32](r)
	// cf := float32(math.Tan(math.Pi * fix.U08ToFloat[float64](l.cutoff)))
	cf := fix.U08ToFloat[float32](l.cutoff)
	for i := range in[0] {
		inp := fix.S17ToFloat[float32](in[0][i])
		// input with half delay
		ih := (0.5 * (inp + l.zi))
		l.zi = fix.S17ToFloat[float32](in[0][i])
		var (
			s0 = l.states[0]
			s1 = l.states[1]
			s2 = l.states[2]
			s3 = l.states[3]
			// TODO: use more fixes?
			t0 = tanhxdx(ih - (rf * s3))
			t1 = tanhxdx(s0)
			t2 = tanhxdx(s1)
			t3 = tanhxdx(s2)
			t4 = tanhxdx(s3)
			// TODO these could probably stand some rearranging
			g0 = 1 / (1 + cf*t1)
			g1 = 1 / (1 + cf*t2)
			g2 = 1 / (1 + cf*t3)
			g3 = 1 / (1 + cf*t4)

			f3 = cf * t3 * g3
			f2 = cf * t2 * g2 * f3
			f1 = cf * t1 * g1 * f2
			f0 = cf * t0 * g0 * f1

			y3 = (g3*s3 + f3*g2*s2 + f2*g1*s1 + f1*g0*s0 + f0*inp) / (1 + rf*f0)

			xx = t0 * (inp - rf*y3)
			y0 = t1 * g0 * (s0 + cf*xx)
			y1 = t2 * g1 * (s1 + cf*y0)
			y2 = t3 * g2 * (s2 + cf*y1)
		)
		l.states[0] += (2 * cf * (xx - y0))
		l.states[1] += (2 * cf * (y0 - y1))
		l.states[2] += (2 * cf * (y1 - y2))
		l.states[3] += (2 * cf * (y2 - t4*y3))

		out[0][i] = fix.S17FromFloat(y3)
	}
	l.i++
	if l.i%10 == 0 {
		l.cutoff++
	}
	fmt.Println(math.Atan(float64(cf))/math.Pi*44100, l.states)
}

type TwoP struct {
	states [2]fix.S17
	// cutoff is tan(pi*hz/samplerate), noting that with a samplerate of
	// 44100, a value of 1 is roughly 11000 Hz, so the filter is never
	// totally wide open. This does however mean tan is operating in a
	// fairly linear range, which helps us get away with using a fixed point
	// type for this and skipping the tan.
	cutoff fix.U08
	i      int
}

func (TwoP) Inputs() int    { return 1 } // audio, cutoff and Q
func (TwoP) Outputs() int   { return 1 }
func (TwoP) String() string { return "TwoP" }

func (l *TwoP) Tick(in, out [][]fix.S17) {
	r := fix.U17FromFloat(1.0)
	rf := fix.U17ToFloat[float32](r)
	// cf := float32(math.Tan(math.Pi * fix.U08ToFloat[float64](l.cutoff)))
	cf := fix.U08ToFloat[float32](l.cutoff)
	// c2 := l.cutoff.SMul(l.cutoff)
	f := rf + rf/(1.0-cf)
	for i := range in[0] {
		inps := fix.S17ToFloat[float32](in[0][i].SSub(l.states[0]))
		d := fix.S17ToFloat[float32](l.states[0].SSub(l.states[1]))
		l.states[0] = l.states[0].SAdd(fix.S17FromFloat(cf * (inps - f*d)))
		l.states[1] = l.states[1].SAdd(fix.S17FromFloat(cf * d))
		out[0][i] = l.states[1]
		if i == 0 {
			fmt.Println(f, l.states)
		}
	}
	l.i++
	if l.i%10 == 0 {
		l.cutoff++
	}
	// fmt.Println(math.Atan(float64(cf))/math.Pi*44100, l.states)
}

func tanhxdx(x float32) float32 {
	if math.Abs(float64(x)) < 1e-4 {
		return 1
	}
	return float32(math.Tanh(float64(x))) / x
}
