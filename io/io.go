// package io does audio in and out.
package io

import (
	"context"
	"encoding/binary"
	"fmt"
	"math"
	"os"

	"github.com/gen2brain/malgo"

	"github.com/pfcm/fxp"
	"github.com/pfcm/fxp/fix"
)

// PlayWithDefaults uses the default input and outputs to run the provided
// Ticker. It blocks until the provided context is cancelled.
func PlayWithDefaults(ctx context.Context, t fxp.Ticker) error {
	mctx, err := malgo.InitContext(nil, malgo.ContextConfig{}, func(msg string) {
		fmt.Fprint(os.Stderr, msg)
	})
	if err != nil {
		return err
	}
	defer func() {
		mctx.Uninit()
		mctx.Free()
	}()
	cfg := malgo.DefaultDeviceConfig(malgo.Duplex)
	// TODO: ???
	inps := max(1, t.Inputs())
	cfg.Capture.Format = malgo.FormatF32
	cfg.Capture.Channels = uint32(inps)
	cfg.Playback.Format = malgo.FormatF32
	cfg.Playback.Channels = uint32(t.Outputs())
	cfg.SampleRate = 44100

	// TODO: do we know the sizes ahead of the first recv call?
	inputs := make([][]fix.S17, inps)
	for i := range inputs {
		inputs[i] = make([]fix.S17, 4096)
	}
	outputs := make([][]fix.S17, t.Outputs())
	for i := range outputs {
		outputs[i] = make([]fix.S17, 4096)
	}

	recv := func(out, in []byte, framecount uint32) {
		if framecount == 0 {
			return
		}
		// de-interleave and reformat the samples. Each input sample is
		// 4 bytes.
		frameSize := 4 * inps
		for i := 0; i < len(in); i += frameSize {
			for c := range inputs {
				// Convert from float 32 to unsigned 8
				j := i + c*4
				u := binary.LittleEndian.Uint32(in[j:])
				f := math.Float32frombits(u)
				inputs[c][i/frameSize] = fix.FromFloat(f)
			}
		}
		for i, inp := range inputs {
			// Make sure the bounds are correct.
			inputs[i] = inp[:framecount]
		}
		for i, outp := range outputs {
			outputs[i] = outp[:framecount]
		}
		// Run the ticker.
		t.Tick(inputs, outputs)

		// reformat the output to float32 and re-interleave.
		o := out[:0]
		for i := 0; i < int(framecount); i++ {
			for c := range outputs {
				f := fix.Float[float32](outputs[c][i])
				o = binary.LittleEndian.AppendUint32(o, math.Float32bits(f))
			}
		}
	}

	device, err := malgo.InitDevice(mctx.Context, cfg, malgo.DeviceCallbacks{
		Data: recv,
	})
	if err != nil {
		return err
	}
	if err := device.Start(); err != nil {
		return err
	}

	<-ctx.Done()

	device.Uninit()
	return nil
}
