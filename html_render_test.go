package word

import (
	"bytes"
	"errors"
	"html"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/yunkeweb/go-word/element"
	"github.com/yunkeweb/go-word/style"
)

func FuzzRenderHTMLTextEscaping(f *testing.F) {
	for _, seed := range []string{"", "plain", "<script>alert(1)</script>", "quotes & ampersands", "中文"} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, input string) {
		doc := New()
		doc.AddSection().AddText(input)
		got, err := doc.RenderHTML(HTMLOptions{})
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(got), html.EscapeString(input)) {
			t.Fatalf("escaped text missing from HTML: %q", got)
		}
		assertHTMLStructure(t, string(got))
	})
}

func TestRenderHTMLBasicDocument(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	sec.AddTitle("A <title>", 2)
	run := sec.AddTextRun(style.Paragraph{Alignment: style.JcCenter})
	run.AddText("Hello & <world>", style.Font{Bold: true, Color: "#123456"})
	run.AddLink("https://example.test/?a=1&b=2", "link & text")
	got, err := doc.RenderHTML(HTMLOptions{Standalone: true, IncludeCSS: true})
	if err != nil {
		t.Fatal(err)
	}
	s := string(got)
	for _, want := range []string{"<!doctype html>", "<h2>A &lt;title&gt;</h2>", "Hello &amp; &lt;world&gt;", "font-weight:700", "href=\"https://example.test/?a=1&amp;b=2\"", "<style>"} {
		if !strings.Contains(s, want) {
			t.Fatalf("HTML missing %q: %s", want, s)
		}
	}
}

func TestRenderHTMLStandaloneUsesSectionPageGeometry(t *testing.T) {
	doc := New()
	sec := doc.AddSection(style.Section{
		PageSizeW:    12000,
		PageSizeH:    16800,
		MarginTop:    1440,
		MarginRight:  1800,
		MarginBottom: 1440,
		MarginLeft:   1800,
	})
	sec.AddText("page geometry")
	got, err := doc.RenderHTML(HTMLOptions{Standalone: true, IncludeCSS: true})
	if err != nil {
		t.Fatal(err)
	}
	s := html.UnescapeString(string(got))
	for _, want := range []string{
		"width:800px;min-height:1120px",
		"padding:96px 120px 96px 120px",
		"<p>page geometry</p>",
		"background:#e7e7e7",
	} {
		if !strings.Contains(s, want) {
			t.Fatalf("HTML missing %q: %s", want, s)
		}
	}
	assertHTMLStructure(t, s)
}

func TestRenderHTMLParagraphPaginationStyles(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	sec.AddText("keep with next", style.Font{}, style.Paragraph{KeepNext: true})
	sec.AddText("keep lines", style.Font{}, style.Paragraph{KeepLines: true})
	falseWidowControl := false
	sec.AddText("allow widow", style.Font{}, style.Paragraph{WidowControl: &falseWidowControl})
	trueWidowControl := true
	sec.AddText("control widow", style.Font{}, style.Paragraph{WidowControl: &trueWidowControl})
	got, err := doc.RenderHTML(HTMLOptions{})
	if err != nil {
		t.Fatal(err)
	}
	s := string(got)
	for _, want := range []string{
		`<p style="break-after:avoid">keep with next</p>`,
		`<p style="break-inside:avoid">keep lines</p>`,
		`<p style="widows:1;orphans:1">allow widow</p>`,
		`<p style="widows:2;orphans:2">control widow</p>`,
	} {
		if !strings.Contains(s, want) {
			t.Fatalf("HTML missing %q: %s", want, s)
		}
	}
}

