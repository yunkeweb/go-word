package element

import (
	"path/filepath"
	"strings"

	"github.com/yunkeweb/go-word/metadata"
	"github.com/yunkeweb/go-word/style"
)

// Header types (PHPWord Element\Header).
const (
	HeaderAuto  = "default"
	HeaderFirst = "first"
	HeaderEven  = "even"
)

// Section is a document section (PHPWord Element\Section).
type Section struct {
	Container
	Style              style.Section
	Headers            []*Header
	Footers            []*Footer
	FootnoteProperties *metadata.FootnoteProperties
}

func (s *Section) Type() string { return "Section" }

// NewSection constructs a section with the given 1-based index.
func NewSection(index int, st any) *Section {
	sec := &Section{Style: style.NewSection()}
	sec.Kind = "Section"
	sec.SectionID = index
	sec.DocPart = "Section"
	applySectionStyle(&sec.Style, st)
	return sec
}

func applySectionStyle(dst *style.Section, st any) {
	switch v := st.(type) {
	case style.Section:
		*dst = v
	case *style.Section:
		if v != nil {
			*dst = *v
		}
	}
}

// AddHeader appends a header of the given type (default, first, even).
func (s *Section) AddHeader(typ ...string) *Header {
	t := HeaderAuto
	if len(typ) > 0 && typ[0] != "" {
		t = typ[0]
	}
	h := &Header{HeaderType: t}
	h.Kind = "Header"
	h.SectionID = s.SectionID
	s.Headers = append(s.Headers, h)
	setParent(h, s)
	return h
}

// AddFooter appends a footer of the given type.
func (s *Section) AddFooter(typ ...string) *Footer {
	t := HeaderAuto
	if len(typ) > 0 && typ[0] != "" {
		t = typ[0]
	}
	f := &Footer{HeaderType: t}
	f.Kind = "Footer"
	f.SectionID = s.SectionID
	s.Footers = append(s.Footers, f)
	setParent(f, s)
	return f
}

// Header is a header part.
type Header struct {
	Container
	HeaderType string
}

func (h *Header) Type() string { return "Header" }

func (h *Header) SetType(t string) { h.HeaderType = headerTypeOrDefault(t) }
func (h *Header) GetType() string  { return h.HeaderType }
func (h *Header) ResetType() string {
	h.HeaderType = HeaderAuto
	return h.HeaderType
}
func (h *Header) FirstPage() string {
	h.HeaderType = HeaderFirst
	return h.HeaderType
}
func (h *Header) EvenPage() string {
	h.HeaderType = HeaderEven
	return h.HeaderType
}

// Footer is a footer part.
type Footer struct {
	Container
	HeaderType string
}

func (f *Footer) Type() string { return "Footer" }

func (f *Footer) SetType(t string) { f.HeaderType = headerTypeOrDefault(t) }
func (f *Footer) GetType() string  { return f.HeaderType }
func (f *Footer) ResetType() string {
	f.HeaderType = HeaderAuto
	return f.HeaderType
}
func (f *Footer) FirstPage() string {
	f.HeaderType = HeaderFirst
	return f.HeaderType
}
func (f *Footer) EvenPage() string {
	f.HeaderType = HeaderEven
	return f.HeaderType
}

func headerTypeOrDefault(t string) string {
	switch t {
	case HeaderFirst, HeaderEven, HeaderAuto:
		return t
	default:
		return HeaderAuto
	}
}

// HasDifferentFirstPage reports whether a first-page header or footer exists.
func (s *Section) HasDifferentFirstPage() bool {
	for _, h := range s.Headers {
		if h.HeaderType == HeaderFirst {
			return true
		}
	}
	for _, f := range s.Footers {
		if f.HeaderType == HeaderFirst {
			return true
		}
	}
	return false
}

// SetFootnoteProperties sets section footnote properties.
func (s *Section) SetFootnoteProperties(p *metadata.FootnoteProperties) {
	s.FootnoteProperties = p
}

// GetFootnoteProperties returns section footnote properties.
func (s *Section) GetFootnoteProperties() *metadata.FootnoteProperties {
	return s.FootnoteProperties
}

// GetHeaders returns section headers.
func (s *Section) GetHeaders() []*Header { return s.Headers }

// GetFooters returns section footers.
func (s *Section) GetFooters() []*Footer { return s.Footers }

// AddWatermark appends a watermark image to a header (PHPWord Header::addWatermark).
func (h *Header) AddWatermark(src string, st ...any) *Image {
	img := h.AddImage(src, st...)
	img.IsWatermark = true
	img.Style.IsWatermark = true
	img.Style.WrappingStyle = style.WrappingBehind
	return img
}

// AddWatermarkBytes appends an in-memory image watermark to a header.
func (h *Header) AddWatermarkBytes(name string, data []byte, st ...any) *Image {
	img := h.AddImageBytes(name, data, st...)
	img.IsWatermark = true
	img.Style.IsWatermark = true
	img.Style.WrappingStyle = style.WrappingBehind
	return img
}

// AddTextWatermark appends a VML text watermark to a header.
func (h *Header) AddTextWatermark(text string) *TextWatermark {
	tw := &TextWatermark{Text: text}
	h.add(tw)
	return tw
}

// EnsureTextWatermark sets or updates the header's text watermark.
func (h *Header) EnsureTextWatermark(text string) *TextWatermark {
	for _, el := range h.Elements() {
		if tw, ok := el.(*TextWatermark); ok {
			tw.Text = text
			return tw
		}
	}
	return h.AddTextWatermark(text)
}

// EnsureImageWatermark sets or updates the header's image watermark.
func (h *Header) EnsureImageWatermark(data []byte, name ...string) *Image {
	fileName := "watermark.png"
	if len(name) > 0 && name[0] != "" {
		fileName = name[0]
	}
	for _, el := range h.Elements() {
		if img, ok := el.(*Image); ok && img.IsWatermark {
			img.Data = append([]byte(nil), data...)
			img.Media.Data = img.Data
			ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(fileName), "."))
			if ext == "jpg" {
				ext = "jpeg"
			}
			if ext != "" {
				img.Media.Ext = ext
			}
			return img
		}
	}
	return h.AddWatermarkBytes(fileName, data, style.Image{Width: 400, Height: 400})
}

// HasDifferentEvenPage reports whether an even-page header or footer exists.
func (s *Section) HasDifferentEvenPage() bool {
	for _, h := range s.Headers {
		if h.HeaderType == HeaderEven {
			return true
		}
	}
	for _, f := range s.Footers {
		if f.HeaderType == HeaderEven {
			return true
		}
	}
	return false
}

// SetColumns sets section multi-column layout (w:cols num/space/sep).
func (s *Section) SetColumns(num int, space int, showLine bool) {
	if num < 1 {
		num = 1
	}
	s.Style.ColsNum = num
	if space > 0 {
		s.Style.ColsSpace = space
	}
	s.Style.ColsSeparator = showLine
}

// SetOrientation sets portrait or landscape page orientation.
func (s *Section) SetOrientation(orient string) {
	s.Style.Orientation = orient
}

// SetDifferentFirstPage adds a first-page header when enable is true.
func (s *Section) SetDifferentFirstPage(enable bool) {
	if enable && !s.HasDifferentFirstPage() {
		s.AddHeader(HeaderFirst)
	}
}
