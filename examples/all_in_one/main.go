// Command all_in_one builds all_features_demo.docx, exercising every major
// GoWord surface: styles, merged tables, StreamWriter, charts, Markdown/HTML
// import, the template engine, comments, and track changes.
package main

import (
	"bytes"
	"fmt"
	"log"

	"github.com/yunkeweb/go-word"
	"github.com/yunkeweb/go-word/element"
	"github.com/yunkeweb/go-word/style"
)

const outFile = "all_features_demo.docx"

func main() {
	if err := demoStreamWriter(); err != nil {
		log.Fatal(err)
	}

	doc := word.New()
	info := doc.GetDocInfo()
	info.Title = "GoWord All Features Demo"
	info.Creator = "GoWord"
	info.Subject = "v0.4.0 integration sample"
	info.Keywords = "charts, markdown, html, comments, track changes, template, stream"
	info.Company = "yunkeweb"

	doc.SetDefaultFontName("Calibri")
	doc.SetDefaultAsianFontName("Microsoft YaHei")
	doc.SetDefaultFontSize(11)

	registerStyles(doc)

	sec := doc.AddSection(style.Section{
		Orientation:  style.OrientationPortrait,
		PageSizeW:    style.DefaultPageWidth,
		PageSizeH:    style.DefaultPageHeight,
		MarginTop:    1134, // 2 cm
		MarginBottom: 1134,
		MarginLeft:   1134,
		MarginRight:  1134,
		HeaderHeight: style.DefaultHeaderHeight,
		FooterHeight: style.DefaultFooterHeight,
	})

	addCoverAndParagraphs(sec)
	sec.AddPageBreak()
	addMergedTable(sec)
	sec.AddPageBreak()
	addStreamedSection(sec)
	sec.AddPageBreak()
	addCharts(doc, sec)
	sec.AddPageBreak()
	addRichText(doc, sec)
	sec.AddPageBreak()
	addTemplatePlaceholders(sec)
	sec.AddPageBreak()
	addCommentsAndRevisions(doc, sec)

	raw, err := doc.Bytes()
	if err != nil {
		log.Fatal(err)
	}
	tp, err := word.NewTemplateProcessorBytes(raw)
	if err != nil {
		log.Fatal(err)
	}
	if err := fillTemplate(tp); err != nil {
		log.Fatal(err)
	}
	if err := tp.Save(outFile); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("wrote %s\n", outFile)
}

func registerStyles(doc *word.Document) {
	doc.AddTitleStyle(1, style.Font{
		Bold: true, Size: 22, Color: "1F4E79", Name: "Microsoft YaHei",
	}, style.Paragraph{Spacing: style.Spacing{Before: 240, After: 160}})
	doc.AddTitleStyle(2, style.Font{
		Bold: true, Size: 16, Color: "2E75B6", Name: "Microsoft YaHei",
	}, style.Paragraph{Spacing: style.Spacing{Before: 200, After: 120}})
	doc.AddTitleStyle(3, style.Font{
		Bold: true, Size: 13, Color: "5B9BD5", Name: "Microsoft YaHei",
	}, style.Paragraph{Spacing: style.Spacing{Before: 160, After: 80}})
	doc.AddFontStyle("lead", style.Font{Size: 12, Color: "44546A", Italic: true})
	doc.AddParagraphStyle("center", style.Paragraph{Alignment: style.JcCenter})
}

func addCoverAndParagraphs(sec *element.Section) {
	sec.AddTitle("GoWord 全功能集成演示", 1)
	sec.AddText("v0.4.0 — Charts, Markdown/HTML, Comments, Track Changes, StreamWriter, TemplateProcessor", "lead", "center")
	sec.AddTextBreak()

	sec.AddTitle("1. 基础样式与段落", 1)
	sec.AddTitle("1.1 标题层级", 2)
	sec.AddTitle("第三级标题 Heading 3", 3)
	sec.AddText("本节设置了 2 cm 页边距，并注册 H1–H3 标题样式。下面用不同字号、颜色与强调展示段落格式。")

	sec.AddTitle("1.2 字体强调", 2)
	sec.AddText("常规正文 11pt Calibri / 微软雅黑。", style.Font{Size: 11})
	sec.AddText("18pt 大号标题色段落。", style.Font{Size: 18, Color: "1F4E79", Bold: true})
	sec.AddText("加粗段落。", style.Font{Bold: true, Size: 12})
	sec.AddText("斜体段落。", style.Font{Italic: true, Size: 12, Color: "C45911"})
	sec.AddText("单下划线段落。", style.Font{Underline: style.UnderlineSingle, Size: 12, Color: "548235"})
	sec.AddText("加粗 + 斜体 + 下划线。", style.Font{
		Bold: true, Italic: true, Underline: style.UnderlineSingle, Size: 12, Color: "833C0C",
	})

	tr := sec.AddTextRun()
	tr.AddText("同一段落内联：")
	tr.AddText("红", style.Font{Color: "C00000", Bold: true})
	tr.AddText(" / ")
	tr.AddText("绿", style.Font{Color: "548235", Italic: true})
	tr.AddText(" / ")
	tr.AddText("蓝下划线", style.Font{Color: "2E75B6", Underline: style.UnderlineSingle})
	tr.AddText("。")

	sec.AddLink("https://pkg.go.dev/github.com/yunkeweb/go-word@v0.4.0", "GoWord on pkg.go.dev")
}

