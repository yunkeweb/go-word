package math

import (
	"encoding/xml"
	"fmt"
	"io"
	"strings"

	"github.com/yunkeweb/go-word/pkg/common"
)

const ommlNS = "http://schemas.openxmlformats.org/officeDocument/2006/math"

// WriteOMML serializes m as a display equation (m:oMathPara / m:oMath).
func WriteOMML(m *Math) ([]byte, error) {
	w := common.GetXMLWriter()
	w.Start("m:oMathPara", "xmlns:m", ommlNS)
	w.Start("m:oMath")
	writeMathBody(w, m)
	w.End()
	w.End()
	return common.FinishXML(w), nil
}

// WriteOMath serializes m as an inline m:oMath (no oMathPara wrapper).
func WriteOMath(m *Math) ([]byte, error) {
	w := common.GetXMLWriter()
	w.Start("m:oMath")
	writeMathBody(w, m)
	w.End()
	return common.FinishXML(w), nil
}

func writeMathBody(w *common.XMLWriter, m *Math) {
	if m == nil {
		return
	}
	for _, el := range m.Elements {
		writeElement(w, el)
	}
}

func writeElement(w *common.XMLWriter, el Element) {
	switch v := el.(type) {
	case *Row:
		for _, c := range v.Elements {
			writeElement(w, c)
		}
	case *Fraction:
		w.Start("m:f")
		w.Start("m:num")
		if v.Numerator != nil {
			writeElement(w, v.Numerator)
		}
		w.End()
		w.Start("m:den")
		if v.Denominator != nil {
			writeElement(w, v.Denominator)
		}
		w.End()
		w.End()
	case *Superscript:
		w.Start("m:sSup")
		w.Start("m:e")
		if v.Base != nil {
			writeElement(w, v.Base)
		}
		w.End()
		w.Start("m:sup")
		if v.Sup != nil {
			writeElement(w, v.Sup)
		}
		w.End()
		w.End()
	case *Subscript:
		w.Start("m:sSub")
		w.Start("m:e")
		if v.Base != nil {
			writeElement(w, v.Base)
		}
		w.End()
		w.Start("m:sub")
		if v.Sub != nil {
			writeElement(w, v.Sub)
		}
		w.End()
		w.End()
	case *Radical:
		w.Start("m:rad")
		w.Start("m:radPr")
		w.Empty("m:degHide", "m:val", "1")
		w.End()
		w.Start("m:deg")
		w.End()
		w.Start("m:e")
		if v.Base != nil {
			writeElement(w, v.Base)
		}
		w.End()
		w.End()
	case *Delimiter:
		w.Start("m:d")
		w.Start("m:dPr")
		if v.Beg != "" {
			w.Empty("m:begChr", "m:val", v.Beg)
		}
		if v.End != "" {
			w.Empty("m:endChr", "m:val", v.End)
		}
		w.End()
		w.Start("m:e")
		if v.Content != nil {
			writeElement(w, v.Content)
		}
		w.End()
		w.End()
	case *Identifier:
		writeRun(w, v.Value)
	case *Numeric:
		writeRun(w, v.Value)
	case *Operator:
		writeRun(w, v.Value)
	case *Semantics:
		for _, c := range v.Elements {
			writeElement(w, c)
		}
	case *Math:
		for _, c := range v.Elements {
			writeElement(w, c)
		}
	}
}

func writeRun(w *common.XMLWriter, text string) {
	w.Start("m:r")
	w.Start("m:t", "xml:space", "preserve")
	w.Text(text)
	w.End()
	w.End()
}

// ReadOMML parses an OMML fragment into a Math tree.
func ReadOMML(r io.Reader) (*Math, error) {
	dec := xml.NewDecoder(r)
	m := New()
	var stack []Element
	push := func(el Element) {
		if len(stack) == 0 {
			m.Add(el)
			stack = append(stack, el)
			return
		}
		switch g := stack[len(stack)-1].(type) {
		case *Row:
			g.Add(el)
		case *Semantics:
			g.Add(el)
		case *Math:
			g.Add(el)
		case *Fraction:
			if g.Numerator == nil {
				g.Numerator = el
			} else if g.Denominator == nil {
				g.Denominator = el
			}
		case *Superscript:
			if g.Base == nil {
				g.Base = el
			} else if g.Sup == nil {
				g.Sup = el
			}
		case *Subscript:
			if g.Base == nil {
				g.Base = el
			} else if g.Sub == nil {
				g.Sub = el
			}
		case *Radical:
			if g.Base == nil {
				g.Base = el
			}
		case *Delimiter:
			if g.Content == nil {
				g.Content = el
			}
		}
		stack = append(stack, el)
	}
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
			local := t.Name.Local
			if i := strings.IndexByte(local, ':'); i >= 0 {
				local = local[i+1:]
			}
			switch local {
			case "oMathPara", "oMath":
				// container, ignore
			case "f":
				push(NewFraction(nil, nil))
			case "sSup":
				push(NewSuperscript(nil, nil))
			case "sSub":
				push(NewSubscript(nil, nil))
			case "rad":
				push(NewRadical(nil))
			case "d":
				push(NewDelimiter("", "", nil))
			case "r":
				// wait for t
			case "t":
				var text string
				if err := dec.DecodeElement(&text, &t); err != nil {
					return nil, err
				}
				text = strings.TrimSpace(text)
				el := Element(NewIdentifier(text))
				if looksNumeric(text) {
					el = NewNumeric(text)
				}
				if len(stack) == 0 {
					m.Add(el)
				} else {
					push(el)
					if len(stack) > 0 {
						stack = stack[:len(stack)-1]
					}
				}
			case "num", "den", "e", "sup", "sub", "deg", "dPr", "radPr", "fPr":
				// structural, ignore
			default:
				// skip unknown
			}
		case xml.EndElement:
			local := t.Name.Local
			if i := strings.IndexByte(local, ':'); i >= 0 {
				local = local[i+1:]
			}
			switch local {
			case "f", "sSup", "sSub", "rad", "d":
				if len(stack) > 0 {
					stack = stack[:len(stack)-1]
				}
			}
		}
	}
	if m == nil {
		return nil, fmt.Errorf("math: empty OMML")
	}
	return m, nil
}

func looksNumeric(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r != '.' && r != '-' && (r < '0' || r > '9') {
			return false
		}
	}
	return true
}
