package word

import (
	"strconv"
	"time"

	"github.com/yunkeweb/go-word/element"
	"github.com/yunkeweb/go-word/ooxml"
	"github.com/yunkeweb/go-word/pkg/common"
	"github.com/yunkeweb/go-word/style"
)

func (w *word2007Writer) contentTypes() []byte {
	ct := ooxml.NewContentTypes()
	for _, img := range w.images {
		ct.AddDefault(img.Ext, mimeForExt(img.Ext))
	}
	ct.AddOverride("/word/document.xml", ooxml.CTDocument)
	ct.AddOverride("/word/styles.xml", ooxml.CTStyles)
	ct.AddOverride("/word/numbering.xml", ooxml.CTNumbering)
	ct.AddOverride("/word/settings.xml", ooxml.CTSettings)
	ct.AddOverride("/word/webSettings.xml", ooxml.CTWebSettings)
	ct.AddOverride("/word/fontTable.xml", ooxml.CTFontTable)
	ct.AddOverride("/word/theme/theme1.xml", ooxml.CTTheme)
	ct.AddOverride("/docProps/core.xml", ooxml.CTCore)
	ct.AddOverride("/docProps/app.xml", ooxml.CTApp)
	if len(w.doc.info.Custom) > 0 {
		ct.AddOverride("/docProps/custom.xml", ooxml.CTCustom)
	}
	for _, h := range w.headers {
		ct.AddOverride("/"+h.Name, ooxml.CTHeader)
	}
	for _, f := range w.footers {
		ct.AddOverride("/"+f.Name, ooxml.CTFooter)
	}
	if len(w.doc.footnotes) > 0 {
		ct.AddOverride("/word/footnotes.xml", ooxml.CTFootnotes)
	}
	if len(w.doc.endnotes) > 0 {
		ct.AddOverride("/word/endnotes.xml", ooxml.CTEndnotes)
	}
	if len(w.comments) > 0 {
		ct.AddOverride("/word/comments.xml", ooxml.CTComments)
	}
	for _, ch := range w.charts {
		ct.AddOverride("/"+ch.Name, ooxml.CTChart)
	}
	b, _ := common.MarshalXML(ct)
	return b
}

func (w *word2007Writer) pkgRels() []byte {
	r := ooxml.NewRelationships()
	r.Add("rId1", ooxml.NSOfficeRelOfficeDoc, "word/document.xml", "")
	r.Add("rId2", ooxml.NSCorePropsRel, "docProps/core.xml", "")
	r.Add("rId3", ooxml.NSExtPropsRel, "docProps/app.xml", "")
	if len(w.doc.info.Custom) > 0 {
		r.Add("rId4", ooxml.NSCustomPropsRel, "docProps/custom.xml", "")
	}
	b, _ := common.MarshalXML(r)
	return b
}

func (w *word2007Writer) docRels() []byte {
	r := ooxml.NewRelationships()
	r.Rel = append(r.Rel, w.rels...)
	b, _ := common.MarshalXML(r)
	return b
}

func (w *word2007Writer) coreProps() []byte {
	info := w.doc.info
	fmtTime := func(t time.Time) *ooxml.W3Time {
		if t.IsZero() {
			t = time.Now().UTC()
		}
		return &ooxml.W3Time{XSIType: "dcterms:W3CDTF", Value: t.UTC().Format(time.RFC3339)}
	}
	cp := &ooxml.CoreProperties{
		XmlnsCP:        ooxml.NSCP,
		XmlnsDC:        ooxml.NSDC,
		XmlnsDCTERMS:   ooxml.NSDCTERMS,
		XmlnsDCMIType:  ooxml.NSDCMI,
		XmlnsXSI:       ooxml.NSXSI,
		Title:          info.Title,
		Subject:        info.Subject,
		Creator:        info.Creator,
		Keywords:       info.Keywords,
		Description:    info.Description,
		LastModifiedBy: info.LastModifiedBy,
		Category:       info.Category,
		Revision:       "1",
		Created:        fmtTime(info.Created),
		Modified:       fmtTime(info.Modified),
	}
	b, _ := common.MarshalXML(cp)
	return b
}