func TestRenderHTMLStandaloneHeaderFooterContainers(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	sec.AddHeader(element.HeaderFirst).AddText("first header")
	sec.AddFooter(element.HeaderEven).AddText("even footer")
	sec.AddText("body")
	got, err := doc.RenderHTML(HTMLOptions{
		Standalone:            true,
		IncludeCSS:            true,
		IncludeHeadersFooters: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	s := string(got)
	for _, want := range []string{
		`<header data-type="first" class="goword-header" data-section="1">`,
		`<footer data-type="even" class="goword-footer" data-section="1">`,
	} {
		if !strings.Contains(s, want) {
			t.Fatalf("HTML missing %q: %s", want, s)
		}
	}
	fragment, err := doc.RenderHTML(HTMLOptions{IncludeHeadersFooters: true})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(fragment), "goword-header") || strings.Contains(string(fragment), "data-section=\"1\"") {
		t.Fatalf("fragment output unexpectedly gained page-only header metadata: %s", fragment)
	}
}

func TestRenderHTMLStandaloneRepeatsOnlySimpleDefaultHeaderFooter(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	sec.AddHeader().AddText("header")
	sec.AddFooter().AddText("footer")
	got, err := doc.RenderHTML(HTMLOptions{Standalone: true, IncludeCSS: true, IncludeHeadersFooters: true})
	if err != nil {
		t.Fatal(err)
	}
	s := string(got)
	if !strings.Contains(s, `<body class="goword-document" data-repeat-header="true" data-repeat-footer="true">`) {
		t.Fatalf("simple default header/footer did not enable print repetition: %s", s)
	}

	doc = New()
	sec = doc.AddSection()
	sec.AddHeader(element.HeaderFirst).AddText("first header")
	sec.AddFooter(element.HeaderEven).AddText("even footer")
	got, err = doc.RenderHTML(HTMLOptions{Standalone: true, IncludeCSS: true, IncludeHeadersFooters: true})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(got), `<body class="goword-document" data-repeat-header=`) || strings.Contains(string(got), `<body class="goword-document" data-repeat-footer=`) {
		t.Fatalf("variant header/footer incorrectly enabled repetition: %s", got)
	}
}

func TestRenderHTMLTablePaginationAndFixedLayout(t *testing.T) {
	doc := New()
	table := doc.AddSection().AddTable(style.Table{Width: 3000, Layout: "fixed"})
	header := table.AddRow()
	header.SetHeader(true)
	header.AddCell(1500).AddText("header")
	body := table.AddRow()
	body.SetCantSplit(true)
	body.AddCell(1500).AddText("body")
	got, err := doc.RenderHTML(HTMLOptions{})
	if err != nil {
		t.Fatal(err)
	}
	s := string(got)
	for _, want := range []string{
		`<table style="border-collapse:collapse;width:200px;table-layout:fixed">`,
		`<thead><tr><td style="width:100px">`,
		`</td></tr></thead><tbody><tr style="break-inside:avoid"><td style="width:100px">`,
	} {
		if !strings.Contains(s, want) {
			t.Fatalf("HTML missing %q: %s", want, s)
		}
	}
}

func TestRenderHTMLTableBordersSpacingAndCellPadding(t *testing.T) {
	doc := New()
	table := doc.AddSection().AddTable(style.Table{
		Width:          3000,
		CellSpacingVal: 60,
		Borders:        style.Borders{Top: style.Border{Style: "single", Size: 8, Color: "112233"}},
	})
	cell := table.AddRow().AddCell(1500, style.Cell{
		PaddingTop:    60,
		PaddingRight:  120,
		PaddingBottom: 180,
		PaddingLeft:   240,
		Borders:       style.Borders{Bottom: style.Border{Style: "single", Size: 4, Color: "445566"}},
	})
	cell.AddText("padded")
	got, err := doc.RenderHTML(HTMLOptions{})
	if err != nil {
		t.Fatal(err)
	}
	s := string(got)
	for _, want := range []string{
		`border-spacing:4px`,
		`border-top:1pt solid #112233`,
		`padding-top:4px`,
		`padding-right:8px`,
		`padding-bottom:12px`,
		`padding-left:16px`,
		`border-bottom:0.5pt solid #445566`,
	} {
		if !strings.Contains(s, want) {
			t.Fatalf("HTML missing %q: %s", want, s)
		}
	}
}

func TestRenderHTMLImageLayoutStylesAndEMUFallback(t *testing.T) {
	doc := New()
	doc.AddSection().AddImageBytes("photo.png", []byte{1, 2, 3}, style.Image{
		WidthEMU:      914400,
		HeightEMU:     457200,
		Alignment:     "right",
		WrappingStyle: style.WrappingSquare,
		MarginTop:     4,
		MarginRight:   8,
		OffsetX:       2,
		OffsetY:       3,
		AltText:       "photo",
	})
	got, err := doc.RenderHTML(HTMLOptions{})
	if err != nil {
		t.Fatal(err)
	}
	s := string(got)
	for _, want := range []string{
		`width="96"`,
		`height="48"`,
		`style="float:right;margin-top:4px;margin-right:8px;position:relative;left:2px;top:3px"`,
	} {
		if !strings.Contains(s, want) {
			t.Fatalf("HTML missing %q: %s", want, s)
		}
	}
}

