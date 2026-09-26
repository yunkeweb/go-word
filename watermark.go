package word

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/yunkeweb/go-word/element"
	"github.com/yunkeweb/go-word/style"
)

// WatermarkOptions configures a VML text watermark (PowerPlusWaterMarkObject).
type WatermarkOptions struct {
	Angle    float64 // degrees; 0 defaults to -45
	Color    string  // VML fillcolor, e.g. silver or C0C0C0
	FontSize int     // points; 0 uses Word's 1pt + fitshape
	FontName string  // default Calibri
	Opacity  float64 // 0–1; 0 defaults to 0.5. Values > 1 are treated as percent
	Tile     bool    // full-page grid
	Rows     int     // tile rows, default 3
	Cols     int     // tile columns, default 3
}

// ImageWatermarkOptions configures a VML picture watermark (WordPictureWatermark).
type ImageWatermarkOptions struct {
	Washout bool    // Word washout (gain/blacklevel)
	Scale   float64 // size multiplier; 0 leaves the default size
	Opacity float64 // 0–1; 0 omits v:fill opacity
}

// SetTextWatermark stores a document-wide text watermark applied to every
// section header as Word-native VML (PowerPlusWaterMarkObject).
func (d *Document) SetTextWatermark(text string, opts ...WatermarkOptions) {
	d.textWatermark = text
	d.textWatermarkOpts = WatermarkOptions{}
	if len(opts) > 0 {
		d.textWatermarkOpts = opts[0]
	}
}

// SetImageWatermark stores a document-wide image watermark applied to every
// section header as VML v:imagedata (Word watermark drawing).
func (d *Document) SetImageWatermark(imageBytes []byte, opts ...ImageWatermarkOptions) {
	if len(imageBytes) == 0 {
		d.imageWatermark = nil
		d.imageWatermarkName = ""
		d.imageWatermarkOpts = ImageWatermarkOptions{}
		return
	}
	d.imageWatermark = append([]byte(nil), imageBytes...)
	if d.imageWatermarkName == "" {
		d.imageWatermarkName = "watermark.png"
	}
	d.imageWatermarkOpts = ImageWatermarkOptions{}
	if len(opts) > 0 {
		d.imageWatermarkOpts = opts[0]
	}
}

// SetImageWatermarkFile loads a PNG or JPEG from disk as a page watermark.
func (d *Document) SetImageWatermarkFile(path string, opts ...ImageWatermarkOptions) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	name := filepath.Base(path)
	if name == "" || name == "." {
		name = "watermark.png"
	}
	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(name), "."))
	if ext == "jpg" {
		name = strings.TrimSuffix(name, filepath.Ext(name)) + ".jpeg"
	}
	d.imageWatermarkName = name
	d.SetImageWatermark(data, opts...)
	d.imageWatermarkName = name
	return nil
}

func applyTextWatermarkOptions(tw *element.TextWatermark, opts WatermarkOptions) {
	tw.Angle = opts.Angle
	tw.Color = opts.Color
	tw.FontSize = opts.FontSize
	tw.FontName = opts.FontName
	tw.Opacity = opts.Opacity
	tw.Tile = opts.Tile
	tw.TileRows = opts.Rows
	tw.TileCols = opts.Cols
}

func applyImageWatermarkOptions(img *element.Image, opts ImageWatermarkOptions) {
	img.Style.Washout = opts.Washout
	img.Style.Opacity = opts.Opacity
	img.Style.Scale = opts.Scale
	if img.Style.WrappingStyle == "" {
		img.Style.WrappingStyle = style.WrappingBehind
	}
}
