package element

import (
	"path/filepath"
	"strings"

	"github.com/yunkeweb/go-word/style"
)

// Image is a picture (PHPWord Element\Image).
type Image struct {
	Base
	Source      string
	Data        []byte
	Style       style.Image
	IsWatermark bool
	Media       Media
}

func (i *Image) Type() string { return "Image" }

func (i *Image) GetSource() string { return i.Source }
func (i *Image) GetName() string {
	if i.Style.Name != "" {
		return i.Style.Name
	}
	return i.Source
}
func (i *Image) SetName(name string) { i.Style.Name = name }
func (i *Image) GetAltText() string  { return i.Style.AltText }
func (i *Image) SetAltText(s string) { i.Style.AltText = s }
func (i *Image) SetIsWatermark(v bool) {
	i.IsWatermark = v
	i.Style.IsWatermark = v
}

// NewImage constructs an image from a path.
func NewImage(source string, st any) *Image {
	img := &Image{Source: source}
	applyImageStyle(&img.Style, st)
	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(source), "."))
	if ext == "jpg" {
		ext = "jpeg"
	}
	img.Media = Media{
		MediaType: "image",
		Source:    source,
		Target:    source,
		Ext:       ext,
	}
	return img
}

// NewImageBytes constructs an image from memory.
func NewImageBytes(name string, data []byte, st any) *Image {
	img := NewImage(name, st)
	img.Data = data
	img.Media.Data = data
	return img
}

func applyImageStyle(dst *style.Image, st any) {
	switch v := st.(type) {
	case style.Image:
		*dst = v
	case *style.Image:
		if v != nil {
			*dst = *v
		}
	}
}
