package word

import (
	"bytes"
	"encoding/xml"
	"errors"
	"io"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/yunkeweb/go-word/element"
	"github.com/yunkeweb/go-word/metadata"
	"github.com/yunkeweb/go-word/pkg/common"
	"github.com/yunkeweb/go-word/style"
)

type word2007Reader struct{}

type documentRelationship struct {
	Target string
	Type   string
}

type documentImage struct {
	Name   string
	Target string
	Data   []byte
}

type drawingInfo struct {
	rid, geom, fill, line, text, alt string
	width, height, lineWidth         int64
	rotation                         float64
	flipH, flipV, textBox            bool
	bold                             bool
}

func (word2007Reader) Load(filename string) (*Document, error) {
	zr, err := common.OpenZipFile(filename)
	if err != nil {
		return nil, err
	}
	defer zr.Close()
	return loadWord2007(zr, ReadOptions{})
}

func LoadBytes(data []byte) (*Document, error) {
	return LoadBytesWithOptions(data, ReadOptions{})
}

// LoadBytesWithOptions loads a document with optional ZIP read budgets.
func LoadBytesWithOptions(data []byte, opts ReadOptions) (*Document, error) {
	if err := opts.validate(); err != nil {
		return nil, err
	}
	if err := validateZipLimits(nil, int64(len(data)), opts); err != nil {
		return nil, err
	}
	zr, err := common.OpenZipBytes(data)
	if err != nil {
		return nil, err
	}
	return loadWord2007(zr, opts)
}

func loadWord2007(zr *common.ZipReader, opts ReadOptions) (*Document, error) {
	if err := validateZipLimits(zr.Files(), -1, opts); err != nil {
		return nil, err
	}
	doc := New()
	raw, err := zr.ReadFileWithLimit("word/document.xml", opts.partLimit())
	if err != nil {
		return nil, err
	}
	rels := map[string]documentRelationship{}
	if relRaw, err := zr.ReadFileWithLimit("word/_rels/document.xml.rels", opts.partLimit()); err == nil {
		rels = parseRelationshipDetails(relRaw)
	} else if !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}
	listStyles := map[int]map[int]style.ListItem{}
	if numbering, e := zr.ReadFileWithLimit("word/numbering.xml", opts.partLimit()); e == nil {
		listStyles = parseNumberingPart(numbering)
	} else if !errors.Is(e, os.ErrNotExist) {
		return nil, e
	}
	var styles *wordStyleSheet
	if stylesRaw, e := zr.ReadFileWithLimit("word/styles.xml", opts.partLimit()); e == nil {
		styles, err = parseWordStyles(stylesRaw)
		if err != nil {
			return nil, err
		}
	} else if !errors.Is(e, os.ErrNotExist) {
		return nil, e
	}
	if styles != nil {
		for _, rel := range rels {
			if !strings.HasSuffix(rel.Type, "/theme") || strings.Contains(rel.Target, "://") {
				continue
			}
			target := path.Clean(path.Join("word", rel.Target))
			themeRaw, e := zr.ReadFileWithLimit(target, opts.partLimit())
			if e == nil {
				styles.theme, e = parseWordTheme(themeRaw)
				if e != nil {
					return nil, e
				}
			} else if !errors.Is(e, os.ErrNotExist) {
				return nil, e
			}
			break
		}
	}
	images := map[string]documentImage{}
	charts := map[string]*element.Chart{}
	for id, rel := range rels {
		if !strings.HasSuffix(rel.Type, "/image") || strings.Contains(rel.Target, "://") {
			continue
		}
		target := path.Clean(path.Join("word", rel.Target))
		data, e := zr.ReadFileWithLimit(target, opts.partLimit())
		if e != nil {
			if errors.Is(e, os.ErrNotExist) {
				continue
			}
			return nil, e
		}
		images[id] = documentImage{Name: path.Base(target), Target: target, Data: data}
	}
	for id, rel := range rels {
		if !strings.HasSuffix(rel.Type, "/chart") || strings.Contains(rel.Target, "://") {
			continue
		}
		target := path.Clean(path.Join("word", rel.Target))
		data, e := zr.ReadFileWithLimit(target, opts.partLimit())
		if e != nil {
			if errors.Is(e, os.ErrNotExist) {
				continue
			}
			return nil, e
		}
		charts[id] = parseChartPart(data)
		charts[id].DocPart = target
	}
	sections, err := parseDocumentSections(raw)
	if err != nil {
		return nil, err
	}
	if len(sections) == 0 {
		sections = append(sections, parsedSection{})
	}
	for _, parsed := range sections {
		sec := doc.AddSection(parsed.style)
		if err := parseDocumentXMLRelsImagesWithStylesAndCharts(parsed.body, sec, relationshipTargets(rels), listStyles, images, charts, styles); err != nil {
			return nil, err
		}
		if err := loadSectionParts(zr, opts, sec, parsed.sectPr, rels); err != nil {
			return nil, err
		}
	}
	if err := loadNoteParts(zr, opts, doc, rels); err != nil {
		return nil, err
	}
	commentsByID := make(map[int]*element.Comment, len(doc.comments))
	for _, comment := range doc.comments {
		if comment != nil && comment.CommentID >= 0 && commentsByID[comment.CommentID] == nil {
			commentsByID[comment.CommentID] = comment
		}
	}
	walkDocument(doc, func(el element.Element) {
		if ranged, ok := el.(interface {
			GetCommentRangeStart() *element.Comment
			SetCommentRangeStart(*element.Comment)
		}); ok {
			if marker := ranged.GetCommentRangeStart(); marker != nil {
				if comment := commentsByID[marker.CommentID]; comment != nil {
					ranged.SetCommentRangeStart(comment)
				}
			}
		}
		if ranged, ok := el.(interface {
			GetCommentRangeEnd() *element.Comment
			SetCommentRangeEnd(*element.Comment)
		}); ok {
			if marker := ranged.GetCommentRangeEnd(); marker != nil {
				if comment := commentsByID[marker.CommentID]; comment != nil {
					ranged.SetCommentRangeEnd(comment)
				}
			}
		}
	})
	if core, err := zr.ReadFileWithLimit("docProps/core.xml", opts.partLimit()); err == nil {
		parseCoreProperties(core, doc.info)
	} else if !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}
	doc.imagesLoaded = true
	for _, f := range zr.Files() {
		name := strings.ReplaceAll(f.Name, "\\", "/")
		lower := strings.ToLower(name)
		if !strings.HasPrefix(lower, "word/media/") || strings.HasSuffix(name, "/") {
			continue
		}
		data, err := zr.ReadFileWithLimit(f.Name, opts.partLimit())
		if err != nil {
			return nil, err
		}
		base := name
		if i := strings.LastIndex(base, "/"); i >= 0 {
			base = base[i+1:]
		}
		ext := ""
		if i := strings.LastIndex(base, "."); i >= 0 {
			ext = base[i+1:]
		}
		doc.extractedImages = append(doc.extractedImages, ImageFile{
			Name: base,
			MIME: mimeForExt(ext),
			Data: data,
		})
	}
	return doc, nil
}

