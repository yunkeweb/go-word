package word

import (
	"bytes"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"path"
	"strconv"
	"strings"

	"github.com/yunkeweb/go-word/element"
	"github.com/yunkeweb/go-word/ooxml"
	"github.com/yunkeweb/go-word/pkg/common"
)

type word2007Writer struct {
	doc          *Document
	rels         []ooxml.Relationship
	nextRel      int
	images       []pkgImage
	headers      []pkgHF
	footers      []pkgHF
	hyper        []ooxml.Relationship
	charts       []pkgChart
	comments     []*element.Comment
	oles         []pkgOLE
	imgIndex     int
	hfIndex      int
	bkIndex      int
	chartIndex   int
	commentIndex int
	oleIndex     int
}

type pkgChart struct {
	RelID string
	Name  string
	El    *element.Chart
}

type pkgOLE struct {
	RelID string
	Name  string
	Data  []byte
	El    *element.OLEObject
}

type pkgImage struct {
	RelID string
	Name  string
	Data  []byte
	Ext   string
	El    *element.Image
}

type pkgHF struct {
	RelID string
	Name  string
	Type  string // default, first, even
	Kind  string // header, footer
	El    element.Element
}

func newWord2007Writer(doc *Document) *word2007Writer {
	return &word2007Writer{doc: doc, nextRel: 1}
}

func (w *word2007Writer) Save(filename string) error {
	var buf bytes.Buffer
	if _, err := w.WriteTo(&buf); err != nil {
		return err
	}
	return writeFile(filename, buf.Bytes())
}

func (w *word2007Writer) WriteTo(dest io.Writer) (int64, error) {
	if err := w.prepare(); err != nil {
		return 0, err
	}
	var buf bytes.Buffer
	zw := common.NewZipWriter(&buf)

	parts := []struct {
		name string
		data []byte
	}{
		{"[Content_Types].xml", w.contentTypes()},
		{"_rels/.rels", w.pkgRels()},
		{"docProps/core.xml", w.coreProps()},
		{"docProps/app.xml", w.appProps()},
		{"word/document.xml", w.documentXML()},
		{"word/_rels/document.xml.rels", w.docRels()},
		{"word/styles.xml", w.stylesXML()},
		{"word/numbering.xml", w.numberingXML()},
		{"word/settings.xml", w.settingsXML()},
		{"word/webSettings.xml", []byte(webSettingsXML)},
		{"word/fontTable.xml", []byte(fontTableXML)},
		{"word/theme/theme1.xml", []byte(themeXML)},
	}
	if len(w.doc.info.Custom) > 0 {
		parts = append(parts, struct {
			name string
			data []byte
		}{"docProps/custom.xml", w.customProps()})
	}
	for _, p := range parts {
		if err := zw.AddFile(p.name, p.data); err != nil {
			return 0, err
		}
	}
	for _, h := range w.headers {
		data := w.hdrFtrXML("w:hdr", h.El)
		if err := zw.AddFile(h.Name, data); err != nil {
			return 0, err
		}
	}
	for _, f := range w.footers {
		data := w.hdrFtrXML("w:ftr", f.El)
		if err := zw.AddFile(f.Name, data); err != nil {
			return 0, err
		}
	}
	if len(w.doc.footnotes) > 0 {
		if err := zw.AddFile("word/footnotes.xml", w.notesXML(true)); err != nil {
			return 0, err
		}
	}
	if len(w.doc.endnotes) > 0 {
		if err := zw.AddFile("word/endnotes.xml", w.notesXML(false)); err != nil {
			return 0, err
		}
	}
	if len(w.comments) > 0 {
		if err := zw.AddFile("word/comments.xml", w.commentsXML()); err != nil {
			return 0, err
		}
	}
	for _, ch := range w.charts {
		if err := zw.AddFile(ch.Name, chartPartXML(ch.El)); err != nil {
			return 0, err
		}
	}
	for _, o := range w.oles {
		if err := zw.AddFile(o.Name, o.Data); err != nil {
			return 0, err
		}
	}
	for _, img := range w.images {
		if err := zw.AddFile(img.Name, img.Data); err != nil {
			return 0, err
		}
	}
	if err := zw.Close(); err != nil {
		return 0, err
	}
	n, err := dest.Write(buf.Bytes())
	return int64(n), err
}

func (w *word2007Writer) addRel(typ, target, mode string) string {
	id := "rId" + strconv.Itoa(w.nextRel)
	w.nextRel++
	w.rels = append(w.rels, ooxml.Relationship{ID: id, Type: typ, Target: target, TargetMode: mode})
	return id
}

