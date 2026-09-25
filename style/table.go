package style

// Table is a table property set (PHPWord Style\Table).
type Table struct {
	Width            int    // twips
	Unit             string // dxa, pct, auto, nil
	Alignment        string
	CellMarginTop    int
	CellMarginLeft   int
	CellMarginRight  int
	CellMarginBottom int
	CellSpacing      int
	Layout           string // autofit, fixed
	BidiVisual       bool
	Borders          Borders
	Shading          Shading
	CellSpacingVal   int
	StyleName        string
	FirstRow         *Table
	Position         *TablePosition
	Indent           int
}

// TablePosition is floating-table placement (PHPWord Style\TablePosition).
type TablePosition struct {
	LeftFromText   int
	RightFromText  int
	TopFromText    int
	BottomFromText int
	VertAnchor     string
	HorzAnchor     string
	TblpXSpec      string
	TblpX          int
	TblpYSpec      string
	TblpY          int
}

const (
	VAnchorText   = "text"
	VAnchorMargin = "margin"
	VAnchorPage   = "page"
	HAnchorText   = "text"
	HAnchorMargin = "margin"
	HAnchorPage   = "page"
)

// Cell is a table cell property set.
type Cell struct {
	Width        int
	Unit         string
	VAlign       string
	TextDir      string
	GridSpan     int
	VMerge       string // restart, continue
	Shading      Shading
	Borders      Borders
	BgColor      string
	NoWrap       bool
	PaddingTop   int
	PaddingLeft  int
	PaddingRight int
	PaddingBottom int
}

// Row is a table row property set.
type Row struct {
	Height    int
	Rule      string // auto, exact, atLeast
	Header    bool
	CantSplit bool
	TblHeader bool
}