func TestRenderHTMLUnknownImageWrappingProducesDiagnostic(t *testing.T) {
	doc := New()
	doc.AddSection().AddImageBytes("photo.png", []byte{1, 2, 3}, style.Image{WrappingStyle: "unsupported-wrap"})
	result, err := doc.RenderHTMLWithDiagnostics(HTMLOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Diagnostics) != 1 || result.Diagnostics[0].ElementType != "Image" || !strings.Contains(result.Diagnostics[0].Message, "wrapping") {
		t.Fatalf("diagnostics=%+v", result.Diagnostics)
	}
	if !strings.Contains(string(result.HTML), "<img") {
		t.Fatalf("image was dropped after layout diagnostic: %s", result.HTML)
	}
	if _, err := doc.RenderHTML(HTMLOptions{Strict: true}); err == nil {
		t.Fatal("strict mode accepted unsupported image wrapping")
	}
}

func TestRenderHTMLAdvancedFontFallbackAndMetrics(t *testing.T) {
	doc := New()
	doc.AddSection().AddText("metrics", style.Font{
		Name:         "Primary Font",
		FallbackFont: "Fallback Font",
		Spacing:      30,
		Scale:        80,
		Position:     2,
		WhiteSpace:   "preserve",
	})
	got, err := doc.RenderHTML(HTMLOptions{})
	if err != nil {
		t.Fatal(err)
	}
	s := html.UnescapeString(string(got))
	for _, want := range []string{
		`font-family:"Primary Font","Fallback Font"`,
		"letter-spacing:2px",
		"font-stretch:80%",
		"vertical-align:1pt",
		"white-space:pre-wrap",
	} {
		if !strings.Contains(s, want) {
			t.Fatalf("HTML missing %q: %s", want, s)
		}
	}
}

func TestWordThemeFontsAndColorsResolveForHTML(t *testing.T) {
	themeXML := `<a:theme xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main"><a:themeElements><a:clrScheme name="x"><a:accent1><a:srgbClr val="336699"/></a:accent1></a:clrScheme><a:fontScheme name="x"><a:majorFont><a:latin typeface="Aptos Display"/><a:ea typeface="Noto Sans CJK"/></a:majorFont><a:minorFont><a:latin typeface="Aptos"/><a:ea typeface="Noto Sans CJK SC"/></a:minorFont></a:fontScheme></a:themeElements></a:theme>`
	theme, err := parseWordTheme([]byte(themeXML))
	if err != nil {
		t.Fatal(err)
	}
	sheet := &wordStyleSheet{theme: theme}
	font := sheet.font(style.Font{}, wordRunProperties{
		Fonts: &struct {
			ASCII         string `xml:"ascii,attr"`
			HighANSI      string `xml:"hAnsi,attr"`
			EastAsia      string `xml:"eastAsia,attr"`
			ASCIITheme    string `xml:"asciiTheme,attr"`
			HighANSITheme string `xml:"hAnsiTheme,attr"`
			EastAsiaTheme string `xml:"eastAsiaTheme,attr"`
		}{ASCIITheme: "majorAscii", EastAsiaTheme: "minorEastAsia"},
		Color: &wordColor{ThemeColor: "accent1"},
	})
	if font.Name != "Noto Sans CJK SC" || font.Color != "336699" {
		t.Fatalf("theme resolution = %+v", font)
	}
	major := sheet.font(style.Font{}, wordRunProperties{Fonts: &struct {
		ASCII         string `xml:"ascii,attr"`
		HighANSI      string `xml:"hAnsi,attr"`
		EastAsia      string `xml:"eastAsia,attr"`
		ASCIITheme    string `xml:"asciiTheme,attr"`
		HighANSITheme string `xml:"hAnsiTheme,attr"`
		EastAsiaTheme string `xml:"eastAsiaTheme,attr"`
	}{ASCIITheme: "majorAscii"}})
	if major.Name != "Aptos Display" {
		t.Fatalf("major theme font = %q", major.Name)
	}
	doc := New()
	doc.AddSection().AddText("theme", font)
	got, err := doc.RenderHTML(HTMLOptions{})
	if err != nil {
		t.Fatal(err)
	}
	s := html.UnescapeString(string(got))
	for _, want := range []string{`font-family:"Noto Sans CJK SC"`, `color:#336699`} {
		if !strings.Contains(s, want) {
			t.Fatalf("HTML missing %q: %s", want, s)
		}
	}
}

