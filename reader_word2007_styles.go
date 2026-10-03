package word

import (
	"encoding/xml"
	"strconv"
	"strings"

	"github.com/yunkeweb/go-word/style"
)

// Pointers distinguish omitted properties from explicit false/none/zero values.
// Applying each property separately also preserves partially overridden spacing
// and indentation from the parent style.
type wordValue struct {
	Val string `xml:"val,attr"`
}

type wordRunProperties struct {
	Style *wordValue `xml:"rStyle"`
	Fonts *struct {
		ASCII         string `xml:"ascii,attr"`
		HighANSI      string `xml:"hAnsi,attr"`
		EastAsia      string `xml:"eastAsia,attr"`
		ASCIITheme    string `xml:"asciiTheme,attr"`
		HighANSITheme string `xml:"hAnsiTheme,attr"`
		EastAsiaTheme string `xml:"eastAsiaTheme,attr"`
	} `xml:"rFonts"`
	Size      *wordValue   `xml:"sz"`
	Color     *wordColor   `xml:"color"`
	Bold      *wordValue   `xml:"b"`
	Italic    *wordValue   `xml:"i"`
	Underline *wordValue   `xml:"u"`
	Strike    *wordValue   `xml:"strike"`
	DStrike   *wordValue   `xml:"dstrike"`
	SmallCaps *wordValue   `xml:"smallCaps"`
	Caps      *wordValue   `xml:"caps"`
	Hidden    *wordValue   `xml:"vanish"`
	RTL       *wordValue   `xml:"rtl"`
	VertAlign *wordValue   `xml:"vertAlign"`
	Shading   *wordShading `xml:"shd"`
}

type wordColor struct {
	Val        string `xml:"val,attr"`
	ThemeColor string `xml:"themeColor,attr"`
	Tint       string `xml:"themeTint,attr"`
	Shade      string `xml:"themeShade,attr"`
}

type wordShading struct {
	Val   string `xml:"val,attr"`
	Color string `xml:"color,attr"`
	Fill  string `xml:"fill,attr"`
}

type wordParagraphProperties struct {
	Style   *wordValue `xml:"pStyle"`
	Align   *wordValue `xml:"jc"`
	Outline *wordValue `xml:"outlineLvl"`
	Spacing *struct {
		Before *int    `xml:"before,attr"`
		After  *int    `xml:"after,attr"`
		Line   *int    `xml:"line,attr"`
		Rule   *string `xml:"lineRule,attr"`
	} `xml:"spacing"`
	Indent *struct {
		Left      *int `xml:"left,attr"`
		Right     *int `xml:"right,attr"`
		FirstLine *int `xml:"firstLine,attr"`
		Hanging   *int `xml:"hanging,attr"`
	} `xml:"ind"`
	Numbering *struct {
		ID    *wordValue `xml:"numId"`
		Level *wordValue `xml:"ilvl"`
	} `xml:"numPr"`
	PageBreak *wordValue   `xml:"pageBreakBefore"`
	Bidi      *wordValue   `xml:"bidi"`
	Shading   *wordShading `xml:"shd"`
}

type wordNamedStyle struct {
	ID      string                  `xml:"styleId,attr"`
	Kind    string                  `xml:"type,attr"`
	Default string                  `xml:"default,attr"`
	Name    wordValue               `xml:"name"`
	BasedOn wordValue               `xml:"basedOn"`
	Run     wordRunProperties       `xml:"rPr"`
	Para    wordParagraphProperties `xml:"pPr"`
}

type wordStyleSheet struct {
	styles           map[string]*wordNamedStyle
	chains           map[string][]*wordNamedStyle
	defaultPara      string
	defaultFont      wordRunProperties
	defaultParagraph wordParagraphProperties
	theme            *wordTheme
}

type wordTheme struct {
	colors map[string]string
	fonts  map[string]string
}

