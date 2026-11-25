package word

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/yunkeweb/go-word/element"
	"github.com/yunkeweb/go-word/style"
)

const drawingNS = ` xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main" xmlns:wp="http://schemas.openxmlformats.org/drawingml/2006/wordprocessingDrawing" xmlns:wps="http://schemas.microsoft.com/office/word/2010/wordprocessingShape" xmlns:c="http://schemas.openxmlformats.org/drawingml/2006/chart" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships"`

func complexDrawing(content string) string {
	return `<w:r><w:drawing` + drawingNS + `><wp:inline><wp:extent cx="952500" cy="476250"/><a:graphic><a:graphicData>` + content + `</a:graphicData></a:graphic></wp:inline></w:drawing></w:r>`
}

func TestDOCXComplexObjectBaseline(t *testing.T) {
	shape := `<wps:wsp><wps:spPr><a:xfrm rot="5400000" flipH="1"/><a:prstGeom prst="roundRect"/><a:solidFill><a:srgbClr val="D9EAF7"/></a:solidFill><a:ln w="12700"><a:solidFill><a:srgbClr val="205493"/></a:solidFill></a:ln></wps:spPr><wps:txbx><w:txbxContent><w:p><w:r><w:rPr><w:b/></w:rPr><w:t>shape &amp; text</w:t></w:r></w:p><w:p><w:r><w:t>second paragraph</w:t></w:r></w:p></w:txbxContent></wps:txbx></wps:wsp>`
	raw := contentDOCXFixture(t, `<w:p><w:r><w:t>before</w:t></w:r>`+complexDrawing(shape)+`<w:r><w:t>after</w:t></w:r></w:p>`, "", nil)
	doc, err := LoadBytes(raw)
	if err != nil {
		t.Fatal(err)
	}
	var sh *element.DMLShape
	walkDocument(doc, func(el element.Element) {
		if v, ok := el.(*element.DMLShape); ok {
			sh = v
		}
	})
	if sh == nil {
		t.Fatal("DrawingML shape missing from DOM")
	}
	if sh.Width != 952500 || sh.Height != 476250 || sh.Rotation != 90 || !sh.FlipH || sh.FillColor != "D9EAF7" || sh.LineColor != "205493" || sh.LineWidth != 12700 {
		t.Fatalf("shape properties: %+v", sh)
	}
	got, err := doc.RenderHTML(HTMLOptions{Strict: true})
	if err != nil {
		t.Fatal(err)
	}
	s := string(got)
	for _, want := range []string{"before", "shape &amp; text", "second paragraph", "after", "font-weight:700", "width:100px", "border:1pt solid #205493"} {
		if !strings.Contains(s, want) {
			t.Errorf("missing %q: %s", want, s)
		}
	}
	if !(strings.Index(s, "before") < strings.Index(s, "shape &amp;") && strings.Index(s, "second paragraph") < strings.Index(s, "after")) {
		t.Fatal("drawing reordered body text")
	}
	assertHTMLStructure(t, s)
}

func TestReaderDrawingMLRoundTrip(t *testing.T) {
	doc := New()
	sh := doc.AddSection().AddDMLShape("roundRect", 952500, 476250, "D9EAF7", "205493", 12700)
	sh.Rotation = 30
	sh.FlipV = true
	sh.AddText("preserved", style.Font{Bold: true})
	for i := 0; i < 2; i++ {
		doc = cycleDocument(t, doc)
		var loaded *element.DMLShape
		walkDocument(doc, func(el element.Element) {
			if v, ok := el.(*element.DMLShape); ok {
				loaded = v
			}
		})
		if loaded == nil || loaded.Width != sh.Width || loaded.Rotation != 30 || !loaded.FlipV {
			t.Fatalf("cycle %d: %+v", i, loaded)
		}
	}
}

const cachedChart = `<c:chartSpace xmlns:c="http://schemas.openxmlformats.org/drawingml/2006/chart" xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main"><c:chart><c:title><c:tx><c:rich><a:p><a:r><a:t>Sales &amp; costs</a:t></a:r></a:p></c:rich></c:tx></c:title><c:plotArea><c:barChart><c:ser><c:tx><c:strRef><c:strCache><c:pt idx="0"><c:v>Revenue</c:v></c:pt></c:strCache></c:strRef></c:tx><c:cat><c:strRef><c:strCache><c:ptCount val="3"/><c:pt idx="2"><c:v>Q3</c:v></c:pt><c:pt idx="0"><c:v>Q1</c:v></c:pt><c:pt idx="1"><c:v>Q2</c:v></c:pt></c:strCache></c:strRef></c:cat><c:val><c:numRef><c:numCache><c:ptCount val="3"/><c:pt idx="2"><c:v>30.5</c:v></c:pt><c:pt idx="0"><c:v>0</c:v></c:pt></c:numCache></c:numRef></c:val></c:ser><c:axId val="1"/><c:axId val="2"/></c:barChart><c:lineChart><c:ser><c:tx><c:v>Costs</c:v></c:tx><c:val><c:numLit><c:pt idx="0"><c:v>5</c:v></c:pt></c:numLit></c:val></c:ser><c:axId val="3"/><c:axId val="4"/></c:lineChart><c:catAx><c:title><c:tx><c:rich><a:p><a:r><a:t>Quarter</a:t></a:r></a:p></c:rich></c:tx></c:title></c:catAx><c:valAx><c:title><c:tx><c:rich><a:p><a:r><a:t>USD</a:t></a:r></a:p></c:rich></c:tx></c:title></c:valAx></c:plotArea><c:legend><c:legendPos val="b"/></c:legend></c:chart></c:chartSpace>`