func TestRenderHTMLParagraphTabsAndHyphenation(t *testing.T) {
	doc := New()
	doc.AddSection().AddText("tabs", style.Font{}, style.Paragraph{
		Tabs:                []style.Tab{{Pos: 720}},
		SuppressAutoHyphens: true,
	})
	got, err := doc.RenderHTML(HTMLOptions{})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"tab-size:48px", "hyphens:none"} {
		if !strings.Contains(string(got), want) {
			t.Fatalf("HTML missing %q: %s", want, got)
		}
	}
}

func TestRenderHTMLParagraphIndentSpacingAndBorders(t *testing.T) {
	doc := New()
	doc.AddSection().AddText("layout", style.Font{}, style.Paragraph{
		Indentation: style.Indentation{Left: 360, FirstLine: 240},
		Spacing:     style.Spacing{Before: 120, After: 180, Line: 360, Rule: "exact"},
		Borders:     style.Borders{Bottom: style.Border{Style: "single", Size: 8, Color: "123456"}},
	})
	got, err := doc.RenderHTML(HTMLOptions{})
	if err != nil {
		t.Fatal(err)
	}
	s := string(got)
	for _, want := range []string{
		`margin-left:24px`,
		`text-indent:16px`,
		`margin-top:8px`,
		`margin-bottom:12px`,
		`line-height:24px`,
		`border-bottom:1pt solid #123456`,
	} {
		if !strings.Contains(s, want) {
			t.Fatalf("HTML missing %q: %s", want, s)
		}
	}
}

func TestRenderHTMLListsTablesAndImage(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	sec.AddListItem("one", 0, style.Font{}, nil, style.ListItem{ListType: style.ListTypeBullet})
	sec.AddListItem("two", 0, style.Font{}, nil, style.ListItem{ListType: style.ListTypeBullet})
	tbl := sec.AddTable(style.Table{Width: 1500})
	row := tbl.AddRow()
	row.AddCell(750).AddText("A")
	row.AddCell(750).AddText("B")
	sec.AddImageBytes("pixel.png", []byte{0x89, 0x50, 0x4e, 0x47}, style.Image{AltText: "pixel", Width: 20})
	got, err := doc.RenderHTML(HTMLOptions{})
	if err != nil {
		t.Fatal(err)
	}
	s := string(got)
	for _, want := range []string{"<ul>", "<li>one", "<table", "<td", "data:image/png;base64,", "alt=\"pixel\"", "width=\"20\""} {
		if !strings.Contains(s, want) {
			t.Fatalf("HTML missing %q: %s", want, s)
		}
	}
}

func TestRenderHTMLStrictDiagnostics(t *testing.T) {
	doc := New()
	doc.AddSection().AddChart("bar", []string{"A"}, []float64{1})
	res, err := doc.RenderHTMLWithDiagnostics(HTMLOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Diagnostics) != 1 || res.Diagnostics[0].ElementType != "Chart" {
		t.Fatalf("diagnostics=%+v", res.Diagnostics)
	}
	if !strings.Contains(string(res.HTML), "data-element=\"Chart\"") {
		t.Fatalf("fallback missing: %s", res.HTML)
	}
	if _, err := doc.RenderHTML(HTMLOptions{Strict: true}); err == nil {
		t.Fatal("strict mode accepted unsupported chart")
	}
}

func TestRenderHTMLReaderEntrypoints(t *testing.T) {
	doc := New()
	doc.AddSection().AddText("reader entry")
	raw, err := doc.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	got, err := RenderHTML(bytes.NewReader(raw), HTMLOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(got), "reader entry") {
		t.Fatalf("reader output=%s", got)
	}
}

