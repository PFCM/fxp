package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"os/signal"
	"runtime"
	"runtime/pprof"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/pfcm/fxp"
	"github.com/pfcm/fxp/delay"
	"github.com/pfcm/fxp/env"
	"github.com/pfcm/fxp/fix"
	"github.com/pfcm/fxp/hid"
	"github.com/pfcm/fxp/io"
	"github.com/pfcm/fxp/osc"
)

var (
	profileFlag = flag.Bool("profile", false, "whether to write pprof profiles to the current working directory")
)

func s17s(fs ...float32) []fix.S17 {
	out := make([]fix.S17, len(fs))
	for i, f := range fs {
		out[i] = fix.FromFloat(f)
	}
	return out
}

// c makes a fix.S17 const ticker with the nearest value.
func c(f float32) fxp.Ticker {
	return fxp.Const{fix.FromFloat(f)}
}

// uc makes an fix.U62 const ticker.
func uc(f float32) fxp.Ticker {
	return fxp.Const{fix.S17(fix.U62FromFloat(f))}
}

// o makes a sine wave oscillator with the provided serial chain of tickers
// as the frequency input.
func o(min int, inps ...fxp.Ticker) fxp.Ticker {
	os := osc.Square(44100, -128, 127, min)
	if len(inps) == 0 {
		return os
	}
	return fxp.Serially(append(inps, os)...)
}

func delays() fxp.Ticker {
	return fxp.Serially(
		fxp.Concurrently(
			// some oscillators
			o(48,
				// lfo
				// fxp.Const{0},
				// osc.RatSine(44100, -60, 0.5),
				// osc.Sine(44100, -60),
				// fxp.Scale{
				// 	Mul:   fix.FromFloat[float32](0.01),
				// 	Shift: fix.FromFloat[float32](0.01),
				// },
				uc(4),
			),
			o(48, uc(0)),
			o(48, uc(7)),
		),
		fxp.Concurrently(
			// generate an envelope
			fxp.Serially(
				// fxp.Every(1, 1000*time.Millisecond, 44100),
				fxp.Once(1),
				env.AttackDecay(2000*time.Millisecond, 2000*time.Millisecond, 44100),
			),
			// mix the oscillators together
			// fxp.Sum(2),
			fxp.Mix(s17s(0.4, 0.4, 0.4)),
		),
		// apply the envelope
		fxp.Amp{},
		// delay.NewDelay(500*time.Millisecond, 44100),
		// send it to a delay with some wet/dry mix
		fxp.Mult{2}, // broadcast
		fxp.Concurrently(
			fxp.Serially(
				fxp.Mult{2}, // broadcast
				fxp.Concurrently(
					fxp.Serially(
						fxp.Concurrently(
							fxp.Noop{1},
							fxp.Serially(
								// fxp.Const{fix.S17(fix.FloatToRat44(float32(0.5)))},
								fxp.Const{fix.S17(0)},
								// osc.Square(44100,
								// 	fix.S17(fix.FloatToRat44(float32(1.1))),
								// 	fix.S17(fix.FloatToRat44(float32(1.0/1.1))),
								// 	-48,
								// ),
								osc.RatSine(44100, -66, 1.05),
								// osc.Sine(44100, -60),
								// fxp.Scale{
								// 	Mul:   fix.FromFloat(float32(0.5)),
								// 	Shift: fix.FromFloat(float32(1)),
								// },
							),
						),
						delay.NewDelay(1000*time.Millisecond,
							44100, fxp.Serially(
								fxp.Mix(s17s(0.9, 0.5)),
							)),
					),
					fxp.Noop{1},
				),
				fxp.Mix(s17s(0.3, 0.9)),
			),
			fxp.Serially(
				fxp.Mult{2},
				fxp.Concurrently(
					fxp.Serially(
						fxp.Concurrently(
							fxp.Noop{1},
							fxp.Serially(
								fxp.Const{fix.S17(0)},
								osc.RatSine(44100, -60, 1.1),
								// osc.Sine(44100, -60),
								// fxp.Scale{
								// 	Mul:   fix.FromFloat(float32(0.5)),
								// 	Shift: fix.FromFloat(float32(1)),
								// },
							),
						),
						delay.NewDelay(1100*time.Millisecond,
							44100, fxp.Serially(
								fxp.Mix(s17s(0.9, 0.5)),
							)),
					),
					fxp.Noop{1},
				),
				fxp.Mix(s17s(0.3, 0.9)),
			),
		),
	)
}

