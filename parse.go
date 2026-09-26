package word

import (
	"path/filepath"
	"strings"

	"github.com/yunkeweb/go-word/element"
	"github.com/yunkeweb/go-word/metadata"
	"github.com/yunkeweb/go-word/style"
)

// Vertical cell alignment (ST_VerticalJc), re-exported for document APIs.
const (
	VAlignTop    = style.VAlignTop
	VAlignCenter = style.VAlignCenter
	VAlignBottom = style.VAlignBottom
)

// Text direction (ST_TextDirection).
const (
	TextDirectionHorizontal = style.TextDirectionLrTb
	TextDirectionVertical   = style.TextDirectionTbRl
)

// Paragraph is a rich-text paragraph used by bookmark and hyperlink APIs.
type Paragraph = element.TextRun

// ImageFile is a media part extracted from word/media/.
type ImageFile struct {
	Name string
	MIME string
	Data []byte
}

// ExtractText returns document body text in paragraph and table order.
func (d *Document) ExtractText() string {
	if d == nil {
		return ""
	}
	var b strings.Builder
	for _, sec := range d.sections {
		extractContainer(&b, sec.Elements())
	}
	return strings.TrimRight(b.String(), "\n")
}

func extractContainer(b *strings.Builder, els []element.Element) {
	for _, el := range els {
		switch v := el.(type) {
		case *element.Table:
			extractTable(b, v)
		case *element.Text:
			writeExtractLine(b, v.Content)
		case *element.TextRun:
			writeExtractLine(b, v.GetText())
		case *element.Title:
			writeExtractLine(b, v.Text)
		case *element.Link:
			writeExtractLine(b, v.Text)
		case *element.ListItem:
			writeExtractLine(b, v.Text)
		case *element.ListItemRun:
			writeExtractLine(b, v.GetText())
		case *element.PreserveText:
			writeExtractLine(b, v.Content)
		case *element.CheckBox:
			writeExtractLine(b, v.Content)
		}
	}
}

func extractTable(b *strings.Builder, tbl *element.Table) {
	if tbl == nil {
		return
	}
	for _, row := range tbl.Rows {
		cells := make([]string, 0, len(row.Cells))
		for _, cell := range row.Cells {
			cells = append(cells, extractCellText(cell))
		}
		writeExtractLine(b, strings.Join(cells, "\t"))
	}
}

func extractCellText(c *element.Cell) string {
	if c == nil {
		return ""
	}
	var parts []string
	for _, el := range c.Elements() {
		switch v := el.(type) {
		case *element.Table:
			var nb strings.Builder
			extractTable(&nb, v)
			if s := strings.TrimRight(nb.String(), "\n"); s != "" {
				parts = append(parts, s)
			}
		case *element.Text:
			if v.Content != "" {
				parts = append(parts, v.Content)
			}
		case *element.TextRun:
			if s := v.GetText(); s != "" {
				parts = append(parts, s)
			}
		case *element.Title:
			if v.Text != "" {
				parts = append(parts, v.Text)
			}
		case *element.Link:
			if v.Text != "" {
				parts = append(parts, v.Text)
			}
		case *element.ListItem:
			if v.Text != "" {
				parts = append(parts, v.Text)
			}
		case *element.PreserveText:
			if v.Content != "" {
				parts = append(parts, v.Content)
			}
		case *element.CheckBox:
			if v.Content != "" {
				parts = append(parts, v.Content)
			}
		}
	}
	return strings.Join(parts, " ")
}

func writeExtractLine(b *strings.Builder, s string) {
	b.WriteString(s)
	b.WriteByte('\n')
}

// ExtractImages returns pictures from word/media/ (after Open/Read) or from
// in-memory image elements on a newly built document.
func (d *Document) ExtractImages() ([]ImageFile, error) {
	if d == nil {
		return nil, nil
	}
	if d.imagesLoaded {
		out := make([]ImageFile, len(d.extractedImages))
		copy(out, d.extractedImages)
		return out, nil
	}
	var out []ImageFile
	walkDocument(d, func(el element.Element) {
		img, ok := el.(*element.Image)
		if !ok || len(img.Data) == 0 {
			return
		}
		name := img.GetName()
		if name == "" {
			name = img.Source
		}
		base := filepath.Base(name)
		ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(base), "."))
		out = append(out, ImageFile{
			Name: base,
			MIME: mimeForExt(ext),
			Data: append([]byte(nil), img.Data...),
		})
	})
	return out, nil
}

// GetMetadata returns core document properties (author, created, modified, ...).
func (d *Document) GetMetadata() *metadata.DocInfo {
	if d == nil {
		return nil
	}
	return d.info
}

// AddBookmark writes a bookmarkStart/bookmarkEnd pair on paragraph p.
// If p is nil, the bookmark is appended to the last section.
func (d *Document) AddBookmark(p *Paragraph, name string) *element.Bookmark {
	if name == "" {
		return nil
	}
	if p != nil {
		return p.AddBookmark(name)
	}
	return d.lastOrNewSection().AddBookmark(name)
}

// AddHyperlinkToBookmark appends an internal hyperlink (w:hyperlink w:anchor)
// that jumps to bookmarkName. If p is nil, the link is added to the last section.
func (d *Document) AddHyperlinkToBookmark(p *Paragraph, text, bookmarkName string) *element.Link {
	if bookmarkName == "" {
		return nil
	}
	if p != nil {
		return p.AddLink(bookmarkName, text, nil, nil, true)
	}
	return d.lastOrNewSection().AddLink(bookmarkName, text, nil, nil, true)
}