func loadNoteParts(zr *common.ZipReader, opts ReadOptions, doc *Document, rels map[string]documentRelationship) error {
	for _, rel := range rels {
		var dst string
		switch {
		case strings.HasSuffix(rel.Type, "/footnotes"):
			dst = "footnotes"
		case strings.HasSuffix(rel.Type, "/endnotes"):
			dst = "endnotes"
		case strings.HasSuffix(rel.Type, "/comments"):
			dst = "comments"
		default:
			continue
		}
		target := path.Clean(path.Join("word", rel.Target))
		raw, err := zr.ReadFileWithLimit(target, opts.partLimit())
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			return err
		}
		notes := parseNotePart(raw, dst)
		switch dst {
		case "footnotes":
			doc.footnotes = append(doc.footnotes, notes.footnotes...)
		case "endnotes":
			doc.endnotes = append(doc.endnotes, notes.endnotes...)
		case "comments":
			doc.comments = append(doc.comments, notes.comments...)
		}
	}
	return nil
}

type parsedNotes struct {
	footnotes []*element.Footnote
	endnotes  []*element.Endnote
	comments  []*element.Comment
}

func parseNotePart(data []byte, kind string) parsedNotes {
	var out parsedNotes
	dec := xml.NewDecoder(bytes.NewReader(data))
	var noteID int
	var author, initials, date string
	var text strings.Builder
	var active bool
	flush := func() {
		value := text.String()
		text.Reset()
		if value == "" {
			return
		}
		switch kind {
		case "footnotes":
			n := element.NewFootnote(nil)
			n.NoteID = noteID
			n.AddText(value)
			out.footnotes = append(out.footnotes, n)
		case "endnotes":
			n := element.NewEndnote(nil)
			n.NoteID = noteID
			n.AddText(value)
			out.endnotes = append(out.endnotes, n)
		case "comments":
			c := &element.Comment{Author: author, Initials: initials, Date: date, CommentID: noteID}
			c.Kind = "Comment"
			c.AddText(value)
			out.comments = append(out.comments, c)
		}
	}
	for {
		tok, err := dec.Token()
		if err != nil {
			break
		}
		switch t := tok.(type) {
		case xml.StartElement:
			switch localName(t.Name) {
			case "footnote", "endnote", "comment":
				noteID = atoi(attr(t, "id"))
				active = noteID >= 0
				author, initials, date = attr(t, "author"), attr(t, "initials"), attr(t, "date")
			case "t", "delText":
				var s string
				if dec.DecodeElement(&s, &t) == nil && active {
					text.WriteString(s)
				}
			}
		case xml.EndElement:
			switch localName(t.Name) {
			case "footnote", "endnote", "comment":
				if active {
					flush()
				}
				active = false
			}
		}
	}
	return out
}

func parseSimpleField(dec *xml.Decoder, start xml.StartElement) (*element.Field, error) {
	instruction := strings.TrimSpace(attr(start, "instr"))
	var result strings.Builder
	depth := 1
	for depth > 0 {
		tok, err := dec.Token()
		if err != nil {
			return nil, err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			depth++
			if localName(t.Name) == "t" || localName(t.Name) == "delText" {
				var s string
				if err := dec.DecodeElement(&s, &t); err != nil {
					return nil, err
				}
				result.WriteString(s)
				depth--
			}
		case xml.EndElement:
			depth--
		}
	}
	fieldType := fieldTypeFromInstruction(instruction)
	f := element.NewField(fieldType, nil, nil, result.String())
	f.Instruction = instruction
	f.HasResult = result.Len() > 0
	return f, nil
}

func fieldTypeFromInstruction(instruction string) string {
	parts := strings.Fields(instruction)
	if len(parts) == 0 {
		return ""
	}
	return strings.ToUpper(parts[0])
}

func parseSDT(dec *xml.Decoder, start xml.StartElement) (*element.SDT, error) {
	s := &element.SDT{SDTType: element.SDTTypeRichText}
	s.Kind = "SDT"
	depth := 1
	inContent := false
	var text strings.Builder
	var contentRun *element.TextRun
	var runText strings.Builder
	var runFont style.Font
	knownType := false
	flushRun := func() {
		if runText.Len() == 0 {
			return
		}
		if contentRun == nil {
			contentRun = s.AddTextRun()
		}
		font := any(nil)
		if !runFont.IsZero() {
			font = runFont
		}
		contentRun.AddText(runText.String(), font)
		text.WriteString(runText.String())
		runText.Reset()
	}
	for depth > 0 {
		tok, err := dec.Token()
		if err != nil {
			return nil, err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			local := localName(t.Name)
			depth++
			switch local {
			case "sdtPr":
				if err := parseSDTProperties(dec, s, &knownType); err != nil {
					return nil, err
				}
				depth--
			case "sdtContent":
				inContent = true
			case "p":
				if inContent {
					s.IsBlock = true
					flushRun()
					contentRun = nil
				}
			case "r":
				if inContent {
					flushRun()
					runFont = style.Font{}
				}
			case "rPr":
				if inContent {
					var properties wordRunProperties
					if err := dec.DecodeElement(&properties, &t); err != nil {
						return nil, err
					}
					properties.apply(&runFont, false, nil)
					depth--
				}
			case "t", "delText":
				var value string
				if err := dec.DecodeElement(&value, &t); err != nil {
					return nil, err
				}
				if inContent {
					runText.WriteString(value)
				}
				depth--
			case "br", "cr":
				if inContent {
					flushRun()
					if contentRun == nil {
						contentRun = s.AddTextRun()
					}
					contentRun.AddTextBreak()
					text.WriteByte('\n')
				}
			}
		case xml.EndElement:
			local := localName(t.Name)
			if local == "sdtContent" {
				flushRun()
				inContent = false
			}
			if local == "p" && inContent {
				flushRun()
				contentRun = nil
			}
			if local == "r" && inContent {
				flushRun()
				runFont = style.Font{}
			}
			depth--
		}
	}
	s.Value = text.String()
	if !knownType {
		s.Diagnostic = "unknown content control type; preserved visible content as text"
	}
	// The content runs above are the source of truth. Keep a plain fallback only
	// for controls whose content contains no supported run element.
	if s.Value != "" && len(s.Elements()) == 0 {
		if s.IsBlock {
			s.AddText(s.Value)
		} else {
			s.AddTextRun().AddText(s.Value)
		}
	}
	if s.SDTType == element.SDTTypeCheckbox {
		if s.Value == "" {
			s.SetChecked(s.Checked)
		}
	} else if s.Value == "" && s.ShowingPlcHdr {
		s.Placeholder = sdtPlaceholder(s.SDTType)
		s.Value = s.Placeholder
	} else if s.Value != "" && !s.ShowingPlcHdr {
		s.ShowingPlcHdr = false
	}
	return s, nil
}

