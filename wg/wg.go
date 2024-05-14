// package wg implements digital waveguide-ish algorithms.
package wg

import (
	"math"
	"math/rand"

	"github.com/pfcm/fxp"
	"github.com/pfcm/fxp/fix"
	"github.com/pfcm/fxp/interp"
)

// KS implements a fairly straightforward, original Karplus-Strong algorithm
// focused on string synthesis. It accepts two inputs: the first is a trigger
// which will restart the excitation and the second is a MIDI note that decides
// the fundamental frequency.
type KS struct {
	samplePeriod float32

	ex []fix.S17 // the excitation
	// TODO position in ex
	tab []fix.S17 // delay buffer
	// TODO internal/buffer?
	// TODO fractional tpos for varying pitch?
	tpos        float32
	prev        fix.S17
	prevV       fix.U08
	prevPreHigh fix.S17
}

// TODO: provide the excitation etc.
func NewKS(samplerate float32) *KS {
	tab := make([]fix.S17, 1<<16)
	ex := make([]fix.S17, len(tab))

	// for i := range ex[:len(ex)/3] {
	// 	switch i % 3 { //rand.Intn(3) {
	// 	case 0:
	// 		ex[i] = 0x7f
	// 	case 1:
	// 		// not -0x80, try and avoid any dc to
	// 		// keep the signal centered.
	// 		ex[i] = -0x7f
	// 	}
	// }
	l := fxp.Noise()
	l.Tick(nil, [][]fix.S17{ex})
	return &KS{
		samplePeriod: 1.0 / samplerate,
		tab:          tab,
		ex:           ex,
	}
}

func (*KS) Inputs() int    { return 2 } // TODO: fine-tuning?
func (*KS) Outputs() int   { return 1 }
func (*KS) String() string { return "KS" }

const r fix.S17 = fix.MaxS17 - 2

func (k *KS) Tick(in, out [][]fix.S17) {
	// TODO: varying frequencies
	// TODO: initial excitation.
	for i := range out[0] {
		v := fix.U08(in[1][i])
		if k.prevV == 0 && v != 0 {
			// round the position in the table down, just to make it
			// easy to copy. We're starting fresh so it shouldn't
			// make any audible difference.
			// TODO: is copying the excitation without resampling
			// the right thing to do? Probably not?
			t := int(k.tpos)
			k.tpos = float32(t)
			rand.Shuffle(len(k.ex), func(i, j int) {
				k.ex[i], k.ex[j] = k.ex[j], k.ex[i]
			})
			ringCopy(k.tab, t, k.ex)
		}
		k.prevV = v
		// v = v.SAdd(64)
		// TODO: do this better
		b := fix.U08FromFloat(1.0 - fix.U08ToFloat[float32](v))
		// follow with a very low high pass to try and block any DC from
		// swamping the signal.
		// This works at rest, but we still get a weird offset when
		// v > 0.
		preHigh := get(k.tab, k.tpos).SMulU08(v).SAdd(k.prev.SMulU08(b).SAdd(1))
		output := preHigh.SSub(k.prevPreHigh).SAdd(r.SMul(k.prev))
		k.prevPreHigh = preHigh
		out[0][i] = output

		step := k.step(fix.U71(in[0][i]))
		nextTpos := k.tpos + step
		for wp := int(k.tpos) + 1; float32(wp) <= nextTpos; wp++ {
			c := (float32(wp) - k.tpos) / step
			a := interp.L(output, k.prev, fix.S17FromFloat(c))
			k.tab[wp%len(k.tab)] = a
		}

		k.prev = output
		k.tpos = nextTpos
		if l := float32(len(k.tab)); k.tpos >= l {
			k.tpos -= l
		}
	}
}

func get(src []fix.S17, pos float32) fix.S17 {
	var (
		i = int(pos)
		j = (i + 1) % len(src)
		c = pos - float32(i)
	)
	return interp.L(src[i], src[j], fix.S17FromFloat(c))
}

func (k *KS) step(note fix.U71) float32 {
	f := u71Freq[note]
	// f is wavetables per second, calculate target wavetable samples per
	// second.
	ts := f * float32(len(k.tab))
	// Final result is the number of wavetable samples per output sample.
	return k.samplePeriod * ts
}

// tables of fix.U71 to frequencies in Hertz.
var u71Freq = func() (fs [256]float32) {
	for i := 0; i < 256; i++ {
		n := fix.U71ToFloat[float64](fix.U71(i))
		fs[i] = float32(math.Pow(2.0, (n-69)/12)) * 440
	}
	return fs
}()

func ringCopy(dst []fix.S17, dpos int, src []fix.S17) {
	// assumes len(src) < = len(dst)
	n := copy(dst[dpos:], src)
	if n < len(src) {
		copy(dst, src[n:])
	}
}

// KS2 implements a fairly straightforward, original Karplus-Strong algorithm
// focused on string synthesis. It accepts two inputs: the first is a trigger
// which will restart the excitation and the second is a MIDI note that decides
// the fundamental frequency.
type KS2 struct {
	samplePeriod float32
	// TODO position in ex
	tab []fix.S17 // delay buffer
	// TODO internal/buffer?
	// TODO fractional tpos for varying pitch?
	tpos        float32
	prev        fix.S17
	prevV       fix.U08
	prevPreHigh fix.S17
}

func NewKS2(samplerate float32) *KS2 {
	tab := make([]fix.S17, 1<<11)
	return &KS2{
		samplePeriod: 1.0 / samplerate,
		tab:          tab,
	}
}

func (*KS2) Inputs() int    { return 3 } // TODO: fine-tuning?
func (*KS2) Outputs() int   { return 1 }
func (*KS2) String() string { return "KS2" }

func (k *KS2) Tick(in, out [][]fix.S17) {
	for i := range out[0] {
		v := fix.U08(in[2][i])
		k.prevV = v
		// v = v.SAdd(64)
		// TODO: do this better
		b := fix.U08FromFloat(1.0 - fix.U08ToFloat[float32](v))
		// follow with a very low high pass to try and block any DC from
		// swamping the signal.
		// This works at rest, but we still get a weird offset when
		// v > 0.
		y, x := in[0][i], get(k.tab, k.tpos)
		z := y.SAdd(x)
		preHigh := z.SMulU08(v).SAdd(k.prev.SMulU08(b).SAdd(1))
		output := preHigh.SSub(k.prevPreHigh).SAdd(r.SMul(k.prev))
		k.prevPreHigh = preHigh
		out[0][i] = output

		step := k.step(fix.U71(in[1][i]))
		nextTpos := k.tpos + step
		for wp := int(k.tpos) + 1; float32(wp) <= nextTpos; wp++ {
			c := (float32(wp) - k.tpos) / step
			a := interp.L(output, k.prev, fix.S17FromFloat(c))
			k.tab[wp%len(k.tab)] = a
		}
		k.prev = output
		k.tpos = nextTpos
		if l := float32(len(k.tab)); k.tpos >= l {
			k.tpos -= l
		}
	}
}

func (k *KS2) step(note fix.U71) float32 {
	f := u71Freq[note]
	// f is wavetables per second, calculate target wavetable samples per
	// second.
	ts := f * float32(len(k.tab))
	// Final result is the number of wavetable samples per output sample.
	return k.samplePeriod * ts
}