func TestRenderHTMLRoundTripFromDocx(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	sec.AddTitle("Roundtrip", 1)
	sec.AddText("paragraph")
	sec.AddListItem("item", 0, style.Font{}, nil, style.ListItem{ListType: style.ListTypeBullet})
	tbl := sec.AddTable()
	row := tbl.AddRow()
	row.AddCell(900).AddText("cell")
	sec.AddImageBytes("pixel.png", []byte{1, 2, 3}, style.Image{AltText: "roundtrip"})
	raw, err := doc.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	loaded, err := LoadBytes(raw)
	if err != nil {
		t.Fatal(err)
	}
	got, err := loaded.RenderHTML(HTMLOptions{})
	if err != nil {
		t.Fatal(err)
	}
	s := string(got)
	for _, want := range []string{"Roundtrip", "paragraph", "<ul>", "cell", "data:image/"} {
		if !strings.Contains(s, want) {
			t.Fatalf("roundtrip HTML missing %q: %s", want, s)
		}
	}
}

func TestRenderHTMLImageURLCallback(t *testing.T) {
	doc := New()
	img := doc.AddSection().AddImageBytes("pixel.png", []byte{1, 2, 3}, nil)
	got, err := doc.RenderHTML(HTMLOptions{ImageURL: func(got *element.Image) (string, error) {
		if got != img {
			t.Fatal("unexpected image callback")
		}
		return "/assets/pixel.png", nil
	}})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(got), "src=\"/assets/pixel.png\"") {
		t.Fatalf("callback URL missing: %s", got)
	}
}

func TestRenderHTMLImageURLCallbackRejectsUnsafeURL(t *testing.T) {
	doc := New()
	doc.AddSection().AddImageBytes("pixel.png", []byte{1, 2, 3}, nil)
	res, err := doc.RenderHTMLWithDiagnostics(HTMLOptions{ImageURL: func(*element.Image) (string, error) {
		return "javascript:alert(1)", nil
	}})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Diagnostics) != 1 || !strings.Contains(res.Diagnostics[0].Message, "unsafe URL") {
		t.Fatalf("diagnostics=%+v", res.Diagnostics)
	}
	if strings.Contains(string(res.HTML), "javascript:") {
		t.Fatalf("unsafe callback URL leaked: %s", res.HTML)
	}
}

func TestRenderHTMLSanitizesURLsAndSupportsHeaders(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	sec.AddLink("javascript:alert(1)", "unsafe")
	sec.AddHeader().AddText("Header")
	sec.AddFooter().AddText("Footer")
	got, err := doc.RenderHTML(HTMLOptions{IncludeHeadersFooters: true})
	if err != nil {
		t.Fatal(err)
	}
	s := string(got)
	if strings.Contains(s, "javascript:") || !strings.Contains(s, "Header") || !strings.Contains(s, "Footer") {
		t.Fatalf("unsafe or header output: %s", s)
	}
}

func TestRenderHTMLBookmarksAndInternalLinks(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	title := sec.AddTitle("Target", 2)
	title.BookmarkName = "target-heading"
	sec.AddBookmark("target-paragraph")
	run := sec.AddTextRun()
	run.AddBookmark("target-inline")
	run.AddText("inline target")
	doc.AddHyperlinkToBookmark(nil, "jump to target", "target-paragraph")

	got, err := doc.RenderHTML(HTMLOptions{})
	if err != nil {
		t.Fatal(err)
	}
	s := string(got)
	for _, want := range []string{
		`<h2 id="target-heading">Target</h2>`,
		`<span id="target-paragraph"></span>`,
		`<span id="target-inline"></span>inline target`,
		`<a href="#target-paragraph">jump to target</a>`,
	} {
		if !strings.Contains(s, want) {
			t.Fatalf("HTML missing %q: %s", want, s)
		}
	}
	assertHTMLStructure(t, s)
}

func TestRenderHTMLTableMerges(t *testing.T) {
	doc := New()
	tbl := doc.AddSection().AddTable()
	first := tbl.AddRow()
	first.AddCell(1000, style.Cell{VMerge: "restart"}).AddText("A")
	first.AddCell(2000, style.Cell{GridSpan: 2, VMerge: "restart"}).AddText("B")
	second := tbl.AddRow()
	second.AddCell(1000, style.Cell{VMerge: "continue"})
	second.AddCell(2000, style.Cell{GridSpan: 2, VMerge: "continue"})

	got, err := doc.RenderHTML(HTMLOptions{})
	if err != nil {
		t.Fatal(err)
	}
	s := string(got)
	for _, want := range []string{`rowspan="2"`, `colspan="2"`, ">A</p>", ">B</p>"} {
		if !strings.Contains(s, want) {
			t.Fatalf("HTML missing %q: %s", want, s)
		}
	}
	if strings.Count(s, "<td") != 2 {
		t.Fatalf("vertical merge continuation cells should be omitted: %s", s)
	}
	assertHTMLStructure(t, s)
}

