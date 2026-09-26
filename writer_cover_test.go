package word

import (
	"bytes"
	"image"
	"image/color"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/yunkeweb/go-word/element"
	"github.com/yunkeweb/go-word/metadata"
	"github.com/yunkeweb/go-word/pkg/common"
	"github.com/yunkeweb/go-word/pkg/math"
	"github.com/yunkeweb/go-word/style"
)

func pngBytes(t *testing.T) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 4, 4))
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func jpegBytes(t *testing.T) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 4, 4))
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, nil); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func gifBytes(t *testing.T) []byte {
	t.Helper()
	img := image.NewPaletted(image.Rect(0, 0, 2, 2), color.Palette{color.Black, color.White})
	var buf bytes.Buffer
	if err := gif.Encode(&buf, img, nil); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func TestWriterAllElementsAndSettings(t *testing.T) {
	doc := New()
	info := doc.GetDocInfo()
	info.SetTitle("T")
	info.SetSubject("S")
	info.SetKeywords("k")
	info.SetDescription("d")
	info.SetCategory("c")
	info.SetCompany("co")
	info.SetManager("m")
	info.Created = time.Time{}
	info.Modified = time.Time{}
	info.SetCustomProperty("str", "string", "v")
	info.SetCustomProperty("i", "int", "1")
	info.SetCustomProperty("f", "float", "1.5")
	info.SetCustomProperty("b", "bool", "true")

	doc.AddFontStyle("r", style.Font{Bold: true}, style.Paragraph{KeepNext: true})
	doc.AddFontStyle("r2", &style.Font{Italic: true})
	doc.AddFontStyle("r3", "nope")
	doc.AddParagraphStyle("p", style.Paragraph{Alignment: style.JcCenter})
	doc.AddParagraphStyle("p2", &style.Paragraph{KeepLines: true})
	doc.AddTableStyle("tbl", style.Table{Width: 1000}, style.Table{Width: 1})
	doc.AddTableStyle("tbl2", &style.Table{Width: 2}, &style.Table{Width: 3})
	doc.AddTableStyle("tbl3", "x")
	doc.AddNumberingStyle("num", style.Numbering{
		Levels: []style.NumberingLevel{{Format: style.NumberDecimal, Text: "%1.", Alignment: "left", Start: 1}},
	})
	doc.AddNumberingStyle("num2", &style.Numbering{Type: "multilevel", Levels: []style.NumberingLevel{{}}})
	doc.AddNumberingStyle("num3", 1)
	doc.AddTitleStyle(0, style.Font{Bold: true})
	for d := 1; d <= 9; d++ {
		doc.AddTitleStyle(d, style.Font{Bold: true}, style.Paragraph{Spacing: style.Spacing{After: 100}})
	}
	doc.AddLinkStyle("", style.Font{Color: "00F"})
	doc.AddLinkStyle("MyLink", &style.Font{Underline: style.UnderlineSingle})
	doc.SetDefaultParagraphStyle(style.Paragraph{Alignment: style.JcLeft})
	if doc.GetStyles() == nil || doc.GetStyle("r") == nil {
		t.Fatal("styles")
	}

	s := doc.Settings()
	s.ZoomPreset = style.ZoomBestFit
	s.HideSpellingErrors = true
	s.HideGrammaticalErrors = true
	s.TrackRevisions = true
	s.UpdateFields = true
	s.MirrorMargins = true
	s.EvenAndOddHeaders = true
	s.ThemeFontLang = "en-US"
	s.ThemeFont = &style.Language{EastAsia: "zh-CN", Bidirectional: "ar-SA", Latin: "en-US"}
	s.ProofState = metadata.ProofState{Spelling: metadata.ProofDirty, Grammar: metadata.ProofClean}
	s.DoNotTrackMoves = true
	s.DoNotTrackFormatting = true
	s.AutoHyphenation = true
	s.ConsecutiveHyphenLimit = 2
	s.HyphenationZone = 360
	s.DoNotHyphenateCaps = true
	s.BookFoldPrinting = true
	s.DecimalSymbol = ","
	yes, no := true, false
	s.RevisionView = &metadata.TrackChangesView{Markup: &yes, Comments: &no, InsDel: &yes, Formatting: &no, InkAnnotations: &yes}
	s.DocumentProtection = &metadata.Protection{Editing: "readOnly", Hash: "abc"}
	doc.Compatibility().SetOoxmlVersion(0)

	sec := doc.AddSection(style.Section{
		Orientation:        style.OrientationLandscape,
		PageNumberingStart: 3,
		BreakType:          "continuous",
		ColsNum:            0,
		PageSizeW:          10000,
		PageSizeH:          8000,
		MarginTop:          720,
	})
	sec2 := doc.AddSection()
	sec2.AddText("second")

	hdr := sec.AddHeader()
	hdr.HeaderType = ""
	hdr.AddText("header")
	hdr.AddPreserveText("PAGE")
	sec.AddHeader(element.HeaderFirst).AddText("first")
	sec.AddHeader(element.HeaderEven).AddText("even")
	ftr := sec.AddFooter()
	ftr.HeaderType = ""
	ftr.AddText("footer")
	sec.AddFooter(element.HeaderFirst).AddText("ff")

	tx := sec.AddText("styled", style.Font{
		Name: "Arial", Size: 12, Color: "FF0000", Bold: true, Italic: true,
		SmallCaps: true, AllCaps: true, Strikethrough: true, DoubleStrikethrough: true,
		Hidden: true, FgColor: "yellow", Underline: style.UnderlineSingle,
		SuperScript: true, BgColor: "FFFF00", RTL: true, Lang: "en-US", NoProof: true,
		Spacing: 20, Hint: "eastAsia", Shading: style.Shading{Val: "solid", Fill: "AAA", Color: "000"},
	}, style.Paragraph{
		KeepNext: true, KeepLines: true, PageBreakBefore: true, Alignment: style.JcCenter,
		OutlineLevel: 1, Bidi: true, ContextualSpacing: true, TextAlignment: style.TextAlignCenter,
		SuppressAutoHyphens: true, NumStyle: "num", NumLevel: 1,
		Spacing:     style.Spacing{Before: 10, After: 20, Line: 240},
		Indentation: style.Indentation{Left: 1, Right: 2, FirstLine: 3, Hanging: 4, FirstLineChars: 5},
		Borders:     style.Borders{Top: style.Border{Size: 4, Color: "FF0000"}},
		Shading:     style.Shading{Fill: "EEEEEE"},
		Tabs:        []style.Tab{{Val: "right", Leader: "dot", Pos: 1000}},
	})
	wc := false
	tx2 := sec.AddText("widow", nil, style.Paragraph{WidowControl: &wc, Spacing: style.Spacing{Line: 240, Rule: "exact"}})
	_ = tx2
	cmt := sec.AddComment("Ann", "A", "2020-01-01T00:00:00Z")
	cmt.AddText("note")
	tx.SetCommentRangeStart(cmt)
	tx.SetCommentRangeEnd(cmt)

	tr := sec.AddTextRun("p")
	tr.AddText("run", "r")
	tr.AddText("font", &style.Font{Bold: true})
	sec.AddTextBreak(1)
	sec.AddPageBreak()
	sec.AddLink("https://example.com", "ex")
	sec.AddLink("bm", "internal", style.Font{Color: "00F"}, nil, true)
	sec.AddLink("https://x", "named", "MyLink")
	sec.AddBookmark("bm")
	sec.AddTitle("H1", 1)
	sec.AddListItem("bullet", 0, nil, nil, style.ListTypeBullet)
	sec.AddListItem("num", 0, nil, nil, style.ListTypeNumber)
	sec.AddListItem("named", 0, nil, nil, "num")
	sec.AddListItem("li", 0, nil, style.Paragraph{Alignment: style.JcLeft}, style.ListItem{NumId: 2, Depth: 1, Format: style.NumberDecimal, ListType: style.ListTypeNumber})
	lir := sec.AddListItemRun(0, style.ListTypeNumber, style.Paragraph{KeepNext: true})
	lir.AddText("rich list")

	tbl := sec.AddTable(style.Table{
		StyleName: "tbl", Width: 5000, Alignment: style.JcTableCenter, Layout: "fixed",
		CellMarginTop: 10, CellMarginLeft: 10, CellMarginRight: 10, CellMarginBottom: 10,
		Shading: style.Shading{Fill: "DDDDDD"}, Indent: 100,
		Borders: style.Borders{
			Top: style.Border{Style: "single", Size: 4}, Left: style.Border{Style: "single"},
			Right: style.Border{Color: "F00"}, Bottom: style.Border{Style: "dashed", Size: 8, Color: "0F0", Space: 1},
			InsideH: style.Border{Style: "single"}, InsideV: style.Border{Style: "single"},
		},
		Position: &style.TablePosition{
			LeftFromText: 1, RightFromText: 2, TopFromText: 3, BottomFromText: 4,
			VertAnchor: "page", HorzAnchor: "margin", TblpXSpec: "center", TblpX: 5, TblpYSpec: "top", TblpY: 6,
		},
	})
	row := tbl.AddRow(300)
	row.Style.Header = true
	row.Style.CantSplit = true
	row.Style.Rule = "exact"
	cell := row.AddCell(2000, style.Cell{
		VAlign: style.VAlignCenter, GridSpan: 2, VMerge: "restart", BgColor: "FFFF00",
		NoWrap: true, PaddingTop: 1, PaddingLeft: 2, PaddingRight: 3, PaddingBottom: 4,
		TextDir: "tbRl", Borders: style.Borders{Top: style.Border{Style: "single"}},
	})
	cell.AddText("c1")
	row.AddCell(1000, style.Cell{VMerge: "continue", Width: 1000, Unit: "dxa"})
	tbl.AddRow().AddCell(0)

	pngData := pngBytes(t)
	sec.AddImageBytes("dot.png", pngData, style.Image{Width: 4, Height: 4, Name: "Dot", AltText: "a"})
	sec.AddImageBytes("dot.jpeg", jpegBytes(t))
	sec.AddImageBytes("dot.gif", gifBytes(t))
	sec.AddImageBytes("dot.emf", []byte("not-an-image"))
	sec.AddImageBytes("dot.wmf", []byte("not-an-image"))
	sec.AddImageBytes("dot.bin", []byte("x"))
	sec.AddImageBytes("dot.jpg", pngData)
	dir := t.TempDir()
	imgPath := filepath.Join(dir, "f.png")
	if err := os.WriteFile(imgPath, pngData, 0o644); err != nil {
		t.Fatal(err)
	}
	sec.AddImage(imgPath)
	sec.AddImage("missing.png")
	inlineRun := sec.AddTextRun()
	inlineRun.AddImageBytes("in.png", pngData)

	cb := sec.AddCheckBox("n", "check")
	cb.Checked = true
	sec.AddField("DATE", nil, []string{`\@ "yyyy"`}, "now")
	sec.AddField("", nil, nil, "")
	fn := sec.AddFootnote()
	fn.AddText("footnote body")
	en := sec.AddEndnote()
	en.AddText("endnote body")
	m := math.New()
	m.Add(math.NewFraction(math.NewNumeric("1"), math.NewNumeric("2")))
	sec.AddFormula(m)
	sec.AddFormula(nil)
	sec.AddTOC(nil, nil, 1, 9)
	sdt := sec.AddSDT("rich")
	sdt.Alias = "a"
	sdt.Tag = "t"
	sdt.Value = "v"
	sdt.AddText("sdt body")
	emptySDT := sec.AddSDT("plain")
	emptySDT.Value = "empty"
	tbox := sec.AddTextBox(style.TextBox{Width: 100})
	tbox.AddText("box")
	sec.AddLine(style.Line{Width: 200, Height: 0, Color: "F00", Weight: 2})
	sec.AddShape("rect", style.Shape{Fill: style.Fill{Color: "0F0"}})
	sec.AddShape("oval")
	ff := sec.AddFormField("checkbox")
	ff.Name = "cb"
	dd := sec.AddFormField("dropdown")
	dd.Name = "dd"
	dd.Items = []string{"a", "b"}
	ti := sec.AddFormField("textinput")
	ti.Name = "ti"
	ti.Value = "v"
	rbBase := element.NewTextRun(nil)
	rbBase.AddText("汉")
	rbRuby := element.NewTextRun(nil)
	rbRuby.AddText("han")
	sec.AddRuby(rbBase, rbRuby, element.RubyProperties{Alignment: element.RubyAlignLeft})
	sec.AddRuby(nil, nil, element.RubyProperties{})

	for _, kind := range []string{"pie", "doughnut", "bar", "stacked_bar", "percent_stacked_bar", "column", "line", "area", "stacked_area", "percent_stacked_area", "radar", "scatter", "unknown"} {
		ch := sec.AddChart(kind, []string{"A", "B"}, []float64{1, 2}, style.Chart{
			Title: "T", ShowLegend: true, ThreeD: true, ShowAxisLabels: true,
			ShowGridX: true, ShowGridY: true, CategoryAxisTitle: "X", ValueAxisTitle: "Y",
			Colors: []string{"FF0000", "00FF00"}, Width: 2000000, Height: 1000000,
			DataLabelsSet: true, DataLabels: style.DefaultDataLabelOptions(),
			LegendPosition: "b", MajorTickPosition: "out",
			CategoryLabelPosition: "low", ValueLabelPosition: "high",
		})
		_ = ch
	}
	ch2 := sec.AddChart("bar", nil, nil)
	ch2.Series = nil
	ch2.Categories = []string{"A"}
	ch2.Values = []float64{1}
	ch2.Style.ShowLegend = true
	ch2.Style.ShowAxisLabels = false

	emptyCmt := sec.AddComment("Bob", "", "")
	_ = emptyCmt

	ole := sec.AddOLEObject("obj.bin")
	ole.Media.Data = []byte("oledata")
	olePath := filepath.Join(dir, "o.bin")
	if err := os.WriteFile(olePath, []byte("fileole"), 0o644); err != nil {
		t.Fatal(err)
	}
	sec.AddOLEObject(olePath)
	sec.AddOLEObject("missing.ole")

	b, err := doc.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	sxml := string(b)
	for _, need := range []string{
		"word/footnotes.xml", "word/endnotes.xml", "word/comments.xml",
		"docProps/custom.xml", "word/media/", "word/charts/chart",
	} {
		if !strings.Contains(sxml, need) {
			t.Fatalf("missing %s", need)
		}
	}
	dir2 := t.TempDir()
	if err := doc.Save(filepath.Join(dir2, "all.docx")); err != nil {
		t.Fatal(err)
	}
}

func TestWriterUnexportedHelpers(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	w := newWord2007Writer(doc)
	xw := common.NewXMLWriter()
	w.writeElement(xw, &element.TrackChange{}, false)
	w.writeChart(xw, &element.Chart{}, false)
	w.writeOLE(xw, &element.OLEObject{}, false)
	w.writeCommentRef(xw, &element.Comment{}, false)
	w.writeCommentRef(xw, &element.Comment{CommentID: 1}, true)
	w.writeImage(xw, &element.Image{})
	w.writeLink(xw, &element.Link{Target: "x", Internal: true}, true)
	w.writeLink(xw, &element.Link{Target: "x", FontStyle: &style.Font{Bold: true}}, true)
	w.writePPrFrom(xw, nil)
	w.writePPrFrom(xw, "")
	w.writePPrFromInner(xw, "name")
	w.writePPrFromInner(xw, (*style.Paragraph)(nil))
	w.writeRPrFrom(xw, nil)
	w.writeRPrFrom(xw, "")
	w.writeRPrFrom(xw, style.Font{})
	w.writeNumPr(xw, "decimal", 0)
	w.writeNumPr(xw, style.ListItem{ListType: style.ListTypeNumber}, 0)
	w.numberingID("missing")
	if mimeForExt("PNG") != "image/png" || mimeForExt("jpg") != "image/jpeg" || mimeForExt("gif") != "image/gif" {
		t.Fatal("mime")
	}
	if mimeForExt("emf") == "" || mimeForExt("wmf") == "" || mimeForExt("bin") != "application/octet-stream" {
		t.Fatal("mime2")
	}
	if nonEmpty("", "d") != "d" || nonEmpty("x", "d") != "x" {
		t.Fatal("nonEmpty")
	}
	if nonzero(0, 5) != 5 || nonzero(3, 5) != 3 {
		t.Fatal("nonzero")
	}
	if bool01(true) != "1" || bool01(false) != "0" {
		t.Fatal("bool01")
	}
	if itoa(12) != "12" || itoa64(12) != "12" {
		t.Fatal("itoa")
	}
	name, p := splitPara("x")
	if name != "x" || p != nil {
		t.Fatal("splitPara str")
	}
	name, p = splitPara(style.Paragraph{StyleName: "p"})
	if name != "p" || p == nil {
		t.Fatal("splitPara val")
	}
	name, f := splitFont("r")
	if name != "r" || f != nil {
		t.Fatal("splitFont")
	}
	name, f = splitFont((*style.Font)(nil))
	if f != nil {
		t.Fatal("nil font")
	}
	_, _ = splitFont(1)
	_, _ = splitPara(1)
	w.writeBorders(xw, "w:pBdr", style.Borders{})
	w.writeShd(xw, style.Shading{})
	w.writeTblPr(xw, style.Table{})
	w.writeContainer(xw, nil, false)
	w.relFor(&element.Image{})
	w.relFor(&element.Header{})
	w.relFor(&element.Footer{})
	w.relFor(&element.Link{})
	w.relFor(&element.Chart{})
	w.relFor(&element.OLEObject{})
	w.relFor(sec)
	if err := w.registerImage(&element.Image{}); err == nil {
		t.Fatal("empty image")
	}
	w.registerOLE(&element.OLEObject{})
	if qFormatName("table") {
		t.Fatal("qFormat")
	}
	_ = headingStyleName(9)
	_ = headingStyleName(10)

	s := doc.Settings()
	s.ZoomPreset = ""
	s.Zoom = 0
	s.ThemeFontLang = ""
	s.ThemeFont = &style.Language{Latin: "en-US"}
	s.RevisionView = &metadata.TrackChangesView{}
	s.DocumentProtection = &metadata.Protection{}
	if _, err := w.WriteTo(&bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}

	s.ThemeFont = &style.Language{}
	s.ProofState = metadata.ProofState{}
	doc.Compatibility().SetOoxmlVersion(15)
	w2 := newWord2007Writer(doc)
	if _, err := w2.WriteTo(&bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}

	type failW struct{}
	if _, err := newWord2007Writer(doc).WriteTo(failWriter{}); err == nil {
		t.Fatal("expected write error")
	}
	_ = io.Discard
}

type failWriter struct{}

func (failWriter) Write([]byte) (int, error) { return 0, io.ErrClosedPipe }

func TestSettingsXMLThemeFontBranches(t *testing.T) {
	doc := New()
	doc.AddSection().AddText("x")
	s := doc.Settings()
	s.ThemeFontLang = "en-GB"
	s.ThemeFont = &style.Language{EastAsia: "ja-JP"}
	b, err := doc.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(b, []byte("word/settings.xml")) {
		t.Fatal("settings")
	}
	s.ThemeFontLang = ""
	s.ThemeFont = &style.Language{EastAsia: "zh-CN", Bidirectional: "ar-SA"}
	if _, err := doc.Bytes(); err != nil {
		t.Fatal(err)
	}
}

func TestFontPtrParaPtrTablePtr(t *testing.T) {
	if fontPtr(nil) != nil || paraPtr(nil) != nil || tablePtr(nil) != nil || numberingPtr(nil) != nil {
		t.Fatal("nil")
	}
	if fontPtr(style.Font{Bold: true}) == nil || fontPtr(&style.Font{}) == nil {
		t.Fatal("font")
	}
	if paraPtr(style.Paragraph{}) == nil || paraPtr(&style.Paragraph{}) == nil {
		t.Fatal("para")
	}
	if tablePtr(style.Table{}) == nil || tablePtr(&style.Table{}) == nil {
		t.Fatal("table")
	}
	if numberingPtr(style.Numbering{}) == nil || numberingPtr(&style.Numbering{}) == nil {
		t.Fatal("num")
	}
}

func TestWritePPrFromNamedStyle(t *testing.T) {
	w := newWord2007Writer(New())
	xw := common.NewXMLWriter()
	w.writeRPrFrom(xw, "Hyperlink")
	w.writePPrFrom(xw, "Normal")
	p := &style.Paragraph{StyleName: "p"}
	w.writePPrFrom(xw, p)
	w.writePPrFromInner(xw, p)
	f := &style.Font{Bold: true, SubScript: true, Underline: style.UnderlineNone}
	w.writeRPrFrom(xw, f)
	w.writeRPr(xw, style.Font{SubScript: true}, "")
}
