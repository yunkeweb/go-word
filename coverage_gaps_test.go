package word

import (
	"archive/zip"
	"bytes"
	"strings"
	"testing"

	"github.com/yunkeweb/go-word/element"
	"github.com/yunkeweb/go-word/pkg/common"
	"github.com/yunkeweb/go-word/style"
)

func TestDocumentWriteToError(t *testing.T) {
	doc := New()
	doc.AddSection().AddText("x")
	if _, err := doc.WriteTo(failWriter{}); err == nil {
		t.Fatal("expected write error")
	}
}

func TestWriterNumberingIDAndTableGrid(t *testing.T) {
	doc := New()
	doc.AddNumberingStyle("num", style.Numbering{})
	sec := doc.AddSection()
	sec.AddListItem("x", 0, nil, nil, "missing-num")
	tbl := sec.AddTable()
	row := tbl.AddRow(120)
	c0 := row.AddCell(0)
	c0.Width = 0
	c0.Style.Width = 800
	c1 := row.AddCell(0)
	c1.Width = 0
	c1.Style.Width = 0
	if _, err := doc.Bytes(); err != nil {
		t.Fatal(err)
	}
	w := newWord2007Writer(doc)
	if w.numberingID("missing-num") != 1 {
		t.Fatal("numbering fallback")
	}
}

func TestWriterRelForHeaderFooterAndImageBranches(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	h := sec.AddHeader()
	h.AddText("h")
	f := sec.AddFooter()
	f.AddText("f")
	sec.AddImageBytes("bad.jpg", []byte("not-an-image"))
	sec.AddImageBytes("ok.png", pngBytes(t), style.Image{WidthEMU: 100, HeightEMU: 100})
	w := newWord2007Writer(doc)
	if err := w.prepare(); err != nil {
		t.Fatal(err)
	}
	if w.relFor(h) == "" {
		t.Fatal("header rel")
	}
	if w.relFor(f) == "" {
		t.Fatal("footer rel")
	}

	emptyExt := &element.Image{Data: []byte("xxxx")}
	if err := w.registerImage(emptyExt); err != nil {
		t.Fatal(err)
	}
	jpgExt := &element.Image{Data: []byte("xxxx"), Media: element.Media{Ext: "jpg"}}
	if err := w.registerImage(jpgExt); err != nil {
		t.Fatal(err)
	}

	xw := common.NewXMLWriter()
	img := &element.Image{Style: style.Image{WidthEMU: 1, HeightEMU: 1}}
	w.images = append(w.images, pkgImage{RelID: "rId99", El: img})
	w.writeImage(xw, img)
}

func TestWriterChartSeriesNameAndAxisLabels(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	ch := sec.AddChart("bar", []string{"A"}, []float64{1})
	ch.SeriesName = "Sales"
	ch.Style.ShowAxisLabels = true
	ch.Style.CategoryAxisTitle = ""
	ch.Style.ValueAxisTitle = ""
	ch.Style.CategoryLabelPosition = ""
	ch.Style.ValueLabelPosition = ""
	ch2 := sec.AddChart("scatter", []string{"A"}, []float64{1})
	ch2.Series = []element.ChartSeries{{
		Name:       "S",
		Categories: []string{"A"},
		Values:     []float64{1},
	}}
	if _, err := doc.Bytes(); err != nil {
		t.Fatal(err)
	}
}

func TestWalkElementTextKinds(t *testing.T) {
	if elementText(&element.Link{Text: "L"}) != "L" {
		t.Fatal("link")
	}
	if elementText(&element.Title{Text: "T"}) != "T" {
		t.Fatal("title")
	}
	if elementText(&element.ListItem{Text: "I"}) != "I" {
		t.Fatal("list")
	}
}

func TestReaderHyperlinkAndFontFlush(t *testing.T) {
	xml := `<?xml version="1.0"?><w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
	<w:body>
		<w:p><w:r><w:rPr><w:b/><w:i w:val="true"/><w:color w:val="FF0000"/></w:rPr><w:t>Bold</w:t><w:br w:type="page"/></w:r></w:p>
		<w:p><w:hyperlink r:id="rId1" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships"><w:r><w:t>open</w:t><w:br w:type="page"/></w:r></w:hyperlink></w:p>
	</w:body></w:document>`
	doc, err := LoadBytes(zipDocXML(t, xml))
	if err != nil {
		t.Fatal(err)
	}
	if doc.GetSection(0) == nil {
		t.Fatal("section")
	}
}

