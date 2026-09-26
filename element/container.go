package element

import (
	"github.com/yunkeweb/go-word/pkg/math"
	"github.com/yunkeweb/go-word/style"
)

// Container is a parent that holds child elements (section, cell, textrun, ...).
type Container struct {
	Base
	Kind     string
	elements []Element
}

func (c *Container) Type() string { return c.Kind }

// Elements returns child elements in document order.
func (c *Container) Elements() []Element { return c.elements }

// GetElements is the PHPWord name for Elements.
func (c *Container) GetElements() []Element { return c.elements }

// CountElements returns the number of children (PHPWord countElements).
func (c *Container) CountElements() int { return len(c.elements) }

// GetElement returns the child at index, or nil.
func (c *Container) GetElement(index int) Element {
	if index < 0 || index >= len(c.elements) {
		return nil
	}
	return c.elements[index]
}

// RemoveElement removes a child by index or by identity (PHPWord removeElement).
func (c *Container) RemoveElement(toRemove any) {
	switch v := toRemove.(type) {
	case int:
		if v >= 0 && v < len(c.elements) {
			c.elements = append(c.elements[:v], c.elements[v+1:]...)
		}
	case Element:
		for i, el := range c.elements {
			if el == v {
				c.elements = append(c.elements[:i], c.elements[i+1:]...)
				return
			}
		}
	}
}

func (c *Container) add(el Element) {
	setParent(el, c)
	c.elements = append(c.elements, el)
}

func pick2(styles []any) (any, any) {
	var a, b any
	if len(styles) > 0 {
		a = styles[0]
	}
	if len(styles) > 1 {
		b = styles[1]
	}
	return a, b
}

// AddText appends a text run wrapped in a paragraph (PHPWord addText).
func (c *Container) AddText(text string, styles ...any) *Text {
	font, para := pick2(styles)
	t := NewText(text, font, para)
	c.add(t)
	return t
}

// AddTextRun appends a rich-text paragraph.
func (c *Container) AddTextRun(styles ...any) *TextRun {
	_, para := pick2(styles)
	if len(styles) == 1 {
		para = styles[0]
	}
	tr := NewTextRun(para)
	c.add(tr)
	return tr
}

// AddTextBreak inserts count empty paragraphs (default 1).
func (c *Container) AddTextBreak(count ...int) {
	n := 1
	if len(count) > 0 && count[0] > 0 {
		n = count[0]
	}
	var font, para any
	if len(count) > 1 {
		// PHPWord: addTextBreak($count, $font, $para) — not used here.
	}
	_ = font
	_ = para
	for i := 0; i < n; i++ {
		c.add(&TextBreak{})
	}
}

// AddPageBreak inserts a page break.
func (c *Container) AddPageBreak() *PageBreak {
	p := &PageBreak{}
	c.add(p)
	return p
}

// AddLink appends a hyperlink.
func (c *Container) AddLink(target, text string, styles ...any) *Link {
	font, para := pick2(styles)
	internal := false
	if len(styles) > 2 {
		if v, ok := styles[2].(bool); ok {
			internal = v
		}
	}
	l := NewLink(target, text, font, para, internal)
	c.add(l)
	return l
}

// AddBookmark appends a bookmark.
func (c *Container) AddBookmark(name string) *Bookmark {
	b := &Bookmark{Name: name}
	c.add(b)
	return b
}

// AddTitle appends a heading. An optional page number is accepted (PHPWord addTitle).
func (c *Container) AddTitle(text string, depth int, page ...int) *Title {
	if depth < 1 {
		depth = 1
	}
	t := NewTitle(text, depth)
	if len(page) > 0 {
		t.Page = page[0]
	}
	c.add(t)
	return t
}

// AddListItem appends a numbered or bulleted paragraph.
func (c *Container) AddListItem(text string, depth int, styles ...any) *ListItem {
	font, para := pick2(styles)
	var list any
	if len(styles) > 2 {
		list = styles[2]
	}
	item := NewListItem(text, depth, font, list, para)
	c.add(item)
	return item
}

