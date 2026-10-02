package word

import (
	"archive/zip"
	"bytes"
	"html"
	"regexp"
	"strings"
	"testing"
)

// These fixtures model OOXML formatting independently of the DOCX writer. That
// is important for imports from Word, whose style IDs need not match our names.
func htmlDOCXFormattingFixture(t *testing.T, body, styles, numbering string) []byte {
	t.Helper()
	const ns = ` xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main"`
	parts := []struct{ name, content string }{
		{"[Content_Types].xml", `<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types"><Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/><Default Extension="xml" ContentType="application/xml"/><Override PartName="/word/document.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"/><Override PartName="/word/styles.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.styles+xml"/><Override PartName="/word/numbering.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.numbering+xml"/></Types>`},
		{"_rels/.rels", `<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="word/document.xml"/></Relationships>`},
		{"word/_rels/document.xml.rels", `<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="rStyles" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/styles" Target="styles.xml"/><Relationship Id="rNumbering" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/numbering" Target="numbering.xml"/></Relationships>`},
		{"word/document.xml", `<w:document` + ns + `><w:body>` + body + `</w:body></w:document>`},
		{"word/styles.xml", `<w:styles` + ns + `>` + styles + `</w:styles>`},
		{"word/numbering.xml", `<w:numbering` + ns + `>` + numbering + `</w:numbering>`},
	}
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	for _, part := range parts {
		w, err := zw.Create(part.name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := w.Write([]byte(part.content)); err != nil {
			t.Fatal(err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func renderHTMLDOCXFormatting(t *testing.T, body, styles, numbering string) string {
	t.Helper()
	raw := htmlDOCXFormattingFixture(t, body, styles, numbering)
	got, err := RenderHTML(bytes.NewReader(raw), HTMLOptions{})
	if err != nil {
		t.Fatal(err)
	}
	assertHTMLStructure(t, string(got))
	return html.UnescapeString(string(got))
}

func requireHTMLDOCXFormatting(t *testing.T, got string, wants ...string) {
	t.Helper()
	for _, want := range wants {
		if !strings.Contains(got, want) {
			t.Errorf("HTML missing %q: %s", want, got)
		}
	}
}

func TestHTMLDOCXColorsUseCSSValues(t *testing.T) {
	for _, tc := range []struct{ name, color, want string }{
		{"hexadecimal", "333333", "color:#333333"},
		{"hexadecimal_with_letters", "00AAFF", "color:#00aaff"},
		{"automatic", "auto", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := renderHTMLDOCXFormatting(t, `<w:p><w:r><w:rPr><w:color w:val="`+tc.color+`"/></w:rPr><w:t>sample</w:t></w:r></w:p>`, "", "")
			got = strings.ToLower(got)
			if tc.want != "" {
				requireHTMLDOCXFormatting(t, got, tc.want)
			}
			if strings.Contains(got, "color:auto") {
				t.Errorf("OOXML automatic color is not a CSS color: %s", got)
			}
		})
	}
}

func TestHTMLDOCXDirectParagraphFormatting(t *testing.T) {
	const paragraph = `<w:p><w:pPr><w:shd w:val="clear" w:color="auto" w:fill="FFFFFF"/><w:spacing w:before="75" w:after="75"/><w:ind w:left="-360"/><w:jc w:val="right"/></w:pPr><w:r><w:t>aligned</w:t></w:r></w:p>`
	for _, tc := range []struct{ name, body string }{
		{"body", paragraph},
		{"table_cell", `<w:tbl><w:tr><w:tc>` + paragraph + `</w:tc></w:tr></w:tbl>`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := renderHTMLDOCXFormatting(t, tc.body, "", "")
			// 75 twips = 5 CSS px, and -360 twips = -24 CSS px.
			requireHTMLDOCXFormatting(t, strings.ToLower(got), "text-align:right", "margin-top:5px", "margin-bottom:5px", "margin-left:-24px", "background-color:#ffffff")
		})
	}
}

func TestHTMLDOCXDirectRunFormatting(t *testing.T) {
	got := renderHTMLDOCXFormatting(t, `<w:p><w:r><w:rPr><w:rFonts w:ascii="SimHei" w:hAnsi="SimHei" w:eastAsia="SimHei"/><w:sz w:val="24"/><w:u w:val="single"/><w:shd w:val="clear" w:color="auto" w:fill="FFFFFF"/></w:rPr><w:t>formatted</w:t></w:r></w:p>`, "", "")
	// w:sz is measured in half-points, unlike the point value in CSS.
	requireHTMLDOCXFormatting(t, got, `font-family:"SimHei"`, "font-size:12pt", "text-decoration:underline")
	requireHTMLDOCXFormatting(t, strings.ToLower(got), "background-color:#ffffff")
}

func TestHTMLDOCXDocumentDefaults(t *testing.T) {
	const defaults = `<w:docDefaults><w:rPrDefault><w:rPr><w:rFonts w:ascii="Example Serif" w:hAnsi="Example Serif"/><w:sz w:val="22"/><w:color w:val="333333"/></w:rPr></w:rPrDefault><w:pPrDefault><w:pPr><w:spacing w:after="150"/></w:pPr></w:pPrDefault></w:docDefaults>`
	got := renderHTMLDOCXFormatting(t, `<w:p><w:r><w:t>defaulted</w:t></w:r></w:p>`, defaults, "")
	requireHTMLDOCXFormatting(t, got, `font-family:"Example Serif"`, "font-size:11pt", "color:#333333", "margin-bottom:10px")
}

func TestHTMLDOCXDefaultParagraphStyleHasArbitraryID(t *testing.T) {
	const styles = `<w:style w:type="paragraph" w:default="1" w:styleId="a"><w:name w:val="Normal"/><w:pPr><w:jc w:val="both"/></w:pPr><w:rPr><w:sz w:val="21"/></w:rPr></w:style>`
	got := renderHTMLDOCXFormatting(t, `<w:p><w:r><w:t>normal</w:t></w:r></w:p>`, styles, "")
	requireHTMLDOCXFormatting(t, got, "text-align:justify", "font-size:10.5pt")
}

func TestHTMLDOCXNamedStylesAndDirectOverrides(t *testing.T) {
	const styles = `<w:style w:type="paragraph" w:default="1" w:styleId="base"><w:name w:val="Normal"/><w:pPr><w:spacing w:after="120"/><w:jc w:val="both"/></w:pPr><w:rPr><w:rFonts w:ascii="Example Serif" w:hAnsi="Example Serif"/><w:sz w:val="21"/></w:rPr></w:style>
	<w:style w:type="paragraph" w:styleId="web"><w:name w:val="Normal (Web)"/><w:basedOn w:val="base"/><w:pPr><w:jc w:val="left"/></w:pPr><w:rPr><w:sz w:val="24"/></w:rPr></w:style>
	<w:style w:type="character" w:styleId="strong"><w:name w:val="Strong"/><w:rPr><w:b/><w:u w:val="single"/></w:rPr></w:style>`
	for _, tc := range []struct {
		name, paragraphProperties, runProperties string
		wants                                    []string
		clearsInherited                          bool
	}{
		{"inherits_named_and_character_styles", "", "", []string{"text-align:left", "font-size:12pt", "font-weight:700", "text-decoration:underline", "margin-bottom:8px"}, false},
		{"direct_properties_override_styles", `<w:spacing w:after="0"/><w:jc w:val="right"/>`, `<w:b w:val="0"/><w:sz w:val="28"/><w:u w:val="none"/>`, []string{"text-align:right", "font-size:14pt"}, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			body := `<w:p><w:pPr><w:pStyle w:val="web"/>` + tc.paragraphProperties + `</w:pPr><w:r><w:rPr><w:rStyle w:val="strong"/>` + tc.runProperties + `</w:rPr><w:t>styled</w:t></w:r></w:p>`
			got := renderHTMLDOCXFormatting(t, body, styles, "")
			requireHTMLDOCXFormatting(t, got, `font-family:"Example Serif"`)
			requireHTMLDOCXFormatting(t, got, tc.wants...)
			if tc.clearsInherited {
				for _, forbidden := range []string{"font-weight:700", "text-decoration:underline", "margin-bottom:8px"} {
					if strings.Contains(got, forbidden) {
						t.Errorf("explicit false/none/zero must clear inherited %q: %s", forbidden, got)
					}
				}
			}
		})
	}
}

func TestHTMLDOCXStyleInheritanceMissingAndCyclicParents(t *testing.T) {
	for _, tc := range []struct{ name, styles string }{
		{"missing_parent", `<w:style w:type="paragraph" w:styleId="child"><w:basedOn w:val="missing"/><w:pPr><w:jc w:val="right"/></w:pPr><w:rPr><w:sz w:val="26"/></w:rPr></w:style>`},
		{"cyclic_parents", `<w:style w:type="paragraph" w:styleId="child"><w:basedOn w:val="parent"/><w:pPr><w:jc w:val="right"/></w:pPr></w:style><w:style w:type="paragraph" w:styleId="parent"><w:basedOn w:val="child"/><w:rPr><w:sz w:val="26"/></w:rPr></w:style>`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := renderHTMLDOCXFormatting(t, `<w:p><w:pPr><w:pStyle w:val="child"/></w:pPr><w:r><w:t>bounded inheritance</w:t></w:r></w:p>`, tc.styles, "")
			requireHTMLDOCXFormatting(t, got, "bounded inheritance", "text-align:right", "font-size:13pt")
		})
	}
}

func TestHTMLDOCXParagraphLineSpacing(t *testing.T) {
	for _, tc := range []struct{ name, rule, line, want string }{
		{"multiple", "auto", "360", "line-height:1.5"},
		{"exact", "exact", "240", "line-height:16px"},
		{"minimum", "atLeast", "180", "line-height:max("},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := renderHTMLDOCXFormatting(t, `<w:p><w:pPr><w:spacing w:line="`+tc.line+`" w:lineRule="`+tc.rule+`"/></w:pPr><w:r><w:rPr><w:sz w:val="24"/></w:rPr><w:t>spaced</w:t></w:r></w:p>`, "", "")
			requireHTMLDOCXFormatting(t, got, tc.want)
			if tc.rule == "atLeast" {
				// The requested 180 twip minimum must not clip a 12 point font.
				if !regexp.MustCompile(`line-height:max\([^;)]*12px`).MatchString(got) {
					t.Errorf("atLeast spacing should preserve its 12px minimum without forcing a short fixed line box: %s", got)
				}
			}
		})
	}
}

