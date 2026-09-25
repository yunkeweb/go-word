package word

import (
	"strconv"

	"github.com/yunkeweb/go-word/element"
	"github.com/yunkeweb/go-word/pkg/common"
	"github.com/yunkeweb/go-word/pkg/math"
	"github.com/yunkeweb/go-word/style"
)

func (w *word2007Writer) writeContainer(xw *common.XMLWriter, els []element.Element, inline bool) {
	for _, el := range els {
		w.writeElement(xw, el, inline)
	}
}

func (w *word2007Writer) writeElement(xw *common.XMLWriter, el element.Element, inline bool) {
	if cr, ok := el.(interface {
		GetCommentRangeStart() *element.Comment
		GetCommentRangeEnd() *element.Comment
	}); ok {
		if c := cr.GetCommentRangeStart(); c != nil && c.CommentID > 0 {
			xw.Empty("w:commentRangeStart", "w:id", itoa(c.CommentID))
		}
		defer func() {
			if c := cr.GetCommentRangeEnd(); c != nil && c.CommentID > 0 {
				xw.Empty("w:commentRangeEnd", "w:id", itoa(c.CommentID))
				xw.Start("w:r")
				xw.Empty("w:commentReference", "w:id", itoa(c.CommentID))
				xw.End()
			}
		}()
	}
	switch v := el.(type) {
	case *element.Text:
		w.writeText(xw, v, inline)
	case *element.TextRun:
		w.writeTextRun(xw, v)
	case *element.TextBreak:
		xw.Start("w:p")
		xw.End()
	case *element.PageBreak:
		xw.Start("w:p")
		xw.Start("w:r")
		xw.Empty("w:br", "w:type", "page")
		xw.End()
		xw.End()
	case *element.Link:
		w.writeLink(xw, v, inline)
	case *element.Bookmark:
		w.writeBookmark(xw, v)
	case *element.Title:
		w.writeTitle(xw, v)
	case *element.ListItem:
		w.writeListItem(xw, v)
	case *element.ListItemRun:
		w.writeListItemRun(xw, v)
	case *element.Table:
		w.writeTable(xw, v)
	case *element.Image:
		if inline {
			w.writeImage(xw, v)
		} else {
			xw.Start("w:p")
			w.writeImage(xw, v)
			xw.End()
		}
	case *element.PreserveText:
		w.writePreserve(xw, v)
	case *element.CheckBox:
		w.writeCheckBox(xw, v)
	case *element.Field:
		w.writeField(xw, v)
	case *element.Footnote:
		w.writeNoteRef(xw, v, true)
	case *element.Endnote:
		w.writeNoteRef(xw, v, false)
	case *element.Formula:
		w.writeFormula(xw, v)
	case *element.TOC:
		w.writeTOC(xw, v)
	case *element.SDT:
		w.writeSDT(xw, v)
	case *element.TextBox:
		w.writeTextBox(xw, v)
	case *element.Line:
		w.writeLine(xw, v)
	case *element.Shape:
		w.writeShape(xw, v)
	case *element.FormField:
		w.writeFormField(xw, v)
	case *element.Ruby:
		w.writeRuby(xw, v)
	case *element.Chart:
		w.writeChart(xw, v, inline)
	case *element.Comment:
		w.writeCommentRef(xw, v, inline)
	case *element.OLEObject:
		w.writeOLE(xw, v, inline)
	default:
		// skip unknown
	}
}

func (w *word2007Writer) writeChart(xw *common.XMLWriter, ch *element.Chart, inline bool) {
	rid := w.relFor(ch)
	if rid == "" {
		return
	}
	xw.Raw(chartDrawingXML(rid, ch, !inline))
}

func (w *word2007Writer) writeCommentRef(xw *common.XMLWriter, c *element.Comment, inline bool) {
	if c.CommentID == 0 {
		return
	}
	if !inline {
		xw.Start("w:p")
	}
	xw.Empty("w:commentRangeStart", "w:id", itoa(c.CommentID))
	xw.Empty("w:commentRangeEnd", "w:id", itoa(c.CommentID))
	xw.Start("w:r")
	xw.Empty("w:commentReference", "w:id", itoa(c.CommentID))
	xw.End()
	if !inline {
		xw.End()
	}
}

