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
		Layout:        "autofit",
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
	assertSeq(t, barKids, "c:barDir", "c:grouping", "c:varyColors", "c:ser", "c:dLbls", "c:axId")
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
	assertSeq(t, serKids, "c:idx", "c:order", "c:cat", "c:val")
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

func TestOpenXMLAreaChartChildOrder(t *testing.T) {
	doc := New()
	ch := doc.AddChart(ChartTypeArea, []string{"Q1", "Q2"}, []float64{4, 6})
	ch.Series[0].Name = "Revenue"
	ch.SetDataLabels(ChartDataLabelOptions{ShowVal: true, ShowCatName: true, Position: DataLabelPosTop})
	raw, err := doc.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	xml := readZipFile(t, raw, "word/charts/chart1.xml")
	assertWellFormedXML(t, "area", xml)
	if !strings.Contains(xml, `<c:grouping val="standard"`) {
		t.Fatal("areaChart grouping must be standard")
	}
	areaKids := innerChildNames(xml, "c:areaChart")
	assertSeq(t, areaKids, "c:grouping", "c:ser", "c:dLbls", "c:axId")
	areaInner := innerXML(xml, "c:areaChart")
	if strings.Contains(areaInner, "<c:varyColors") {
		t.Fatal("minimal areaChart must not emit c:varyColors")
	}
	if !strings.Contains(areaInner, "<c:dLbls") {
		t.Fatal("SetDataLabels must emit c:dLbls")
	}
	assertSeq(t, innerChildNames(xml, "c:dLbls"), "c:showLegendKey", "c:showVal", "c:showCatName", "c:showSerName", "c:showPercent")
	if strings.Contains(innerXML(xml, "c:dLbls"), "showBubbleSize") || strings.Contains(innerXML(xml, "c:dLbls"), "showLeaderLines") {
		t.Fatal("area dLbls must omit showBubbleSize/showLeaderLines")
	}
	if strings.Contains(xml, "<c:dLblPos") {
		t.Fatal("area dLbls must omit c:dLblPos")
	}
	if ids := axIDVals(areaInner); len(ids) != 2 {
		t.Fatalf("areaChart must have exactly 2 axId, got %v", ids)
	}
	if strings.Contains(areaInner, "<c:smooth") || strings.Contains(areaInner, "<c:marker") {
		t.Fatal("areaChart must not contain line-only c:smooth or c:marker")
	}
	serKids := innerChildNames(xml, "c:ser")
	assertSeq(t, serKids, "c:idx", "c:order", "c:tx", "c:spPr", "c:cat", "c:val")
	cat := innerXML(xml, "c:cat")
	if strings.Contains(cat, "<c:numLit") || !strings.Contains(cat, "<c:strLit>") {
		t.Fatalf("cat must use strLit, got %s", cat)
	}
	val := innerXML(xml, "c:val")
	if !strings.Contains(val, "<c:numLit>") || strings.Contains(val, "<c:strLit") {
		t.Fatalf("val must use numLit, got %s", val)
	}
	fi := strings.Index(val, "<c:formatCode>General</c:formatCode>")
	pi := strings.Index(val, "<c:ptCount")
	if fi < 0 || pi < 0 || fi > pi {
		t.Fatalf("formatCode must precede ptCount in numLit: %s", val)
	}
	sp := innerXML(xml, "c:spPr")
	if !strings.Contains(sp, `<a:srgbClr val="5B9BD5"`) || !strings.Contains(sp, "<a:noFill") {
		t.Fatalf("area ser spPr: %s", sp)
	}
	plot := innerChildNames(xml, "c:plotArea")
	assertSeq(t, plot, "c:layout", "c:areaChart", "c:catAx", "c:valAx")
	if strings.Contains(innerXML(xml, "c:plotArea"), "<c:legend") {
		t.Fatal("legend must not be inside plotArea")
	}
	assertChartAxIDsAligned(t, xml)
}

