package jpegxl

import (
	"bytes"
	"compress/gzip"
	"context"
	_ "embed"
	"fmt"
	"image"
	"image/color"
	"io"
	"os"
	"sync/atomic"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

//go:embed lib/jxl.wasm.gz
var jxlWasm []byte

func decode(r io.Reader, configOnly, decodeAll bool) (*JXL, image.Config, error) {
	if !initialized.Load() {
		initialize()
	}

	var cfg image.Config
	var jxl bytes.Buffer

	_, err := jxl.ReadFrom(r)
	if err != nil {
		return nil, cfg, fmt.Errorf("read: %w", err)
	}

	inSize := jxl.Len()
	ctx := context.Background()

	res, err := _alloc.Call(ctx, uint64(inSize))
	if err != nil {
		return nil, cfg, fmt.Errorf("alloc: %w", err)
	}
	inPtr := res[0]
	defer _free.Call(ctx, inPtr)

	ok := mod.Memory().Write(uint32(inPtr), jxl.Bytes())
	if !ok {
		return nil, cfg, ErrMemWrite
	}

	res, err = _alloc.Call(ctx, 4)
	if err != nil {
		return nil, cfg, fmt.Errorf("alloc: %w", err)
	}
	widthPtr := res[0]
	defer _free.Call(ctx, widthPtr)

	res, err = _alloc.Call(ctx, 4)
	if err != nil {
		return nil, cfg, fmt.Errorf("alloc: %w", err)
	}
	heightPtr := res[0]
	defer _free.Call(ctx, heightPtr)

	res, err = _alloc.Call(ctx, 4)
	if err != nil {
		return nil, cfg, fmt.Errorf("alloc: %w", err)
	}
	depthPtr := res[0]
	defer _free.Call(ctx, depthPtr)

	res, err = _alloc.Call(ctx, 4)
	if err != nil {
		return nil, cfg, fmt.Errorf("alloc: %w", err)
	}
	countPtr := res[0]
	defer _free.Call(ctx, countPtr)

	res, err = _decode.Call(ctx, inPtr, uint64(inSize), 1, 0, widthPtr, heightPtr, depthPtr, countPtr, 0, 0)
	if err != nil {
		return nil, cfg, fmt.Errorf("decode: %w", err)
	}

	if res[0] == 0 {
		return nil, cfg, ErrDecode
	}

	width, ok := mod.Memory().ReadUint32Le(uint32(widthPtr))
	if !ok {
		return nil, cfg, ErrMemRead
	}

	height, ok := mod.Memory().ReadUint32Le(uint32(heightPtr))
	if !ok {
		return nil, cfg, ErrMemRead
	}

	depth, ok := mod.Memory().ReadUint32Le(uint32(depthPtr))
	if !ok {
		return nil, cfg, ErrMemRead
	}

	count, ok := mod.Memory().ReadUint32Le(uint32(countPtr))
	if !ok {
		return nil, cfg, ErrMemRead
	}

	cfg.Width = int(width)
	cfg.Height = int(height)

	cfg.ColorModel = color.NRGBAModel
	if depth == 16 {
		cfg.ColorModel = color.NRGBA64Model
	}

	if configOnly {
		return nil, cfg, nil
	}

	size := cfg.Width * cfg.Height * 4
	if depth == 16 {
		size = cfg.Width * cfg.Height * 8
	}

	outSize := size
	if decodeAll {
		outSize = size * int(count)
	}

	res, err = _alloc.Call(ctx, uint64(outSize))
	if err != nil {
		return nil, cfg, fmt.Errorf("alloc: %w", err)
	}
	outPtr := res[0]
	defer _free.Call(ctx, outPtr)

	delaySize := 4
	if decodeAll {
		delaySize = 4 * int(count)
	}

	res, err = _alloc.Call(ctx, uint64(delaySize))
	if err != nil {
		return nil, cfg, fmt.Errorf("alloc: %w", err)
	}
	delayPtr := res[0]
	defer _free.Call(ctx, delayPtr)

	all := 0
	if decodeAll {
		all = 1
	}

	res, err = _decode.Call(ctx, inPtr, uint64(inSize), 0, uint64(all), widthPtr, heightPtr, depthPtr, countPtr, delayPtr, outPtr)
	if err != nil {
		return nil, cfg, fmt.Errorf("decode: %w", err)
	}

	if res[0] == 0 {
		return nil, cfg, ErrDecode
	}

	delay := make([]int, 0)
	images := make([]image.Image, 0)

	for i := 0; i < int(count); i++ {
		out, ok := mod.Memory().Read(uint32(outPtr)+uint32(i*size), uint32(size))
		if !ok {
			return nil, cfg, ErrMemRead
		}

		if depth == 16 {
			img := image.NewNRGBA64(image.Rect(0, 0, cfg.Width, cfg.Height))
			img.Pix = out
			images = append(images, img)
		} else {
			img := image.NewNRGBA(image.Rect(0, 0, cfg.Width, cfg.Height))
			img.Pix = out
			images = append(images, img)
		}

		d, ok := mod.Memory().ReadUint32Le(uint32(delayPtr) + uint32(i*4))
		if !ok {
			return nil, cfg, ErrMemRead
		}

		delay = append(delay, int(d))

		if !decodeAll {
			break
		}
	}

	ret := &JXL{
		Image: images,
		Delay: delay,
	}

	return ret, cfg, nil
}

var (
	mod api.Module

	_alloc  api.Function
	_free   api.Function
	_decode api.Function

	initialized atomic.Bool
)

func initialize() {
	if initialized.Load() {
		return
	}

	ctx := context.Background()
	rt := wazero.NewRuntime(ctx)

	r, err := gzip.NewReader(bytes.NewReader(jxlWasm))
	if err != nil {
		panic(err)
	}

	var data bytes.Buffer
	_, err = data.ReadFrom(r)
	if err != nil {
		panic(err)
	}

	compiled, err := rt.CompileModule(ctx, data.Bytes())
	if err != nil {
		panic(err)
	}

	wasi_snapshot_preview1.MustInstantiate(ctx, rt)

	mod, err = rt.InstantiateModule(ctx, compiled, wazero.NewModuleConfig().WithStderr(os.Stderr).WithStdout(os.Stdout))
	if err != nil {
		panic(err)
	}

	_alloc = mod.ExportedFunction("allocate")
	_free = mod.ExportedFunction("deallocate")
	_decode = mod.ExportedFunction("decode")

	initialized.Store(true)
}