func (w *word2007Writer) appProps() []byte {
	ap := &ooxml.AppProperties{
		Xmlns:       ooxml.NSEPov,
		XmlnsVT:     ooxml.NSVT,
		Application: "GoWord",
		Company:     w.doc.info.Company,
		Manager:     w.doc.info.Manager,
		AppVersion:  "0.3",
	}
	b, _ := common.MarshalXML(ap)
	return b
}

func (w *word2007Writer) customProps() []byte {
	cp := &ooxml.CustomProperties{
		Xmlns:   ooxml.NSCust,
		XmlnsVT: ooxml.NSVT,
	}
	for i, p := range w.doc.info.Custom {
		item := ooxml.CustomProperty{
			FmtID: "{D5CDD505-2E9C-101B-9397-08002B2CF9AE}",
			PID:   i + 2,
			Name:  p.Name,
		}
		switch p.Type {
		case "int":
			item.I4 = p.Value
		case "float":
			item.R8 = p.Value
		case "bool":
			item.Bool = p.Value
		default:
			item.Lpwstr = p.Value
		}
		cp.Props = append(cp.Props, item)
	}
	b, _ := common.MarshalXML(cp)
	return b
}

func (w *word2007Writer) documentXML() []byte {
	xw := common.GetXMLWriter()
	w.writeDocument(xw)
	return common.FinishXML(xw)
}

func (w *word2007Writer) writeDocumentStart(xw *common.XMLWriter) {
	xw.StartDocument()
	xw.Start("w:document",
		"xmlns:wpc", "http://schemas.microsoft.com/office/word/2010/wordprocessingCanvas",
		"xmlns:mc", "http://schemas.openxmlformats.org/markup-compatibility/2006",
		"xmlns:o", ooxml.NSO,
		"xmlns:r", ooxml.NSR,
		"xmlns:m", ooxml.NSM,
		"xmlns:v", ooxml.NSV,
		"xmlns:wp", ooxml.NSWP,
		"xmlns:w10", ooxml.NSW10,
		"xmlns:w", ooxml.NSW,
		"xmlns:wne", ooxml.NSWNE,
		"xmlns:a", ooxml.NSA,
		"xmlns:pic", ooxml.NSPic,
	)
	xw.Start("w:body")
}

func (w *word2007Writer) writeDocumentEnd(xw *common.XMLWriter) {
	xw.End() // body
	xw.End() // document
}

func (w *word2007Writer) writeDocument(xw *common.XMLWriter) {
	w.writeDocumentStart(xw)
	secs := w.doc.sections
	for i, sec := range secs {
		w.writeContainer(xw, sec.Elements(), false)
		if i == len(secs)-1 {
			w.writeSectPr(xw, sec)
		} else {
			xw.Start("w:p")
			xw.Start("w:pPr")
			w.writeSectPr(xw, sec)
			xw.End()
			xw.End()
		}
	}
	w.writeDocumentEnd(xw)
	_ = xw.Flush()
}

