# Quick Start & Installation

GoWord writes native **OpenXML Word 2007 (`.docx`)** packages. The public API keeps PHPWord names (`AddSection`, `AddText`, `IOFactory`, `TemplateProcessor`) with idiomatic Go types and `error` returns.

Requires **Go 1.21+**. License: [GNU LGPL v3](https://github.com/yunkeweb/go-word/blob/main/LICENSE).

## Install

```sh
go get github.com/yunkeweb/go-word@v0.8.0
```

`go.mod` has no third-party `require`. Serialization uses `encoding/xml`; packages use `archive/zip`.

## First document

Save as `main.go` and run `go run .`. The file `hello.docx` opens in Microsoft Word without a repair dialog.

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

OpenXML produced by this program:

| Feature | Node |
| --- | --- |
| Display formula | `w:p` / `m:oMathPara` / `m:oMath` / `m:f` |
| Inline formula | `m:oMath` beside `w:r` |
| Rounded rectangle | `wps:wsp` / `a:prstGeom prst="roundRect"` |
| Two columns | `w:cols w:num="2" w:space="720" w:sep="1"` |

`Save` streams `word/document.xml` into the ZIP. `CreateWriter(doc, "Word2007")` is the PHPWord-compatible alias.

## Next

| Topic | Page |
| --- | --- |
| Paragraph, nested table, image, header | [Core DOM](./basics) |
| LaTeX → Word equations | [Office Math](./math) |
| Charts and shapes | [DrawingML](./drawing) |
| Academic report, merger, finance template | [Examples & Recipes](./examples) |
