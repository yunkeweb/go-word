// Command word_diag writes six single-module .docx files so Microsoft Word
// can isolate which OpenXML feature refuses to open.
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/yunkeweb/go-word"
	"github.com/yunkeweb/go-word/element"
	"github.com/yunkeweb/go-word/style"
)

func main() {
	gens := []struct {
		name string
		fn   func(string) error
	}{
		{"diag_a_paragraphs_tables.docx", genParagraphsTables},
		{"diag_b_streamwriter.docx", genStreamWriter},
		{"diag_c_charts.docx", genCharts},
		{"diag_d_markdown_html.docx", genMarkdownHTML},
		{"diag_e_template.docx", genTemplate},
		{"diag_f_comments_revisions.docx", genCommentsRevisions},
		{"diag_combo.docx", genComboChart},
		{"diag_area.docx", genAreaChart},
		{"diag_template.docx", genTemplateFilters},
	}
	for _, g := range gens {
		if err := g.fn(g.name); err != nil {
			log.Fatalf("%s: %v", g.name, err)
		}
		fmt.Printf("wrote %s\n", g.name)
	}
	fmt.Println("Open each file in Microsoft Word. The module that Word refuses is the one to fix.")
}

func genParagraphsTables(path string) error {
	doc := word.New()
	info := doc.GetDocInfo()
	info.Title = "diag A paragraphs tables"
	info.Creator = "GoWord"
	sec := doc.AddSection(style.Section{
		Orientation: style.OrientationPortrait,
		PageSizeW:   style.DefaultPageWidth,
		PageSizeH:   style.DefaultPageHeight,
		MarginTop:   1134, MarginBottom: 1134, MarginLeft: 1134, MarginRight: 1134,
	})
	doc.AddTitleStyle(1, style.Font{Bold: true, Size: 18, Color: "1F4E79"}, style.Paragraph{})
	sec.AddTitle("Paragraphs and tables", 1)
	sec.AddText("Plain paragraph.")
	sec.AddText("Bold italic underline.", style.Font{
		Bold: true, Italic: true, Underline: style.UnderlineSingle, Size: 12, Color: "833C0C",
	})
	tr := sec.AddTextRun()
	tr.AddText("inline ")
	tr.AddText("red", style.Font{Color: "C00000", Bold: true})
	line := style.Border{Style: "single", Size: 8, Color: "1F4E79"}
	tbl := sec.AddTable(style.Table{
		Width: 9360, Alignment: style.JcTableCenter,
		Borders: style.Borders{Top: line, Left: line, Right: line, Bottom: line, InsideH: line, InsideV: line},
	})
	center := style.Paragraph{Alignment: style.JcCenter}
	r0 := tbl.AddRow(400)
	r0.AddCell(9360, style.Cell{GridSpan: 4, BgColor: "1F4E79", VAlign: style.VAlignCenter, Width: 9360}).
		AddText("Title spanning four columns", style.Font{Bold: true, Color: "FFFFFF"}, center)
	r1 := tbl.AddRow(280)
	r1.AddCell(2340, style.Cell{BgColor: "2E75B6", VMerge: "restart", Width: 2340}).AddText("A", nil, center)
	r1.AddCell(2340, style.Cell{BgColor: "2E75B6", Width: 2340}).AddText("B", nil, center)
	r1.AddCell(2340, style.Cell{BgColor: "2E75B6", Width: 2340}).AddText("C", nil, center)
	r1.AddCell(2340, style.Cell{BgColor: "2E75B6", VMerge: "restart", Width: 2340}).AddText("D", nil, center)
	r2 := tbl.AddRow(280)
	r2.AddCell(2340, style.Cell{VMerge: "continue", BgColor: "2E75B6", Width: 2340})
	r2.AddCell(2340, style.Cell{Width: 2340}).AddText("1", nil, center)
	r2.AddCell(2340, style.Cell{Width: 2340}).AddText("2", nil, center)
	r2.AddCell(2340, style.Cell{VMerge: "continue", BgColor: "2E75B6", Width: 2340})
	return doc.Save(path)
}

