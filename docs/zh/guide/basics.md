# 核心 DOM

`Document` 保存命名样式与一个或多个 `Section`。正文元素挂在容器上（`Section`、`Header`、`Footer`、`Cell`、`TextRun`）。

OpenXML 要点：

- 段落是 `w:p` / `w:r` / `w:t`。
- 表格遵循 `w:tbl` → `w:tr` → `w:tc`。每个 `w:tc` 至少要有一个 `w:p`（单元格无子节点时 GoWord 会补空段落）。
- 纵向合并用 `w:vMerge w:val="restart"` 再跟 `"continue"`。横向合并用 `w:gridSpan`。
- 页眉页脚是独立部件（`word/header1.xml`），由 `w:sectPr` 通过 `r:id` 引用。

```go
package main

import (
	"log"

	"github.com/yunkeweb/go-word"
	"github.com/yunkeweb/go-word/element"
	"github.com/yunkeweb/go-word/style"
)

func main() {
	doc := word.New()
	doc.SetDefaultFontName("Calibri")
	doc.SetDefaultFontSize(11)

	doc.AddFontStyle("strong", style.Font{Bold: true, Size: 14, Name: "Calibri"})
	doc.AddParagraphStyle("center", style.Paragraph{Alignment: style.JcCenter})
	doc.AddTitleStyle(1, style.Font{Bold: true, Size: 18}, style.Paragraph{
		Spacing: style.Spacing{After: 240},
	})

	sec := doc.AddSection(style.Section{
		Orientation: style.OrientationPortrait,
		MarginTop:   1440, MarginBottom: 1440,
		MarginLeft:  1440, MarginRight:  1440,
	})

	h := sec.AddHeader()
	h.AddText("GoWord — core DOM")
	f := sec.AddFooter()
	f.AddPreserveText("PAGE")
	first := sec.AddHeader(element.HeaderFirst)
	first.AddText("Cover header")
	doc.SetDifferentFirstPage(true)

	sec.AddTitle("Paragraphs", 1)
	sec.AddText("Centered heading style.", "strong", "center")
	run := sec.AddTextRun()
	run.AddText("Bold ", style.Font{Bold: true})
	run.AddText("and italic.", style.Font{Italic: true})
	sec.AddLink("https://github.com/yunkeweb/go-word", "GoWord on GitHub")
	sec.AddBookmark("intro")

	sec.AddTitle("Table with nested cell", 1)
	tbl := sec.AddTable(style.Table{Width: 9000})
	hdr := tbl.AddRow()
	hdr.AddCell(3000, style.Cell{Shading: style.Shading{Fill: "1F4E79"}}).
		AddText("Region", style.Font{Bold: true, Color: "FFFFFF"})
	hdr.AddCell(6000, style.Cell{Shading: style.Shading{Fill: "1F4E79"}}).
		AddText("Breakdown", style.Font{Bold: true, Color: "FFFFFF"})

	row := tbl.AddRow()
	merged := row.AddCell(3000, style.Cell{VMerge: "restart", VAlign: "center"})
	merged.AddText("APAC")
	host := row.AddCell(6000)
	inner := host.AddTable(style.Table{Width: 5800})
	ir := inner.AddRow()
	ir.AddCell(2900).AddText("Hardware")
	ir.AddCell(2900).AddText("120")
	ir2 := inner.AddRow()
	ir2.AddCell(2900).AddText("Software")
	ir2.AddCell(2900).AddText("80")

	cont := tbl.AddRow()
	cont.AddCell(3000, style.Cell{VMerge: "continue"})
	cont.AddCell(6000).AddText("Continued APAC row uses w:vMerge continue.")

	span := tbl.AddRow()
	span.AddCell(9000, style.Cell{GridSpan: 2}).AddText("Footer spans both columns (w:gridSpan=2).")

	sec.AddTitle("Image", 1)
	sec.AddImageBytes("dot.png", tinyPNG(), style.Image{Width: 32, Height: 32, AltText: "swatch"})

	if err := doc.Save("core-dom.docx"); err != nil {
		log.Fatal(err)
	}
}

func tinyPNG() []byte {
	return []byte{
		0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a, 0x00, 0x00, 0x00, 0x0d,
		0x49, 0x48, 0x44, 0x52, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
		0x08, 0x02, 0x00, 0x00, 0x00, 0x90, 0x77, 0x53, 0xde, 0x00, 0x00, 0x00,
		0x0c, 0x49, 0x44, 0x41, 0x54, 0x08, 0xd7, 0x63, 0xf8, 0xcf, 0xc0, 0x00,
		0x00, 0x00, 0x03, 0x00, 0x01, 0x00, 0x05, 0xfe, 0xd4, 0xef, 0x00, 0x00,
		0x00, 0x00, 0x49, 0x45, 0x4e, 0x44, 0xae, 0x42, 0x60, 0x82,
	}
}
```

`SetDifferentFirstPage(true)` 会在 `w:sectPr` 写出 `w:titlePg`，并加上 `w:headerReference w:type="first"`。媒体在保存时写入 `word/media/image1.png`，并分配唯一 `rId`。

加载已有文件：`CreateReader("Word2007").Load("input.docx")`。若只要抽取文本/图片、不建 DOM，见 [流式解析器](./streaming)。
