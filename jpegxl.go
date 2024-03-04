// Package jpegxl implements an JPEG XL image decoder based on libjxl compiled to WASM.
package jpegxl

import (
	"image"
)

// JXL represents the possibly multiple images stored in a JXL file.
type JXL struct {
	// Decoded images, NRGBA or NRGBA64.
	Image []image.Image
	// Delay times, one per frame, in seconds of a tick.
	Delay []int
}

func init() {
	image.RegisterFormat("jxl", "????JXL", Decode, DecodeConfig)
	image.RegisterFormat("jxl", "\xff\x0a", Decode, DecodeConfig)
}
