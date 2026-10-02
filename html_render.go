package word

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"html"
	"io"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/yunkeweb/go-word/element"
	"github.com/yunkeweb/go-word/pkg/common"
	"github.com/yunkeweb/go-word/style"
)

// HTMLImageMode controls image references in rendered HTML.
type HTMLImageMode string

const (
	HTMLImageDataURI HTMLImageMode = "data"
	HTMLImageURL     HTMLImageMode = "url"
)

// HTMLOptions controls DOCX to HTML rendering.
type HTMLOptions struct {
	Standalone            bool
	Title                 string
	ImageMode             HTMLImageMode
	ImageURL              func(*element.Image) (string, error)
	IncludeHeadersFooters bool
	IncludeCSS            bool
	Strict                bool
}

// HTMLDiagnostic describes a lossy or unsupported conversion.
type HTMLDiagnostic struct {
	ElementType string
	Message     string
}

// HTMLRenderResult contains rendered HTML and non-fatal diagnostics.
type HTMLRenderResult struct {
	HTML        []byte
	Diagnostics []HTMLDiagnostic
}

// HTMLUnsupportedError reports unsupported elements in strict mode.
type HTMLUnsupportedError struct{ Diagnostics []HTMLDiagnostic }

func (e *HTMLUnsupportedError) Error() string {
	if e == nil || len(e.Diagnostics) == 0 {
		return "word: unsupported HTML elements"
	}
	return fmt.Sprintf("word: unsupported HTML elements: %s", e.Diagnostics[0].ElementType)
}

// RenderHTML renders a loaded document as an HTML fragment or standalone page.
func (d *Document) RenderHTML(opts HTMLOptions) ([]byte, error) {
	res, err := d.RenderHTMLWithDiagnostics(opts)
	if err != nil {
		return nil, err
	}
	return res.HTML, nil
}

// RenderHTMLWithDiagnostics renders a document and returns conversion diagnostics.
func (d *Document) RenderHTMLWithDiagnostics(opts HTMLOptions) (HTMLRenderResult, error) {
	if d == nil {
		return HTMLRenderResult{}, fmt.Errorf("word: nil document")
	}
	var buf bytes.Buffer
	r := newHTMLRenderer(&buf, opts)
	r.renderDocument(d)
	result := HTMLRenderResult{HTML: append([]byte(nil), buf.Bytes()...), Diagnostics: append([]HTMLDiagnostic(nil), r.diagnostics...)}
	if r.err != nil {
		return result, r.err
	}
	if r.opts.Strict && len(r.diagnostics) > 0 {
		return result, &HTMLUnsupportedError{Diagnostics: result.Diagnostics}
	}
	return result, nil
}

// RenderHTML reads a DOCX from r and renders it as HTML.
func RenderHTML(r io.Reader, opts HTMLOptions) ([]byte, error) {
	d, err := Read(r)
	if err != nil {
		return nil, err
	}
	return d.RenderHTML(opts)
}

// RenderHTMLFile loads a DOCX file and renders it as HTML.
func RenderHTMLFile(path string, opts HTMLOptions) ([]byte, error) {
	d, err := Open(path)
	if err != nil {
		return nil, err
	}
	return d.RenderHTML(opts)
}

// RenderHTMLWithOptions reads a DOCX with read budgets and renders it as HTML.
func RenderHTMLWithOptions(r io.Reader, readOpts ReadOptions, htmlOpts HTMLOptions) ([]byte, error) {
	d, err := ReadWithOptions(r, readOpts)
	if err != nil {
		return nil, err
	}
	return d.RenderHTML(htmlOpts)
}

// RenderHTMLFileWithOptions loads a DOCX with read budgets and renders it as HTML.
func RenderHTMLFileWithOptions(path string, readOpts ReadOptions, htmlOpts HTMLOptions) ([]byte, error) {
	d, err := LoadWithOptions(path, readOpts)
	if err != nil {
		return nil, err
	}
	return d.RenderHTML(htmlOpts)
}

