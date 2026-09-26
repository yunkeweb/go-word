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
- **SDT 结构化表单控件** — `AddSDTText`、`AddSDTDropdown`、`AddSDTDate`、`AddSDTCheckbox` 写出 Word 内容控件（`w:sdt` → `w:sdtPr` → `w:sdtContent`），复选框使用 Word 2010 `w14:checkbox`。
- **表格高级版式** — 跨页重复页眉（`w:tblHeader`，`SetHeader` / `SetHeaderRow`）、行禁止跨页断裂（`w:cantSplit`）、单元格垂直对齐（`SetVAlign`）与文本方向（`SetTextDirection`）。
- **平铺 / 图片水印** — `SetTextWatermark(text, WatermarkOptions{Tile, Angle, …})` 写出 3×3 VML 网格；`SetImageWatermark` / `SetImageWatermarkFile` 写入洗白图片水印。
- **区域编辑例外** — `Protect` 仍写出 `w:documentProtection`；段落、单元格或表格上的 `AllowEdit("Everyone")` 包裹 `w:permStart` / `w:permEnd`，这些区域在保护开启后仍可编辑。
- **Office Math (OMML)** — `AddMath` 将基础 LaTeX（`\frac{a}{b}`、`x^{2}`、`\sqrt{x_1}`、`\pi`）转译为 Word 原生 `m:oMathPara` / `m:oMath` 公式，可在公式编辑器中双击编辑。
- **Document Merger（无损文档合并）** — `AppendDocument` 克隆源节，并重映射冲突的样式 ID、书签名以及图片 `rId` / 媒体部件，多份文档拼接后资源彼此隔离。
- **DrawingML & Charts** — 柱状、条形、折线、饼图、面积图、堆叠图与双轴组合图，以及矢量形状与文本框（`wps:wsp`、`w:txbxContent`），支持填充、边框与内嵌文字。
- **Streaming Parser（流式解析）** — `StreamExtractText` / `StreamExtractImages` 以 \(O(1)\) 额外内存遍历 `.docx` ZIP，边读边输出段落文本与图片。
- **Template Engine v2** — `${variable}` 占位符、嵌套 `${block}` 循环、二元比较 `${if}` / `${endif}`，以及 `${var | pipe}` 链式过滤器（`formatDate`、`formatCurrency`、`trim`、`upper`、`lower`、`truncate`、`default`）。
- **Advanced Layout（高级排版）** — 多节横纵向混排、多栏排版（`w:cols`）、TOC 自动目录、VML 水印与只读文档保护。

此外还包括：嵌套表格、页眉页脚、图片、列表、脚注/尾注、书签与内部超链接、Markdown/HTML 导入、批注、修订、流式 `Save` / `StreamWriter`，以及 `sync.Pool` 缓冲复用。

## 安装

```sh
go get github.com/yunkeweb/go-word@v0.9.0
```

需要 **Go 1.21** 或更高版本。

## 快速开始

创建一份受保护的表单：SDT 控件、跨页重复表头、平铺斜向水印，以及可编辑例外区域：

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
	row.AddCell(3000).SetTextDirection("tbRl").AddText("备注")
	row.AddCell(6000).AllowEdit("Everyone").AddText("CN-2026-001")

	if err := doc.Save("hello.docx"); err != nil {
		log.Fatal(err)
	}
}
```

可运行示例：

| 示例 | 内容 |
| --- | --- |
| [`examples/v0.9.0_sdt`](examples/v0.9.0_sdt) | 纯文本、下拉、日期、复选框 SDT |
| [`examples/v0.9.0_table_advanced`](examples/v0.9.0_table_advanced) | `tblHeader`、`cantSplit`、`vAlign`、`textDirection` |
| [`examples/v0.9.0_watermark_security`](examples/v0.9.0_watermark_security) | 平铺文字水印、图片洗白、`AllowEdit` |
| [`examples/v0.8.0_demo`](examples/v0.8.0_demo) | OMML、DrawingML 形状、分栏、`AppendDocument` |
| [`examples/simple`](examples/simple) | 样式、标题与第一份 `.docx` |

包文档：[pkg.go.dev/github.com/yunkeweb/go-word](https://pkg.go.dev/github.com/yunkeweb/go-word)。站点：[yunkeweb.github.io/go-word](https://yunkeweb.github.io/go-word/zh/)。

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