func (w *word2007Writer) writeOLE(xw *common.XMLWriter, o *element.OLEObject, inline bool) {
	rid := w.relFor(o)
	if rid == "" {
		return
	}
	if !inline {
		xw.Start("w:p")
	}
	shapeID := "ole_" + rid
	xw.Start("w:r")
	xw.Start("w:object", "w:dxaOrig", "249", "w:dyaOrig", "160")
	xw.Start("v:shape", "id", shapeID, "type", "#_x0000_t75", "style", "width:104px;height:67px")
	xw.End()
	xw.Empty("o:OLEObject",
		"Type", "Embed",
		"ProgID", "Package",
		"ShapeID", shapeID,
		"DrawAspect", "Icon",
		"ObjectID", "_"+rid,
		"r:id", rid)
	xw.End()
	xw.End()
	if !inline {
		xw.End()
	}
}

func (w *word2007Writer) writeText(xw *common.XMLWriter, t *element.Text, inline bool) {
	if !inline {
		xw.Start("w:p")
		w.writePPrFrom(xw, t.ParagraphStyle)
	}
	xw.Start("w:r")
	w.writeRPrFrom(xw, t.FontStyle)
	xw.WT(t.Content)
	xw.End()
	if !inline {
		xw.End()
	}
}

func (w *word2007Writer) writeTextRun(xw *common.XMLWriter, tr *element.TextRun) {
	xw.Start("w:p")
	w.writePPrFrom(xw, tr.ParagraphStyle)
	w.writeContainer(xw, tr.Elements(), true)
	xw.End()
}

func (w *word2007Writer) writeLink(xw *common.XMLWriter, l *element.Link, inline bool) {
	if !inline {
		xw.Start("w:p")
		w.writePPrFrom(xw, l.ParagraphStyle)
	}
	attrs := []string{"w:history", "1"}
	if l.Internal {
		attrs = append(attrs, "w:anchor", l.Target)
	} else if id := w.relFor(l); id != "" {
		attrs = append(attrs, "r:id", id)
	}
	xw.Start("w:hyperlink", attrs...)
	xw.Start("w:r")
	fontName := "Hyperlink"
	if s, ok := l.FontStyle.(string); ok && s != "" {
		fontName = s
	}
	if f, ok := l.FontStyle.(style.Font); ok {
		w.writeRPr(xw, f, fontName)
	} else if f, ok := l.FontStyle.(*style.Font); ok && f != nil {
		w.writeRPr(xw, *f, fontName)
	} else {
		w.writeRPr(xw, style.Font{}, fontName)
	}
	xw.WT(l.Text)
	xw.End()
	xw.End()
	if !inline {
		xw.End()
	}
}

func (w *word2007Writer) writeBookmark(xw *common.XMLWriter, b *element.Bookmark) {
	w.bkIndex++
	id := itoa(w.bkIndex)
	xw.Empty("w:bookmarkStart", "w:id", id, "w:name", b.Name)
	xw.Empty("w:bookmarkEnd", "w:id", id)
}

func (w *word2007Writer) writeTitle(xw *common.XMLWriter, t *element.Title) {
	xw.Start("w:p")
	xw.Start("w:pPr")
	xw.Empty("w:pStyle", "w:val", headingStyleName(t.Depth))
	xw.Empty("w:outlineLvl", "w:val", itoa(t.Depth-1))
	xw.End()
	xw.Start("w:r")
	xw.WT(t.Text)
	xw.End()
	xw.End()
}

func (w *word2007Writer) writeListItem(xw *common.XMLWriter, item *element.ListItem) {
	xw.Start("w:p")
	xw.Start("w:pPr")
	w.writeNumPr(xw, item.ListStyle, item.Depth)
	if p, ok := item.ParagraphStyle.(style.Paragraph); ok {
		w.writePPrInner(xw, p)
	}
	xw.End()
	xw.Start("w:r")
	w.writeRPrFrom(xw, item.FontStyle)
	xw.WT(item.Text)
	xw.End()
	xw.End()
}

func (w *word2007Writer) writeListItemRun(xw *common.XMLWriter, item *element.ListItemRun) {
	xw.Start("w:p")
	xw.Start("w:pPr")
	w.writeNumPr(xw, item.ListStyle, item.Depth)
	w.writePPrFromInner(xw, item.ParagraphStyle)
	xw.End()
	w.writeContainer(xw, item.Elements(), true)
	xw.End()
}