func addMergedTable(sec *element.Section) {
	sec.AddTitle("2. 表格与单元格合并", 1)
	sec.AddText("复杂表头使用 gridSpan；区域列使用 vMerge；表头与数据行带自定义边框与背景色。")

	line := style.Border{Style: "single", Size: 8, Color: "1F4E79"}
	tbl := sec.AddTable(style.Table{
		Width:     9360,
		Alignment: style.JcTableCenter,
		Borders: style.Borders{
			Top: line, Left: line, Right: line, Bottom: line,
			InsideH: line, InsideV: line,
		},
	})

	titleFont := style.Font{Bold: true, Color: "FFFFFF", Size: 14, Name: "Microsoft YaHei"}
	headFont := style.Font{Bold: true, Color: "FFFFFF", Size: 11, Name: "Microsoft YaHei"}
	center := style.Paragraph{Alignment: style.JcCenter}

	r0 := tbl.AddRow(400)
	r0.AddCell(9360, style.Cell{
		GridSpan: 4, BgColor: "1F4E79", VAlign: style.VAlignCenter, Width: 9360,
	}).AddText("2026 季度销售总览", titleFont, center)

	r1 := tbl.AddRow(320)
	r1.AddCell(2340, style.Cell{
		BgColor: "2E75B6", VMerge: "restart", VAlign: style.VAlignCenter, Width: 2340,
	}).AddText("区域", headFont, center)
	r1.AddCell(4680, style.Cell{
		GridSpan: 2, BgColor: "2E75B6", VAlign: style.VAlignCenter, Width: 4680,
	}).AddText("2026 年", headFont, center)
	r1.AddCell(2340, style.Cell{
		BgColor: "2E75B6", VMerge: "restart", VAlign: style.VAlignCenter, Width: 2340,
	}).AddText("合计", headFont, center)

	r2 := tbl.AddRow(280)
	r2.AddCell(2340, style.Cell{VMerge: "continue", BgColor: "2E75B6", Width: 2340})
	r2.AddCell(2340, style.Cell{BgColor: "5B9BD5", VAlign: style.VAlignCenter, Width: 2340}).
		AddText("上半年", headFont, center)
	r2.AddCell(2340, style.Cell{BgColor: "5B9BD5", VAlign: style.VAlignCenter, Width: 2340}).
		AddText("下半年", headFont, center)
	r2.AddCell(2340, style.Cell{VMerge: "continue", BgColor: "2E75B6", Width: 2340})

	addSalesRow(tbl, "华东", "120", "148", "268", "D6EAF8")
	addSalesRow(tbl, "华南", "90", "110", "200", "FFFFFF")
	addSalesRow(tbl, "华北", "150", "166", "316", "D6EAF8")

	note := tbl.AddRow(280)
	note.AddCell(2340, style.Cell{BgColor: "FFF2CC", VAlign: style.VAlignCenter, Width: 2340}).
		AddText("备注", style.Font{Bold: true, Size: 10}, center)
	note.AddCell(7020, style.Cell{GridSpan: 3, BgColor: "FFF2CC", Width: 7020}).
		AddText("金额单位：万元。合计列为上半年 + 下半年。", style.Font{Size: 10, Italic: true})
}

func addSalesRow(tbl *element.Table, region, h1, h2, total, bg string) {
	center := style.Paragraph{Alignment: style.JcCenter}
	row := tbl.AddRow(280)
	row.AddCell(2340, style.Cell{BgColor: bg, VAlign: style.VAlignCenter, Width: 2340}).
		AddText(region, style.Font{Bold: true, Size: 11}, center)
	row.AddCell(2340, style.Cell{BgColor: bg, VAlign: style.VAlignCenter, Width: 2340}).
		AddText(h1, style.Font{Size: 11}, center)
	row.AddCell(2340, style.Cell{BgColor: bg, VAlign: style.VAlignCenter, Width: 2340}).
		AddText(h2, style.Font{Size: 11}, center)
	row.AddCell(2340, style.Cell{BgColor: bg, VAlign: style.VAlignCenter, Width: 2340}).
		AddText(total, style.Font{Bold: true, Size: 11}, center)
}