func TestRenderHTMLListNumberingFormats(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	sec.AddListItem("chapter", 0, nil, nil, style.ListItem{ListType: style.ListTypeNumber, Format: style.NumberUpperRoman})
	sec.AddListItem("section", 1, nil, nil, style.ListItem{ListType: style.ListTypeNumber, Format: style.NumberLowerLetter})

	got, err := doc.RenderHTML(HTMLOptions{})
	if err != nil {
		t.Fatal(err)
	}
	s := string(got)
	for _, want := range []string{`<ol type="I">`, `<ol type="a">`, "chapter", "section"} {
		if !strings.Contains(s, want) {
			t.Fatalf("HTML missing %q: %s", want, s)
		}
	}
	assertHTMLStructure(t, s)
}

func TestRenderHTMLConcurrentIsolation(t *testing.T) {
	doc := New()
	doc.AddSection().AddText("concurrent")
	var wg sync.WaitGroup
	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			got, err := doc.RenderHTML(HTMLOptions{Standalone: true})
			if err != nil || !strings.Contains(string(got), "concurrent") {
				t.Errorf("render failed: %v %s", err, got)
			}
		}()
	}
	wg.Wait()
}

func TestRenderHTMLWithOptionsAndWriteHTML(t *testing.T) {
	doc := New()
	doc.AddSection().AddText("limited")
	raw, err := doc.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	got, err := RenderHTMLWithOptions(bytes.NewReader(raw), ReadOptions{MaxArchiveSize: int64(len(raw) + 1)}, HTMLOptions{})
	if err != nil || !strings.Contains(string(got), "limited") {
		t.Fatalf("limited render failed: %v %s", err, got)
	}
	var out bytes.Buffer
	if err := doc.WriteHTML(&out, HTMLOptions{Standalone: true}); err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(out.String(), "<!doctype html>") {
		t.Fatalf("stream output=%s", out.String())
	}
}

func TestRenderHTMLProducesBalancedStructure(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	sec.AddTitle("Structure", 1)
	sec.AddText("escaped <text> & unicode ✓")
	sec.AddListItem("item", 0, style.Font{}, nil, style.ListItem{ListType: style.ListTypeBullet})
	tbl := sec.AddTable()
	tbl.AddRow().AddCell(900).AddText("cell")
	sec.AddImageBytes("pixel.png", []byte{1, 2, 3}, style.Image{AltText: "pixel"})

	got, err := doc.RenderHTML(HTMLOptions{Standalone: true, IncludeCSS: true})
	if err != nil {
		t.Fatal(err)
	}
	assertHTMLStructure(t, string(got))
}