func (w *word2007Writer) writeNumPr(xw *common.XMLWriter, list any, depth int) {
	numID := 1
	switch v := list.(type) {
	case string:
		switch v {
		case style.ListTypeNumber, style.ListTypeNumberNE, "decimal":
			numID = 2
		case style.ListTypeBullet, "":
			numID = 1
		default:
			if ns := w.doc.styleByName(v); ns != nil && ns.Kind == "numbering" {
				numID = w.numberingID(v)
			}
		}
	case style.ListItem:
		if v.NumId > 0 {
			numID = v.NumId
		} else if v.Format == style.NumberDecimal || v.ListType == style.ListTypeNumber {
			numID = 2
		}
		if v.Depth > 0 {
			depth = v.Depth
		}
	}
	xw.Start("w:numPr")
	xw.Empty("w:ilvl", "w:val", itoa(depth))
	xw.Empty("w:numId", "w:val", itoa(numID))
	xw.End()
}

func (w *word2007Writer) numberingID(name string) int {
	id := 3
	for _, ns := range w.doc.styles {
		if ns.Kind != "numbering" {
			continue
		}
		if ns.Name == name {
			return id
		}
		id++
	}
	return 1
}

func (w *word2007Writer) writeTable(xw *common.XMLWriter, tbl *element.Table) {
	xw.Start("w:tbl")
	w.writeTblPr(xw, tbl.Style)
	xw.Start("w:tblGrid")
	if len(tbl.Rows) > 0 {
		for _, c := range tbl.Rows[0].Cells {
			wd := c.Width
			if wd == 0 {
				wd = c.Style.Width
			}
			if wd == 0 {
				wd = 1440
			}
			xw.Empty("w:gridCol", "w:w", itoa(wd))
		}
	}
	xw.End()
	for _, row := range tbl.Rows {
		xw.Start("w:tr")
		if row.Style.Height > 0 || row.Style.Header || row.Style.CantSplit {
			xw.Start("w:trPr")
			if row.Style.Height > 0 {
				rule := row.Style.Rule
				if rule == "" {
					rule = "atLeast"
				}
				xw.Empty("w:trHeight", "w:val", itoa(row.Style.Height), "w:hRule", rule)
			}
			if row.Style.Header {
				xw.Empty("w:tblHeader")
			}
			if row.Style.CantSplit {
				xw.Empty("w:cantSplit")
			}
			xw.End()
		}
		for _, cell := range row.Cells {
			xw.Start("w:tc")
			w.writeTcPr(xw, cell)
			if len(cell.Elements()) == 0 {
				xw.Start("w:p")
				xw.End()
			} else {
				w.writeContainer(xw, cell.Elements(), false)
			}
			xw.End()
		}
		xw.End()
		_ = xw.Flush()
	}
	xw.End()
}

func (w *word2007Writer) writeTblPr(xw *common.XMLWriter, st style.Table) {
	xw.Start("w:tblPr")
	if st.StyleName != "" {
		xw.Empty("w:tblStyle", "w:val", st.StyleName)
	}
	unit := st.Unit
	if unit == "" {
		unit = "dxa"
	}
	if st.Width > 0 {
		xw.Empty("w:tblW", "w:w", itoa(st.Width), "w:type", unit)
	} else {
		xw.Empty("w:tblW", "w:w", "0", "w:type", "auto")
	}
	if st.Alignment != "" {
		xw.Empty("w:jc", "w:val", st.Alignment)
	}
	if st.Layout != "" {
		xw.Empty("w:tblLayout", "w:type", st.Layout)
	}
	if st.CellMarginTop != 0 || st.CellMarginLeft != 0 || st.CellMarginRight != 0 || st.CellMarginBottom != 0 {
		xw.Start("w:tblCellMar")
		if st.CellMarginTop != 0 {
			xw.Empty("w:top", "w:w", itoa(st.CellMarginTop), "w:type", "dxa")
		}
		if st.CellMarginLeft != 0 {
			xw.Empty("w:left", "w:w", itoa(st.CellMarginLeft), "w:type", "dxa")
		}
		if st.CellMarginRight != 0 {
			xw.Empty("w:right", "w:w", itoa(st.CellMarginRight), "w:type", "dxa")
		}
		if st.CellMarginBottom != 0 {
			xw.Empty("w:bottom", "w:w", itoa(st.CellMarginBottom), "w:type", "dxa")
		}
		xw.End()
	}
	w.writeBorders(xw, "w:tblBorders", st.Borders)
	if st.Shading.Fill != "" {
		w.writeShd(xw, st.Shading)
	}
	if st.Indent != 0 {
		xw.Empty("w:tblInd", "w:w", itoa(st.Indent), "w:type", "dxa")
	}
	if p := st.Position; p != nil {
		attrs := []string{}
		if p.LeftFromText != 0 {
			attrs = append(attrs, "w:leftFromText", itoa(p.LeftFromText))
		}
		if p.RightFromText != 0 {
			attrs = append(attrs, "w:rightFromText", itoa(p.RightFromText))
		}
		if p.TopFromText != 0 {
			attrs = append(attrs, "w:topFromText", itoa(p.TopFromText))
		}
		if p.BottomFromText != 0 {
			attrs = append(attrs, "w:bottomFromText", itoa(p.BottomFromText))
		}
		if p.VertAnchor != "" {
			attrs = append(attrs, "w:vertAnchor", p.VertAnchor)
		}
		if p.HorzAnchor != "" {
			attrs = append(attrs, "w:horzAnchor", p.HorzAnchor)
		}
		if p.TblpXSpec != "" {
			attrs = append(attrs, "w:tblpXSpec", p.TblpXSpec)
		}
		if p.TblpX != 0 {
			attrs = append(attrs, "w:tblpX", itoa(p.TblpX))
		}
		if p.TblpYSpec != "" {
			attrs = append(attrs, "w:tblpYSpec", p.TblpYSpec)
		}
		if p.TblpY != 0 {
			attrs = append(attrs, "w:tblpY", itoa(p.TblpY))
		}
		if len(attrs) > 0 {
			xw.Empty("w:tblpPr", attrs...)
		}
	}
	xw.End()
}

