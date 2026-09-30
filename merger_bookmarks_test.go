package word

import (
	"testing"

	"github.com/yunkeweb/go-word/element"
)

func TestMergeRemapsCrossSectionBookmarks(t *testing.T) {
	for _, forward := range []bool{false, true} {
		dst := New()
		dst.AddSection().AddBookmark("target")
		src := New()
		first, second := src.AddSection(), src.AddSection()
		target, from := first, second
		if forward {
			target, from = second, first
		}
		target.AddBookmark("target")
		original := from.AddLink("target", "jump", nil, nil, true)
		if err := dst.AppendDocument(src, MergeOptions{}); err != nil {
			t.Fatal(err)
		}
		var bookmark *element.Bookmark
		var link *element.Link
		for _, sec := range dst.Sections()[1:] {
			for _, el := range sec.Elements() {
				switch v := el.(type) {
				case *element.Bookmark:
					bookmark = v
				case *element.Link:
					link = v
				}
			}
		}
		if link == nil || bookmark == nil {
			t.Fatalf("forward=%v: merged document is missing its link or bookmark", forward)
		}
		if link.Target != bookmark.Name || link.Target == "target" {
			t.Fatalf("forward=%v: link=%q bookmark=%q", forward, link.Target, bookmark.Name)
		}
		if original.Target != "target" {
			t.Fatal("source mutated")
		}
	}
}
