package word

import "github.com/yunkeweb/go-word/style"

// namedStyle is a document-scoped style definition.
type namedStyle struct {
	Name      string
	Kind      string // font, paragraph, table, numbering, title, link
	Font      *style.Font
	Paragraph *style.Paragraph
	Table     *style.Table
	FirstRow  *style.Table
	Numbering *style.Numbering
	Depth     int
}

func fontPtr(v any) *style.Font {
	switch x := v.(type) {
	case style.Font:
		f := x
		return &f
	case *style.Font:
		return x
	default:
		return nil
	}
}

func paraPtr(v any) *style.Paragraph {
	switch x := v.(type) {
	case style.Paragraph:
		p := x
		return &p
	case *style.Paragraph:
		return x
	default:
		return nil
	}
}

func tablePtr(v any) *style.Table {
	switch x := v.(type) {
	case style.Table:
		t := x
		return &t
	case *style.Table:
		return x
	default:
		return nil
	}
}

func numberingPtr(v any) *style.Numbering {
	switch x := v.(type) {
	case style.Numbering:
		n := x
		return &n
	case *style.Numbering:
		return x
	default:
		return nil
	}
}

// AddFontStyle registers a named character style (PHPWord addFontStyle).
func (d *Document) AddFontStyle(name string, font any, para ...any) {
	ns := namedStyle{Name: name, Kind: "font", Font: fontPtr(font)}
	if len(para) > 0 {
		ns.Paragraph = paraPtr(para[0])
	}
	d.styles = append(d.styles, ns)
}

// AddParagraphStyle registers a named paragraph style.
func (d *Document) AddParagraphStyle(name string, para any) {
	d.styles = append(d.styles, namedStyle{Name: name, Kind: "paragraph", Paragraph: paraPtr(para)})
}

// AddTableStyle registers a named table style.
func (d *Document) AddTableStyle(name string, table any, firstRow ...any) {
	ns := namedStyle{Name: name, Kind: "table", Table: tablePtr(table)}
	if len(firstRow) > 0 {
		ns.FirstRow = tablePtr(firstRow[0])
	}
	d.styles = append(d.styles, ns)
}

// AddNumberingStyle registers a named numbering definition.
func (d *Document) AddNumberingStyle(name string, num any) {
	d.styles = append(d.styles, namedStyle{Name: name, Kind: "numbering", Numbering: numberingPtr(num)})
}

// AddTitleStyle registers HeadingN (depth 1-9). depth 0 is Title.
func (d *Document) AddTitleStyle(depth int, font any, para ...any) {
	name := "Title"
	if depth > 0 {
		name = headingStyleName(depth)
	}
	ns := namedStyle{Name: name, Kind: "title", Font: fontPtr(font), Depth: depth}
	if len(para) > 0 {
		ns.Paragraph = paraPtr(para[0])
	}
	d.styles = append(d.styles, ns)
}

// AddLinkStyle registers the Hyperlink character style.
func (d *Document) AddLinkStyle(name string, font any) {
	if name == "" {
		name = "Hyperlink"
	}
	d.styles = append(d.styles, namedStyle{Name: name, Kind: "link", Font: fontPtr(font)})
}

// SetDefaultParagraphStyle sets the Normal style.
func (d *Document) SetDefaultParagraphStyle(para any) {
	d.defaultParagraph = paraPtr(para)
}

func headingStyleName(depth int) string {
	switch depth {
	case 1:
		return "Heading1"
	case 2:
		return "Heading2"
	case 3:
		return "Heading3"
	case 4:
		return "Heading4"
	case 5:
		return "Heading5"
	case 6:
		return "Heading6"
	case 7:
		return "Heading7"
	case 8:
		return "Heading8"
	default:
		return "Heading9"
	}
}

// NamedStyle is a document-scoped style definition (PHPWord Style registry entry).
type NamedStyle struct {
	Name      string
	Kind      string
	Font      *style.Font
	Paragraph *style.Paragraph
	Table     *style.Table
	FirstRow  *style.Table
	Numbering *style.Numbering
	Depth     int
}

func (n namedStyle) export() NamedStyle {
	return NamedStyle{
		Name: n.Name, Kind: n.Kind, Font: n.Font, Paragraph: n.Paragraph,
		Table: n.Table, FirstRow: n.FirstRow, Numbering: n.Numbering, Depth: n.Depth,
	}
}

// GetStyles returns registered named styles (PHPWord Style::getStyles).
func (d *Document) GetStyles() []NamedStyle {
	out := make([]NamedStyle, len(d.styles))
	for i, s := range d.styles {
		out[i] = s.export()
	}
	return out
}

// GetStyle returns a named style, or nil.
func (d *Document) GetStyle(name string) *NamedStyle {
	if s := d.styleByName(name); s != nil {
		ns := s.export()
		return &ns
	}
	return nil
}

// CountStyles returns the number of registered named styles.
func (d *Document) CountStyles() int { return len(d.styles) }

// ResetStyles clears the named-style registry (PHPWord Style::resetStyles).
func (d *Document) ResetStyles() { d.styles = nil }

func (d *Document) styleByName(name string) *namedStyle {
	for i := range d.styles {
		if d.styles[i].Name == name {
			return &d.styles[i]
		}
	}
	return nil
}