func (w *word2007Writer) writeTcPr(xw *common.XMLWriter, c *element.Cell) {
	st := c.Style
	xw.Start("w:tcPr")
	wd := c.Width
	if st.Width != 0 {
		wd = st.Width
	}
	unit := st.Unit
	if unit == "" {
		unit = "dxa"
	}
	if wd > 0 {
		xw.Empty("w:tcW", "w:w", itoa(wd), "w:type", unit)
	}
	if st.GridSpan > 1 {
		xw.Empty("w:gridSpan", "w:val", itoa(st.GridSpan))
	}
	if st.VMerge != "" {
		if st.VMerge == "continue" {
			xw.Empty("w:vMerge")
		} else {
			xw.Empty("w:vMerge", "w:val", st.VMerge)
		}
	}
	if st.VAlign != "" {
		xw.Empty("w:vAlign", "w:val", st.VAlign)
	}
	w.writeBorders(xw, "w:tcBorders", st.Borders)
	sh := st.Shading
	if st.BgColor != "" && sh.Fill == "" {
		sh.Fill = st.BgColor
		sh.Val = "clear"
	}
	if sh.Fill != "" {
		w.writeShd(xw, sh)
	}
	if st.NoWrap {
		xw.Empty("w:noWrap")
	}
	if st.PaddingTop != 0 || st.PaddingLeft != 0 || st.PaddingRight != 0 || st.PaddingBottom != 0 {
		xw.Start("w:tcMar")
		if st.PaddingTop != 0 {
			xw.Empty("w:top", "w:w", itoa(st.PaddingTop), "w:type", "dxa")
		}
		if st.PaddingLeft != 0 {
			xw.Empty("w:left", "w:w", itoa(st.PaddingLeft), "w:type", "dxa")
		}
		if st.PaddingRight != 0 {
			xw.Empty("w:right", "w:w", itoa(st.PaddingRight), "w:type", "dxa")
		}
		if st.PaddingBottom != 0 {
			xw.Empty("w:bottom", "w:w", itoa(st.PaddingBottom), "w:type", "dxa")
		}
		xw.End()
	}
	if st.TextDir != "" {
		xw.Empty("w:textDirection", "w:val", st.TextDir)
	}
	xw.End()
}

