# 企业级实战案例

三份完整程序，复制到 `main.go` 后执行 `go run .`。每份都会写出一份 Microsoft Word 2007 至 Microsoft 365 可直接打开、无需修复对话框的 `.docx`。

| 案例 | 用到的 API | 输出 |
| --- | --- | --- |
| [1. 合同 / 财务报表](#1-合同--财务报表) | 模板 Pipe、`${block}` / `${if}`、嵌套 `w:tbl` | `contract-report.docx` |
| [2. 学术 / 工程论文排版](#2-学术--工程论文排版) | `AddMath` OMML、`SetColumns`、`AddTOC` | `paper.docx` |
| [3. 跨文档无损拼接](#3-跨文档无损拼接) | `AppendDocument` + 样式 / 书签 / `rId` 隔离 | `dossier.docx` |

相关参考：[模板引擎 v2](./template)、[表格](./table)、[Office Math](./math)、[分栏](./columns)、[TOC](./toc)、[文档合并](./merger)。

---

## 1. 合同 / 财务报表

封面留给法务在 Word 里继续改，P&L 的单元格里再嵌一张表。占位符始终落在同一个 `w:t` 里，因为模板由 GoWord 生成（`NewTemplateProcessorBytes` 不会碰到 GUI 把 `${name}` 拆开的情况）。

下方用到的过滤器：`upper`、`trim`、`formatDate`、`formatCurrency`、`default`。嵌套表走 `Cell.AddTable`——`Cell` 嵌入了 `Container`。

保存为 `main.go`，执行 `go run .`。

```go
package main

import (
	"log"

	"github.com/yunkeweb/go-word"
	"github.com/yunkeweb/go-word/style"
)

func main() {
	src := word.New()
	src.SetDefaultFontName("Calibri")
	src.SetDefaultAsianFontName("Microsoft YaHei")
	src.SetTextWatermark("CONFIDENTIAL")

	sec := src.AddSection()
	hdr := sec.AddHeader()
	hdr.AddText("GoWord · 主服务协议", style.Font{Size: 9, Color: "666666"})
	ftr := sec.AddFooter()
	ftr.AddText("第 ", style.Font{Size: 9})
	ftr.AddPageNumber()
	ftr.AddText(" 页", style.Font{Size: 9})

	sec.AddTitle("主服务协议", 1)
	sec.AddText("甲方：${party_a | upper}")
	sec.AddText("乙方：${party_b | trim | upper}")
	sec.AddText("合同编号 ${contract_no | default:DRAFT}")
	sec.AddText("签署日期 ${signed_at | formatDate:\"2006-01-02\"}")
	sec.AddText("合同金额 ${total | formatCurrency:¥}")
	sec.AddText("${if confidential}本副本依保密协议发放，禁止外传。${endif}")

	sec.AddTitle("明细行", 2)
	lines := sec.AddTable(style.Table{Width: 9000})
	lh := lines.AddRow()
	lh.AddCell(4500, style.Cell{Shading: style.Shading{Fill: "1F4E79"}}).
		AddText("SKU", style.Font{Bold: true, Color: "FFFFFF"})
	lh.AddCell(4500, style.Cell{Shading: style.Shading{Fill: "1F4E79"}}).
		AddText("金额", style.Font{Bold: true, Color: "FFFFFF"})
	lr := lines.AddRow()
	lr.AddCell(4500).AddText("${sku}")
	lr.AddCell(4500).AddText("${amount}")

	sec.AddTitle("一季度分区损益", 2)
	outer := sec.AddTable(style.Table{Width: 9000})
	oh := outer.AddRow()
	oh.AddCell(3000, style.Cell{Shading: style.Shading{Fill: "1F4E79"}}).
		AddText("区域", style.Font{Bold: true, Color: "FFFFFF"})
	oh.AddCell(6000, style.Cell{Shading: style.Shading{Fill: "1F4E79"}}).
		AddText("构成", style.Font{Bold: true, Color: "FFFFFF"})

	apac := outer.AddRow()
	apac.AddCell(3000, style.Cell{VMerge: "restart", VAlign: "center"}).
		AddText("${region | upper}")
	host := apac.AddCell(6000)
	inner := host.AddTable(style.Table{Width: 5800})
	r1 := inner.AddRow()
	r1.AddCell(2900).AddText("硬件")
	r1.AddCell(2900).AddText("${q1_hw | formatCurrency:¥}")
	r2 := inner.AddRow()
	r2.AddCell(2900).AddText("软件")
	r2.AddCell(2900).AddText("${q1_sw | formatCurrency:¥}")

	cont := outer.AddRow()
	cont.AddCell(3000, style.Cell{VMerge: "continue"})
	cont.AddCell(6000).AddText("APAC 在下一视觉行继续（w:vMerge）。")

	total := outer.AddRow()
	total.AddCell(9000, style.Cell{GridSpan: 2, Shading: style.Shading{Fill: "D6DCE4"}}).
		AddText("一季度合计 ${q1_total | formatCurrency:¥}", style.Font{Bold: true})

	sec.AddText("${clauses}")
	sec.AddText("${title | upper}。${body}")
	sec.AddText("${/clauses}")

	if err := src.Protect(word.ProtectTypeReadOnly, "goword"); err != nil {
		log.Fatal(err)
	}

	raw, err := src.Bytes()
	if err != nil {
		log.Fatal(err)
	}
	tp, err := word.NewTemplateProcessorBytes(raw)
	if err != nil {
		log.Fatal(err)
	}
	tp.SetValues(map[string]string{
		"party_a":     "northwind trading",
		"party_b":     "  contoso labs  ",
		"contract_no": "MSA-2026-0418",
		"signed_at":   "2026-09-26T08:00:00Z",
		"total":       "1280000",
		"region":      "apac",
		"q1_hw":       "420000",
		"q1_sw":       "180000",
		"q1_total":    "600000",
	})
	if err := tp.SetCondition("confidential", true); err != nil {
		log.Fatal(err)
	}
	if err := tp.CloneRowAndSetValues("sku", []map[string]string{
		{"sku": "HW-4401 服务器", "amount": "¥420,000.00"},
		{"sku": "SW-2208 许可", "amount": "¥180,000.00"},
		{"sku": "SV-1102 支持", "amount": "¥80,000.00"},
	}); err != nil {
		log.Fatal(err)
	}
	if err := tp.CloneBlockAndSetValues("clauses", []map[string]string{
		{"title": "付款", "body": "发票日起 30 日内电汇至附件 A 账户。"},
		{"title": "责任", "body": "任一方累计责任上限为合同金额。"},
	}); err != nil {
		log.Fatal(err)
	}
	if err := tp.Save("contract-report.docx"); err != nil {
		log.Fatal(err)
	}
}
```

### Word 里会看到什么

- 封面解析为 `NORTHWIND TRADING`、`CONTOSO LABS`、`MSA-2026-0418`、`2026-09-26`、`¥1280000.00`。
- 保密句保留（`confidential` 为 true）。三行 SKU 由同一模板行克隆。
- APAC 占两行视觉高度（`vMerge`）。右侧单元格内嵌 2×2 表，金额走了 `formatCurrency`。
- 合计单元格横跨两列（`gridSpan=2`）。
- 斜向 `CONFIDENTIAL` 水印。编辑时 Word 会要密码 `goword`（`w:documentProtection`——ZIP 本身不加密）。见 [水印与保护](./protect)。

---

## 2. 学术 / 工程论文排版

扉页与目录保持单栏。正文节写成两等栏（`w:cols w:num="2" w:sep="1"`），公式是原生 OMML，审稿人双击即可编辑，而不是打开一张图。标题走 `AddTitle`，这样 `AddTOC` 才有 `w:outlineLvl` 可收集。

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
	doc.AddTitleStyle(1, style.Font{Bold: true, Size: 16})
	doc.AddTitleStyle(2, style.Font{Bold: true, Size: 13})
	doc.SetDifferentFirstPage(true)

	cover := doc.AddSection()
	cover.AddHeader().AddText("GoWord 工程简报")
	cover.AddHeader("first").AddText("封面 — 不编页码")
	cover.AddFooter().AddPageNumber()

	cover.AddTitle("简支梁跨中集中力的最大弯曲应力", 1)
	cover.AddText("云克研究  ·  GoWord v0.8.0  ·  2026 年 9 月 26 日", style.Font{Italic: true, Size: 11})
	cover.AddText("摘要。跨中集中力化为闭式峰值应力。公式是 Office Math（OMML），不是图片。打开文件后请右键目录域，选择“更新域”。")
	cover.AddTitle("目录", 1)
	cover.AddTOC(nil, nil, 1, 2)

	body := doc.AddSection(style.Section{
		BreakType: "nextPage",
		PageSizeW: style.DefaultPageWidth,
		PageSizeH: style.DefaultPageHeight,
		MarginTop: 1134, MarginBottom: 1134, MarginLeft: 1134, MarginRight: 1134,
	})
	body.SetColumns(2, 720, true)
	body.AddHeader().AddText("弯曲应力  ·  双栏正文")
	body.AddFooter().AddPageNumber()

	body.AddTitle("控制方程", 1)
	body.AddText("梁长 L、惯性矩 I、到最外纤维距离 c，跨中集中力 P 下的峰值弯曲应力为")
	body.AddMath(`\sigma = \frac{M c}{I}`)
	body.AddText("其中跨中弯矩")
	body.AddMath(`M = \frac{P L}{4}`)

	body.AddTitle("算例", 2)
	body.AddText("把弯矩代入弯曲公式。嵌套上下标与根号仍可编辑：")
	body.AddMath(`\sigma = \frac{P L c}{4 I}`)
	body.AddMath(`\sqrt{x_1} + \pi`)
	p := body.AddTextRun()
	p.AddText("欧拉恒等式 ")
	p.AddMath(`e^{i\pi} + 1 = 0`)
	p.AddText(" 与前后文字落在同一段落。")

	body.AddTitle("结果", 1)
	body.AddText("峰值应力低于屈服。第一栏填满后，Word 把本段流进第二栏。")
	tbl := body.AddTable(style.Table{Width: 4200})
	th := tbl.AddRow()
	th.AddCell(2100, style.Cell{Shading: style.Shading{Fill: "1F4E79"}}).
		AddText("符号", style.Font{Bold: true, Color: "FFFFFF"})
	th.AddCell(2100, style.Cell{Shading: style.Shading{Fill: "1F4E79"}}).
		AddText("取值", style.Font{Bold: true, Color: "FFFFFF"})
	for _, pair := range [][2]string{{"P", "12 kN"}, {"L", "4.0 m"}, {"c", "75 mm"}} {
		tr := tbl.AddRow()
		tr.AddCell(2100).AddText(pair[0])
		tr.AddCell(2100).AddText(pair[1])
	}

	if err := doc.Save("paper.docx"); err != nil {
		log.Fatal(err)
	}
}
```

### Word 里会看到什么

- 第 1 页是单栏封面：标题、摘要、TOC 域（`TOC \o "1-3" \h \z \u`）。右键该域 → 更新域 → 更新整个表格，页码才会出现。
- 第 2 页开启新节。两等栏带分隔线（属性是 `w:sep`，不是 `w:separator`）。公式在 Word 公式编辑器中打开（`m:oMathPara` / 行内 `m:oMath`）。
- 封面页眉独立，因为 `SetDifferentFirstPage(true)` 写出了 `w:titlePg`。

分栏宽度是节属性。目录若要通栏，请留在上一节。见 [多栏排版](./columns)、[TOC](./toc)、[LaTeX 转 OMML](./math)。

---

## 3. 跨文档无损拼接

三棵独立撰写的 `.docx` 树（封面信、技术说明、附录）共用样式 ID（`Note`）、书签名（`shared`）和媒体名（`word/media/image1.png`）。`AppendDocument` 重映射这些冲突，合并后的 ZIP 不会覆盖图片或样式定义。

`rId` 在写出时分配（`word2007Writer.nextRel`）。克隆图片时清空 `RelationID` 即可——`Save` 时写出器会分配 `imageN` 和新的 `rId`。

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
	cover := word.New()
	cover.AddParagraphStyle("Note", style.Paragraph{Spacing: style.Spacing{After: 120}})
	c := cover.AddSection()
	c.AddTitle("封面信", 1)
	c.AddText("来自封面的书签与藏青色色块。", nil, "Note")
	c.AddBookmark("shared")
	c.AddImageBytes("cover.png", swatch(color.RGBA{R: 31, G: 78, B: 121, A: 255}), style.Image{Width: 48, Height: 24})

	note := word.New()
	note.AddParagraphStyle("Note", style.Paragraph{Spacing: style.Spacing{After: 240}})
	n := note.AddSection()
	n.AddTitle("技术说明", 1)
	n.AddText("同样的样式 id Note、同样的书签名、橙色色块、原生公式。", nil, "Note")
	n.AddBookmark("shared")
	n.AddImageBytes("note.png", swatch(color.RGBA{R: 237, G: 125, B: 49, A: 255}), style.Image{Width: 48, Height: 24})
	n.AddMath(`\frac{P L}{4}`)

	annex := word.New()
	annex.AddParagraphStyle("Note", style.Paragraph{Spacing: style.Spacing{After: 80}})
	a := annex.AddSection()
	a.AddTitle("附录 A — 测试矩阵", 1)
	a.AddText("绿色色块与第三份 Note 样式。", nil, "Note")
	a.AddBookmark("shared")
	a.AddImageBytes("annex.png", swatch(color.RGBA{R: 112, G: 173, B: 71, A: 255}), style.Image{Width: 48, Height: 24})

	if err := cover.AppendDocument(note, word.MergeOptions{
		StylePrefix:    "note_",
		BookmarkPrefix: "note_",
		SectionBreak:   "nextPage",
	}); err != nil {
		log.Fatal(err)
	}
	if err := cover.AppendDocument(annex, word.MergeOptions{
		StylePrefix:    "annex_",
		BookmarkPrefix: "annex_",
		SectionBreak:   "nextPage",
	}); err != nil {
		log.Fatal(err)
	}
	if err := cover.Save("dossier.docx"); err != nil {
		log.Fatal(err)
	}
}

func swatch(c color.RGBA) []byte {
	img := image.NewRGBA(image.Rect(0, 0, 48, 24))
	for y := 0; y < 24; y++ {
		for x := 0; x < 48; x++ {
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

### Word 里会看到什么

三页，每页都以下一页分节符开头。

| 部分 | 合并后的样式 ID | 书签 | 媒体 |
| --- | --- | --- | --- |
| 封面 | `Note`（段后距 120） | `shared` | `word/media/image1.png` 藏青 |
| 技术说明 | `note_Note`（段后距 240） | `note_shared` | `image2.png` 橙色 |
| 附录 | `annex_Note`（段后距 80） | `annex_shared` | `image3.png` 绿色 |

解压该包时图片不会互相覆盖。克隆树里指向 `shared` 的内部超链接会改写到带前缀的名字。技术说明里的 OMML 分数仍可编辑。

选项为空时的默认前缀：样式与书签都是 `src_`。空的 `SectionBreak` 视为 `nextPage`。见 [文档合并](./merger) 与 [FAQ — 样式冲突重映射](./faq)。
