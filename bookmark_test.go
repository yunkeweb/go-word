package word

import (
	"strings"
	"testing"
)

func TestAddBookmarkAndAnchorHyperlink(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	target := sec.AddTextRun()
	target.AddText("Destination")
	bm := doc.AddBookmark(target, "ch2")
	if bm == nil || bm.Name != "ch2" {
		t.Fatal("bookmark")
	}
	linkPara := sec.AddTextRun()
	linkPara.AddText("See ")
	link := doc.AddHyperlinkToBookmark(linkPara, "chapter 2", "ch2")
	if link == nil || !link.Internal || link.Target != "ch2" || link.Text != "chapter 2" {
		t.Fatalf("link=%+v", link)
	}

	raw, err := doc.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	xml := readZipFile(t, raw, "word/document.xml")
	if !strings.Contains(xml, `w:bookmarkStart`) || !strings.Contains(xml, `w:name="ch2"`) {
		t.Fatalf("missing bookmarkStart: %s", xml)
	}
	if !strings.Contains(xml, `w:bookmarkEnd`) {
		t.Fatal("missing bookmarkEnd")
	}
	if !strings.Contains(xml, `w:hyperlink`) || !strings.Contains(xml, `w:anchor="ch2"`) {
		t.Fatalf("missing anchor hyperlink: %s", xml)
	}
	if strings.Contains(xml, `r:id=`) && strings.Contains(xml, `w:anchor="ch2"`) {
		// internal links must not require a relationship id
		anchorAt := strings.Index(xml, `w:anchor="ch2"`)
		chunkStart := anchorAt - 80
		if chunkStart < 0 {
			chunkStart = 0
		}
		chunk := xml[chunkStart : anchorAt+40]
		if strings.Contains(chunk, `r:id=`) {
			t.Fatalf("anchor hyperlink should not use r:id: %s", chunk)
		}
	}

	loaded, err := LoadBytes(raw)
	if err != nil {
		t.Fatal(err)
	}
	bms := loaded.GetBookmarks()
	found := false
	for _, b := range bms {
		if b.Name == "ch2" {
			found = true
		}
	}
	if !found {
		t.Fatal("parsed bookmark missing")
	}
	text := loaded.ExtractText()
	if !strings.Contains(text, "Destination") || !strings.Contains(text, "chapter 2") {
		t.Fatalf("text=%q", text)
	}
}

func TestAddBookmarkNilParagraphUsesSection(t *testing.T) {
	doc := New()
	if doc.AddBookmark(nil, "") != nil {
		t.Fatal("empty name")
	}
	bm := doc.AddBookmark(nil, "secbm")
	if bm == nil || bm.Name != "secbm" {
		t.Fatal("section bookmark")
	}
	if doc.AddHyperlinkToBookmark(nil, "x", "") != nil {
		t.Fatal("empty bookmark name")
	}
	l := doc.AddHyperlinkToBookmark(nil, "go", "secbm")
	if l == nil || !l.Internal {
		t.Fatal("section hyperlink")
	}
	raw, err := doc.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	xml := readZipFile(t, raw, "word/document.xml")
	if !strings.Contains(xml, `w:name="secbm"`) || !strings.Contains(xml, `w:anchor="secbm"`) {
		t.Fatal(xml)
	}
}
