package style

// Line is a VML line style (PHPWord Style\Line).
type Line struct {
	Weight     int
	Color      string
	BeginArrow string
	EndArrow   string
	Dash       string
	Connector  string
	Width      int
	Height     int
	Left       int
	Top        int
	Wrapping   string
}

// Shape is a VML shape style.
type Shape struct {
	Type      string
	Width     int
	Height    int
	Left      int
	Top       int
	Fill      Fill
	Outline   Outline
	Shadow    Shadow
	Extrusion Extrusion
	Wrapping  string
}

// Fill is a shape fill.
type Fill struct {
	Color  string
	Color2 string
	Type   string
}

// Outline is a shape outline.
type Outline struct {
	Color  string
	Weight int
	Line   string
}

// Shadow is a shape shadow.
type Shadow struct {
	Color  string
	Offset string
}

// Extrusion is a 3-D extrusion.
type Extrusion struct {
	Type  string
	Color string
}

// TextBox is a text-frame style.
type TextBox struct {
	Width       int
	Height      int
	Alignment   string
	Wrapping    string
	InnerMargin int
	BorderColor string
	BorderSize  int
	BgColor     string
}

// DataLabelOptions is chart data-label visibility (PHPWord Style\Chart).
type DataLabelOptions struct {
	ShowVal         bool
	ShowCatName     bool
	ShowLegendKey   bool
	ShowSerName     bool
	ShowPercent     bool
	ShowLeaderLines bool
	ShowBubbleSize  bool
	Position        string
}

// DefaultDataLabelOptions matches PHPWord Chart constructor defaults.
func DefaultDataLabelOptions() DataLabelOptions {
	return DataLabelOptions{ShowVal: true, ShowCatName: true}
}

// Chart is a chart style.
type Chart struct {
	Width                 int
	Height                int
	ShowLegend            bool
	LegendPosition        string
	CategoryAxisTitle     string
	ValueAxisTitle        string
	Colors                []string
	ThreeD                bool
	Title                 string
	ShowAxisLabels        bool
	ShowGridX             bool
	ShowGridY             bool
	CategoryLabelPosition string
	ValueLabelPosition    string
	MajorTickPosition     string
	DataLabels            DataLabelOptions
	DataLabelsSet         bool
	LineSmooth            bool
	LineMarker            string
	ValueNumFmt           string
	SecondaryValueNumFmt  string
}

// TOC is a table-of-contents tab style (PHPWord Style\TOC).
type TOC struct {
	TabPos    int
	TabLeader string
	Indent    int
}

// NewTOC returns the PHPWord default TOC tab (right, 9062 twips, dot leader).
func NewTOC() TOC {
	return TOC{TabPos: 9062, TabLeader: "dot", Indent: 200}
}

// Frame is shared positioning for images, text boxes and shapes (PHPWord Style\Frame).
type Frame struct {
	Alignment          string
	Unit               string
	Width              int
	Height             int
	Left               int
	Top                int
	Pos                string
	HPos               string
	VPos               string
	HPosRelTo          string
	VPosRelTo          string
	Wrap               string
	WrapDistanceTop    int
	WrapDistanceBottom int
	WrapDistanceLeft   int
	WrapDistanceRight  int
}

// Language is a theme font language.
type Language struct {
	Latin         string
	EastAsia      string
	Bidirectional string
}