func TestOpenXMLComboChartAxIDs(t *testing.T) {
	doc := New()
	cats := []string{"A", "B"}
	ch := doc.AddChart(ChartTypeCombo, cats, []float64{10, 20})
	ch.Series[0].Kind = ChartTypeColumn
	ch.AddComboSeries(ChartTypeLine, cats, []float64{1, 2}, "Rate", true)
	ch.SetDataLabels(ChartDataLabelOptions{ShowVal: true, Position: DataLabelPosTop})
	raw, err := doc.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	xml := readZipFile(t, raw, "word/charts/chart1.xml")
	assertWellFormedXML(t, "combo", xml)
	barKids := innerChildNames(xml, "c:barChart")
	assertSeq(t, barKids, "c:barDir", "c:grouping", "c:varyColors", "c:ser", "c:dLbls", "c:axId")
	lineKids := innerChildNames(xml, "c:lineChart")
	assertSeq(t, lineKids, "c:grouping", "c:varyColors", "c:ser", "c:dLbls", "c:axId")
	plot := innerChildNames(xml, "c:plotArea")
	assertSeq(t, plot, "c:layout", "c:barChart", "c:lineChart", "c:catAx", "c:valAx")
	assertChartAxIDsAligned(t, xml)
	barIDs := axIDVals(innerXML(xml, "c:barChart"))
	lineIDs := axIDVals(innerXML(xml, "c:lineChart"))
	if len(barIDs) != 2 || barIDs[0] != "1" || barIDs[1] != "2" {
		t.Fatalf("barChart axId=%v", barIDs)
	}
	if len(lineIDs) != 2 || lineIDs[0] != "1" || lineIDs[1] != "3" {
		t.Fatalf("lineChart axId=%v", lineIDs)
	}
	cat := axisBlockByID(xml, "c:catAx", "1")
	if !strings.Contains(cat, `<c:crossAx val="2"`) {
		t.Fatalf("catAx must cross primary valAx 2: %s", cat)
	}
	if strings.Contains(cat, "<c:crossesAt") {
		t.Fatal("catAx must not emit c:crossesAt")
	}
	prim := axisBlockByID(xml, "c:valAx", "2")
	if !strings.Contains(prim, `<c:axPos val="l"`) || !strings.Contains(prim, `<c:crossAx val="1"`) {
		t.Fatalf("primary valAx: %s", prim)
	}
	sec := axisBlockByID(xml, "c:valAx", "3")
	if sec == "" {
		t.Fatal("missing secondary valAx 3")
	}
	if !strings.Contains(sec, `<c:axPos val="r"`) {
		t.Fatalf("secondary axPos r: %s", sec)
	}
	if !strings.Contains(sec, `<c:crossAx val="1"`) {
		t.Fatalf("secondary crossAx must point at catAx 1: %s", sec)
	}
	if !strings.Contains(sec, `<c:crosses val="max"`) {
		t.Fatalf("secondary crosses max: %s", sec)
	}
	if strings.Contains(sec, "<c:crossesAt") {
		t.Fatal("secondary valAx must not emit c:crossesAt")
	}
	bar := innerXML(xml, "c:barChart")
	if strings.Contains(bar, `<c:dLblPos val="t"`) {
		t.Fatal("barChart dLblPos t is not Word-safe")
	}
	if !strings.Contains(bar, `<c:dLblPos val="outEnd"`) {
		t.Fatal("barChart dLblPos should map t to outEnd")
	}
}

func TestOpenXMLNoDefaultDLbls(t *testing.T) {
	doc := New()
	doc.AddChart(ChartTypeBar, []string{"A"}, []float64{1})
	raw, err := doc.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	xml := readZipFile(t, raw, "word/charts/chart1.xml")
	if strings.Contains(xml, "<c:dLbls") {
		t.Fatal("default chart must not emit c:dLbls")
	}
}

func TestOpenXMLAreaChartOmitsDLblsWithoutSet(t *testing.T) {
	doc := New()
	doc.AddChart(ChartTypeArea, []string{"Q1"}, []float64{1})
	raw, err := doc.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	xml := readZipFile(t, raw, "word/charts/chart1.xml")
	assertSeq(t, innerChildNames(xml, "c:areaChart"), "c:grouping", "c:ser", "c:axId")
	if strings.Contains(innerXML(xml, "c:areaChart"), "<c:dLbls") {
		t.Fatal("areaChart must not emit c:dLbls unless SetDataLabels")
	}
}

