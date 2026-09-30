package word

import (
	"bytes"
	"encoding/xml"
	"errors"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/yunkeweb/go-word/element"
	"github.com/yunkeweb/go-word/metadata"
	"github.com/yunkeweb/go-word/pkg/common"
	"github.com/yunkeweb/go-word/style"
)

type word2007Reader struct{}

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
	rels := map[string]string{}
	if relRaw, err := zr.ReadFileWithLimit("word/_rels/document.xml.rels", opts.partLimit()); err == nil {
		rels = parseRelationshipTargets(relRaw)
	} else if !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}
	sec := doc.AddSection()
	if err := parseDocumentXMLRels(raw, sec, rels); err != nil {
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

func parseDocumentXML(data []byte, sec *element.Section) error {
	return parseDocumentXMLRels(data, sec, nil)
}

type tblFrame struct {
	tbl  *element.Table
	row  *element.Row
	cell *element.Cell
	para *element.TextRun
}

func parseDocumentXMLRels(data []byte, sec *element.Section, rels map[string]string) error {
	dec := xml.NewDecoder(bytes.NewReader(data))
	var (
		frames        []tblFrame
		inHyper       bool
		runBuf        strings.Builder
		hyperTarget   string
		hyperInternal bool
		hyperText     strings.Builder
		pStyle        string
		bodyRun       *element.TextRun
		bold, italic  bool
		color         string
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
		currentRun().AddText(t, runFont())
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
		if c := currentCell(); c != nil {
			f := &frames[len(frames)-1]
			if f.para != nil {
				f.para.SetParagraphStyle(pStyle)
				els := f.para.Elements()
				if len(els) == 0 {
					c.RemoveElement(f.para)
				}
				if len(els) == 1 {
					if tx, ok := els[0].(*element.Text); ok {
						c.RemoveElement(f.para)
						c.AddText(tx.Content, tx.FontStyle, pStyle)
					}
				}
				f.para = nil
			}
			pStyle = ""
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
			case len(els) == 1:
				if tx, ok := els[0].(*element.Text); ok {
					sec.RemoveElement(bodyRun)
					var para any
					if pStyle != "" {
						para = pStyle
					}
					sec.AddText(tx.Content, tx.FontStyle, para)
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
			case "t":
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
			case "drawing", "sectPr":
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

func parseRelationshipTargets(data []byte) map[string]string {
	out := map[string]string{}
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
		if id != "" && target != "" {
			out[id] = target
		}
	}
	return out
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
