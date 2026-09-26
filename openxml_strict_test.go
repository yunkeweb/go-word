package word

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"io"
	"regexp"
	"strings"
	"testing"

	"github.com/yunkeweb/go-word/element"
	"github.com/yunkeweb/go-word/ooxml"
	"github.com/yunkeweb/go-word/style"
)

func readZipFile(t *testing.T, raw []byte, name string) string {
	t.Helper()
	zr, err := zip.NewReader(bytes.NewReader(raw), int64(len(raw)))
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range zr.File {
		if f.Name != name {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			t.Fatal(err)
		}
		b, err := io.ReadAll(rc)
		_ = rc.Close()
		if err != nil {
			t.Fatal(err)
		}
		return string(b)
	}
	t.Fatalf("missing zip part %s", name)
	return ""
}

func assertWellFormedXML(t *testing.T, name, s string) {
	t.Helper()
	dec := xml.NewDecoder(strings.NewReader(s))
	for {
		_, err := dec.Token()
		if err == io.EOF {
			return
		}
		if err != nil {
			t.Fatalf("%s not well-formed: %v", name, err)
		}
	}
}

func innerChildNames(block, parent string) []string {
	open := "<" + parent
	i := strings.Index(block, open)
	if i < 0 {
		return nil
	}
	gt := strings.Index(block[i:], ">")
	if gt < 0 {
		return nil
	}
	start := i + gt + 1
	close := "</" + parent + ">"
	end := strings.Index(block[start:], close)
	if end < 0 {
		return nil
	}
	inner := block[start : start+end]
	re := regexp.MustCompile(`<([A-Za-z0-9]+:[A-Za-z0-9]+)`)
	var out []string
	for _, m := range re.FindAllStringSubmatch(inner, -1) {
		out = append(out, m[1])
	}
	return out
}

func indexOf(seq []string, name string) int {
	for i, s := range seq {
		if s == name {
			return i
		}
	}
	return -1
}

func assertSeq(t *testing.T, seq []string, names ...string) {
	t.Helper()
	last := -1
	var lastName string
	for _, n := range names {
		i := indexOf(seq, n)
		if i < 0 {
			continue
		}
		if i < last {
			t.Fatalf("child order: %s at %d after %s at %d in %v", n, i, lastName, last, seq)
		}
		last, lastName = i, n
	}
}

func TestOpenXMLContentTypesAndRels(t *testing.T) {
	doc := New()
	doc.AddChart(ChartTypeBar, []string{"A", "B"}, []float64{1, 2})
	doc.CommentOn("text", "note", "Ann", "A", "2026-01-01T00:00:00Z")
	raw, err := doc.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	ct := readZipFile(t, raw, "[Content_Types].xml")
	assertWellFormedXML(t, "[Content_Types].xml", ct)
	if !strings.Contains(ct, `ContentType="`+ooxml.CTChart+`"`) {
		t.Fatal("chart content type")
	}
	if !strings.Contains(ct, `PartName="/word/charts/chart1.xml"`) {
		t.Fatal("chart override")
	}
	if !strings.Contains(ct, `ContentType="`+ooxml.CTComments+`"`) {
		t.Fatal("comments content type")
	}
	if !strings.Contains(ct, `PartName="/word/comments.xml"`) {
		t.Fatal("comments override")
	}

	rels := readZipFile(t, raw, "word/_rels/document.xml.rels")
	assertWellFormedXML(t, "document.xml.rels", rels)
	if !strings.Contains(rels, ooxml.NSOfficeRelChart) {
		t.Fatal("chart relationship type URI")
	}
	if !strings.Contains(rels, `Target="charts/chart1.xml"`) {
		t.Fatal("chart relationship target")
	}
	if !strings.Contains(rels, ooxml.NSOfficeRelComments) {
		t.Fatal("comments relationship type URI")
	}
	if !strings.Contains(rels, `Target="comments.xml"`) {
		t.Fatal("comments relationship target")
	}
}

