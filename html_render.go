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
	IncludeCSS            bool // with Standalone, include built-in CSS and section page geometry
	Strict                bool
	RevisionMode          HTMLRevisionMode
}

// HTMLRevisionMode controls how tracked insertions and deletions are shown.
type HTMLRevisionMode string

const (
	HTMLRevisionFinal    HTMLRevisionMode = "final"
	HTMLRevisionOriginal HTMLRevisionMode = "original"
	HTMLRevisionMarkup   HTMLRevisionMode = "markup"
)

// HTMLDiagnostic describes a lossy or unsupported conversion.
type HTMLDiagnostic struct {
	ElementType string
	Message     string
	Part        string
	Path        string
	ID          string
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
	opts             HTMLOptions
	out              io.Writer
	err              error
	diagnostics      []HTMLDiagnostic
	document         *Document
	listCounts       map[htmlListKey]int
	diagnosticKeys   map[string]bool
	footnoteIDs      map[int]bool
	endnoteIDs       map[int]bool
	commentIDs       map[int]bool
	noteRefCounts    map[string]int
	commentRefCounts map[int]int
}

func newHTMLRenderer(out io.Writer, opts HTMLOptions) *htmlRenderer {
	return &htmlRenderer{opts: normalizeHTMLOptions(opts), out: out}
}

type htmlAttr struct{ name, value string }

func (r *htmlRenderer) renderDocument(d *Document) {
	r.document = d
	r.listCounts = make(map[htmlListKey]int)
	r.diagnosticKeys = make(map[string]bool)
	r.footnoteIDs, r.endnoteIDs, r.commentIDs = make(map[int]bool), make(map[int]bool), make(map[int]bool)
	r.noteRefCounts, r.commentRefCounts = make(map[string]int), make(map[int]int)
	for _, note := range d.GetFootnotes() {
		if note != nil && note.NoteID >= 0 {
			r.footnoteIDs[note.NoteID] = true
		}
	}
	for _, note := range d.GetEndnotes() {
		if note != nil && note.NoteID >= 0 {
			r.endnoteIDs[note.NoteID] = true
		}
	}
	for _, comment := range d.GetComments() {
		if comment != nil && comment.CommentID >= 0 {
			r.commentIDs[comment.CommentID] = true
		}
	}
	var pageNames []string
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
			r.void("meta", htmlAttr{"name", "viewport"}, htmlAttr{"content", "width=device-width, initial-scale=1"})
			r.writeString("<style>")
			r.writeString(defaultHTMLCSS)
			r.writeString(pageHTMLCSS)
			pageNames = r.writePageRules(d.sections)
			r.writeString("</style>")
		}
		r.writeString("</head>")
		if r.opts.IncludeCSS {
			attrs := []htmlAttr{{"class", "goword-document"}}
			if repeatHeader, repeatFooter := repeatableDefaultHeaderFooter(d.sections); repeatHeader {
				attrs = append(attrs, htmlAttr{"data-repeat-header", "true"})
				if repeatFooter {
					attrs = append(attrs, htmlAttr{"data-repeat-footer", "true"})
				}
			} else if repeatFooter {
				attrs = append(attrs, htmlAttr{"data-repeat-footer", "true"})
			}
			r.open("body", attrs...)
		} else {
			r.open("body")
		}
	}
	multiSection := len(d.sections) > 1
	pageLayout := r.opts.Standalone && r.opts.IncludeCSS
	for i, sec := range d.sections {
		if sec == nil {
			continue
		}
		if multiSection || pageLayout {
			attrs := []htmlAttr{}
			if sec.Style.BreakType != "" {
				attrs = append(attrs, htmlAttr{"data-break-type", sec.Style.BreakType})
			}
			if pageLayout {
				attrs = append(attrs, htmlAttr{"class", "goword-page"}, htmlAttr{"style", sectionPageGeometry(sec.Style).screenStyle() + ";page:" + pageNames[i]})
			}
			r.open("section", attrs...)
		}
		if r.opts.IncludeHeadersFooters {
			for _, h := range sec.Headers {
				attrs := []htmlAttr{{"data-type", h.HeaderType}}
				if pageLayout {
					attrs = append(attrs, htmlAttr{"class", "goword-header"}, htmlAttr{"data-section", strconv.Itoa(i + 1)})
				}
				r.open("header", attrs...)
				r.renderElements(h.Elements())
				r.close("header")
			}
		}
		r.renderElements(sec.Elements())
		if r.opts.IncludeHeadersFooters {
			for _, f := range sec.Footers {
				attrs := []htmlAttr{{"data-type", f.HeaderType}}
				if pageLayout {
					attrs = append(attrs, htmlAttr{"class", "goword-footer"}, htmlAttr{"data-section", strconv.Itoa(i + 1)})
				}
				r.open("footer", attrs...)
				r.renderElements(f.Elements())
				r.close("footer")
			}
		}
		if multiSection || pageLayout {
			r.close("section")
		}
	}
	r.renderDocumentNotes(d)
	if r.opts.Standalone {
		r.writeString("</body></html>")
	}
}

