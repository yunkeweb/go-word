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
	images := map[string]documentImage{}
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
	sections, err := parseDocumentSections(raw)
	if err != nil {
		return nil, err
	}
	if len(sections) == 0 {
		sections = append(sections, parsedSection{})
	}
	for _, parsed := range sections {
		sec := doc.AddSection(parsed.style)
		if err := parseDocumentXMLRelsImages(parsed.body, sec, relationshipTargets(rels), listStyles, images); err != nil {
			return nil, err
		}
		if err := loadSectionParts(zr, opts, sec, parsed.sectPr, rels); err != nil {
			return nil, err
		}
	}
	if err := loadNoteParts(zr, opts, doc, rels); err != nil {
		return nil, err
	}
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
				active = noteID > 0
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
	dec := xml.NewDecoder(bytes.NewReader(data))
	var (
		frames        []tblFrame
		inHyper       bool
		runBuf        strings.Builder
		hyperTarget   string
		hyperInternal bool
		hyperText     strings.Builder
		pStyle        string
		pNumID        int
		pLevel        int
		bodyRun       *element.TextRun
		bold, italic  bool
		color         string
		track         *element.TrackChange
	)
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
	headingDepth := func(name string) int {
		depth := 1
		if len(name) > 7 {
			depth = int(name[7] - '0')
			if depth < 1 {
				depth = 1
			}
		}
		return depth
	}
	runFont := func() any {
		if bold || italic || color != "" {
			return style.Font{Bold: bold, Italic: italic, Color: color}
		}
		return nil
	}
	flushRun := func() {
		t := runBuf.String()
		runBuf.Reset()
		if t == "" {
			return
		}
		if inHyper {
			hyperText.WriteString(t)
			return
		}
		tx := currentRun().AddText(t, runFont())
		if track != nil {
			tx.SetTrackChange(track)
		}
	}
	flushHyperlink := func() {
		flushRun()
		target, text, internal := hyperTarget, hyperText.String(), hyperInternal
		hyperText.Reset()
		hyperTarget = ""
		hyperInternal = false
		inHyper = false
		if target == "" && text == "" {
			return
		}
		currentRun().AddLink(target, text, nil, nil, internal)
	}
	flushPara := func() {
		flushRun()
		listStyle := style.ListItem{NumId: pNumID, Depth: pLevel, ListType: style.ListTypeNumber, Format: style.NumberDecimal}
		if levels := listStyles[pNumID]; levels != nil {
			if level, ok := levels[pLevel]; ok {
				listStyle = level
			}
		}
		if c := currentCell(); c != nil {
			f := &frames[len(frames)-1]
			if f.para != nil {
				f.para.SetParagraphStyle(pStyle)
				els := f.para.Elements()
				if len(els) == 0 {
					c.RemoveElement(f.para)
				} else if pNumID > 0 {
					c.RemoveElement(f.para)
					item := c.AddListItemRun(pLevel, listStyle, pStyle)
					for _, el := range els {
						item.AppendElement(el)
					}
				} else if len(els) == 1 {
					if tx, ok := els[0].(*element.Text); ok {
						c.RemoveElement(f.para)
						out := c.AddText(tx.Content, tx.FontStyle, pStyle)
						out.SetTrackChange(tx.GetTrackChange())
					}
				}
				f.para = nil
			}
			pStyle = ""
			pNumID, pLevel = 0, 0
			bold, italic, color = false, false, ""
			return
		}
		text := ""
		if bodyRun != nil {
			text = bodyRun.GetText()
		}
		if strings.HasPrefix(pStyle, "Heading") {
			if bodyRun != nil {
				sec.RemoveElement(bodyRun)
				bodyRun = nil
			}
			if text != "" {
				sec.AddTitle(text, headingDepth(pStyle))
			}
			pStyle = ""
			bold, italic, color = false, false, ""
			return
		}
		if bodyRun != nil {
			els := bodyRun.Elements()
			switch {
			case len(els) == 0:
				sec.RemoveElement(bodyRun)
			case pNumID > 0 && (len(els) != 1 || !isTextElement(els[0])):
				sec.RemoveElement(bodyRun)
				item := sec.AddListItemRun(pLevel, listStyle, pStyle)
				for _, el := range els {
					item.AppendElement(el)
				}
			case len(els) == 1:
				if tx, ok := els[0].(*element.Text); ok {
					sec.RemoveElement(bodyRun)
					var para any
					if pStyle != "" {
						para = pStyle
					}
					if pNumID > 0 {
						item := sec.AddListItem(tx.Content, pLevel, tx.FontStyle, para, listStyle)
						item.SetTrackChange(tx.GetTrackChange())
					} else {
						out := sec.AddText(tx.Content, tx.FontStyle, para)
						out.SetTrackChange(tx.GetTrackChange())
					}
				} else if pStyle != "" {
					bodyRun.SetParagraphStyle(pStyle)
				}
			default:
				if pStyle != "" {
					bodyRun.SetParagraphStyle(pStyle)
				}
			}
			bodyRun = nil
		}
		pStyle = ""
		pNumID, pLevel = 0, 0
		bold, italic, color = false, false, ""
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
				pStyle = ""
				if c := currentCell(); c != nil {
					frames[len(frames)-1].para = c.AddTextRun()
				} else {
					bodyRun = sec.AddTextRun()
				}
			case "pStyle":
				pStyle = attr(t, "val")
			case "ins", "del":
				track = &element.TrackChange{ChangeType: local, Author: attr(t, "author"), Date: attr(t, "date")}
			case "ilvl":
				pLevel = atoi(attr(t, "val"))
			case "numId":
				pNumID = atoi(attr(t, "val"))
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
			case "tcW":
				if c := currentCell(); c != nil {
					if w := atoi(attr(t, "w")); w > 0 {
						c.Width = w
						c.Style.Width = w
					}
					if u := attr(t, "type"); u != "" {
						c.Style.Unit = u
					}
				}
			case "vAlign":
				if c := currentCell(); c != nil {
					if v := attr(t, "val"); v != "" {
						c.Style.VAlign = v
					}
				}
			case "textDirection":
				if c := currentCell(); c != nil {
					if v := attr(t, "val"); v != "" {
						c.Style.TextDir = v
					}
				}
			case "tblHeader":
				if n := len(frames); n > 0 && frames[n-1].row != nil && onOffTrue(t) {
					frames[n-1].row.Style.Header = true
					frames[n-1].row.Style.TblHeader = true
				}
			case "cantSplit":
				if n := len(frames); n > 0 && frames[n-1].row != nil && onOffTrue(t) {
					frames[n-1].row.Style.CantSplit = true
				}
			case "trHeight":
				if n := len(frames); n > 0 && frames[n-1].row != nil {
					row := frames[n-1].row
					if h := atoi(attr(t, "val")); h > 0 {
						row.Style.Height = h
					}
					if rule := attr(t, "hRule"); rule != "" {
						row.Style.Rule = rule
					}
				}
			case "t", "delText":
				var s string
				if err := dec.DecodeElement(&s, &t); err != nil {
					return err
				}
				runBuf.WriteString(s)
			case "b":
				if attr(t, "val") != "0" && attr(t, "val") != "false" {
					bold = true
				}
			case "i":
				if attr(t, "val") != "0" && attr(t, "val") != "false" {
					italic = true
				}
			case "color":
				color = attr(t, "val")
			case "br":
				if attr(t, "type") == "page" {
					if currentCell() != nil {
						flushPara()
						currentCell().AddPageBreak()
					} else {
						flushPara()
						sec.AddPageBreak()
					}
				}
			case "drawing":
				rid, width, height, alt := parseDrawing(dec)
				if img, ok := images[rid]; ok {
					st := style.Image{WidthEMU: width, HeightEMU: height, AltText: alt}
					if width > 0 {
						st.Width = common.EMUToPixel(width)
					}
					if height > 0 {
						st.Height = common.EMUToPixel(height)
					}
					out := currentRun().AddImageBytes(img.Name, img.Data, st)
					out.Media.Target = img.Target
					out.Media.Ext = strings.TrimPrefix(strings.ToLower(path.Ext(img.Name)), ".")
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
				bold, italic, color = false, false, ""
			case "p":
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

func parseDrawing(dec *xml.Decoder) (rid string, width, height int64, alt string) {
	depth := 1
	for depth > 0 {
		tok, err := dec.Token()
		if err != nil {
			return rid, width, height, alt
		}
		switch t := tok.(type) {
		case xml.StartElement:
			depth++
			switch localName(t.Name) {
			case "blip":
				if rid == "" {
					rid = attr(t, "embed")
				}
			case "extent":
				if width == 0 {
					width = atoi64(attr(t, "cx"))
					height = atoi64(attr(t, "cy"))
				}
			case "docPr":
				alt = attr(t, "descr")
			}
		case xml.EndElement:
			depth--
		}
	}
	return rid, width, height, alt
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