func (w *word2007Writer) writeSectPr(xw *common.XMLWriter, sec *element.Section) {
	xw.Start("w:sectPr")
	for _, h := range w.headers {
		if hdr, ok := h.El.(*element.Header); ok && hdr.SectionID == sec.SectionID {
			typ := h.Type
			if typ == "" {
				typ = "default"
			}
			xw.Empty("w:headerReference", "w:type", typ, "r:id", h.RelID)
		}
	}
	for _, f := range w.footers {
		if ftr, ok := f.El.(*element.Footer); ok && ftr.SectionID == sec.SectionID {
			typ := f.Type
			if typ == "" {
				typ = "default"
			}
			xw.Empty("w:footerReference", "w:type", typ, "r:id", f.RelID)
		}
	}
	st := sec.Style
	orient := st.Orientation
	wAttr := itoa(st.PageSizeW)
	hAttr := itoa(st.PageSizeH)
	if orient == style.OrientationLandscape {
		xw.Empty("w:pgSz", "w:w", wAttr, "w:h", hAttr, "w:orient", "landscape")
	} else {
		xw.Empty("w:pgSz", "w:w", wAttr, "w:h", hAttr)
	}
	xw.Empty("w:pgMar",
		"w:top", itoa(st.MarginTop),
		"w:right", itoa(st.MarginRight),
		"w:bottom", itoa(st.MarginBottom),
		"w:left", itoa(st.MarginLeft),
		"w:header", itoa(st.HeaderHeight),
		"w:footer", itoa(st.FooterHeight),
		"w:gutter", itoa(st.Gutter),
	)
	cols := st.ColsNum
	if cols < 1 {
		cols = 1
	}
	xw.Empty("w:cols", "w:space", itoa(st.ColsSpace), "w:num", itoa(cols))
	if st.PageNumberingStart > 0 {
		xw.Empty("w:pgNumType", "w:start", itoa(st.PageNumberingStart))
	}
	if st.BreakType != "" {
		xw.Empty("w:type", "w:val", st.BreakType)
	}
	xw.Empty("w:docGrid", "w:linePitch", "360")
	xw.End()
}

func (w *word2007Writer) hdrFtrXML(tag string, el element.Element) []byte {
	xw := common.GetXMLWriter()
	w.writeHdrFtr(xw, tag, el)
	return common.FinishXML(xw)
}

func (w *word2007Writer) writeHdrFtr(xw *common.XMLWriter, tag string, el element.Element) {
	xw.StartDocument()
	xw.Start(tag,
		"xmlns:w", ooxml.NSW,
		"xmlns:r", ooxml.NSR,
		"xmlns:wp", ooxml.NSWP,
		"xmlns:a", ooxml.NSA,
		"xmlns:pic", ooxml.NSPic,
		"xmlns:v", ooxml.NSV,
	)
	var kids []element.Element
	switch v := el.(type) {
	case *element.Header:
		kids = v.Elements()
	case *element.Footer:
		kids = v.Elements()
	}
	w.writeContainer(xw, kids, false)
	xw.End()
}

func (w *word2007Writer) notesXML(foot bool) []byte {
	xw := common.GetXMLWriter()
	w.writeNotesXML(xw, foot)
	return common.FinishXML(xw)
}

func (w *word2007Writer) writeNotesXML(xw *common.XMLWriter, foot bool) {
	xw.StartDocument()
	tag := "w:footnotes"
	if !foot {
		tag = "w:endnotes"
	}
	xw.Start(tag, "xmlns:w", ooxml.NSW, "xmlns:r", ooxml.NSR)
	sep := "footnote"
	if !foot {
		sep = "endnote"
	}
	xw.Start("w:"+sep, "w:id", "-1", "w:type", "separator")
	xw.Start("w:p")
	xw.Start("w:r")
	xw.Empty("w:separator")
	xw.End()
	xw.End()
	xw.End()
	xw.Start("w:"+sep, "w:id", "0", "w:type", "continuationSeparator")
	xw.Start("w:p")
	xw.Start("w:r")
	xw.Empty("w:continuationSeparator")
	xw.End()
	xw.End()
	xw.End()
	if foot {
		for i, n := range w.doc.footnotes {
			xw.Start("w:footnote", "w:id", itoa(i+1))
			w.writeContainer(xw, n.Elements(), false)
			xw.End()
		}
	} else {
		for i, n := range w.doc.endnotes {
			xw.Start("w:endnote", "w:id", itoa(i+1))
			w.writeContainer(xw, n.Elements(), false)
			xw.End()
		}
	}
	xw.End()
}

