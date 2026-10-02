package word

import (
	"bytes"
	"strings"
	"testing"

	"github.com/yunkeweb/go-word/style"
)

func TestHTMLNumberingDOCX(t *testing.T) {
	doc := New()
	doc.AddNumberingStyle("outline", style.Numbering{Type: "multilevel", Levels: []style.NumberingLevel{
		{Start: 3, Format: style.NumberUpperRoman, Text: "%1."},
		{Start: 1, Format: style.NumberLowerLetter, Text: "%2."},
	}})
	sec := doc.AddSection()
	sec.AddListItem("Chapter", 0, nil, nil, "outline")
	child := sec.AddListItemRun(1, "outline", nil)
	child.AddText("Bold", style.Font{Bold: true})
	child.AddLink("https://example.test", "Link")
	sec.AddText("interruption")
	sec.AddListItem("Next chapter", 0, nil, nil, "outline")
	cell := sec.AddTable().AddRow().AddCell(0)
	cell.AddListItem("Cell list", 0, nil, nil, "outline")
	raw, err := doc.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	loaded, err := LoadBytes(raw)
	if err != nil {
		t.Fatal(err)
	}
	for name, d := range map[string]*Document{"memory": doc, "docx": loaded} {
		t.Run(name, func(t *testing.T) {
			got, err := d.RenderHTML(HTMLOptions{Strict: true})
			if err != nil {
				t.Fatal(err)
			}
			for _, want := range []string{`<ol type="I" start="3">`, `<ol type="a">`, `<ol type="I" start="4">`, `<ol type="I" start="5">`, `font-weight:700`, `href="https://example.test"`} {
				if !strings.Contains(string(got), want) {
					t.Errorf("missing %q: %s", want, got)
				}
			}
			assertHTMLStructure(t, string(got))
		})
	}
}

func TestHTMLMixedNestedLists(t *testing.T) {
	d := New()
	s := d.AddSection()
	s.AddListItem("parent", 0, nil, nil, style.ListTypeNumber)
	s.AddListItem("bullet", 1, nil, nil, style.ListTypeBullet)
	s.AddListItem("number", 1, nil, nil, style.ListTypeNumber)
	s.AddListItem("sibling", 0, nil, nil, style.ListTypeNumber)
	got, err := d.RenderHTML(HTMLOptions{})
	if err != nil {
		t.Fatal(err)
	}
	want := "<ol><li>parent<ul><li>bullet</li></ul><ol><li>number</li></ol></li><li>sibling</li></ol>"
	if !bytes.Equal(got, []byte(want)) {
		t.Fatalf("got %s", got)
	}
}