func (w *word2007Writer) writeImage(xw *common.XMLWriter, img *element.Image) {
	rid := w.relFor(img)
	if rid == "" {
		return
	}
	cx := img.Style.WidthEMU
	cy := img.Style.HeightEMU
	if cx == 0 {
		cx = common.PixelToEMU(img.Style.Width)
	}
	if cy == 0 {
		cy = common.PixelToEMU(img.Style.Height)
	}
	if cx == 0 {
		cx = common.PixelToEMU(100)
	}
	if cy == 0 {
		cy = common.PixelToEMU(100)
	}
	docPrID := img.RelationID
	if docPrID == 0 {
		docPrID = 1
	}
	name := img.Style.Name
	if name == "" {
		name = "Picture"
	}
	xw.Start("w:r")
	xw.Start("w:drawing")
	xw.Start("wp:inline", "distT", "0", "distB", "0", "distL", "0", "distR", "0")
	xw.Empty("wp:extent", "cx", itoa64(cx), "cy", itoa64(cy))
	xw.Empty("wp:effectExtent", "l", "0", "t", "0", "r", "0", "b", "0")
	xw.Start("wp:docPr", "id", itoa(docPrID), "name", name, "descr", img.Style.AltText)
	xw.End()
	xw.Start("wp:cNvGraphicFramePr")
	xw.Empty("a:graphicFrameLocks", "xmlns:a", "http://schemas.openxmlformats.org/drawingml/2006/main", "noChangeAspect", "1")
	xw.End()
	xw.Start("a:graphic", "xmlns:a", "http://schemas.openxmlformats.org/drawingml/2006/main")
	xw.Start("a:graphicData", "uri", "http://schemas.openxmlformats.org/drawingml/2006/picture")
	xw.Start("pic:pic", "xmlns:pic", "http://schemas.openxmlformats.org/drawingml/2006/picture")
	xw.Start("pic:nvPicPr")
	xw.Empty("pic:cNvPr", "id", "0", "name", name)
	xw.Start("pic:cNvPicPr")
	xw.Empty("a:picLocks", "noChangeAspect", "1")
	xw.End()
	xw.End()
	xw.Start("pic:blipFill")
	xw.Empty("a:blip", "r:embed", rid)
	xw.Start("a:stretch")
	xw.Empty("a:fillRect")
	xw.End()
	xw.End()
	xw.Start("pic:spPr")
	xw.Start("a:xfrm")
	xw.Empty("a:off", "x", "0", "y", "0")
	xw.Empty("a:ext", "cx", itoa64(cx), "cy", itoa64(cy))
	xw.End()
	xw.Start("a:prstGeom", "prst", "rect")
	xw.Empty("a:avLst")
	xw.End()
	xw.End()
	xw.End() // pic
	xw.End() // graphicData
	xw.End() // graphic
	xw.End() // inline
	xw.End() // drawing
	xw.End() // r
}

func itoa64(n int64) string { return strconv.FormatInt(n, 10) }

func (w *word2007Writer) writePreserve(xw *common.XMLWriter, p *element.PreserveText) {
	xw.Start("w:p")
	w.writePPrFrom(xw, p.ParagraphStyle)
	// PHPWord splits ${VAR} into fldChar begin / instrText / fldChar end
	xw.Start("w:r")
	w.writeRPrFrom(xw, p.FontStyle)
	xw.Start("w:instrText", "xml:space", "preserve")
	xw.Text(p.Content)
	xw.End()
	xw.End()
	xw.End()
}

func (w *word2007Writer) writeCheckBox(xw *common.XMLWriter, cb *element.CheckBox) {
	xw.Start("w:p")
	w.writePPrFrom(xw, cb.ParagraphStyle)
	xw.Start("w:r")
	xw.Start("w:fldChar", "w:fldCharType", "begin")
	xw.Start("w:ffData")
	xw.Empty("w:name", "w:val", cb.Name)
	xw.Empty("w:enabled")
	xw.Start("w:checkBox")
	if cb.Checked {
		xw.Empty("w:checked")
	}
	xw.End()
	xw.End()
	xw.End()
	xw.End()
	xw.Start("w:r")
	w.writeRPrFrom(xw, cb.FontStyle)
	xw.WT(cb.Content)
	xw.End()
	xw.End()
}

func (w *word2007Writer) writeField(xw *common.XMLWriter, f *element.Field) {
	instr := f.FieldType
	if instr == "" {
		instr = "PAGE"
	}
	for _, o := range f.Options {
		instr += " " + o
	}
	if f.Text != "" {
		instr += " " + f.Text
	}
	xw.Start("w:p")
	writeFieldRun(xw, instr)
	xw.End()
}

func writeFieldRun(xw *common.XMLWriter, instr string) {
	xw.Start("w:r")
	xw.Empty("w:fldChar", "w:fldCharType", "begin")
	xw.End()
	xw.Start("w:r")
	xw.Start("w:instrText", "xml:space", "preserve")
	xw.Text(" " + instr + " ")
	xw.End()
	xw.End()
	xw.Start("w:r")
	xw.Empty("w:fldChar", "w:fldCharType", "end")
	xw.End()
}

func (w *word2007Writer) writeNoteRef(xw *common.XMLWriter, el element.Element, foot bool) {
	id := 1
	if foot {
		for i, n := range w.doc.footnotes {
			if n == el {
				id = i + 1
				break
			}
		}
	} else {
		for i, n := range w.doc.endnotes {
			if n == el {
				id = i + 1
				break
			}
		}
	}
	xw.Start("w:r")
	xw.Start("w:rPr")
	xw.Empty("w:vertAlign", "w:val", "superscript")
	xw.End()
	if foot {
		xw.Empty("w:footnoteReference", "w:id", itoa(id))
	} else {
		xw.Empty("w:endnoteReference", "w:id", itoa(id))
	}
	xw.End()
}

