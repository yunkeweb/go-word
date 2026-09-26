# 实战案例库

三份完整程序，保存为 `main.go` 后执行 `go run .`。均使用 v0.8.0 公开 API。

| 案例 | 输出 | 功能 |
| --- | --- | --- |
| A | `academic-report.docx` | OMML 公式、双栏正文、目录、页眉页脚 |
| B | `spliced.docx` | `AppendDocument`，冲突样式 / 书签 / PNG 部件隔离 |
| C | `invoice-INV-1042.docx`、`invoice-INV-1043.docx` | 内存模板、Pipe 过滤器、`${block}` / `${if}` |

## 案例 A — 学术 / 工程报告

公式写成 Word 原生 Office Math，方法节使用双栏（`w:cols w:num="2" w:space="720" w:sep="1"`），并插入 TOC 域。

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
	info := doc.GetDocInfo()
	info.Title = "Stress analysis of a simply supported beam"
	info.Creator = "GoWord"
	info.Subject = "Academic report with OMML and two-column layout"

	doc.SetDefaultFontName("Calibri")
	doc.SetDefaultAsianFontName("Microsoft YaHei")
	doc.SetDefaultFontSize(11)
	doc.AddTitleStyle(1, style.Font{Bold: true, Size: 16})
	doc.AddTitleStyle(2, style.Font{Bold: true, Size: 13})

	cover := doc.AddSection()
	h := cover.AddHeader()
	h.AddText("GoWord academic report")
	f := cover.AddFooter()
	f.AddText("Page ")
	f.AddPreserveText("PAGE")
	cover.AddHeader(element.HeaderFirst).AddText("Cover")
	doc.SetDifferentFirstPage(true)

	cover.AddTitle("Stress analysis of a simply supported beam", 1)
	cover.AddText("An engineering note that mixes native Office Math with a two-column methods section.")
	cover.AddTitle("Contents", 2)
	cover.AddTOC(nil, nil, 1, 2)
	cover.AddText("Right-click the TOC field in Microsoft Word and choose Update Field.")

	cover.AddTitle("Governing equations", 1)
	cover.AddText("Maximum bending stress:")
	cover.AddMath(`\sigma = \frac{M y}{I}`)
	p := cover.AddTextRun()
	p.AddText("Second moment of area for a rectangle: ")
	p.AddMath(`I = \frac{b h^{3}}{12}`)
	cover.AddText("Shear formula:")
	cover.AddMath(`\tau = \frac{V Q}{I t}`)

	methods := doc.AddSection(style.Section{
		Orientation: style.OrientationPortrait,
		BreakType:   "nextPage",
		PageSizeW:   style.DefaultPageWidth,
		PageSizeH:   style.DefaultPageHeight,
		MarginTop:   1134, MarginBottom: 1134, MarginLeft: 1134, MarginRight: 1134,
	})
	methods.SetColumns(2, 720, true)
	methods.AddTitle("Methods", 1)
	methods.AddText("The left column records the loading: a concentrated force at mid-span on a simply supported beam of length L.")
	methods.AddText("The right column continues with the section properties. Word flows these paragraphs into two equal columns with a separator line.")
	run := methods.AddTextRun()
	run.AddText("Check: ")
	run.AddMath(`\sqrt{x_1} + \pi`)

	if err := doc.Save("academic-report.docx"); err != nil {
		log.Fatal(err)
	}
}
```

OpenXML：展示公式为 `m:oMathPara`；行内公式为与 `w:r` 并列的 `m:oMath`；栏设置写在第二节 `w:sectPr` 的 `w:cols` 上。

## 案例 B — 多文档自动化无损拼接

两份源文档共用段落样式 `Note` 与书签 `shared`。`AppendDocument` 给冲突标识符加前缀，并在写出时分配独立的图片 `rId`。

```go
package main

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"log"

	"github.com/yunkeweb/go-word"
	"github.com/yunkeweb/go-word/style"
)

func main() {
	chapter1 := word.New()
	chapter1.AddParagraphStyle("Note", style.Paragraph{Spacing: style.Spacing{After: 120}})
	s1 := chapter1.AddSection()
	s1.AddTitle("Chapter 1 — Specimens", 1)
	s1.AddText("Macrograph from the first lab.", nil, "Note")
	s1.AddBookmark("shared")
	s1.AddImageBytes("lab-a.png", swatch(color.RGBA{R: 31, G: 78, B: 121, A: 255}), style.Image{Width: 80, Height: 40, AltText: "lab A"})

	chapter2 := word.New()
	chapter2.AddParagraphStyle("Note", style.Paragraph{Spacing: style.Spacing{After: 240}})
	s2 := chapter2.AddSection()
	s2.AddTitle("Chapter 2 — Fracture surface", 1)
	s2.AddText("Macrograph from the second lab, with a different Note style.", nil, "Note")
	s2.AddBookmark("shared")
	s2.AddImageBytes("lab-b.png", swatch(color.RGBA{R: 237, G: 125, B: 49, A: 255}), style.Image{Width: 80, Height: 40, AltText: "lab B"})
	s2.AddMath(`\frac{\sigma}{E}`)

	if err := chapter1.AppendDocument(chapter2, word.MergeOptions{
		StylePrefix:    "ch2_",
		BookmarkPrefix: "ch2_",
		SectionBreak:   "nextPage",
	}); err != nil {
		log.Fatal(err)
	}
	if err := chapter1.Save("spliced.docx"); err != nil {
		log.Fatal(err)
	}
}

