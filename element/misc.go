package element

import (
	"github.com/yunkeweb/go-word/pkg/math"
	"github.com/yunkeweb/go-word/style"
)

// TextBreak is an empty paragraph.
type TextBreak struct{ Base }

func (t *TextBreak) Type() string { return "TextBreak" }

// PageBreak is a page break.
type PageBreak struct{ Base }

func (p *PageBreak) Type() string { return "PageBreak" }

// Bookmark is a bookmark start/end pair.
type Bookmark struct {
	Base
	Name string
}

func (b *Bookmark) Type() string { return "Bookmark" }

// Title is a heading paragraph.
type Title struct {
	Base
	Text         string
	Depth        int
	Page         int
	Run          *TextRun
	BookmarkName string
}

// TextWatermark is a VML diagonal text watermark drawn in a header.
type TextWatermark struct {
	Base
	Text string
}

func (t *TextWatermark) Type() string { return "TextWatermark" }

func (t *Title) Type() string { return "Title" }

// NewTitle constructs a heading.
func NewTitle(text string, depth int) *Title {
	return &Title{Text: text, Depth: depth}
}

// ListItem is a numbered or bulleted paragraph.
type ListItem struct {
	Base
	Text           string
	Depth          int
	FontStyle      any
	ParagraphStyle any
	ListStyle      any
}

func (l *ListItem) Type() string { return "ListItem" }

// NewListItem constructs a list paragraph.
func NewListItem(text string, depth int, font, list, para any) *ListItem {
	return &ListItem{
		Text:           text,
		Depth:          depth,
		FontStyle:      normalizeFont(font),
		ParagraphStyle: normalizePara(para),
		ListStyle:      list,
	}
}

// ListItemRun is a rich list paragraph.
type ListItemRun struct {
	TextRun
	Depth     int
	ListStyle any
}

func (l *ListItemRun) Type() string { return "ListItemRun" }

// NewListItemRun constructs a rich list paragraph.
func NewListItemRun(depth int, list, para any) *ListItemRun {
	r := &ListItemRun{Depth: depth, ListStyle: list}
	r.Kind = "ListItemRun"
	r.ParagraphStyle = normalizePara(para)
	return r
}

// PreserveText is a field character run used in headers.
type PreserveText struct {
	Text
}

func (p *PreserveText) Type() string { return "PreserveText" }

func NewPreserveText(text string, font, para any) *PreserveText {
	return &PreserveText{Text: *NewText(text, font, para)}
}

// CheckBox is a checkbox form field.
type CheckBox struct {
	Text
	Name    string
	Checked bool
}

func (c *CheckBox) Type() string { return "CheckBox" }

func NewCheckBox(name, text string, font, para any) *CheckBox {
	return &CheckBox{Text: *NewText(text, font, para), Name: name}
}

// Field is a Word field (PAGE, DATE, ...).
type Field struct {
	Base
	FieldType  string
	Properties map[string]string
	Options    []string
	Text       string
	FontStyle  any
}

func (f *Field) Type() string { return "Field" }

func NewField(typ string, props map[string]string, options []string, text string) *Field {
	return &Field{FieldType: typ, Properties: props, Options: options, Text: text}
}

// Line is a VML line.
type Line struct {
	Base
	Style style.Line
}

func (l *Line) Type() string { return "Line" }

// Shape is a VML shape.
type Shape struct {
	Base
	ShapeType string
	Style     style.Shape
}

func (s *Shape) Type() string { return "Shape" }

// TextBox is a floating text frame.
type TextBox struct {
	Container
	BoxStyle style.TextBox
}

func (t *TextBox) Type() string { return "TextBox" }

func NewTextBox() *TextBox {
	tb := &TextBox{}
	tb.Kind = "TextBox"
	return tb
}

// Formula is an OMML equation.
type Formula struct {
	Base
	Math *math.Math
}

func (f *Formula) Type() string { return "Formula" }

// Footnote is a footnote container.
type Footnote struct {
	Container
	NoteID int
}

func (f *Footnote) Type() string { return "Footnote" }

func NewFootnote(para any) *Footnote {
	fn := &Footnote{}
	fn.Kind = "Footnote"
	_ = para
	return fn
}

// Endnote is an endnote container.
type Endnote struct {
	Container
	NoteID int
}

func (e *Endnote) Type() string { return "Endnote" }

func NewEndnote(para any) *Endnote {
	en := &Endnote{}
	en.Kind = "Endnote"
	_ = para
	return en
}

// SDT is a structured document tag.
type SDT struct {
	Container
	SDTType   string
	Alias     string
	Tag       string
	Value     string
	ListItems []string
}

func (s *SDT) Type() string { return "SDT" }