func TestRenderHTMLLargeDocument(t *testing.T) {
	doc := buildLargeHTMLDocument(500)
	got, err := doc.RenderHTML(HTMLOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Count(string(got), "</p>") < 500 {
		t.Fatalf("large document rendered too few paragraphs: %d", strings.Count(string(got), "</p>"))
	}
	assertHTMLStructure(t, string(got))
}

func buildLargeHTMLDocument(paragraphs int) *Document {
	doc := New()
	sec := doc.AddSection()
	for i := 0; i < paragraphs; i++ {
		sec.AddText("The quick brown fox jumps over the lazy dog.", style.Font{Size: 11})
	}
	for r := 0; r < 10; r++ {
		tbl := sec.AddTable(style.Table{Width: 9000})
		row := tbl.AddRow()
		for c := 0; c < 5; c++ {
			row.AddCell(1800).AddText("cell")
		}
	}
	return doc
}

func assertHTMLStructure(t *testing.T, s string) {
	t.Helper()
	voidTags := map[string]bool{
		"area": true, "base": true, "br": true, "col": true, "embed": true,
		"hr": true, "img": true, "input": true, "link": true, "meta": true,
		"param": true, "source": true, "track": true, "wbr": true,
	}
	var stack []string
	for pos := 0; pos < len(s); {
		relStart := strings.IndexByte(s[pos:], '<')
		if relStart < 0 {
			break
		}
		start := pos + relStart
		relEnd := strings.IndexByte(s[start+1:], '>')
		if relEnd < 0 {
			t.Fatalf("unterminated tag at byte %d", start)
		}
		end := start + 1 + relEnd
		token := strings.TrimSpace(s[start+1 : end])
		pos = end + 1
		if token == "" || strings.HasPrefix(token, "!") || strings.HasPrefix(token, "?") {
			continue
		}
		if strings.HasPrefix(token, "/") {
			name := strings.Fields(strings.TrimSpace(strings.TrimPrefix(token, "/")))[0]
			if len(stack) == 0 || stack[len(stack)-1] != name {
				t.Fatalf("mismatched closing tag </%s>, stack=%v", name, stack)
			}
			stack = stack[:len(stack)-1]
			continue
		}
		selfClosing := strings.HasSuffix(token, "/")
		name := strings.Fields(strings.TrimSpace(strings.TrimSuffix(token, "/")))[0]
		if !voidTags[name] && !selfClosing {
			stack = append(stack, name)
		}
	}
	if len(stack) != 0 {
		t.Fatalf("unclosed tags: %v", stack)
	}
}

var htmlBenchmarkSink []byte

func BenchmarkRenderHTMLLargeDocument(b *testing.B) {
	doc := buildLargeHTMLDocument(2000)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		got, err := doc.RenderHTML(HTMLOptions{})
		if err != nil {
			b.Fatal(err)
		}
		htmlBenchmarkSink = got
	}
}

func TestWriteHTMLMatchesBufferedRender(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	sec.AddTitle("stream", 1)
	sec.AddText("streamed output")
	want, err := doc.RenderHTML(HTMLOptions{Standalone: true, IncludeCSS: true})
	if err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	if err := doc.WriteHTML(&out, HTMLOptions{Standalone: true, IncludeCSS: true}); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(want, out.Bytes()) {
		t.Fatalf("stream output differs from buffered output\nwant=%s\ngot=%s", want, out.Bytes())
	}
}

func TestWriteHTMLPropagatesWriterErrors(t *testing.T) {
	doc := New()
	doc.AddSection().AddText("writer failure")
	wantErr := errors.New("sink failed")
	if err := doc.WriteHTML(errorHTMLWriter{err: wantErr}, HTMLOptions{}); !errors.Is(err, wantErr) {
		t.Fatalf("error=%v, want %v", err, wantErr)
	}
	if err := doc.WriteHTML(shortHTMLWriter{}, HTMLOptions{}); !errors.Is(err, io.ErrShortWrite) {
		t.Fatalf("short write error=%v", err)
	}
}

type errorHTMLWriter struct{ err error }

func (w errorHTMLWriter) Write([]byte) (int, error) { return 0, w.err }

type shortHTMLWriter struct{}

func (shortHTMLWriter) Write(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	return len(p) - 1, nil
}

func TestRenderHTMLEntrypointParityAndDeterminism(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	sec.AddTitle("Parity", 1)
	sec.AddText("same output")
	raw, err := doc.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	opts := HTMLOptions{Standalone: true, IncludeCSS: true}
	want, err := doc.RenderHTML(opts)
	if err != nil {
		t.Fatal(err)
	}
	repeat, err := doc.RenderHTML(opts)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(want, repeat) {
		t.Fatal("repeated Document rendering is not deterministic")
	}
	readerHTML, err := RenderHTML(bytes.NewReader(raw), opts)
	if err != nil {
		t.Fatal(err)
	}
	limitedHTML, err := RenderHTMLWithOptions(bytes.NewReader(raw), ReadOptions{}, opts)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "parity.docx")
	if err := os.WriteFile(path, raw, 0o644); err != nil {
		t.Fatal(err)
	}
	fileHTML, err := RenderHTMLFile(path, opts)
	if err != nil {
		t.Fatal(err)
	}
	for name, got := range map[string][]byte{"reader": readerHTML, "limited": limitedHTML, "file": fileHTML} {
		s := string(got)
		if !strings.Contains(s, "Parity") || !strings.Contains(s, "same output") {
			t.Errorf("%s entrypoint lost document content: %s", name, s)
		}
		assertHTMLStructure(t, s)
	}
}