func TestReaderChartCache(t *testing.T) {
	body := `<w:p><w:r><w:t>before</w:t></w:r>` + complexDrawing(`<c:chart r:id="chart1"/>`) + `<w:r><w:t>after</w:t></w:r></w:p>`
	rel := `<Relationship Id="chart1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/chart" Target="charts/chart1.xml"/>`
	raw := contentDOCXFixture(t, body, rel, map[string]string{"word/charts/chart1.xml": cachedChart})
	doc, err := LoadBytes(raw)
	if err != nil {
		t.Fatal(err)
	}
	var ch *element.Chart
	walkDocument(doc, func(el element.Element) {
		if v, ok := el.(*element.Chart); ok {
			ch = v
		}
	})
	if ch == nil || len(ch.Series) != 2 {
		t.Fatalf("chart=%+v", ch)
	}
	if ch.Style.Title != "Sales & costs" || ch.Series[0].Name != "Revenue" || ch.Categories[1] != "Q2" || len(ch.Series[0].Values) != 3 || ch.Series[0].Values[2] != 30.5 || !ch.Series[0].MissingValues[1] || ch.Series[0].MissingValues[0] || !ch.Series[1].SecondaryAxis || ch.Series[1].Kind != "line" || ch.Style.CategoryAxisTitle != "Quarter" || ch.Style.ValueAxisTitle != "USD" || !ch.Style.ShowLegend {
		t.Fatalf("chart data=%+v series=%+v", ch, ch.Series)
	}
	result, err := doc.RenderHTMLWithDiagnostics(HTMLOptions{})
	if err != nil {
		t.Fatal(err)
	}
	s := string(result.HTML)
	for _, want := range []string{"Sales &amp; costs", "Revenue", "Costs", "Q2", "30.5", "scope=\"col\"", "scope=\"row\"", "data-secondary-axis=\"true\""} {
		if !strings.Contains(s, want) {
			t.Errorf("missing %q: %s", want, s)
		}
	}
	if strings.Contains(s, "<p><figure") {
		t.Fatal("invalid block figure inside paragraph")
	}
	assertHTMLStructure(t, s)
	if len(result.Diagnostics) == 0 || result.Diagnostics[0].Part != "word/charts/chart1.xml" || result.Diagnostics[0].ID == "" {
		t.Fatalf("missing cache source diagnostic: %+v", result.Diagnostics)
	}
	if _, err = doc.RenderHTML(HTMLOptions{Strict: true}); err == nil {
		t.Fatal("strict mode accepted missing chart cache values")
	}
	if _, err = LoadBytesWithOptions(raw, ReadOptions{MaxPartSize: int64(len(cachedChart) - 1)}); !errors.Is(err, ErrReadLimitExceeded) {
		t.Fatalf("chart part exceeded read budget: %v", err)
	}
}

func TestReaderTextBoxAndUnknownDrawing(t *testing.T) {
	for _, content := range []string{`<wps:wsp><wps:cNvSpPr txBox="1"/><wps:spPr><a:prstGeom prst="rect"/></wps:spPr><wps:txbx><w:txbxContent><w:p><w:r><w:t>textbox</w:t></w:r></w:p></w:txbxContent></wps:txbx></wps:wsp>`, `<a:unknown><a:t>unknown text</a:t></a:unknown>`, `<a:unknown/>`} {
		raw := contentDOCXFixture(t, `<w:p>`+complexDrawing(content)+`<w:r><w:t>after</w:t></w:r></w:p>`, "", nil)
		doc, err := LoadBytes(raw)
		if err != nil {
			t.Fatal(err)
		}
		result, err := doc.RenderHTMLWithDiagnostics(HTMLOptions{})
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Contains(result.HTML, []byte("after")) {
			t.Fatal("drawing consumed subsequent text")
		}
		if strings.Contains(content, "textbox") {
			if !bytes.Contains(result.HTML, []byte("textbox")) {
				t.Fatal("textbox lost")
			}
		} else if len(result.Diagnostics) == 0 {
			t.Fatal("unknown drawing not diagnosed")
		}
	}
}

