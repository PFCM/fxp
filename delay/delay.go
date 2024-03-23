// package delay provides some delay lines.
package delay

import (
	"fmt"
	"time"

	"github.com/pfcm/fxp/fix"
)

// ring is an interpolating ring buffer.
// TODO: figure out an api for a multi-tap
type ring struct {
	buf    []fix.S17
	writep int
	readp  int
}

// newRing allocates a new ring buffer with the given number of samples. The
// write head and the read head start off in the same position to provide
// maximum delay; so be sure to read before writing. Or move the heads around,
// I'm not your mum.
func newRing(size int) *ring {
	return &ring{
		buf: make([]fix.S17, size),
	}
}

// write writes a chunk of samples to the buffer at the current position of the
// write head. Updates the write head.
func (r *ring) write(in []fix.S17) {
	if len(in) > len(r.buf) {
		panic(fmt.Errorf("input %d larger than buffer %d", len(in), len(r.buf)))
	}
	copied := copy(r.buf[r.writep:], in)
	if copied < len(in) {
		// we couldn't fit it all on the end.
		r.writep = copy(r.buf, in[copied:])
	} else {
		r.writep += copied
	}
}

// read reads a chunk of samples from the buffer at the current read
// head. Advances the read head.
func (r *ring) read(out []fix.S17) {
	if len(out) > len(r.buf) {
		panic(fmt.Errorf("output %d larger than buffer %d", len(out), len(r.buf)))
	}
	copied := copy(out, r.buf[r.readp:])
	if copied < len(out) {
		// reached the end of the buffer, wrap
		r.readp = copy(out[copied:], r.buf)
	} else {
		r.readp += copied
	}
}

// Delay is an fxp.Ticker that provides a simple tape-style delay.
// TODO: input parameters.
// TODO: external feedback
type Delay struct {
	rb *ring
}

func NewDelay(maxTime time.Duration, samplerate float32) *Delay {
	samps := int(maxTime.Seconds() * float64(samplerate))
	return &Delay{rb: newRing(samps)}
}

func (*Delay) Inputs() int      { return 1 }
func (*Delay) Outputs() int     { return 1 }
func (d *Delay) String() string { return fmt.Sprintf("Delay(%d)", len(d.rb.buf)) }

func (d *Delay) Tick(in, out [][]fix.S17) {
	d.rb.read(out[0])
	d.rb.write(in[0])
}
