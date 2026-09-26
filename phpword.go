package word

import (
	"io"

	"github.com/yunkeweb/go-word/element"
	"github.com/yunkeweb/go-word/metadata"
	"github.com/yunkeweb/go-word/pkg/common"
	"github.com/yunkeweb/go-word/style"
)

// Document is the PHPWord PhpWord class: an in-memory word-processing document.
type Document struct {
	sections         []*element.Section
	styles           []namedStyle
	info             *metadata.DocInfo
	settings         *metadata.Settings
	compatibility    *metadata.Compatibility
	defaultFontName  string
	defaultAsianFont string
	defaultFontSize  float64
	defaultFontColor string
	defaultParagraph *style.Paragraph
	nextBookmarkID   int
	footnotes        []*element.Footnote
	endnotes         []*element.Endnote
	comments         []*element.Comment
	titles           []*element.Title
	charts           []*element.Chart
	textWatermark    string
	imageWatermark   []byte
}

// PhpWord is an alias for Document, matching the PHP class name.
type PhpWord = Document

// New creates an empty document (PHPWord constructor).
func New() *Document {
	return &Document{
		info:             metadata.NewDocInfo(),
		settings:         metadata.NewSettings(),
		compatibility:    metadata.NewCompatibility(),
		defaultFontName:  defaultFontName,
		defaultAsianFont: defaultAsianFontName,
		defaultFontSize:  defaultFontSize,
		defaultFontColor: defaultFontColor,
	}
}

// DocInfo returns document properties.
func (d *Document) DocInfo() *metadata.DocInfo { return d.info }

// GetDocInfo is the PHPWord name for DocInfo.
func (d *Document) GetDocInfo() *metadata.DocInfo { return d.info }

// Settings returns document settings.
func (d *Document) Settings() *metadata.Settings { return d.settings }

// GetSettings is the PHPWord name for Settings.
func (d *Document) GetSettings() *metadata.Settings { return d.settings }

// Compatibility returns OOXML compatibility settings.
func (d *Document) Compatibility() *metadata.Compatibility { return d.compatibility }

// GetCompatibility is the PHPWord name for Compatibility.
func (d *Document) GetCompatibility() *metadata.Compatibility { return d.compatibility }

// Sections returns all sections.
func (d *Document) Sections() []*element.Section { return d.sections }

// GetSections is the PHPWord name for Sections.
func (d *Document) GetSections() []*element.Section { return d.sections }

// GetTitles returns heading elements in document order.
func (d *Document) GetTitles() []*element.Title {
	var out []*element.Title
	walkDocument(d, func(el element.Element) {
		if t, ok := el.(*element.Title); ok {
			out = append(out, t)
		}
	})
	return out
}

// GetFootnotes returns footnote elements.
func (d *Document) GetFootnotes() []*element.Footnote {
	var out []*element.Footnote
	walkDocument(d, func(el element.Element) {
		if t, ok := el.(*element.Footnote); ok {
			out = append(out, t)
		}
	})
	return out
}

// GetEndnotes returns endnote elements.
func (d *Document) GetEndnotes() []*element.Endnote {
	var out []*element.Endnote
	walkDocument(d, func(el element.Element) {
		if t, ok := el.(*element.Endnote); ok {
			out = append(out, t)
		}
	})
	return out
}

// GetCharts returns chart elements.
func (d *Document) GetCharts() []*element.Chart {
	var out []*element.Chart
	walkDocument(d, func(el element.Element) {
		if t, ok := el.(*element.Chart); ok {
			out = append(out, t)
		}
	})
	return out
}

// GetComments returns comment elements, including those attached as ranges.
func (d *Document) GetComments() []*element.Comment {
	seen := map[*element.Comment]bool{}
	var out []*element.Comment
	add := func(c *element.Comment) {
		if c == nil || seen[c] {
			return
		}
		seen[c] = true
		out = append(out, c)
	}
	walkDocument(d, func(el element.Element) {
		if t, ok := el.(*element.Comment); ok {
			add(t)
		}
		if g, ok := el.(interface{ GetCommentRangeStart() *element.Comment }); ok {
			add(g.GetCommentRangeStart())
		}
		if g, ok := el.(interface{ GetCommentRangeEnd() *element.Comment }); ok {
			add(g.GetCommentRangeEnd())
		}
	})
	return out
}

// GetBookmarks returns bookmark elements.
func (d *Document) GetBookmarks() []*element.Bookmark {
	var out []*element.Bookmark
	walkDocument(d, func(el element.Element) {
		if t, ok := el.(*element.Bookmark); ok {
			out = append(out, t)
		}
	})
	return out
}

// GetSection returns the section at index, or nil.
func (d *Document) GetSection(index int) *element.Section {
	if index < 0 || index >= len(d.sections) {
		return nil
	}
	return d.sections[index]
}

// AddSection appends a section (PHPWord addSection).
func (d *Document) AddSection(st ...any) *element.Section {
	var s any
	if len(st) > 0 {
		s = st[0]
	}
	sec := element.NewSection(len(d.sections)+1, s)
	d.sections = append(d.sections, sec)
	return sec
}

func (d *Document) lastOrNewSection() *element.Section {
	if len(d.sections) == 0 {
		return d.AddSection()
	}
	return d.sections[len(d.sections)-1]
}

// SortSections sorts sections with the given comparison.
func (d *Document) SortSections(less func(a, b *element.Section) bool) {
	secs := d.sections
	for i := 1; i < len(secs); i++ {
		for j := i; j > 0 && less(secs[j], secs[j-1]); j-- {
			secs[j], secs[j-1] = secs[j-1], secs[j]
		}
	}
}

func (d *Document) GetDefaultFontName() string      { return d.defaultFontName }
func (d *Document) SetDefaultFontName(name string)  { d.defaultFontName = name }
func (d *Document) GetDefaultAsianFontName() string { return d.defaultAsianFont }
func (d *Document) SetDefaultAsianFontName(name string) {
	d.defaultAsianFont = name
}
func (d *Document) GetDefaultFontSize() float64     { return d.defaultFontSize }
func (d *Document) SetDefaultFontSize(size float64) { d.defaultFontSize = size }
func (d *Document) GetDefaultFontColor() string     { return d.defaultFontColor }
func (d *Document) SetDefaultFontColor(color string) {
	d.defaultFontColor = color
}

// Save writes the document as Word2007 (.docx).
func (d *Document) Save(filename string) error {
	return d.SaveAs(filename, "Word2007")
}

// SaveAs writes the document in the named format.
func (d *Document) SaveAs(filename, format string) error {
	w, err := CreateWriter(d, format)
	if err != nil {
		return err
	}
	return w.Save(filename)
}

// WriteTo writes a Word2007 package to dest.
func (d *Document) WriteTo(dest io.Writer) (int64, error) {
	w, err := CreateWriter(d, "Word2007")
	if err != nil {
		return 0, err
	}
	return w.WriteTo(dest)
}

// Bytes returns the Word2007 package as a byte slice.
func (d *Document) Bytes() ([]byte, error) {
	buf := common.GetBuffer()
	defer common.PutBuffer(buf)
	if _, err := d.WriteTo(buf); err != nil {
		return nil, err
	}
	return common.CloneBytes(buf.Bytes()), nil
}

// Writer writes a document to a file or stream.
type Writer interface {
	Save(filename string) error
	WriteTo(w io.Writer) (int64, error)
}

// Reader loads a document from a file.
type Reader interface {
	Load(filename string) (*Document, error)
}