func TestOpenXMLDocumentAndChartNamespaces(t *testing.T) {
	doc := New()
	doc.AddChartStyled(ChartTypeLine, []string{"Q1", "Q2"}, []float64{3, 4}, style.Chart{
		Title: "T", ShowLegend: true, ShowGridY: true, Width: 2000000, Height: 1000000,
	})
	raw, err := doc.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	document := readZipFile(t, raw, "word/document.xml")
	assertWellFormedXML(t, "document.xml", document)
	for _, ns := range []string{
		`xmlns:w="` + ooxml.NSW + `"`,
		`xmlns:r="` + ooxml.NSR + `"`,
		`xmlns:a="` + ooxml.NSA + `"`,
		`xmlns:c="` + ooxml.NSC + `"`,
		`xmlns:wp="` + ooxml.NSWP + `"`,
	} {
		if !strings.Contains(document, ns) {
			t.Fatalf("document.xml missing %s", ns)
		}
	}
	chart := readZipFile(t, raw, "word/charts/chart1.xml")
	assertWellFormedXML(t, "chart1.xml", chart)
	for _, ns := range []string{
		`xmlns:c="` + ooxml.NSC + `"`,
		`xmlns:a="` + ooxml.NSA + `"`,
		`xmlns:r="` + ooxml.NSR + `"`,
	} {
		if !strings.Contains(chart, ns) {
			t.Fatalf("chart1.xml missing %s", ns)
		}
	}
}

func TestOpenXMLPropertyChildOrder(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	falseWC := false
	sec.AddText("para", style.Font{
		Name: "Calibri", Bold: true, Italic: true, AllCaps: true, SmallCaps: true,
		NoProof: true, Color: "FF0000", Underline: style.UnderlineSingle,
		BgColor: "FFFF00", SuperScript: true, Lang: "en-US", Size: 12,
	}, style.Paragraph{
		KeepNext: true, Alignment: style.JcCenter, OutlineLevel: 2,
		WidowControl: &falseWC,
		Spacing:      style.Spacing{Before: 120, After: 120},
		Indentation:  style.Indentation{Left: 200},
		Borders:      style.Borders{Bottom: style.Border{Style: "single", Size: 4, Color: "000000"}},
		Shading:      style.Shading{Fill: "EEEEEE"},
		Tabs:         []style.Tab{{Val: "left", Pos: 1440}},
		Bidi:         true, ContextualSpacing: true, TextAlignment: "center",
		SuppressAutoHyphens: true,
	})
	tbl := sec.AddTable(style.Table{
		Width: 4000, Alignment: style.JcTableCenter, Indent: 100,
		Layout: "autofit",
		CellMarginTop: 40, CellMarginLeft: 40, CellMarginBottom: 40, CellMarginRight: 40,
		Position: &style.TablePosition{LeftFromText: 10, TblpX: 20},
		Borders:  style.Borders{Top: style.Border{Style: "single", Size: 4, Color: "000000"}},
	})
	row := tbl.AddRow(200)
	row.AddCell(4000, style.Cell{
		Width: 4000, GridSpan: 2, VMerge: "restart", VAlign: style.VAlignCenter,
		NoWrap: true, TextDir: "btLr",
		PaddingTop: 20, PaddingLeft: 20, PaddingBottom: 20, PaddingRight: 20,
		BgColor: "DDEEFF",
		Borders: style.Borders{Top: style.Border{Style: "single", Size: 4, Color: "111111"}},
	}).AddText("cell")

	raw, err := doc.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	document := readZipFile(t, raw, "word/document.xml")
	pPr := innerChildNames(document, "w:pPr")
	assertSeq(t, pPr,
		"w:keepNext", "w:widowControl", "w:pBdr", "w:shd", "w:tabs",
		"w:suppressAutoHyphens", "w:bidi", "w:spacing", "w:ind",
		"w:contextualSpacing", "w:jc", "w:textAlignment", "w:outlineLvl")
	rPr := innerChildNames(document, "w:rPr")
	assertSeq(t, rPr,
		"w:rFonts", "w:b", "w:i", "w:caps", "w:smallCaps", "w:noProof",
		"w:color", "w:sz", "w:u", "w:shd", "w:vertAlign", "w:lang")
	tblPr := innerChildNames(document, "w:tblPr")
	assertSeq(t, tblPr, "w:tblpPr", "w:tblW", "w:jc", "w:tblInd", "w:tblBorders", "w:tblLayout", "w:tblCellMar")
	tcPr := innerChildNames(document, "w:tcPr")
	assertSeq(t, tcPr, "w:tcW", "w:gridSpan", "w:vMerge", "w:tcBorders", "w:shd", "w:noWrap", "w:tcMar", "w:textDirection", "w:vAlign")
}