func sdtPlaceholder(typ string) string {
	switch typ {
	case element.SDTTypeDate:
		return "Click or tap to enter a date."
	case element.SDTTypeDropDown, element.SDTTypeComboBox:
		return "Choose an item."
	default:
		return "Click or tap to enter text."
	}
}

func parseSDTProperties(dec *xml.Decoder, s *element.SDT, knownType *bool) error {
	depth := 1
	for depth > 0 {
		tok, err := dec.Token()
		if err != nil {
			return err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			local := localName(t.Name)
			depth++
			switch local {
			case "alias":
				s.Alias = attr(t, "val")
			case "tag":
				s.Tag = attr(t, "val")
			case "id":
				s.ID = atoi(attr(t, "val"))
			case "text":
				s.SDTType = element.SDTTypePlainText
				*knownType = true
				s.MultiLine = onOffTrue(t)
			case "richText":
				s.SDTType = element.SDTTypeRichText
				*knownType = true
			case "comboBox":
				s.SDTType = element.SDTTypeComboBox
				*knownType = true
			case "dropDownList":
				s.SDTType = element.SDTTypeDropDown
				*knownType = true
			case "date":
				s.SDTType = element.SDTTypeDate
				*knownType = true
			case "dateFormat":
				s.DateFormat = attr(t, "val")
			case "lock":
				s.Lock = attr(t, "val")
			case "placeholder":
				// The document-part metadata is separate from currently showing the placeholder.
			case "showingPlcHdr":
				s.ShowingPlcHdr = true
			case "docPart":
				// Metadata-only child; its value is not needed for read-only HTML.
			case "checkbox":
				s.SDTType = element.SDTTypeCheckbox
				*knownType = true
				s.Checked = false
			case "checked":
				s.Checked = onOffTrue(t)
			case "listItem":
				s.ListItems = append(s.ListItems, element.SDTListItem{Value: attr(t, "value"), DisplayText: attr(t, "displayText")})
			}
		case xml.EndElement:
			depth--
		}
	}
	return nil
}

func parseNumberingPart(data []byte) map[int]map[int]style.ListItem {
	abs := map[int]map[int]style.ListItem{}
	nums := map[int]int{}
	dec := xml.NewDecoder(bytes.NewReader(data))
	currentAbs, currentNum, currentLevel := -1, -1, -1
	for {
		tok, err := dec.Token()
		if err != nil {
			break
		}
		switch t := tok.(type) {
		case xml.StartElement:
			switch localName(t.Name) {
			case "abstractNum":
				currentAbs = atoi(attr(t, "abstractNumId"))
				abs[currentAbs] = make(map[int]style.ListItem)
			case "lvl":
				if currentAbs >= 0 {
					currentLevel = atoi(attr(t, "ilvl"))
				}
			case "start":
				if currentAbs >= 0 && currentLevel >= 0 {
					level := abs[currentAbs][currentLevel]
					level.Start = atoi(attr(t, "val"))
					abs[currentAbs][currentLevel] = level
				}
			case "numFmt":
				if currentAbs >= 0 && currentLevel >= 0 {
					level := abs[currentAbs][currentLevel]
					level.Format = attr(t, "val")
					abs[currentAbs][currentLevel] = level
				}
			case "num":
				currentNum = atoi(attr(t, "numId"))
			case "abstractNumId":
				if currentNum >= 0 {
					nums[currentNum] = atoi(attr(t, "val"))
				}
			}
		case xml.EndElement:
			switch localName(t.Name) {
			case "lvl":
				currentLevel = -1
			case "abstractNum":
				currentAbs = -1
			case "num":
				currentNum = -1
			}
		}
	}
	out := map[int]map[int]style.ListItem{}
	for id, abstractID := range nums {
		out[id] = make(map[int]style.ListItem)
		for depth, level := range abs[abstractID] {
			level.NumId = id
			level.Depth = depth
			level.ListType = style.ListTypeNumber
			if level.Format == style.NumberBullet {
				level.ListType = style.ListTypeBullet
			}
			out[id][depth] = level
		}
	}
	return out
}

func parseDocumentXML(data []byte, sec *element.Section) error {
	return parseDocumentXMLRels(data, sec, nil, nil)
}

type parsedSection struct {
	body   []byte
	sectPr []byte
	style  style.Section
}

