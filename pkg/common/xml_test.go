package common

import (
	"bytes"
	"strings"
	"testing"
)

func TestXMLWriterAPI(t *testing.T) {
	w := NewXMLWriter()
	w.End() // empty stack
	w.StartDocument()
	w.Start("root", "a", "1", "skip", "")
	w.Empty("child", "x", "y")
	w.Element("t", "hi")
	w.Element("empty", "")
	w.Text("")
	w.WT("  spaced  ")
	w.WT("plain")
	w.WriteElementIf(false, "no", "x")
	w.WriteElementIf(true, "yes", "x")
	n, v := w.WriteAttributeIf(false, "a", "b")
	if n != "" || v != "" {
		t.Fatal("WriteAttributeIf false")
	}
	n, v = w.WriteAttributeIf(true, "a", "")
	if n != "" {
		t.Fatal("empty value")
	}
	n, v = w.WriteAttributeIf(true, "a", "b")
	if n != "a" || v != "b" {
		t.Fatalf("%s %s", n, v)
	}
	w.WriteElementBlock("blk", []string{"k", "v"}, func() {
		w.Empty("inner")
	})
	w.WriteElementBlock("blk2", nil, nil)
	w.Raw("<!--raw-->")
	s := w.String()
	if !strings.Contains(s, "<?xml") || !strings.Contains(s, "<!--raw-->") {
		t.Fatalf("%s", s)
	}
	if !strings.Contains(s, "xml:space") {
		t.Fatal("preserve")
	}
	var buf bytes.Buffer
	if _, err := w.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}
	if buf.Len() == 0 {
		t.Fatal("WriteTo empty")
	}
}

func TestMarshalXML(t *testing.T) {
	type n struct {
		XMLName struct{} `xml:"n"`
		V       string   `xml:"v,attr"`
	}
	b, err := MarshalXML(n{V: "x"})
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(b, []byte("<?xml")) || !bytes.Contains(b, []byte("v=")) {
		t.Fatalf("%s", b)
	}
	if _, err := MarshalXML(make(chan int)); err == nil {
		t.Fatal("expected marshal error")
	}
}
