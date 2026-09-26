package element

import (
	"github.com/yunkeweb/go-word/metadata"
	"github.com/yunkeweb/go-word/pkg/math"
	"github.com/yunkeweb/go-word/style"
	"testing"
)

func TestContainerCRUDAndAdders(t *testing.T) {
	sec := NewSection(1, nil)
	if sec.Type() != "Section" {
		t.Fatal("type")
	}
	if sec.CountElements() != 0 || sec.GetElement(0) != nil || sec.GetElement(-1) != nil {
		t.Fatal("empty")
	}
	tx := sec.AddText("a", style.Font{Bold: true}, style.Paragraph{Alignment: style.JcCenter})
	if tx.Type() != "Text" || tx.Text() != "a" {
		t.Fatal("text")
	}
	tx.SetText("b")
	tr := sec.AddTextRun(style.Paragraph{Alignment: style.JcLeft})
	tr.AddText("run")
	if tr.GetText() != "run" {
		t.Fatal("textrun")
	}
	tr.SetParagraphStyle("p")
	if tr.GetParagraphStyle() != "p" {
		t.Fatal("para style")
	}
	sec.AddTextRun()
	sec.AddTextBreak()
	sec.AddTextBreak(2)
	sec.AddTextBreak(0)
	sec.AddTextBreak(1, 0)
	sec.AddPageBreak()
	sec.AddLink("https://x", "X", style.Font{Color: "00F"}, style.Paragraph{}, true)
	sec.AddLink("https://y", "")
	sec.AddBookmark("bm")
	sec.AddTitle("H", 0)
	sec.AddTitle("H2", 2, 3)
	sec.AddListItem("li", 0)
	sec.AddListItem("li2", 1, style.Font{Bold: true}, style.Paragraph{}, style.ListTypeNumber)
	sec.AddListItemRun(0, style.ListTypeBullet, style.Paragraph{})
	tbl := sec.AddTable()
	sec.AddTable(style.Table{Width: 1})
	img := sec.AddImage("pic.jpg", style.Image{Width: 1})
	if img.Media.Ext != "jpeg" {
		t.Fatalf("ext=%s", img.Media.Ext)
	}
	sec.AddImage("pic.png")
	sec.AddImageBytes("x.png", []byte{1}, &style.Image{Height: 2})
	sec.AddImageBytes("y.png", []byte{1})
	sec.AddPreserveText("{X}")
	sec.AddCheckBox("n", "t")
	sec.AddField("PAGE", map[string]string{"a": "b"}, []string{"\\*"}, "1")
	sec.AddPageNumber()
	sec.AddNumPages()
	sec.AddLine()
	sec.AddLine(style.Line{Weight: 1})
	sec.AddShape("rect")
	sec.AddShape("rect", style.Shape{Type: "rect"})
	tb := sec.AddTextBox()
	tb.AddText("in box")
	sec.AddTextBox(style.TextBox{Width: 1})
	sec.AddFormula(math.New())
	if sec.AddMath(`\frac{a}{b}`).Type() != "Formula" {
		t.Fatal("add math")
	}
	sh := sec.AddDMLShape("rect", 100, 50, "5B9BD5", "2E75B6", 12700)
	if sh.Type() != "DMLShape" {
		t.Fatal("dml shape")
	}
	sh.AddText("inside")
	sec.SetColumns(2, 720, true)
	if sec.Style.ColsNum != 2 || !sec.Style.ColsSeparator {
		t.Fatal("columns")
	}
	sec.SetColumns(0, 0, false)
	cloned := CloneSection(sec)
	if cloned == nil || cloned.CountElements() == 0 {
		t.Fatal("clone section")
	}
	if CloneElement(nil) != nil {
		t.Fatal("clone nil")
	}
	fn := sec.AddFootnote()
	fn.AddText("fn")
	sec.AddFootnote(style.Paragraph{})
	en := sec.AddEndnote()
	en.AddText("en")
	sec.AddEndnote(style.Paragraph{})
	sdt := sec.AddSDT("rich")
	sdt.AddText("sdt")
	sec.AddTOC(nil, nil, 0, 0)
	sec.AddTOC(style.Font{Bold: true}, style.TOC{}, 1, 3)
	ch := sec.AddChart("pie", []string{"A"}, []float64{1})
	sec.AddChart("bar", []string{"A"}, []float64{1}, style.Chart{Width: 1})
	base := NewTextRun(nil)
	base.AddText("基")
	ruby := NewTextRun(nil)
	ruby.AddText("ji")
	sec.AddRuby(base, ruby, RubyProperties{Alignment: RubyAlignCenter})
	sec.AddFormField("textinput")
	cm := sec.AddComment("Ann", "A", "2020")
	cm.AddText("note")
	sec.AddObject("x.bin")
	sec.AddOLEObject("x.bin", style.Image{Width: 1})

	if sec.GetElements() == nil || len(sec.Elements()) == 0 {
		t.Fatal("elements")
	}
	if sec.GetElement(0) == nil {
		t.Fatal("get")
	}
	n := sec.CountElements()
	sec.RemoveElement(0)
	if sec.CountElements() != n-1 {
		t.Fatal("remove index")
	}
	sec.RemoveElement(-1)
	sec.RemoveElement(999)
	el := sec.GetElement(0)
	sec.RemoveElement(el)
	sec.RemoveElement(NewText("missing", nil, nil))
	_ = tbl
	_ = ch
	_ = cm
}

