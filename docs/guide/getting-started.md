# Quick Start

GoWord writes native **OpenXML Word 2007 (`.docx`)** packages. The public API keeps PHPWord names (`AddSection`, `AddText`, `IOFactory`, `TemplateProcessor`) with idiomatic Go types and `error` returns.

Requires **Go 1.21+**. License: [GNU LGPL v3](https://github.com/yunkeweb/go-word/blob/main/LICENSE).

## Install

```sh
go get github.com/yunkeweb/go-word@v0.8.0
```

## Hello, Word

Create a document, insert a native OMML formula, draw a DrawingML shape, and enable two-column layout:

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

## Next steps

| Topic | Page |
| --- | --- |
| Paragraphs, tables, images, headers | [Basic DOM](./basics) |
| LaTeX → Word equations | [Office Math](./math) |
| Shapes and charts | [DrawingML & Charts](./drawing) |
| \(O(1)\) ZIP extractors | [Streaming Parser](./streaming) |
| `${block}` / `${if}` / pipes | [Template Engine v2](./template) |
| Splice documents | [Document Merger](./merger) |
| Columns, watermark, protect, TOC | [Advanced Layout](./layout) |

Runnable samples live under [`examples/`](https://github.com/yunkeweb/go-word/tree/main/examples). Full API: [pkg.go.dev/github.com/yunkeweb/go-word](https://pkg.go.dev/github.com/yunkeweb/go-word).