// WriteHTML renders the document directly to an output stream.
func (d *Document) WriteHTML(w io.Writer, opts HTMLOptions) error {
	if w == nil {
		return fmt.Errorf("word: nil HTML writer")
	}
	if d == nil {
		return fmt.Errorf("word: nil document")
	}
	r := newHTMLRenderer(w, opts)
	r.renderDocument(d)
	if r.err != nil {
		return r.err
	}
	if r.opts.Strict && len(r.diagnostics) > 0 {
		return &HTMLUnsupportedError{Diagnostics: append([]HTMLDiagnostic(nil), r.diagnostics...)}
	}
	return nil
}

func normalizeHTMLOptions(opts HTMLOptions) HTMLOptions {
	if opts.ImageMode == "" {
		opts.ImageMode = HTMLImageDataURI
	}
	return opts
}

type htmlRenderer struct {
	opts        HTMLOptions
	out         io.Writer
	err         error
	diagnostics []HTMLDiagnostic
	document    *Document
	listCounts  map[htmlListKey]int
}

func newHTMLRenderer(out io.Writer, opts HTMLOptions) *htmlRenderer {
	return &htmlRenderer{opts: normalizeHTMLOptions(opts), out: out}
}

type htmlAttr struct{ name, value string }

func (r *htmlRenderer) renderDocument(d *Document) {
	r.document = d
	r.listCounts = make(map[htmlListKey]int)
	if r.opts.Standalone {
		r.writeString("<!doctype html><html><head><meta charset=\"utf-8\">")
		title := r.opts.Title
		if title == "" && d.info != nil {
			title = d.info.Title
		}
		if title != "" {
			r.tagText("title", title)
		}
		if r.opts.IncludeCSS {
			r.writeString("<style>")
			r.writeString(defaultHTMLCSS)
			r.writeString("</style>")
		}
		r.writeString("</head><body>")
	}
	for _, sec := range d.sections {
		if sec == nil {
			continue
		}
		if r.opts.IncludeHeadersFooters {
			for _, h := range sec.Headers {
				r.open("header", htmlAttr{"data-type", h.HeaderType})
				r.renderElements(h.Elements())
				r.close("header")
			}
		}
		r.renderElements(sec.Elements())
		if r.opts.IncludeHeadersFooters {
			for _, f := range sec.Footers {
				r.open("footer", htmlAttr{"data-type", f.HeaderType})
				r.renderElements(f.Elements())
				r.close("footer")
			}
		}
	}
	if r.opts.Standalone {
		r.writeString("</body></html>")
	}
}

func (r *htmlRenderer) renderElements(elements []element.Element) {
	for i := 0; i < len(elements); {
		if isListElement(elements[i]) {
			next := r.renderList(elements, i, listDepth(elements[i]))
			if next > i {
				i = next
				continue
			}
		}
		r.renderBlock(elements[i])
		i++
	}
}