type wordThemeColor struct {
	SRGB *struct {
		Val string `xml:"val,attr"`
	} `xml:"srgbClr"`
	Sys *struct {
		Last string `xml:"lastClr,attr"`
		Val  string `xml:"val,attr"`
	} `xml:"sysClr"`
}

type wordThemeFontGroup struct {
	Latin struct {
		Typeface string `xml:"typeface,attr"`
	} `xml:"latin"`
	EastAsia struct {
		Typeface string `xml:"typeface,attr"`
	} `xml:"ea"`
	CS struct {
		Typeface string `xml:"typeface,attr"`
	} `xml:"cs"`
}

func parseWordTheme(data []byte) (*wordTheme, error) {
	var part struct {
		Elements struct {
			Colors struct {
				DK1      wordThemeColor `xml:"dk1"`
				LT1      wordThemeColor `xml:"lt1"`
				DK2      wordThemeColor `xml:"dk2"`
				LT2      wordThemeColor `xml:"lt2"`
				Accent1  wordThemeColor `xml:"accent1"`
				Accent2  wordThemeColor `xml:"accent2"`
				Accent3  wordThemeColor `xml:"accent3"`
				Accent4  wordThemeColor `xml:"accent4"`
				Accent5  wordThemeColor `xml:"accent5"`
				Accent6  wordThemeColor `xml:"accent6"`
				Hlink    wordThemeColor `xml:"hlink"`
				FolHlink wordThemeColor `xml:"folHlink"`
			} `xml:"clrScheme"`
			Fonts struct {
				Major wordThemeFontGroup `xml:"majorFont"`
				Minor wordThemeFontGroup `xml:"minorFont"`
			} `xml:"fontScheme"`
		} `xml:"themeElements"`
	}
	if err := xml.Unmarshal(data, &part); err != nil {
		return nil, err
	}
	t := &wordTheme{colors: make(map[string]string), fonts: make(map[string]string)}
	colorValues := map[string]wordThemeColor{
		"dk1": part.Elements.Colors.DK1, "lt1": part.Elements.Colors.LT1,
		"dk2": part.Elements.Colors.DK2, "lt2": part.Elements.Colors.LT2,
		"accent1": part.Elements.Colors.Accent1, "accent2": part.Elements.Colors.Accent2,
		"accent3": part.Elements.Colors.Accent3, "accent4": part.Elements.Colors.Accent4,
		"accent5": part.Elements.Colors.Accent5, "accent6": part.Elements.Colors.Accent6,
		"hlink": part.Elements.Colors.Hlink, "folhlink": part.Elements.Colors.FolHlink,
	}
	for name, c := range colorValues {
		if c.SRGB != nil {
			t.colors[name] = c.SRGB.Val
		} else if c.Sys != nil {
			t.colors[name] = c.Sys.Last
		}
	}
	for prefix, group := range map[string]wordThemeFontGroup{"major": part.Elements.Fonts.Major, "minor": part.Elements.Fonts.Minor} {
		t.fonts[prefix+"HAnsi"] = group.Latin.Typeface
		t.fonts[prefix+"Ascii"] = group.Latin.Typeface
		t.fonts[prefix+"EastAsia"] = group.EastAsia.Typeface
		t.fonts[prefix+"CS"] = group.CS.Typeface
	}
	return t, nil
}

func (t *wordTheme) font(name string) string {
	if t == nil {
		return ""
	}
	return t.fonts[name]
}

func (t *wordTheme) color(name string) string {
	if t == nil {
		return ""
	}
	return t.colors[strings.ToLower(strings.TrimSpace(name))]
}

func parseWordStyles(data []byte) (*wordStyleSheet, error) {
	var part struct {
		Defaults struct {
			Run  wordRunProperties       `xml:"rPrDefault>rPr"`
			Para wordParagraphProperties `xml:"pPrDefault>pPr"`
		} `xml:"docDefaults"`
		Styles []wordNamedStyle `xml:"style"`
	}
	if err := xml.Unmarshal(data, &part); err != nil {
		return nil, err
	}
	s := &wordStyleSheet{
		styles: make(map[string]*wordNamedStyle), chains: make(map[string][]*wordNamedStyle),
		defaultFont: part.Defaults.Run, defaultParagraph: part.Defaults.Para,
	}
	for i := range part.Styles {
		st := &part.Styles[i]
		s.styles[st.ID] = st
		if st.Kind == "paragraph" && wordOn(st.Default) && st.Default != "" {
			s.defaultPara = st.ID
		}
	}
	return s, nil
}

