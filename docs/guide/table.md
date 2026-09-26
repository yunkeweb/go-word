# Tables & Nested Cells

Tables follow `w:tbl` → `w:tr` → `w:tc`. Every `w:tc` must contain at least one `w:p`. GoWord writes an empty paragraph when a cell has no children. `Cell` embeds `Container`, so a cell can hold another table.

## AddTable / AddRow / AddCell

### Signature

```go
func (c *Container) AddTable(styles ...any) *Table
func (t *Table) AddRow(height ...int) *Row
func (r *Row) AddCell(width int, st ...any) *Cell
func (t *Table) SetWidth(w int)
```

### Parameters

| Name | Type | Description |
| --- | --- | --- |
| `styles` | `...any` | Optional `style.Table` (`Width` in twips, `Alignment`, borders, shading) or a style name. |
| `height` | `...int` | Optional row height in twips (`w:trHeight`). |
| `width` | `int` | Preferred cell width in twips (`w:tcW`). |
| `st` | `...any` | Optional `style.Cell`. |

### Cell style

| Field | OpenXML | Meaning |
| --- | --- | --- |
| `GridSpan` | `w:gridSpan` | Horizontal merge. Span counts toward `CountColumns`. |
| `VMerge` | `w:vMerge` | `"restart"` on the first row, `"continue"` on the following rows. |
| `VAlign` | `w:vAlign` | `top` / `center` / `bottom`. |
| `Shading.Fill` | `w:shd w:fill` | Hex fill without `#`. |

## Nested tables

Call `AddTable` on the cell:

```go
host := row.AddCell(6000)
inner := host.AddTable(style.Table{Width: 5800})
```

The inner `w:tbl` is a child of `w:tc`, which still ends with a `w:p` after the nested table.

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
	sec.AddTitle("Tables with nested cells", 1)

	tbl := sec.AddTable(style.Table{Width: 9000})
	hdr := tbl.AddRow()
	hdr.AddCell(3000, style.Cell{Shading: style.Shading{Fill: "1F4E79"}}).
		AddText("Region", style.Font{Bold: true, Color: "FFFFFF"})
	hdr.AddCell(6000, style.Cell{Shading: style.Shading{Fill: "1F4E79"}}).
		AddText("Breakdown", style.Font{Bold: true, Color: "FFFFFF"})

	row := tbl.AddRow()
	merged := row.AddCell(3000, style.Cell{VMerge: "restart", VAlign: "center"})
	merged.AddText("APAC")
	host := row.AddCell(6000)
	inner := host.AddTable(style.Table{Width: 5800})
	ir := inner.AddRow()
	ir.AddCell(2900).AddText("Hardware")
	ir.AddCell(2900).AddText("120")
	ir2 := inner.AddRow()
	ir2.AddCell(2900).AddText("Software")
	ir2.AddCell(2900).AddText("80")

	cont := tbl.AddRow()
	cont.AddCell(3000, style.Cell{VMerge: "continue"})
	cont.AddCell(6000).AddText("Continued APAC row uses w:vMerge continue.")

	span := tbl.AddRow()
	span.AddCell(9000, style.Cell{GridSpan: 2}).
		AddText("Footer spans both columns (w:gridSpan=2).")

	if err := doc.Save("table.docx"); err != nil {
		log.Fatal(err)
	}
}
```

### What Word shows

A two-column header in navy, an APAC label that spans two body rows, a nested 2×2 grid in the right column, and a footer cell that spans the full table width. `CloneRow` on a [template](./template) keeps `vMerge` groups and `gridSpan` together.
