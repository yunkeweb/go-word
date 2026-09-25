# GoWord

[English](#english) | [中文](#chinese)

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)](https://go.dev/)
[![Go Reference](https://pkg.go.dev/badge/github.com/yunkeweb/go-word.svg)](https://pkg.go.dev/github.com/yunkeweb/go-word)
[![CI](https://github.com/yunkeweb/go-word/actions/workflows/test.yml/badge.svg)](https://github.com/yunkeweb/go-word/actions/workflows/test.yml)
[![License: LGPL v3](https://img.shields.io/badge/License-LGPL%20v3-blue.svg)](LICENSE)

---

<a id="english"></a>

## English

### Introduction

**GoWord** is a pure-Go library for creating, reading, and filling Microsoft Word documents. It writes native **OpenXML Word2007 (`.docx`)** packages and ports the core architecture of [PHPWord](https://github.com/PHPOffice/PHPWord) / PHPOffice.

The public API keeps PHPWord names (`AddSection`, `AddText`, `IOFactory`, `TemplateProcessor`) while using idiomatic Go types and `error` returns.

**Zero third-party dependencies:** `go.mod` has no external `require`. OOXML is serialized with the standard library `encoding/xml`; `.docx` packages are packed and unpacked with `archive/zip`.

### Features

- **Zero dependencies** — Go standard library only (`encoding/xml`, `archive/zip`, `image`, …)
- **Word 2007 / OpenXML** — generate and load `.docx` / `.docm` (WordprocessingML)
- **Paragraphs & rich text** — `AddText`, `TextRun`, headings, breaks, hyperlinks, bookmarks
- **Tables** — rows, cells, width, merge, borders, shading
- **Images** — from file path or in-memory bytes
- **Headers & footers** — default, first page, even page; watermarks
- **Formulas** — Office Math (OMML) via `pkg/math` (fractions, superscripts, identifiers, operators)
- **Styles** — font, paragraph, table, numbering, section, paper size and margins
- **Template fill** — `${variable}` placeholders, nested `${block}` loops, `${if}` / `${endif}` clipping, merged-cell row cloning, images and charts
- **Streaming writer** — `NewStreamWriter` writes paragraphs and tables into the ZIP stream as they are produced
- **Memory pooling** — `sync.Pool` reuses buffers, XML writers, and core style structs
- **More** — lists, charts, footnotes/endnotes, HTML import, document properties

### Installation

```sh
go get github.com/yunkeweb/go-word@v0.3.0
```

Requires **Go 1.21+**.

### Quick Start

Create a document, register paragraph styles, and save `output.docx`:

```go
package main

import (
	"log"

	"github.com/yunkeweb/go-word"
	"github.com/yunkeweb/go-word/style"
)

func main() {
	doc := word.New()

	doc.AddFontStyle("rStyle", style.Font{
		Bold: true,
		Size: 16,
		Name: "Calibri",
	})
	doc.AddParagraphStyle("pStyle", style.Paragraph{
		Alignment: style.JcCenter,
		Spacing:   style.Spacing{After: 200},
	})
	doc.AddTitleStyle(1, style.Font{Bold: true, Size: 18}, style.Paragraph{
		Spacing: style.Spacing{After: 240},
	})

	section := doc.AddSection()
	section.AddTitle("Welcome to GoWord", 1)
	section.AddText("Hello, Word 2007.", "rStyle", "pStyle")
	section.AddText("This document was built with the Go standard library.")

	if err := doc.Save("output.docx"); err != nil {
		log.Fatal(err)
	}
}
```

A longer runnable sample lives in [`examples/simple`](examples/simple). API docs: [pkg.go.dev/github.com/yunkeweb/go-word](https://pkg.go.dev/github.com/yunkeweb/go-word).

### Streaming writer

For large tables or reports, write body content incrementally so `document.xml` is never fully buffered:

```go
package main

import (
	"log"
	"os"

	"github.com/yunkeweb/go-word"
	"github.com/yunkeweb/go-word/element"
	"github.com/yunkeweb/go-word/style"
)

func main() {
	f, err := os.Create("report.docx")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	sw := word.NewStreamWriter(f)
	if err := sw.WriteParagraph("Quarterly report", style.Font{Bold: true, Size: 16}); err != nil {
		log.Fatal(err)
	}

	tbl := element.NewTable(style.Table{Width: 9000})
	row := tbl.AddRow()
	row.AddCell(4500).AddText("Item")
	row.AddCell(4500).AddText("Amount")
	if err := sw.WriteTable(tbl); err != nil {
		log.Fatal(err)
	}
	if err := sw.Close(); err != nil {
		log.Fatal(err)
	}
}
```

`Document.Save` and `Document.WriteTo` also stream `word/document.xml` into the ZIP package.

### What's new in v0.3.0

- **Nested block loops** — `${block_a}` may contain `${block_b}`; `CloneNestedBlock` fills hierarchical lists
- **Conditional blocks** — `${if condition}` … `${endif}` drops the covering paragraphs or table rows when false or empty
- **Merged-cell row cloning** — `CloneRow` / `DeleteRow` keep `w:vMerge` restart/continue groups and `gridSpan` together

### What's new in v0.2.0

- **StreamWriter** for incremental paragraphs, tables, and other body elements
- **Streaming `Save` / `WriteTo`** so the full document XML is not held in one buffer
- **`sync.Pool`** reuse of `bytes.Buffer`, XML writers, and Font / Paragraph / Table styles
- Benchmarks versus v0.1.1 (median of 3): `BenchmarkTableRender` about 74% fewer B/op and 37% fewer allocs/op; `BenchmarkSaveDocx` about 52% fewer B/op and 34% fewer allocs/op

### License

GNU Lesser General Public License version 3, same family as PHPWord. See [LICENSE](LICENSE).

---

<a id="chinese"></a>

## 中文

### 项目简介

**GoWord** 是纯 Go 实现的 Word 文档库，面向原生 **OpenXML Word2007（`.docx`）** 读写，移植自 [PHPWord](https://github.com/PHPOffice/PHPWord) / PHPOffice 的核心架构与文档模型。

公开 API 沿用 PHPWord 命名（`AddSection`、`AddText`、`IOFactory`、`TemplateProcessor`），同时采用惯用的 Go 类型与 `error` 返回。

**零第三方依赖：** `go.mod` 不含任何外部 `require`。OpenXML 仅使用标准库 `encoding/xml` 序列化；`.docx` 仅使用 `archive/zip` 打包与解包。

### 核心特性

- **零依赖** — 仅使用 Go 标准库（`encoding/xml`、`archive/zip`、`image` 等）
- **Word 2007 / OpenXML** — 生成与加载 `.docx` / `.docm`（WordprocessingML）
- **段落与富文本** — `AddText`、`TextRun`、标题、换行、超链接、书签
- **表格** — 行、单元格、宽度、合并、边框、底纹
- **图片** — 本地路径或内存字节
- **页眉页脚** — 默认 / 首页 / 偶数页，以及水印
- **公式** — 通过 `pkg/math` 写入 Office Math（OMML）：分数、上下标、标识符与运算符
- **样式配置** — 字体、段落、表格、编号、节、纸张与页边距
- **模板变量** — `${variable}` 占位符、嵌套 `${block}` 循环、`${if}` / `${endif}` 裁剪、合并单元格行克隆，以及图片与图表
- **流式写入** — `NewStreamWriter` 边生成边写入 ZIP 流中的段落与表格
- **对象池** — 通过 `sync.Pool` 复用缓冲区、XML 写入器与核心样式结构
- **更多** — 列表、图表、脚注/尾注、HTML 导入、文档属性

### 安装

```sh
go get github.com/yunkeweb/go-word@v0.3.0
```

需要 **Go 1.21** 或更高版本。

### 快速开始

新建文档、添加段落样式，并保存为 `output.docx`：

```go
package main

import (
	"log"

	"github.com/yunkeweb/go-word"
	"github.com/yunkeweb/go-word/style"
)

func main() {
	doc := word.New()

	doc.AddFontStyle("rStyle", style.Font{
		Bold: true,
		Size: 16,
		Name: "Calibri",
	})
	doc.AddParagraphStyle("pStyle", style.Paragraph{
		Alignment: style.JcCenter,
		Spacing:   style.Spacing{After: 200},
	})
	doc.AddTitleStyle(1, style.Font{Bold: true, Size: 18}, style.Paragraph{
		Spacing: style.Spacing{After: 240},
	})

	section := doc.AddSection()
	section.AddTitle("Welcome to GoWord", 1)
	section.AddText("Hello, Word 2007.", "rStyle", "pStyle")
	section.AddText("This document was built with the Go standard library.")

	if err := doc.Save("output.docx"); err != nil {
		log.Fatal(err)
	}
}
```

更完整的可运行示例见 [`examples/simple`](examples/simple)。包文档：[pkg.go.dev/github.com/yunkeweb/go-word](https://pkg.go.dev/github.com/yunkeweb/go-word)。

### 流式写入器

大表格或长报告可按行写入，避免把整份 `document.xml` 缓存在内存中：

```go
package main

import (
	"log"
	"os"

	"github.com/yunkeweb/go-word"
	"github.com/yunkeweb/go-word/element"
	"github.com/yunkeweb/go-word/style"
)

func main() {
	f, err := os.Create("report.docx")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	sw := word.NewStreamWriter(f)
	if err := sw.WriteParagraph("季度报告", style.Font{Bold: true, Size: 16}); err != nil {
		log.Fatal(err)
	}

	tbl := element.NewTable(style.Table{Width: 9000})
	row := tbl.AddRow()
	row.AddCell(4500).AddText("项目")
	row.AddCell(4500).AddText("金额")
	if err := sw.WriteTable(tbl); err != nil {
		log.Fatal(err)
	}
	if err := sw.Close(); err != nil {
		log.Fatal(err)
	}
}
```

`Document.Save` 与 `Document.WriteTo` 同样将 `word/document.xml` 流式写入 ZIP 包。

### v0.3.0 更新

- **嵌套块循环**：`${block_a}` 内可再套 `${block_b}`；`CloneNestedBlock` 填充多级列表
- **条件块**：`${if condition}` … `${endif}` 在条件为假或空时裁掉覆盖的段落或表格行
- **合并单元格行克隆**：`CloneRow` / `DeleteRow` 将 `w:vMerge` 的 restart/continue 组与 `gridSpan` 一并处理

### v0.2.0 更新

- **StreamWriter**：按段写入段落、表格及其他正文元素
- **流式 `Save` / `WriteTo`**：不再把整份文档 XML 放进单一缓冲区
- **`sync.Pool`**：复用 `bytes.Buffer`、XML 写入器以及 Font / Paragraph / Table 样式
- 相对 v0.1.1 的基准（三次中位数）：`BenchmarkTableRender` 约减少 74% B/op、37% allocs/op；`BenchmarkSaveDocx` 约减少 52% B/op、34% allocs/op

### 开源协议

GNU Lesser General Public License version 3，与 PHPWord 同族。详见 [LICENSE](LICENSE)。
