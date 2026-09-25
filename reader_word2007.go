package word

import (
	"encoding/xml"
	"io"
	"strings"

	"github.com/yunkeweb/go-word/element"
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
	return loadWord2007(zr)
}

func LoadBytes(data []byte) (*Document, error) {
	zr, err := common.OpenZipBytes(data)
	if err != nil {
		return nil, err
	}
	return loadWord2007(zr)
}

func loadWord2007(zr *common.ZipReader) (*Document, error) {
	doc := New()
	raw, err := zr.ReadFile("word/document.xml")
	if err != nil {
		return nil, err
	}
	sec := doc.AddSection()
	if err := parseDocumentXML(raw, sec); err != nil {
		return nil, err
	}
	return doc, nil
}

func parseDocumentXML(data []byte, sec *element.Section) error {
	dec := xml.NewDecoder(strings.NewReader(string(data)))
	var (
		inP, inHyper, inTbl, inTr, inTc bool
		runBuf                          strings.Builder
		paraBuf                         strings.Builder
		hyperTarget                     string
		hyperText                       strings.Builder
		tbl                             *element.Table
		row                             *element.Row
		cell                            *element.Cell
		pStyle                          string
		rStyle                          style.Font
		bold, italic                    bool
		color                           string
	)
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
		if inTc && cell != nil {
			f := style.Font{Bold: bold, Italic: italic, Color: color}
			cell.AddText(t, f)
			return
		}
		paraBuf.WriteString(t)
	}
	flushPara := func() {
		flushRun()
		text := paraBuf.String()
		paraBuf.Reset()
		if inHyper {
			sec.AddLink(hyperTarget, hyperText.String())
			hyperText.Reset()
			hyperTarget = ""
			inHyper = false
			return
		}
		if text == "" {
			if pStyle != "" {
				// empty styled paragraph still skipped
			}
			pStyle = ""
			return
		}
		var font any
		var para any
		if bold || italic || color != "" {
			font = style.Font{Bold: bold, Italic: italic, Color: color}
		}
		if strings.HasPrefix(pStyle, "Heading") {
			depth := 1
			if len(pStyle) > 7 {
				depth = int(pStyle[7] - '0')
				if depth < 1 {
					depth = 1
				}
			}
			sec.AddTitle(text, depth)
			pStyle = ""
			bold, italic, color = false, false, ""
			return
		}
		if pStyle != "" {
			para = pStyle
		}
		sec.AddText(text, font, para)
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
				inP = true
				paraBuf.Reset()
				pStyle = ""
			case "pStyle":
				pStyle = attr(t, "val")
			case "hyperlink":
				flushRun()
				inHyper = true
				hyperTarget = attr(t, "id")
				if hyperTarget == "" {
					hyperTarget = attr(t, "anchor")
				}
			case "tbl":
				flushPara()
				inTbl = true
				tbl = sec.AddTable()
			case "tr":
				if tbl != nil {
					row = tbl.AddRow()
					inTr = true
				}
			case "tc":
				if row != nil {
					cell = row.AddCell(0)
					inTc = true
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
					flushPara()
					sec.AddPageBreak()
				}
			case "drawing", "sectPr":
				if err := skip(dec, t); err != nil {
					return err
				}
			}
			_ = rStyle
		case xml.EndElement:
			local := localName(t.Name)
			switch local {
			case "r":
				flushRun()
				bold, italic, color = false, false, ""
			case "p":
				if inTc {
					flushRun()
					inP = false
					break
				}
				flushPara()
				inP = false
			case "hyperlink":
				inHyper = false
			case "tbl":
				inTbl = false
				tbl = nil
			case "tr":
				inTr = false
				row = nil
			case "tc":
				inTc = false
				cell = nil
			}
			_ = inP
			_ = inTbl
			_ = inTr
		}
	}
	return nil
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
