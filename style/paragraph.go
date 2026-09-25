package style

// Indentation is paragraph indentation in twips.
type Indentation struct {
	Left           int
	Right          int
	FirstLine      int
	FirstLineChars int
	Hanging        int
}

// Spacing is paragraph spacing in twips (before/after) and line rule.
type Spacing struct {
	Before int
	After  int
	Line   int
	Rule   string // auto, exact, atLeast
}

// Tab is a custom tab stop.
type Tab struct {
	Val    string // left, center, right, decimal, bar, clear
	Leader string
	Pos    int // twips
}

// Border is a single-side border.
type Border struct {
	Style string // single, dashed, dotted, double, nil, ...
	Size  int    // eighths of a point
	Color string
	Space int
}

// Borders is a four-side plus inside border set.
type Borders struct {
	Top     Border
	Left    Border
	Right   Border
	Bottom  Border
	InsideH Border
	InsideV Border
}

// Shading is a background fill.
type Shading struct {
	Val   string // clear, solid, ...
	Color string
	Fill  string
}

// Paragraph is a paragraph property set (PHPWord Style\Paragraph).
type Paragraph struct {
	Alignment       string
	BasedOn         string
	Next            string
	Indentation     Indentation
	Spacing         Spacing
	WidowControl    *bool
	KeepNext        bool
	KeepLines       bool
	PageBreakBefore bool
	Bidi            bool
	OutlineLevel    int // 0 = unset; 1-9 = heading level
	Tabs            []Tab
	Shading         Shading
	Borders         Borders
	NumStyle             string
	NumLevel             int
	StyleName            string
	ContextualSpacing    bool
	TextAlignment        string
	SuppressAutoHyphens  bool
}

// IsZero reports whether p has no paragraph properties set.
func (p Paragraph) IsZero() bool {
	return p.Alignment == "" && p.BasedOn == "" && p.Next == "" &&
		p.Indentation == (Indentation{}) && p.Spacing == (Spacing{}) &&
		p.WidowControl == nil && !p.KeepNext && !p.KeepLines && !p.PageBreakBefore &&
		!p.Bidi && p.OutlineLevel == 0 && len(p.Tabs) == 0 &&
		p.Shading == (Shading{}) && p.Borders == (Borders{}) &&
		p.NumStyle == "" && p.NumLevel == 0 && p.StyleName == ""
}

// SpaceAfter is a convenience setter used by PHPWord samples.
func (p Paragraph) WithSpaceAfter(twips int) Paragraph {
	p.Spacing.After = twips
	return p
}