func (w *word2007Writer) settingsXML() []byte {
	xw := common.GetXMLWriter()
	w.writeSettingsXML(xw)
	return common.FinishXML(xw)
}

func (w *word2007Writer) writeSettingsXML(xw *common.XMLWriter) {
	xw.StartDocument()
	xw.Start("w:settings", "xmlns:w", ooxml.NSW, "xmlns:m", ooxml.NSM, "xmlns:o", ooxml.NSO, "xmlns:r", ooxml.NSR)
	s := w.doc.settings
	if s.ZoomPreset != "" {
		xw.Empty("w:zoom", "w:val", s.ZoomPreset)
	} else {
		xw.Empty("w:zoom", "w:percent", itoa(nonzero(s.Zoom, 100)))
	}
	if s.HideSpellingErrors {
		xw.Empty("w:hideSpellingErrors")
	}
	if s.HideGrammaticalErrors {
		xw.Empty("w:hideGrammaticalErrors")
	}
	if s.TrackRevisions {
		xw.Empty("w:trackRevisions")
	}
	if s.UpdateFields {
		xw.Empty("w:updateFields")
	}
	if s.MirrorMargins {
		xw.Empty("w:mirrorMargins")
	}
	if s.EvenAndOddHeaders {
		xw.Empty("w:evenAndOddHeaders")
	}
	if s.ThemeFontLang != "" {
		attrs := []string{"w:val", s.ThemeFontLang}
		if s.ThemeFont != nil {
			if s.ThemeFont.EastAsia != "" {
				attrs = append(attrs, "w:eastAsia", s.ThemeFont.EastAsia)
			}
			if s.ThemeFont.Bidirectional != "" {
				attrs = append(attrs, "w:bidi", s.ThemeFont.Bidirectional)
			}
		}
		xw.Empty("w:themeFontLang", attrs...)
	} else if s.ThemeFont != nil {
		attrs := []string{}
		if s.ThemeFont.Latin != "" {
			attrs = append(attrs, "w:val", s.ThemeFont.Latin)
		}
		if s.ThemeFont.EastAsia != "" {
			attrs = append(attrs, "w:eastAsia", s.ThemeFont.EastAsia)
		}
		if s.ThemeFont.Bidirectional != "" {
			attrs = append(attrs, "w:bidi", s.ThemeFont.Bidirectional)
		}
		if len(attrs) > 0 {
			xw.Empty("w:themeFontLang", attrs...)
		}
	}
	if s.ProofState.Spelling != "" || s.ProofState.Grammar != "" {
		attrs := []string{}
		if s.ProofState.Spelling != "" {
			attrs = append(attrs, "w:spelling", s.ProofState.Spelling)
		}
		if s.ProofState.Grammar != "" {
			attrs = append(attrs, "w:grammar", s.ProofState.Grammar)
		}
		xw.Empty("w:proofState", attrs...)
	}
	if s.DoNotTrackMoves {
		xw.Empty("w:doNotTrackMoves")
	}
	if s.DoNotTrackFormatting {
		xw.Empty("w:doNotTrackFormatting")
	}
	if s.AutoHyphenation {
		xw.Empty("w:autoHyphenation")
	}
	if s.ConsecutiveHyphenLimit > 0 {
		xw.Empty("w:consecutiveHyphenLimit", "w:val", itoa(s.ConsecutiveHyphenLimit))
	}
	if s.HyphenationZone > 0 {
		xw.Empty("w:hyphenationZone", "w:val", itoa(s.HyphenationZone))
	}
	if s.DoNotHyphenateCaps {
		xw.Empty("w:doNotHyphenateCaps")
	}
	if s.BookFoldPrinting {
		xw.Empty("w:bookFoldPrinting")
	}
	if s.DecimalSymbol != "" && s.DecimalSymbol != "." {
		xw.Empty("w:decimalSymbol", "w:val", s.DecimalSymbol)
	}
	if rv := s.RevisionView; rv != nil {
		attrs := []string{}
		if rv.Markup != nil {
			attrs = append(attrs, "w:markup", bool01(*rv.Markup))
		}
		if rv.Comments != nil {
			attrs = append(attrs, "w:comments", bool01(*rv.Comments))
		}
		if rv.InsDel != nil {
			attrs = append(attrs, "w:insDel", bool01(*rv.InsDel))
		}
		if rv.Formatting != nil {
			attrs = append(attrs, "w:formatting", bool01(*rv.Formatting))
		}
		if rv.InkAnnotations != nil {
			attrs = append(attrs, "w:inkAnnotations", bool01(*rv.InkAnnotations))
		}
		if len(attrs) > 0 {
			xw.Empty("w:revisionView", attrs...)
		}
	}
	if p := s.DocumentProtection; p != nil && p.Editing != "" {
		attrs := []string{"w:edit", p.Editing, "w:enforcement", "1"}
		if p.Hash != "" {
			attrs = append(attrs, "w:hash", p.Hash)
		}
		xw.Empty("w:documentProtection", attrs...)
	}
	ver := w.doc.compatibility.OOXMLVersion
	if ver == 0 {
		ver = 12
	}
	xw.Start("w:compat")
	xw.Empty("w:compatSetting",
		"w:name", "compatibilityMode",
		"w:uri", "http://schemas.microsoft.com/office/word",
		"w:val", itoa(ver))
	xw.End()
	xw.Empty("w:clrSchemeMapping",
		"w:bg1", "light1", "w:t1", "dark1", "w:bg2", "light2", "w:t2", "dark2",
		"w:accent1", "accent1", "w:accent2", "accent2", "w:accent3", "accent3",
		"w:accent4", "accent4", "w:accent5", "accent5", "w:accent6", "accent6",
		"w:hyperlink", "hyperlink", "w:followedHyperlink", "followedHyperlink")
	xw.End()
}