func swatch(c color.RGBA) []byte {
	img := image.NewRGBA(image.Rect(0, 0, 80, 40))
	for y := 0; y < 40; y++ {
		for x := 0; x < 80; x++ {
			img.Set(x, y, c)
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		log.Fatal(err)
	}
	return buf.Bytes()
}
```

拼接后：第一章保留 `Note` 与 `shared`；第二章变为 `ch2_Note` 与 `ch2_shared`；媒体分别落在 `word/media/image1.png` 与 `image2.png`。

## 案例 C — 财务 / 合同报表批量导出

用 GoWord 生成模板，经 `NewTemplateProcessorBytes` 填充，按记录导出发票。Pipe 过滤器格式化日期与货币；`${block}` 克隆明细行；`${if}` 裁剪逾期 / 已付文案。

```go
package main

import (
	"fmt"
	"log"

	"github.com/yunkeweb/go-word"
)

func main() {
	raw, err := invoiceTemplate()
	if err != nil {
		log.Fatal(err)
	}

	type line struct{ SKU, Qty, Amount string }
	type invoice struct {
		ID, Client, When, Total string
		Paid, Overdue           bool
		Lines                   []line
	}
	batch := []invoice{
		{
			ID: "INV-1042", Client: "  acme labs  ", When: "2026-09-01T00:00:00Z",
			Total: "279.50", Paid: true, Overdue: false,
			Lines: []line{{"A-01", "2", "199.50"}, {"B-02", "1", "80.00"}},
		},
		{
			ID: "INV-1043", Client: "beta works", When: "2026-08-15T00:00:00Z",
			Total: "50.00", Paid: false, Overdue: true,
			Lines: []line{{"C-09", "1", "50.00"}},
		},
	}

	for _, inv := range batch {
		tp, err := word.NewTemplateProcessorBytes(raw)
		if err != nil {
			log.Fatal(err)
		}
		tp.SetValues(map[string]string{
			"invoice": inv.ID,
			"client":  inv.Client,
			"when":    inv.When,
			"total":   inv.Total,
		})
		items := make([]word.BlockData, 0, len(inv.Lines))
		for _, ln := range inv.Lines {
			items = append(items, word.BlockData{
				Values: map[string]string{"sku": ln.SKU, "qty": ln.Qty, "amount": ln.Amount},
			})
		}
		if err := tp.CloneNestedBlock("items", items); err != nil {
			log.Fatal(err)
		}
		if err := tp.SetCondition("paid", inv.Paid); err != nil {
			log.Fatal(err)
		}
		if err := tp.SetCondition("overdue", inv.Overdue); err != nil {
			log.Fatal(err)
		}
		out := fmt.Sprintf("invoice-%s.docx", inv.ID)
		if err := tp.Save(out); err != nil {
			log.Fatal(err)
		}
		fmt.Println("wrote", out)
	}
}

func invoiceTemplate() ([]byte, error) {
	src := word.New()
	sec := src.AddSection()
	sec.AddTitle("Invoice ${invoice | trim | upper}", 1)
	sec.AddText("Client: ${client | trim | upper}")
	sec.AddText("Date: ${when | formatDate:\"2006-01-02\"}")
	sec.AddText("${items}")
	sec.AddText("${sku} × ${qty} = ${amount | formatCurrency:¥}")
	sec.AddText("${/items}")
	sec.AddText("Total ${total | formatCurrency:¥}")
	sec.AddText("${if paid}Thank you for your payment. This invoice is closed.${endif}")
	sec.AddText("${if overdue}OVERDUE — please settle this contract immediately.${endif}")
	return src.Bytes()
}
```

每个 `${...}` 都落在同一个 `w:t` run 内。`CloneNestedBlock` 给嵌套宏加上 `#n` 后缀，明细值不会串到下一张发票。

## 相关指南

- [Office Math 原生公式](./math) — 案例 A 使用的 LaTeX → OMML
- [文档无损合并](./merger) — 案例 B 使用的样式 / 书签 / 媒体隔离
- [模板引擎 v2](./template) — 案例 C 使用的管道、块与 `${if}`
- [高级排版与安全](./layout) — 分栏、目录、页眉