func parseDocumentSections(data []byte) ([]parsedSection, error) {
	dec := xml.NewDecoder(bytes.NewReader(data))
	var body bytes.Buffer
	var sections []parsedSection
	depth := 0
	skipEnds := 0
	inBody := false
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			local := localName(t.Name)
			if local == "body" {
				inBody = true
				continue
			}
			if !inBody {
				continue
			}
			if local == "sectPr" {
				// A non-final section stores sectPr inside a dedicated paragraph.
				// Do not expose that structural paragraph as document content.
				lastP := bytes.LastIndex(body.Bytes(), []byte("<p"))
				if lastP >= 0 {
					tail := body.Bytes()[lastP:]
					if bytes.Contains(tail, []byte("<pPr")) && !bytes.Contains(tail, []byte("<r")) {
						i := bytes.LastIndex(body.Bytes()[:lastP], []byte("<p"))
						if i < 0 {
							i = lastP
						}
						body.Truncate(i)
						skipEnds = 2
						depth = 0
					}
				}
				var raw struct {
					XML string `xml:",innerxml"`
				}
				if err := dec.DecodeElement(&raw, &t); err != nil {
					return nil, err
				}
				pr := []byte("<w:sectPr xmlns:w=\"http://schemas.openxmlformats.org/wordprocessingml/2006/main\">")
				pr = append(pr, []byte(raw.XML)...)
				pr = append(pr, []byte("</w:sectPr>")...)
				st := style.NewSection()
				parseSectPr(pr, &st)
				sections = append(sections, parsedSection{body: append([]byte(nil), body.Bytes()...), sectPr: pr, style: st})
				body.Reset()
				continue
			}
			depth++
			if err := appendXMLToken(&body, t); err != nil {
				return nil, err
			}
		case xml.EndElement:
			if !inBody {
				continue
			}
			if localName(t.Name) == "body" {
				inBody = false
				continue
			}
			if skipEnds > 0 {
				skipEnds--
				continue
			}
			if depth > 0 {
				body.WriteString("</")
				body.WriteString(t.Name.Local)
				body.WriteString(">")
				depth--
			}
		case xml.CharData:
			if inBody {
				if err := xml.EscapeText(&body, t); err != nil {
					return nil, err
				}
			}
		case xml.Comment:
			if inBody {
				body.WriteString("<!--")
				body.Write(t)
				body.WriteString("-->")
			}
		}
	}
	if len(sections) == 0 || body.Len() > 0 {
		st := style.NewSection()
		sections = append(sections, parsedSection{body: body.Bytes(), style: st})
	}
	return sections, nil
}

func appendXMLToken(dst *bytes.Buffer, tok xml.Token) error {
	var buf bytes.Buffer
	enc := xml.NewEncoder(&buf)
	if err := enc.EncodeToken(tok); err != nil {
		return err
	}
	if err := enc.Flush(); err != nil {
		return err
	}
	dst.Write(buf.Bytes())
	return nil
}

func escapeXMLAttr(s string) string {
	var b strings.Builder
	_ = xml.EscapeText(&b, []byte(s))
	return b.String()
}

func parseSectPr(data []byte, st *style.Section) {
	dec := xml.NewDecoder(bytes.NewReader(data))
	for {
		tok, err := dec.Token()
		if err != nil {
			return
		}
		se, ok := tok.(xml.StartElement)
		if !ok {
			continue
		}
		switch localName(se.Name) {
		case "type":
			st.BreakType = attr(se, "val")
		case "pgSz":
			st.PageSizeW, st.PageSizeH = atoi(attr(se, "w")), atoi(attr(se, "h"))
			if attr(se, "orient") == "landscape" {
				st.Orientation = style.OrientationLandscape
			}
		case "pgMar":
			st.MarginTop, st.MarginRight, st.MarginBottom, st.MarginLeft = atoi(attr(se, "top")), atoi(attr(se, "right")), atoi(attr(se, "bottom")), atoi(attr(se, "left"))
			st.HeaderHeight, st.FooterHeight, st.Gutter = atoi(attr(se, "header")), atoi(attr(se, "footer")), atoi(attr(se, "gutter"))
		case "pgNumType":
			st.PageNumberingStart = atoi(attr(se, "start"))
		case "cols":
			st.ColsNum, st.ColsSpace = atoi(attr(se, "num")), atoi(attr(se, "space"))
			st.ColsSeparator = attr(se, "sep") == "1" || attr(se, "sep") == "true"
		}
	}
}

func relationshipTargets(in map[string]documentRelationship) map[string]string {
	out := map[string]string{}
	for id, r := range in {
		out[id] = r.Target
	}
	return out
}

func loadSectionParts(zr *common.ZipReader, opts ReadOptions, sec *element.Section, sectPr []byte, rels map[string]documentRelationship) error {
	dec := xml.NewDecoder(bytes.NewReader(sectPr))
	for {
		tok, err := dec.Token()
		if err != nil {
			break
		}
		se, ok := tok.(xml.StartElement)
		if !ok {
			continue
		}
		local := localName(se.Name)
		if local != "headerReference" && local != "footerReference" {
			continue
		}
		typ := attr(se, "type")
		id := attr(se, "id")
		if id == "" {
			continue
		}
		r, ok := rels[id]
		if !ok {
			continue
		}
		target := path.Clean(path.Join("word", r.Target))
		raw, err := zr.ReadFileWithLimit(target, opts.partLimit())
		if err != nil {
			continue
		}
		partRels := map[string]string{}
		base := path.Base(target)
		if rr, e := zr.ReadFileWithLimit(path.Join(path.Dir(target), "_rels", base+".rels"), opts.partLimit()); e == nil {
			partRels = parseRelationshipTargets(rr)
		}
		if strings.Contains(r.Type, "/header") {
			h := sec.AddHeader(typ)
			parseHeaderFooter(raw, h, partRels)
		} else if strings.Contains(r.Type, "/footer") {
			f := sec.AddFooter(typ)
			parseHeaderFooter(raw, f, partRels)
		}
	}
	return nil
}

type tblFrame struct {
	tbl  *element.Table
	row  *element.Row
	cell *element.Cell
	para *element.TextRun
}

func parseDocumentXMLRels(data []byte, sec *element.Section, rels map[string]string, listStyles map[int]map[int]style.ListItem) error {
	return parseDocumentXMLRelsImages(data, sec, rels, listStyles, nil)
}

func parseDocumentXMLRelsImages(data []byte, sec *element.Section, rels map[string]string, listStyles map[int]map[int]style.ListItem, images map[string]documentImage) error {
	return parseDocumentXMLRelsImagesWithStyles(data, sec, rels, listStyles, images, nil)
}

func parseDocumentXMLRelsImagesWithStyles(data []byte, sec *element.Section, rels map[string]string, listStyles map[int]map[int]style.ListItem, images map[string]documentImage, styles *wordStyleSheet) error {
	return parseDocumentXMLRelsImagesWithStylesAndCharts(data, sec, rels, listStyles, images, nil, styles)
}

