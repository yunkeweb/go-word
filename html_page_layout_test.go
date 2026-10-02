package word

import (
	"bytes"
	"html"
	"math"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

func renderHTMLPageLayoutFixture(t *testing.T, body string, opts HTMLOptions) string {
	t.Helper()
	got, err := RenderHTML(bytes.NewReader(htmlDOCXFormattingFixture(t, body, "", "")), opts)
	if err != nil {
		t.Fatal(err)
	}
	assertHTMLStructure(t, string(got))
	return string(got)
}

func pageSectPr(width, height, top, right, bottom, left int, orientation, breakType string) string {
	attrs, breakXML := "", ""
	if orientation != "" {
		attrs = ` w:orient="` + orientation + `"`
	}
	if breakType != "" {
		breakXML = `<w:type w:val="` + breakType + `"/>`
	}
	return `<w:sectPr>` + breakXML + `<w:pgSz w:w="` + strconv.Itoa(width) + `" w:h="` + strconv.Itoa(height) + `"` + attrs + `/><w:pgMar w:top="` + strconv.Itoa(top) + `" w:right="` + strconv.Itoa(right) + `" w:bottom="` + strconv.Itoa(bottom) + `" w:left="` + strconv.Itoa(left) + `"/></w:sectPr>`
}

func TestHTMLPageLayoutStandaloneUsesPaperPreview(t *testing.T) {
	body := `<w:p><w:r><w:t>paper</w:t></w:r></w:p>` + pageSectPr(12000, 16800, 1440, 1800, 1440, 1800, "", "")
	got := renderHTMLPageLayoutFixture(t, body, HTMLOptions{Standalone: true, IncludeCSS: true})
	pages := htmlPageLayoutSections(t, got, 1)
	bodyTag := regexp.MustCompile(`<body\b([^>]*)>`).FindStringSubmatch(got)
	if len(bodyTag) != 2 || !strings.Contains(" "+htmlPageLayoutAttrs(bodyTag[1])["class"]+" ", " goword-document ") {
		t.Errorf("standalone preview body is missing goword-document class: %s", got)
	}
	css := htmlPageLayoutDeclarations(pages[0].attrs["style"])
	assertHTMLPageLayoutPixels(t, "width", css["width"], 800)
	assertHTMLPageLayoutPixels(t, "min-height", css["min-height"], 1120)
	assertHTMLPageLayoutPixels(t, "padding", css["padding"], 96, 120, 96, 120)
	if !strings.Contains(pages[0].content, "paper") {
		t.Error("preview lost the paragraph content")
	}
}

func TestHTMLPageLayoutHonorsZeroMargins(t *testing.T) {
	body := `<w:p><w:r><w:t>edge to edge</w:t></w:r></w:p>` + pageSectPr(12000, 16800, 0, 0, 0, 0, "", "")
	got := renderHTMLPageLayoutFixture(t, body, HTMLOptions{Standalone: true, IncludeCSS: true})
	pages := htmlPageLayoutSections(t, got, 1)
	css := htmlPageLayoutDeclarations(pages[0].attrs["style"])
	assertHTMLPageLayoutPixels(t, "padding", css["padding"], 0, 0, 0, 0)
	assertHTMLPageLayoutPixels(t, "print margin", htmlPageLayoutPrintPages(got)[css["page"]]["margin"], 0, 0, 0, 0)
}

func TestHTMLPageLayoutNormalizesLandscapeDimensions(t *testing.T) {
	for _, size := range []struct {
		name          string
		width, height int
	}{{"portrait_values", 12000, 16800}, {"already_landscape_values", 16800, 12000}} {
		t.Run(size.name, func(t *testing.T) {
			body := `<w:p><w:r><w:t>landscape</w:t></w:r></w:p>` + pageSectPr(size.width, size.height, 1440, 1440, 1440, 1440, "landscape", "")
			got := renderHTMLPageLayoutFixture(t, body, HTMLOptions{Standalone: true, IncludeCSS: true})
			pages := htmlPageLayoutSections(t, got, 1)
			css := htmlPageLayoutDeclarations(pages[0].attrs["style"])
			assertHTMLPageLayoutPixels(t, "width", css["width"], 1120)
			assertHTMLPageLayoutPixels(t, "min-height", css["min-height"], 800)
			assertHTMLPageLayoutPixels(t, "print size", htmlPageLayoutPrintPages(got)[css["page"]]["size"], 1120, 800)
		})
	}
}

func TestHTMLPageLayoutPrintRulesUseSectionGeometry(t *testing.T) {
	body := `<w:p><w:r><w:t>print</w:t></w:r></w:p>` + pageSectPr(12000, 16800, 1440, 1800, 1440, 1800, "", "")
	got := renderHTMLPageLayoutFixture(t, body, HTMLOptions{Standalone: true, IncludeCSS: true})
	pages := htmlPageLayoutSections(t, got, 1)
	pageName := htmlPageLayoutDeclarations(pages[0].attrs["style"])["page"]
	if pageName == "" {
		t.Fatal("preview section must select its named print page")
	}
	rule := htmlPageLayoutPrintPages(got)[pageName]
	assertHTMLPageLayoutPixels(t, "print size", rule["size"], 800, 1120)
	assertHTMLPageLayoutPixels(t, "print margin", rule["margin"], 96, 120, 96, 120)
	printStart := regexp.MustCompile(`@media\s+print\s*\{`).FindStringIndex(got)
	if printStart == nil {
		t.Fatal("missing print media rules")
	}
	depth, printEnd := 1, printStart[1]
	for printEnd < len(got) && depth > 0 {
		switch got[printEnd] {
		case '{':
			depth++
		case '}':
			depth--
		}
		printEnd++
	}
	pageRule := regexp.MustCompile(`\.goword-page\s*\{([^{}]*)\}`).FindStringSubmatch(got[printStart[1]:printEnd])
	if len(pageRule) != 2 {
		t.Fatal("print media must override the page preview dimensions and padding")
	}
	printCSS := htmlPageLayoutDeclarations(pageRule[1])
	for _, prop := range []string{"padding", "min-height"} {
		if !strings.Contains(printCSS[prop], "!important") {
			t.Errorf("print %s must override the inline preview style, got %q", prop, printCSS[prop])
		}
		assertHTMLPageLayoutPixels(t, "print "+prop, printCSS[prop], 0)
	}
	if compact := strings.ReplaceAll(printCSS["width"], " ", ""); compact != "auto!important" {
		t.Errorf("printed content must fit the @page content area, got width %q", printCSS["width"])
	}
}

func TestHTMLPageLayoutMultipleSectionsKeepWidthsAndBreakMetadata(t *testing.T) {
	first := `<w:p><w:r><w:t>first page</w:t></w:r></w:p><w:p><w:pPr>` + pageSectPr(12000, 16800, 1440, 1440, 1440, 1440, "", "nextPage") + `</w:pPr></w:p>`
	second := `<w:p><w:r><w:t>second page</w:t></w:r></w:p>` + pageSectPr(10000, 14000, 720, 720, 720, 720, "", "continuous")
	got := renderHTMLPageLayoutFixture(t, first+second, HTMLOptions{Standalone: true, IncludeCSS: true})
	pages := htmlPageLayoutSections(t, got, 2)
	if pages[0].attrs["data-break-type"] != "nextPage" || pages[1].attrs["data-break-type"] != "continuous" {
		t.Errorf("section break metadata lost: first=%v, second=%v", pages[0].attrs, pages[1].attrs)
	}
	for i, want := range []struct {
		width, height, margin float64
		text                  string
	}{{800, 1120, 96, "first page"}, {10000.0 / 15, 14000.0 / 15, 48, "second page"}} {
		css := htmlPageLayoutDeclarations(pages[i].attrs["style"])
		assertHTMLPageLayoutPixels(t, "width", css["width"], want.width)
		assertHTMLPageLayoutPixels(t, "min-height", css["min-height"], want.height)
		assertHTMLPageLayoutPixels(t, "padding", css["padding"], want.margin, want.margin, want.margin, want.margin)
		assertHTMLPageLayoutPixels(t, "print size", htmlPageLayoutPrintPages(got)[css["page"]]["size"], want.width, want.height)
		if !strings.Contains(pages[i].content, want.text) {
			t.Errorf("section %d lost or moved its text %q: %s", i, want.text, pages[i].content)
		}
	}
}

func TestHTMLPageLayoutContinuousSameGeometryKeepsBreakMetadata(t *testing.T) {
	first := `<w:p><w:r><w:t>continuous first</w:t></w:r></w:p><w:p><w:pPr>` + pageSectPr(12000, 16800, 1440, 1440, 1440, 1440, "", "") + `</w:pPr></w:p>`
	second := `<w:p><w:r><w:t>continuous second</w:t></w:r></w:p>` + pageSectPr(12000, 16800, 1440, 1440, 1440, 1440, "", "continuous")
	got := renderHTMLPageLayoutFixture(t, first+second, HTMLOptions{Standalone: true, IncludeCSS: true})
	pages := htmlPageLayoutSections(t, got, 2)
	if pages[1].attrs["data-break-type"] != "continuous" {
		t.Errorf("continuous section break metadata lost: %v", pages[1].attrs)
	}
	for i, text := range []string{"continuous first", "continuous second"} {
		if !strings.Contains(pages[i].content, text) {
			t.Errorf("section %d lost %q: %s", i, text, pages[i].content)
		}
	}
	styles := htmlPageLayoutPrintPages(got)
	if len(styles) != 1 {
		t.Errorf("same geometry should reuse one named @page rule, got %d: %v", len(styles), styles)
	}
}

func TestHTMLPageLayoutOnlyAppliesToStandaloneWithCSS(t *testing.T) {
	body := `<w:p><w:r><w:t>fragment</w:t></w:r></w:p>` + pageSectPr(12000, 16800, 1440, 1440, 1440, 1440, "", "")
	for _, tc := range []struct {
		name string
		opts HTMLOptions
	}{{"fragment", HTMLOptions{}}, {"fragment_with_css_option", HTMLOptions{IncludeCSS: true}}, {"standalone_without_css", HTMLOptions{Standalone: true}}} {
		t.Run(tc.name, func(t *testing.T) {
			got := renderHTMLPageLayoutFixture(t, body, tc.opts)
			for _, forbidden := range []string{"goword-page", "goword-document", "<section", "@page", "min-height:"} {
				if strings.Contains(got, forbidden) {
					t.Errorf("%s unexpectedly gained paper layout %q: %s", tc.name, forbidden, got)
				}
			}
			if !strings.Contains(got, "fragment") || strings.Contains(got, "<html>") != tc.opts.Standalone {
				t.Errorf("content or existing standalone wrapper changed: %s", got)
			}
		})
	}
}

func TestHTMLPageLayoutRenderAndWriteEntrypointsMatch(t *testing.T) {
	body := `<w:p><w:r><w:t>entry parity</w:t></w:r></w:p>` + pageSectPr(12000, 16800, 1440, 1800, 1440, 1800, "", "")
	raw := htmlDOCXFormattingFixture(t, body, "", "")
	opts := HTMLOptions{Standalone: true, IncludeCSS: true}
	readerHTML, err := RenderHTML(bytes.NewReader(raw), opts)
	if err != nil {
		t.Fatal(err)
	}
	doc, err := LoadBytes(raw)
	if err != nil {
		t.Fatal(err)
	}
	var written bytes.Buffer
	if err := doc.WriteHTML(&written, opts); err != nil {
		t.Fatal(err)
	}
	documentHTML, err := doc.RenderHTML(opts)
	if err != nil {
		t.Fatal(err)
	}
	withOptions, err := RenderHTMLWithOptions(bytes.NewReader(raw), ReadOptions{}, opts)
	if err != nil {
		t.Fatal(err)
	}
	for name, got := range map[string][]byte{"WriteHTML": written.Bytes(), "Document.RenderHTML": documentHTML, "RenderHTMLWithOptions": withOptions} {
		if !bytes.Equal(readerHTML, got) {
			t.Errorf("%s differs from RenderHTML:\nreader=%s\nactual=%s", name, readerHTML, got)
		}
	}
	htmlPageLayoutSections(t, string(readerHTML), 1)
}

type htmlPageLayoutSection struct {
	attrs   map[string]string
	content string
}

func htmlPageLayoutSections(t *testing.T, got string, count int) []htmlPageLayoutSection {
	t.Helper()
	matches := regexp.MustCompile(`(?s)<section\b([^>]*)>(.*?)</section>`).FindAllStringSubmatch(got, -1)
	if len(matches) != count {
		t.Fatalf("section count=%d, want %d: %s", len(matches), count, got)
	}
	sections := make([]htmlPageLayoutSection, len(matches))
	for i, match := range matches {
		sections[i] = htmlPageLayoutSection{htmlPageLayoutAttrs(match[1]), match[2]}
		if !strings.Contains(" "+sections[i].attrs["class"]+" ", " goword-page ") {
			t.Errorf("section %d lacks goword-page class: %s", i, match[0])
		}
	}
	return sections
}

func htmlPageLayoutAttrs(raw string) map[string]string {
	attrs := make(map[string]string)
	for _, match := range regexp.MustCompile(`([\w:-]+)="([^"]*)"`).FindAllStringSubmatch(raw, -1) {
		attrs[match[1]] = html.UnescapeString(match[2])
	}
	return attrs
}

func htmlPageLayoutDeclarations(raw string) map[string]string {
	css := make(map[string]string)
	for _, decl := range strings.Split(raw, ";") {
		if key, value, ok := strings.Cut(decl, ":"); ok {
			css[strings.TrimSpace(key)] = strings.TrimSpace(value)
		}
	}
	return css
}

func htmlPageLayoutPrintPages(got string) map[string]map[string]string {
	pages := make(map[string]map[string]string)
	for _, match := range regexp.MustCompile(`@page\s+([\w-]+)\s*\{([^{}]*)\}`).FindAllStringSubmatch(got, -1) {
		pages[match[1]] = htmlPageLayoutDeclarations(match[2])
	}
	return pages
}

func assertHTMLPageLayoutPixels(t *testing.T, name, got string, want ...float64) {
	t.Helper()
	fields := strings.Fields(strings.ReplaceAll(got, "!important", ""))
	if len(want) == 4 {
		switch len(fields) {
		case 1:
			fields = []string{fields[0], fields[0], fields[0], fields[0]}
		case 2:
			fields = []string{fields[0], fields[1], fields[0], fields[1]}
		case 3:
			fields = []string{fields[0], fields[1], fields[2], fields[1]}
		}
	}
	if len(fields) != len(want) {
		t.Errorf("%s=%q, want pixel values %v", name, got, want)
		return
	}
	for i, field := range fields {
		value, err := strconv.ParseFloat(strings.TrimSuffix(field, "px"), 64)
		if err != nil || math.Abs(value-want[i]) > 0.000001 || (value != 0 && !strings.HasSuffix(field, "px")) {
			t.Errorf("%s=%q, want pixel values %v", name, got, want)
			return
		}
	}
}
