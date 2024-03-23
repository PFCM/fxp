package main

import (
	"context"
	"fmt"
	"log"
	"math"
	"os"
	"os/signal"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/pfcm/fxp"
	"github.com/pfcm/fxp/delay"
	"github.com/pfcm/fxp/env"
	"github.com/pfcm/fxp/fix"
	"github.com/pfcm/fxp/io"
	"github.com/pfcm/fxp/osc"
)

func s17s(fs ...float32) []fix.S17 {
	out := make([]fix.S17, len(fs))
	for i, f := range fs {
		out[i] = fix.FromFloat(f)
	}
	return out
}

func main() {
	g, ctx := errgroup.WithContext(interruptContext())
	c := newCopier(1)
	ch := fxp.Serially(
		fxp.Concurrently(
			// some oscillators
			fxp.Serially(
				fxp.Const{fix.S17(fix.U62FromFloat(float32(48)))},
				osc.Sine(44100, 0),
			),
			// fxp.Serially(
			// 	fxp.Const{fix.S17(fix.U62FromFloat(float32(55)))},
			// 	osc.Sine(44100, 0),
			// ),
			// fxp.Serially(
			// 	fxp.Const{fix.S17(fix.U62FromFloat(float32(60)))},
			// 	osc.Sine(44100, 12),
			// ),
		),
		fxp.Concurrently(
			// generate an envelope
			fxp.Serially(
				fxp.Every(1, 1*time.Second, 44100),
				env.AttackDecay(50*time.Millisecond, 200*time.Millisecond, 44100),
			),
			// mix the oscillators together
			fxp.Sum(1),
		),
		// apply the envelope
		fxp.Amp{},
		// send it to a delay with some wet/dry mix
		// TODO: feedback :(
		fxp.Mult{2}, // broadcast
		fxp.Concurrently(
			delay.NewDelay(1700*time.Millisecond, 44100),
			fxp.Noop{1},
		),
		fxp.Mixer{Gains: s17s(0.5, 0.5)},
		c,
	)

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
				fmt.Printf("\r%v: %v", time.Since(t0), s)
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
