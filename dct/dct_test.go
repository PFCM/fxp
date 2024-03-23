package dct

import (
	"fmt"
	"math"
	"testing"

	"github.com/pfcm/fxp/fix"
)

func TestMatrix(t *testing.T) {
	got := Matrix(5)
	for row := 0; row < 5; row++ {
		for col := 0; col < 5; col++ {
			fmt.Print(got[row][col], " ")
		}
		fmt.Println()
	}
	t.Fatal("no")
}

func TestTransform(t *testing.T) {
	const n = 128
	m := Matrix(n)

	data := make([]fix.S17, n)
	for i := range data {
		f := math.Sin((math.Pi / float64(n/2)) * float64(i))
		data[i] = fix.FromFloat(f)
	}
	out := make([]fix.S17, n)

	Transform(data, out, m)

	t.Fatal(out)
}
