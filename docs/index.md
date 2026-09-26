---
layout: home

hero:
  name: GoWord
  text: Industrial Word (.docx) Engine
  tagline: 100% pure Go standard library, zero third-party dependencies — an industrial-grade engine for Microsoft Word (.docx) processing.
  actions:
    - theme: brand
      text: Get Started
      link: /guide/getting-started
    - theme: alt
      text: Recipes
      link: /guide/recipes
    - theme: alt
      text: GitHub
      link: https://github.com/yunkeweb/go-word

features:
  - title: Zero Dependencies
    details: encoding/xml, archive/zip, image, and sync only. go.mod has no third-party require. Drop the module into air-gapped and regulated environments.
  - title: OMML Formula Engine
    details: AddMath compiles LaTeX such as \frac{a}{b} and x^{2} into Word-native m:oMathPara equations that open in the built-in equation editor.
  - title: Lossless Document Merger
    details: AppendDocument remaps colliding style IDs, bookmark names, and image rIds so several .docx trees splice without resource clashes.
  - title: DrawingML Charts
    details: Bar, column, line, pie, area, stacked, and dual-axis combo charts, plus vector shapes and text boxes (wps:wsp / w:txbxContent).
  - title: O(1) Streaming Parser
    details: StreamExtractText and StreamExtractImages walk a .docx ZIP with xml.Decoder. Paragraph buffers are discarded after each callback.
  - title: Template Engine v2
    details: Nested ${block} loops, ${if} comparisons, and chained ${var | pipe} filters (formatDate, formatCurrency, trim, upper, truncate, default).
---

<p align="center">
  <a href="https://pkg.go.dev/github.com/yunkeweb/go-word"><img src="https://pkg.go.dev/badge/github.com/yunkeweb/go-word.svg" alt="Go Reference" /></a>
  <a href="https://github.com/yunkeweb/go-word/actions/workflows/test.yml"><img src="https://github.com/yunkeweb/go-word/actions/workflows/test.yml/badge.svg" alt="CI" /></a>
  <a href="https://github.com/yunkeweb/go-word/blob/main/LICENSE"><img src="https://img.shields.io/badge/License-LGPL%20v3-blue.svg" alt="License: LGPL v3" /></a>
  <a href="https://go.dev/"><img src="https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go" alt="Go 1.21+" /></a>
  <a href="https://github.com/yunkeweb/go-word/releases/tag/v0.8.0"><img src="https://img.shields.io/badge/release-v0.8.0-green.svg" alt="v0.8.0" /></a>
</p>

## Core advantages

Production Word pipelines usually stall on four constraints: shipping a binary into an air-gapped network, keeping RAM flat on multi-hundred-page extracts, emitting equations Microsoft Word can edit, and filling legal templates without a second templating stack.

| Constraint | GoWord | Typical PHP / Python stack | Other Go `.docx` writers |
| --- | --- | --- | --- |
| Language & deps | Go 1.21+ **standard library only** | runtime + XML / ZIP extras | often stdlib, write-only surface |
| Extract memory | `StreamExtractText` **O(1)** extra heap | load the whole package / DOM | full document tree |
| Equations | LaTeX → native **OMML** (`m:oMathPara`) | image, OLE, or omit | omit |
| Templates | `${var \| pipe}` + `${block}` / `${if}` | PHPWord macros or Jinja→docx | omit or string replace |
| Merge isolation | `AppendDocument` remaps styles, bookmarks, `rId` | ZIP copy, colliding `image1.png` | omit |
| License | LGPL v3 | mixed | AGPL or MIT write-only |

GoWord keeps the PHPWord names (`AddSection`, `AddText`, `IOFactory`, `TemplateProcessor`) on a pure-Go OpenXML writer. Three copy-paste programs that combine these advantages live under [Enterprise Recipes](/guide/recipes).

## Benchmarks

Numbers from `go test -benchmem` on Intel Core i7-10870H, Windows amd64. Allocations are the stable signal; wall time moves with disk and CPU. Full tables, hardware notes, and a streaming-extract occupancy example: [Benchmarks](/guide/benchmarks).

| Workload | ns/op | B/op | allocs/op |
| --- | ---: | ---: | ---: |
| `BenchmarkSaveDocx` — 30 paragraphs + 20×5 table | 1.88e6 | 163 KiB | 949 |
| `BenchmarkTableRender` — 50×8 table | 2.76e6 | 180 KiB | 1820 |
| `BenchmarkTemplateProcess` — two `${}` replacements | 1.43e6 | 238 KiB | 703 |
| `BenchmarkStreamWriter` — incremental ZIP | 1.55e6 | 142 KiB | 627 |

`B/op` is extra heap per iteration, not the size of the `.docx`. Stream extract discards each paragraph buffer after the callback, so extra memory stays O(1) relative to file size.

## Feature comparison

GoWord ports the PHPWord public API (`AddSection`, `AddText`, `IOFactory`, `TemplateProcessor`) onto a pure-Go OpenXML writer. The table below compares the surface that production Word pipelines usually need.

| Capability | GoWord | PHPWord | unioffice | fumiama/go-docx | gingfrederik/docx |
| --- | :---: | :---: | :---: | :---: | :---: |
| Language | Go 1.21+ | PHP | Go (commercial) | Go | Go |
| License | LGPL v3 | LGPL v3 | Commercial | AGPL-3.0 | MIT |
| Third-party Go deps | **None** | n/a | None (paid SDK) | None | None |
| Create `.docx` | ✓ | ✓ | ✓ | ✓ | ✓ |
| Read / round-trip `.docx` | ✓ | ✓ | ✓ | ✓ | |
| Native OMML from LaTeX | ✓ | | ✓ | | |
| DrawingML charts (bar / line / pie / area / combo) | ✓ | ✓ | ✓ | | |
| Vector shapes `wps:wsp` | ✓ | ✓ | ✓ | ✓ | |
| Multi-column `w:cols` | ✓ | ✓ | ✓ | | |
| Nested tables + `vMerge` / `gridSpan` | ✓ | ✓ | ✓ | ✓ | |
| Template `${var \| pipe}` + `${block}` / `${if}` | ✓ | ✓ | templates | | |
| O(1) stream extract | ✓ | | | | |
| Merge with style / bookmark / `rId` isolation | ✓ | | ✓ | | |
| Watermark + `w:documentProtection` | ✓ | ✓ | ✓ | | |

PHPWord is the API ancestor. unioffice is a paid, multi-format Office SDK. The two lightweight Go writers cover paragraphs (and, for go-docx, pictures and tables) and stop short of OMML, charts, streaming extract, and identifier-safe merge.

## Quick start

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
	sec.AddMath(`\frac{a}{b}`)
	doc.AddShape(word.ShapeRoundRect, word.ShapeOptions{
		FillColor: "5B9BD5", Text: "DrawingML",
		Font: style.Font{Bold: true, Color: "FFFFFF"},
	})
	if err := doc.Save("hello.docx"); err != nil {
		log.Fatal(err)
	}
}
```

```sh
go get github.com/yunkeweb/go-word@v0.8.0
```

Continue with [Installation](/guide/installation), [Quick Start](/guide/getting-started), and the three [Enterprise Recipes](/guide/recipes) (contracts, academic papers, lossless splice). Full signatures live on [pkg.go.dev](https://pkg.go.dev/github.com/yunkeweb/go-word).