func TestHTMLDOCXNumericHeadingStyleKeepsFormatting(t *testing.T) {
	const styles = `<w:style w:type="paragraph" w:styleId="1"><w:name w:val="heading 1"/><w:pPr><w:jc w:val="left"/><w:outlineLvl w:val="0"/></w:pPr><w:rPr><w:b/><w:sz w:val="48"/></w:rPr></w:style>`
	body := `<w:p><w:pPr><w:pStyle w:val="1"/><w:jc w:val="center"/></w:pPr><w:r><w:rPr><w:sz w:val="24"/></w:rPr><w:t>heading</w:t></w:r></w:p>`
	got := renderHTMLDOCXFormatting(t, body, styles, "")
	requireHTMLDOCXFormatting(t, got, "<h1", "</h1>", "text-align:center", "font-weight:700", "font-size:12pt")
}

func TestHTMLDOCXChineseCountingList(t *testing.T) {
	const numbering = `<w:abstractNum w:abstractNumId="0"><w:lvl w:ilvl="0"><w:start w:val="5"/><w:numFmt w:val="chineseCounting"/><w:suff w:val="nothing"/><w:lvlText w:val="%1、"/></w:lvl></w:abstractNum><w:num w:numId="1"><w:abstractNumId w:val="0"/></w:num>`
	const listProperties = `<w:pPr><w:numPr><w:ilvl w:val="0"/><w:numId w:val="1"/></w:numPr></w:pPr>`
	body := `<w:p>` + listProperties + `<w:r><w:t>fifth</w:t></w:r></w:p><w:p>` + listProperties + `<w:r><w:t>sixth</w:t></w:r></w:p>`
	got := renderHTMLDOCXFormatting(t, body, "", numbering)
	requireHTMLDOCXFormatting(t, got, "<ol", `<li`, "fifth", "sixth")
	// Either browser counters or explicit markers can preserve this numbering.
	cssCounter := strings.Contains(got, "cjk-ideographic") || strings.Contains(got, "simp-chinese-informal") || strings.Contains(got, "trad-chinese-informal")
	explicitMarkers := strings.Contains(got, "五、") && strings.Contains(got, "六、")
	if !cssCounter && !explicitMarkers {
		t.Errorf("chineseCounting must not silently become decimal numbering: %s", got)
	}
	if cssCounter && !strings.Contains(got, `start="5"`) {
		t.Errorf("CSS list counters must preserve the OOXML start value 5: %s", got)
	}
}

