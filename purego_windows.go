//go:build windows && !nodynamic

package jpegxl

import (
	"fmt"

	"golang.org/x/sys/windows"
)

const (
	libname = "libjxl.dll"
)

func loadLibrary() (uintptr, error) {
	handle, err := windows.LoadLibrary(libname)
	if err != nil {
		return 0, fmt.Errorf("cannot load library %s: %w", libname, err)
	}

	return uintptr(handle), nil
}
