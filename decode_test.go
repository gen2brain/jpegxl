package jpegxl_test

import (
	"bytes"
	_ "embed"
	"image"
	"image/jpeg"
	"io"
	"testing"

	"github.com/gen2brain/jpegxl"
)

//go:embed testdata/test8.jxl
var testJxl8 []byte

//go:embed testdata/test16.jxl
var testJxl16 []byte

func TestDecode(t *testing.T) {
	img, err := jpegxl.Decode(bytes.NewReader(testJxl8))
	if err != nil {
		t.Fatal(err)
	}

	err = jpeg.Encode(io.Discard, img, nil)
	if err != nil {
		t.Error(err)
	}
}

func TestDecode16(t *testing.T) {
	img, err := jpegxl.Decode(bytes.NewReader(testJxl16))
	if err != nil {
		t.Fatal(err)
	}

	err = jpeg.Encode(io.Discard, img, nil)
	if err != nil {
		t.Error(err)
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

func TestDecodeConfig(t *testing.T) {
	cfg, err := jpegxl.DecodeConfig(bytes.NewReader(testJxl8))
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

func BenchmarkDecodeJPEG(b *testing.B) {
	img, _, err := image.Decode(bytes.NewReader(testJxl8))
	if err != nil {
		b.Error(err)
	}

	var testJpeg bytes.Buffer
	err = jpeg.Encode(&testJpeg, img, nil)
	if err != nil {
		b.Error(err)
	}

	for i := 0; i < b.N; i++ {
		_, _, err := image.Decode(bytes.NewReader(testJpeg.Bytes()))
		if err != nil {
			b.Error(err)
		}
	}
}

func BenchmarkDecodeJPEGXL(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _, err := image.Decode(bytes.NewReader(testJxl8))
		if err != nil {
			b.Error(err)
		}
	}
}

func BenchmarkDecodeConfig(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := jpegxl.DecodeConfig(bytes.NewReader(testJxl8))
		if err != nil {
			b.Error(err)
		}
	}
}
