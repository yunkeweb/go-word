package element

// TextRun is a paragraph that holds inline elements (PHPWord Element\TextRun).
type TextRun struct {
	Container
	ParagraphStyle any
}

func (t *TextRun) Type() string { return "TextRun" }

// NewTextRun constructs an empty run container.
func NewTextRun(para any) *TextRun {
	tr := &TextRun{ParagraphStyle: normalizePara(para)}
	tr.Kind = "TextRun"
	return tr
}

// AddText appends an inline run (no paragraph of its own).
func (t *TextRun) AddText(text string, styles ...any) *Text {
	font, _ := pick2(styles)
	el := NewText(text, font, nil)
	t.add(el)
	return el
}

// GetText concatenates child text (PHPWord TextRun::getText).
func (t *TextRun) GetText() string {
	var b []byte
	for _, el := range t.elements {
		if tx, ok := el.(*Text); ok {
			b = append(b, tx.Content...)
		}
		if l, ok := el.(*Link); ok {
			b = append(b, l.Text...)
		}
	}
	return string(b)
}

// SetParagraphStyle sets the paragraph style of this run.
func (t *TextRun) SetParagraphStyle(para any) { t.ParagraphStyle = normalizePara(para) }

// GetParagraphStyle returns the paragraph style.
func (t *TextRun) GetParagraphStyle() any { return t.ParagraphStyle }
