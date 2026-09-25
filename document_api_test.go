package word

import (
	"strings"
	"testing"

	"github.com/yunkeweb/go-word/element"
	"github.com/yunkeweb/go-word/style"
)

func TestStyleRegistryAndCollections(t *testing.T) {
	doc := New()
	doc.AddFontStyle("rStyle", style.Font{Bold: true})
	doc.AddParagraphStyle("pStyle", style.Paragraph{Alignment: style.JcCenter})
	if doc.CountStyles() != 2 {
		t.Fatalf("CountStyles=%d", doc.CountStyles())
	}
	if doc.GetStyle("rStyle") == nil || doc.GetStyle("rStyle").Font == nil || !doc.GetStyle("rStyle").Font.Bold {
		t.Fatal("GetStyle rStyle")
	}
	if doc.GetStyle("missing") != nil {
		t.Fatal("missing style")
	}

	sec := doc.AddSection()
	sec.AddTitle("Hello", 1)
	sec.AddBookmark("bm")
	sec.AddFootnote()
	sec.AddChart("pie", []string{"A"}, []float64{1})
	if len(doc.GetTitles()) != 1 || doc.GetTitles()[0].Text != "Hello" {
		t.Fatal("GetTitles")
	}
	if len(doc.GetBookmarks()) != 1 {
		t.Fatal("GetBookmarks")
	}
	if len(doc.GetFootnotes()) != 1 {
		t.Fatal("GetFootnotes")
	}
	if len(doc.GetCharts()) != 1 {
		t.Fatal("GetCharts")
	}
}

func TestContainerElementCRUD(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	sec.AddText("a")
	sec.AddText("b")
	sec.AddText("c")
	if sec.CountElements() != 3 {
		t.Fatalf("count=%d", sec.CountElements())
	}
	if tx, ok := sec.GetElement(1).(*element.Text); !ok || tx.Content != "b" {
		t.Fatal("GetElement")
	}
	sec.RemoveElement(1)
	if sec.CountElements() != 2 {
		t.Fatal("RemoveElement index")
	}
	sec.RemoveElement(sec.GetElement(0))
	if sec.CountElements() != 1 {
		t.Fatal("RemoveElement identity")
	}
}

func TestTableHelpersAndWatermark(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	tbl := sec.AddTable()
	tbl.SetWidth(5000)
	r := tbl.AddRow(300)
	r.AddCell(1000)
	r.AddCell(2000)
	if tbl.CountColumns() != 2 || tbl.GetWidth() != 5000 {
		t.Fatal("table width/cols")
	}
	if w := tbl.FindFirstDefinedCellWidths(); len(w) != 2 || w[1] != 2000 {
		t.Fatalf("widths=%v", w)
	}
	if r.GetHeight() != 300 {
		t.Fatal("row height")
	}
	h := sec.AddHeader(element.HeaderFirst)
	h.AddWatermark("mark.png")
	if !sec.HasDifferentFirstPage() {
		t.Fatal("first page header")
	}
	if !h.Elements()[0].(*element.Image).IsWatermark {
		t.Fatal("watermark flag")
	}
}

func TestAddHTML(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	html := `<p>Hello <strong>World</strong></p><h1>Title</h1><ul><li>one</li></ul>`
	if err := AddHTML(sec, html); err != nil {
		t.Fatal(err)
	}
	var texts []string
	for _, e := range sec.Elements() {
		switch v := e.(type) {
		case *element.TextRun:
			texts = append(texts, v.GetText())
		case *element.Title:
			texts = append(texts, v.Text)
		case *element.ListItem:
			texts = append(texts, v.Text)
		}
	}
	joined := strings.Join(texts, "|")
	if !strings.Contains(joined, "Hello") || !strings.Contains(joined, "World") {
		t.Fatalf("html text=%q", joined)
	}
	if !strings.Contains(joined, "Title") || !strings.Contains(joined, "one") {
		t.Fatalf("html structure=%q", joined)
	}
}

func TestSettingsExtras(t *testing.T) {
	SetDefaultPaper("Letter")
	if DefaultPaper() != "Letter" {
		t.Fatal("paper")
	}
	SetMeasurementUnit(UnitCM)
	if MeasurementUnit() != UnitCM {
		t.Fatal("unit")
	}
	SetCompatibility(false)
	if HasCompatibility() {
		t.Fatal("compat")
	}
	SetCompatibility(true)
	SetDefaultPaper("A4")
	SetMeasurementUnit(UnitTwip)
}

func TestPaperSize(t *testing.T) {
	p := style.NewPaper("Letter")
	if p.Width == 0 || p.Height == 0 {
		t.Fatal("letter geometry")
	}
	sec := style.NewSection()
	p.ApplyToSection(&sec)
	if sec.PaperSize != "Letter" {
		t.Fatal("apply")
	}
}

func TestChartAddSeries(t *testing.T) {
	ch := &element.Chart{ChartType: "bar"}
	ch.AddSeries([]string{"A", "B"}, []float64{1, 2}, "s1")
	if len(ch.Series) != 1 || ch.SeriesName != "s1" {
		t.Fatal("series")
	}
}

func TestResetStylesAndHeaderTypes(t *testing.T) {
	doc := New()
	doc.AddFontStyle("r", style.Font{Bold: true})
	if doc.CountStyles() != 1 {
		t.Fatal("count")
	}
	doc.ResetStyles()
	if doc.CountStyles() != 0 {
		t.Fatal("reset")
	}
	sec := doc.AddSection()
	h := sec.AddHeader()
	if h.FirstPage() != element.HeaderFirst || h.GetType() != element.HeaderFirst {
		t.Fatal("first")
	}
	if h.EvenPage() != element.HeaderEven {
		t.Fatal("even")
	}
	if h.ResetType() != element.HeaderAuto {
		t.Fatal("reset type")
	}
	f := sec.AddFooter()
	f.SetType(element.HeaderFirst)
	if f.GetType() != element.HeaderFirst {
		t.Fatal("footer type")
	}
}

func TestChartAndCommentParts(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	sec.AddChart("pie", []string{"A", "B"}, []float64{3, 1})
	c := sec.AddComment("Ann", "A", "2020-01-01T00:00:00Z")
	c.AddText("note")
	data, err := doc.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	s := string(data)
	if !strings.Contains(s, "word/charts/chart1.xml") {
		t.Fatal("chart part missing")
	}
	if !strings.Contains(s, "word/comments.xml") {
		t.Fatal("comments part missing")
	}
}
