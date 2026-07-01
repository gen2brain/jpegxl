package jpegxl

import (
	"bytes"
	_ "embed"
	"testing"
)

//go:embed testdata/test_exif.jxl
var testJxlExif []byte

// TestDecodeExif covers the brotli-compressed brob path.
func TestDecodeExif(t *testing.T) {
	ex, err := DecodeExif(bytes.NewReader(testJxlExif))
	if err != nil {
		t.Fatal(err)
	}

	if ex.Orientation != 6 {
		t.Errorf("Orientation = %d, want 6", ex.Orientation)
	}

	if ex.Make != "TestCam" {
		t.Errorf("Make = %q, want %q", ex.Make, "TestCam")
	}

	if ex.Model != "Model123" {
		t.Errorf("Model = %q, want %q", ex.Model, "Model123")
	}

	if ex.ISOSpeed != 800 {
		t.Errorf("ISOSpeed = %d, want 800", ex.ISOSpeed)
	}
}

// TestDecodeExifPlain covers the uncompressed Exif box path.
func TestDecodeExifPlain(t *testing.T) {
	if _, err := DecodeExif(bytes.NewReader(testJxl8)); err != nil {
		t.Fatal(err)
	}
}

// TestDecodeExifNone covers a raw codestream with no container boxes.
func TestDecodeExifNone(t *testing.T) {
	if _, err := DecodeExif(bytes.NewReader(testJxl16)); err != ErrNoExif {
		t.Errorf("got %v, want ErrNoExif", err)
	}
}
