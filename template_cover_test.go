package word

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yunkeweb/go-word/element"
	"github.com/yunkeweb/go-word/style"
)

func TestTemplateProcessorCoverage(t *testing.T) {
	src := New()
	sec := src.AddSection()
	sec.AddText("Hello ${name}")
	sec.AddText("See ${chart}")
	sec.AddText("Img ${img}")
	sec.AddText("Val ${val}")
	sec.AddText("Blk ${blk}X${/blk}")
	sec.AddText("Cx ${cx}")
	sec.AddText("Cb ${cb}")
	tbl := sec.AddTable()
	row := tbl.AddRow()
	row.AddCell(1000).AddText("${item}")
	row.AddCell(1000).AddText("${qty}")
	tbl2 := sec.AddTable()
	r2 := tbl2.AddRow()
	r2.AddCell(1000).AddText("${only}")
	b, err := src.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "t.docx")
	if err := os.WriteFile(path, b, 0o644); err != nil {
		t.Fatal(err)
	}
	tp, err := NewTemplateProcessor(path)
	if err != nil {
		t.Fatal(err)
	}
	tp.SetMacroOpeningChars("${")
	tp.SetMacroClosingChars("}")
	tp.SetMacroChars("${", "}")
	SetMacroOpeningChars("${")
	SetMacroClosingChars("}")
	SetMacroChars("${", "}")

	tp.SetValues(map[string]string{"name": "World"})
	tp.SetValueLimit("val", "X", 1)
	tp.SetValueLimit("missing", "Y", 0)
	if err := tp.CloneBlockAndSetValues("blk", []map[string]string{{"blk": "1"}, {"blk": "2"}}); err != nil {
		t.Fatal(err)
	}
	if err := tp.ReplaceBlock("nope", "x"); err == nil {
		t.Fatal("missing block")
	}
	tp2, err := NewTemplateProcessorBytes(b)
	if err != nil {
		t.Fatal(err)
	}
	if err := tp2.DeleteBlock("blk"); err != nil {
		t.Fatal(err)
	}
	if err := tp2.ReplaceBlock("blk", "Z"); err != nil {
		// already deleted
	}

	tp3, err := NewTemplateProcessorBytes(b)
	if err != nil {
		t.Fatal(err)
	}
	if err := tp3.CloneBlock("missing", 1); err == nil {
		t.Fatal("clone missing")
	}
	if err := tp3.CloneRow("missing", 1); err == nil {
		t.Fatal("clone row missing")
	}

	imgPath := filepath.Join(dir, "i.png")
	if err := os.WriteFile(imgPath, pngBytes(t), 0o644); err != nil {
		t.Fatal(err)
	}
	tp4, err := NewTemplateProcessorBytes(b)
	if err != nil {
		t.Fatal(err)
	}
	if err := tp4.SetImageValue("img", imgPath); err != nil {
		t.Fatal(err)
	}
	if err := tp4.SetImageValue("img", filepath.Join(dir, "no.png")); err == nil {
		t.Fatal("missing image")
	}
	if err := tp4.SetImageValueBytes("img", "pic.jpg", pngBytes(t)); err != nil {
		t.Fatal(err)
	}
	if err := tp4.SetComplexValue("cx", element.NewText("hi", nil, nil)); err != nil {
		t.Fatal(err)
	}
	tpCx, err := NewTemplateProcessorBytes(b)
	if err != nil {
		t.Fatal(err)
	}
	if err := tpCx.SetComplexBlock("cx", element.NewText("p", nil, nil)); err != nil {
		t.Fatal(err)
	}
	if err := tp4.ReplaceXmlBlock("nope", "x", "w:p"); err == nil {
		t.Fatal("missing xml block")
	}
	tp4.SetCheckbox("cb", true)
	tp4.SetCheckbox("cb", false)
	tp4.SetCheckbox("missing", true)
	tp4.SetUpdateFields(true)
	tp4.SetUpdateFields(false)
	if err := tp4.SetChart("chart", nil); err == nil {
		t.Fatal("nil chart")
	}
	if err := tp4.SetChart("chart", &element.Chart{ChartType: "bar", Categories: []string{"A"}, Values: []float64{1}}); err != nil {
		t.Fatal(err)
	}
	tpChart, err := NewTemplateProcessorBytes(b)
	if err != nil {
		t.Fatal(err)
	}
	if err := tpChart.SetChart("chart", &element.Chart{ChartType: "pie", Categories: []string{"A"}, Values: []float64{1}}); err != nil {
		t.Fatal(err)
	}
	_ = tpChart.SetChart("chart", &element.Chart{ChartType: "pie", Categories: []string{"A"}, Values: []float64{1}})
	outPath := filepath.Join(dir, "out.docx")
	if err := tp4.Save(outPath); err != nil {
		t.Fatal(err)
	}
	if err := tp4.SaveAs(filepath.Join(dir, "out2.docx")); err != nil {
		t.Fatal(err)
	}
	if _, err := tp4.Bytes(); err != nil {
		t.Fatal(err)
	}

	tp5, err := NewTemplateProcessorBytes(b)
	if err != nil {
		t.Fatal(err)
	}
	if err := tp5.DeleteRow("only"); err != nil {
		t.Fatal(err)
	}
	if err := tp5.DeleteRow("missing"); err == nil {
		t.Fatal("delete missing")
	}
	if err := tp5.CloneRowAndSetValues("missing", nil); err == nil {
		t.Fatal("clone values missing")
	}

	if _, err := NewTemplateProcessor(filepath.Join(dir, "no.docx")); err == nil {
		t.Fatal("missing template")
	}
	if _, err := NewTemplateProcessorBytes([]byte("nope")); err == nil {
		t.Fatal("bad bytes")
	}

	if extractRowContaining("abc", "zz") != "" {
		t.Fatal("no needle")
	}
	if extractRowContaining("<w:p>${x}</w:p>", "${x}") != "" {
		t.Fatal("no row")
	}
	if lastTagStart("hello", "w:tr") >= 0 {
		t.Fatal("no tag")
	}
	if lastTagStart("<w:tr>x", "w:tr") < 0 {
		t.Fatal("tag")
	}
	if replaceRowMacros("${/blk}", 1) != "${/blk}" {
		t.Fatal("slash macro")
	}
	start, end := findXMLBlock("abc", "zz", "w:p")
	if start >= 0 || end >= 0 {
		t.Fatal("find missing")
	}
	start, end = findXMLBlock("<w:t>${x}</w:t>", "${x}", "w:p")
	if start >= 0 {
		t.Fatal("no parent")
	}
	start, end = findXMLBlock("<w:p>${x}", "${x}", "w:p")
	if start >= 0 {
		t.Fatal("no close")
	}

	tp6, err := NewTemplateProcessorBytes(b)
	if err != nil {
		t.Fatal(err)
	}
	delete(tp6.files, "word/settings.xml")
	tp6.SetUpdateFields(true)

	src2 := New()
	s2 := src2.AddSection()
	s2.AddText("${cb}")
	b2, err := src2.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	tp7, err := NewTemplateProcessorBytes(b2)
	if err != nil {
		t.Fatal(err)
	}
	xml := string(tp7.files["word/document.xml"])
	xml = strings.Replace(xml, "${cb}", `<w:sdt><w:sdtPr><w14:checked w14:val="0"/></w:sdtPr><w:sdtContent><w:r><w:t>☐</w:t></w:r></w:sdtContent></w:sdt>`, 1)
	tp7.files["word/document.xml"] = []byte(xml)
	tp7.SetCheckbox("cb", true)

	settings := string(tp7.files["word/settings.xml"])
	if strings.Contains(settings, "updateFields") {
		tp7.SetUpdateFields(true)
	} else {
		tp7.SetUpdateFields(true)
		tp7.SetUpdateFields(false)
	}

	_ = style.Font{}
}

func TestTemplateCloneRowMacros(t *testing.T) {
	src := New()
	sec := src.AddSection()
	tbl := sec.AddTable()
	r := tbl.AddRow()
	r.AddCell(100).AddText("${item}")
	r.AddCell(100).AddText("${/skip}")
	b, err := src.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	tp, err := NewTemplateProcessorBytes(b)
	if err != nil {
		t.Fatal(err)
	}
	if err := tp.CloneRow("item", 2); err != nil {
		t.Fatal(err)
	}
	out := string(tp.files["word/document.xml"])
	if !strings.Contains(out, "${item#1}") {
		t.Fatalf("%s", out)
	}
}

func TestUnwrapMacroAndXmlEscape(t *testing.T) {
	if unwrapMacro(" ${x} ") != "x" {
		t.Fatal("unwrap")
	}
	s := xmlEscape(`a&b<c>"`)
	if !strings.Contains(s, "&amp;") || !strings.Contains(s, "&lt;") {
		t.Fatalf("%s", s)
	}
	if nextRelID(`Id="rId3" Id="rId1"`) != "rId4" {
		t.Fatal("nextRelID")
	}
}
