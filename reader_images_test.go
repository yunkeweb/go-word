package word

import (
	"bytes"
	"testing"

	"github.com/yunkeweb/go-word/element"
	"github.com/yunkeweb/go-word/style"
)

func TestReaderDrawingImageRoundTrip(t *testing.T) {
	d := New()
	img := d.AddSection().AddImageBytes("source.png", []byte{0x89, 0x50, 0x4e, 0x47, 0x00}, style.Image{Width: 32, Height: 24, AltText: "logo"})
	raw, err := d.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	got, err := LoadBytes(raw)
	if err != nil {
		t.Fatal(err)
	}
	var loaded *element.Image
	walkDocument(got, func(el element.Element) {
		if v, ok := el.(*element.Image); ok {
			loaded = v
		}
	})
	if loaded == nil {
		t.Fatal("drawing image was not parsed")
	}
	if !bytes.Equal(loaded.Data, img.Data) {
		t.Fatalf("image data mismatch: %v", loaded.Data)
	}
	if loaded.GetAltText() != "logo" {
		t.Fatalf("alt text=%q", loaded.GetAltText())
	}
	if loaded.Style.WidthEMU == 0 || loaded.Style.HeightEMU == 0 || loaded.Style.Width != 32 || loaded.Style.Height != 24 {
		t.Fatalf("dimensions lost: %+v", loaded.Style)
	}
	second, err := got.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	for _, diag := range ValidatePackage(second) {
		if diag.Severity == "error" {
			t.Fatalf("round-trip package error: %+v", diag)
		}
	}
}