func (w *word2007Writer) writeFormula(xw *common.XMLWriter, f *element.Formula) {
	xw.Start("w:p")
	if f.Math != nil {
		b, err := math.WriteOMML(f.Math)
		if err == nil {
			xw.Raw(string(b))
		}
	}
	xw.End()
}

func (w *word2007Writer) writeTOC(xw *common.XMLWriter, toc *element.TOC) {
	_ = toc
	xw.Start("w:p")
	writeFieldRun(xw, `TOC \o "1-9" \h \z \u`)
	xw.End()
}

func (w *word2007Writer) writeSDT(xw *common.XMLWriter, s *element.SDT) {
	xw.Start("w:sdt")
	xw.Start("w:sdtPr")
	if s.Alias != "" {
		xw.Empty("w:alias", "w:val", s.Alias)
	}
	if s.Tag != "" {
		xw.Empty("w:tag", "w:val", s.Tag)
	}
	xw.Empty("w:id", "w:val", "1")
	xw.End()
	xw.Start("w:sdtContent")
	if len(s.Elements()) == 0 {
		xw.Start("w:p")
		xw.Start("w:r")
		xw.WT(s.Value)
		xw.End()
		xw.End()
	} else {
		w.writeContainer(xw, s.Elements(), false)
	}
	xw.End()
	xw.End()
}

func (w *word2007Writer) writeTextBox(xw *common.XMLWriter, tb *element.TextBox) {
	xw.Start("w:p")
	xw.Start("w:r")
	xw.Start("w:pict")
	xw.Start("v:shape", "type", "#_x0000_t202", "style", "width:100pt;height:50pt")
	xw.Start("v:textbox")
	xw.Start("w:txbxContent")
	w.writeContainer(xw, tb.Elements(), false)
	xw.End()
	xw.End()
	xw.End()
	xw.End()
	xw.End()
	xw.End()
}

func (w *word2007Writer) writeLine(xw *common.XMLWriter, ln *element.Line) {
	xw.Start("w:p")
	xw.Start("w:r")
	xw.Start("w:pict")
	xw.Empty("v:line",
		"from", "0,0",
		"to", itoa(nonzero(ln.Style.Width, 100))+","+itoa(ln.Style.Height),
		"strokecolor", nonEmpty(ln.Style.Color, "000000"),
		"strokeweight", itoa(nonzero(ln.Style.Weight, 1))+"pt")
	xw.End()
	xw.End()
	xw.End()
}

func (w *word2007Writer) writeShape(xw *common.XMLWriter, sh *element.Shape) {
	xw.Start("w:p")
	xw.Start("w:r")
	xw.Start("w:pict")
	xw.Start("v:shape", "type", "#_x0000_t"+sh.ShapeType, "style", "width:100pt;height:50pt")
	if sh.Style.Fill.Color != "" {
		xw.Empty("v:fill", "color", "#"+sh.Style.Fill.Color)
	}
	xw.End()
	xw.End()
	xw.End()
	xw.End()
}

func (w *word2007Writer) writeFormField(xw *common.XMLWriter, f *element.FormField) {
	xw.Start("w:p")
	xw.Start("w:r")
	xw.Start("w:fldChar", "w:fldCharType", "begin")
	xw.Start("w:ffData")
	xw.Empty("w:name", "w:val", f.Name)
	switch f.FormType {
	case "checkbox":
		xw.Start("w:checkBox")
		xw.End()
	case "dropdown":
		xw.Start("w:ddList")
		for _, it := range f.Items {
			xw.Empty("w:listEntry", "w:val", it)
		}
		xw.End()
	default:
		xw.Start("w:textInput")
		if f.Value != "" {
			xw.Empty("w:default", "w:val", f.Value)
		}
		xw.End()
	}
	xw.End()
	xw.End()
	xw.End()
	xw.End()
}

func (w *word2007Writer) writeRuby(xw *common.XMLWriter, r *element.Ruby) {
	xw.Start("w:p")
	xw.Start("w:r")
	xw.Start("w:ruby")
	xw.Start("w:rubyPr")
	xw.Empty("w:rubyAlign", "w:val", nonEmpty(r.Properties.Alignment, "center"))
	xw.End()
	xw.Start("w:rt")
	if r.RubyText != nil {
		w.writeContainer(xw, r.RubyText.Elements(), true)
	}
	xw.End()
	xw.Start("w:rubyBase")
	if r.BaseText != nil {
		w.writeContainer(xw, r.BaseText.Elements(), true)
	}
	xw.End()
	xw.End()
	xw.End()
	xw.End()
}