func (r *htmlRenderer) renderDocumentNotes(d *Document) {
	seenFootnotes := make(map[int]bool)
	for _, note := range d.GetFootnotes() {
		if note == nil || note.NoteID < 0 || seenFootnotes[note.NoteID] {
			continue
		}
		seenFootnotes[note.NoteID] = true
		r.open("aside", htmlAttr{"id", fmt.Sprintf("goword-footnote-%d", note.NoteID)}, htmlAttr{"class", "goword-footnote"})
		r.renderElements(note.Elements())
		if r.noteRefCounts[fmt.Sprintf("goword-footnote-ref-%d", note.NoteID)] > 0 {
			r.open("a", htmlAttr{"class", "goword-note-backlink"}, htmlAttr{"href", fmt.Sprintf("#goword-footnote-ref-%d", note.NoteID)})
			r.text("back")
			r.close("a")
		}
		r.close("aside")
	}
	seenEndnotes := make(map[int]bool)
	for _, note := range d.GetEndnotes() {
		if note == nil || note.NoteID < 0 || seenEndnotes[note.NoteID] {
			continue
		}
		seenEndnotes[note.NoteID] = true
		r.open("aside", htmlAttr{"id", fmt.Sprintf("goword-endnote-%d", note.NoteID)}, htmlAttr{"class", "goword-endnote"})
		r.renderElements(note.Elements())
		if r.noteRefCounts[fmt.Sprintf("goword-endnote-ref-%d", note.NoteID)] > 0 {
			r.open("a", htmlAttr{"class", "goword-note-backlink"}, htmlAttr{"href", fmt.Sprintf("#goword-endnote-ref-%d", note.NoteID)})
			r.text("back")
			r.close("a")
		}
		r.close("aside")
	}
	seenComments := make(map[int]bool)
	for _, comment := range d.GetComments() {
		if comment == nil || comment.CommentID < 0 || seenComments[comment.CommentID] {
			continue
		}
		seenComments[comment.CommentID] = true
		attrs := []htmlAttr{{"id", fmt.Sprintf("goword-comment-%d", comment.CommentID)}, {"class", "goword-comment"}}
		if comment.Author != "" {
			attrs = append(attrs, htmlAttr{"data-author", comment.Author})
		}
		if comment.Date != "" {
			attrs = append(attrs, htmlAttr{"data-date", comment.Date})
		}
		r.open("aside", attrs...)
		if comment.Author != "" {
			r.open("p", htmlAttr{"class", "goword-comment-author"})
			r.text(comment.Author)
			r.close("p")
		}
		r.renderElements(comment.Elements())
		if r.commentRefCounts[comment.CommentID] > 0 {
			r.open("a", htmlAttr{"class", "goword-comment-backlink"}, htmlAttr{"href", fmt.Sprintf("#goword-comment-ref-%d", comment.CommentID)})
			r.text("back")
			r.close("a")
		}
		r.close("aside")
	}
}

func repeatableDefaultHeaderFooter(sections []*element.Section) (bool, bool) {
	if len(sections) != 1 || sections[0] == nil {
		return false, false
	}
	var headers, footers int
	for _, h := range sections[0].Headers {
		if h == nil || h.HeaderType != element.HeaderAuto {
			return false, false
		}
		headers++
	}
	for _, f := range sections[0].Footers {
		if f == nil || f.HeaderType != element.HeaderAuto {
			return false, false
		}
		footers++
	}
	return headers == 1, footers == 1
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
		r.openStyled("p", paragraphTextStyle(v.ParagraphStyle, []element.Element{v}))
		r.renderRevisionText(v.Content, v.GetTrackChange(), v.FontStyle)
		r.close("p")
	case *element.PreserveText:
		r.open("p")
		r.renderText(v.Content, v.FontStyle)
		r.close("p")
	case *element.TextRun:
		if hasBlockInline(v.Elements()) {
			r.renderMixedTextRun(v)
		} else {
			r.openStyled("p", paragraphTextStyle(v.ParagraphStyle, v.Elements()))
			r.renderInlineElements(v.Elements())
			r.close("p")
		}
	case *element.ListItem:
		r.open("p")
		r.renderRevisionText(v.Text, v.GetTrackChange(), v.FontStyle)
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
		if v.Run != nil {
			// A loaded heading's run formatting overrides the browser's h1-h6
			// defaults, including the font size used to calculate its line box.
			css := paragraphTextStyle(v.Run.ParagraphStyle, v.Run.Elements())
			attrs = append(attrs, htmlAttr{"style", "font-size:inherit;font-weight:normal;" + css})
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
	case *element.Field:
		r.renderField(v)
	case *element.NoteReference:
		r.renderNoteReference(v)
	case *element.CommentReference:
		r.renderCommentReference(v)
	case *element.SDT:
		r.renderSDT(v)
	case *element.Link:
		r.open("p")
		r.renderRevisionLink(v)
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
		r.open("div", htmlAttr{"class", "goword-page-break"})
		r.close("div")
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
		attrs := []htmlAttr{{"class", "goword-textbox"}}
		if v.BoxStyle.Width > 0 {
			attrs = append(attrs, htmlAttr{"data-width", strconv.Itoa(v.BoxStyle.Width)})
		}
		if v.BoxStyle.Height > 0 {
			attrs = append(attrs, htmlAttr{"data-height", strconv.Itoa(v.BoxStyle.Height)})
		}
		if css := textBoxStyle(v.BoxStyle); css != "" {
			attrs = append(attrs, htmlAttr{"style", css})
		}
		r.open("div", attrs...)
		r.renderElements(v.Elements())
		r.close("div")
	case *element.DMLShape:
		if v.Diagnostic != "" {
			r.recordDiagnostic(v, v.Diagnostic, "word/document.xml", v.ElementID)
		}
		attrs := []htmlAttr{{"class", "goword-shape"}, {"data-shape", v.PrstGeom}}
		if v.Width > 0 {
			attrs = append(attrs, htmlAttr{"data-width", strconv.Itoa(v.Width)})
		}
		if v.Height > 0 {
			attrs = append(attrs, htmlAttr{"data-height", strconv.Itoa(v.Height)})
		}
		if v.FillColor != "" {
			attrs = append(attrs, htmlAttr{"data-fill", v.FillColor})
		}
		if v.LineColor != "" {
			attrs = append(attrs, htmlAttr{"data-line", v.LineColor})
		}
		if css := shapeStyle(v); css != "" {
			attrs = append(attrs, htmlAttr{"style", css})
		}
		r.open("div", attrs...)
		r.renderElements(v.Elements())
		r.close("div")
	case *element.Chart:
		r.renderChart(v)
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
			r.renderRevisionText(v.Content, v.GetTrackChange(), v.FontStyle)
		case *element.PreserveText:
			r.renderRevisionText(v.Content, v.GetTrackChange(), v.FontStyle)
		case *element.Link:
			r.renderRevisionLink(v)
		case *element.Image:
			r.renderImage(v)
		case *element.TextBreak:
			r.void("br")
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
		case *element.Field:
			r.renderField(v)
		case *element.NoteReference:
			r.renderNoteReference(v)
		case *element.CommentReference:
			r.renderCommentReference(v)
		case *element.SDT:
			r.renderSDT(v)
		case *element.TextBox:
			r.renderTextBoxInline(v)
		case *element.DMLShape:
			r.renderShapeInline(v)
		case *element.Chart:
			r.renderChart(v)
		default:
			r.unsupported(el, "inline element is not mapped to HTML")
		}
	}
}