func (r *htmlRenderer) renderBlock(el element.Element) {
	switch v := el.(type) {
	case *element.Text:
		r.openStyled("p", paragraphStyle(v.ParagraphStyle))
		r.renderText(v.Content, v.FontStyle)
		r.close("p")
	case *element.PreserveText:
		r.open("p")
		r.renderText(v.Content, v.FontStyle)
		r.close("p")
	case *element.TextRun:
		r.openStyled("p", paragraphStyle(v.ParagraphStyle))
		r.renderInlineElements(v.Elements())
		r.close("p")
	case *element.ListItem:
		r.open("p")
		r.renderText(v.Text, v.FontStyle)
		r.close("p")
	case *element.ListItemRun:
		r.openStyled("p", paragraphStyle(v.ParagraphStyle))
		r.renderInlineElements(v.Elements())
		r.close("p")
	case *element.Title:
		depth := v.Depth
		if depth < 1 {
			depth = 1
		}
		if depth > 6 {
			depth = 6
		}
		tag := "h" + strconv.Itoa(depth)
		attrs := []htmlAttr{}
		if v.BookmarkName != "" {
			attrs = append(attrs, htmlAttr{"id", v.BookmarkName})
		}
		r.open(tag, attrs...)
		if v.Run != nil {
			r.renderInlineElements(v.Run.Elements())
		} else {
			r.text(v.Text)
		}
		r.close(tag)
	case *element.Bookmark:
		r.open("span", htmlAttr{"id", v.Name})
		r.close("span")
	case *element.Link:
		r.open("p")
		r.renderLink(v)
		r.close("p")
	case *element.Table:
		r.renderTable(v)
	case *element.Image:
		r.open("p")
		r.renderImage(v)
		r.close("p")
	case *element.TextBreak:
		r.open("p")
		r.close("p")
	case *element.PageBreak:
		r.void("div", htmlAttr{"class", "goword-page-break"})
	case *element.Footnote:
		r.open("aside", htmlAttr{"class", "goword-footnote"})
		r.renderElements(v.Elements())
		r.close("aside")
	case *element.Endnote:
		r.open("aside", htmlAttr{"class", "goword-endnote"})
		r.renderElements(v.Elements())
		r.close("aside")
	case *element.Comment:
		r.open("aside", htmlAttr{"class", "goword-comment"})
		r.renderElements(v.Elements())
		r.close("aside")
	case *element.TextBox:
		r.open("div", htmlAttr{"class", "goword-textbox"})
		r.renderElements(v.Elements())
		r.close("div")
	case *element.DMLShape:
		r.open("div", htmlAttr{"class", "goword-shape"})
		r.renderElements(v.Elements())
		r.close("div")
	case *element.Formula:
		if v.Source != "" {
			r.open("span", htmlAttr{"class", "goword-formula"})
			r.text(v.Source)
			r.close("span")
		} else {
			r.unsupported(v, "formula has no source text")
		}
	case *element.CheckBox:
		r.open("p")
		r.void("input", htmlAttr{"type", "checkbox"}, htmlAttr{"disabled", "disabled"})
		r.renderText(v.Content, v.FontStyle)
		r.close("p")
	default:
		r.unsupported(el, "element is not mapped to HTML")
	}
}

func (r *htmlRenderer) renderInlineElements(elements []element.Element) {
	for _, el := range elements {
		switch v := el.(type) {
		case *element.Text:
			r.renderText(v.Content, v.FontStyle)
		case *element.PreserveText:
			r.renderText(v.Content, v.FontStyle)
		case *element.Link:
			r.renderLink(v)
		case *element.Image:
			r.renderImage(v)
		case *element.TextRun:
			r.renderInlineElements(v.Elements())
		case *element.ListItemRun:
			r.renderInlineElements(v.Elements())
		case *element.Bookmark:
			r.open("span", htmlAttr{"id", v.Name})
			r.close("span")
		case *element.CheckBox:
			r.void("input", htmlAttr{"type", "checkbox"}, htmlAttr{"disabled", "disabled"})
			r.renderText(v.Content, v.FontStyle)
		case *element.FormField:
			r.renderText(v.Value, v.FontStyle)
		default:
			r.unsupported(el, "inline element is not mapped to HTML")
		}
	}
}

func (r *htmlRenderer) renderText(text string, font any) {
	if st := fontStyle(font); st != "" {
		r.openStyled("span", st)
		r.text(text)
		r.close("span")
	} else {
		r.text(text)
	}
}

func (r *htmlRenderer) renderLink(v *element.Link) {
	target := v.Target
	if v.Internal && !strings.HasPrefix(target, "#") {
		target = "#" + target
	}
	if !safeURL(target) {
		r.unsupported(v, "unsafe hyperlink URL")
		r.renderText(v.Text, v.FontStyle)
		return
	}
	r.open("a", htmlAttr{"href", target})
	r.renderText(v.Text, v.FontStyle)
	r.close("a")
}

func (r *htmlRenderer) renderTable(t *element.Table) {
	rowspan, skip := tableVerticalMerges(t)
	attrs := []htmlAttr{}
	if st := tableStyle(t.Style); st != "" {
		attrs = append(attrs, htmlAttr{"style", st})
	}
	r.open("table", attrs...)
	r.open("tbody")
	for _, row := range t.Rows {
		r.open("tr")
		for _, cell := range row.Cells {
			if skip[cell] {
				continue
			}
			attrs := []htmlAttr{}
			if cell.Style.GridSpan > 1 {
				attrs = append(attrs, htmlAttr{"colspan", strconv.Itoa(cell.Style.GridSpan)})
			}
			if rowspan[cell] > 1 {
				attrs = append(attrs, htmlAttr{"rowspan", strconv.Itoa(rowspan[cell])})
			}
			if st := cellStyle(cell.Style); st != "" {
				attrs = append(attrs, htmlAttr{"style", st})
			}
			r.open("td", attrs...)
			r.renderElements(cell.Elements())
			r.close("td")
		}
		r.close("tr")
	}
	r.close("tbody")
	r.close("table")
}

