# Shapes & Text Boxes (wps:wsp)

GoWord writes Word 2010 **wordprocessingShape** drawings (`wps:wsp`) with DrawingML preset geometry (`a:prstGeom`). Inner text uses `wps:txbx` / `w:txbxContent` — the same structure Word creates from Insert → Text Box.

## AddShape

### Signature

```go
func (d *Document) AddShape(shapeType ShapeType, opts ShapeOptions) *element.DMLShape
```

### Parameters

| Name | Type | Description |
| --- | --- | --- |
| `shapeType` | `ShapeType` | `ShapeRect`, `ShapeRoundRect`, `ShapeArrow`, `ShapeTextBox`. |
| `opts.Width` / `Height` | `int` | EMU. 1 inch = 914400. Zero uses the writer default. |
| `opts.FillColor` | `string` | Hex without `#` (`srgbClr`). |
| `opts.LineColor` | `string` | Outline hex. |
| `opts.LineWidth` | `int` | Outline width in EMU. |
| `opts.Text` | `string` | Body of `w:txbxContent`. |
| `opts.Font` | `style.Font` | Bold, size, color of the inner run. |

### Geometry

| Constant | `a:prstGeom prst` | Extra |
| --- | --- | --- |
| `ShapeRect` | `rect` | |
| `ShapeRoundRect` | `roundRect` | |
| `ShapeArrow` | `rightArrow` | |
| `ShapeTextBox` | `rect` | Always emits `wps:txbx`. Empty text becomes a space. |

Low-level: `Section.AddDMLShape(prst, width, height, fill, line, lineWidth)`.

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
	doc.SetDefaultFontName("Calibri")
	sec := doc.AddSection()
	sec.AddTitle("DrawingML shapes", 1)

	doc.AddShape(word.ShapeRect, word.ShapeOptions{
		FillColor: "5B9BD5", LineColor: "2E75B6",
		Width: 1828800, Height: 914400,
	})
	doc.AddShape(word.ShapeRoundRect, word.ShapeOptions{
		FillColor: "ED7D31", LineColor: "C45911",
		Text: "Rounded", Font: style.Font{Bold: true, Color: "FFFFFF"},
	})
	doc.AddShape(word.ShapeArrow, word.ShapeOptions{
		FillColor: "70AD47", LineColor: "548235",
	})
	doc.AddShape(word.ShapeTextBox, word.ShapeOptions{
		FillColor: "FFF2CC", LineColor: "BF8F00",
		Text: "w:txbxContent", Font: style.Font{Size: 12, Color: "595959"},
	})

	if err := doc.Save("shapes.docx"); err != nil {
		log.Fatal(err)
	}
}
```

### What Word shows

Four inline drawings: a blue rectangle (2 in × 1 in), an orange pill with white **Rounded**, a green right arrow, and a cream text box whose contents you can click into and type. Selecting a shape shows Word’s Drawing Tools; this is `wps:wsp`, not VML `w:pict` (except watermarks, which remain VML).