func (w *word2007Writer) writePPrFrom(xw *common.XMLWriter, v any) {
	name, p := splitPara(v)
	if name == "" && (p == nil || p.IsZero()) {
		return
	}
	if p == nil {
		p = &style.Paragraph{}
	}
	w.writePPr(xw, *p, name)
}

func (w *word2007Writer) writePPrFromInner(xw *common.XMLWriter, v any) {
	name, p := splitPara(v)
	if name != "" {
		xw.Empty("w:pStyle", "w:val", name)
	}
	if p != nil {
		w.writePPrInner(xw, *p)
	}
}

func splitPara(v any) (string, *style.Paragraph) {
	switch x := v.(type) {
	case string:
		return x, nil
	case style.Paragraph:
		return x.StyleName, &x
	case *style.Paragraph:
		if x == nil {
			return "", nil
		}
		return x.StyleName, x
	default:
		return "", nil
	}
}

func (w *word2007Writer) writePPr(xw *common.XMLWriter, p style.Paragraph, styleName string) {
	xw.Start("w:pPr")
	if styleName != "" {
		xw.Empty("w:pStyle", "w:val", styleName)
	}
	w.writePPrInner(xw, p)
	xw.End()
}

func (w *word2007Writer) writePPrInner(xw *common.XMLWriter, p style.Paragraph) {
	if p.KeepNext {
		xw.Empty("w:keepNext")
	}
	if p.KeepLines {
		xw.Empty("w:keepLines")
	}
	if p.PageBreakBefore {
		xw.Empty("w:pageBreakBefore")
	}
	if p.WidowControl != nil && !*p.WidowControl {
		xw.Empty("w:widowControl", "w:val", "0")
	}
	if p.NumStyle != "" {
		w.writeNumPr(xw, p.NumStyle, p.NumLevel)
	}
	if p.Spacing.Before != 0 || p.Spacing.After != 0 || p.Spacing.Line != 0 {
		attrs := []string{}
		if p.Spacing.Before != 0 {
			attrs = append(attrs, "w:before", itoa(p.Spacing.Before))
		}
		if p.Spacing.After != 0 {
			attrs = append(attrs, "w:after", itoa(p.Spacing.After))
		}
		if p.Spacing.Line != 0 {
			attrs = append(attrs, "w:line", itoa(p.Spacing.Line))
			rule := p.Spacing.Rule
			if rule == "" {
				rule = "auto"
			}
			attrs = append(attrs, "w:lineRule", rule)
		}
		xw.Empty("w:spacing", attrs...)
	}
	ind := p.Indentation
	if ind != (style.Indentation{}) {
		attrs := []string{}
		if ind.Left != 0 {
			attrs = append(attrs, "w:left", itoa(ind.Left))
		}
		if ind.Right != 0 {
			attrs = append(attrs, "w:right", itoa(ind.Right))
		}
		if ind.FirstLine != 0 {
			attrs = append(attrs, "w:firstLine", itoa(ind.FirstLine))
		}
		if ind.Hanging != 0 {
			attrs = append(attrs, "w:hanging", itoa(ind.Hanging))
		}
		if ind.FirstLineChars != 0 {
			attrs = append(attrs, "w:firstLineChars", itoa(ind.FirstLineChars))
		}
		xw.Empty("w:ind", attrs...)
	}
	if p.Alignment != "" {
		xw.Empty("w:jc", "w:val", p.Alignment)
	}
	if p.OutlineLevel > 0 {
		xw.Empty("w:outlineLvl", "w:val", itoa(p.OutlineLevel-1))
	}
	if p.Bidi {
		xw.Empty("w:bidi")
	}
	if p.ContextualSpacing {
		xw.Empty("w:contextualSpacing")
	}
	if p.TextAlignment != "" {
		xw.Empty("w:textAlignment", "w:val", p.TextAlignment)
	}
	if p.SuppressAutoHyphens {
		xw.Empty("w:suppressAutoHyphens")
	}
	w.writeBorders(xw, "w:pBdr", p.Borders)
	if p.Shading.Fill != "" {
		w.writeShd(xw, p.Shading)
	}
	if len(p.Tabs) > 0 {
		xw.Start("w:tabs")
		for _, tab := range p.Tabs {
			xw.Empty("w:tab", "w:val", nonEmpty(tab.Val, "left"), "w:leader", tab.Leader, "w:pos", itoa(tab.Pos))
		}
		xw.End()
	}
}

