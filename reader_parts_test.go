package word

import (
	"testing"

	"github.com/yunkeweb/go-word/element"
	"github.com/yunkeweb/go-word/style"
)

func cycleDocument(t *testing.T, d *Document) *Document {
	t.Helper()
	raw, err := d.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	loaded, err := LoadBytes(raw)
	if err != nil {
		t.Fatal(err)
	}
	return loaded
}

func TestReaderSectionPartsRoundTrip(t *testing.T) {
	d := New()
	first := d.AddSection()
	first.AddText("first body")
	first.AddHeader(element.HeaderFirst).AddText("first header")
	first.AddHeader(element.HeaderEven).AddLink("https://example.com/header", "header link")
	first.AddFooter().AddText("first footer")
	second := d.AddSection()
	second.SetOrientation(style.OrientationLandscape)
	second.SetColumns(2, 500, true)
	second.Style.MarginLeft = 720
	second.AddText("second body")
	second.AddHeader().AddText("second header")
	for cycle := 0; cycle < 2; cycle++ {
		d = cycleDocument(t, d)
		if len(d.Sections()) != 2 {
			t.Fatalf("cycle %d: sections=%d", cycle, len(d.Sections()))
		}
		s := d.Sections()[0]
		if len(s.Headers) != 2 || len(s.Footers) != 1 {
			t.Fatalf("lost section parts")
		}
		if s.Headers[0].HeaderType != element.HeaderFirst {
			t.Fatal("first header lost")
		}
		s = d.Sections()[1]
		if s.Style.Orientation != style.OrientationLandscape || s.Style.ColsNum != 2 || s.Style.MarginLeft != 720 {
			t.Fatalf("section style lost: %+v", s.Style)
		}
	}
}

func TestReaderNotesRoundTrip(t *testing.T) {
	d := New()
	s := d.AddSection()
	fn := s.AddFootnote()
	fn.NoteID = 7
	fn.AddText("foot text")
	en := s.AddEndnote()
	en.NoteID = 8
	en.AddText("end text")
	raw, err := d.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	got, err := LoadBytes(raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.GetFootnotes()) != 1 || got.GetFootnotes()[0].NoteID != 1 {
		t.Fatalf("footnotes=%+v", got.GetFootnotes())
	}
	if len(got.GetEndnotes()) != 1 || got.GetEndnotes()[0].NoteID != 1 {
		t.Fatalf("endnotes=%+v", got.GetEndnotes())
	}
	_ = element.HeaderAuto
}

func TestReaderCommentsRoundTrip(t *testing.T) {
	d := New()
	s := d.AddSection()
	tx := s.CommentOn("visible", "comment body", "Alice", "A", "2026-01-01T00:00:00Z")
	raw, err := d.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	got, err := LoadBytes(raw)
	if err != nil {
		t.Fatal(err)
	}
	comments := got.GetComments()
	if len(comments) == 0 || comments[0].Author != "Alice" || comments[0].Initials != "A" {
		t.Fatalf("comments=%+v", comments)
	}
	if tx == nil {
		t.Fatal("comment text missing")
	}
}

func TestReaderListRoundTrip(t *testing.T) {
	d := New()
	s := d.AddSection()
	s.AddListItem("one", 0, nil, style.ListItem{ListType: style.ListTypeNumber, Format: style.NumberDecimal}, nil)
	s.AddListItem("two", 1, nil, style.ListItem{ListType: style.ListTypeBullet, Format: "bullet"}, nil)
	raw, err := d.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	got, err := LoadBytes(raw)
	if err != nil {
		t.Fatal(err)
	}
	var lists int
	for _, el := range got.Sections()[0].Elements() {
		if el.Type() == "ListItem" {
			lists++
		}
	}
	if lists != 2 {
		t.Fatalf("list elements=%d: %+v", lists, got.Sections()[0].Elements())
	}
}

func TestReaderRevisionRoundTrip(t *testing.T) {
	d := New()
	s := d.AddSection()
	s.AddInsertion("added", "Alice", "2026-01-01")
	s.AddDeletion("removed", "Bob", "2026-01-02")
	raw, err := d.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	got, err := LoadBytes(raw)
	if err != nil {
		t.Fatal(err)
	}
	var changes int
	walkDocument(got, func(el element.Element) {
		if v, ok := el.(*element.Text); ok && v.GetTrackChange() != nil {
			changes++
		}
	})
	if changes != 2 {
		t.Fatalf("track changes=%d", changes)
	}
}