func hasBlockInline(elements []element.Element) bool {
	for _, el := range elements {
		if _, ok := el.(*element.Chart); ok {
			return true
		}
	}
	return false
}

func (r *htmlRenderer) renderMixedTextRun(run *element.TextRun) {
	open := false
	for _, el := range run.Elements() {
		if _, ok := el.(*element.Chart); ok {
			if open {
				r.close("p")
				open = false
			}
			r.renderChart(el.(*element.Chart))
			continue
		}
		if !open {
			r.openStyled("p", paragraphTextStyle(run.ParagraphStyle, run.Elements()))
			open = true
		}
		r.renderInlineElements([]element.Element{el})
	}
	if open {
		r.close("p")
	}
}

func (r *htmlRenderer) renderChart(chart *element.Chart) {
	if chart == nil {
		return
	}
	series := chart.Series
	if len(series) == 0 && (len(chart.Categories) > 0 || len(chart.Values) > 0) {
		series = []element.ChartSeries{{Name: chart.SeriesName, Categories: chart.Categories, Values: chart.Values}}
	}
	attrs := []htmlAttr{{"class", "goword-chart"}}
	if chart.ChartType != "" {
		attrs = append(attrs, htmlAttr{"data-chart-type", chart.ChartType})
	}
	r.open("figure", attrs...)
	if chart.Style.Title != "" {
		r.tagText("figcaption", chart.Style.Title)
	}
	for _, item := range series {
		for _, missing := range item.MissingValues {
			if missing {
				r.recordDiagnostic(chart, "chart cache contains missing values", chart.DocPart, "")
				break
			}
		}
	}
	if len(series) == 0 {
		r.recordDiagnostic(chart, "chart has no cached series data", "word/charts", chart.ElementID)
		r.open("p", htmlAttr{"class", "goword-chart-fallback"})
		r.text("Chart data is unavailable.")
		r.close("p")
		r.close("figure")
		return
	}
	r.open("table", htmlAttr{"class", "goword-chart-data"})
	r.open("thead")
	r.open("tr")
	r.tagTextAttrs("th", []htmlAttr{{"scope", "col"}}, "Category")
	for i, item := range series {
		name := strings.TrimSpace(item.Name)
		if name == "" && i == 0 {
			name = strings.TrimSpace(chart.SeriesName)
		}
		if name == "" {
			name = "Series " + strconv.Itoa(i+1)
		}
		attrs := []htmlAttr{{"scope", "col"}}
		if item.SecondaryAxis {
			attrs = append(attrs, htmlAttr{"data-secondary-axis", "true"})
		}
		r.tagTextAttrs("th", attrs, name)
	}
	r.close("tr")
	r.close("thead")
	r.open("tbody")
	rowCount := len(chart.Categories)
	for _, item := range series {
		if len(item.Categories) > rowCount {
			rowCount = len(item.Categories)
		}
		if len(item.Values) > rowCount {
			rowCount = len(item.Values)
		}
	}
	for row := 0; row < rowCount; row++ {
		category := ""
		if row < len(chart.Categories) {
			category = chart.Categories[row]
		}
		if category == "" {
			for _, item := range series {
				if row < len(item.Categories) && item.Categories[row] != "" {
					category = item.Categories[row]
					break
				}
			}
		}
		if category == "" {
			category = "Category " + strconv.Itoa(row+1)
		}
		r.open("tr")
		r.tagTextAttrs("th", []htmlAttr{{"scope", "row"}}, category)
		for _, item := range series {
			if row < len(item.Values) {
				value := strconv.FormatFloat(item.Values[row], 'f', -1, 64)
				if row < len(item.MissingValues) && item.MissingValues[row] {
					value = ""
				}
				r.tagText("td", value)
			} else {
				r.tagText("td", "")
			}
		}
		r.close("tr")
	}
	r.close("tbody")
	r.close("table")
	r.close("figure")
}