func TestOpenXMLChartSchema(t *testing.T) {
	doc := New()
	cats := []string{"A", "B"}
	doc.AddChartStyled(ChartTypeBar, cats, []float64{1, 2}, style.Chart{Title: "Bar", ShowLegend: true, ShowGridY: true})
	doc.AddChartStyled(ChartTypeLine, cats, []float64{3, 4}, style.Chart{Title: "Line", ShowLegend: true})
	doc.AddChartStyled(ChartTypePie, cats, []float64{5, 6}, style.Chart{Title: "Pie", ShowLegend: true})
	raw, err := doc.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	bar := readZipFile(t, raw, "word/charts/chart1.xml")
	line := readZipFile(t, raw, "word/charts/chart2.xml")
	pie := readZipFile(t, raw, "word/charts/chart3.xml")
	for name, xml := range map[string]string{"bar": bar, "line": line, "pie": pie} {
		assertWellFormedXML(t, name, xml)
		if !strings.Contains(xml, `xmlns:c="`+ooxml.NSC+`"`) {
			t.Fatalf("%s missing chart ns", name)
		}
		chartKids := innerChildNames(xml, "c:chart")
		assertSeq(t, chartKids, "c:title", "c:plotArea", "c:legend")
		if strings.Contains(xml, "<c:strRef>") && !strings.Contains(xml, "<c:f>") {
			t.Fatalf("%s strRef without required c:f", name)
		}
	}
	barKids := innerChildNames(bar, "c:barChart")
	assertSeq(t, barKids, "c:barDir", "c:grouping", "c:varyColors", "c:ser", "c:axId")
	if strings.Contains(bar, "<c:overlap") {
		t.Fatal("clustered bar must not emit c:overlap")
	}
	if strings.Contains(line, "<c:overlap") {
		t.Fatal("line chart must not emit c:overlap")
	}
	if strings.Contains(pie, "<c:overlap") {
		t.Fatal("pie chart must not emit c:overlap")
	}
	serKids := innerChildNames(bar, "c:ser")
	assertSeq(t, serKids, "c:idx", "c:order", "c:dLbls", "c:cat", "c:val")
	axKids := innerChildNames(bar, "c:catAx")
	assertSeq(t, axKids, "c:axId", "c:scaling", "c:delete", "c:axPos", "c:majorTickMark", "c:crossAx", "c:auto")
	if !strings.Contains(bar, "<c:v>") {
		// series name may be empty; values still use c:v
		t.Fatal("bar series values")
	}
	if !strings.Contains(bar, "<c:formatCode>General</c:formatCode>") {
		t.Fatal("numLit formatCode")
	}
}

func TestOpenXMLCommentsInsideParagraph(t *testing.T) {
	doc := New()
	doc.EnableTrackChanges(true)
	doc.CommentOn("visible", "body", "Ann", "A", "2026-01-01T00:00:00Z")
	doc.AddInsertion("ins", "Bob", "2026-01-01T00:00:00Z")
	doc.AddDeletion("del", "Bob", "2026-01-01T00:00:00Z")
	raw, err := doc.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	document := readZipFile(t, raw, "word/document.xml")
	assertWellFormedXML(t, "document.xml", document)
	if !strings.Contains(document, "<w:commentRangeStart") || !strings.Contains(document, "<w:commentReference") {
		t.Fatal("comment markers")
	}
	stripped := regexp.MustCompile(`(?s)<w:p\b.*?</w:p>`).ReplaceAllString(document, "")
	if strings.Contains(stripped, "commentReference") || strings.Contains(stripped, "<w:r>") {
		t.Fatal("commentReference or run escaped w:p")
	}
	comments := readZipFile(t, raw, "word/comments.xml")
	assertWellFormedXML(t, "comments.xml", comments)
	if !strings.Contains(comments, `xmlns:w="`+ooxml.NSW+`"`) {
		t.Fatal("comments xmlns:w")
	}
	if !strings.Contains(readZipFile(t, raw, "word/settings.xml"), "<w:trackRevisions") {
		t.Fatal("trackRevisions")
	}
}