func (w *word2007Writer) stylesXML() []byte {
	xw := common.GetXMLWriter()
	w.writeStylesXML(xw)
	return common.FinishXML(xw)
}

func (w *word2007Writer) writeStylesXML(xw *common.XMLWriter) {
	xw.StartDocument()
	xw.Start("w:styles", "xmlns:w", ooxml.NSW, "xmlns:r", ooxml.NSR)
	xw.Start("w:docDefaults")
	xw.Start("w:rPrDefault")
	xw.Start("w:rPr")
	xw.Empty("w:rFonts",
		"w:ascii", w.doc.defaultFontName,
		"w:hAnsi", w.doc.defaultFontName,
		"w:eastAsia", w.doc.defaultAsianFont,
		"w:cs", w.doc.defaultFontName)
	xw.Empty("w:sz", "w:val", itoa(int(w.doc.defaultFontSize*2)))
	xw.Empty("w:szCs", "w:val", itoa(int(w.doc.defaultFontSize*2)))
	xw.Empty("w:color", "w:val", w.doc.defaultFontColor)
	xw.Empty("w:lang", "w:val", "en-US", "w:eastAsia", "en-US", "w:bidi", "ar-SA")
	xw.End()
	xw.End()
	xw.Start("w:pPrDefault")
	if w.doc.defaultParagraph != nil {
		w.writePPr(xw, *w.doc.defaultParagraph, "")
	}
	xw.End()
	xw.End() // docDefaults

	w.writeStyleDef(xw, "Normal", "paragraph", "Normal", true, nil, w.doc.defaultParagraph)
	w.writeStyleDef(xw, "Hyperlink", "character", "Hyperlink", false, &style.Font{Color: "0563C1", Underline: style.UnderlineSingle}, nil)

	for _, ns := range w.doc.styles {
		switch ns.Kind {
		case "font", "link":
			w.writeStyleDef(xw, ns.Name, "character", ns.Name, false, ns.Font, ns.Paragraph)
		case "paragraph":
			w.writeStyleDef(xw, ns.Name, "paragraph", ns.Name, false, ns.Font, ns.Paragraph)
		case "title":
			ui := ns.Name
			if ns.Depth > 0 {
				ui = "heading " + itoa(ns.Depth)
			}
			w.writeHeadingStyle(xw, ns, ui)
		case "table":
			w.writeTableStyle(xw, ns)
		}
	}
	xw.End()
}

