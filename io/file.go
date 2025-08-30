package io

import (
	"os"
	"sync"

	"github.com/pfcm/audiofile/wav"
	"github.com/pfcm/fxp/fix"
)

// WavWriter is an fxp.Ticker that writes its input to a wav file. It
// has an input for each channel, and the actual writes and sample
// conversion is done in a separate thread.
// Close must always be called when there is nothing left to write,
// to make sure the file is properly finalised.
// Note Tick may panic if it is called after Close.
type WavWriter struct {
	fmt wav.FileFormat
	w   *wav.Writer

	c    chan [][]fix.S17
	errs chan error
}

func NewWavWriter(fmt wav.FileFormat, path string) (*WavWriter, error) {
	f, err := os.Create(path)
	if err != nil {
		return nil, err
	}
	w, err := wav.NewWriter(f, fmt)
	if err != nil {
		return nil, err
	}

	return &WavWriter{
		fmt:  fmt,
		w:    w,
		c:    make(chan [][]fix.S17, 256),
		errs: make(chan error, 1),
	}, nil
}

func (w *WavWriter) Inputs() int    { return w.fmt.Channels }
func (w *WavWriter) Outputs() int   { return 0 }
func (w *WavWriter) String() string { return "WavWriter" }

func (w *WavWriter) Tick(inputs, _ [][]fix.S17) {
	select {
	case err := <-w.errs:
		panic(err)
	default:
	}
	buf := get(w.fmt.Channels, len(inputs[0]))
	for i := range buf {
		copy(buf[i], inputs[i])
	}
	w.c <- buf
}

// writeLoop writes everything sent over w.c to the file. Any errors
// will break the loop and send the error on w.errs. When w.c is
// closed, the file is finalised and a nil is sent on w.errs.
func (w *WavWriter) writeLoop() {
	tmp := make([][]int16, w.fmt.Channels)
	for buf := range w.c {
		// TODO: convert into the actual format of the file.
		for i := range buf {
			if l := len(buf[i]); cap(tmp[i]) < l {
				tmp[i] = make([]int16, l)
			}
			tmp[i] = tmp[i][:len(buf[i])]
			for j := range buf[i] {
				tmp[i][j] = int16(buf[i][j]) << 8
			}
		}
		if _, err := w.w.Write16PCM(tmp); err != nil {
			w.errs <- err
			close(w.c)
			free(buf)
			return
		}
		free(buf)
	}
	w.errs <- w.Close()
}

// Close finishes any pending writing and closes the file.
func (w *WavWriter) Close() error {
	close(w.c)
	return <-w.errs
}

var pool = sync.Pool{
	New: func() any {
		b := make([]fix.S17, 1024)
		return &b
	},
}

func get(channels, size int) [][]fix.S17 {
	buf := make([][]fix.S17, channels)
	for i := range buf {
		b := *(pool.Get().(*[]fix.S17))
		if cap(b) < size {
			b = make([]fix.S17, 1024)
		}
		b = b[:size]
		buf[i] = b
	}
	return buf
}

func free(buf [][]fix.S17) {
	for i := range buf {
		pool.Put(&buf[i])
	}
}
