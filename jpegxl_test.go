package jpegxl

import (
	"bytes"
	_ "embed"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"os"
	"sync"
	"testing"
)

//go:embed testdata/test8.jxl
var testJxl8 []byte

//go:embed testdata/test16.jxl
var testJxl16 []byte

//go:embed testdata/test.jxl
var testJxlAnim []byte

func TestDecode(t *testing.T) {
	img, _, err := decode(bytes.NewReader(testJxl8), false, false)
	if err != nil {
		t.Fatal(err)
	}

	w, err := writeCloser()
	if err != nil {
		t.Fatal(err)
	}
	defer w.Close()

	err = jpeg.Encode(w, img.Image[0], nil)
	if err != nil {
		t.Error(err)
	}
}

func TestDecode16(t *testing.T) {
	img, _, err := decode(bytes.NewReader(testJxl16), false, false)
	if err != nil {
		t.Fatal(err)
	}

	w, err := writeCloser()
	if err != nil {
		t.Fatal(err)
	}
	defer w.Close()

	err = jpeg.Encode(w, img.Image[0], nil)
	if err != nil {
		t.Error(err)
	}
}

func TestDecodeDynamic(t *testing.T) {
	if err := Dynamic(); err != nil {
		fmt.Println(err)
		t.Skip()
	}

	img, _, err := decodeDynamic(bytes.NewReader(testJxl8), false, false)
	if err != nil {
		t.Fatal(err)
	}

	w, err := writeCloser()
	if err != nil {
		t.Fatal(err)
	}
	defer w.Close()

	err = jpeg.Encode(w, img.Image[0], nil)
	if err != nil {
		t.Error(err)
	}
}

func TestDecode16Dynamic(t *testing.T) {
	if err := Dynamic(); err != nil {
		fmt.Println(err)
		t.Skip()
	}

	img, _, err := decodeDynamic(bytes.NewReader(testJxl16), false, false)
	if err != nil {
		t.Fatal(err)
	}

	w, err := writeCloser()
	if err != nil {
		t.Fatal(err)
	}
	defer w.Close()

	err = jpeg.Encode(w, img.Image[0], nil)
	if err != nil {
		t.Error(err)
	}
}

func TestDecodeAnim(t *testing.T) {
	ret, err := DecodeAll(bytes.NewReader(testJxlAnim))
	if err != nil {
		t.Fatal(err)
	}

	if len(ret.Image) != len(ret.Delay) {
		t.Errorf("not equal, got %d, want %d", len(ret.Delay), len(ret.Image))
	}

	if len(ret.Image) != 48 {
		t.Errorf("got %d, want %d", len(ret.Image), 48)
	}

	for _, img := range ret.Image {
		err = jpeg.Encode(io.Discard, img, nil)
		if err != nil {
			t.Error(err)
		}
	}
}

func TestImageDecode(t *testing.T) {
	img, _, err := image.Decode(bytes.NewReader(testJxl8))
	if err != nil {
		t.Fatal(err)
	}

	w, err := writeCloser()
	if err != nil {
		t.Fatal(err)
	}
	defer w.Close()

	err = jpeg.Encode(w, img, nil)
	if err != nil {
		t.Error(err)
	}
}

func TestImageDecodeAnim(t *testing.T) {
	img, _, err := image.Decode(bytes.NewReader(testJxlAnim))
	if err != nil {
		t.Fatal(err)
	}

	w, err := writeCloser()
	if err != nil {
		t.Fatal(err)
	}
	defer w.Close()

	err = jpeg.Encode(w, img, nil)
	if err != nil {
		t.Error(err)
	}
}

func TestDecodeConfig(t *testing.T) {
	cfg, err := DecodeConfig(bytes.NewReader(testJxl8))
	if err != nil {
		t.Fatal(err)
	}

	if cfg.Width != 512 {
		t.Errorf("width: got %d, want %d", cfg.Width, 512)
	}

	if cfg.Height != 512 {
		t.Errorf("height: got %d, want %d", cfg.Height, 512)
	}
}

func TestEncode(t *testing.T) {
	img, err := Decode(bytes.NewReader(testJxl8))
	if err != nil {
		t.Fatal(err)
	}

	w, err := writeCloser()
	if err != nil {
		t.Fatal(err)
	}
	defer w.Close()

	err = encode(w, img, DefaultQuality, DefaultEffort)
	if err != nil {
		t.Fatal(err)
	}
}

func TestEncodeDynamic(t *testing.T) {
	if err := Dynamic(); err != nil {
		fmt.Println(err)
		t.Skip()
	}

	img, err := Decode(bytes.NewReader(testJxl8))
	if err != nil {
		t.Fatal(err)
	}

	w, err := writeCloser()
	if err != nil {
		t.Fatal(err)
	}
	defer w.Close()

	err = encodeDynamic(w, img, DefaultQuality, DefaultEffort)
	if err != nil {
		t.Fatal(err)
	}
}

func TestEncodeSync(t *testing.T) {
	wg := sync.WaitGroup{}
	ch := make(chan bool, 2)

	img, err := Decode(bytes.NewReader(testJxl8))
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			ch <- true
			defer func() { <-ch; wg.Done() }()

			err = encode(io.Discard, img, DefaultQuality, DefaultEffort)
			if err != nil {
				t.Error(err)
			}
		}()
	}

	wg.Wait()
}

func BenchmarkDecode(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _, err := decode(bytes.NewReader(testJxl8), false, false)
		if err != nil {
			b.Error(err)
		}
	}
}

func BenchmarkDecodeDynamic(b *testing.B) {
	if err := Dynamic(); err != nil {
		fmt.Println(err)
		b.Skip()
	}

	for i := 0; i < b.N; i++ {
		_, _, err := decodeDynamic(bytes.NewReader(testJxl8), false, false)
		if err != nil {
			b.Error(err)
		}
	}
}

func BenchmarkDecodeConfig(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _, err := decode(bytes.NewReader(testJxl8), true, false)
		if err != nil {
			b.Error(err)
		}
	}
}

func BenchmarkDecodeConfigDynamic(b *testing.B) {
	if err := Dynamic(); err != nil {
		fmt.Println(err)
		b.Skip()
	}

	for i := 0; i < b.N; i++ {
		_, _, err := decodeDynamic(bytes.NewReader(testJxl8), true, false)
		if err != nil {
			b.Error(err)
		}
	}
}

func BenchmarkEncode(b *testing.B) {
	img, err := Decode(bytes.NewReader(testJxl8))
	if err != nil {
		b.Fatal(err)
	}

	for i := 0; i < b.N; i++ {
		err := encode(io.Discard, img, DefaultQuality, DefaultEffort)
		if err != nil {
			b.Error(err)
		}
	}
}

func BenchmarkEncodeDynamic(b *testing.B) {
	if err := Dynamic(); err != nil {
		fmt.Println(err)
		b.Skip()
	}

	img, err := Decode(bytes.NewReader(testJxl8))
	if err != nil {
		b.Fatal(err)
	}

	for i := 0; i < b.N; i++ {
		err := encodeDynamic(io.Discard, img, DefaultQuality, DefaultEffort)
		if err != nil {
			b.Error(err)
		}
	}
}

type discard struct{}

func (d discard) Close() error {
	return nil
}

func (discard) Write(p []byte) (int, error) {
	return len(p), nil
}

var discardCloser io.WriteCloser = discard{}

func writeCloser(s ...string) (io.WriteCloser, error) {
	if len(s) > 0 {
		f, err := os.Create(s[0])
		if err != nil {
			return nil, err
		}

		return f, nil
	}

	return discardCloser, nil
}
