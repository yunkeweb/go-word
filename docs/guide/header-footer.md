# Headers, Footers & Page Numbers

Headers and footers are separate ZIP parts (`word/header1.xml`, `word/footer1.xml`) referenced from `w:sectPr` via `r:id`. Three types exist, matching PHPWord: default (odd), first page, and even page.

## AddHeader / AddFooter

### Signature

```go
func (s *Section) AddHeader(typ ...string) *Header
func (s *Section) AddFooter(typ ...string) *Footer
```

### Parameters

| Name | Type | Description |
| --- | --- | --- |
| `typ` | `...string` | `element.HeaderAuto` (`"default"`), `element.HeaderFirst` (`"first"`), `element.HeaderEven` (`"even"`). Omitted → default. |

`Header` and `Footer` embed `Container`, so `AddText`, `AddImageBytes`, `AddPreserveText`, `AddPageNumber`, and `AddNumPages` all work.

## Page fields

### Signature

```go
func (c *Container) AddPreserveText(text string, styles ...any) *PreserveText
func (c *Container) AddPageNumber() *Field
func (c *Container) AddNumPages() *Field
func (d *Document) SetDifferentFirstPage(enable bool)
func (d *Document) SetEvenAndOddHeaders(enable bool)
```

| Call | OpenXML |
| --- | --- |
| `AddPreserveText("PAGE")` | `w:instrText` with `PAGE` |
| `AddPageNumber()` | `w:fldChar` + `PAGE` field |
| `AddNumPages()` | `NUMPAGES` field |
| `SetDifferentFirstPage(true)` | `w:titlePg` on `w:sectPr` |
| `SetEvenAndOddHeaders(true)` | `w:evenAndOddHeaders` in settings |

### Notes

- First-page headers are ignored unless `SetDifferentFirstPage(true)`.
- Even-page headers are ignored unless `SetEvenAndOddHeaders(true)`.
- Fields show their cached value until the user right-clicks → Update Field (or prints).

## Complete example

```go
package main

import (
	"log"

	"github.com/yunkeweb/go-word"
	"github.com/yunkeweb/go-word/element"
)

func main() {
	doc := word.New()
	doc.SetDefaultFontName("Calibri")
	doc.SetDifferentFirstPage(true)
	doc.SetEvenAndOddHeaders(true)

	sec := doc.AddSection()
	sec.AddHeader().AddText("GoWord — odd pages")
	sec.AddHeader(element.HeaderFirst).AddText("GoWord — cover")
	sec.AddHeader(element.HeaderEven).AddText("GoWord — even pages")

	f := sec.AddFooter()
	f.AddText("Page ")
	f.AddPreserveText("PAGE")
	f.AddText(" of ")
	f.AddNumPages()
	sec.AddFooter(element.HeaderFirst).AddText("Cover footer")

	sec.AddTitle("Headers and page numbers", 1)
	sec.AddText("This is the first page (cover header).")
	sec.AddPageBreak()
	sec.AddText("Second page uses the odd header.")
	sec.AddPageBreak()
	sec.AddText("Third page uses the even header.")

	if err := doc.Save("headers.docx"); err != nil {
		log.Fatal(err)
	}
}
```

Word shows three different headers across three pages, with `PAGE` / `NUMPAGES` in the default footer. Print layout or print preview refreshes the fields.