func TestOpenXMLAxisChildOrder(t *testing.T) {
	doc := New()
	ch := doc.AddChart(ChartTypeArea, []string{"Q1"}, []float64{1})
	ch.SetMajorGridlines(true)
	raw, err := doc.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	xml := readZipFile(t, raw, "word/charts/chart1.xml")
	for _, tag := range []string{"c:catAx", "c:valAx"} {
		kids := innerChildNames(xml, tag)
		assertSeq(t, kids, "c:axId", "c:scaling", "c:delete", "c:axPos", "c:majorGridlines", "c:numFmt", "c:majorTickMark", "c:minorTickMark", "c:tickLblPos", "c:spPr", "c:txPr", "c:crossAx", "c:crosses")
	}
	if strings.Contains(xml, "<c:crossesAt") {
		t.Fatal("axes must not emit c:crossesAt")
	}
}

func TestTemplateFilterOutputHasNoNumericEntities(t *testing.T) {
	tp := mustTemplate(t, func(sec *element.Section) {
		sec.AddText("When ${created_at | formatDate:\"2006-01-02\"}")
		sec.AddText("Hello ${name | upper}")
		sec.AddText("Amt ${amount | formatCurrency:¥}")
	})
	tp.SetValue("created_at", "2026-09-26T08:00:00Z")
	tp.SetValue("name", "alice")
	tp.SetValue("amount", "1999.5")
	raw, err := tp.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	docxml := readZipFile(t, raw, "word/document.xml")
	if i := strings.Index(docxml, "&#"); i >= 0 {
		end := i + 24
		if end > len(docxml) {
			end = len(docxml)
		}
		t.Fatalf("numeric character reference in template output: %s", docxml[i:end])
	}
	for _, want := range []string{"When 2026-09-26", "ALICE", "¥1999.50"} {
		if !strings.Contains(docxml, want) {
			t.Fatalf("missing %q in %s", want, docxml)
		}
	}
}

var axIDValRe = regexp.MustCompile(`<c:axId val="([^"]+)"`)

func innerXML(block, parent string) string {
	open := "<" + parent
	i := strings.Index(block, open)
	if i < 0 {
		return ""
	}
	gt := strings.Index(block[i:], ">")
	if gt < 0 {
		return ""
	}
	start := i + gt + 1
	close := "</" + parent + ">"
	end := strings.Index(block[start:], close)
	if end < 0 {
		return ""
	}
	return block[start : start+end]
}

func axIDVals(block string) []string {
	var out []string
	for _, m := range axIDValRe.FindAllStringSubmatch(block, -1) {
		out = append(out, m[1])
	}
	return out
}

func axisBlockByID(chartXML, tag, id string) string {
	rest := chartXML
	open := "<" + tag
	close := "</" + tag + ">"
	for {
		i := strings.Index(rest, open)
		if i < 0 {
			return ""
		}
		end := strings.Index(rest[i:], close)
		if end < 0 {
			return ""
		}
		block := rest[i : i+end+len(close)]
		if strings.Contains(block, `<c:axId val="`+id+`"`) {
			return block
		}
		rest = rest[i+end+len(close):]
	}
}

func assertChartAxIDsAligned(t *testing.T, chartXML string) {
	t.Helper()
	declared := map[string]struct{}{}
	for _, tag := range []string{"c:catAx", "c:valAx"} {
		rest := chartXML
		for {
			block := innerXML(rest, tag)
			if block == "" {
				break
			}
			ids := axIDVals(block)
			if len(ids) == 0 {
				t.Fatalf("%s missing c:axId", tag)
			}
			declared[ids[0]] = struct{}{}
			close := "</" + tag + ">"
			i := strings.Index(rest, close)
			if i < 0 {
				break
			}
			rest = rest[i+len(close):]
		}
	}
	for _, tag := range []string{"c:barChart", "c:lineChart", "c:areaChart", "c:pieChart"} {
		block := innerXML(chartXML, tag)
		if block == "" {
			continue
		}
		for _, id := range axIDVals(block) {
			if _, ok := declared[id]; !ok {
				t.Fatalf("%s references axId %s with no matching axis", tag, id)
			}
		}
	}
	if len(declared) == 0 {
		t.Fatal("no axes declared")
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