func parseDocumentXMLRelsImagesWithStylesAndCharts(data []byte, sec *element.Section, rels map[string]string, listStyles map[int]map[int]style.ListItem, images map[string]documentImage, charts map[string]*element.Chart, styles *wordStyleSheet) error {
	dec := xml.NewDecoder(bytes.NewReader(data))
	var (
		frames              []tblFrame
		inHyper             bool
		runBuf              strings.Builder
		hyperTarget         string
		hyperInternal       bool
		paraFormat          wordParagraphFormat
		bodyRun             *element.TextRun
		font                style.Font
		track               *element.TrackChange
		field               *element.Field
		fieldStage          string
		fieldMalformed      bool
		fieldInstr          strings.Builder
		fieldResult         strings.Builder
		contentNodeID       int
		pendingCommentStart *element.Comment
		lastInline          element.Element
	)
	nextContentID := func(kind string) string {
		contentNodeID++
		return kind + "-" + strconv.Itoa(contentNodeID)
	}
	currentCell := func() *element.Cell {
		if len(frames) == 0 {
			return nil
		}
		return frames[len(frames)-1].cell
	}
	currentRun := func() *element.TextRun {
		if c := currentCell(); c != nil {
			f := &frames[len(frames)-1]
			if f.para == nil {
				f.para = c.AddTextRun()
			}
			return f.para
		}
		if bodyRun == nil {
			bodyRun = sec.AddTextRun()
		}
		return bodyRun
	}
	runFont := func() any {
		if !font.IsZero() {
			return font
		}
		return nil
	}
	flushRun := func() {
		t := runBuf.String()
		runBuf.Reset()
		if t == "" {
			return
		}
		if field != nil {
			if fieldStage == "result" {
				fieldResult.WriteString(t)
				return
			}
		}
		if inHyper {
			link := currentRun().AddLink(hyperTarget, t, runFont(), nil, hyperInternal)
			link.SetTrackChange(track)
			if pendingCommentStart != nil {
				link.SetCommentRangeStart(pendingCommentStart)
				pendingCommentStart = nil
			}
			lastInline = link
			return
		}
		tx := currentRun().AddText(t, runFont())
		if track != nil {
			tx.SetTrackChange(track)
		}
		if pendingCommentStart != nil {
			tx.SetCommentRangeStart(pendingCommentStart)
			pendingCommentStart = nil
		}
		lastInline = tx
	}
	appendInline := func(el element.Element) {
		if el == nil {
			return
		}
		if c := currentCell(); c != nil {
			currentRun().AppendElement(el)
			return
		}
		if bodyRun != nil {
			currentRun().AppendElement(el)
			return
		}
		sec.AppendElement(el)
	}
	finishField := func() {
		if field == nil {
			return
		}
		field.Instruction = strings.TrimSpace(fieldInstr.String())
		if field.Instruction == "" {
			field.Instruction = field.FieldType
		}
		if field.FieldType == "" {
			field.FieldType = fieldTypeFromInstruction(field.Instruction)
		}
		field.Text = fieldResult.String()
		field.HasResult = fieldStage == "result" && field.Text != ""
		if fieldMalformed {
			field.Diagnostic = "malformed complex field structure; preserved available result text"
		}
		field.FontStyle = runFont()
		field.DocPart = "word/document.xml"
		field.ElementID = nextContentID("field")
		appendInline(field)
		field = nil
		fieldStage = ""
		fieldMalformed = false
		fieldInstr.Reset()
		fieldResult.Reset()
	}
	flushHyperlink := func() {
		flushRun()
		hyperTarget = ""
		hyperInternal = false
		inHyper = false
	}
	flushPara := func() {
		flushRun()
		if field != nil {
			fieldMalformed = true
			finishField()
		}
		pNumID, pLevel := paraFormat.numID, paraFormat.level
		para := paraFormat.para
		listStyle := style.ListItem{NumId: pNumID, Depth: pLevel, ListType: style.ListTypeNumber, Format: style.NumberDecimal}
		if levels := listStyles[pNumID]; levels != nil {
			if level, ok := levels[pLevel]; ok {
				listStyle = level
			}
		}
		container, run := &sec.Container, bodyRun
		if c := currentCell(); c != nil {
			container = &c.Container
			run = frames[len(frames)-1].para
			frames[len(frames)-1].para = nil
		} else {
			bodyRun = nil
		}
		if run != nil {
			run.SetParagraphStyle(para)
			els := run.Elements()
			switch {
			case len(els) == 0:
				container.RemoveElement(run)
			case para.OutlineLevel > 0 && pNumID == 0:
				container.RemoveElement(run)
				title := container.AddTitle(run.GetText(), para.OutlineLevel)
				title.Run = run
			case pNumID > 0 && (len(els) != 1 || !isTextElement(els[0])):
				container.RemoveElement(run)
				item := container.AddListItemRun(pLevel, listStyle, para)
				for _, el := range els {
					item.AppendElement(el)
				}
			case len(els) == 1:
				if tx, ok := els[0].(*element.Text); ok {
					container.RemoveElement(run)
					if pNumID > 0 {
						item := container.AddListItem(tx.Content, pLevel, tx.FontStyle, para, listStyle)
						item.SetTrackChange(tx.GetTrackChange())
					} else {
						out := container.AddText(tx.Content, tx.FontStyle, para)
						out.SetTrackChange(tx.GetTrackChange())
					}
				}
			}
		}
		paraFormat = wordParagraphFormat{}
		font = style.Font{}
	}

	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			local := localName(t.Name)
			switch local {
			case "p":
				if field != nil {
					flushRun()
					fieldMalformed = true
					finishField()
				}
				paraFormat = styles.paragraph(wordParagraphProperties{})
				if c := currentCell(); c != nil {
					frames[len(frames)-1].para = c.AddTextRun()
				} else {
					bodyRun = sec.AddTextRun()
				}
			case "pPr":
				var properties wordParagraphProperties
				if err := dec.DecodeElement(&properties, &t); err != nil {
					return err
				}
				paraFormat = styles.paragraph(properties)
			case "r":
				font = paraFormat.font
			case "rPr":
				var properties wordRunProperties
				if err := dec.DecodeElement(&properties, &t); err != nil {
					return err
				}
				font = styles.font(paraFormat.font, properties)
			case "ins", "del":
				track = &element.TrackChange{ChangeType: local, Author: attr(t, "author"), Date: attr(t, "date")}
			case "fldSimple":
				flushRun()
				f, err := parseSimpleField(dec, t)
				if err != nil {
					return err
				}
				f.DocPart = "word/document.xml"
				f.ElementID = nextContentID("field")
				f.FontStyle = runFont()
				appendInline(f)
			case "fldChar":
				switch attr(t, "fldCharType") {
				case "begin":
					flushRun()
					if field != nil {
						fieldMalformed = true
					}
					finishField()
					field = &element.Field{}
					fieldStage = "instruction"
				case "separate":
					if field != nil {
						fieldStage = "result"
					}
				case "end":
					flushRun()
					finishField()
				}
			case "instrText":
				var s string
				if err := dec.DecodeElement(&s, &t); err != nil {
					return err
				}
				if field != nil && fieldStage == "instruction" {
					fieldInstr.WriteString(s)
				}
			case "footnoteReference", "endnoteReference":
				flushRun()
				kind := "footnote"
				if local == "endnoteReference" {
					kind = "endnote"
				}
				ref := &element.NoteReference{NoteID: atoi(attr(t, "id")), Kind: kind}
				ref.DocPart, ref.ElementID = "word/document.xml", nextContentID(kind+"-reference")
				appendInline(ref)
			case "commentReference":
				flushRun()
				ref := &element.CommentReference{CommentID: atoi(attr(t, "id"))}
				ref.DocPart, ref.ElementID = "word/document.xml", nextContentID("comment-reference")
				appendInline(ref)
				lastInline = ref
			case "commentRangeStart":
				pendingCommentStart = &element.Comment{CommentID: atoi(attr(t, "id"))}
				pendingCommentStart.Kind = "Comment"
			case "commentRangeEnd":
				if target, ok := lastInline.(interface{ SetCommentRangeEnd(*element.Comment) }); ok {
					marker := &element.Comment{CommentID: atoi(attr(t, "id"))}
					marker.Kind = "Comment"
					target.SetCommentRangeEnd(marker)
				}
			case "sdt":
				flushRun()
				s, err := parseSDT(dec, t)
				if err != nil {
					return err
				}
				s.DocPart, s.ElementID = "word/document.xml", nextContentID("sdt")
				s.SetTrackChange(track)
				if currentCell() != nil || bodyRun != nil {
					appendInline(s)
				} else {
					sec.AppendElement(s)
				}
			case "hyperlink":
				flushRun()
				inHyper = true
				hyperInternal = false
				hyperTarget = attr(t, "anchor")
				if hyperTarget != "" {
					hyperInternal = true
				} else {
					id := attr(t, "id")
					if rels != nil && rels[id] != "" {
						hyperTarget = rels[id]
					} else {
						hyperTarget = id
					}
				}
			case "bookmarkStart":
				name := attr(t, "name")
				if name == "" {
					break
				}
				if c := currentCell(); c != nil {
					currentRun().AddBookmark(name)
				} else if bodyRun != nil {
					bodyRun.AddBookmark(name)
				} else {
					sec.AddBookmark(name)
				}
			case "tbl":
				if currentCell() != nil {
					flushRun()
				} else {
					flushPara()
				}
				var tbl *element.Table
				if c := currentCell(); c != nil {
					tbl = c.AddTable()
				} else {
					tbl = sec.AddTable()
				}
				frames = append(frames, tblFrame{tbl: tbl})
			case "tr":
				if n := len(frames); n > 0 && frames[n-1].tbl != nil {
					frames[n-1].row = frames[n-1].tbl.AddRow()
				}
			case "tc":
				if n := len(frames); n > 0 && frames[n-1].row != nil {
					frames[n-1].cell = frames[n-1].row.AddCell(0)
				}
			case "tblPr":
				if n := len(frames); n > 0 && frames[n-1].tbl != nil {
					var properties wordTableProperties
					if err := decodeWordTableProperties(dec, t, &properties); err != nil {
						return err
					}
					properties.apply(frames[n-1].tbl, themeFromStyles(styles))
				}
			case "trPr":
				if n := len(frames); n > 0 && frames[n-1].row != nil {
					var properties wordTableRowProperties
					if err := decodeWordTableProperties(dec, t, &properties); err != nil {
						return err
					}
					properties.apply(frames[n-1].row)
				}
			case "tcPr":
				if c := currentCell(); c != nil {
					var properties wordTableCellProperties
					if err := decodeWordTableProperties(dec, t, &properties); err != nil {
						return err
					}
					properties.apply(c, themeFromStyles(styles))
				}
			case "gridCol":
				if n := len(frames); n > 0 && frames[n-1].tbl != nil {
					frames[n-1].tbl.Style.GridWidths = append(frames[n-1].tbl.Style.GridWidths, atoi(attr(t, "w")))
				}
			case "t", "delText":
				var s string
				if err := dec.DecodeElement(&s, &t); err != nil {
					return err
				}
				if field != nil && fieldStage == "result" {
					fieldResult.WriteString(s)
				} else {
					runBuf.WriteString(s)
				}
			case "br", "cr":
				if attr(t, "type") == "page" {
					if currentCell() != nil {
						flushPara()
						currentCell().AddPageBreak()
					} else {
						flushPara()
						sec.AddPageBreak()
					}
				} else {
					flushRun()
					currentRun().AddTextBreak()
				}
			case "drawing":
				info := parseDrawing(dec)
				if img, ok := images[info.rid]; ok {
					width, height := info.width, info.height
					st := style.Image{WidthEMU: width, HeightEMU: height, AltText: info.alt}
					if width > 0 {
						st.Width = common.EMUToPixel(width)
					}
					if height > 0 {
						st.Height = common.EMUToPixel(height)
					}
					out := currentRun().AddImageBytes(img.Name, img.Data, st)
					out.Media.Target = img.Target
					out.Media.Ext = strings.TrimPrefix(strings.ToLower(path.Ext(img.Name)), ".")
				} else if ch, ok := charts[info.rid]; ok {
					flushRun()
					clone := cloneChartForReader(ch)
					if clone.DocPart == "" {
						clone.DocPart = "word/charts"
					}
					clone.ElementID = nextContentID("chart")
					currentRun().AppendElement(clone)
				} else if info.geom != "" || info.text != "" {
					flushRun()
					shape := &element.DMLShape{PrstGeom: info.geom, Width: int(info.width), Height: int(info.height), FillColor: info.fill, LineColor: info.line, LineWidth: int(info.lineWidth), Rotation: info.rotation, FlipH: info.flipH, FlipV: info.flipV}
					if info.geom == "" {
						shape.Diagnostic = "unknown DrawingML geometry"
					}
					if info.text != "" {
						if info.bold {
							shape.AddText(info.text, style.Font{Bold: true})
						} else {
							shape.AddText(info.text)
						}
					}
					currentRun().AppendElement(shape)
				} else if info.rid != "" {
					// Keep the drawing in the DOM so HTML conversion can report the
					// missing relationship or media part instead of dropping it.
					st := style.Image{AltText: info.alt}
					if info.width > 0 {
						st.Width = common.EMUToPixel(info.width)
						st.WidthEMU = info.width
					}
					if info.height > 0 {
						st.Height = common.EMUToPixel(info.height)
						st.HeightEMU = info.height
					}
					out := currentRun().AddImageBytes("missing-"+info.rid, nil, st)
					out.Media.Target = info.rid
				} else {
					currentRun().AppendElement(&element.DMLShape{Diagnostic: "unknown DrawingML object"})
				}
			case "sectPr":
				if err := skip(dec, t); err != nil {
					return err
				}
			}
		case xml.EndElement:
			local := localName(t.Name)
			switch local {
			case "r":
				flushRun()
				font = style.Font{}
			case "p":
				if field != nil {
					flushRun()
					fieldMalformed = true
					finishField()
				}
				flushPara()
			case "hyperlink":
				flushHyperlink()
			case "ins", "del":
				track = nil
			case "tbl":
				if n := len(frames); n > 0 {
					frames = frames[:n-1]
				}
			case "tr":
				if n := len(frames); n > 0 {
					frames[n-1].row = nil
				}
			case "tc":
				if n := len(frames); n > 0 {
					frames[n-1].cell = nil
				}
			}
		}
	}
	return nil
}