// Resolve iteratively so broken or cyclic basedOn references cannot recurse
// forever. Each style is applied once, ancestors first, then the selected style.
func (s *wordStyleSheet) chain(id string) []*wordNamedStyle {
	if s == nil || id == "" {
		return nil
	}
	if chain, ok := s.chains[id]; ok {
		return chain
	}
	var chain []*wordNamedStyle
	seen := make(map[string]bool)
	for next := id; next != "" && !seen[next]; {
		st := s.styles[next]
		if st == nil {
			break
		}
		seen[next] = true
		chain = append(chain, st)
		next = st.BasedOn.Val
	}
	for i, j := 0, len(chain)-1; i < j; i, j = i+1, j-1 {
		chain[i], chain[j] = chain[j], chain[i]
	}
	s.chains[id] = chain
	return chain
}

type wordParagraphFormat struct {
	font         style.Font
	para         style.Paragraph
	numID, level int
}

func (s *wordStyleSheet) paragraph(direct wordParagraphProperties) wordParagraphFormat {
	var out wordParagraphFormat
	// Word defaults to no inter-paragraph spacing, unlike browser p/h margins.
	out.para.Spacing.BeforeSet = true
	out.para.Spacing.AfterSet = true
	id := ""
	if s != nil {
		id = s.defaultPara
		s.defaultFont.apply(&out.font, false, s.theme)
		s.defaultParagraph.apply(&out)
	}
	if direct.Style != nil {
		id = direct.Style.Val
	}
	out.para.OutlineLevel = wordHeadingLevel(id)
	var chain []*wordNamedStyle
	if s != nil {
		chain = s.chain(id)
	}
	for _, st := range chain {
		if st.Kind == "paragraph" {
			st.Run.apply(&out.font, true, s.theme)
			if depth := wordHeadingLevel(st.Name.Val); depth > 0 {
				out.para.OutlineLevel = depth
			}
			st.Para.apply(&out)
		}
	}
	direct.apply(&out)
	return out
}

func wordHeadingLevel(name string) int {
	name = strings.ToLower(strings.TrimSpace(name))
	if !strings.HasPrefix(name, "heading") {
		return 0
	}
	n, _ := strconv.Atoi(strings.TrimSpace(strings.TrimPrefix(name, "heading")))
	if n >= 1 && n <= 9 {
		return n
	}
	return 0
}

func (s *wordStyleSheet) font(base style.Font, direct wordRunProperties) style.Font {
	var theme *wordTheme
	if s != nil {
		theme = s.theme
	}
	if s != nil && direct.Style != nil {
		for _, st := range s.chain(direct.Style.Val) {
			if st.Kind == "character" {
				st.Run.apply(&base, true, theme)
			}
		}
	}
	direct.apply(&base, false, theme)
	return base
}

func wordOn(val string) bool {
	return val != "0" && val != "false" && val != "off"
}

func wordToggle(dst *bool, value *wordValue, inStyle bool) {
	if value == nil {
		return
	}
	on := wordOn(value.Val)
	if inStyle {
		if on {
			*dst = !*dst
		}
	} else {
		*dst = on
	}
}