func genStreamWriter(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	sw := word.NewStreamWriter(f)
	if err := sw.WriteParagraph("StreamWriter diagnostic", style.Font{Bold: true, Size: 14}); err != nil {
		return err
	}
	for i := 1; i <= 8; i++ {
		if err := sw.WriteParagraph(fmt.Sprintf("row %02d", i), style.Font{Size: 11}); err != nil {
			return err
		}
	}
	line := style.Border{Style: "single", Size: 4, Color: "8FAADC"}
	tbl := element.NewTable(style.Table{
		Width:   9000,
		Borders: style.Borders{Top: line, Left: line, Right: line, Bottom: line, InsideH: line, InsideV: line},
	})
	hr := tbl.AddRow(280)
	for _, h := range []string{"No", "Name", "Qty", "Status"} {
		hr.AddCell(2250, style.Cell{BgColor: "2E75B6", Width: 2250}).
			AddText(h, style.Font{Bold: true, Color: "FFFFFF"})
	}
	for i := 1; i <= 12; i++ {
		row := tbl.AddRow(240)
		row.AddCell(2250, style.Cell{Width: 2250}).AddText(fmt.Sprintf("%02d", i))
		row.AddCell(2250, style.Cell{Width: 2250}).AddText(fmt.Sprintf("LOT-%03d", i))
		row.AddCell(2250, style.Cell{Width: 2250}).AddText(fmt.Sprintf("%d", 10*i))
		row.AddCell(2250, style.Cell{Width: 2250}).AddText("OK")
	}
	if err := sw.WriteTable(tbl); err != nil {
		return err
	}
	return sw.Close()
}

func genCharts(path string) error {
	doc := word.New()
	sec := doc.AddSection()
	sec.AddText("Charts diagnostic")
	cats := []string{"East", "South", "North"}
	sz := style.Chart{Width: 5486400, Height: 3200400, ShowLegend: true, ShowGridY: true}
	bar := sz
	bar.Title = "Bar"
	doc.AddChartStyled(word.ChartTypeBar, cats, []float64{120, 90, 150}, bar)
	line := sz
	line.Title = "Line"
	doc.AddChartStyled(word.ChartTypeLine, cats, []float64{148, 110, 166}, line)
	pie := sz
	pie.Title = "Pie"
	doc.AddChartStyled(word.ChartTypePie, cats, []float64{268, 200, 316}, pie)
	return doc.Save(path)
}

func genMarkdownHTML(path string) error {
	doc := word.New()
	_ = doc.AddSection()
	if err := doc.AddMarkdown("### Markdown\n\n**bold** *italic* ++under++ `code`\n\n- one\n- two\n"); err != nil {
		return err
	}
	if err := doc.AddHTML(`<h3>HTML</h3><p><strong>bold</strong> <em>italic</em> <u>under</u></p><ul><li>item</li></ul>`); err != nil {
		return err
	}
	return doc.Save(path)
}