func isTextElement(el element.Element) bool {
	_, ok := el.(*element.Text)
	return ok
}

func parseDrawing(dec *xml.Decoder) drawingInfo {
	var out drawingInfo
	depth := 1
	for depth > 0 {
		tok, err := dec.Token()
		if err != nil {
			return out
		}
		switch t := tok.(type) {
		case xml.StartElement:
			depth++
			switch localName(t.Name) {
			case "chart":
				if out.rid == "" {
					out.rid = attr(t, "id")
				}
			case "blip":
				if out.rid == "" {
					out.rid = attr(t, "embed")
				}
			case "extent":
				if out.width == 0 {
					out.width = atoi64(attr(t, "cx"))
					out.height = atoi64(attr(t, "cy"))
				}
			case "docPr":
				out.alt = attr(t, "descr")
			case "prstGeom":
				out.geom = attr(t, "prst")
			case "solidFill":
				if out.fill == "" {
					out.fill = attr(t, "val")
				}
			case "srgbClr":
				if out.fill == "" {
					out.fill = attr(t, "val")
				} else if out.line == "" {
					out.line = attr(t, "val")
				}
			case "ln":
				out.lineWidth = atoi64(attr(t, "w"))
			case "wsp":
				out.textBox = false
			case "xfrm":
				if v := atoi64(attr(t, "rot")); v != 0 {
					out.rotation = float64(v) / 60000
				}
				out.flipH = attr(t, "flipH") == "1" || strings.EqualFold(attr(t, "flipH"), "true")
				out.flipV = attr(t, "flipV") == "1" || strings.EqualFold(attr(t, "flipV"), "true")
			case "t":
				var s string
				if err := dec.DecodeElement(&s, &t); err == nil {
					out.text += s
					depth--
				}
			case "b":
				out.bold = true
			}
		case xml.EndElement:
			depth--
		}
	}
	return out
}

