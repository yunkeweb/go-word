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
