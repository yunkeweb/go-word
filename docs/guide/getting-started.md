# Quick Start

This page builds one `.docx` that exercises the four features most teams touch on day one: a native Office Math formula, a DrawingML shape, a two-column section, and `Save`.

The public API keeps PHPWord names (`AddSection`, `AddText`, `IOFactory`) with idiomatic Go types and `error` returns.

## New / AddSection / Save

### Signature

```go
func New() *Document
func (d *Document) AddSection(style ...any) *element.Section
func (d *Document) Save(filename string) error
```

### Parameters

| Name | Type | Description |
| --- | --- | --- |
| `style` | `...any` | Optional `style.Section` (page size, margins, orientation, break type). |
| `filename` | `string` | Destination path. Parent directories must already exist. |

### Notes

- `New` allocates an empty document with Calibri 11 pt as the default run font.
- `AddSection` appends a `w:sectPr` block. Body elements attach to the returned `*element.Section`.
- `Save` is `CreateWriter(doc, "Word2007")` plus a file. Relationship IDs (`rIdN`) and media names (`word/media/imageN`) are assigned at write time.

## Complete example

Save as `main.go` and run `go run .`. Microsoft Word opens `hello.docx` without a repair dialog.

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
	doc.SetDefaultFontSize(11)

	info := doc.GetDocInfo()
	info.Title = "GoWord v0.8.0"
	info.Creator = "GoWord"

	sec := doc.AddSection()
	sec.AddTitle("GoWord v0.8.0", 1)
	sec.AddText("Native Office Math:")
	sec.AddMath(`\frac{a}{b}`)

	p := sec.AddTextRun()
	p.AddText("Pythagoras: ")
	p.AddMath(`x^{2} + y^{2} = z^{2}`)

	doc.AddShape(word.ShapeRoundRect, word.ShapeOptions{
		FillColor: "5B9BD5",
		LineColor: "2E75B6",
		Text:      "DrawingML",
		Font:      style.Font{Bold: true, Color: "FFFFFF"},
	})

	cols := doc.AddSection()
	cols.SetColumns(2, 720, true)
	cols.AddText("The left column starts here. Word flows this section into two equal columns.")
	cols.AddText("A separator line is emitted as w:cols w:sep.")

	if err := doc.Save("hello.docx"); err != nil {
		log.Fatal(err)
	}
}
```

### What Word shows

| Feature | On screen | OpenXML |
| --- | --- | --- |
| Display formula | A centered equation you can double-click | `w:p` / `m:oMathPara` / `m:oMath` / `m:f` |
| Inline formula | Pythagoras sits in the same line as the label | `m:oMath` beside `w:r` |
| Rounded rectangle | Blue pill with white “DrawingML” | `wps:wsp` / `a:prstGeom prst="roundRect"` |
| Two columns | Equal columns with a vertical separator | `w:cols w:num="2" w:space="720" w:sep="1"` |

## IOFactory aliases

### Signature

```go
func CreateWriter(doc *Document, name string) (Writer, error)
func CreateReader(name string) (Reader, error)
func Load(filename string, readerName ...string) (*Document, error)
func Open(filePath string) (*Document, error)
```

`name` is `"Word2007"` (the default). `Load` / `Open` build a full DOM; for O(1) text extraction see [Streaming Parser](./streaming).

## Next

| Topic | Page |
| --- | --- |
| How the ZIP is assembled | [Architecture](./architecture) |
| Paragraphs, tables, images | [Paragraphs & Runs](./paragraph) |
| LaTeX → Word equations | [Office Math](./math) |
| Charts | [DrawingML Charts](./charts) |