func TestOpenXMLTableGridMatchesColumns(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	tbl := sec.AddTable(style.Table{Width: 8000})
	r0 := tbl.AddRow(200)
	r0.AddCell(8000, style.Cell{GridSpan: 4, Width: 8000}).AddText("span")
	r1 := tbl.AddRow(200)
	for i := 0; i < 4; i++ {
		r1.AddCell(2000, style.Cell{Width: 2000}).AddText("c")
	}
	raw, err := doc.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	document := readZipFile(t, raw, "word/document.xml")
	grid := innerChildNames(document, "w:tblGrid")
	n := 0
	for _, c := range grid {
		if c == "w:gridCol" {
			n++
		}
	}
	if n != 4 {
		t.Fatalf("tblGrid cols=%d want 4", n)
	}
}

func TestIsolatedModulesWellFormed(t *testing.T) {
	type gen func() ([]byte, error)
	cases := []struct {
		name string
		fn   gen
	}{
		{"a-paragraphs-tables", func() ([]byte, error) {
			doc := New()
			sec := doc.AddSection()
			sec.AddText("hello", style.Font{Bold: true, Size: 12})
			tbl := sec.AddTable(style.Table{Width: 4000})
			r := tbl.AddRow(200)
			r.AddCell(4000, style.Cell{GridSpan: 2, Width: 4000}).AddText("t")
			r2 := tbl.AddRow(200)
			r2.AddCell(2000, style.Cell{Width: 2000}).AddText("a")
			r2.AddCell(2000, style.Cell{Width: 2000}).AddText("b")
			return doc.Bytes()
		}},
		{"b-streamwriter", func() ([]byte, error) {
			var buf bytes.Buffer
			sw := NewStreamWriter(&buf)
			if err := sw.WriteParagraph("stream"); err != nil {
				return nil, err
			}
			tbl := element.NewTable(style.Table{Width: 4000})
			row := tbl.AddRow(200)
			row.AddCell(4000).AddText("c")
			if err := sw.WriteTable(tbl); err != nil {
				return nil, err
			}
			if err := sw.Close(); err != nil {
				return nil, err
			}
			return buf.Bytes(), nil
		}},
		{"c-charts", func() ([]byte, error) {
			doc := New()
			doc.AddChart(ChartTypePie, []string{"A", "B"}, []float64{1, 2})
			return doc.Bytes()
		}},
		{"d-markdown-html", func() ([]byte, error) {
			doc := New()
			if err := doc.AddMarkdown("# H\n\n**b**"); err != nil {
				return nil, err
			}
			if err := doc.AddHTML("<p><em>i</em></p>"); err != nil {
				return nil, err
			}
			return doc.Bytes()
		}},
		{"e-template", func() ([]byte, error) {
			doc := New()
			sec := doc.AddSection()
			sec.AddText("${block_a}")
			sec.AddText("x ${region}")
			sec.AddText("${/block_a}")
			sec.AddText("${if show}")
			sec.AddText("keep")
			sec.AddText("${endif}")
			raw, err := doc.Bytes()
			if err != nil {
				return nil, err
			}
			tp, err := NewTemplateProcessorBytes(raw)
			if err != nil {
				return nil, err
			}
			if err := tp.CloneNestedBlock("block_a", []BlockData{
				{Values: map[string]string{"region": "east"}},
			}); err != nil {
				return nil, err
			}
			if err := tp.SetCondition("show", true); err != nil {
				return nil, err
			}
			return tp.Bytes()
		}},
		{"f-comments-revisions", func() ([]byte, error) {
			doc := New()
			doc.EnableTrackChanges(true)
			doc.CommentOn("t", "c", "A")
			doc.AddInsertion("i", "B", "2026-01-01T00:00:00Z")
			doc.AddDeletion("d", "B", "2026-01-01T00:00:00Z")
			return doc.Bytes()
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			raw, err := tc.fn()
			if err != nil {
				t.Fatal(err)
			}
			document := readZipFile(t, raw, "word/document.xml")
			assertWellFormedXML(t, tc.name+"/document.xml", document)
			assertWellFormedXML(t, tc.name+"/types", readZipFile(t, raw, "[Content_Types].xml"))
			assertWellFormedXML(t, tc.name+"/rels", readZipFile(t, raw, "word/_rels/document.xml.rels"))
		})
	}
}