func TestSectionHeaderFooter(t *testing.T) {
	sec := NewSection(2, style.Section{Orientation: style.OrientationLandscape})
	sec2 := NewSection(3, &style.Section{PageSizeW: 1})
	NewSection(4, "ignore")
	if sec2.Style.PageSizeW != 1 {
		t.Fatal("ptr style")
	}
	h := sec.AddHeader()
	if h.Type() != "Header" || h.GetType() != HeaderAuto {
		t.Fatal("header default")
	}
	h.SetType("bogus")
	if h.GetType() != HeaderAuto {
		t.Fatal("invalid type")
	}
	h.SetType(HeaderFirst)
	if h.FirstPage() != HeaderFirst {
		t.Fatal("first")
	}
	if h.EvenPage() != HeaderEven {
		t.Fatal("even")
	}
	if h.ResetType() != HeaderAuto {
		t.Fatal("reset")
	}
	h2 := sec.AddHeader(HeaderFirst)
	wm := h2.AddWatermark("w.png")
	if !wm.IsWatermark {
		t.Fatal("watermark")
	}
	h2.AddWatermark("w2.png", style.Image{Width: 10})
	f := sec.AddFooter()
	f.SetType(HeaderEven)
	if f.GetType() != HeaderEven || f.Type() != "Footer" {
		t.Fatal("footer")
	}
	if f.FirstPage() != HeaderFirst {
		t.Fatal("footer first")
	}
	if f.EvenPage() != HeaderEven {
		t.Fatal("footer even")
	}
	if f.ResetType() != HeaderAuto {
		t.Fatal("footer reset")
	}
	sec.AddFooter("")
	sec.AddFooter(HeaderFirst)
	if !sec.HasDifferentFirstPage() {
		t.Fatal("first page header")
	}
	sec3 := NewSection(5, nil)
	if sec3.HasDifferentFirstPage() {
		t.Fatal("no first")
	}
	ff := sec3.AddFooter(HeaderFirst)
	if !sec3.HasDifferentFirstPage() {
		t.Fatal("first page footer")
	}
	_ = ff
	fp := &metadata.FootnoteProperties{Pos: metadata.FootnotePosPageBottom}
	sec.SetFootnoteProperties(fp)
	if sec.GetFootnoteProperties() != fp {
		t.Fatal("fn props")
	}
	if len(sec.GetHeaders()) == 0 || len(sec.GetFooters()) == 0 {
		t.Fatal("get hf")
	}
}