func cloneChartForReader(ch *element.Chart) *element.Chart {
	if ch == nil {
		return &element.Chart{}
	}
	out := *ch
	out.Categories = append([]string(nil), ch.Categories...)
	out.Values = append([]float64(nil), ch.Values...)
	out.Series = append([]element.ChartSeries(nil), ch.Series...)
	for i := range out.Series {
		out.Series[i].Categories = append([]string(nil), ch.Series[i].Categories...)
		out.Series[i].Values = append([]float64(nil), ch.Series[i].Values...)
	}
	return &out
}

func parseChartPart(data []byte) *element.Chart {
	ch := &element.Chart{}
	dec := xml.NewDecoder(bytes.NewReader(data))
	var current *element.ChartSeries
	var inCat, inVal, inSeriesTitle, inChartTitle, inCatAxisTitle, inValAxisTitle bool
	var currentChartKind string
	var catIndex, valIndex int
	var pointIndex int
	for {
		tok, err := dec.Token()
		if err != nil {
			break
		}
		switch t := tok.(type) {
		case xml.StartElement:
			switch localName(t.Name) {
			case "barChart":
				if ch.ChartType == "" {
					ch.ChartType = "bar"
				}
				currentChartKind = "bar"
			case "lineChart":
				if ch.ChartType == "" {
					ch.ChartType = "line"
				}
				currentChartKind = "line"
			case "pieChart":
				if ch.ChartType == "" {
					ch.ChartType = "pie"
				}
				currentChartKind = "pie"
			case "areaChart":
				if ch.ChartType == "" {
					ch.ChartType = "area"
				}
				currentChartKind = "area"
			case "scatterChart":
				if ch.ChartType == "" {
					ch.ChartType = "scatter"
				}
				currentChartKind = "scatter"
			case "ser":
				current = &element.ChartSeries{}
				ch.Series = append(ch.Series, *current)
				current = &ch.Series[len(ch.Series)-1]
				current.Kind = currentChartKind
				if len(ch.Series) > 1 && currentChartKind != ch.ChartType {
					current.SecondaryAxis = true
				}
			case "tx":
				inSeriesTitle = current != nil
				inChartTitle = current == nil && ch.Style.Title == ""
			case "cat":
				inCat = current != nil
				catIndex = 0
			case "val":
				inVal = current != nil
				valIndex = 0
			case "catAx":
				inCatAxisTitle = true
			case "valAx":
				inValAxisTitle = true
			case "legend":
				ch.Style.ShowLegend = true
			case "legendPos":
				ch.Style.LegendPosition = attr(t, "val")
			case "catAxTitle":
				inCatAxisTitle = true
			case "valAxTitle":
				inValAxisTitle = true
			case "pt":
				pointIndex = atoi(attr(t, "idx"))
			case "crosses":
				if current != nil && attr(t, "val") == "max" {
					current.SecondaryAxis = true
				}
			case "t":
				var s string
				if dec.DecodeElement(&s, &t) == nil {
					if inSeriesTitle && current != nil {
						current.Name += s
					}
					if inChartTitle {
						ch.Style.Title += s
					}
					if inCatAxisTitle {
						ch.Style.CategoryAxisTitle += s
					}
					if inValAxisTitle {
						ch.Style.ValueAxisTitle += s
					}
					if inCat && current != nil {
						for len(current.Categories) <= pointIndex {
							current.Categories = append(current.Categories, "")
						}
						current.Categories[pointIndex] = s
					}
				}
			case "v":
				var s string
				if dec.DecodeElement(&s, &t) == nil && current != nil {
					if inCat && current != nil {
						for len(current.Categories) <= pointIndex {
							current.Categories = append(current.Categories, "")
						}
						current.Categories[pointIndex] = strings.TrimSpace(s)
					}
					if inSeriesTitle && current != nil {
						current.Name += strings.TrimSpace(s)
					}
					if inVal {
						for len(current.Values) <= pointIndex {
							current.Values = append(current.Values, 0)
							current.MissingValues = append(current.MissingValues, true)
						}
						if n, e := strconv.ParseFloat(strings.TrimSpace(s), 64); e == nil {
							current.Values[pointIndex] = n
							current.MissingValues[pointIndex] = false
						}
					}
				}
			}
		case xml.EndElement:
			switch localName(t.Name) {
			case "tx":
				inSeriesTitle = false
				inChartTitle = false
			case "cat":
				inCat = false
			case "val":
				inVal = false
			case "catAxTitle":
				inCatAxisTitle = false
			case "valAxTitle":
				inValAxisTitle = false
			case "catAx":
				inCatAxisTitle = false
			case "valAx":
				inValAxisTitle = false
			case "ser":
				current = nil
			}
		}
	}
	if len(ch.Series) > 0 {
		ch.Categories = ch.Series[0].Categories
		ch.Values = ch.Series[0].Values
		ch.SeriesName = ch.Series[0].Name
	}
	_ = catIndex
	_ = valIndex
	_ = pointIndex
	return ch
}

