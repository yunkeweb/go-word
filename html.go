package word

import (
	"encoding/xml"
	"io"
	"regexp"
	"strconv"
	"strings"

	"github.com/yunkeweb/go-word/element"
	"github.com/yunkeweb/go-word/style"
)

// htmlContainer is the PHPWord AbstractContainer surface used by Shared\Html::addHtml.
type htmlContainer interface {
	AddText(text string, styles ...any) *element.Text
	AddTextRun(styles ...any) *element.TextRun
	AddTextBreak(count ...int)
	AddLink(target, text string, styles ...any) *element.Link
	AddTable(styles ...any) *element.Table
	AddListItem(text string, depth int, styles ...any) *element.ListItem
	AddImage(source string, styles ...any) *element.Image
	AddTitle(text string, depth int, page ...int) *element.Title
	AddCheckBox(name, text string, styles ...any) *element.CheckBox
	AddLine(s ...style.Line) *element.Line
	AddPageBreak() *element.PageBreak
}

// AddHTML parses HTML and appends equivalent elements to dst
// (PHPWord Shared\Html::addHtml).
func AddHTML(dst htmlContainer, html string, fullHTML ...bool) error {
	full := false
	if len(fullHTML) > 0 {
		full = fullHTML[0]
	}
	prepared := prepareHTML(html, full)
	dec := xml.NewDecoder(strings.NewReader(prepared))
	dec.Strict = false
	dec.AutoClose = xml.HTMLAutoClose
	dec.Entity = xml.HTMLEntity
	root, err := decodeHTML(dec)
	if err != nil {
		return err
	}
	parseHTMLNode(root, dst, htmlStyle{})
	return nil
}

type htmlNode struct {
	tag  string
	attr map[string]string
	text string
	kids []*htmlNode
}

type htmlStyle struct {
	font style.Font
	para style.Paragraph
}

func prepareHTML(s string, full bool) string {
	s = strings.ReplaceAll(s, "\r", "")
	s = strings.ReplaceAll(s, "\n", "")
	void := regexp.MustCompile(`(?i)<(br|img|hr|input)(\s[^>]*)?>`)
	s = void.ReplaceAllStringFunc(s, func(m string) string {
		if strings.HasSuffix(strings.TrimSpace(m), "/>") {
			return m
		}
		return strings.TrimSuffix(m, ">") + "/>"
	})
	if !full && !strings.Contains(strings.ToLower(s), "<body") {
		s = "<body>" + s + "</body>"
	}
	return s
}

func decodeHTML(dec *xml.Decoder) (*htmlNode, error) {
	root := &htmlNode{tag: "root"}
	stack := []*htmlNode{root}
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			n := &htmlNode{tag: strings.ToLower(t.Name.Local), attr: map[string]string{}}
			for _, a := range t.Attr {
				n.attr[strings.ToLower(a.Name.Local)] = a.Value
			}
			parent := stack[len(stack)-1]
			parent.kids = append(parent.kids, n)
			stack = append(stack, n)
		case xml.EndElement:
			if len(stack) > 1 {
				stack = stack[:len(stack)-1]
			}
		case xml.CharData:
			text := string(t)
			if strings.TrimSpace(text) == "" && !strings.Contains(text, " ") {
				continue
			}
			parent := stack[len(stack)-1]
			parent.kids = append(parent.kids, &htmlNode{tag: "#text", text: text})
		}
	}
	return root, nil
}

func parseHTMLNode(n *htmlNode, dst htmlContainer, st htmlStyle) {
	if n == nil {
		return
	}
	st = applyHTMLStyle(n, st)
	switch n.tag {
	case "root", "html", "body", "div":
		for _, k := range n.kids {
			parseHTMLNode(k, dst, st)
		}
	case "p":
		tr := dst.AddTextRun(st.para)
		for _, k := range n.kids {
			parseInline(k, tr, st)
		}
	case "h1", "h2", "h3", "h4", "h5", "h6":
		depth, _ := strconv.Atoi(n.tag[1:])
		dst.AddTitle(collectText(n), depth)
	case "br":
		dst.AddTextBreak(1)
	case "hr":
		dst.AddLine(style.Line{Weight: 1, Color: "000000", Width: 4500})
	case "table":
		tbl := dst.AddTable()
		for _, k := range n.kids {
			parseTableNode(k, tbl, st)
		}
	case "ul":
		parseList(n, dst, st, 0, style.ListTypeBullet)
	case "ol":
		parseList(n, dst, st, 0, style.ListTypeNumber)
	case "img":
		src := n.attr["src"]
		if src != "" {
			img := dst.AddImage(src)
			if w := cssSize(n.attr["width"]); w > 0 {
				img.Style.Width = w
			}
			if h := cssSize(n.attr["height"]); h > 0 {
				img.Style.Height = h
			}
			img.Style.AltText = n.attr["alt"]
		}
	case "input":
		if n.attr["type"] == "checkbox" {
			dst.AddCheckBox(n.attr["name"], n.attr["value"], st.font, st.para)
		}
	case "a", "span", "strong", "b", "em", "i", "u", "s", "strike", "del", "sup", "sub", "font", "#text":
		tr := dst.AddTextRun(st.para)
		parseInline(n, tr, st)
	default:
		for _, k := range n.kids {
			parseHTMLNode(k, dst, st)
		}
	}
}

