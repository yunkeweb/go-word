package style

import "github.com/yunkeweb/go-word/pkg/common"

// Paper sizes from PHPWord Style\Paper (ISO/IEC 29500 subset actually used).
var paperSizes = map[string]struct {
	W, H float64
	Unit string
}{
	"A3":     {297, 420, "mm"},
	"A4":     {210, 297, "mm"},
	"A5":     {148, 210, "mm"},
	"B5":     {176, 250, "mm"},
	"Folio":  {8.5, 13, "in"},
	"Legal":  {8.5, 14, "in"},
	"Letter": {8.5, 11, "in"},
}

// Paper is a named page size (PHPWord Style\Paper).
type Paper struct {
	Size   string
	Width  int // twips
	Height int
}

// NewPaper returns a paper of the given size (default A4).
func NewPaper(size string) Paper {
	p := Paper{}
	p.SetSize(size)
	return p
}

// SetSize updates width/height from a known paper name.
func (p *Paper) SetSize(size string) {
	if size == "" {
		size = "A4"
	}
	spec, ok := paperSizes[size]
	if !ok {
		spec = paperSizes["A4"]
		size = "A4"
	}
	p.Size = size
	if spec.Unit == "mm" {
		p.Width = int(common.CMToTwip(spec.W / 10))
		p.Height = int(common.CMToTwip(spec.H / 10))
		return
	}
	p.Width = int(common.InchToTwip(spec.W))
	p.Height = int(common.InchToTwip(spec.H))
}

// ApplyToSection writes this paper's geometry onto a section style.
func (p Paper) ApplyToSection(s *Section) {
	if s == nil {
		return
	}
	s.PaperSize = p.Size
	s.PageSizeW = p.Width
	s.PageSizeH = p.Height
}