func TestTemplateRemainingBranches(t *testing.T) {
	src := New()
	sec := src.AddSection()
	sec.AddText("${a}${a}${a}")
	sec.AddText("${name} ${name}")
	sec.AddText("Blk ${blk}X${/blk}")
	tbl := sec.AddTable()
	tbl.AddRow().AddCell(100).AddText("${item}")
	tbl.AddRow().AddCell(100).AddText("${other}")
	b, err := src.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	tp, err := NewTemplateProcessorBytes(b)
	if err != nil {
		t.Fatal(err)
	}
	tp.SetValueLimit("a", "X", 1)
	counts := tp.GetVariableCount()
	if counts["name"] < 2 {
		t.Fatalf("counts=%v", counts)
	}
	vars := tp.GetVariables()
	if len(vars) == 0 {
		t.Fatal("vars")
	}
	if err := tp.DeleteRow("item"); err != nil {
		t.Fatal(err)
	}

	tp2, err := NewTemplateProcessorBytes(b)
	if err != nil {
		t.Fatal(err)
	}
	if err := tp2.DeleteRow("name"); err == nil {
		t.Fatal("expected table not found")
	}
	if err := tp2.CloneBlockAndSetValues("missing", []map[string]string{{"x": "1"}}); err == nil {
		t.Fatal("clone block missing")
	}

	tpBad, err := NewTemplateProcessorBytes(b)
	if err != nil {
		t.Fatal(err)
	}
	tpBad.files["word/document.xml"] = []byte(`<w:document><w:body><w:tbl><w:trPr></w:trPr><w:trPr>${item}</w:trPr></w:tbl></w:body></w:document>`)
	if err := tpBad.DeleteRow("item"); err == nil {
		t.Fatal("malformed table row")
	}

	if extractRowContaining("<w:tr>${x}", "${x}") != "" {
		t.Fatal("no row end")
	}
	if lastTagStart(`<w:tr>x<w:tr foo="1">`, "w:tr") < 0 {
		t.Fatal("attr start tag")
	}

	tp3, err := NewTemplateProcessorBytes(b)
	if err != nil {
		t.Fatal(err)
	}
	xml := string(tp3.files["word/document.xml"])
	xml = strings.Replace(xml, "${name}", `<w:sdt><w14:checked w14:val="0"/><w:sdtContent><w:r><w:t>${name}</w:t></w:r></w:sdtContent></w:sdt>`, 1)
	tp3.files["word/document.xml"] = []byte(xml)
	tp3.SetCheckbox("name", true)
	tp4, err := NewTemplateProcessorBytes(b)
	if err != nil {
		t.Fatal(err)
	}
	xml4 := string(tp4.files["word/document.xml"])
	xml4 = strings.Replace(xml4, "${name}", `<w:sdt><w14:checked w14:val="1"/><w:sdtContent><w:r><w:t>${name}</w:t></w:r></w:sdtContent></w:sdt>`, 1)
	tp4.files["word/document.xml"] = []byte(xml4)
	tp4.SetCheckbox("name", false)

	dir := t.TempDir()
	if err := tp.Save(dir); err == nil {
		t.Fatal("save to directory")
	}
}

func TestTemplateEncryptedZip(t *testing.T) {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	w, err := zw.Create("word/document.xml")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := w.Write([]byte("<w:document/>")); err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	raw := append([]byte(nil), buf.Bytes()...)
	for i := 0; i+12 < len(raw); i++ {
		if raw[i] == 'P' && raw[i+1] == 'K' && raw[i+2] == 3 && raw[i+3] == 4 {
			raw[i+8] = 99
			raw[i+9] = 0
		}
		if raw[i] == 'P' && raw[i+1] == 'K' && raw[i+2] == 1 && raw[i+3] == 2 {
			raw[i+10] = 99
			raw[i+11] = 0
		}
	}
	if _, err := NewTemplateProcessorBytes(raw); err == nil {
		t.Fatal("unknown zip method")
	}
}

func TestTemplateCRCMismatch(t *testing.T) {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	h := &zip.FileHeader{Name: "word/document.xml", Method: zip.Store}
	w, err := zw.CreateHeader(h)
	if err != nil {
		t.Fatal(err)
	}
	payload := []byte("<w:document/>")
	if _, err := w.Write(payload); err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	raw := buf.Bytes()
	idx := bytes.Index(raw, payload)
	if idx < 0 {
		t.Fatal("payload missing")
	}
	raw[idx] ^= 0xff
	if _, err := NewTemplateProcessorBytes(raw); err == nil {
		t.Fatal("crc mismatch")
	}
}

func TestTemplateBytesLongName(t *testing.T) {
	src := New()
	src.AddSection().AddText("x")
	b, err := src.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	tp, err := NewTemplateProcessorBytes(b)
	if err != nil {
		t.Fatal(err)
	}
	long := strings.Repeat("a", 70000)
	tp.order = append(tp.order, long)
	tp.files[long] = []byte("x")
	if _, err := tp.Bytes(); err == nil {
		t.Fatal("expected long zip name")
	}
	if err := tp.Save(t.TempDir() + "/x.docx"); err == nil {
		t.Fatal("save via bytes error")
	}
}
