# GoWord

[中文文档](README_zh-CN.md) | English

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)](https://go.dev/)
[![Go Reference](https://pkg.go.dev/badge/github.com/yunkeweb/go-word.svg)](https://pkg.go.dev/github.com/yunkeweb/go-word)
[![CI](https://github.com/yunkeweb/go-word/actions/workflows/test.yml/badge.svg)](https://github.com/yunkeweb/go-word/actions/workflows/test.yml)
[![License: LGPL v3](https://img.shields.io/badge/License-LGPL%20v3-blue.svg)](LICENSE)

**GoWord** is a pure-Go library for creating, reading, filling, and merging Microsoft Word documents. It writes native **OpenXML Word 2007 (`.docx`)** packages and ports the core architecture of [PHPWord](https://github.com/PHPOffice/PHPWord).

The public API keeps PHPWord names (`AddSection`, `AddText`, `IOFactory`, `TemplateProcessor`) with idiomatic Go types and `error` returns.

## Features

- **Zero External Dependencies** — 100% Go standard library (`encoding/xml`, `archive/zip`, `image`, `sync`). `go.mod` has no third-party `require`.
- **Office Math (OMML)** — `AddMath` turns basic LaTeX (`\frac{a}{b}`, `x^{2}`, `\sqrt{x_1}`, `\pi`) into Word-native `m:oMathPara` / `m:oMath` equations that open in the built-in equation editor.
- **Document Merger** — `AppendDocument` clones source sections and remaps colliding style IDs, bookmark names, and image `rId` / media parts so several `.docx` trees splice without resource clashes.
- **DrawingML & Charts** — bar, column, line, pie, area, stacked, and dual-axis combo charts, plus vector shapes and text boxes (`wps:wsp`, `w:txbxContent`) with fill, outline, and inner text.
- **Streaming Parser** — `StreamExtractText` / `StreamExtractImages` walk a `.docx` ZIP with \(O(1)\) extra memory, emitting paragraphs and pictures as they are read.
- **Template Engine v2** — `${variable}` placeholders, nested `${block}` loops, binary `${if}` / `${endif}` clipping, and chained `${var | pipe}` filters (`formatDate`, `formatCurrency`, `trim`, `upper`, `lower`, `truncate`, `default`).
- **Advanced Layout** — mixed portrait/landscape sections, multi-column layout (`w:cols`), automatic TOC, VML watermarks, and read-only document protection.

Also included: tables with nested cells, headers/footers, images, lists, footnotes/endnotes, bookmarks and internal hyperlinks, Markdown/HTML import, comments, track changes, a streaming `Save` / `StreamWriter`, and `sync.Pool` buffer reuse.

## Installation

```sh
go get github.com/yunkeweb/go-word@v0.8.0
```

Requires **Go 1.21+**.

## Quick Start

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

Runnable samples:

| Example | What it shows |
| --- | --- |
| [`examples/simple`](examples/simple) | Styles, titles, and a first `.docx` |
| [`examples/v0.8.0_demo`](examples/v0.8.0_demo) | OMML, DrawingML shapes, columns, `AppendDocument` |
| [`examples/v0.7.0_demo`](examples/v0.7.0_demo) | Area/combo charts and template pipes |
| [`examples/all_in_one`](examples/all_in_one) | Broader Writer surface in one file |

API reference: [pkg.go.dev/github.com/yunkeweb/go-word](https://pkg.go.dev/github.com/yunkeweb/go-word).

## Document merger

```go
dst := word.New()
src := word.New()
// ... fill both documents ...
if err := dst.AppendDocument(src, word.MergeOptions{
	StylePrefix:    "src_",
	BookmarkPrefix: "src_",
	SectionBreak:   "nextPage",
}); err != nil {
	log.Fatal(err)
}
```

Colliding paragraph style names and bookmark names are prefixed; image parts receive fresh relationship IDs when the package is written.

## License

GNU Lesser General Public License version 3, same family as PHPWord. See [LICENSE](LICENSE).
