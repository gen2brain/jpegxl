// Package jpegxl implements an JPEG XL image decoder based on libjxl compiled to WASM.
package jpegxl

import (
	"image"
)

func init() {
	image.RegisterFormat("jxl", "????JXL", Decode, DecodeConfig)
	image.RegisterFormat("jxl", "\xff\x0a", Decode, DecodeConfig)
}