func TestHTMLDOCXInlineBreaksPreserveOrder(t *testing.T) {
	// This independent fixture guards text order around an in-run break. It is
	// intentionally synthetic and does not depend on a user's source document.
	got := renderHTMLDOCXFormatting(t, `<w:p><w:r><w:t>before</w:t><w:br/><w:t>middle</w:t><w:cr/><w:t>after</w:t></w:r></w:p>`, "", "")
	withBreaks := regexp.MustCompile(`<br\s*/?>`).ReplaceAllString(got, "\n")
	text := regexp.MustCompile(`<[^>]+>`).ReplaceAllString(withBreaks, "")
	if text != "before\nmiddle\nafter" {
		t.Errorf("in-run line breaks must stay between their neighboring text: got %q; HTML: %s", text, got)
	}
}

func TestHTMLDOCXExplicitZeroSpacing(t *testing.T) {
	got := renderHTMLDOCXFormatting(t, `<w:p><w:pPr><w:spacing w:before="0" w:after="0"/></w:pPr><w:r><w:t>no margins</w:t></w:r></w:p>`, "", "")
	requireHTMLDOCXFormatting(t, got, "margin-top:0px", "margin-bottom:0px")
}

func TestHTMLDOCXHeadingFontOverridesBrowserDefaults(t *testing.T) {
	const styles = `<w:style w:type="paragraph" w:styleId="1"><w:name w:val="heading 1"/><w:rPr><w:b/><w:sz w:val="48"/></w:rPr></w:style>`
	got := renderHTMLDOCXFormatting(t, `<w:p><w:pPr><w:pStyle w:val="1"/><w:spacing w:line="180" w:lineRule="atLeast"/></w:pPr><w:r><w:rPr><w:b w:val="0"/><w:sz w:val="24"/></w:rPr><w:t>normal weight heading</w:t></w:r></w:p>`, styles, "")
	open := regexp.MustCompile(`<h1[^>]*>`).FindString(got)
	requireHTMLDOCXFormatting(t, open, "font-size:12pt", "font-weight:normal")
}