func (r *htmlRenderer) renderTextBoxInline(v *element.TextBox) {
	attrs := []htmlAttr{{"class", "goword-textbox"}}
	if v.BoxStyle.Width > 0 {
		attrs = append(attrs, htmlAttr{"data-width", strconv.Itoa(v.BoxStyle.Width)})
	}
	if v.BoxStyle.Height > 0 {
		attrs = append(attrs, htmlAttr{"data-height", strconv.Itoa(v.BoxStyle.Height)})
	}
	if css := textBoxStyle(v.BoxStyle); css != "" {
		attrs = append(attrs, htmlAttr{"style", css})
	}
	r.open("span", attrs...)
	r.renderInlineContainer(v.Elements())
	r.close("span")
}

func (r *htmlRenderer) renderShapeInline(v *element.DMLShape) {
	if v.Diagnostic != "" {
		r.recordDiagnostic(v, v.Diagnostic, "word/document.xml", v.ElementID)
	}
	attrs := []htmlAttr{{"class", "goword-shape"}, {"data-shape", v.PrstGeom}}
	if v.Width > 0 {
		attrs = append(attrs, htmlAttr{"data-width", strconv.Itoa(v.Width)})
	}
	if v.Height > 0 {
		attrs = append(attrs, htmlAttr{"data-height", strconv.Itoa(v.Height)})
	}
	if v.FillColor != "" {
		attrs = append(attrs, htmlAttr{"data-fill", v.FillColor})
	}
	if v.LineColor != "" {
		attrs = append(attrs, htmlAttr{"data-line", v.LineColor})
	}
	if css := shapeStyle(v); css != "" {
		attrs = append(attrs, htmlAttr{"style", css})
	}
	r.open("span", attrs...)
	r.renderInlineContainer(v.Elements())
	r.close("span")
}

func (r *htmlRenderer) renderInlineContainer(elements []element.Element) {
	for _, el := range elements {
		switch v := el.(type) {
		case *element.Text:
			r.renderRevisionText(v.Content, v.GetTrackChange(), v.FontStyle)
		case *element.TextRun:
			r.renderInlineElements(v.Elements())
		default:
			r.renderInlineElements([]element.Element{v})
		}
	}
}

func (r *htmlRenderer) revisionVisible(tc *element.TrackChange) bool {
	if tc == nil {
		return true
	}
	mode := r.opts.RevisionMode
	if mode == "" {
		mode = HTMLRevisionFinal
	}
	if mode == HTMLRevisionOriginal {
		return tc.ChangeType == "del"
	}
	return tc.ChangeType != "del" || mode == HTMLRevisionMarkup
}

func (r *htmlRenderer) renderRevisionText(text string, tc *element.TrackChange, font any) {
	if !r.revisionVisible(tc) {
		return
	}
	mode := r.opts.RevisionMode
	if mode == "" {
		mode = HTMLRevisionFinal
	}
	if mode == HTMLRevisionMarkup && tc != nil {
		tag := "ins"
		if tc.ChangeType == "del" {
			tag = "del"
		}
		attrs := []htmlAttr{}
		if tc.Author != "" {
			attrs = append(attrs, htmlAttr{"data-author", tc.Author})
		}
		if tc.Date != "" {
			attrs = append(attrs, htmlAttr{"data-date", tc.Date})
		}
		r.open(tag, attrs...)
		r.renderText(text, font)
		r.close(tag)
		return
	}
	r.renderText(text, font)
}

func (r *htmlRenderer) renderRevisionLink(v *element.Link) {
	tc := v.GetTrackChange()
	if !r.revisionVisible(tc) {
		return
	}
	mode := r.opts.RevisionMode
	if mode == "" {
		mode = HTMLRevisionFinal
	}
	if mode == HTMLRevisionMarkup && tc != nil {
		tag := "ins"
		if tc.ChangeType == "del" {
			tag = "del"
		}
		attrs := []htmlAttr{}
		if tc.Author != "" {
			attrs = append(attrs, htmlAttr{"data-author", tc.Author})
		}
		if tc.Date != "" {
			attrs = append(attrs, htmlAttr{"data-date", tc.Date})
		}
		r.open(tag, attrs...)
		r.renderLink(v)
		r.close(tag)
		return
	}
	r.renderLink(v)
}

func (r *htmlRenderer) renderField(v *element.Field) {
	attrs := []htmlAttr{{"class", "goword-field"}, {"data-field", strings.ToUpper(strings.TrimSpace(v.FieldType))}}
	if v.Instruction != "" {
		attrs = append(attrs, htmlAttr{"data-instruction", v.Instruction})
	}
	r.open("span", attrs...)
	if v.Diagnostic != "" {
		r.recordDiagnostic(v, v.Diagnostic, v.DocPart, v.ElementID)
	}
	if v.HasResult || v.Text != "" {
		r.renderText(v.Text, v.FontStyle)
	} else {
		r.unsupportedWith(v, "field has no cached result", "word/document.xml", "")
		r.text("[" + strings.ToUpper(strings.TrimSpace(v.FieldType)) + "]")
	}
	r.close("span")
}

func (r *htmlRenderer) renderNoteReference(v *element.NoteReference) {
	kind := "footnote"
	if v.Kind == "endnote" {
		kind = "endnote"
	}
	targetID := fmt.Sprintf("goword-%s-%d", kind, v.NoteID)
	refBase := fmt.Sprintf("goword-%s-ref-%d", kind, v.NoteID)
	r.noteRefCounts[refBase]++
	refID := refBase
	if r.noteRefCounts[refBase] > 1 {
		refID = fmt.Sprintf("%s-%d", refBase, r.noteRefCounts[refBase])
	}
	id := targetID
	valid := (kind == "footnote" && r.footnoteIDs[v.NoteID]) || (kind == "endnote" && r.endnoteIDs[v.NoteID])
	if !valid {
		r.recordDiagnostic(v, "reference points to a missing or invalid note", v.DocPart, v.ElementID)
	}
	r.open("sup", htmlAttr{"class", "goword-note-reference"}, htmlAttr{"data-note-id", strconv.Itoa(v.NoteID)}, htmlAttr{"data-note-kind", kind})
	if valid {
		r.open("a", htmlAttr{"id", refID}, htmlAttr{"href", "#" + id})
	} else {
		r.open("span", htmlAttr{"class", "goword-missing-reference"})
	}
	r.text(strconv.Itoa(v.NoteID))
	if valid {
		r.close("a")
	} else {
		r.close("span")
	}
	r.close("sup")
}