func midikeys(n int) fxp.Ticker {
	voice := func() fxp.Ticker {
		return fxp.Serially(
			fxp.Concurrently(
				o(0),
				env.NewADSR(
					200*time.Millisecond,
					200*time.Millisecond,
					fix.FromFloat(float32(0.5)),
					2000*time.Millisecond,
					44100),
			),
			fxp.Amp{},
		)
	}
	voices := make([]fxp.Ticker, n)
	for i := range voices {
		voices[i] = voice()
	}
	return fxp.Serially(
		hid.NewMidiNotes(n),
		fxp.Concurrently(voices...),
		fxp.Sum(n),
	)
}

func main() {
	flag.Parse()

	if *profileFlag {
		finish, err := startProfiles()
		if err != nil {
			log.Fatalf("Starting profiling: %v", err)
		}
		defer func() {
			if err := finish(); err != nil {
				log.Fatalf("Finishing profiles: %v", err)
			}
		}()
	}

	g, ctx := errgroup.WithContext(interruptContext())

	//t := delays()
	t := midikeys(4)

	c := newCopier(t.Outputs())
	ch := fxp.Serially(t, c)

	g.Go(func() error {
		return io.PlayWithDefaults(ctx, ch)
	})
	g.Go(func() error {
		t0 := time.Now()
		t := time.NewTicker(100 * time.Millisecond)
		defer t.Stop()
		for {
			select {
			case <-ctx.Done():
				return nil
			case <-t.C:
				var s []string
				for _, f := range c.getRMS() {
					s = append(s, fmt.Sprintf("%.2f", f))
				}
				fmt.Printf("\r%.4f: %v", time.Since(t0).Seconds(), s)
			}
		}
	})

	if err := g.Wait(); err != nil {
		log.Fatal(err)
	}
}

type copier struct {
	channels int

	mu  sync.Mutex
	rms []float32
}

func newCopier(channels int) *copier {
	return &copier{
		channels: channels,
		rms:      make([]float32, channels),
	}
}

func (c *copier) Inputs() int    { return c.channels }
func (c *copier) Outputs() int   { return c.channels }
func (c *copier) String() string { return fmt.Sprintf("copier(%d)", c.channels) }

func (c *copier) Tick(in, out [][]fix.S17) {
	for i, inp := range in {
		copy(out[i], inp)
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	for i, channel := range in {
		rms := float64(0)
		for _, s := range channel {
			rms += float64(s) * float64(s)
		}
		rms /= float64(len(channel))
		c.rms[i] = 0.01*c.rms[i] + 0.99*float32(math.Sqrt(rms))
	}
}

func (c *copier) getRMS() []float32 {
	results := make([]float32, c.channels)
	c.mu.Lock()
	defer c.mu.Unlock()
	copy(results, c.rms)
	return results
}

func interruptContext() context.Context {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	go func() {
		<-ctx.Done()
		cancel()
	}()
	return ctx
}

func startProfiles() (func() error, error) {
	cpu, err := os.Create("cpu.pprof")
	if err != nil {
		return nil, err
	}
	if err := pprof.StartCPUProfile(cpu); err != nil {
		return nil, fmt.Errorf("starting cpu profile: %w", err)
	}

	mem, err := os.Create("mem.pprof")
	if err != nil {
		return nil, err
	}
	return func() error {
		pprof.StopCPUProfile()
		if err := cpu.Close(); err != nil {
			return err
		}
		runtime.GC()
		if err := pprof.WriteHeapProfile(mem); err != nil {
			return err
		}
		return mem.Close()
	}, nil
}
