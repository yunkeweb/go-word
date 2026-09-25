package metadata

// FootnoteProperties is section footnote configuration
// (PHPWord ComplexType\FootnoteProperties).
type FootnoteProperties struct {
	Pos        string // pageBottom, beneathText, sectEnd, docEnd
	NumFmt     string
	NumStart   int
	NumRestart string // continuous, eachSect, eachPage
}

const (
	FootnotePosPageBottom   = "pageBottom"
	FootnotePosBeneathText  = "beneathText"
	FootnotePosSectionEnd   = "sectEnd"
	FootnotePosDocEnd       = "docEnd"
	FootnoteRestartContinuous = "continuous"
	FootnoteRestartEachSect   = "eachSect"
	FootnoteRestartEachPage   = "eachPage"
)

// TblWidth is a typed table/cell width (PHPWord ComplexType\TblWidth).
type TblWidth struct {
	Type  string // nil, auto, pct, dxa
	Value int
}