func TestReaderTableGridAndMergedCells(t *testing.T) {
	body := `<w:tbl><w:tblPr><w:tblW w:w="6000" w:type="dxa"/><w:tblLayout w:type="fixed"/><w:tblCellSpacing w:w="40" w:type="dxa"/><w:tblCellMar><w:top w:w="120" w:type="dxa"/><w:left w:w="80" w:type="dxa"/></w:tblCellMar><w:tblBorders><w:top w:val="single" w:sz="8" w:color="112233"/></w:tblBorders></w:tblPr><w:tblGrid><w:gridCol w:w="2000"/><w:gridCol w:w="4000"/></w:tblGrid><w:tr><w:trPr><w:tblHeader/><w:cantSplit/><w:trHeight w:val="360" w:hRule="exact"/></w:trPr><w:tc><w:tcPr><w:gridSpan w:val="2"/><w:vMerge w:val="restart"/><w:shd w:fill="FFEEDD"/><w:tcMar><w:bottom w:w="100" w:type="dxa"/></w:tcMar><w:tcBorders><w:left w:val="double" w:sz="12" w:color="445566"/></w:tcBorders></w:tcPr><w:p><w:r><w:t>merged root</w:t></w:r></w:p></w:tc></w:tr><w:tr><w:tc><w:tcPr><w:gridSpan w:val="2"/><w:vMerge/></w:tcPr><w:p/></w:tc></w:tr><w:tr><w:tc><w:tcPr><w:tcW w:w="2000" w:type="dxa"/></w:tcPr><w:p><w:r><w:t>left</w:t></w:r></w:p><w:tbl><w:tblGrid><w:gridCol w:w="500"/></w:tblGrid><w:tr><w:tc><w:p><w:r><w:t>nested</w:t></w:r></w:p></w:tc></w:tr></w:tbl></w:tc><w:tc><w:tcPr><w:tcW w:w="4000" w:type="dxa"/></w:tcPr><w:p><w:r><w:t>right</w:t></w:r></w:p></w:tc></w:tr></w:tbl>`
	raw := contentDOCXFixture(t, body, "", nil)
	doc, err := LoadBytes(raw)
	if err != nil {
		t.Fatal(err)
	}
	var tables []*element.Table
	walkDocument(doc, func(el element.Element) {
		if table, ok := el.(*element.Table); ok {
			tables = append(tables, table)
		}
	})
	if len(tables) != 2 {
		t.Fatalf("got %d tables, want outer and nested tables", len(tables))
	}
	outer := tables[0]
	if outer.Style.Width != 6000 || outer.Style.Unit != "dxa" || outer.Style.Layout != "fixed" {
		t.Errorf("outer table properties not retained: %+v", outer.Style)
	}
	if outer.Style.CellMarginTop != 120 || outer.Style.CellMarginLeft != 80 || outer.Style.Borders.Top.Color != "112233" || outer.Style.CellSpacingVal != 40 {
		t.Errorf("table margins/borders not retained: %+v", outer.Style)
	}
	if len(outer.Style.GridWidths) != 2 || outer.Style.GridWidths[0] != 2000 || outer.Style.GridWidths[1] != 4000 {
		t.Errorf("table grid not retained: %v", outer.Style.GridWidths)
	}
	if len(outer.Rows) != 3 || len(outer.Rows[0].Cells) != 1 || outer.Rows[0].Cells[0].Style.GridSpan != 2 || outer.Rows[0].Cells[0].Style.VMerge != "restart" {
		t.Fatalf("merged root not retained: %+v", outer.Rows)
	}
	if !outer.Rows[0].IsHeader() || !outer.Rows[0].IsCantSplit() || outer.Rows[0].Style.Height != 360 || outer.Rows[0].Style.Rule != "exact" {
		t.Errorf("row properties not retained: %+v", outer.Rows[0].Style)
	}
	rootCell := outer.Rows[0].Cells[0].Style
	if rootCell.Shading.Fill != "FFEEDD" || rootCell.PaddingTop != 0 || rootCell.PaddingBottom != 100 || rootCell.Borders.Left.Style != "double" || rootCell.Borders.Left.Color != "445566" {
		t.Errorf("cell style not retained: %+v", rootCell)
	}
	if len(outer.Rows[1].Cells) != 1 || outer.Rows[1].Cells[0].Style.GridSpan != 2 || outer.Rows[1].Cells[0].Style.VMerge != "continue" {
		t.Fatalf("empty vMerge must mean continue: %+v", outer.Rows[1].Cells)
	}
	doc = cycleDocument(t, doc)
	tables = nil
	walkDocument(doc, func(el element.Element) {
		if table, ok := el.(*element.Table); ok {
			tables = append(tables, table)
		}
	})
	if len(tables) != 2 || len(tables[0].Style.GridWidths) != 2 || tables[0].Style.GridWidths[0] != 2000 || tables[0].Style.CellSpacingVal != 40 || tables[0].Rows[1].Cells[0].Style.VMerge != "continue" {
		t.Fatalf("table grid/merge was not preserved across DOCX round trip: %+v", tables)
	}
	result, err := doc.RenderHTMLWithDiagnostics(HTMLOptions{})
	if err != nil {
		t.Fatal(err)
	}
	got := string(result.HTML)
	for _, want := range []string{`<table style="border-collapse:separate;width:400px;table-layout:fixed;border-spacing:2.6666666666666665px`, `<colgroup><col style="width:133.33333333333334px"><col style="width:266.6666666666667px"></colgroup>`, `colspan="2"`, `rowspan="2"`, "merged root", "nested", "left", "right", "border-top:1pt solid #112233", "padding-bottom:6.666666666666667px", "background-color:#FFEEDD", "border-left:1.5pt double #445566"} {
		if !strings.Contains(got, want) {
			t.Errorf("HTML missing %q: %s", want, got)
		}
	}
	assertHTMLStructure(t, got)
}