func (r wordRunProperties) apply(f *style.Font, inStyle bool, theme *wordTheme) {
	if r.Fonts != nil {
		switch {
		case r.Fonts.EastAsia != "":
			f.Name = r.Fonts.EastAsia
		case r.Fonts.EastAsiaTheme != "":
			f.Name = theme.font(themeFontKey(r.Fonts.EastAsiaTheme, "EastAsia"))
		case r.Fonts.ASCII != "":
			f.Name = r.Fonts.ASCII
		case r.Fonts.ASCIITheme != "":
			f.Name = theme.font(themeFontKey(r.Fonts.ASCIITheme, "Ascii"))
		case r.Fonts.HighANSI != "":
			f.Name = r.Fonts.HighANSI
		case r.Fonts.HighANSITheme != "":
			f.Name = theme.font(themeFontKey(r.Fonts.HighANSITheme, "HAnsi"))
		}
	}
	if r.Size != nil {
		n, _ := strconv.ParseFloat(r.Size.Val, 64)
		f.Size = n / 2
	}
	if r.Color != nil {
		if r.Color.Val != "" {
			f.Color = r.Color.Val
		} else {
			f.Color = theme.color(r.Color.ThemeColor)
		}
	}
	wordToggle(&f.Bold, r.Bold, inStyle)
	wordToggle(&f.Italic, r.Italic, inStyle)
	wordToggle(&f.Strikethrough, r.Strike, inStyle)
	wordToggle(&f.DoubleStrikethrough, r.DStrike, inStyle)
	wordToggle(&f.SmallCaps, r.SmallCaps, inStyle)
	wordToggle(&f.AllCaps, r.Caps, inStyle)
	wordToggle(&f.Hidden, r.Hidden, inStyle)
	wordToggle(&f.RTL, r.RTL, false)
	if r.Underline != nil {
		f.Underline = r.Underline.Val
	}
	if r.VertAlign != nil {
		f.SuperScript = r.VertAlign.Val == "superscript"
		f.SubScript = r.VertAlign.Val == "subscript"
	}
	if r.Shading != nil {
		f.Shading = style.Shading(*r.Shading)
		f.BgColor = r.Shading.Fill
	}
}

func themeFontKey(ref, script string) string {
	ref = strings.ToLower(strings.TrimSpace(ref))
	if strings.HasPrefix(ref, "major") {
		return "major" + script
	}
	return "minor" + script
}

func wordInt(dst *int, src *int) {
	if src != nil {
		*dst = *src
	}
}

func (p wordParagraphProperties) apply(out *wordParagraphFormat) {
	para := &out.para
	if p.Align != nil {
		para.Alignment = p.Align.Val
	}
	if p.Outline != nil {
		para.OutlineLevel = atoi(p.Outline.Val) + 1
		if para.OutlineLevel > 9 {
			para.OutlineLevel = 0
		}
	}
	if s := p.Spacing; s != nil {
		wordInt(&para.Spacing.Before, s.Before)
		wordInt(&para.Spacing.After, s.After)
		para.Spacing.BeforeSet = para.Spacing.BeforeSet || s.Before != nil
		para.Spacing.AfterSet = para.Spacing.AfterSet || s.After != nil
		wordInt(&para.Spacing.Line, s.Line)
		if s.Rule != nil {
			para.Spacing.Rule = *s.Rule
		}
	}
	if i := p.Indent; i != nil {
		wordInt(&para.Indentation.Left, i.Left)
		wordInt(&para.Indentation.Right, i.Right)
		wordInt(&para.Indentation.FirstLine, i.FirstLine)
		wordInt(&para.Indentation.Hanging, i.Hanging)
		if i.FirstLine != nil && i.Hanging == nil {
			para.Indentation.Hanging = 0
		}
	}
	if p.Numbering != nil {
		if p.Numbering.ID != nil {
			out.numID = atoi(p.Numbering.ID.Val)
		}
		if p.Numbering.Level != nil {
			out.level = atoi(p.Numbering.Level.Val)
		}
	}
	if p.PageBreak != nil {
		para.PageBreakBefore = wordOn(p.PageBreak.Val)
	}
	if p.Bidi != nil {
		para.Bidi = wordOn(p.Bidi.Val)
	}
	if p.Shading != nil {
		para.Shading = style.Shading(*p.Shading)
	}
}