// AddListItemRun appends a rich list paragraph.
func (c *Container) AddListItemRun(depth int, listStyle any, para any) *ListItemRun {
	item := NewListItemRun(depth, listStyle, para)
	c.add(item)
	return item
}

// AddTable appends a table.
func (c *Container) AddTable(styles ...any) *Table {
	var s any
	if len(styles) > 0 {
		s = styles[0]
	}
	tbl := NewTable(s)
	c.add(tbl)
	return tbl
}

// AddImage appends an image from a filesystem path.
func (c *Container) AddImage(source string, styles ...any) *Image {
	var s any
	if len(styles) > 0 {
		s = styles[0]
	}
	img := NewImage(source, s)
	c.add(img)
	return img
}

// AddImageBytes appends an image from memory.
func (c *Container) AddImageBytes(name string, data []byte, styles ...any) *Image {
	var s any
	if len(styles) > 0 {
		s = styles[0]
	}
	img := NewImageBytes(name, data, s)
	c.add(img)
	return img
}

// AddPreserveText appends a mail-merge / header field run.
func (c *Container) AddPreserveText(text string, styles ...any) *PreserveText {
	font, para := pick2(styles)
	p := NewPreserveText(text, font, para)
	c.add(p)
	return p
}

// AddCheckBox appends a checkbox form field.
func (c *Container) AddCheckBox(name, text string, styles ...any) *CheckBox {
	font, para := pick2(styles)
	cb := NewCheckBox(name, text, font, para)
	c.add(cb)
	return cb
}

// AddField appends a Word field.
func (c *Container) AddField(typ string, properties map[string]string, options []string, text string) *Field {
	f := NewField(typ, properties, options, text)
	c.add(f)
	return f
}

// AddPageNumber inserts a PAGE field.
func (c *Container) AddPageNumber() *Field {
	return c.AddField("PAGE", nil, nil, "")
}

// AddNumPages inserts a NUMPAGES field.
func (c *Container) AddNumPages() *Field {
	return c.AddField("NUMPAGES", nil, nil, "")
}

// AddLine appends a VML line shape.
func (c *Container) AddLine(s ...style.Line) *Line {
	ln := &Line{}
	if len(s) > 0 {
		ln.Style = s[0]
	}
	c.add(ln)
	return ln
}

// AddShape appends a VML shape.
func (c *Container) AddShape(typ string, s ...style.Shape) *Shape {
	sh := &Shape{ShapeType: typ}
	if len(s) > 0 {
		sh.Style = s[0]
	}
	c.add(sh)
	return sh
}

// AddDMLShape appends a DrawingML preset shape, optionally with textbox content.
func (c *Container) AddDMLShape(prst string, width, height int, fill, line string, lineWidth int) *DMLShape {
	sh := &DMLShape{
		PrstGeom:  prst,
		Width:     width,
		Height:    height,
		FillColor: fill,
		LineColor: line,
		LineWidth: lineWidth,
	}
	sh.Kind = "DMLShape"
	c.add(sh)
	return sh
}

// AddTextBox appends a text box.
func (c *Container) AddTextBox(s ...style.TextBox) *TextBox {
	tb := NewTextBox()
	if len(s) > 0 {
		tb.BoxStyle = s[0]
	}
	c.add(tb)
	return tb
}

// AddFormula appends an OMML formula.
func (c *Container) AddFormula(m *math.Math) *Formula {
	f := &Formula{Math: m}
	c.add(f)
	return f
}

// AddMath parses a basic LaTeX expression into a native OMML formula.
func (c *Container) AddMath(formula string) *Formula {
	m, err := math.ParseLaTeX(formula)
	if err != nil || m == nil {
		m = math.New()
		m.Add(math.NewIdentifier(formula))
	}
	f := c.AddFormula(m)
	f.Source = formula
	return f
}