// TOC is a table of contents field.
type TOC struct {
	Base
	FontStyle any
	TOCStyle  any
	MinDepth  int
	MaxDepth  int
}

func (t *TOC) Type() string { return "TOC" }

// ChartSeries is one data series on a chart.
type ChartSeries struct {
	Categories    []string
	Values        []float64
	Name          string
	Kind          string
	SecondaryAxis bool
	Smooth        bool
	Marker        string
}

// Chart is a chart drawing.
type Chart struct {
	Base
	ChartType  string
	Categories []string
	Values     []float64
	Style      style.Chart
	SeriesName string
	Series     []ChartSeries
}

func (c *Chart) Type() string { return "Chart" }

// AddSeries appends a data series (PHPWord Chart::addSeries).
func (c *Chart) AddSeries(categories []string, values []float64, name ...string) {
	s := ChartSeries{Categories: categories, Values: values}
	if len(name) > 0 {
		s.Name = name[0]
	}
	c.Series = append(c.Series, s)
	if len(c.Categories) == 0 {
		c.Categories = categories
		c.Values = values
		c.SeriesName = s.Name
	}
}

// AddComboSeries appends a series with an explicit chart kind and axis.
func (c *Chart) AddComboSeries(kind string, categories []string, values []float64, name string, secondary bool) {
	s := ChartSeries{
		Categories:    categories,
		Values:        values,
		Name:          name,
		Kind:          kind,
		SecondaryAxis: secondary,
	}
	c.Series = append(c.Series, s)
	if len(c.Categories) == 0 {
		c.Categories = categories
		c.Values = values
		c.SeriesName = name
		if c.ChartType == "" {
			c.ChartType = kind
		}
	}
}

// SetDataLabels configures c:dLbls visibility and dLblPos.
func (c *Chart) SetDataLabels(opts style.DataLabelOptions) {
	c.Style.DataLabels = opts
	c.Style.DataLabelsSet = true
}

// SetMajorGridlines toggles major gridlines on both axes.
func (c *Chart) SetMajorGridlines(visible bool) {
	c.Style.ShowGridX = visible
	c.Style.ShowGridY = visible
}

// SetLegendPosition shows the legend at pos (t/b/l/r).
func (c *Chart) SetLegendPosition(pos string) {
	c.Style.ShowLegend = true
	c.Style.LegendPosition = pos
}

// SetLineSmooth enables c:smooth on line series.
func (c *Chart) SetLineSmooth(on bool) {
	c.Style.LineSmooth = on
}

// SetLineMarker sets the c:marker symbol for line series (circle, diamond, none, …).
func (c *Chart) SetLineMarker(symbol string) {
	c.Style.LineMarker = symbol
}

// Ruby is phonetic guide text.
type Ruby struct {
	Base
	BaseText   *TextRun
	RubyText   *TextRun
	Properties RubyProperties
}

func (r *Ruby) Type() string { return "Ruby" }

// Ruby alignment values (PHPWord ComplexType\RubyProperties).
const (
	RubyAlignCenter           = "center"
	RubyAlignDistributeLetter = "distributeLetter"
	RubyAlignDistributeSpace  = "distributeSpace"
	RubyAlignLeft             = "left"
	RubyAlignRight            = "right"
	RubyAlignRightVertical    = "rightVertical"
)

// RubyProperties controls ruby layout.
type RubyProperties struct {
	Alignment    string
	Hanging      int
	FontFace     string
	FontSize     float64
	Raise        float64
	BaseFontSize float64
	LanguageID   string
}

// FormField is a Word form field.
type FormField struct {
	Text
	FormType string // textinput, checkbox, dropdown
	Name     string
	Value    string
	Default  string
	Enabled  bool
	Items    []string
}

func (f *FormField) Type() string { return "FormField" }

func NewFormField(typ string, font, para any) *FormField {
	return &FormField{Text: *NewText("", font, para), FormType: typ, Enabled: true}
}

// OLEObject is an embedded OLE object.
type OLEObject struct {
	Base
	Source string
	Style  style.Image
	Media  Media
}

func (o *OLEObject) Type() string { return "OLEObject" }

func NewOLEObject(source string, st any) *OLEObject {
	o := &OLEObject{Source: source}
	applyImageStyle(&o.Style, st)
	o.Media = Media{MediaType: "object", Source: source, Target: source}
	return o
}

// Comment is a comment range.
type Comment struct {
	Container
	Author       string
	Initials     string
	Date         string
	CommentID    int
	StartElement Element
	EndElement   Element
}

func (c *Comment) Type() string { return "Comment" }

// TrackChange marks an insertion or deletion.
type TrackChange struct {
	Base
	ChangeType string // ins, del
	Author     string
	Date       string
}

func (t *TrackChange) Type() string { return "TrackChange" }
