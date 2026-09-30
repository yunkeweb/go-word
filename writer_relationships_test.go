package word

import (
	"bytes"
	"github.com/yunkeweb/go-word/pkg/common"
	"strings"
	"testing"
)

func TestHeaderHyperlinkUsesHeaderRelationships(t *testing.T) {
	d := New()
	h := d.AddSection().AddHeader()
	h.AddLink("https://example.com/header", "header link")
	b, err := d.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	zr, err := common.OpenZipBytes(b)
	if err != nil {
		t.Fatal(err)
	}
	defer zr.Close()
	header, _ := zr.ReadFile("word/header1.xml")
	headerRels, _ := zr.ReadFile("word/_rels/header1.xml.rels")
	docRels, _ := zr.ReadFile("word/_rels/document.xml.rels")
	if !bytes.Contains(header, []byte(`r:id="`)) {
		t.Fatal("header hyperlink has no relation id")
	}
	if !bytes.Contains(headerRels, []byte("https://example.com/header")) {
		t.Fatal("header relationship missing")
	}
	if bytes.Contains(docRels, []byte("https://example.com/header")) {
		t.Fatal("header hyperlink leaked into document relationships")
	}
	if !strings.Contains(string(headerRels), `Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/hyperlink"`) {
		t.Fatal("wrong relationship type")
	}
}
