package osc

import (
	"testing"

	"github.com/pfcm/fxp/fix"
)

func TestMakeStep(t *testing.T) {
	tab := &Table{samplerate: 44100, tab: make([]fix.S17, 128)}
	for _, f := range []float32{
		20,
		100,
		440,
		1024,
		40000,
	} {

		t.Error(tab.MakeStep(f))
	}
}
