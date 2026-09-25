package style

// Page orientation.
const (
	OrientationPortrait  = "portrait"
	OrientationLandscape = "landscape"
)

// Default section geometry in twips (A4), matching PHPWord Style\Section.
const (
	DefaultPageWidth     = 11906
	DefaultPageHeight    = 16838
	DefaultGutter        = 0
	DefaultHeaderHeight  = 720
	DefaultFooterHeight  = 720
	DefaultColumnCount   = 1
	DefaultColumnSpacing = 720
	DefaultMargin        = 1440
)

// Section is a section property set (PHPWord Style\Section).
type Section struct {
	Orientation        string
	PageSizeW          int
	PageSizeH          int
	Gutter             int
	HeaderHeight       int
	FooterHeight       int
	MarginTop          int
	MarginLeft         int
	MarginRight        int
	MarginBottom       int
	ColsNum            int
	ColsSpace          int
	BreakType          string // nextPage, continuous, evenPage, oddPage
	PageNumberingStart int
	Borders            Borders
	PaperSize          string
	VAlign             string
	RtlGutter          bool
	LineNumbering      *LineNumbering
}

// LineNumbering is section line numbering (PHPWord Style\LineNumbering).
type LineNumbering struct {
	Start     int
	Increment int
	Distance  int    // twips
	Restart   string // continuous, newPage, newSection
}

const (
	LineNumberingContinuous = "continuous"
	LineNumberingNewPage    = "newPage"
	LineNumberingNewSection = "newSection"
)

// NewSection returns A4 portrait with 1-inch margins.
func NewSection() Section {
	return Section{
		Orientation:  OrientationPortrait,
		PageSizeW:    DefaultPageWidth,
		PageSizeH:    DefaultPageHeight,
		HeaderHeight: DefaultHeaderHeight,
		FooterHeight: DefaultFooterHeight,
		MarginTop:    DefaultMargin,
		MarginLeft:   DefaultMargin,
		MarginRight:  DefaultMargin,
		MarginBottom: DefaultMargin,
		ColsNum:      DefaultColumnCount,
		ColsSpace:    DefaultColumnSpacing,
	}
}