func TestReaderFontScriptGroupsAndLanguages(t *testing.T) {
	body := `<w:p><w:r><w:rPr><w:rFonts w:ascii="Arial" w:hAnsi="Liberation Sans" w:eastAsia="SimSun" w:cs="Noto Naskh Arabic" w:hint="eastAsia"/><w:lang w:val="en-US" w:eastAsia="zh-CN" w:bidi="ar-SA"/></w:rPr><w:t>mixed scripts</w:t></w:r></w:p>`
	raw := contentDOCXFixture(t, body, "", nil)
	doc, err := LoadBytes(raw)
	if err != nil {
		t.Fatal(err)
	}
	readFont := func() style.Font {
		var found style.Font
		walkDocument(doc, func(el element.Element) {
			if text, ok := el.(*element.Text); ok && text.Content == "mixed scripts" {
				found, _ = text.FontStyle.(style.Font)
			}
		})
		return found
	}
	want := style.Font{ASCII: "Arial", HighANSI: "Liberation Sans", EastAsia: "SimSun", ComplexScript: "Noto Naskh Arabic", Hint: "eastAsia", Lang: "en-US", EastAsiaLang: "zh-CN", ComplexScriptLang: "ar-SA"}
	if got := readFont(); got.ASCII != want.ASCII || got.HighANSI != want.HighANSI || got.EastAsia != want.EastAsia || got.ComplexScript != want.ComplexScript || got.Hint != want.Hint || got.Lang != want.Lang || got.EastAsiaLang != want.EastAsiaLang || got.ComplexScriptLang != want.ComplexScriptLang {
		t.Fatalf("script fonts/languages not retained: %+v", got)
	}
	doc = cycleDocument(t, doc)
	got, err := doc.RenderHTML(HTMLOptions{})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"SimSun", "Arial", "Liberation Sans", "Noto Naskh Arabic"} {
		if !strings.Contains(string(got), want) {
			t.Errorf("HTML font fallback missing %q: %s", want, got)
		}
	}
	font := readFont()
	if font.ASCII != "Arial" || font.HighANSI != "Liberation Sans" || font.EastAsia != "SimSun" || font.ComplexScript != "Noto Naskh Arabic" || font.Lang != "en-US" || font.EastAsiaLang != "zh-CN" || font.ComplexScriptLang != "ar-SA" {
		t.Fatalf("script fonts/languages changed across round trip: %+v", font)
	}
}

func TestReaderTablePercentageWidth(t *testing.T) {
	raw := contentDOCXFixture(t, `<w:tbl><w:tblPr><w:tblW w:w="2500" w:type="pct"/><w:tblLayout w:type="fixed"/></w:tblPr><w:tblGrid><w:gridCol w:w="2500"/></w:tblGrid><w:tr><w:tc><w:p><w:r><w:t>half width</w:t></w:r></w:p></w:tc></w:tr></w:tbl>`, "", nil)
	doc, err := LoadBytes(raw)
	if err != nil {
		t.Fatal(err)
	}
	var table *element.Table
	walkDocument(doc, func(el element.Element) {
		if value, ok := el.(*element.Table); ok {
			table = value
		}
	})
	if table == nil || table.Style.Width != 2500 || table.Style.Unit != "pct" {
		t.Fatalf("percentage width not retained: %+v", table)
	}
	doc = cycleDocument(t, doc)
	got, err := doc.RenderHTML(HTMLOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(got), "width:50%") {
		t.Fatalf("percentage width did not survive round trip: %s", got)
	}
}
