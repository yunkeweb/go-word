# Table of Contents

`AddTOC` / `AddTableOfContents` write a Word `TOC` field. Heading paragraphs from `AddTitle` carry `w:outlineLvl`, so Word can collect them when the user updates the field.

## AddTOC / AddTableOfContents

### Signature

```go
func (c *Container) AddTOC(font any, tocStyle any, minDepth, maxDepth int) *TOC
func (d *Document) AddTableOfContents(depth ...int) *element.TOC
```

### Parameters

| Name | Type | Description |
| --- | --- | --- |
| `font` | `any` | Optional `style.Font` for the field paragraph. `nil` uses the default. |
| `tocStyle` | `any` | Optional paragraph style. `nil` is fine. |
| `minDepth` / `maxDepth` | `int` | Outline range. `0` becomes 1 and 9. |
| `depth` | `...int` | `AddTableOfContents(3)` collects levels 1–3. |

The instruction text is `TOC \o "1-3" \h \z \u` (range follows min/max).

### Notes

- Word shows “Update Field” on first open. Right-click the TOC → Update Field → Update entire table.
- `SetUpdateFields(true)` on a [TemplateProcessor](./template) asks Word to refresh fields on open (`w:updateFields`).
- Titles must use `AddTitle`, not styled `AddText`, so the outline level is present.

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
	doc.AddTitleStyle(1, style.Font{Bold: true, Size: 16})
	doc.AddTitleStyle(2, style.Font{Bold: true, Size: 13})

	sec := doc.AddSection()
	sec.AddTitle("Table of Contents", 1)
	sec.AddTOC(nil, nil, 1, 2)
	sec.AddText("Right-click the TOC field in Microsoft Word and choose Update Field.")

	sec.AddTitle("Governing equations", 1)
	sec.AddText("Maximum bending stress.")
	sec.AddTitle("Methods", 2)
	sec.AddText("A concentrated force at mid-span.")
	sec.AddTitle("Results", 1)
	sec.AddText("Peak stress stays below yield.")

	if err := doc.Save("toc.docx"); err != nil {
		log.Fatal(err)
	}
}
```

### What Word shows

After Update Field, the TOC lists “Governing equations”, “Methods”, and “Results” with page numbers and hyperlinks (`\h`). Levels follow `AddTitle` depth.
