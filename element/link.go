package element

// Link is a hyperlink (PHPWord Element\Link).
type Link struct {
	Base
	Target         string
	Text           string
	Internal       bool
	FontStyle      any
	ParagraphStyle any
}

func (l *Link) Type() string { return "Link" }

// NewLink constructs a hyperlink.
func NewLink(target, text string, font, para any, internal bool) *Link {
	if text == "" {
		text = target
	}
	return &Link{
		Target:         target,
		Text:           text,
		Internal:       internal,
		FontStyle:      normalizeFont(font),
		ParagraphStyle: normalizePara(para),
	}
}