func TestHTMLDOCXHyperlinkRunFormattingAndBreakOrder(t *testing.T) {
	const styles = `<w:style w:type="character" w:styleId="link"><w:rPr><w:sz w:val="24"/><w:color w:val="0066CC"/><w:u w:val="single"/></w:rPr></w:style>`
	got := renderHTMLDOCXFormatting(t, `<w:p><w:hyperlink w:anchor="destination"><w:r><w:rPr><w:rStyle w:val="link"/></w:rPr><w:t>before</w:t><w:br/><w:t>after</w:t></w:r></w:hyperlink></w:p>`, styles, "")
	requireHTMLDOCXFormatting(t, strings.ToLower(got), `href="#destination"`, "font-size:12pt", "color:#0066cc", "text-decoration:underline")
	withBreaks := regexp.MustCompile(`<br\s*/?>`).ReplaceAllString(got, "\n")
	if text := regexp.MustCompile(`<[^>]+>`).ReplaceAllString(withBreaks, ""); text != "before\nafter" {
		t.Fatalf("link text order: %q", text)
	}
}

func TestHTMLDOCXHyperlinkExplicitAppearanceOnAnchor(t *testing.T) {
	// A child span cannot remove the anchor's default underline, and inheriting
	// color inside it inherits the browser's blue/purple link color, not the body.
	got := renderHTMLDOCXFormatting(t, `<w:p><w:hyperlink w:anchor="destination"><w:r><w:rPr><w:color w:val="auto"/><w:u w:val="none"/></w:rPr><w:t>plain-looking link</w:t></w:r></w:hyperlink></w:p>`, "", "")
	anchor := regexp.MustCompile(`<a\s[^>]*>`).FindString(got)
	requireHTMLDOCXFormatting(t, anchor, `href="#destination"`, "color:inherit", "text-decoration:none")
}

