//go:build !unix && !darwin && !windows

package jpegxl

import (
	"fmt"
	"runtime"
)

func loadLibrary() (uintptr, error) {
	return 0, fmt.Errorf("unsupported os: %s", runtime.GOOS)
}
