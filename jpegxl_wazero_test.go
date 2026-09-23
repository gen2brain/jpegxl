//go:build !wasm2go

package jpegxl

import (
	"bytes"
	_ "embed"
	"image"
	"testing"
)

// test16_linear.jxl is a lossless 32x32 16-bit image tagged linear Rec.2100 (as DxO PureRAW's DNG tiles are), whose
// samples follow linear16Sample exactly.
//
//go:embed testdata/test16_linear.jxl
var testJxl16Linear []byte

func linear16Sample(x, y int) [3]uint16 {
	return [3]uint16{uint16(x*512 + y), uint16(40000 - x*300 - y*7), uint16(y*1000 + x)}
}

// Decoding this color encoding makes libjxl allocate in a way that left the freed output buffer's first pixels
// overwritten by allocator metadata when the result aliased the module's memory.
func TestDecode16LinearPixels(t *testing.T) {
	ret, _, err := decode(bytes.NewReader(testJxl16Linear), false, false)
	if err != nil {
		t.Fatal(err)
	}

	img, ok := ret.Image[0].(*image.NRGBA64)
	if !ok {
		t.Fatalf("decoded %T, want *image.NRGBA64", ret.Image[0])
	}

	for y := range 32 {
		for x := range 32 {
			c := img.NRGBA64At(x, y)
			if got, want := [3]uint16{c.R, c.G, c.B}, linear16Sample(x, y); got != want {
				t.Fatalf("pixel (%d,%d) = %v, want %v", x, y, got, want)
			}
		}
	}
}
