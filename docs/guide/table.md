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
| `VAlign` | `w:vAlign` | `top` / `center` / `bottom`. Prefer `Cell.SetVAlign`. |
| `TextDir` | `w:textDirection` | `lrTb`, `tbRl`, `btLr`, … Prefer `Cell.SetTextDirection`. |
| `Shading.Fill` | `w:shd w:fill` | Hex fill without `#`. |

## SetHeader / SetCantSplit / SetVAlign / SetTextDirection {#setheader-setcantsplit-setvalign-settextdirection}

Repeating headers (`w:tblHeader`) and unbreakable rows (`w:cantSplit`) live on `w:trPr`. Vertical alignment (`w:vAlign`) and text direction (`w:textDirection`) live on `w:tcPr`. Child order on the row is `cantSplit` → `trHeight` → `tblHeader` (CT_TrPrBase).

### Signature

```go
func (t *Table) SetHeaderRow(row *Row) *Table
func (r *Row) SetHeader(v bool) *Row
func (r *Row) SetCantSplit(v bool) *Row
func (c *Cell) SetVAlign(v string) *Cell
func (c *Cell) SetTextDirection(v string) *Cell
```

### Parameters

| Name | Type | Description |
| --- | --- | --- |
| `row` | `*Row` | Header row. `SetHeaderRow` is `row.SetHeader(true)`. Word repeats consecutive header rows at the top of each page. |
| `v` (SetHeader) | `bool` | Writes `w:tblHeader`. |
| `v` (SetCantSplit) | `bool` | Writes `w:cantSplit` so the row stays on one page. |
| `v` (SetVAlign) | `string` | `top`, `center` / `middle`, `bottom`, `both` / `justify`. |
| `v` (SetTextDirection) | `string` | `lrTb` / `horizontal`, `tbRl` / `vertical`, `btLr`, `lrTbV`, `tbRlV`, `tbLrV`. |

### Notes

- `SetHeader(true)` and `SetHeaderRow` are equivalent. Call either.
- `SetCantSplit` on a header row is common: the repeating header stays intact.
- `w:textDirection` is written **before** `w:vAlign` on `w:tcPr`.

### Example

```go
tbl := sec.AddTable(style.Table{Width: 9000})
hdr := tbl.AddRow(400)
tbl.SetHeaderRow(hdr)
hdr.SetCantSplit(true)
hdr.AddCell(3000).SetVAlign("center").AddText("Item", style.Font{Bold: true})
hdr.AddCell(6000).SetVAlign("center").AddText("Note", style.Font{Bold: true})

row := tbl.AddRow(1600)
row.SetCantSplit(true)
row.AddCell(1500).SetVAlign("center").SetTextDirection("tbRl").AddText("竖排")
row.AddCell(7500).SetVAlign("bottom").AddText("This row stays on one page.")
```

A 48-row sample with repeating headers lives in [`examples/v0.9.0_table_advanced`](https://github.com/yunkeweb/go-word/tree/main/examples/v0.9.0_table_advanced).

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
