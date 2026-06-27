//go:build wasm2go

package jpegxl

import (
	"fmt"
	"image"
	"image/color"
	"io"
	"sync"
)

var modPool = sync.Pool{New: func() any { return newModuleRaw() }}

// There is no runtime to set up; modules are pooled per call.
func initDecoderOnce() {}
func initEncoderOnce() {}

func decode(r io.Reader, configOnly, decodeAll bool) (*JXL, image.Config, error) {
	var cfg image.Config

	data, err := io.ReadAll(r)
	if err != nil {
		return nil, cfg, fmt.Errorf("read: %w", err)
	}

	mod := modPool.Get().(*module)
	defer modPool.Put(mod)

	inPtr := mod.Xmalloc(int32(len(data)))
	if inPtr == 0 {
		return nil, cfg, ErrMemWrite
	}
	defer mod.Xfree(inPtr)
	if !mod.write(inPtr, data) {
		return nil, cfg, ErrMemWrite
	}

	info := mod.Xmalloc(16)
	if info == 0 {
		return nil, cfg, ErrMemWrite
	}
	defer mod.Xfree(info)

	cfgOnly := int32(0)
	if configOnly {
		cfgOnly = 1
	}

	out := mod.Xdecode(inPtr, int32(len(data)), cfgOnly, info)

	width := int(load32(mod.memory[info:]))
	height := int(load32(mod.memory[info+4:]))
	count := int(load32(mod.memory[info+12:]))

	cfg.Width = width
	cfg.Height = height
	cfg.ColorModel = color.NRGBAModel

	if configOnly {
		return nil, cfg, nil
	}
	if out == 0 {
		return nil, cfg, ErrDecode
	}
	defer mod.Xfree(out)

	size := width * height * 4
	images := make([]image.Image, 0, count)
	delay := make([]int, 0, count)

	for i := 0; i < count; i++ {
		src, ok := mod.read(out+int32(i*size), int32(size))
		if !ok {
			return nil, cfg, ErrMemRead
		}

		img := image.NewNRGBA(image.Rect(0, 0, width, height))
		img.Pix = make([]byte, size)
		copy(img.Pix, src)

		images = append(images, img)
		delay = append(delay, 0)

		if !decodeAll {
			break
		}
	}

	return &JXL{Image: images, Delay: delay}, cfg, nil
}

// encode always produces lossless JXL; zune-jpegxl has no quality knob.
func encode(w io.Writer, m image.Image, quality, effort int, lossless bool) error {
	img := imageToNRGBA(m)

	mod := modPool.Get().(*module)
	defer modPool.Put(mod)

	inPtr := mod.Xmalloc(int32(len(img.Pix)))
	if inPtr == 0 {
		return ErrMemWrite
	}
	defer mod.Xfree(inPtr)
	if !mod.write(inPtr, img.Pix) {
		return ErrMemWrite
	}

	sizePtr := mod.Xmalloc(8)
	if sizePtr == 0 {
		return ErrMemWrite
	}
	defer mod.Xfree(sizePtr)

	out := mod.Xencode(inPtr, int32(img.Bounds().Dx()), int32(img.Bounds().Dy()), sizePtr, 0, 0)
	if out == 0 {
		return ErrEncode
	}
	defer mod.Xfree(out)

	size := int(load64(mod.memory[sizePtr:]))
	src, ok := mod.read(out, int32(size))
	if !ok {
		return ErrMemRead
	}

	if _, err := w.Write(src); err != nil {
		return fmt.Errorf("write: %w", err)
	}

	return nil
}

func (m *module) write(ptr int32, data []byte) bool {
	if ptr < 0 || int(ptr)+len(data) > len(m.memory) {
		return false
	}
	copy(m.memory[ptr:], data)
	return true
}

func (m *module) read(ptr, size int32) ([]byte, bool) {
	if ptr < 0 || size < 0 || int(ptr)+int(size) > len(m.memory) {
		return nil, false
	}
	return m.memory[ptr : ptr+size : ptr+size], true
}
