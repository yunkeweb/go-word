package ooxml

import "testing"

func TestContentTypesAndRelationships(t *testing.T) {
	ct := NewContentTypes()
	if ct.Xmlns != NSContentTypes || len(ct.Defaults) != 2 {
		t.Fatal("defaults")
	}
	ct.AddDefault("rels", CTRels)
	if len(ct.Defaults) != 2 {
		t.Fatal("dup default")
	}
	ct.AddDefault("png", "image/png")
	if len(ct.Defaults) != 3 {
		t.Fatal("add default")
	}
	ct.AddOverride("/word/document.xml", CTDocument)
	if len(ct.Override) != 1 {
		t.Fatal("override")
	}
	r := NewRelationships()
	r.Add("rId1", NSOfficeRelOfficeDoc, "word/document.xml", "")
	if len(r.Rel) != 1 || r.Rel[0].ID != "rId1" {
		t.Fatal("rel")
	}
	_ = CoreProperties{Title: "t"}
	_ = AppProperties{Application: "GoWord"}
	_ = CustomProperties{Props: []CustomProperty{{Name: "n", Lpwstr: "v"}}}
}
