package word

import (
	"bytes"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"os"
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
	revIndex     int
	shapeIndex   int
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
	Rels  []ooxml.Relationship
}

func newWord2007Writer(doc *Document) *word2007Writer {
	return &word2007Writer{doc: doc, nextRel: 1}
}

func (w *word2007Writer) Save(filename string) error {
	if filename == "" {
		return fmt.Errorf("word: empty filename")
	}
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	_, err = w.WriteTo(f)
	if cerr := f.Close(); err == nil {
		err = cerr
	}
	return err
}

type countingWriter struct {
	w io.Writer
	n int64
}

func (c *countingWriter) Write(p []byte) (int, error) {
	n, err := c.w.Write(p)
	c.n += int64(n)
	return n, err
}

func addXML(zw *common.ZipWriter, name string, write func(*common.XMLWriter)) error {
	xw := common.GetXMLWriter()
	write(xw)
	err := zw.AddFile(name, xw.Bytes())
	common.PutXMLWriter(xw)
	return err
}

type zipPart struct {
	name  string
	data  []byte
	write func(*common.XMLWriter)
}

func addParts(zw *common.ZipWriter, parts []zipPart) error {
	for _, p := range parts {
		if p.write != nil {
			if err := addXML(zw, p.name, p.write); err != nil {
				return err
			}
			continue
		}
		if err := zw.AddFile(p.name, p.data); err != nil {
			return err
		}
	}
	return nil
}

func (w *word2007Writer) WriteTo(dest io.Writer) (int64, error) {
	if err := w.prepare(); err != nil {
		return 0, err
	}
	cw := &countingWriter{w: dest}
	zw := common.NewZipWriter(cw)
	if err := w.writeZip(zw); err != nil {
		_ = zw.Close()
		return cw.n, err
	}
	if err := zw.Close(); err != nil {
		return cw.n, err
	}
	return cw.n, nil
}

func (w *word2007Writer) writeZip(zw *common.ZipWriter) error {
	if err := w.writeZipMeta(zw); err != nil {
		return err
	}
	fw, err := zw.Create("word/document.xml")
	if err != nil {
		return err
	}
	xw := common.NewXMLWriterTo(fw)
	w.writeDocument(xw)
	if err := xw.Flush(); err != nil {
		return err
	}
	return w.writeZipAfterDocument(zw)
}

func (w *word2007Writer) writeSupportingParts(zw *common.ZipWriter) error {
	if err := w.writeZipMeta(zw); err != nil {
		return err
	}
	return w.writeZipAfterDocument(zw)
}

func (w *word2007Writer) writeZipMeta(zw *common.ZipWriter) error {
	return addParts(zw, []zipPart{
		{name: "[Content_Types].xml", data: w.contentTypes()},
		{name: "_rels/.rels", data: w.pkgRels()},
		{name: "docProps/core.xml", data: w.coreProps()},
		{name: "docProps/app.xml", data: w.appProps()},
	})
}