func genTemplate(path string) error {
	doc := word.New()
	sec := doc.AddSection()
	sec.AddText("${block_a}")
	sec.AddText("region ${region}")
	sec.AddText("${cities}")
	sec.AddText("city ${city}")
	sec.AddText("${/cities}")
	sec.AddText("${if featured}")
	sec.AddText("featured")
	sec.AddText("${endif}")
	sec.AddText("${/block_a}")
	sec.AddText("${if show}")
	sec.AddText("shown")
	sec.AddText("${endif}")
	sec.AddText("${if hide}")
	sec.AddText("hidden")
	sec.AddText("${endif}")
	raw, err := doc.Bytes()
	if err != nil {
		return err
	}
	tp, err := word.NewTemplateProcessorBytes(raw)
	if err != nil {
		return err
	}
	if err := tp.CloneNestedBlock("block_a", []word.BlockData{
		{
			Values: map[string]string{"region": "East"},
			Blocks: map[string][]word.BlockData{
				"cities": {{Values: map[string]string{"city": "Shanghai"}}},
			},
			If: map[string]bool{"featured": true},
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
	return tp.Save(path)
}

func genComboChart(path string) error {
	doc := word.New()
	sec := doc.AddSection()
	sec.AddText("Combo chart diagnostic")
	cats := []string{"Jan", "Feb", "Mar", "Apr"}
	ch := doc.AddChart(word.ChartTypeCombo, cats, []float64{120, 150, 140, 180})
	ch.Series[0].Name = "Sales"
	ch.Series[0].Kind = word.ChartTypeColumn
	ch.AddComboSeries(word.ChartTypeLine, cats, []float64{8.2, 9.1, 8.7, 10.4}, "Margin %", true)
	ch.Style.Title = "Sales vs Margin"
	ch.Style.Width = 5486400
	ch.Style.Height = 3200400
	ch.Style.ValueNumFmt = "0"
	ch.Style.SecondaryValueNumFmt = "0.0"
	ch.SetLegendPosition(word.LegendRight)
	ch.SetMajorGridlines(true)
	ch.SetLineSmooth(true)
	ch.SetLineMarker("circle")
	ch.SetDataLabels(word.ChartDataLabelOptions{ShowVal: true, Position: word.DataLabelPosTop})
	return doc.Save(path)
}

func genAreaChart(path string) error {
	doc := word.New()
	sec := doc.AddSection()
	sec.AddText("Area chart diagnostic")
	ch := doc.AddChart(word.ChartTypeStackedArea, []string{"Q1", "Q2", "Q3", "Q4"}, []float64{12, 18, 15, 22}, []float64{8, 11, 13, 16})
	ch.Series[0].Name = "Hardware"
	ch.Series[1].Name = "Software"
	ch.Style.Title = "Quarterly Revenue Mix"
	ch.Style.Width = 5486400
	ch.Style.Height = 3200400
	ch.SetLegendPosition(word.LegendBottom)
	ch.SetMajorGridlines(true)
	ch.SetDataLabels(word.ChartDataLabelOptions{
		ShowVal:  true,
		Position: word.DataLabelPosCenter,
	})
	return doc.Save(path)
}

func genTemplateFilters(path string) error {
	src := word.New()
	sec := src.AddSection()
	sec.AddText("When ${created_at | formatDate:\"2006-01-02\"}")
	sec.AddText("Hello ${name | upper}")
	sec.AddText("City ${city | lower}")
	sec.AddText("Title ${title | trim | upper}")
	sec.AddText("Blurb ${blurb | truncate:12}")
	sec.AddText("Amount ${amount | formatCurrency:¥}")
	sec.AddText("Empty ${missing | default:N/A}")
	sec.AddText("${if age >= 18}Adult content is visible.${endif}")
	raw, err := src.Bytes()
	if err != nil {
		return err
	}
	tp, err := word.NewTemplateProcessorBytes(raw)
	if err != nil {
		return err
	}
	tp.SetValue("created_at", "2026-09-26T08:00:00Z")
	tp.SetValue("name", "alice")
	tp.SetValue("city", "Shanghai")
	tp.SetValue("title", "  go-word  ")
	tp.SetValue("blurb", "Streaming parser, charts and filters")
	tp.SetValue("amount", "1999.5")
	tp.SetValue("age", "21")
	return tp.Save(path)
}

func genCommentsRevisions(path string) error {
	doc := word.New()
	sec := doc.AddSection()
	sec.AddText("Comments and track changes")
	doc.EnableTrackChanges(true)
	doc.CommentOn("please review this total.", "Does this include tax?", "Ann", "A", "2026-09-26T10:00:00Z")
	doc.AddInsertion("inserted clause.", "Bob", "2026-09-26T10:05:00Z", style.Font{Color: "006600"})
	doc.AddDeletion("removed clause.", "Bob", "2026-09-26T10:06:00Z", style.Font{Color: "C00000"})
	return doc.Save(path)
}