func TestTableHelpers(t *testing.T) {
	if NewTable("Grid").Style.StyleName != "Grid" {
		t.Fatal("name")
	}
	if NewTable(style.Table{Width: 3}).Style.Width != 3 {
		t.Fatal("val")
	}
	st := &style.Table{Width: 4}
	if NewTable(st).Style.Width != 4 {
		t.Fatal("ptr")
	}
	NewTable((*style.Table)(nil))
	NewTable(1)
	tbl := NewTable(nil)
	if tbl.Type() != "Table" {
		t.Fatal("type")
	}
	if tbl.FindFirstDefinedCellWidths() != nil {
		t.Fatal("empty widths")
	}
	if tbl.CountColumns() != 0 {
		t.Fatal("cols")
	}
	c := tbl.AddCell(100)
	if c.Type() != "Cell" || c.Width != 100 {
		t.Fatal("addcell creates row")
	}
	r := tbl.AddRow(200, 1)
	if r.Type() != "Row" || r.GetHeight() != 200 {
		t.Fatal("row")
	}
	r.AddCell(50, style.Cell{GridSpan: 2, Width: 0})
	r.AddCell(60, &style.Cell{Width: 60})
	r.AddCell(70, (*style.Cell)(nil))
	empty := tbl.AddRow()
	_ = empty
	tbl.SetWidth(5000)
	if tbl.GetWidth() != 5000 || len(tbl.GetRows()) == 0 {
		t.Fatal("width/rows")
	}
	if tbl.CountColumns() < 3 {
		t.Fatalf("cols=%d", tbl.CountColumns())
	}
	w := tbl.FindFirstDefinedCellWidths()
	if len(w) != 1 || w[0] != 100 {
		t.Fatalf("widths=%v", w)
	}
	tbl2 := NewTable(nil)
	tbl2.AddRow()
	if tbl2.FindFirstDefinedCellWidths() != nil {
		t.Fatal("empty cells")
	}
	_ = r.GetCells()
}

func TestImageLinkTextNormalize(t *testing.T) {
	img := NewImage("a.JPG", style.Image{Name: "n", AltText: "a"})
	if img.Type() != "Image" || img.GetSource() != "a.JPG" || img.GetName() != "n" {
		t.Fatal("image")
	}
	img.SetName("")
	if img.GetName() != "a.JPG" {
		t.Fatal("fallback name")
	}
	img.SetAltText("b")
	if img.GetAltText() != "b" {
		t.Fatal("alt")
	}
	img.SetIsWatermark(true)
	if !img.IsWatermark || !img.Style.IsWatermark {
		t.Fatal("wm")
	}
	NewImage("x.png", &style.Image{Width: 1})
	NewImage("x.png", (*style.Image)(nil))
	NewImage("x.png", "nope")
	ib := NewImageBytes("z.png", []byte{1, 2}, nil)
	if len(ib.Data) != 2 {
		t.Fatal("bytes")
	}

	l := NewLink("t", "", "font", "para", false)
	if l.Text != "t" || l.Type() != "Link" {
		t.Fatal("link")
	}

	tx := NewText("hi", style.Font{Bold: true}, style.Paragraph{Alignment: "left"})
	if tx.Font() == nil || !tx.Font().Bold {
		t.Fatal("font val")
	}
	tx2 := NewText("hi", &style.Font{Italic: true}, &style.Paragraph{KeepNext: true})
	if tx2.Font() == nil || tx2.Paragraph() == nil {
		t.Fatal("ptr")
	}
	tx3 := NewText("hi", "rStyle", "pStyle")
	if tx3.FontName() != "rStyle" || tx3.ParagraphName() != "pStyle" {
		t.Fatal("names")
	}
	if tx3.Font() != nil || tx3.Paragraph() != nil {
		t.Fatal("nil structs")
	}
	tx4 := NewText("hi", "", "")
	if tx4.FontStyle != nil || tx4.ParagraphStyle != nil {
		t.Fatal("empty string")
	}
	tx5 := NewText("hi", (*style.Font)(nil), (*style.Paragraph)(nil))
	if tx5.FontStyle != nil {
		t.Fatal("nil ptr")
	}
	tx6 := NewText("hi", 1, 2)
	if tx6.FontStyle != 1 {
		t.Fatal("passthrough")
	}
	if NewText("hi", nil, nil).FontName() != "" || NewText("hi", nil, nil).ParagraphName() != "" {
		t.Fatal("empty names")
	}

	tr := NewTextRun(nil)
	tr.AddText("a")
	tr.AddLink("u", "L")
	if tr.GetText() != "aL" || tr.Type() != "TextRun" {
		t.Fatal("getText")
	}

	ptrFont := NewText("hi", nil, nil)
	ptrFont.FontStyle = &style.Font{Bold: true}
	if f := ptrFont.Font(); f == nil || !f.Bold {
		t.Fatal("font pointer")
	}
	ptrPara := NewText("hi", nil, nil)
	ptrPara.ParagraphStyle = &style.Paragraph{KeepNext: true}
	if p := ptrPara.Paragraph(); p == nil || !p.KeepNext {
		t.Fatal("paragraph pointer")
	}
}

