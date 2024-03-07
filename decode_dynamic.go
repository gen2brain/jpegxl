package jpegxl

import (
	"image"
	"image/color"
	"io"
	"runtime"

	"github.com/ebitengine/purego"
)

func decodeDynamic(r io.Reader, configOnly, decodeAll bool) (*JXL, image.Config, error) {
	var cfg image.Config

	decoder := jxlDecoderCreate()
	defer jxlDecoderDestroy(decoder)

	if !jxlDecoderSubscribeEvents(decoder, jxlDecBasicInfo|jxlDecFrame|jxlDecFullImage) {
		return nil, cfg, ErrDecode
	}

	var info jxlBasicInfo
	var header jxlFrameHeader

	var format jxlPixelFormat
	format.NumChannels = 4
	format.DataType = jxlTypeUint8
	format.Endianness = jxlNativeEndian

	data, err := io.ReadAll(r)
	if err != nil {
		return nil, cfg, err
	}

	jxlDecoderSetInput(decoder, data)
	jxlDecoderCloseInput(decoder)

	delay := make([]int, 0)
	images := make([]image.Image, 0)

	for {
		status := jxlDecoderProcessInput(decoder)

		switch status {
		case jxlDecError:
			return nil, cfg, ErrDecode
		case jxlDecNeedMoreInput:
			return nil, cfg, ErrDecode
		case jxlDecBasicInfo:
			if !jxlDecoderGetBasicInfo(decoder, &info) {
				return nil, cfg, ErrDecode
			}

			cfg.Width = int(info.Xsize)
			cfg.Height = int(info.Ysize)
			cfg.ColorModel = color.NRGBAModel

			if configOnly && info.HaveAnimation == 0 {
				return nil, cfg, nil
			}

			if info.BitsPerSample == 16 {
				format.DataType = jxlTypeUint16
				format.Endianness = jxlBigEndian
			}
		case jxlDecFrame:
			if !jxlDecoderGetFrameHeader(decoder, &header) {
				return nil, cfg, ErrDecode
			}

			delay = append(delay, int(header.Duration))
		case jxlDecNeedImageOutBuffer:
			if configOnly {
				jxlDecoderSkipCurrentFrame(decoder)

				continue
			}

			var bufSize uint64
			if !jxlDecoderImageOutBufferSize(decoder, &format, &bufSize) {
				return nil, cfg, ErrDecode
			}

			if info.BitsPerSample == 16 {
				img := image.NewNRGBA64(image.Rect(0, 0, cfg.Width, cfg.Height))
				images = append(images, img)

				if !jxlDecoderSetImageOutBuffer(decoder, &format, img.Pix, bufSize) {
					return nil, cfg, ErrDecode
				}
			} else {
				img := image.NewNRGBA(image.Rect(0, 0, cfg.Width, cfg.Height))
				images = append(images, img)

				if !jxlDecoderSetImageOutBuffer(decoder, &format, img.Pix, bufSize) {
					return nil, cfg, ErrDecode
				}
			}
		case jxlDecFullImage:
			if !decodeAll || (info.HaveAnimation == 1 && header.IsLast == 1) {
				ret := &JXL{
					Image: images,
					Delay: delay,
				}

				return ret, cfg, nil
			}
		case jxlDecSuccess:
			runtime.KeepAlive(data)

			ret := &JXL{
				Image: images,
				Delay: delay,
			}

			return ret, cfg, nil
		}
	}
}

func init() {
	var err error

	libjxl, err = loadLibrary()
	if err == nil {
		dynamic = true
	} else {
		return
	}

	purego.RegisterLibFunc(&_jxlDecoderCreate, libjxl, "JxlDecoderCreate")
	purego.RegisterLibFunc(&_jxlDecoderDestroy, libjxl, "JxlDecoderDestroy")
	purego.RegisterLibFunc(&_jxlDecoderSubscribeEvents, libjxl, "JxlDecoderSubscribeEvents")
	purego.RegisterLibFunc(&_jxlDecoderSetInput, libjxl, "JxlDecoderSetInput")
	purego.RegisterLibFunc(&_jxlDecoderCloseInput, libjxl, "JxlDecoderCloseInput")
	purego.RegisterLibFunc(&_jxlDecoderProcessInput, libjxl, "JxlDecoderProcessInput")
	purego.RegisterLibFunc(&_jxlDecoderGetBasicInfo, libjxl, "JxlDecoderGetBasicInfo")
	purego.RegisterLibFunc(&_jxlDecoderGetFrameHeader, libjxl, "JxlDecoderGetFrameHeader")
	purego.RegisterLibFunc(&_jxlDecoderSkipCurrentFrame, libjxl, "JxlDecoderSkipCurrentFrame")
	purego.RegisterLibFunc(&_jxlDecoderImageOutBufferSize, libjxl, "JxlDecoderImageOutBufferSize")
	purego.RegisterLibFunc(&_jxlDecoderSetImageOutBuffer, libjxl, "JxlDecoderSetImageOutBuffer")
}

var (
	libjxl  uintptr
	dynamic bool
)