func tableVerticalMerges(t *element.Table) (map[*element.Cell]int, map[*element.Cell]bool) {
	rowspan := make(map[*element.Cell]int)
	skip := make(map[*element.Cell]bool)
	active := make(map[int]*htmlMergeState)
	for _, row := range t.Rows {
		seen := make(map[*htmlMergeState]bool)
		for _, state := range active {
			if !seen[state] {
				state.counted = false
				seen[state] = true
			}
		}
		next := make(map[int]*htmlMergeState)
		col := 0
		for _, cell := range row.Cells {
			if cell.Style.VMerge == "continue" {
				if state := active[col]; state != nil {
					if !state.counted {
						rowspan[state.root]++
						state.counted = true
					}
					skip[cell] = true
					for offset := 0; offset < state.span; offset++ {
						next[col+offset] = state
					}
					col += state.span
					continue
				}
			}
			span := cell.Style.GridSpan
			if span < 1 {
				span = 1
			}
			if cell.Style.VMerge == "restart" {
				rowspan[cell] = 1
				state := &htmlMergeState{root: cell, span: span}
				for offset := 0; offset < span; offset++ {
					next[col+offset] = state
				}
			}
			col += span
		}
		active = next
	}
	return rowspan, skip
}

type htmlMergeState struct {
	root    *element.Cell
	span    int
	counted bool
}

func (r *htmlRenderer) renderList(elements []element.Element, start, depth int) int {
	list := r.listStyle(elements[start])
	tag := "ul"
	if list.ListType == style.ListTypeNumber || list.ListType == style.ListTypeNumberNE || list.ListType == style.ListTypeMultilevel {
		tag = "ol"
	}
	var attrs []htmlAttr
	switch list.Format {
	case style.NumberUpperRoman:
		attrs = append(attrs, htmlAttr{"type", "I"})
	case style.NumberLowerRoman:
		attrs = append(attrs, htmlAttr{"type", "i"})
	case style.NumberUpperLetter:
		attrs = append(attrs, htmlAttr{"type", "A"})
	case style.NumberLowerLetter:
		attrs = append(attrs, htmlAttr{"type", "a"})
	case style.NumberDecimalZero:
		attrs = append(attrs, htmlAttr{"style", "list-style-type:decimal-leading-zero"})
	}
	if list.ListType == style.ListTypeNone || list.Format == style.NumberNone {
		attrs = []htmlAttr{{"style", "list-style-type:none"}}
	}
	key := htmlListKey{list.NumId, depth}
	number := maxInt(list.Start, 1)
	if list.NumId > 0 && r.listCounts[key] > 0 {
		number = r.listCounts[key]
	}
	if tag == "ol" && number != 1 {
		attrs = append(attrs, htmlAttr{"start", strconv.Itoa(number)})
	}
	r.open(tag, attrs...)
	i := start
	for i < len(elements) {
		if !isListElement(elements[i]) || listDepth(elements[i]) != depth || r.listStyle(elements[i]) != list {
			break
		}
		r.open("li")
		switch v := elements[i].(type) {
		case *element.ListItem:
			r.renderText(v.Text, v.FontStyle)
		case *element.ListItemRun:
			r.renderInlineElements(v.Elements())
		}
		if list.NumId > 0 {
			r.listCounts[key] = number + 1
			for child := range r.listCounts {
				if child.numID == list.NumId && child.depth > depth {
					delete(r.listCounts, child)
				}
			}
		}
		number++
		i++
		for i < len(elements) && isListElement(elements[i]) && listDepth(elements[i]) > depth {
			i = r.renderList(elements, i, listDepth(elements[i]))
		}
		r.close("li")
	}
	r.close(tag)
	return i
}

