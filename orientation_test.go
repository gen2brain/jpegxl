package jpegxl

import (
	"bytes"
	_ "embed"
	"testing"
)

//go:embed testdata/or6_ll.jxl
var testJxlOrient []byte

func TestOrientation(t *testing.T) {
	img, err := Decode(bytes.NewReader(testJxlOrient))
	if err != nil {
		t.Fatal(err)
	}

	b := img.Bounds()
	if b.Dx() != 480 || b.Dy() != 640 {
		t.Errorf("decoded dims: got %dx%d, want 480x640 (orientation applied)", b.Dx(), b.Dy())
	}
}

func TestConfigStream(t *testing.T) {
	cfg, err := DecodeConfig(bytes.NewReader(testJxlOrient))
	if err != nil {
		t.Fatal(err)
	}

	if cfg.Width != 480 || cfg.Height != 640 {
		t.Errorf("config dims: got %dx%d, want 480x640", cfg.Width, cfg.Height)
	}
}
