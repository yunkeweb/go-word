package metadata

import "testing"

func TestCompatibility(t *testing.T) {
	c := NewCompatibility()
	if c.GetOoxmlVersion() != 12 {
		t.Fatal("default")
	}
	c.SetOoxmlVersion(15)
	if c.GetOoxmlVersion() != 15 {
		t.Fatal("set")
	}
}

func TestDocInfo(t *testing.T) {
	d := NewDocInfo()
	if d.Creator != "GoWord" || d.Created.IsZero() {
		t.Fatal("defaults")
	}
	d.SetCreator("a")
	d.SetLastModifiedBy("b")
	d.SetTitle("t")
	d.SetDescription("d")
	d.SetSubject("s")
	d.SetKeywords("k")
	d.SetCategory("c")
	d.SetCompany("co")
	d.SetManager("m")
	if d.Creator != "a" || d.Title != "t" || d.Manager != "m" {
		t.Fatal("setters")
	}
	d.SetCustomProperty("n", "string", "v")
	d.SetCustomProperty("n", "int", "1")
	if len(d.Custom) != 1 || d.Custom[0].Type != "int" {
		t.Fatalf("%v", d.Custom)
	}
	d.SetCustomProperty("other", "bool", "true")
	if len(d.Custom) != 2 {
		t.Fatal("append")
	}
}

func TestSettings(t *testing.T) {
	s := NewSettings()
	if s.DecimalSymbol != "." || s.Zoom != 100 {
		t.Fatal("defaults")
	}
	_ = FootnoteProperties{Pos: FootnotePosPageBottom, NumRestart: FootnoteRestartEachPage}
	_ = TblWidth{Type: "dxa", Value: 1}
	_ = ProofState{Spelling: ProofDirty, Grammar: ProofClean}
}