func (r *htmlRenderer) renderCommentReference(v *element.CommentReference) {
	id := fmt.Sprintf("goword-comment-%d", v.CommentID)
	r.commentRefCounts[v.CommentID]++
	refID := fmt.Sprintf("goword-comment-ref-%d", v.CommentID)
	if r.commentRefCounts[v.CommentID] > 1 {
		refID = fmt.Sprintf("%s-%d", refID, r.commentRefCounts[v.CommentID])
	}
	valid := r.commentIDs[v.CommentID]
	if !valid {
		r.recordDiagnostic(v, "reference points to a missing or invalid comment", v.DocPart, v.ElementID)
	}
	r.open("sup", htmlAttr{"class", "goword-comment-reference"}, htmlAttr{"data-comment-id", strconv.Itoa(v.CommentID)})
	if valid {
		r.open("a", htmlAttr{"id", refID}, htmlAttr{"href", "#" + id})
	} else {
		r.open("span", htmlAttr{"class", "goword-missing-reference"})
	}
	r.text(strconv.Itoa(v.CommentID))
	if valid {
		r.close("a")
	} else {
		r.close("span")
	}
	r.close("sup")
}

func (r *htmlRenderer) renderSDT(v *element.SDT) {
	tc := v.GetTrackChange()
	if !r.revisionVisible(tc) {
		return
	}
	if v.Diagnostic != "" {
		r.recordDiagnostic(v, v.Diagnostic, v.DocPart, v.ElementID)
	}
	markup := r.opts.RevisionMode == HTMLRevisionMarkup && tc != nil
	if markup {
		tag := "ins"
		if tc.ChangeType == "del" {
			tag = "del"
		}
		attrs := []htmlAttr{}
		if tc.Author != "" {
			attrs = append(attrs, htmlAttr{"data-author", tc.Author})
		}
		if tc.Date != "" {
			attrs = append(attrs, htmlAttr{"data-date", tc.Date})
		}
		r.open(tag, attrs...)
	}
	attrs := []htmlAttr{{"class", "goword-sdt"}, {"data-sdt-type", v.SDTType}}
	if v.Alias != "" {
		attrs = append(attrs, htmlAttr{"data-alias", v.Alias})
	}
	if v.Tag != "" {
		attrs = append(attrs, htmlAttr{"data-tag", v.Tag})
	}
	if v.DateFormat != "" {
		attrs = append(attrs, htmlAttr{"data-date-format", v.DateFormat})
	}
	if len(v.ListItems) > 0 {
		options := make([]string, 0, len(v.ListItems))
		for _, item := range v.ListItems {
			options = append(options, item.Value+":"+item.DisplayText)
		}
		attrs = append(attrs, htmlAttr{"data-options", strings.Join(options, "|")})
	}
	if v.ShowingPlcHdr {
		attrs = append(attrs, htmlAttr{"data-placeholder", v.Value})
	}
	if v.IsBlock {
		r.open("div", attrs...)
	} else {
		r.open("span", attrs...)
	}
	if v.SDTType == element.SDTTypeCheckbox {
		checkboxAttrs := []htmlAttr{{"type", "checkbox"}, {"disabled", "disabled"}}
		if v.Checked {
			checkboxAttrs = append(checkboxAttrs, htmlAttr{"checked", "checked"})
		}
		r.void("input", checkboxAttrs...)
	}
	if len(v.Elements()) > 0 {
		if v.IsBlock {
			r.renderElements(v.Elements())
		} else {
			r.renderInlineElements(v.Elements())
		}
	} else if v.Value != "" {
		r.text(v.Value)
	} else if v.Placeholder != "" {
		r.text(v.Placeholder)
	} else {
		r.unsupportedWith(v, "content control has no visible content", v.DocPart, v.ElementID)
	}
	if v.IsBlock {
		r.close("div")
	} else {
		r.close("span")
	}
	if markup {
		tag := "ins"
		if tc.ChangeType == "del" {
			tag = "del"
		}
		r.close(tag)
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
	attrs := []htmlAttr{{"href", target}}
	if css := fontStyle(v.FontStyle); css != "" {
		// Apply explicit link appearance to the anchor itself. A child span
		// cannot cancel an ancestor's text decoration or its visited color.
		attrs = append(attrs, htmlAttr{"style", css})
	}
	r.open("a", attrs...)
	r.text(v.Text)
	r.close("a")
}

func (r *htmlRenderer) renderTable(t *element.Table) {
	rowspan, skip := tableVerticalMerges(t)
	attrs := []htmlAttr{}
	if st := tableStyle(t.Style); st != "" {
		attrs = append(attrs, htmlAttr{"style", st})
	}
	r.open("table", attrs...)
	headerRows := 0
	for headerRows < len(t.Rows) && t.Rows[headerRows] != nil && t.Rows[headerRows].IsHeader() {
		headerRows++
	}
	bodyTag := "tbody"
	if headerRows > 0 {
		bodyTag = "thead"
	}
	r.open(bodyTag)
	for i, row := range t.Rows {
		if headerRows > 0 && i == headerRows {
			r.close("thead")
			r.open("tbody")
		}
		rowAttrs := []htmlAttr{}
		if st := tableRowStyle(row.Style); st != "" {
			rowAttrs = append(rowAttrs, htmlAttr{"style", st})
		}
		r.open("tr", rowAttrs...)
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
	r.close(bodyTag)
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
	case style.NumberChineseCounting, style.NumberChineseCountingThousand:
		attrs = append(attrs, htmlAttr{"style", "list-style-type:cjk-ideographic"})
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
		var css string
		switch v := elements[i].(type) {
		case *element.ListItem:
			css = paragraphTextStyle(v.ParagraphStyle, []element.Element{v})
		case *element.ListItemRun:
			css = paragraphTextStyle(v.ParagraphStyle, v.Elements())
		}
		r.openStyled("li", css)
		switch v := elements[i].(type) {
		case *element.ListItem:
			r.renderRevisionText(v.Text, v.GetTrackChange(), v.FontStyle)
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
	width, height := img.Style.Width, img.Style.Height
	if width <= 0 && img.Style.WidthEMU > 0 {
		width = common.EMUToPixel(img.Style.WidthEMU)
	}
	if height <= 0 && img.Style.HeightEMU > 0 {
		height = common.EMUToPixel(img.Style.HeightEMU)
	}
	if width > 0 {
		attrs = append(attrs, htmlAttr{"width", formatFloat(width)})
	}
	if height > 0 {
		attrs = append(attrs, htmlAttr{"height", formatFloat(height)})
	}
	if css, supported := imageStyle(img.Style); css != "" {
		attrs = append(attrs, htmlAttr{"style", css})
		if !supported {
			r.diagnostics = append(r.diagnostics, HTMLDiagnostic{ElementType: "Image", Message: "unsupported image wrapping style"})
		}
	} else if img.Style.WrappingStyle != "" && !supportedImageWrapping(img.Style.WrappingStyle) {
		r.diagnostics = append(r.diagnostics, HTMLDiagnostic{ElementType: "Image", Message: "unsupported image wrapping style"})
	}
	r.void("img", attrs...)
}

func supportedImageWrapping(wrapping string) bool {
	switch strings.ToLower(strings.TrimSpace(wrapping)) {
	case "", style.WrappingInline, style.WrappingSquare, style.WrappingTight, style.WrappingBehind, style.WrappingInFront:
		return true
	default:
		return false
	}
}

func imageStyle(st style.Image) (string, bool) {
	wrapping := strings.ToLower(strings.TrimSpace(st.WrappingStyle))
	supported := supportedImageWrapping(wrapping)
	var out []string
	switch wrapping {
	case style.WrappingSquare, style.WrappingTight:
		switch strings.ToLower(strings.TrimSpace(st.Alignment)) {
		case "right":
			out = append(out, "float:right")
		case "center":
			out = append(out, "display:block", "margin-left:auto", "margin-right:auto")
		default:
			out = append(out, "float:left")
		}
	case style.WrappingBehind:
		out = append(out, "position:relative", "z-index:-1")
	case style.WrappingInFront:
		out = append(out, "position:relative", "z-index:1")
	default:
		if strings.EqualFold(strings.TrimSpace(st.Alignment), "center") {
			out = append(out, "display:block", "margin-left:auto", "margin-right:auto")
		}
	}
	if st.MarginTop != 0 {
		out = append(out, "margin-top:"+formatFloat(st.MarginTop)+"px")
	}
	if st.MarginRight != 0 {
		out = append(out, "margin-right:"+formatFloat(st.MarginRight)+"px")
	}
	if st.MarginBottom != 0 {
		out = append(out, "margin-bottom:"+formatFloat(st.MarginBottom)+"px")
	}
	if st.MarginLeft != 0 {
		out = append(out, "margin-left:"+formatFloat(st.MarginLeft)+"px")
	}
	if st.OffsetX != 0 || st.OffsetY != 0 {
		if !hasCSSProperty(out, "position") {
			out = append(out, "position:relative")
		}
		if st.OffsetX != 0 {
			out = append(out, "left:"+formatFloat(st.OffsetX)+"px")
		}
		if st.OffsetY != 0 {
			out = append(out, "top:"+formatFloat(st.OffsetY)+"px")
		}
	}
	return strings.Join(out, ";"), supported
}

func hasCSSProperty(properties []string, name string) bool {
	prefix := name + ":"
	for _, property := range properties {
		if strings.HasPrefix(property, prefix) {
			return true
		}
	}
	return false
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
	if len(data) == 0 && img.Source != "" && !isExternalURL(img.Source) {
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

func isExternalURL(s string) bool {
	lower := strings.ToLower(strings.TrimSpace(s))
	return strings.HasPrefix(lower, "http://") || strings.HasPrefix(lower, "https://") || strings.HasPrefix(lower, "//")
}

func textBoxStyle(st style.TextBox) string {
	var out []string
	if st.Width > 0 {
		out = append(out, "width:"+strconv.Itoa(st.Width)+"px")
	}
	if st.Height > 0 {
		out = append(out, "height:"+strconv.Itoa(st.Height)+"px")
	}
	if c := cssColor(st.BgColor); c != "" {
		out = append(out, "background-color:"+c)
	}
	if c := cssColor(st.BorderColor); c != "" {
		width := st.BorderSize
		if width <= 0 {
			width = 1
		}
		out = append(out, "border:"+strconv.Itoa(width)+"px solid "+c)
	}
	return strings.Join(out, ";")
}

func shapeStyle(s *element.DMLShape) string {
	if s == nil {
		return ""
	}
	var out []string
	if w := shapeDimension(s.Width); w > 0 {
		out = append(out, "width:"+strconv.Itoa(w)+"px")
	}
	if h := shapeDimension(s.Height); h > 0 {
		out = append(out, "height:"+strconv.Itoa(h)+"px")
	}
	if c := cssColor(s.FillColor); c != "" {
		out = append(out, "background-color:"+c)
	}
	if c := cssColor(s.LineColor); c != "" {
		width := s.LineWidth
		if width <= 0 {
			width = 1
		}
		line := formatFloat(float64(width)) + "px"
		if width >= 1000 {
			line = formatFloat(float64(width)/12700) + "pt"
		}
		out = append(out, "border:"+line+" solid "+c)
	}
	return strings.Join(out, ";")
}

func shapeDimension(v int) int {
	if v <= 0 {
		return 0
	}
	if v > 100000 {
		return int(common.EMUToPixel(int64(v)))
	}
	return v
}

func (r *htmlRenderer) unsupported(el element.Element, msg string) {
	r.unsupportedWith(el, msg, "", "")
}

func (r *htmlRenderer) unsupportedWith(el element.Element, msg, part, path string) {
	typ := "unknown"
	if el != nil {
		typ = el.Type()
	}
	id := ""
	if el != nil {
		switch v := el.(type) {
		case *element.Field:
			if part == "" {
				part = v.DocPart
			}
			if part == "" {
				part = "word/document.xml"
			}
			if path == "" {
				path = v.ElementID
			}
		case *element.Chart:
			id = v.ElementID
			if part == "" {
				part = v.DocPart
			}
			id = v.ElementID
		case *element.SDT:
			if part == "" {
				part = v.DocPart
			}
			if part == "" {
				part = "word/document.xml"
			}
			if path == "" {
				path = v.ElementID
			}
			id = v.ElementID
		}
	}
	r.recordDiagnosticFields(typ, msg, part, path, id)
	r.open("span", htmlAttr{"class", "goword-unsupported"}, htmlAttr{"data-element", typ})
	r.text("[Unsupported: " + typ + "]")
	r.close("span")
}

func (r *htmlRenderer) recordDiagnostic(el element.Element, msg, part, path string) {
	typ, id := "unknown", ""
	if el != nil {
		typ = el.Type()
		switch v := el.(type) {
		case *element.NoteReference:
			id = v.ElementID
		case *element.CommentReference:
			id = v.ElementID
		case *element.Field:
			id = v.ElementID
			if part == "" {
				part = v.DocPart
			}
			if path == "" {
				path = v.ElementID
			}
		case *element.SDT:
			id = v.ElementID
			if part == "" {
				part = v.DocPart
			}
			if path == "" {
				path = v.ElementID
			}
		case *element.Chart:
			id = v.ElementID
			if part == "" {
				part = v.DocPart
			}
			if path == "" {
				path = v.ElementID
			}
		}
	}
	r.recordDiagnosticFields(typ, msg, part, path, id)
}

func (r *htmlRenderer) recordDiagnosticFields(typ, msg, part, path, id string) {
	key := typ + "\x00" + msg + "\x00" + part + "\x00" + path + "\x00" + id
	if r.diagnosticKeys[key] {
		return
	}
	r.diagnosticKeys[key] = true
	r.diagnostics = append(r.diagnostics, HTMLDiagnostic{ElementType: typ, Message: msg, Part: part, Path: path, ID: id})
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
func (r *htmlRenderer) tagTextAttrs(name string, attrs []htmlAttr, s string) {
	r.open(name, attrs...)
	r.text(s)
	r.close(name)
}
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
	if f.Name != "" || f.FallbackFont != "" {
		families := make([]string, 0, 2)
		if f.Name != "" {
			families = append(families, "\""+cssString(f.Name)+"\"")
		}
		if f.FallbackFont != "" {
			families = append(families, "\""+cssString(f.FallbackFont)+"\"")
		}
		p = append(p, "font-family:"+strings.Join(families, ","))
	}
	if f.Size > 0 {
		p = append(p, "font-size:"+formatFloat(f.Size)+"pt")
	}
	if color := cssColor(f.Color); color != "" {
		p = append(p, "color:"+color)
	} else if strings.EqualFold(strings.TrimSpace(f.Color), "auto") {
		p = append(p, "color:inherit")
	}
	if color := cssColor(f.BgColor); color != "" {
		p = append(p, "background-color:"+color)
	}
	if f.Bold {
		p = append(p, "font-weight:700")
	}
	if f.Italic {
		p = append(p, "font-style:italic")
	}
	if f.Underline != "" {
		if f.Underline == style.UnderlineNone {
			p = append(p, "text-decoration:none")
		} else {
			p = append(p, "text-decoration:underline")
		}
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
	if f.Spacing != 0 {
		p = append(p, "letter-spacing:"+twipsPx(f.Spacing))
	}
	if f.Scale > 0 {
		p = append(p, "font-stretch:"+strconv.Itoa(f.Scale)+"%")
	}
	if f.Position != 0 {
		p = append(p, "vertical-align:"+formatFloat(float64(f.Position)/2)+"pt")
	}
	switch strings.ToLower(strings.TrimSpace(f.WhiteSpace)) {
	case "preserve", "pre", "pre-wrap":
		p = append(p, "white-space:pre-wrap")
	case "pre-line":
		p = append(p, "white-space:pre-line")
	case "nowrap":
		p = append(p, "white-space:nowrap")
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

// Keep the block's font size consistent with its text, so line-height in em
// units and list markers do not use unrelated browser defaults.
func paragraphTextStyle(para any, elements []element.Element) string {
	css := paragraphStyle(para)
	if size := paragraphFontSize(elements); size > 0 {
		if css != "" {
			css += ";"
		}
		css += "font-size:" + formatFloat(size) + "pt"
	}
	return css
}

func paragraphFontSize(elements []element.Element) float64 {
	var size float64
	for _, el := range elements {
		var font any
		switch v := el.(type) {
		case *element.Text:
			font = v.FontStyle
		case *element.Link:
			font = v.FontStyle
		case *element.ListItem:
			font = v.FontStyle
		case *element.TextRun:
			if n := paragraphFontSize(v.Elements()); n > size {
				size = n
			}
		}
		if f, ok := asFont(font); ok && f.Size > size && !f.Hidden {
			size = f.Size
		}
	}
	return size
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
	if p.Indentation.Hanging != 0 {
		out = append(out, "text-indent:"+twipsPx(-p.Indentation.Hanging))
	} else if p.Indentation.FirstLine != 0 {
		out = append(out, "text-indent:"+twipsPx(p.Indentation.FirstLine))
	}
	if p.Spacing.Before != 0 || p.Spacing.BeforeSet {
		out = append(out, "margin-top:"+twipsPx(p.Spacing.Before))
	}
	if p.Spacing.After != 0 || p.Spacing.AfterSet {
		out = append(out, "margin-bottom:"+twipsPx(p.Spacing.After))
	}
	if p.Spacing.Line > 0 {
		line := ""
		switch p.Spacing.Rule {
		case "exact":
			line = twipsPx(p.Spacing.Line)
		case "atLeast":
			line = "max(1.2em," + twipsPx(p.Spacing.Line) + ")"
		default:
			line = formatFloat(float64(p.Spacing.Line) / 240)
		}
		out = append(out, "line-height:"+line)
	}
	if len(p.Tabs) > 0 && p.Tabs[0].Pos > 0 {
		out = append(out, "tab-size:"+twipsPx(p.Tabs[0].Pos))
	}
	if p.SuppressAutoHyphens {
		out = append(out, "hyphens:none")
	}
	if p.PageBreakBefore {
		out = append(out, "break-before:page")
	}
	if p.KeepNext {
		out = append(out, "break-after:avoid")
	}
	if p.KeepLines {
		out = append(out, "break-inside:avoid")
	}
	if p.WidowControl != nil {
		if *p.WidowControl {
			out = append(out, "widows:2;orphans:2")
		} else {
			out = append(out, "widows:1;orphans:1")
		}
	}
	if p.Bidi {
		out = append(out, "direction:rtl")
	}
	if p.Shading.Fill != "" {
		out = append(out, "background-color:"+cssColor(p.Shading.Fill))
	}
	for _, b := range []struct {
		name string
		b    style.Border
	}{{"top", p.Borders.Top}, {"right", p.Borders.Right}, {"bottom", p.Borders.Bottom}, {"left", p.Borders.Left}} {
		if b.b.Style != "" && !strings.EqualFold(b.b.Style, "nil") {
			out = append(out, "border-"+b.name+":"+borderCSS(b.b))
		}
	}
	return strings.Join(out, ";")
}
func tableStyle(t style.Table) string {
	out := []string{"border-collapse:collapse"}
	if t.Width > 0 {
		out = append(out, "width:"+twipsPx(t.Width))
	}
	if strings.EqualFold(t.Layout, "fixed") {
		out = append(out, "table-layout:fixed")
	}
	if t.CellSpacing > 0 || t.CellSpacingVal > 0 {
		spacing := t.CellSpacing
		if spacing <= 0 {
			spacing = t.CellSpacingVal
		}
		out[0] = "border-collapse:separate"
		out = append(out, "border-spacing:"+twipsPx(spacing))
	}
	for _, b := range []struct {
		name string
		b    style.Border
	}{{"top", t.Borders.Top}, {"right", t.Borders.Right}, {"bottom", t.Borders.Bottom}, {"left", t.Borders.Left}} {
		if b.b.Style != "" && !strings.EqualFold(b.b.Style, "nil") {
			out = append(out, "border-"+b.name+":"+borderCSS(b.b))
		}
	}
	if t.Shading.Fill != "" {
		out = append(out, "background-color:"+cssColor(t.Shading.Fill))
	}
	return strings.Join(out, ";")
}

func tableRowStyle(row style.Row) string {
	var out []string
	if row.CantSplit {
		out = append(out, "break-inside:avoid")
	}
	if row.Height > 0 {
		out = append(out, "height:"+twipsPx(row.Height))
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
	if c.PaddingTop != 0 {
		out = append(out, "padding-top:"+twipsPx(c.PaddingTop))
	}
	if c.PaddingRight != 0 {
		out = append(out, "padding-right:"+twipsPx(c.PaddingRight))
	}
	if c.PaddingBottom != 0 {
		out = append(out, "padding-bottom:"+twipsPx(c.PaddingBottom))
	}
	if c.PaddingLeft != 0 {
		out = append(out, "padding-left:"+twipsPx(c.PaddingLeft))
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
	if s == "" || strings.EqualFold(s, "auto") {
		return ""
	}
	// OOXML RGB colors omit the '#' required by CSS.
	if len(s) == 6 && strings.Trim(s, "0123456789abcdefABCDEF") == "" {
		return "#" + s
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