func isListElement(el element.Element) bool {
	switch el.(type) {
	case *element.ListItem, *element.ListItemRun:
		return true
	default:
		return false
	}
}
func listDepth(el element.Element) int {
	switch v := el.(type) {
	case *element.ListItem:
		return maxInt(v.Depth, 0)
	case *element.ListItemRun:
		return maxInt(v.Depth, 0)
	default:
		return 0
	}
}

type htmlListKey struct{ numID, depth int }

func (r *htmlRenderer) listStyle(el element.Element) style.ListItem {
	var value any
	switch v := el.(type) {
	case *element.ListItem:
		value = v.ListStyle
	case *element.ListItemRun:
		value = v.ListStyle
	}
	list := style.ListItem{ListType: style.ListTypeBullet}
	switch v := value.(type) {
	case style.ListItem:
		list = v
	case *style.ListItem:
		if v != nil {
			list = *v
		}
	case string:
		list.ListType = v
		if r.document != nil {
			id := 2
			for _, ns := range r.document.styles {
				if ns.Kind != "numbering" || ns.Numbering == nil {
					continue
				}
				id++
				if ns.Name == v && listDepth(el) < len(ns.Numbering.Levels) {
					level := ns.Numbering.Levels[listDepth(el)]
					list = style.ListItem{NumId: id, Format: level.Format, Start: level.Start}
					break
				}
			}
		}
	}
	if list.Format == style.NumberBullet {
		list.ListType = style.ListTypeBullet
	} else if list.Format != "" && list.Format != style.NumberNone {
		list.ListType = style.ListTypeNumber
	}
	list.Depth = 0 // The element owns nesting; compare only numbering identity/style.
	return list
}
func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func (r *htmlRenderer) renderImage(img *element.Image) {
	src, err := r.imageSource(img)
	if err != nil {
		r.unsupported(img, err.Error())
		return
	}
	attrs := []htmlAttr{{"src", src}}
	alt := img.Style.AltText
	if alt == "" {
		alt = img.GetName()
	}
	attrs = append(attrs, htmlAttr{"alt", alt})
	if img.Style.Width > 0 {
		attrs = append(attrs, htmlAttr{"width", formatFloat(img.Style.Width)})
	}
	if img.Style.Height > 0 {
		attrs = append(attrs, htmlAttr{"height", formatFloat(img.Style.Height)})
	}
	r.void("img", attrs...)
}

func (r *htmlRenderer) imageSource(img *element.Image) (string, error) {
	if r.opts.ImageURL != nil {
		src, err := r.opts.ImageURL(img)
		if err != nil {
			return "", err
		}
		if !safeURL(src) {
			return "", fmt.Errorf("image %q has an unsafe URL", img.GetName())
		}
		return src, nil
	}
	if r.opts.ImageMode == HTMLImageURL {
		if img.Source != "" && safeURL(img.Source) {
			return img.Source, nil
		}
		if img.Media.Target != "" && safeURL(img.Media.Target) {
			return img.Media.Target, nil
		}
		return "", fmt.Errorf("image %q has an unsafe URL", img.GetName())
	}
	data := img.Data
	if len(data) == 0 {
		data = img.Media.Data
	}
	if len(data) == 0 && img.Source != "" {
		var err error
		data, err = common.ReadFile(img.Source)
		if err != nil {
			return "", err
		}
	}
	if len(data) == 0 {
		return "", fmt.Errorf("image %q has no data", img.GetName())
	}
	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(img.GetName()), "."))
	if ext == "jpg" {
		ext = "jpeg"
	}
	mime := mimeForExt(ext)
	if mime == "" {
		mime = "application/octet-stream"
	}
	return "data:" + mime + ";base64," + base64.StdEncoding.EncodeToString(data), nil
}

