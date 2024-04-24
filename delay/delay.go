// package delay provides some delay lines.
package delay

import (
	"fmt"
	"time"

	"github.com/pfcm/fxp"
	"github.com/pfcm/fxp/fix"
	"github.com/pfcm/fxp/internal/buffer"
)

// Delay is an fxp.Ticker that provides a simple tape-style delay.
// Feedback is controlled by another ticker, that needs to have two inputs
// and one output.
// The inputs are:
//   - 0: the input audio, S17
//   - 1: the rate at which we read to and write from the delay line,
//     reinterpreted as a Rat44.
//
// TODO: multi tap
type Delay struct {
	rb   *buffer.Ring
	Rate float32
	fb   fxp.Ticker
	fbuf []fix.S17
}

func NewDelay(maxTime time.Duration, samplerate float32, fb fxp.Ticker) *Delay {
	if fb != nil && (fb.Inputs() != 2 || fb.Outputs() != 1) {
		panic(fmt.Errorf("%v: wrong inputs/outputs for delay feedback", fb))
	}
	samps := int(maxTime.Seconds() * float64(samplerate))
	rb := buffer.NewRing(2 * samps)
	rb.Writep = float32(samps)
	return &Delay{
		rb:   rb,
		Rate: 1.0,
		fb:   fb,
		fbuf: make([]fix.S17, 4096),
	}
}

func (*Delay) Inputs() int      { return 2 }
func (*Delay) Outputs() int     { return 1 }
func (d *Delay) String() string { return fmt.Sprintf("Delay(%d)", len(d.rb.Buf)) }

func (d *Delay) Tick(in, out [][]fix.S17) {
	// for _, s := range in[1] {
	// 	fmt.Printf("%v ", fix.InterpretAsRat44(s))
	// }
	// fmt.Println()
	d.rb.Read(out[0], in[1])
	if len(d.fbuf) != len(out[0]) {
		d.fbuf = d.fbuf[:len(out[0])]
	}
	toWrite := in[0]
	if d.fb != nil {
		d.fb.Tick([][]fix.S17{in[0], out[0]}, [][]fix.S17{d.fbuf})
		toWrite = d.fbuf
	}
	d.rb.Write(toWrite, in[1])
}