// AddFootnote appends a footnote.
func (c *Container) AddFootnote(styles ...any) *Footnote {
	var para any
	if len(styles) > 0 {
		para = styles[0]
	}
	fn := NewFootnote(para)
	c.add(fn)
	return fn
}

// AddEndnote appends an endnote.
func (c *Container) AddEndnote(styles ...any) *Endnote {
	var para any
	if len(styles) > 0 {
		para = styles[0]
	}
	en := NewEndnote(para)
	c.add(en)
	return en
}

// AddSDT appends a structured document tag.
func (c *Container) AddSDT(typ string) *SDT {
	s := &SDT{SDTType: typ}
	c.add(s)
	return s
}

// AddTOC appends a table of contents field.
func (c *Container) AddTOC(font any, tocStyle any, minDepth, maxDepth int) *TOC {
	toc := &TOC{FontStyle: font, TOCStyle: tocStyle, MinDepth: minDepth, MaxDepth: maxDepth}
	if toc.MinDepth == 0 {
		toc.MinDepth = 1
	}
	if toc.MaxDepth == 0 {
		toc.MaxDepth = 9
	}
	c.add(toc)
	return toc
}

// AddChart appends a simple chart.
func (c *Container) AddChart(typ string, categories []string, values []float64, s ...style.Chart) *Chart {
	ch := &Chart{ChartType: typ, Categories: categories, Values: values}
	ch.AddSeries(categories, values)
	if len(s) > 0 {
		ch.Style = s[0]
	}
	c.add(ch)
	return ch
}

// AddRuby appends phonetic guide text.
func (c *Container) AddRuby(base, ruby *TextRun, props RubyProperties) *Ruby {
	r := &Ruby{BaseText: base, RubyText: ruby, Properties: props}
	c.add(r)
	return r
}

// AddFormField appends a form field of the given type.
func (c *Container) AddFormField(typ string, styles ...any) *FormField {
	font, para := pick2(styles)
	ff := NewFormField(typ, font, para)
	c.add(ff)
	return ff
}

// AddComment appends a comment range container.
func (c *Container) AddComment(author, initials, date string) *Comment {
	cm := &Comment{Author: author, Initials: initials, Date: date}
	cm.Kind = "Comment"
	c.add(cm)
	return cm
}

// CommentOn adds visible text annotated with a comment (author, body).
// extras is optional initials then date (ISO-8601).
func (c *Container) CommentOn(text, body, author string, extras ...string) *Text {
	initials, date := "", ""
	if len(extras) > 0 {
		initials = extras[0]
	}
	if len(extras) > 1 {
		date = extras[1]
	}
	cm := &Comment{Author: author, Initials: initials, Date: date}
	cm.Kind = "Comment"
	if body != "" {
		cm.AddText(body)
	}
	tx := c.AddText(text)
	tx.SetCommentRangeStart(cm)
	tx.SetCommentRangeEnd(cm)
	return tx
}

// AddInsertion appends text marked as a tracked insertion (w:ins).
func (c *Container) AddInsertion(text, author, date string, styles ...any) *Text {
	tx := c.AddText(text, styles...)
	tx.SetChangeInfo("ins", author, date)
	return tx
}

// AddDeletion appends text marked as a tracked deletion (w:del).
func (c *Container) AddDeletion(text, author, date string, styles ...any) *Text {
	tx := c.AddText(text, styles...)
	tx.SetChangeInfo("del", author, date)
	return tx
}

// AddObject is the deprecated PHPWord alias for AddOLEObject.
func (c *Container) AddObject(source string, styles ...any) *OLEObject {
	return c.AddOLEObject(source, styles...)
}

// AddOLEObject appends an embedded OLE object.
func (c *Container) AddOLEObject(source string, styles ...any) *OLEObject {
	var s any
	if len(styles) > 0 {
		s = styles[0]
	}
	o := NewOLEObject(source, s)
	c.add(o)
	return o
}