func (r *htmlRenderer) unsupported(el element.Element, msg string) {
	typ := "unknown"
	if el != nil {
		typ = el.Type()
	}
	r.diagnostics = append(r.diagnostics, HTMLDiagnostic{ElementType: typ, Message: msg})
	r.open("span", htmlAttr{"class", "goword-unsupported"}, htmlAttr{"data-element", typ})
	r.text("[Unsupported: " + typ + "]")
	r.close("span")
}
func (r *htmlRenderer) open(name string, attrs ...htmlAttr) {
	r.writeString("<")
	r.writeString(name)
	r.writeAttrs(attrs)
	r.writeString(">")
}
func (r *htmlRenderer) openStyled(name, css string) {
	if css == "" {
		r.open(name)
	} else {
		r.open(name, htmlAttr{"style", css})
	}
}
func (r *htmlRenderer) void(name string, attrs ...htmlAttr) {
	r.writeString("<")
	r.writeString(name)
	r.writeAttrs(attrs)
	r.writeString(">")
}
func (r *htmlRenderer) close(name string)      { r.writeString("</" + name + ">") }
func (r *htmlRenderer) text(s string)          { r.writeString(html.EscapeString(s)) }
func (r *htmlRenderer) tagText(name, s string) { r.open(name); r.text(s); r.close(name) }
func (r *htmlRenderer) writeAttrs(attrs []htmlAttr) {
	for _, a := range attrs {
		if a.name == "" {
			continue
		}
		r.writeString(" " + a.name + "=\"" + html.EscapeString(a.value) + "\"")
	}
}

