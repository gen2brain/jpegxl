// Package jpegxl implements an JPEG XL image decoder based on libjxl compiled to WASM.
package jpegxl

import (
	"errors"
	"image"
	"io"
)

// JXL represents the possibly multiple images stored in a JXL file.
type JXL struct {
	// Decoded images, NRGBA or NRGBA64.
	Image []image.Image
	// Delay times, one per frame, in seconds of a tick.
	Delay []int
}

// Errors .
var (
	ErrMemRead  = errors.New("jpegxl: mem read failed")
	ErrMemWrite = errors.New("jpegxl: mem write failed")
	ErrDecode   = errors.New("jpegxl: decode failed")
)

// Decode reads a JPEG XL image from r and returns it as an image.Image.
func Decode(r io.Reader) (image.Image, error) {
	var err error
	var ret *JXL

	if dynamic {
		ret, _, err = decodeDynamic(r, false, false)
		if err != nil {
			return nil, err
		}
	} else {
		ret, _, err = decode(r, false, false)
		if err != nil {
			return nil, err
		}
	}

	return ret.Image[0], nil
}

// DecodeConfig returns the color model and dimensions of a JPEG XL image without decoding the entire image.
func DecodeConfig(r io.Reader) (image.Config, error) {
	var err error
	var cfg image.Config

	if dynamic {
		_, cfg, err = decodeDynamic(r, true, false)
		if err != nil {
			return image.Config{}, err
		}
	} else {
		_, cfg, err = decode(r, true, false)
		if err != nil {
			return image.Config{}, err
		}
	}

	return cfg, nil
}

// DecodeAll reads a JPEG XL image from r and returns the sequential frames and timing information.
func DecodeAll(r io.Reader) (*JXL, error) {
	var err error
	var ret *JXL

	if dynamic {
		ret, _, err = decodeDynamic(r, false, true)
		if err != nil {
			return nil, err
		}
	} else {
		ret, _, err = decode(r, false, true)
		if err != nil {
			return nil, err
		}
	}

	return ret, nil
}

// Dynamic returns true when library is using the dynamic/shared library.
func Dynamic() bool {
	return dynamic
}

func init() {
	image.RegisterFormat("jxl", "????JXL", Decode, DecodeConfig)
	image.RegisterFormat("jxl", "\xff\x0a", Decode, DecodeConfig)
}
