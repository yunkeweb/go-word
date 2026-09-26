# Paragraphs & Runs

A paragraph is `w:p`. A run is `w:r` with `w:t` text. GoWord exposes PHPWord names: `AddText` (one run in one paragraph), `AddTextRun` (several runs in one paragraph), `AddTitle`, `AddLink`, `AddBookmark`, `AddListItem`.

## AddText

### Signature

```go
func (c *Container) AddText(text string, styles ...any) *Text
```

### Parameters

| Name | Type | Description |
| --- | --- | --- |
| `text` | `string` | Paragraph contents. XML-special characters are escaped on write. |
| `styles` | `...any` | Optional `style.Font` then paragraph style (`style.Paragraph` or a registered style name). |

### Notes

- Available on `Section`, `Header`, `Footer`, and `Cell`.
- `style.Font` fields: `Name`, `Size` (points), `Bold`, `Italic`, `Color` (hex without `#`), `Underline`, `Highlight`.

## AddTextRun

### Signature

```go
func (c *Container) AddTextRun(styles ...any) *TextRun
func (p *TextRun) AddText(text string, styles ...any) *Text
func (p *TextRun) AddMath(formula string) *Formula
```

Use `AddTextRun` when bold and italic (or text and math) must share one `w:p`.

## AddTitle / AddLink / AddBookmark / AddListItem

### Signature

```go
func (c *Container) AddTitle(text string, depth int, page ...int) *Title
func (c *Container) AddLink(target, text string, styles ...any) *Link
func (c *Container) AddBookmark(name string) *Bookmark
func (c *Container) AddListItem(text string, depth int, styles ...any) *ListItem
func (c *Container) AddPageBreak() *PageBreak
func (c *Container) AddTextBreak(count ...int)
```

| Method | OpenXML |
| --- | --- |
| `AddTitle` | Paragraph with outline level `w:outlineLvl`. Depth 1–9. |
| `AddLink` | `w:hyperlink r:id` (external) or `w:anchor` (internal, pass `true` as the third style). |
| `AddBookmark` | `w:bookmarkStart` / `w:bookmarkEnd` |
| `AddListItem` | Numbering instance on `w:pPr` |
| `AddPageBreak` | `w:br w:type="page"` |

Register reusable styles on the document:

```go
doc.AddFontStyle("strong", style.Font{Bold: true, Size: 14})
doc.AddParagraphStyle("center", style.Paragraph{Alignment: style.JcCenter})
doc.AddTitleStyle(1, style.Font{Bold: true, Size: 18}, style.Paragraph{
	Spacing: style.Spacing{After: 240},
})
```

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
	doc.SetDefaultFontSize(11)
	doc.AddFontStyle("strong", style.Font{Bold: true, Size: 14, Name: "Calibri"})
	doc.AddParagraphStyle("center", style.Paragraph{Alignment: style.JcCenter})
	doc.AddTitleStyle(1, style.Font{Bold: true, Size: 18}, style.Paragraph{
		Spacing: style.Spacing{After: 240},
	})

	sec := doc.AddSection()
	sec.AddTitle("Paragraphs and runs", 1)
	sec.AddText("Centered heading style.", "strong", "center")

	run := sec.AddTextRun()
	run.AddText("Bold ", style.Font{Bold: true})
	run.AddText("and italic.", style.Font{Italic: true, Color: "C00000"})

	sec.AddBookmark("intro")
	sec.AddLink("https://github.com/yunkeweb/go-word", "GoWord on GitHub")
	sec.AddListItem("First item", 0)
	sec.AddListItem("Nested item", 1)
	sec.AddPageBreak()
	sec.AddText("After the page break.")

	if err := doc.Save("paragraphs.docx"); err != nil {
		log.Fatal(err)
	}
}
```

Word shows a Heading 1, a centered bold line, a mixed-style paragraph, a blue hyperlink, a two-level list, then a second page. Outline levels from `AddTitle` feed the [TOC](./toc).