func (w *word2007Writer) prepare() error {
	w.rels = nil
	w.images = nil
	w.headers = nil
	w.footers = nil
	w.charts = nil
	w.comments = nil
	w.oles = nil
	w.nextRel = 1
	w.imgIndex = 0
	w.hfIndex = 0
	w.bkIndex = 0
	w.chartIndex = 0
	w.commentIndex = 0
	w.oleIndex = 0
	w.doc.titles = nil

	w.addRel(ooxml.NSOfficeRelStyles, "styles.xml", "")
	w.addRel(ooxml.NSOfficeRelNumbering, "numbering.xml", "")
	w.addRel(ooxml.NSOfficeRelSettings, "settings.xml", "")
	w.addRel(ooxml.NSOfficeRelWebSettings, "webSettings.xml", "")
	w.addRel(ooxml.NSOfficeRelFontTable, "fontTable.xml", "")
	w.addRel(ooxml.NSOfficeRelTheme, "theme/theme1.xml", "")

	var notes []*element.Footnote
	var ends []*element.Endnote
	walkDocument(w.doc, func(el element.Element) {
		switch v := el.(type) {
		case *element.Header:
			w.hfIndex++
			name := fmt.Sprintf("header%d.xml", w.hfIndex)
			id := w.addRel(ooxml.NSOfficeRelHeader, name, "")
			w.headers = append(w.headers, pkgHF{RelID: id, Name: "word/" + name, Type: v.HeaderType, Kind: "header", El: v})
		case *element.Footer:
			w.hfIndex++
			name := fmt.Sprintf("footer%d.xml", w.hfIndex)
			id := w.addRel(ooxml.NSOfficeRelFooter, name, "")
			w.footers = append(w.footers, pkgHF{RelID: id, Name: "word/" + name, Type: v.HeaderType, Kind: "footer", El: v})
		case *element.Image:
			if err := w.registerImage(v); err != nil {
				// skip unreadable images; still produce a valid package
			}
		case *element.Link:
			if !v.Internal && v.Target != "" {
				w.addRel(ooxml.NSOfficeRelHyperlink, v.Target, "External")
			}
		case *element.Footnote:
			notes = append(notes, v)
		case *element.Endnote:
			ends = append(ends, v)
		case *element.Title:
			w.doc.titles = append(w.doc.titles, v)
		case *element.Chart:
			w.chartIndex++
			name := fmt.Sprintf("charts/chart%d.xml", w.chartIndex)
			id := w.addRel(ooxml.NSOfficeRelChart, name, "")
			v.RelationID = relIDNum(id)
			w.charts = append(w.charts, pkgChart{RelID: id, Name: "word/" + name, El: v})
		case *element.Comment:
			w.commentIndex++
			v.CommentID = w.commentIndex
			w.comments = append(w.comments, v)
		case *element.OLEObject:
			w.registerOLE(v)
		}
	})
	w.doc.footnotes = notes
	w.doc.endnotes = ends
	if len(notes) > 0 {
		w.addRel(ooxml.NSOfficeRelFootnotes, "footnotes.xml", "")
	}
	if len(ends) > 0 {
		w.addRel(ooxml.NSOfficeRelEndnotes, "endnotes.xml", "")
	}
	if len(w.comments) > 0 {
		w.addRel(ooxml.NSOfficeRelComments, "comments.xml", "")
	}
	return nil
}

func (w *word2007Writer) registerOLE(o *element.OLEObject) {
	data := o.Media.Data
	if len(data) == 0 && o.Source != "" {
		data, _ = common.ReadFile(o.Source)
	}
	if len(data) == 0 {
		return
	}
	w.oleIndex++
	name := fmt.Sprintf("embeddings/oleObject%d.bin", w.oleIndex)
	id := w.addRel(ooxml.NSOfficeRelOleObject, name, "")
	o.RelationID = relIDNum(id)
	w.oles = append(w.oles, pkgOLE{RelID: id, Name: "word/" + name, Data: data, El: o})
}

func (w *word2007Writer) registerImage(img *element.Image) error {
	data := img.Data
	var err error
	if len(data) == 0 && img.Source != "" {
		data, err = common.ReadFile(img.Source)
		if err != nil {
			return err
		}
		img.Data = data
	}
	if len(data) == 0 {
		return fmt.Errorf("word: empty image")
	}
	cfg, format, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		format = img.Media.Ext
		if format == "" {
			format = "png"
		}
	} else {
		img.Style.Name = format
		if img.Style.Width == 0 {
			img.Style.Width = float64(cfg.Width)
		}
		if img.Style.Height == 0 {
			img.Style.Height = float64(cfg.Height)
		}
	}
	if format == "jpg" {
		format = "jpeg"
	}
	w.imgIndex++
	name := fmt.Sprintf("word/media/image%d.%s", w.imgIndex, format)
	id := w.addRel(ooxml.NSOfficeRelImage, path.Base(path.Dir(name))+"/"+path.Base(name), "")
	// Target relative to word/: media/imageN.ext
	w.rels[len(w.rels)-1].Target = "media/" + path.Base(name)
	img.RelationID = relIDNum(id)
	w.images = append(w.images, pkgImage{RelID: id, Name: name, Data: data, Ext: format, El: img})
	return nil
}

func relIDNum(id string) int {
	n, _ := strconv.Atoi(strings.TrimPrefix(id, "rId"))
	return n
}

func (w *word2007Writer) relFor(el element.Element) string {
	switch v := el.(type) {
	case *element.Image:
		for _, img := range w.images {
			if img.El == v {
				return img.RelID
			}
		}
	case *element.Header:
		for _, h := range w.headers {
			if h.El == v {
				return h.RelID
			}
		}
	case *element.Footer:
		for _, f := range w.footers {
			if f.El == v {
				return f.RelID
			}
		}
	case *element.Link:
		for _, r := range w.rels {
			if r.Type == ooxml.NSOfficeRelHyperlink && r.Target == v.Target {
				return r.ID
			}
		}
	case *element.Chart:
		for _, ch := range w.charts {
			if ch.El == v {
				return ch.RelID
			}
		}
	case *element.OLEObject:
		for _, o := range w.oles {
			if o.El == v {
				return o.RelID
			}
		}
	}
	return ""
}

func mimeForExt(ext string) string {
	switch strings.ToLower(ext) {
	case "png":
		return "image/png"
	case "jpeg", "jpg":
		return "image/jpeg"
	case "gif":
		return "image/gif"
	case "emf":
		return "image/x-emf"
	case "wmf":
		return "image/x-wmf"
	default:
		return "application/octet-stream"
	}
}