const (
	jxlDecSuccess            = 0
	jxlDecError              = 1
	jxlDecNeedMoreInput      = 2
	jxlDecNeedImageOutBuffer = 5
	jxlDecBasicInfo          = 0x40
	jxlDecFrame              = 0x400
	jxlDecFullImage          = 0x1000

	jxlTypeUint8  = 2
	jxlTypeUint16 = 3

	jxlNativeEndian = 0
	jxlBigEndian    = 2
)

var (
	_jxlDecoderCreate             func(uintptr) *jxlDecoder
	_jxlDecoderDestroy            func(*jxlDecoder)
	_jxlDecoderSubscribeEvents    func(*jxlDecoder, int32) int
	_jxlDecoderSetInput           func(*jxlDecoder, []byte, uint64) int
	_jxlDecoderCloseInput         func(*jxlDecoder)
	_jxlDecoderProcessInput       func(*jxlDecoder) int
	_jxlDecoderGetBasicInfo       func(*jxlDecoder, *jxlBasicInfo) int
	_jxlDecoderGetFrameHeader     func(*jxlDecoder, *jxlFrameHeader) int
	_jxlDecoderSkipCurrentFrame   func(*jxlDecoder)
	_jxlDecoderImageOutBufferSize func(*jxlDecoder, *jxlPixelFormat, *uint64) int
	_jxlDecoderSetImageOutBuffer  func(*jxlDecoder, *jxlPixelFormat, []byte, uint64) int
)

func jxlDecoderCreate() *jxlDecoder {
	return _jxlDecoderCreate(0)
}

func jxlDecoderDestroy(decoder *jxlDecoder) {
	_jxlDecoderDestroy(decoder)
}

func jxlDecoderSubscribeEvents(decoder *jxlDecoder, wanted int) bool {
	ret := _jxlDecoderSubscribeEvents(decoder, int32(wanted))

	return ret == 0
}

func jxlDecoderSetInput(decoder *jxlDecoder, data []byte) bool {
	ret := _jxlDecoderSetInput(decoder, data, uint64(len(data)))

	return ret == 0
}

func jxlDecoderCloseInput(decoder *jxlDecoder) {
	_jxlDecoderCloseInput(decoder)
}

func jxlDecoderProcessInput(decoder *jxlDecoder) int {
	ret := _jxlDecoderProcessInput(decoder)

	return ret
}

func jxlDecoderGetBasicInfo(decoder *jxlDecoder, info *jxlBasicInfo) bool {
	ret := _jxlDecoderGetBasicInfo(decoder, info)

	return ret == 0
}

func jxlDecoderGetFrameHeader(decoder *jxlDecoder, header *jxlFrameHeader) bool {
	ret := _jxlDecoderGetFrameHeader(decoder, header)

	return ret == 0
}

func jxlDecoderSkipCurrentFrame(decoder *jxlDecoder) {
	_jxlDecoderSkipCurrentFrame(decoder)
}

func jxlDecoderImageOutBufferSize(decoder *jxlDecoder, format *jxlPixelFormat, size *uint64) bool {
	ret := _jxlDecoderImageOutBufferSize(decoder, format, size)

	return ret == 0
}

func jxlDecoderSetImageOutBuffer(decoder *jxlDecoder, format *jxlPixelFormat, buffer []byte, size uint64) bool {
	ret := _jxlDecoderSetImageOutBuffer(decoder, format, buffer, size)

	return ret == 0
}

type jxlBasicInfo struct {
	HaveContainer         int32
	Xsize                 uint32
	Ysize                 uint32
	BitsPerSample         uint32
	ExponentBitsPerSample uint32
	IntensityTarget       float32
	MinNits               float32
	RelativeToMaxDisplay  int32
	LinearBelow           float32
	UsesOriginalProfile   int32
	HavePreview           int32
	HaveAnimation         int32
	Orientation           uint32
	NumColorChannels      uint32
	NumExtraChannels      uint32
	AlphaBits             uint32
	AlphaExponentBits     uint32
	AlphaPremultiplied    int32
	Preview               jxlPreviewHeader
	Animation             jxlAnimationHeader
	IntrinsicXsize        uint32
	IntrinsicYsize        uint32
	Padding               [100]uint8
}

type jxlFrameHeader struct {
	Duration   uint32
	Timecode   uint32
	NameLength uint32
	IsLast     int32
	LayerInfo  jxlLayerInfo
}

type jxlPixelFormat struct {
	NumChannels uint32
	DataType    uint32
	Endianness  uint32
	Align       uint64
}

type jxlAnimationHeader struct {
	TpsNumerator   uint32
	TpsDenominator uint32
	NumLoops       uint32
	HaveTimecodes  int32
}

type jxlPreviewHeader struct {
	Xsize uint32
	Ysize uint32
}

type jxlLayerInfo struct {
	HaveCrop        int32
	CropX0          int32
	CropY0          int32
	Xsize           uint32
	Ysize           uint32
	BlendInfo       jxlBlendInfo
	SaveAsReference uint32
}

type jxlBlendInfo struct {
	Blendmode uint32
	Source    uint32
	Alpha     uint32
	Clamp     int32
}

type jxlDecoder struct{}
