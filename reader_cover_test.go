package word

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"strings"
	"testing"

	"github.com/yunkeweb/go-word/element"
	"github.com/yunkeweb/go-word/pkg/common"
)

func zipDocXML(t *testing.T, documentXML string) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	w, err := zw.Create("word/document.xml")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := w.Write([]byte(documentXML)); err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func TestLoadBytesReaderPaths(t *testing.T) {
	xml := `<?xml version="1.0"?><w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
	<w:body>
		<w:p><w:pPr><w:pStyle w:val="Heading1"/></w:pPr><w:r><w:t>Title</w:t></w:r></w:p>
		<w:p><w:pPr><w:pStyle w:val="Heading"/></w:pPr><w:r><w:t>H</w:t></w:r></w:p>
		<w:p><w:pPr><w:pStyle w:val="Heading0"/></w:pPr><w:r><w:t>H0</w:t></w:r></w:p>
		<w:p><w:pPr><w:pStyle w:val="Normal"/></w:pPr>
			<w:r><w:rPr><w:b/><w:i w:val="true"/><w:color w:val="FF0000"/></w:rPr><w:t>Bold</w:t></w:r>
			<w:r><w:rPr><w:b w:val="0"/><w:i w:val="false"/></w:rPr><w:t></w:t></w:r>
		</w:p>
		<w:p><w:hyperlink r:id="rId9" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships"><w:r><w:t>link</w:t></w:r></w:hyperlink></w:p>
		<w:p><w:hyperlink w:anchor="bm"><w:r><w:t>anc</w:t></w:r></w:hyperlink></w:p>
		<w:p><w:r><w:br w:type="page"/></w:r></w:p>
		<w:p><w:r><w:br/></w:r></w:p>
		<w:p></w:p>
		<w:p><w:pPr><w:pStyle w:val="EmptyStyle"/></w:pPr></w:p>
		<w:tbl>
			<w:tr><w:tc><w:p><w:r><w:t>cell</w:t></w:r></w:p></w:tc></w:tr>
			<w:tr></w:tr>
		</w:tbl>
		<w:drawing><w:unused/></w:drawing>
		<w:sectPr><w:pgSz/></w:sectPr>
	</w:body></w:document>`
	doc, err := LoadBytes(zipDocXML(t, xml))
	if err != nil {
		t.Fatal(err)
	}
	sec := doc.GetSection(0)
	if sec == nil || sec.CountElements() == 0 {
		t.Fatal("empty parse")
	}
	var texts []string
	walkDocument(doc, func(el element.Element) {
		if s := elementText(el); s != "" {
			texts = append(texts, s)
		}
	})
	joined := strings.Join(texts, " ")
	if !strings.Contains(joined, "Title") || !strings.Contains(joined, "cell") {
		t.Fatalf("%q", joined)
	}
}

func TestLoadBytesErrors(t *testing.T) {
	if _, err := LoadBytes([]byte("notzip")); err == nil {
		t.Fatal("bad zip")
	}
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	_ = zw.Close()
	if _, err := LoadBytes(buf.Bytes()); err == nil {
		t.Fatal("missing document.xml")
	}
	if _, err := LoadBytes(zipDocXML(t, `<w:document><`)); err == nil {
		t.Fatal("bad xml")
	}
}

func TestParseDocumentXMLEdges(t *testing.T) {
	sec := element.NewSection(1, nil)
	if err := parseDocumentXML([]byte(`<w:document><w:body><w:p><w:r><w:t>ok</w:t></w:r></w:p></w:body></w:document>`), sec); err != nil {
		t.Fatal(err)
	}
	if err := parseDocumentXML([]byte(`<w:document><w:p><w:t>`), sec); err == nil {
		t.Fatal("truncated t")
	}
	if err := parseDocumentXML([]byte(`<w:document><w:drawing>`), sec); err == nil {
		t.Fatal("truncated drawing")
	}
	if localName(xml.Name{Local: "w:p"}) != "p" {
		t.Fatal("local colon")
	}
	if localName(xml.Name{Local: "p"}) != "p" {
		t.Fatal("local")
	}
	st := xml.StartElement{Name: xml.Name{Local: "x"}, Attr: []xml.Attr{
		{Name: xml.Name{Local: "val"}, Value: "1"},
		{Name: xml.Name{Local: "w:id"}, Value: "2"},
	}}
	if attr(st, "val") != "1" || attr(st, "id") != "2" || attr(st, "no") != "" {
		t.Fatal("attr")
	}
}

func TestWalkAndElementText(t *testing.T) {
	walkElement(nil, func(element.Element) {})
	doc := New()
	sec := doc.AddSection()
	tr := sec.AddTextRun()
	tr.AddText("a")
	lir := sec.AddListItemRun(0, nil, nil)
	lir.AddText("b")
	tb := sec.AddTextBox()
	tb.AddText("c")
	fn := sec.AddFootnote()
	fn.AddText("d")
	en := sec.AddEndnote()
	en.AddText("e")
	sdt := sec.AddSDT("x")
	sdt.AddText("f")
	cm := sec.AddComment("A", "A", "")
	cm.AddText("g")
	tbl := sec.AddTable()
	tbl.AddRow().AddCell(1).AddText("h")
	hdr := sec.AddHeader()
	hdr.AddText("hdr")
	ftr := sec.AddFooter()
	ftr.AddText("ftr")
	sec.AddPreserveText("PT")
	sec.AddCheckBox("n", "CB")
	sec.AddLink("u", "L")
	var n int
	walkDocument(doc, func(el element.Element) { n++ })
	if n == 0 {
		t.Fatal("walk")
	}
	if elementText(&element.PreserveText{Text: element.Text{Content: "p"}}) != "p" {
		t.Fatal("preserve")
	}
	if elementText(&element.CheckBox{Text: element.Text{Content: "c"}}) != "c" {
		t.Fatal("checkbox")
	}
	if elementText(&element.PageBreak{}) != "" {
		t.Fatal("default")
	}
	_ = childElements(&element.PageBreak{})
	_ = common.NewXMLWriter()
}
