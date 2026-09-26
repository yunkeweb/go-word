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
- **SDT form controls** — `AddSDTText`, `AddSDTDropdown`, `AddSDTDate`, and `AddSDTCheckbox` emit Word content controls (`w:sdt` → `w:sdtPr` → `w:sdtContent`), including Word 2010 `w14:checkbox`.
- **Table Mechanics Plus** — repeating headers (`w:tblHeader` via `SetHeader` / `SetHeaderRow`), unbreakable rows (`w:cantSplit`), cell vertical align (`SetVAlign`), and text direction (`SetTextDirection`).
- **Tiled / image watermarks** — `SetTextWatermark(text, WatermarkOptions{Tile, Angle, …})` writes a 3×3 VML grid; `SetImageWatermark` / `SetImageWatermarkFile` add a washout picture watermark.
- **Region edit exceptions** — `Protect` still writes `w:documentProtection`; `AllowEdit("Everyone")` on a paragraph, cell, or table wraps `w:permStart` / `w:permEnd` so those ranges stay editable.
- **Office Math (OMML)** — `AddMath` turns basic LaTeX (`\frac{a}{b}`, `x^{2}`, `\sqrt{x_1}`, `\pi`) into Word-native `m:oMathPara` / `m:oMath` equations that open in the built-in equation editor.
- **Document Merger** — `AppendDocument` clones source sections and remaps colliding style IDs, bookmark names, and image `rId` / media parts so several `.docx` trees splice without resource clashes.
- **DrawingML & Charts** — bar, column, line, pie, area, stacked, and dual-axis combo charts, plus vector shapes and text boxes (`wps:wsp`, `w:txbxContent`) with fill, outline, and inner text.
- **Streaming Parser** — `StreamExtractText` / `StreamExtractImages` walk a `.docx` ZIP with \(O(1)\) extra memory, emitting paragraphs and pictures as they are read.
- **Template Engine v2** — `${variable}` placeholders, nested `${block}` loops, binary `${if}` / `${endif}` clipping, and chained `${var | pipe}` filters (`formatDate`, `formatCurrency`, `trim`, `upper`, `lower`, `truncate`, `default`).
- **Advanced Layout** — mixed portrait/landscape sections, multi-column layout (`w:cols`), automatic TOC, VML watermarks, and read-only document protection.

Also included: tables with nested cells, headers/footers, images, lists, footnotes/endnotes, bookmarks and internal hyperlinks, Markdown/HTML import, comments, track changes, a streaming `Save` / `StreamWriter`, and `sync.Pool` buffer reuse.

## Installation

```sh
go get github.com/yunkeweb/go-word@v0.9.0
```

Requires **Go 1.21+**.

## Quick Start

Create a protected form with SDT controls, a repeating table header, a tiled diagonal watermark, and an editable exception range:

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

	doc.SetTextWatermark("CONFIDENTIAL", word.WatermarkOptions{
		Angle: -45, Color: "C0C0C0", FontSize: 36, Opacity: 0.28,
		Tile: true, Rows: 3, Cols: 3,
	})
	if err := doc.Protect(word.ProtectTypeReadOnly, "goword"); err != nil {
		log.Fatal(err)
	}

	sec := doc.AddSection()
	sec.AddTitle("GoWord v0.9.0", 1)
	sec.AddSDTText("Full name", "full_name", "Enter full name")
	sec.AddSDTDropdown("Department", "dept", map[string]string{
		"eng": "Engineering",
		"hr":  "Human Resources",
	})
	sec.AddSDTDate("Start date", "start_date", "yyyy-MM-dd")
	sec.AddText("Party A: ________________").AllowEdit("Everyone")

	tbl := sec.AddTable(style.Table{Width: 9000})
	hdr := tbl.AddRow()
	tbl.SetHeaderRow(hdr)
	hdr.SetCantSplit(true)
	hdr.AddCell(3000).SetVAlign("center").AddText("Field", style.Font{Bold: true})
	hdr.AddCell(6000).SetVAlign("center").AddText("Value", style.Font{Bold: true})
	row := tbl.AddRow()
	row.AddCell(3000).SetTextDirection("tbRl").AddText("Note")
	row.AddCell(6000).AllowEdit("Everyone").AddText("CN-2026-001")

	if err := doc.Save("hello.docx"); err != nil {
		log.Fatal(err)
	}
}
```

Runnable samples:

| Example | What it shows |
| --- | --- |
| [`examples/v0.9.0_sdt`](examples/v0.9.0_sdt) | Plain text, drop-down, date, checkbox SDT |
| [`examples/v0.9.0_table_advanced`](examples/v0.9.0_table_advanced) | `tblHeader`, `cantSplit`, `vAlign`, `textDirection` |
| [`examples/v0.9.0_watermark_security`](examples/v0.9.0_watermark_security) | Tiled text watermark, image washout, `AllowEdit` |
| [`examples/v0.8.0_demo`](examples/v0.8.0_demo) | OMML, DrawingML shapes, columns, `AppendDocument` |
| [`examples/simple`](examples/simple) | Styles, titles, and a first `.docx` |

API reference: [pkg.go.dev/github.com/yunkeweb/go-word](https://pkg.go.dev/github.com/yunkeweb/go-word). Site: [yunkeweb.github.io/go-word](https://yunkeweb.github.io/go-word/).

## Full-feature matrix

[`tests/matrix`](tests/matrix) writes 80 randomly combined `.docx` files covering every public API from v0.1.0 through v0.9.0 (typography, multi-section, headers/footers, tables, images/shapes, TOC/bookmarks/comments, OMML, charts, SDT, watermark/protection). Generated files stay in `./test_output_docs` (gitignored).

```sh
go run ./tests/matrix
go run tests/matrix/validate_reader.go
powershell -NoProfile -ExecutionPolicy Bypass -File tests/matrix/validate_docs.ps1
```

`validate_reader.go` reverse-parses each file through `word.Open` / `word.Read`, `word.StreamExtractText`, and `word.StreamExtractImages` (there is no `ReadDOM`). `validate_docs.ps1` opens the same files in a headless Microsoft Word COM session (`DisplayAlerts = 0`, `OpenNoRepairDialog`). v0.9.0 scored **80 PASS / 0 FAIL** on both engines.

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