func (w *word2007Writer) writeStyleDef(xw *common.XMLWriter, id, typ, name string, def bool, font *style.Font, para *style.Paragraph) {
	attrs := []string{"w:type", typ, "w:styleId", id}
	if def {
		attrs = append(attrs, "w:default", "1")
	}
	xw.Start("w:style", attrs...)
	xw.Empty("w:name", "w:val", name)
	if !def && typ != "character" {
		xw.Empty("w:basedOn", "w:val", "Normal")
	}
	if qn := qFormatName(typ); qn {
		xw.Empty("w:qFormat")
	}
	if para != nil {
		w.writePPr(xw, *para, "")
	}
	if font != nil {
		w.writeRPr(xw, *font, "")
	}
	xw.End()
}

func qFormatName(typ string) bool { return typ == "paragraph" || typ == "character" }

func (w *word2007Writer) writeHeadingStyle(xw *common.XMLWriter, ns namedStyle, uiName string) {
	xw.Start("w:style", "w:type", "paragraph", "w:styleId", ns.Name)
	xw.Empty("w:name", "w:val", uiName)
	xw.Empty("w:basedOn", "w:val", "Normal")
	xw.Empty("w:next", "w:val", "Normal")
	xw.Empty("w:qFormat")
	p := style.Paragraph{}
	if ns.Paragraph != nil {
		p = *ns.Paragraph
	}
	if p.OutlineLevel == 0 && ns.Depth > 0 {
		p.OutlineLevel = ns.Depth
	}
	w.writePPr(xw, p, "")
	if ns.Font != nil {
		w.writeRPr(xw, *ns.Font, "")
	}
	xw.End()
}

func (w *word2007Writer) writeTableStyle(xw *common.XMLWriter, ns namedStyle) {
	xw.Start("w:style", "w:type", "table", "w:styleId", ns.Name)
	xw.Empty("w:name", "w:val", ns.Name)
	xw.Empty("w:basedOn", "w:val", "TableNormal")
	xw.Empty("w:qFormat")
	if ns.Table != nil {
		w.writeTblPr(xw, *ns.Table)
	}
	xw.End()
}

func (w *word2007Writer) numberingXML() []byte {
	xw := common.GetXMLWriter()
	w.writeNumberingXML(xw)
	return common.FinishXML(xw)
}