func TestMiscTypes(t *testing.T) {
	if (&TextBreak{}).Type() != "TextBreak" {
		t.Fatal("tb")
	}
	if (&PageBreak{}).Type() != "PageBreak" {
		t.Fatal("pb")
	}
	if (&Bookmark{Name: "n"}).Type() != "Bookmark" {
		t.Fatal("bm")
	}
	if NewTitle("t", 1).Type() != "Title" {
		t.Fatal("title")
	}
	if NewListItem("t", 0, nil, nil, nil).Type() != "ListItem" {
		t.Fatal("li")
	}
	if NewListItemRun(0, nil, nil).Type() != "ListItemRun" {
		t.Fatal("lir")
	}
	if NewPreserveText("x", nil, nil).Type() != "PreserveText" {
		t.Fatal("pt")
	}
	if NewCheckBox("n", "t", nil, nil).Type() != "CheckBox" {
		t.Fatal("cb")
	}
	if NewField("PAGE", nil, nil, "").Type() != "Field" {
		t.Fatal("field")
	}
	if (&Line{}).Type() != "Line" {
		t.Fatal("line")
	}
	if (&Shape{}).Type() != "Shape" {
		t.Fatal("shape")
	}
	if NewTextBox().Type() != "TextBox" {
		t.Fatal("tbox")
	}
	if (&Formula{}).Type() != "Formula" {
		t.Fatal("formula")
	}
	if NewFootnote(nil).Type() != "Footnote" {
		t.Fatal("fn")
	}
	if NewEndnote(nil).Type() != "Endnote" {
		t.Fatal("en")
	}
	if (&SDT{}).Type() != "SDT" {
		t.Fatal("sdt")
	}
	if (&TOC{}).Type() != "TOC" {
		t.Fatal("toc")
	}
	if (&TextWatermark{Text: "x"}).Type() != "TextWatermark" {
		t.Fatal("text watermark")
	}
	ch := &Chart{}
	if ch.Type() != "Chart" {
		t.Fatal("chart")
	}
	ch.AddSeries([]string{"A"}, []float64{1})
	ch.AddSeries([]string{"B"}, []float64{2}, "s2")
	if ch.SeriesName != "" || len(ch.Series) != 2 {
		t.Fatal("series first nameless")
	}
	ch2 := &Chart{}
	ch2.AddSeries([]string{"A"}, []float64{1}, "s1")
	if ch2.SeriesName != "s1" {
		t.Fatal("named series")
	}
	if (&Ruby{}).Type() != "Ruby" {
		t.Fatal("ruby")
	}
	if NewFormField("checkbox", nil, nil).Type() != "FormField" {
		t.Fatal("ff")
	}
	if NewOLEObject("x", nil).Type() != "OLEObject" {
		t.Fatal("ole")
	}
	if (&Comment{}).Type() != "Comment" {
		t.Fatal("comment")
	}
	if (&TrackChange{}).Type() != "TrackChange" {
		t.Fatal("tc")
	}
}

func TestBaseTrackChangeAndParent(t *testing.T) {
	tx := NewText("x", nil, nil)
	tx.SetChangeInfo("ins", "Ann", "2020")
	if tx.GetTrackChange() == nil || tx.GetTrackChange().Author != "Ann" {
		t.Fatal("change")
	}
	tc := &TrackChange{ChangeType: "del"}
	tx.SetTrackChange(tc)
	if tx.GetTrackChange() != tc {
		t.Fatal("set tc")
	}
	c := &Comment{Author: "A"}
	tx.SetCommentRangeStart(c)
	tx.SetCommentRangeEnd(c)
	if tx.GetCommentRangeStart() != c || tx.GetCommentRangeEnd() != c {
		t.Fatal("comment range")
	}
	sec := NewSection(1, nil)
	sec.AddText("p")
	if sec.GetElement(0).Parent() == nil {
		t.Fatal("parent")
	}
	setParent(nil, sec)
	setParent(tx, sec)
	if tx.Parent() != sec {
		t.Fatal("setParent")
	}
}

func TestContainerType(t *testing.T) {
	c := &Container{Kind: "X"}
	if c.Type() != "X" {
		t.Fatal("kind")
	}
}
