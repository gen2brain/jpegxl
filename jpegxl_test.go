package jpegxl

import (
	"bytes"
	_ "embed"
	"image"
	"image/jpeg"
	"io"
	"testing"
)

//go:embed testdata/test8.jxl
var testJxl8 []byte

//go:embed testdata/test16.jxl
var testJxl16 []byte

//go:embed testdata/test.jxl
var testJxlAnim []byte

func TestDecode(t *testing.T) {
	img, err := Decode(bytes.NewReader(testJxl8))
	if err != nil {
		t.Fatal(err)
	}

	err = jpeg.Encode(io.Discard, img, nil)
	if err != nil {
		t.Error(err)
	}
}

func TestDecode16(t *testing.T) {
	img, err := Decode(bytes.NewReader(testJxl16))
	if err != nil {
		t.Fatal(err)
	}

	err = jpeg.Encode(io.Discard, img, nil)
	if err != nil {
		t.Error(err)
	}
}

func TestDecodeAnim(t *testing.T) {
	ret, err := DecodeAll(bytes.NewReader(testJxlAnim))
	if err != nil {
		t.Fatal(err)
	}

	if len(ret.Image) != len(ret.Delay) {
		t.Errorf("not equal, got %d, want %d", len(ret.Delay), len(ret.Image))
	}

	if len(ret.Image) != 48 {
		t.Errorf("got %d, want %d", len(ret.Image), 48)
	}

	for _, img := range ret.Image {
		err = jpeg.Encode(io.Discard, img, nil)
		if err != nil {
			t.Error(err)
		}
	}
}

func TestImageDecode(t *testing.T) {
	img, _, err := image.Decode(bytes.NewReader(testJxl8))
	if err != nil {
		t.Fatal(err)
	}

	err = jpeg.Encode(io.Discard, img, nil)
	if err != nil {
		t.Error(err)
	}
}

func TestImageDecodeAnim(t *testing.T) {
	img, _, err := image.Decode(bytes.NewReader(testJxlAnim))
	if err != nil {
		t.Fatal(err)
	}

	err = jpeg.Encode(io.Discard, img, nil)
	if err != nil {
		t.Error(err)
	}
}

func TestDecodeConfig(t *testing.T) {
	cfg, err := DecodeConfig(bytes.NewReader(testJxl8))
	if err != nil {
		t.Fatal(err)
	}

	if cfg.Width != 512 {
		t.Errorf("width: got %d, want %d", cfg.Width, 512)
	}

	if cfg.Height != 512 {
		t.Errorf("height: got %d, want %d", cfg.Height, 512)
	}
}

func TestEncode(t *testing.T) {
	img, err := Decode(bytes.NewReader(testJxl8))
	if err != nil {
		t.Fatal(err)
	}

	err = Encode(io.Discard, img)
	if err != nil {
		t.Fatal(err)
	}
}

func BenchmarkDecodeJPEGXL(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _, err := decode(bytes.NewReader(testJxl8), false, false)
		if err != nil {
			b.Error(err)
		}
	}
}

func BenchmarkDecodeJPEGXLDynamic(b *testing.B) {
	if Dynamic() != nil {
		b.Errorf("dynamic/shared library not installed")
		return
	}

	for i := 0; i < b.N; i++ {
		_, _, err := decodeDynamic(bytes.NewReader(testJxl8), false, false)
		if err != nil {
			b.Error(err)
		}
	}
}

func BenchmarkDecodeConfigJPEGXL(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _, err := decode(bytes.NewReader(testJxl8), true, false)
		if err != nil {
			b.Error(err)
		}
	}
}

func BenchmarkDecodeConfigJPEGXLDynamic(b *testing.B) {
	if Dynamic() != nil {
		b.Errorf("dynamic/shared library not installed")
		return
	}

	for i := 0; i < b.N; i++ {
		_, _, err := decodeDynamic(bytes.NewReader(testJxl8), true, false)
		if err != nil {
			b.Error(err)
		}
	}
}

func BenchmarkEncodeJPEGXL(b *testing.B) {
	img, err := Decode(bytes.NewReader(testJxl8))
	if err != nil {
		b.Fatal(err)
	}

	for i := 0; i < b.N; i++ {
		err := encode(io.Discard, img, DefaultQuality, DefaultEffort)
		if err != nil {
			b.Error(err)
		}
	}
}

func BenchmarkEncodeJPEGXLDynamic(b *testing.B) {
	img, err := Decode(bytes.NewReader(testJxl8))
	if err != nil {
		b.Fatal(err)
	}

	for i := 0; i < b.N; i++ {
		err := encodeDynamic(io.Discard, img, DefaultQuality, DefaultEffort)
		if err != nil {
			b.Error(err)
		}
	}
}
