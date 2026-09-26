# Multi-Column Layout (w:cols)

`SetColumns` writes a schema-valid `w:cols` node on the **current section**. Word flows that section’s paragraphs into equal columns. Start a new section when the rest of the document should return to a single column.

## SetColumns

### Signature

```go
func (s *Section) SetColumns(num int, space int, showLine bool)
```

### Parameters

| Name | Type | Description |
| --- | --- | --- |
| `num` | `int` | Column count (`w:num`). |
| `space` | `int` | Gap in twips (`w:space`). `720` is 0.5 inch. |
| `showLine` | `bool` | Separator line (`w:sep="1"`). The attribute is `sep`, not `separator`. |

### Notes

- Columns live on `w:sectPr`. Mixing one-column and two-column content requires `AddSection` with `BreakType: "nextPage"` (or `"continuous"`).
- Landscape plus columns: pass `style.OrientationLandscape` on the same section.
- Default page size is A4 (`style.DefaultPageWidth` = 11906 twips, `DefaultPageHeight` = 16838).

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
	doc.SetDefaultAsianFontName("Microsoft YaHei")

	intro := doc.AddSection()
	intro.AddTitle("Single column", 1)
	intro.AddText("This section is a normal single-column page.")

	cols := doc.AddSection(style.Section{
		Orientation: style.OrientationPortrait,
		BreakType:   "nextPage",
		PageSizeW:   style.DefaultPageWidth,
		PageSizeH:   style.DefaultPageHeight,
		MarginTop:   1134, MarginBottom: 1134, MarginLeft: 1134, MarginRight: 1134,
	})
	cols.SetColumns(2, 720, true)
	cols.AddTitle("Two-column layout", 1)
	cols.AddText("The left column starts here. Word splits this section into two equal columns with a separator line (w:cols w:sep).")
	cols.AddText("The second paragraph continues the flow so the columns fill naturally. Formulas and pictures follow the same column stream.")
	cols.AddMath(`\frac{a}{b}`)

	if err := doc.Save("columns.docx"); err != nil {
		log.Fatal(err)
	}
}
```

### What Word shows

Page 1 is a single column. Page 2 is two equal columns with a vertical rule between them. The fraction sits in the column stream, not in a full-width band. Print layout makes the columns obvious; draft view does not.