func addStreamedSection(sec *element.Section) {
	sec.AddTitle("3. 流式写入 StreamWriter", 1)
	sec.AddText("StreamWriter 按段落 / 表格行边生成边写入 ZIP 流，避免把整份 document.xml 缓存在内存中。下方表格与 main 中 demoStreamWriter() 使用同一套行数据。")
	tbl := sec.AddTable(style.Table{
		Width:     9000,
		Alignment: style.JcTableCenter,
		Borders: style.Borders{
			Top:     style.Border{Style: "single", Size: 4, Color: "8FAADC"},
			Left:    style.Border{Style: "single", Size: 4, Color: "8FAADC"},
			Right:   style.Border{Style: "single", Size: 4, Color: "8FAADC"},
			Bottom:  style.Border{Style: "single", Size: 4, Color: "8FAADC"},
			InsideH: style.Border{Style: "single", Size: 4, Color: "8FAADC"},
			InsideV: style.Border{Style: "single", Size: 4, Color: "8FAADC"},
		},
	})
	fillStreamTable(tbl)
}

func fillStreamTable(tbl *element.Table) {
	head := style.Font{Bold: true, Color: "FFFFFF", Size: 10}
	center := style.Paragraph{Alignment: style.JcCenter}
	hr := tbl.AddRow(280)
	for _, h := range []string{"序号", "批次", "数量", "状态"} {
		hr.AddCell(2250, style.Cell{BgColor: "2E75B6", VAlign: style.VAlignCenter, Width: 2250}).
			AddText(h, head, center)
	}
	for i := 1; i <= 24; i++ {
		bg := "FFFFFF"
		if i%2 == 0 {
			bg = "DEEBF7"
		}
		row := tbl.AddRow(240)
		row.AddCell(2250, style.Cell{BgColor: bg, Width: 2250}).AddText(fmt.Sprintf("%02d", i), nil, center)
		row.AddCell(2250, style.Cell{BgColor: bg, Width: 2250}).AddText(fmt.Sprintf("LOT-%03d", i), nil, center)
		row.AddCell(2250, style.Cell{BgColor: bg, Width: 2250}).AddText(fmt.Sprintf("%d", 10*i), nil, center)
		st := "OK"
		if i%7 == 0 {
			st = "HOLD"
		}
		row.AddCell(2250, style.Cell{BgColor: bg, Width: 2250}).AddText(st, nil, center)
	}
}

func demoStreamWriter() error {
	var buf bytes.Buffer
	sw := word.NewStreamWriter(&buf)
	if err := sw.WriteParagraph("StreamWriter incremental output", style.Font{Bold: true, Size: 14}); err != nil {
		return err
	}
	for i := 1; i <= 24; i++ {
		if err := sw.WriteParagraph(fmt.Sprintf("Streamed paragraph %02d", i), style.Font{Size: 10}); err != nil {
			return err
		}
	}
	tbl := element.NewTable(style.Table{Width: 9000})
	fillStreamTable(tbl)
	if err := sw.WriteTable(tbl); err != nil {
		return err
	}
	if err := sw.Close(); err != nil {
		return err
	}
	fmt.Printf("StreamWriter flushed %d bytes incrementally\n", sw.BytesWritten())
	return nil
}

func addCharts(doc *word.Document, sec *element.Section) {
	sec.AddTitle("4. 图表绘制 Charts & DrawingML", 1)
	sec.AddText("依次插入柱状图、折线图与饼图。包内生成 word/charts/chartN.xml，并注册 Content Types 与关系。")

	cats := []string{"华东", "华南", "华北"}
	chartSize := style.Chart{
		Width: 5486400, Height: 3200400, ShowLegend: true, ShowGridY: true,
	}

	sec.AddTitle("4.1 柱状图 Bar", 2)
	bar := style.Chart{Width: chartSize.Width, Height: chartSize.Height, ShowLegend: true, ShowGridY: true, Title: "季度收入（万元）"}
	doc.AddChartStyled(word.ChartTypeBar, cats, []float64{120, 90, 150}, bar)

	sec.AddTitle("4.2 折线图 Line", 2)
	line := style.Chart{Width: chartSize.Width, Height: chartSize.Height, ShowLegend: true, ShowGridY: true, Title: "下半年走势"}
	doc.AddChartStyled(word.ChartTypeLine, cats, []float64{148, 110, 166}, line)

	sec.AddTitle("4.3 饼图 Pie", 2)
	pie := style.Chart{Width: chartSize.Width, Height: chartSize.Height, ShowLegend: true, Title: "区域占比"}
	doc.AddChartStyled(word.ChartTypePie, cats, []float64{268, 200, 316}, pie)
}

