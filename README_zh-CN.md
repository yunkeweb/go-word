# GoWord

中文文档 | [English](README.md)

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)](https://go.dev/)
[![Go Reference](https://pkg.go.dev/badge/github.com/yunkeweb/go-word.svg)](https://pkg.go.dev/github.com/yunkeweb/go-word)
[![CI](https://github.com/yunkeweb/go-word/actions/workflows/test.yml/badge.svg)](https://github.com/yunkeweb/go-word/actions/workflows/test.yml)
[![License: LGPL v3](https://img.shields.io/badge/License-LGPL%20v3-blue.svg)](LICENSE)

**GoWord** 是纯 Go 实现的 Microsoft Word 文档库，面向原生 **OpenXML Word 2007（`.docx`）** 的创建、读取、填充与合并，移植自 [PHPWord](https://github.com/PHPOffice/PHPWord) 的核心架构。

公开 API 沿用 PHPWord 命名（`AddSection`、`AddText`、`IOFactory`、`TemplateProcessor`），并使用惯用的 Go 类型与 `error` 返回。

## 核心特性

- **Zero External Dependencies（零第三方依赖）** — 100% Go 标准库（`encoding/xml`、`archive/zip`、`image`、`sync`）。`go.mod` 不含任何外部 `require`。
- **Office Math (OMML)** — `AddMath` 将基础 LaTeX（`\frac{a}{b}`、`x^{2}`、`\sqrt{x_1}`、`\pi`）转译为 Word 原生 `m:oMathPara` / `m:oMath` 公式，可在公式编辑器中双击编辑。
- **Document Merger（无损文档合并）** — `AppendDocument` 克隆源节，并重映射冲突的样式 ID、书签名以及图片 `rId` / 媒体部件，多份文档拼接后资源彼此隔离。
- **DrawingML & Charts** — 柱状、条形、折线、饼图、面积图、堆叠图与双轴组合图，以及矢量形状与文本框（`wps:wsp`、`w:txbxContent`），支持填充、边框与内嵌文字。
- **Streaming Parser（流式解析）** — `StreamExtractText` / `StreamExtractImages` 以 \(O(1)\) 额外内存遍历 `.docx` ZIP，边读边输出段落文本与图片。
- **Template Engine v2** — `${variable}` 占位符、嵌套 `${block}` 循环、二元比较 `${if}` / `${endif}`，以及 `${var | pipe}` 链式过滤器（`formatDate`、`formatCurrency`、`trim`、`upper`、`lower`、`truncate`、`default`）。
- **Advanced Layout（高级排版）** — 多节横纵向混排、多栏排版（`w:cols`）、TOC 自动目录、VML 水印与只读文档保护。

此外还包括：嵌套表格、页眉页脚、图片、列表、脚注/尾注、书签与内部超链接、Markdown/HTML 导入、批注、修订、流式 `Save` / `StreamWriter`，以及 `sync.Pool` 缓冲复用。

## 安装

```sh
go get github.com/yunkeweb/go-word@v0.8.0
```

需要 **Go 1.21** 或更高版本。

## 快速开始

创建文档、插入原生 OMML 公式、绘制 DrawingML 形状，并启用两栏排版：

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

可运行示例：

| 示例 | 内容 |
| --- | --- |
| [`examples/simple`](examples/simple) | 样式、标题与第一份 `.docx` |
| [`examples/v0.8.0_demo`](examples/v0.8.0_demo) | OMML、DrawingML 形状、分栏、`AppendDocument` |
| [`examples/v0.7.0_demo`](examples/v0.7.0_demo) | 面积图/组合图与模板管道 |
| [`examples/all_in_one`](examples/all_in_one) | 更广的 Writer 能力合集 |

包文档：[pkg.go.dev/github.com/yunkeweb/go-word](https://pkg.go.dev/github.com/yunkeweb/go-word)。

## 文档合并

```go
dst := word.New()
src := word.New()
// ... 分别填充两份文档 ...
if err := dst.AppendDocument(src, word.MergeOptions{
	StylePrefix:    "src_",
	BookmarkPrefix: "src_",
	SectionBreak:   "nextPage",
}); err != nil {
	log.Fatal(err)
}
```

冲突的段落样式名与书签名会加上前缀；写出 ZIP 时图片部件获得新的关系 ID。

## 开源协议

GNU Lesser General Public License version 3，与 PHPWord 同族。详见 [LICENSE](LICENSE)。