func TestHTMLDOCXHyperlinkExplicitColorAndUnderlineOnAnchor(t *testing.T) {
	got := renderHTMLDOCXFormatting(t, `<w:p><w:hyperlink w:anchor="destination"><w:r><w:rPr><w:color w:val="112233"/><w:u w:val="single"/></w:rPr><w:t>styled link</w:t></w:r></w:hyperlink></w:p>`, "", "")
	anchor := regexp.MustCompile(`<a\s[^>]*>`).FindString(got)
	requireHTMLDOCXFormatting(t, anchor, "color:#112233", "text-decoration:underline")
}

func TestHTMLDOCXInlineBreakRoundTrip(t *testing.T) {
	raw := htmlDOCXFormattingFixture(t, `<w:p><w:r><w:t>before</w:t><w:br/><w:t>after</w:t></w:r></w:p>`, "", "")
	doc, err := LoadBytes(raw)
	if err != nil {
		t.Fatal(err)
	}
	saved, err := doc.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	xml := readZipFile(t, saved, "word/document.xml")
	if !strings.Contains(xml, "<w:br") {
		t.Fatalf("inline break lost when saving: %s", xml)
	}
	got, err := RenderHTML(bytes.NewReader(saved), HTMLOptions{Strict: true})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(got), "<br") {
		t.Fatalf("inline break lost after round trip: %s", got)
	}
}

func TestHTMLDOCXStyleTogglesAndPartialOverrides(t *testing.T) {
	const styles = `<w:style w:type="paragraph" w:styleId="base"><w:pPr><w:spacing w:before="90" w:after="120"/><w:ind w:left="360" w:hanging="180"/></w:pPr><w:rPr><w:b/><w:i/></w:rPr></w:style><w:style w:type="paragraph" w:styleId="child"><w:basedOn w:val="base"/><w:pPr><w:spacing w:after="0"/></w:pPr><w:rPr><w:b/><w:i w:val="0"/></w:rPr></w:style>`
	got := renderHTMLDOCXFormatting(t, `<w:p><w:pPr><w:pStyle w:val="child"/><w:ind w:firstLine="240"/></w:pPr><w:r><w:t>inherited</w:t></w:r></w:p>`, styles, "")
	requireHTMLDOCXFormatting(t, got, "margin-top:6px", "margin-bottom:0px", "margin-left:24px", "text-indent:16px", "font-style:italic")
	if strings.Contains(got, "font-weight:700") {
		t.Fatalf("style toggle did not cancel inherited bold: %s", got)
	}
}

func TestHTMLDOCXHeadingFormattingRoundTrip(t *testing.T) {
	const styles = `<w:style w:type="paragraph" w:styleId="title"><w:name w:val="heading 2"/></w:style>`
	raw := htmlDOCXFormattingFixture(t, `<w:p><w:pPr><w:pStyle w:val="title"/><w:jc w:val="center"/><w:spacing w:before="0" w:after="0"/></w:pPr><w:r><w:rPr><w:b/><w:sz w:val="24"/></w:rPr><w:t>rich title</w:t></w:r><w:r><w:rPr><w:i/><w:sz w:val="24"/></w:rPr><w:t> suffix</w:t></w:r></w:p>`, styles, "")
	doc, err := LoadBytes(raw)
	if err != nil {
		t.Fatal(err)
	}
	saved, err := doc.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	got, err := RenderHTML(bytes.NewReader(saved), HTMLOptions{Strict: true})
	if err != nil {
		t.Fatal(err)
	}
	requireHTMLDOCXFormatting(t, string(got), "<h2", "text-align:center", "margin-top:0px", "margin-bottom:0px", "font-weight:700", "font-style:italic", "rich title", " suffix")
}

func TestHTMLDOCXHeadingOutlineCanBeCleared(t *testing.T) {
	const styles = `<w:style w:type="paragraph" w:styleId="Heading1"><w:name w:val="heading 1"/><w:pPr><w:outlineLvl w:val="9"/></w:pPr></w:style>`
	got := renderHTMLDOCXFormatting(t, `<w:p><w:pPr><w:pStyle w:val="Heading1"/></w:pPr><w:r><w:t>body text</w:t></w:r></w:p>`, styles, "")
	if strings.Contains(got, "<h1") {
		t.Fatalf("explicit body outline level must win over the style name: %s", got)
	}
}
