# Images & Shapes

Pictures become `word/media/imageN` plus a DrawingML `a:blip r:embed`. Relationship IDs are assigned at write time, so two documents that both add `logo.png` still merge cleanly. Vector shapes (`wps:wsp`) have a dedicated page under [DrawingML Shapes](./shapes); this page covers images and the one-line `Document.AddShape` helper.

## AddImage / AddImageBytes

### Signature

```go
func (c *Container) AddImage(source string, styles ...any) *Image
func (c *Container) AddImageBytes(name string, data []byte, styles ...any) *Image
```

### Parameters

| Name | Type | Description |
| --- | --- | --- |
| `source` | `string` | Filesystem path. The extension selects the content type (`png`, `jpeg`, `gif`, `emf`). |
| `name` | `string` | Part file name used until the writer rewrites it as `imageN`. |
| `data` | `[]byte` | Encoded image bytes (not raw pixels). |
| `styles` | `...any` | Optional `style.Image`. |

### style.Image

| Field | Unit | Notes |
| --- | --- | --- |
| `Width` / `Height` | CSS pixels | Converted to EMU on write. `0` keeps the intrinsic size. |
| `WidthEMU` / `HeightEMU` | EMU | 1 inch = 914400 EMU. Overrides pixel size when set. |
| `AltText` | string | `wp:docPr descr`. |
| `WrappingStyle` | string | `inline` (default), `square`, `tight`, `behind`, `infront`. |
| `Alignment` | string | Paragraph alignment of an inline picture. |

### Notes

- Prefer `AddImageBytes` in tests and servers: no extra files on disk.
- PNG, JPEG, and GIF are detected from the bytes. Invalid bytes still create a part; Word then shows a red X.
- Headers accept the same calls (`header.AddImageBytes(...)`) so a logo can live in `word/header1.xml`.

## Document.AddShape (DrawingML shortcut)

### Signature

```go
func (d *Document) AddShape(shapeType ShapeType, opts ShapeOptions) *element.DMLShape
```

| `ShapeType` | Geometry |
| --- | --- |
| `ShapeRect` | `rect` |
| `ShapeRoundRect` | `roundRect` |
| `ShapeArrow` | `rightArrow` |
| `ShapeTextBox` | `rect` + `wps:txbx` / `w:txbxContent` |

`ShapeOptions`: `Width` / `Height` (EMU), `FillColor`, `LineColor`, `LineWidth`, `Text`, `Font`. Empty text on `ShapeTextBox` is replaced with a space so Word still creates `w:txbxContent`.

Full preset, fill, and text-box examples: [Shapes & Text Boxes](./shapes).

## Complete example

```go
package main

import (
	"log"

	"github.com/yunkeweb/go-word"
	"github.com/yunkeweb/go-word/style"
)

func main() {
	doc := word.New()
	sec := doc.AddSection()
	sec.AddTitle("Images and shapes", 1)
	sec.AddImageBytes("dot.png", tinyPNG(), style.Image{
		Width: 48, Height: 48, AltText: "red swatch",
	})
	sec.AddText("The PNG is stored as word/media/image1.png with a unique rId.")

	doc.AddShape(word.ShapeRoundRect, word.ShapeOptions{
		FillColor: "5B9BD5",
		LineColor: "2E75B6",
		Text:      "DrawingML",
		Font:      style.Font{Bold: true, Color: "FFFFFF"},
	})

	if err := doc.Save("image.docx"); err != nil {
		log.Fatal(err)
	}
}

func tinyPNG() []byte {
	return []byte{
		0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a, 0x00, 0x00, 0x00, 0x0d,
		0x49, 0x48, 0x44, 0x52, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
		0x08, 0x02, 0x00, 0x00, 0x00, 0x90, 0x77, 0x53, 0xde, 0x00, 0x00, 0x00,
		0x0c, 0x49, 0x44, 0x41, 0x54, 0x08, 0xd7, 0x63, 0xf8, 0xcf, 0xc0, 0x00,
		0x00, 0x00, 0x03, 0x00, 0x01, 0x00, 0x05, 0xfe, 0xd4, 0xef, 0x00, 0x00,
		0x00, 0x00, 0x49, 0x45, 0x4e, 0x44, 0xae, 0x42, 0x60, 0x82,
	}
}
```

Word shows a small red square inline, then a blue rounded rectangle with white text. Extract the picture later with [StreamExtractImages](./streaming).