func addRichText(doc *word.Document, sec *element.Section) {
	sec.AddTitle("5. 富文本导入 Markdown / HTML", 1)

	sec.AddTitle("5.1 Markdown", 2)
	md := "" +
		"### Markdown 示例\n\n" +
		"这段由 `AddMarkdown` 解析：**加粗**、*斜体*、++下划线++ 与 `inline code`。\n\n" +
		"- 无序列表甲\n" +
		"- 无序列表乙\n\n" +
		"1. 有序第一步\n" +
		"2. 有序第二步\n\n" +
		"```\nfunc Hello() string {\n    return \"GoWord\"\n}\n```\n"
	if err := doc.AddMarkdown(md); err != nil {
		log.Fatal(err)
	}

	sec.AddTitle("5.2 HTML", 2)
	html := `<h3>HTML 示例</h3>` +
		`<p>这段由 <b>AddHTML</b> 解析：<strong>加粗</strong>、<em>斜体</em>、<u>下划线</u> 与 <code>code</code>。</p>` +
		`<ul><li>HTML 列表一项</li><li>HTML 列表二项</li></ul>` +
		"<pre>SELECT id, name\nFROM products\nWHERE qty &gt; 0;</pre>" +
		`<blockquote>引用块：页边缩进。</blockquote>`
	if err := doc.AddHTML(html); err != nil {
		log.Fatal(err)
	}
}

func addTemplatePlaceholders(sec *element.Section) {
	sec.AddTitle("6. 模板引擎 TemplateProcessor", 1)
	sec.AddText("下面的嵌套块 block_a / cities，以及文档级条件 show / hide，会在保存前由 TemplateProcessor 替换。")

	sec.AddText("${block_a}")
	sec.AddText("区域：${region}")
	sec.AddText("${cities}")
	sec.AddText("城市 ${city}，人口 ${pop}")
	sec.AddText("${/cities}")
	sec.AddText("${if featured}")
	sec.AddText("★ 本区域为重点市场。")
	sec.AddText("${endif}")
	sec.AddText("${/block_a}")

	sec.AddText("${if show}")
	sec.AddText("条件块 show=true：本段保留在成品文档中。")
	sec.AddText("${endif}")
	sec.AddText("${if hide}")
	sec.AddText("条件块 hide=false：本段会在渲染时删除。")
	sec.AddText("${endif}")
}

func fillTemplate(tp *word.TemplateProcessor) error {
	if err := tp.CloneNestedBlock("block_a", []word.BlockData{
		{
			Values: map[string]string{"region": "华东"},
			Blocks: map[string][]word.BlockData{
				"cities": {
					{Values: map[string]string{"city": "上海", "pop": "2487 万"}},
					{Values: map[string]string{"city": "杭州", "pop": "1238 万"}},
				},
			},
			If: map[string]bool{"featured": true},
		},
		{
			Values: map[string]string{"region": "华南"},
			Blocks: map[string][]word.BlockData{
				"cities": {
					{Values: map[string]string{"city": "广州", "pop": "1873 万"}},
				},
			},
			If: map[string]bool{"featured": false},
		},
	}); err != nil {
		return err
	}
	if err := tp.SetCondition("show", true); err != nil {
		return err
	}
	if err := tp.SetCondition("hide", false); err != nil {
		return err
	}
	return nil
}

func addCommentsAndRevisions(doc *word.Document, sec *element.Section) {
	sec.AddTitle("7. 批注与修订模式", 1)
	sec.AddText("开启 w:trackRevisions，对指定文本添加批注，并用插入 / 删除标记演示修订线。")

	doc.EnableTrackChanges(true)
	doc.CommentOn(
		"请重点核对华东合计 268 万元。",
		"合计是否含税？建议在备注行写明口径。",
		"Ann",
		"A",
		"2026-09-26T10:00:00Z",
	)
	sec.AddText("修订演示：")
	doc.AddInsertion("新增：华南下半年上调 8%。", "Bob", "2026-09-26T10:05:00Z", style.Font{Color: "006600"})
	doc.AddDeletion("旧稿：华北全年目标 280 万元。", "Bob", "2026-09-26T10:06:00Z", style.Font{Color: "C00000"})
	sec.AddText("成品文档中可在 Word 里显示标记以查看插入线与删除线。")
}
