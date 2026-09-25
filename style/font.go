package style

// Underline values (ST_Underline).
const (
	UnderlineNone            = "none"
	UnderlineDash            = "dash"
	UnderlineDashHeavy       = "dashHeavy"
	UnderlineDashLong        = "dashLong"
	UnderlineDashLongHeavy   = "dashLongHeavy"
	UnderlineDouble          = "dbl"
	UnderlineDotDash         = "dotDash"
	UnderlineDotDashHeavy    = "dotDashHeavy"
	UnderlineDotDotDash      = "dotDotDash"
	UnderlineDotDotDashHeavy = "dotDotDashHeavy"
	UnderlineDotted          = "dotted"
	UnderlineDottedHeavy     = "dottedHeavy"
	UnderlineHeavy           = "heavy"
	UnderlineSingle          = "single"
	UnderlineWavy            = "wavy"
	UnderlineWavyDouble      = "wavyDbl"
	UnderlineWavyHeavy       = "wavyHeavy"
	UnderlineWords           = "words"
)

// Highlight (foreground) colors.
const (
	FgYellow      = "yellow"
	FgGreen       = "green"
	FgCyan        = "cyan"
	FgMagenta     = "magenta"
	FgBlue        = "blue"
	FgRed         = "red"
	FgDarkBlue    = "darkBlue"
	FgDarkCyan    = "darkCyan"
	FgDarkGreen   = "darkGreen"
	FgDarkMagenta = "darkMagenta"
	FgDarkRed     = "darkRed"
	FgDarkYellow  = "darkYellow"
	FgDarkGray    = "darkGray"
	FgLightGray   = "lightGray"
	FgBlack       = "black"
)

// Font is a run property set (PHPWord Style\Font).
type Font struct {
	Name                string
	Hint                string // default, eastAsia, cs
	Size                float64
	Color               string
	Bold                bool
	Italic              bool
	Underline           string
	SuperScript         bool
	SubScript           bool
	Strikethrough       bool
	DoubleStrikethrough bool
	SmallCaps           bool
	AllCaps             bool
	Hidden              bool
	FgColor             string
	BgColor             string
	Spacing             int // twips, signed
	Kerning             int
	Scale               int // percent
	RTL                 bool
	Lang                string
	Position            int // half-points, signed
	NoProof             bool
	WhiteSpace          string
	FallbackFont        string
	Shading             Shading
}

// IsZero reports whether f has no run properties set.
func (f Font) IsZero() bool {
	return f.Name == "" && f.Size == 0 && f.Color == "" &&
		!f.Bold && !f.Italic && (f.Underline == "" || f.Underline == UnderlineNone) &&
		!f.SuperScript && !f.SubScript && !f.Strikethrough && !f.DoubleStrikethrough &&
		!f.SmallCaps && !f.AllCaps && !f.Hidden && f.FgColor == "" && f.BgColor == "" &&
		f.Spacing == 0 && f.Kerning == 0 && f.Scale == 0 && !f.RTL && f.Lang == "" &&
		f.Position == 0 && !f.NoProof && f.Hint == ""
}

// HalfPoints returns the font size in OOXML half-points, or 0 if unset.
func (f Font) HalfPoints() int {
	if f.Size == 0 {
		return 0
	}
	return int(f.Size * 2)
}