func (w *word2007Writer) writeZipAfterDocument(zw *common.ZipWriter) error {
	parts := []zipPart{
		{name: "word/_rels/document.xml.rels", data: w.docRels()},
		{name: "word/styles.xml", write: w.writeStylesXML},
		{name: "word/numbering.xml", write: w.writeNumberingXML},
		{name: "word/settings.xml", write: w.writeSettingsXML},
		{name: "word/webSettings.xml", data: []byte(webSettingsXML)},
		{name: "word/fontTable.xml", data: []byte(fontTableXML)},
		{name: "word/theme/theme1.xml", data: []byte(themeXML)},
	}
	if len(w.doc.info.Custom) > 0 {
		parts = append(parts, zipPart{name: "docProps/custom.xml", data: w.customProps()})
	}
	for _, h := range w.headers {
		h := h
		parts = append(parts, zipPart{name: h.Name, write: func(xw *common.XMLWriter) {
			w.writeHdrFtr(xw, "w:hdr", h.El)
		}})
		if len(h.Rels) > 0 {
			rels := ooxml.NewRelationships()
			rels.Rel = h.Rels
			data, _ := common.MarshalXML(rels)
			parts = append(parts, zipPart{name: "word/_rels/" + path.Base(h.Name) + ".rels", data: data})
		}
	}
	for _, f := range w.footers {
		f := f
		parts = append(parts, zipPart{name: f.Name, write: func(xw *common.XMLWriter) {
			w.writeHdrFtr(xw, "w:ftr", f.El)
		}})
		if len(f.Rels) > 0 {
			rels := ooxml.NewRelationships()
			rels.Rel = f.Rels
			data, _ := common.MarshalXML(rels)
			parts = append(parts, zipPart{name: "word/_rels/" + path.Base(f.Name) + ".rels", data: data})
		}
	}
	if len(w.doc.footnotes) > 0 {
		parts = append(parts, zipPart{name: "word/footnotes.xml", write: func(xw *common.XMLWriter) {
			w.writeNotesXML(xw, true)
		}})
	}
	if len(w.doc.endnotes) > 0 {
		parts = append(parts, zipPart{name: "word/endnotes.xml", write: func(xw *common.XMLWriter) {
			w.writeNotesXML(xw, false)
		}})
	}
	if len(w.comments) > 0 {
		parts = append(parts, zipPart{name: "word/comments.xml", write: w.writeCommentsXML})
	}
	for _, ch := range w.charts {
		ch := ch
		parts = append(parts, zipPart{name: ch.Name, write: func(xw *common.XMLWriter) {
			writeChartPart(xw, ch.El)
		}})
	}
	for _, o := range w.oles {
		parts = append(parts, zipPart{name: o.Name, data: o.Data})
	}
	for _, img := range w.images {
		parts = append(parts, zipPart{name: img.Name, data: img.Data})
	}
	return addParts(zw, parts)
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
	w.revIndex = 0
	w.doc.titles = nil

	w.doc.applyWatermarks()
	w.doc.syncEvenAndOddHeaders()

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
			// Body vs header/footer images are registered after the walk so
			// watermark pictures land in headerN.xml.rels, not document.xml.rels.
		case *element.Link:
			if !v.Internal && v.Target != "" {
				w.addRel(ooxml.NSOfficeRelHyperlink, v.Target, "External")
			}
		case *element.Footnote:
			notes = append(notes, v)
		case *element.Endnote:
			ends = append(ends, v)
		case *element.Title:
			if v.BookmarkName == "" {
				v.BookmarkName = "_Toc" + strconv.Itoa(len(w.doc.titles)+1)
			}
			w.doc.titles = append(w.doc.titles, v)
		case *element.Chart:
			w.chartIndex++
			name := fmt.Sprintf("charts/chart%d.xml", w.chartIndex)
			id := w.addRel(ooxml.NSOfficeRelChart, name, "")
			v.RelationID = relIDNum(id)
			w.charts = append(w.charts, pkgChart{RelID: id, Name: "word/" + name, El: v})
		case *element.Comment:
			w.registerComment(v)
		case *element.OLEObject:
			w.registerOLE(v)
		}
		if g, ok := el.(interface{ GetCommentRangeStart() *element.Comment }); ok {
			w.registerComment(g.GetCommentRangeStart())
		}
		if g, ok := el.(interface{ GetCommentRangeEnd() *element.Comment }); ok {
			w.registerComment(g.GetCommentRangeEnd())
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
	for _, sec := range w.doc.sections {
		for _, el := range sec.Elements() {
			walkElement(el, func(e element.Element) {
				if img, ok := e.(*element.Image); ok {
					_ = w.registerImage(img)
				}
			})
		}
	}
	for i := range w.headers {
		hf := &w.headers[i]
		walkElement(hf.El, func(e element.Element) {
			if img, ok := e.(*element.Image); ok {
				_ = w.registerImageIn(img, hf)
			}
		})
	}
	for i := range w.footers {
		hf := &w.footers[i]
		walkElement(hf.El, func(e element.Element) {
			if img, ok := e.(*element.Image); ok {
				_ = w.registerImageIn(img, hf)
			}
		})
	}
	return nil
}

func (w *word2007Writer) registerComment(c *element.Comment) {
	if c == nil || c.CommentID > 0 {
		return
	}
	w.commentIndex++
	c.CommentID = w.commentIndex
	w.comments = append(w.comments, c)
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
	return w.addImagePart(img, nil)
}

func (w *word2007Writer) registerImageIn(img *element.Image, hf *pkgHF) error {
	return w.addImagePart(img, hf)
}

func (w *word2007Writer) addImagePart(img *element.Image, hf *pkgHF) error {
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
	target := "media/" + path.Base(name)
	var id string
	if hf != nil {
		id = "rId" + strconv.Itoa(len(hf.Rels)+1)
		hf.Rels = append(hf.Rels, ooxml.Relationship{ID: id, Type: ooxml.NSOfficeRelImage, Target: target})
	} else {
		id = w.addRel(ooxml.NSOfficeRelImage, target, "")
		w.rels[len(w.rels)-1].Target = target
	}
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