func (w *word2007Writer) writeNumberingXML(xw *common.XMLWriter) {
	xw.StartDocument()
	xw.Start("w:numbering", "xmlns:w", ooxml.NSW)
	// abstract 1: bullets
	xw.Start("w:abstractNum", "w:abstractNumId", "1")
	xw.Empty("w:multiLevelType", "w:val", "hybridMultilevel")
	for i := 0; i < 9; i++ {
		xw.Start("w:lvl", "w:ilvl", itoa(i))
		xw.Empty("w:start", "w:val", "1")
		xw.Empty("w:numFmt", "w:val", "bullet")
		xw.Empty("w:lvlText", "w:val", "•")
		xw.Empty("w:lvlJc", "w:val", "left")
		xw.Start("w:pPr")
		xw.Empty("w:ind", "w:left", itoa(720*(i+1)), "w:hanging", "360")
		xw.End()
		xw.End()
	}
	xw.End()
	// abstract 2: decimal
	xw.Start("w:abstractNum", "w:abstractNumId", "2")
	xw.Empty("w:multiLevelType", "w:val", "hybridMultilevel")
	for i := 0; i < 9; i++ {
		xw.Start("w:lvl", "w:ilvl", itoa(i))
		xw.Empty("w:start", "w:val", "1")
		xw.Empty("w:numFmt", "w:val", "decimal")
		xw.Empty("w:lvlText", "w:val", "%"+itoa(i+1)+".")
		xw.Empty("w:lvlJc", "w:val", "left")
		xw.Start("w:pPr")
		xw.Empty("w:ind", "w:left", itoa(720*(i+1)), "w:hanging", "360")
		xw.End()
		xw.End()
	}
	xw.End()

	nextAbs := 3
	numID := 3
	for _, ns := range w.doc.styles {
		if ns.Kind != "numbering" || ns.Numbering == nil {
			continue
		}
		abs := nextAbs
		nextAbs++
		xw.Start("w:abstractNum", "w:abstractNumId", itoa(abs))
		typ := ns.Numbering.Type
		if typ == "" {
			typ = "hybridMultilevel"
		}
		xw.Empty("w:multiLevelType", "w:val", typ)
		xw.Empty("w:name", "w:val", ns.Name)
		for i, lvl := range ns.Numbering.Levels {
			xw.Start("w:lvl", "w:ilvl", itoa(i))
			start := lvl.Start
			if start == 0 {
				start = 1
			}
			xw.Empty("w:start", "w:val", itoa(start))
			fmt := lvl.Format
			if fmt == "" {
				fmt = "decimal"
			}
			xw.Empty("w:numFmt", "w:val", fmt)
			txt := lvl.Text
			if txt == "" {
				txt = "%" + itoa(i+1) + "."
			}
			xw.Empty("w:lvlText", "w:val", txt)
			al := lvl.Alignment
			if al == "" {
				al = "left"
			}
			xw.Empty("w:lvlJc", "w:val", al)
			xw.End()
		}
		xw.End()
		xw.Start("w:num", "w:numId", itoa(numID))
		xw.Empty("w:abstractNumId", "w:val", itoa(abs))
		xw.End()
		numID++
	}

	xw.Start("w:num", "w:numId", "1")
	xw.Empty("w:abstractNumId", "w:val", "1")
	xw.End()
	xw.Start("w:num", "w:numId", "2")
	xw.Empty("w:abstractNumId", "w:val", "2")
	xw.End()
	xw.End()
}

func itoa(n int) string { return strconv.Itoa(n) }

func bool01(v bool) string {
	if v {
		return "1"
	}
	return "0"
}

func nonzero(n, def int) int {
	if n == 0 {
		return def
	}
	return n
}

func (w *word2007Writer) commentsXML() []byte {
	xw := common.GetXMLWriter()
	w.writeCommentsXML(xw)
	return common.FinishXML(xw)
}

func (w *word2007Writer) writeCommentsXML(xw *common.XMLWriter) {
	xw.StartDocument()
	xw.Start("w:comments",
		"xmlns:w", ooxml.NSW,
		"xmlns:r", ooxml.NSR,
		"xmlns:o", ooxml.NSO,
		"xmlns:v", ooxml.NSV,
		"xmlns:wp", ooxml.NSWP,
		"xmlns:m", ooxml.NSM)
	for _, c := range w.comments {
		attrs := []string{"w:id", itoa(c.CommentID), "w:author", c.Author}
		if c.Date != "" {
			attrs = append(attrs, "w:date", c.Date)
		}
		if c.Initials != "" {
			attrs = append(attrs, "w:initials", c.Initials)
		}
		xw.Start("w:comment", attrs...)
		if len(c.Elements()) == 0 {
			xw.Start("w:p")
			xw.End()
		} else {
			w.writeContainer(xw, c.Elements(), false)
		}
		xw.End()
	}
	xw.End()
}