func parseInline(n *htmlNode, tr *element.TextRun, st htmlStyle) {
	st = applyHTMLStyle(n, st)
	switch n.tag {
	case "#text":
		if n.text != "" {
			tr.AddText(n.text, st.font)
		}
	case "br":
		tr.AddText("\n", st.font)
	case "a":
		href := n.attr["href"]
		tr.AddLink(href, collectText(n), st.font)
	case "img":
		src := n.attr["src"]
		if src != "" {
			tr.AddImage(src)
		}
	default:
		for _, k := range n.kids {
			parseInline(k, tr, st)
		}
	}
}

func parseTableNode(n *htmlNode, tbl *element.Table, st htmlStyle) {
	switch n.tag {
	case "tbody", "thead", "tfoot":
		for _, k := range n.kids {
			parseTableNode(k, tbl, st)
		}
	case "tr":
		row := tbl.AddRow()
		for _, k := range n.kids {
			if k.tag == "td" || k.tag == "th" {
				width := 0
				if w := cssSize(k.attr["width"]); w > 0 {
					width = int(w * 15) // px → twip-ish
				}
				cell := row.AddCell(width)
				for _, c := range k.kids {
					parseHTMLNode(c, cell, st)
				}
			}
		}
	}
}

func parseList(n *htmlNode, dst htmlContainer, st htmlStyle, depth int, listType string) {
	for _, k := range n.kids {
		switch k.tag {
		case "li":
			dst.AddListItem(strings.TrimSpace(collectText(k)), depth, st.font, listType, st.para)
		case "ul":
			parseList(k, dst, st, depth+1, style.ListTypeBullet)
		case "ol":
			parseList(k, dst, st, depth+1, style.ListTypeNumber)
		}
	}
}

func applyHTMLStyle(n *htmlNode, st htmlStyle) htmlStyle {
	switch n.tag {
	case "strong", "b":
		st.font.Bold = true
	case "em", "i":
		st.font.Italic = true
	case "u":
		st.font.Underline = style.UnderlineSingle
	case "s", "strike", "del":
		st.font.Strikethrough = true
	case "sup":
		st.font.SuperScript = true
	case "sub":
		st.font.SubScript = true
	}
	if n.attr != nil {
		if v := n.attr["align"]; v != "" {
			st.para.Alignment = mapAlign(v)
		}
		if v := n.attr["color"]; v != "" {
			st.font.Color = strings.TrimPrefix(v, "#")
		}
		if v := n.attr["bgcolor"]; v != "" {
			st.font.BgColor = strings.TrimPrefix(v, "#")
		}
		parseCSS(n.attr["style"], &st)
	}
	return st
}

func parseCSS(css string, st *htmlStyle) {
	for _, part := range strings.Split(css, ";") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		k, v, ok := strings.Cut(part, ":")
		if !ok {
			continue
		}
		k = strings.ToLower(strings.TrimSpace(k))
		v = strings.TrimSpace(v)
		switch k {
		case "color":
			st.font.Color = strings.TrimPrefix(v, "#")
		case "background-color", "background":
			st.font.BgColor = strings.TrimPrefix(v, "#")
		case "font-weight":
			if v == "bold" || v == "bolder" {
				st.font.Bold = true
			} else if n, err := strconv.Atoi(v); err == nil && n >= 600 {
				st.font.Bold = true
			}
		case "font-style":
			if v == "italic" {
				st.font.Italic = true
			}
		case "text-decoration":
			if strings.Contains(v, "underline") {
				st.font.Underline = style.UnderlineSingle
			}
			if strings.Contains(v, "line-through") {
				st.font.Strikethrough = true
			}
		case "text-align":
			st.para.Alignment = mapAlign(v)
		case "font-size":
			if pt, ok := cssPoints(v); ok {
				st.font.Size = pt
			}
		case "font-family":
			name := strings.Trim(strings.Split(v, ",")[0], `"' `)
			st.font.Name = name
		}
	}
}

func mapAlign(v string) string {
	switch strings.ToLower(v) {
	case "center":
		return style.JcCenter
	case "right":
		return style.JcRight
	case "justify":
		return style.JcBoth
	default:
		return style.JcLeft
	}
}

func collectText(n *htmlNode) string {
	if n.tag == "#text" {
		return n.text
	}
	var b strings.Builder
	for _, k := range n.kids {
		b.WriteString(collectText(k))
	}
	return b.String()
}

func cssSize(s string) float64 {
	s = strings.TrimSuffix(strings.TrimSpace(s), "px")
	s = strings.TrimSuffix(s, "pt")
	f, _ := strconv.ParseFloat(s, 64)
	return f
}

func cssPoints(v string) (float64, bool) {
	v = strings.TrimSpace(v)
	if strings.HasSuffix(v, "px") {
		f, err := strconv.ParseFloat(strings.TrimSuffix(v, "px"), 64)
		return f * 0.75, err == nil
	}
	if strings.HasSuffix(v, "pt") {
		f, err := strconv.ParseFloat(strings.TrimSuffix(v, "pt"), 64)
		return f, err == nil
	}
	f, err := strconv.ParseFloat(v, 64)
	return f, err == nil
}
