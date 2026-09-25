package word

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/yunkeweb/go-word/element"
	"github.com/yunkeweb/go-word/pkg/common"
	"github.com/yunkeweb/go-word/style"
)

func TestStreamWriterParagraphsAndTable(t *testing.T) {
	var buf bytes.Buffer
	sw := NewStreamWriter(&buf)
	if sw.BytesWritten() != 0 {
		t.Fatal("nothing written yet")
	}
	if err := sw.WriteParagraph("Hello stream"); err != nil {
		t.Fatal(err)
	}
	if err := sw.WriteParagraph("Styled", style.Font{Bold: true}, style.Paragraph{Alignment: style.JcCenter}); err != nil {
		t.Fatal(err)
	}
	tbl := element.NewTable(nil)
	row := tbl.AddRow()
	row.AddCell(1200).AddText("A")
	row.AddCell(1200).AddText("B")
	row2 := tbl.AddRow()
	row2.AddCell(1200).AddText("C")
	row2.AddCell(1200).AddText("D")
	if err := sw.WriteTable(tbl); err != nil {
		t.Fatal(err)
	}
	if err := sw.WriteElement(element.NewText("via element", nil, nil)); err != nil {
		t.Fatal(err)
	}
	if err := sw.Close(); err != nil {
		t.Fatal(err)
	}
	if sw.BytesWritten() == 0 {
		t.Fatal("expected bytes")
	}
	if err := sw.Close(); err != nil {
		t.Fatal("second close")
	}
	if err := sw.WriteParagraph("after close"); err == nil {
		t.Fatal("write after close")
	}

	zr, err := common.OpenZipBytes(buf.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	defer zr.Close()
	for _, name := range []string{
		"word/document.xml",
		"[Content_Types].xml",
		"_rels/.rels",
		"word/styles.xml",
	} {
		if !zr.Has(name) {
			t.Fatalf("missing %s", name)
		}
	}
	docXML, err := zr.ReadFile("word/document.xml")
	if err != nil {
		t.Fatal(err)
	}
	s := string(docXML)
	for _, need := range []string{"Hello stream", "Styled", "<w:tbl", "via element", "w:sectPr"} {
		if !strings.Contains(s, need) {
			t.Fatalf("document.xml missing %s\n%s", need, s)
		}
	}
}

func TestStreamWriterNilInputsAndFailDest(t *testing.T) {
	var buf bytes.Buffer
	sw := NewStreamWriter(&buf)
	if err := sw.WriteTable(nil); err == nil {
		t.Fatal("nil table")
	}
	if err := sw.WriteElement(nil); err == nil {
		t.Fatal("nil element")
	}
	if err := sw.Close(); err != nil {
		t.Fatal(err)
	}

	sw2 := NewStreamWriter(failWriter{})
	_ = sw2.WriteParagraph("x")
	if err := sw2.Close(); err == nil {
		t.Fatal("expected fail dest")
	}
	if err := sw2.WriteParagraph("y"); err == nil {
		t.Fatal("write after fail")
	}
}

func TestStreamWriterEmptyClose(t *testing.T) {
	var buf bytes.Buffer
	sw := NewStreamWriter(&buf)
	if err := sw.Close(); err != nil {
		t.Fatal(err)
	}
	zr, err := common.OpenZipBytes(buf.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	defer zr.Close()
	if !zr.Has("word/document.xml") {
		t.Fatal("empty stream missing document")
	}
}

func TestDocumentWriteToStreamsWithoutFullBuffer(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	for i := 0; i < 20; i++ {
		sec.AddText("row text")
	}
	tbl := sec.AddTable()
	for r := 0; r < 10; r++ {
		row := tbl.AddRow()
		for c := 0; c < 4; c++ {
			row.AddCell(800).AddText("cell")
		}
	}
	var buf bytes.Buffer
	n, err := doc.WriteTo(&buf)
	if err != nil {
		t.Fatal(err)
	}
	if n != int64(buf.Len()) {
		t.Fatalf("n=%d len=%d", n, buf.Len())
	}
	zr, err := common.OpenZipBytes(buf.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	defer zr.Close()
	body, err := zr.ReadFile("word/document.xml")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(body, []byte("<w:tbl")) || !bytes.Contains(body, []byte("cell")) {
		t.Fatal(string(body))
	}
}

func TestPooledDocumentXMLAndHdrFtr(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	sec.AddText("pool")
	sec.AddHeader().AddText("h")
	w := newWord2007Writer(doc)
	if err := w.prepare(); err != nil {
		t.Fatal(err)
	}
	b := w.documentXML()
	if !bytes.Contains(b, []byte("pool")) {
		t.Fatal(string(b))
	}
	h := w.hdrFtrXML("w:hdr", w.headers[0].El)
	if !bytes.Contains(h, []byte("h")) {
		t.Fatal(string(h))
	}
	_ = w.stylesXML()
	_ = w.numberingXML()
	_ = w.settingsXML()
	_ = w.notesXML(true)
	_ = w.commentsXML()
}

func TestAddXMLAndCountingWriter(t *testing.T) {
	cw := &countingWriter{w: io.Discard}
	n, err := cw.Write([]byte("abcd"))
	if err != nil || n != 4 || cw.n != 4 {
		t.Fatalf("%d %d %v", n, cw.n, err)
	}
}

func TestStreamWriterWriteElementKinds(t *testing.T) {
	var buf bytes.Buffer
	sw := NewStreamWriter(&buf)
	img := &element.Image{Data: pngBytes(t), Style: style.Image{Width: 10, Height: 10}}
	if err := sw.WriteElement(img); err != nil {
		t.Fatal(err)
	}
	ch := &element.Chart{ChartType: "pie", Categories: []string{"A"}, Values: []float64{1}}
	if err := sw.WriteElement(ch); err != nil {
		t.Fatal(err)
	}
	if err := sw.WriteElement(element.NewLink("https://example.com", "ex", nil, nil, false)); err != nil {
		t.Fatal(err)
	}
	ole := element.NewOLEObject("", nil)
	ole.Media.Data = []byte("ole-bytes")
	if err := sw.WriteElement(ole); err != nil {
		t.Fatal(err)
	}
	if err := sw.Close(); err != nil {
		t.Fatal(err)
	}
	zr, err := common.OpenZipBytes(buf.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	defer zr.Close()
	foundMedia, foundChart, foundOLE := false, false, false
	for _, f := range zr.Files() {
		if strings.Contains(f.Name, "word/media/") {
			foundMedia = true
		}
		if strings.Contains(f.Name, "word/charts/") {
			foundChart = true
		}
		if strings.Contains(f.Name, "embeddings/") {
			foundOLE = true
		}
	}
	if !foundMedia || !foundChart || !foundOLE {
		t.Fatalf("media=%v chart=%v ole=%v", foundMedia, foundChart, foundOLE)
	}
}

func TestWord2007SaveCreateError(t *testing.T) {
	doc := New()
	doc.AddSection().AddText("x")
	if err := doc.Save(t.TempDir()); err == nil {
		t.Fatal("expected save to directory to fail")
	}
}

type nthFailWriter struct {
	n, i int
}

func (w *nthFailWriter) Write(p []byte) (int, error) {
	w.i++
	if w.i >= w.n {
		return 0, io.ErrClosedPipe
	}
	return len(p), nil
}

func TestWriteToNthFail(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	sec.AddText("x")
	sec.AddHeader().AddText("h")
	sec.AddFooter().AddText("f")
	failed := 0
	for n := 1; n <= 40; n++ {
		if _, err := doc.WriteTo(&nthFailWriter{n: n}); err != nil {
			failed++
		}
	}
	if failed == 0 {
		t.Fatal("expected some write failures")
	}
}

func TestStreamWriterBeginAndCloseErrors(t *testing.T) {
	sw := NewStreamWriter(&nthFailWriter{n: 1})
	_ = sw.WriteParagraph("x")
	_ = sw.Close()
	sw2 := NewStreamWriter(&nthFailWriter{n: 3})
	_ = sw2.WriteParagraph("x")
	_ = sw2.WriteTable(element.NewTable(nil))
	_ = sw2.Close()
}