func (w *word2007Writer) writeRPrFrom(xw *common.XMLWriter, v any) {
	name, f := splitFont(v)
	if name == "" && (f == nil || f.IsZero()) {
		return
	}
	if f == nil {
		f = &style.Font{}
	}
	w.writeRPr(xw, *f, name)
}

func splitFont(v any) (string, *style.Font) {
	switch x := v.(type) {
	case string:
		return x, nil
	case style.Font:
		return "", &x
	case *style.Font:
		return "", x
	default:
		return "", nil
	}
}

func (w *word2007Writer) writeRPr(xw *common.XMLWriter, f style.Font, styleName string) {
	xw.Start("w:rPr")
	if styleName != "" {
		xw.Empty("w:rStyle", "w:val", styleName)
	}
	if f.Name != "" {
		xw.Empty("w:rFonts", "w:ascii", f.Name, "w:hAnsi", f.Name, "w:eastAsia", f.Name, "w:cs", f.Name, "w:hint", f.Hint)
	}
	if f.Bold {
		xw.Empty("w:b")
		xw.Empty("w:bCs")
	}
	if f.Italic {
		xw.Empty("w:i")
		xw.Empty("w:iCs")
	}
	if f.SmallCaps {
		xw.Empty("w:smallCaps")
	}
	if f.AllCaps {
		xw.Empty("w:caps")
	}
	if f.Strikethrough {
		xw.Empty("w:strike")
	}
	if f.DoubleStrikethrough {
		xw.Empty("w:dstrike")
	}
	if f.Hidden {
		xw.Empty("w:vanish")
	}
	if f.Color != "" {
		xw.Empty("w:color", "w:val", f.Color)
	}
	if f.Spacing != 0 {
		xw.Empty("w:spacing", "w:val", itoa(f.Spacing))
	}
	if hp := f.HalfPoints(); hp > 0 {
		xw.Empty("w:sz", "w:val", itoa(hp))
		xw.Empty("w:szCs", "w:val", itoa(hp))
	}
	if f.FgColor != "" {
		xw.Empty("w:highlight", "w:val", f.FgColor)
	}
	if f.Underline != "" && f.Underline != style.UnderlineNone {
		xw.Empty("w:u", "w:val", f.Underline)
	}
	if f.SuperScript {
		xw.Empty("w:vertAlign", "w:val", "superscript")
	} else if f.SubScript {
		xw.Empty("w:vertAlign", "w:val", "subscript")
	}
	if f.BgColor != "" {
		w.writeShd(xw, style.Shading{Val: "clear", Fill: f.BgColor})
	}
	if f.RTL {
		xw.Empty("w:rtl")
	}
	if f.Lang != "" {
		xw.Empty("w:lang", "w:val", f.Lang)
	}
	if f.NoProof {
		xw.Empty("w:noProof")
	}
	if f.Shading.Fill != "" || f.Shading.Val != "" {
		w.writeShd(xw, f.Shading)
	}
	xw.End()
}

func (w *word2007Writer) writeBorders(xw *common.XMLWriter, tag string, b style.Borders) {
	if b == (style.Borders{}) {
		return
	}
	xw.Start(tag)
	writeBorder := func(name string, side style.Border) {
		if side.Style == "" && side.Size == 0 && side.Color == "" {
			return
		}
		st := side.Style
		if st == "" {
			st = "single"
		}
		xw.Empty(name, "w:val", st, "w:sz", itoa(nonzero(side.Size, 4)), "w:space", itoa(side.Space), "w:color", nonEmpty(side.Color, "auto"))
	}
	writeBorder("w:top", b.Top)
	writeBorder("w:left", b.Left)
	writeBorder("w:bottom", b.Bottom)
	writeBorder("w:right", b.Right)
	writeBorder("w:insideH", b.InsideH)
	writeBorder("w:insideV", b.InsideV)
	xw.End()
}

func (w *word2007Writer) writeShd(xw *common.XMLWriter, s style.Shading) {
	val := s.Val
	if val == "" {
		val = "clear"
	}
	xw.Empty("w:shd", "w:val", val, "w:color", nonEmpty(s.Color, "auto"), "w:fill", s.Fill)
}

func nonEmpty(s, def string) string {
	if s == "" {
		return def
	}
	return s
}
