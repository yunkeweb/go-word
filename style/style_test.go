package style

import "testing"

func TestFontIsZeroAndHalfPoints(t *testing.T) {
	if !(Font{}.IsZero()) {
		t.Fatal("zero")
	}
	if (Font{Underline: UnderlineNone}).IsZero() == false {
		t.Fatal("none underline still zero")
	}
	if (Font{Name: "A"}).IsZero() {
		t.Fatal("name")
	}
	if (Font{Size: 12}).IsZero() {
		t.Fatal("size")
	}
	if (Font{Color: "FF0000"}).IsZero() {
		t.Fatal("color")
	}
	if (Font{Bold: true}).IsZero() {
		t.Fatal("bold")
	}
	if (Font{Italic: true}).IsZero() {
		t.Fatal("italic")
	}
	if (Font{Underline: UnderlineSingle}).IsZero() {
		t.Fatal("underline")
	}
	if (Font{SuperScript: true}).IsZero() {
		t.Fatal("super")
	}
	if (Font{SubScript: true}).IsZero() {
		t.Fatal("sub")
	}
	if (Font{Strikethrough: true}).IsZero() {
		t.Fatal("strike")
	}
	if (Font{DoubleStrikethrough: true}).IsZero() {
		t.Fatal("dstrike")
	}
	if (Font{SmallCaps: true}).IsZero() {
		t.Fatal("smallcaps")
	}
	if (Font{AllCaps: true}).IsZero() {
		t.Fatal("allcaps")
	}
	if (Font{Hidden: true}).IsZero() {
		t.Fatal("hidden")
	}
	if (Font{FgColor: "yellow"}).IsZero() {
		t.Fatal("fg")
	}
	if (Font{BgColor: "FFFFFF"}).IsZero() {
		t.Fatal("bg")
	}
	if (Font{Spacing: 1}).IsZero() {
		t.Fatal("spacing")
	}
	if (Font{Kerning: 1}).IsZero() {
		t.Fatal("kerning")
	}
	if (Font{Scale: 1}).IsZero() {
		t.Fatal("scale")
	}
	if (Font{RTL: true}).IsZero() {
		t.Fatal("rtl")
	}
	if (Font{Lang: "en"}).IsZero() {
		t.Fatal("lang")
	}
	if (Font{Position: 1}).IsZero() {
		t.Fatal("position")
	}
	if (Font{NoProof: true}).IsZero() {
		t.Fatal("noproof")
	}
	if (Font{Hint: "eastAsia"}).IsZero() {
		t.Fatal("hint")
	}
	if (Font{}.HalfPoints() != 0) {
		t.Fatal("half 0")
	}
	if (Font{Size: 11}).HalfPoints() != 22 {
		t.Fatal("half")
	}
}

func TestParagraphIsZeroAndSpaceAfter(t *testing.T) {
	if !(Paragraph{}.IsZero()) {
		t.Fatal("zero")
	}
	if (Paragraph{Alignment: JcCenter}).IsZero() {
		t.Fatal("align")
	}
	if (Paragraph{BasedOn: "Normal"}).IsZero() {
		t.Fatal("based")
	}
	if (Paragraph{Next: "Normal"}).IsZero() {
		t.Fatal("next")
	}
	if (Paragraph{Indentation: Indentation{Left: 1}}).IsZero() {
		t.Fatal("ind")
	}
	if (Paragraph{Spacing: Spacing{After: 1}}).IsZero() {
		t.Fatal("spacing")
	}
	f := false
	if (Paragraph{WidowControl: &f}).IsZero() {
		t.Fatal("widow")
	}
	if (Paragraph{KeepNext: true}).IsZero() {
		t.Fatal("keepnext")
	}
	if (Paragraph{KeepLines: true}).IsZero() {
		t.Fatal("keeplines")
	}
	if (Paragraph{PageBreakBefore: true}).IsZero() {
		t.Fatal("pagebreak")
	}
	if (Paragraph{Bidi: true}).IsZero() {
		t.Fatal("bidi")
	}
	if (Paragraph{OutlineLevel: 1}).IsZero() {
		t.Fatal("outline")
	}
	if (Paragraph{Tabs: []Tab{{Pos: 1}}}).IsZero() {
		t.Fatal("tabs")
	}
	if (Paragraph{Shading: Shading{Fill: "F"}}).IsZero() {
		t.Fatal("shd")
	}
	if (Paragraph{Borders: Borders{Top: Border{Style: "single"}}}).IsZero() {
		t.Fatal("bdr")
	}
	if (Paragraph{NumStyle: "n"}).IsZero() {
		t.Fatal("num")
	}
	if (Paragraph{NumLevel: 1}).IsZero() {
		t.Fatal("numlevel")
	}
	if (Paragraph{StyleName: "p"}).IsZero() {
		t.Fatal("name")
	}
	p := Paragraph{}.WithSpaceAfter(240)
	if p.Spacing.After != 240 {
		t.Fatal("space after")
	}
}

func TestNewSectionAndPaper(t *testing.T) {
	s := NewSection()
	if s.PageSizeW != DefaultPageWidth || s.Orientation != OrientationPortrait {
		t.Fatal("defaults")
	}
	p := NewPaper("")
	if p.Size != "A4" || p.Width == 0 {
		t.Fatal("default paper")
	}
	p.SetSize("Letter")
	if p.Size != "Letter" {
		t.Fatal("letter")
	}
	p.SetSize("Nope")
	if p.Size != "A4" {
		t.Fatal("unknown -> A4")
	}
	p.ApplyToSection(nil)
	sec := NewSection()
	NewPaper("Legal").ApplyToSection(&sec)
	if sec.PaperSize != "Legal" {
		t.Fatal("apply")
	}
	for _, name := range []string{"A3", "A5", "B5", "Folio"} {
		if NewPaper(name).Width == 0 {
			t.Fatalf("%s", name)
		}
	}
}

func TestChartAndTOCDefaults(t *testing.T) {
	d := DefaultDataLabelOptions()
	if !d.ShowVal || !d.ShowCatName {
		t.Fatal("labels")
	}
	toc := NewTOC()
	if toc.TabPos != 9062 || toc.TabLeader != "dot" {
		t.Fatal("toc")
	}
}
