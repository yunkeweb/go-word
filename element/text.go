package element

import "github.com/yunkeweb/go-word/style"

// Text is a paragraph containing a single run (PHPWord Element\Text).
type Text struct {
	Base
	Content        string
	FontStyle      any // string name or style.Font
	ParagraphStyle any // string name or style.Paragraph
}

func (t *Text) Type() string { return "Text" }

// NewText constructs a text element.
func NewText(text string, font, para any) *Text {
	return &Text{
		Content:        text,
		FontStyle:      normalizeFont(font),
		ParagraphStyle: normalizePara(para),
	}
}

func (t *Text) Text() string { return t.Content }

func (t *Text) SetText(s string) { t.Content = s }

func (t *Text) Font() *style.Font {
	if f, ok := t.FontStyle.(style.Font); ok {
		return &f
	}
	if f, ok := t.FontStyle.(*style.Font); ok {
		return f
	}
	return nil
}

func (t *Text) FontName() string {
	if s, ok := t.FontStyle.(string); ok {
		return s
	}
	return ""
}

func (t *Text) Paragraph() *style.Paragraph {
	if p, ok := t.ParagraphStyle.(style.Paragraph); ok {
		return &p
	}
	if p, ok := t.ParagraphStyle.(*style.Paragraph); ok {
		return p
	}
	return nil
}

func (t *Text) ParagraphName() string {
	if s, ok := t.ParagraphStyle.(string); ok {
		return s
	}
	return ""
}

func normalizeFont(v any) any {
	switch x := v.(type) {
	case nil:
		return nil
	case string:
		if x == "" {
			return nil
		}
		return x
	case style.Font:
		return x
	case *style.Font:
		if x == nil {
			return nil
		}
		return *x
	default:
		return v
	}
}

func normalizePara(v any) any {
	switch x := v.(type) {
	case nil:
		return nil
	case string:
		if x == "" {
			return nil
		}
		return x
	case style.Paragraph:
		return x
	case *style.Paragraph:
		if x == nil {
			return nil
		}
		return *x
	default:
		return v
	}
}