func parseRelationshipDetails(data []byte) map[string]documentRelationship {
	out := map[string]documentRelationship{}
	dec := xml.NewDecoder(bytes.NewReader(data))
	for {
		tok, err := dec.Token()
		if err != nil {
			break
		}
		se, ok := tok.(xml.StartElement)
		if !ok {
			continue
		}
		if localName(se.Name) != "Relationship" {
			continue
		}
		id, target := attr(se, "Id"), attr(se, "Target")
		if id == "" {
			id = attr(se, "id")
		}
		typ := attr(se, "Type")
		if id != "" && target != "" {
			out[id] = documentRelationship{Target: target, Type: typ}
		}
	}
	return out
}

func parseRelationshipTargets(data []byte) map[string]string {
	out := map[string]string{}
	for id, rel := range parseRelationshipDetails(data) {
		out[id] = rel.Target
	}
	return out
}

func parseHeaderFooter(data []byte, c interface{ AddTextRun(...any) *element.TextRun }, rels map[string]string) {
	dec := xml.NewDecoder(bytes.NewReader(data))
	var run *element.TextRun
	var inHyper bool
	var target string
	var text strings.Builder
	flushText := func() {
		if run != nil && text.Len() > 0 {
			run.AddLink(target, text.String())
			text.Reset()
		}
	}
	for {
		tok, err := dec.Token()
		if err != nil {
			break
		}
		switch t := tok.(type) {
		case xml.StartElement:
			switch localName(t.Name) {
			case "p":
				run = c.AddTextRun()
			case "t":
				var s string
				if dec.DecodeElement(&s, &t) == nil {
					if inHyper {
						text.WriteString(s)
					} else {
						if run == nil {
							run = c.AddTextRun()
						}
						run.AddText(s)
					}
				}
			case "hyperlink":
				inHyper = true
				target = attr(t, "anchor")
				if target == "" && rels != nil {
					target = rels[attr(t, "id")]
				}
			}
		case xml.EndElement:
			switch localName(t.Name) {
			case "hyperlink":
				flushText()
				inHyper = false
				target = ""
			case "p":
				flushText()
				run = nil
			}
		}
	}
}

func parseCoreProperties(data []byte, info *metadata.DocInfo) {
	if info == nil {
		return
	}
	dec := xml.NewDecoder(bytes.NewReader(data))
	var buf strings.Builder
	for {
		tok, err := dec.Token()
		if err != nil {
			return
		}
		switch t := tok.(type) {
		case xml.StartElement:
			buf.Reset()
		case xml.CharData:
			buf.Write(t)
		case xml.EndElement:
			v := strings.TrimSpace(buf.String())
			switch localName(t.Name) {
			case "creator":
				if v != "" {
					info.Creator = v
				}
			case "lastModifiedBy":
				if v != "" {
					info.LastModifiedBy = v
				}
			case "title":
				info.Title = v
			case "subject":
				info.Subject = v
			case "description":
				info.Description = v
			case "keywords":
				info.Keywords = v
			case "category":
				info.Category = v
			case "created":
				if tm := parseW3Time(v); !tm.IsZero() {
					info.Created = tm
				}
			case "modified":
				if tm := parseW3Time(v); !tm.IsZero() {
					info.Modified = tm
				}
			}
			buf.Reset()
		}
	}
}

func parseW3Time(s string) time.Time {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}
	}
	for _, layout := range []string{
		time.RFC3339,
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05",
		"2006-01-02",
	} {
		if t, err := time.Parse(layout, s); err == nil {
			return t
		}
	}
	return time.Time{}
}

func atoi(s string) int {
	n, _ := strconv.Atoi(s)
	return n
}

func atoi64(s string) int64 {
	n, _ := strconv.ParseInt(s, 10, 64)
	return n
}

func localName(n xml.Name) string {
	s := n.Local
	if i := strings.IndexByte(s, ':'); i >= 0 {
		return s[i+1:]
	}
	return s
}

func attr(t xml.StartElement, name string) string {
	for _, a := range t.Attr {
		if a.Name.Local == name || strings.HasSuffix(a.Name.Local, ":"+name) {
			return a.Value
		}
	}
	return ""
}

func onOffTrue(t xml.StartElement) bool {
	v := strings.ToLower(attr(t, "val"))
	return v != "0" && v != "false" && v != "off"
}

func skip(dec *xml.Decoder, start xml.StartElement) error {
	depth := 1
	for depth > 0 {
		tok, err := dec.Token()
		if err != nil {
			return err
		}
		switch tok.(type) {
		case xml.StartElement:
			depth++
		case xml.EndElement:
			depth--
		}
	}
	return nil
}