func fontStyle(v any) string {
	f, ok := asFont(v)
	if !ok {
		return ""
	}
	var p []string
	if f.Name != "" {
		p = append(p, "font-family:\""+cssString(f.Name)+"\"")
	}
	if f.Size > 0 {
		p = append(p, "font-size:"+formatFloat(f.Size)+"pt")
	}
	if f.Color != "" {
		p = append(p, "color:"+cssColor(f.Color))
	}
	if f.BgColor != "" {
		p = append(p, "background-color:"+cssColor(f.BgColor))
	}
	if f.Bold {
		p = append(p, "font-weight:700")
	}
	if f.Italic {
		p = append(p, "font-style:italic")
	}
	if f.Underline != "" && f.Underline != style.UnderlineNone {
		p = append(p, "text-decoration:underline")
	}
	if f.Strikethrough || f.DoubleStrikethrough {
		p = append(p, "text-decoration:line-through")
	}
	if f.SuperScript {
		p = append(p, "vertical-align:super;font-size:smaller")
	}
	if f.SubScript {
		p = append(p, "vertical-align:sub;font-size:smaller")
	}
	if f.SmallCaps {
		p = append(p, "font-variant:small-caps")
	}
	if f.AllCaps {
		p = append(p, "text-transform:uppercase")
	}
	if f.Hidden {
		p = append(p, "display:none")
	}
	if f.RTL {
		p = append(p, "direction:rtl")
	}
	return strings.Join(p, ";")
}
func paragraphStyle(v any) string {
	p, ok := asParagraph(v)
	if !ok {
		return ""
	}
	var out []string
	if p.Alignment != "" {
		out = append(out, "text-align:"+cssAlign(p.Alignment))
	}
	if p.Indentation.Left != 0 {
		out = append(out, "margin-left:"+twipsPx(p.Indentation.Left))
	}
	if p.Indentation.Right != 0 {
		out = append(out, "margin-right:"+twipsPx(p.Indentation.Right))
	}
	if p.Indentation.FirstLine != 0 {
		out = append(out, "text-indent:"+twipsPx(p.Indentation.FirstLine))
	}
	if p.Spacing.Before != 0 {
		out = append(out, "margin-top:"+twipsPx(p.Spacing.Before))
	}
	if p.Spacing.After != 0 {
		out = append(out, "margin-bottom:"+twipsPx(p.Spacing.After))
	}
	if p.PageBreakBefore {
		out = append(out, "break-before:page")
	}
	if p.Bidi {
		out = append(out, "direction:rtl")
	}
	if p.Shading.Fill != "" {
		out = append(out, "background-color:"+cssColor(p.Shading.Fill))
	}
	return strings.Join(out, ";")
}
func tableStyle(t style.Table) string {
	out := []string{"border-collapse:collapse"}
	if t.Width > 0 {
		out = append(out, "width:"+twipsPx(t.Width))
	}
	if t.Shading.Fill != "" {
		out = append(out, "background-color:"+cssColor(t.Shading.Fill))
	}
	return strings.Join(out, ";")
}
func cellStyle(c style.Cell) string {
	var out []string
	if c.Width > 0 {
		out = append(out, "width:"+twipsPx(c.Width))
	}
	if c.VAlign != "" {
		out = append(out, "vertical-align:"+cssVAlign(c.VAlign))
	}
	if c.NoWrap {
		out = append(out, "white-space:nowrap")
	}
	if c.BgColor != "" {
		out = append(out, "background-color:"+cssColor(c.BgColor))
	}
	if c.Shading.Fill != "" {
		out = append(out, "background-color:"+cssColor(c.Shading.Fill))
	}
	if b := c.Borders.Top; b.Style != "" && b.Style != "nil" {
		out = append(out, "border-top:"+borderCSS(b))
	}
	if b := c.Borders.Right; b.Style != "" && b.Style != "nil" {
		out = append(out, "border-right:"+borderCSS(b))
	}
	if b := c.Borders.Bottom; b.Style != "" && b.Style != "nil" {
		out = append(out, "border-bottom:"+borderCSS(b))
	}
	if b := c.Borders.Left; b.Style != "" && b.Style != "nil" {
		out = append(out, "border-left:"+borderCSS(b))
	}
	return strings.Join(out, ";")
}
func asFont(v any) (style.Font, bool) {
	switch x := v.(type) {
	case style.Font:
		return x, true
	case *style.Font:
		if x != nil {
			return *x, true
		}
	}
	return style.Font{}, false
}
func asParagraph(v any) (style.Paragraph, bool) {
	switch x := v.(type) {
	case style.Paragraph:
		return x, true
	case *style.Paragraph:
		if x != nil {
			return *x, true
		}
	}
	return style.Paragraph{}, false
}
func cssString(s string) string {
	return strings.ReplaceAll(strings.ReplaceAll(s, "\\", "\\\\"), "\"", "\\\"")
}
func cssColor(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || strings.ContainsRune("#%,. -+", r) {
			continue
		}
		return ""
	}
	return s
}
func cssAlign(s string) string {
	switch strings.ToLower(s) {
	case "center":
		return "center"
	case "right":
		return "right"
	case "both", "justify":
		return "justify"
	default:
		return "left"
	}
}
func cssVAlign(s string) string {
	switch strings.ToLower(s) {
	case "center", "middle":
		return "middle"
	case "bottom":
		return "bottom"
	default:
		return "top"
	}
}
func twipsPx(v int) string         { return formatFloat(float64(v)/15) + "px" }
func formatFloat(v float64) string { return strconv.FormatFloat(v, 'f', -1, 64) }
func borderCSS(b style.Border) string {
	color := b.Color
	if color == "" {
		color = "#000"
	}
	width := float64(b.Size) / 8
	if width <= 0 {
		width = 1
	}
	return formatFloat(width) + "pt " + borderStyle(b.Style) + " " + cssColor(color)
}
func borderStyle(s string) string {
	switch strings.ToLower(s) {
	case "double":
		return "double"
	case "dashed":
		return "dashed"
	case "dotted":
		return "dotted"
	default:
		return "solid"
	}
}

func safeURL(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" {
		return false
	}
	lower := strings.ToLower(s)
	for _, prefix := range []string{"javascript:", "vbscript:", "data:"} {
		if strings.HasPrefix(lower, prefix) {
			return false
		}
	}
	return !strings.ContainsAny(s, "\r\n\x00")
}

const defaultHTMLCSS = ".goword-page-break{break-before:page;height:0}.goword-unsupported{color:#a00;font-style:italic}.goword-comment,.goword-footnote,.goword-endnote{margin:.5em 0;padding:.4em;border-left:3px solid #aaa}.goword-textbox,.goword-shape{padding:.25em}"

func (r *htmlRenderer) writeString(s string) {
	if r.err != nil || len(s) == 0 {
		return
	}
	n, err := io.WriteString(r.out, s)
	if err != nil {
		r.err = err
		return
	}
	if n != len(s) {
		r.err = io.ErrShortWrite
	}
}
