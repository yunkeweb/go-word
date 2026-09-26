package math

import (
	"encoding/xml"
	"fmt"
	"io"
	"strings"

	"github.com/yunkeweb/go-word/pkg/common"
)

const mathMLNS = "http://www.w3.org/1998/Math/MathML"

// WriteMathML serializes m as W3C MathML 2.0 (PHP Writer\MathML).
func WriteMathML(m *Math) ([]byte, error) {
	w := common.GetXMLWriter()
	w.StartDocument()
	w.Start("math", "xmlns", mathMLNS)
	for _, el := range m.Elements {
		writeMathMLElement(w, el)
	}
	w.End()
	return common.FinishXML(w), nil
}

func writeMathMLElement(w *common.XMLWriter, el Element) {
	switch v := el.(type) {
	case *Semantics:
		w.Start("semantics")
		for _, c := range v.Elements {
			writeMathMLElement(w, c)
		}
		for enc, ann := range v.Annotations {
			w.Element("annotation", ann, "encoding", enc)
		}
		w.End()
	case *Row:
		w.Start("mrow")
		for _, c := range v.Elements {
			writeMathMLElement(w, c)
		}
		w.End()
	case *Group:
		for _, c := range v.Elements {
			writeMathMLElement(w, c)
		}
	case *Math:
		for _, c := range v.Elements {
			writeMathMLElement(w, c)
		}
	case *Fraction:
		w.Start("mfrac")
		if v.Numerator != nil {
			writeMathMLElement(w, v.Numerator)
		}
		if v.Denominator != nil {
			writeMathMLElement(w, v.Denominator)
		}
		w.End()
	case *Superscript:
		w.Start("msup")
		if v.Base != nil {
			writeMathMLElement(w, v.Base)
		}
		if v.Sup != nil {
			writeMathMLElement(w, v.Sup)
		}
		w.End()
	case *Subscript:
		w.Start("msub")
		if v.Base != nil {
			writeMathMLElement(w, v.Base)
		}
		if v.Sub != nil {
			writeMathMLElement(w, v.Sub)
		}
		w.End()
	case *Radical:
		w.Start("msqrt")
		if v.Base != nil {
			writeMathMLElement(w, v.Base)
		}
		w.End()
	case *Delimiter:
		w.Start("mrow")
		if v.Beg != "" {
			w.Element("mo", v.Beg)
		}
		if v.Content != nil {
			writeMathMLElement(w, v.Content)
		}
		if v.End != "" {
			w.Element("mo", v.End)
		}
		w.End()
	case *Identifier:
		w.Element("mi", v.Value)
	case *Numeric:
		w.Element("mn", v.Value)
	case *Operator:
		w.Element("mo", v.Value)
	}
}

// ReadMathML parses a MathML document into a Math tree (PHP Reader\MathML).
func ReadMathML(r io.Reader) (*Math, error) {
	dec := xml.NewDecoder(r)
	m := New()
	var stack []Element
	var curText strings.Builder
	inText := false

	push := func(el Element) {
		if len(stack) == 0 {
			m.Add(el)
		} else {
			attachChild(stack[len(stack)-1], el)
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
			curText.Reset()
			inText = false
			switch local {
			case "math":
				// root
			case "mrow":
				push(NewRow())
			case "semantics":
				push(NewSemantics())
			case "mfrac":
				push(NewFraction(nil, nil))
			case "msup":
				push(NewSuperscript(nil, nil))
			case "mi", "mn", "mo", "annotation":
				inText = true
				push(&textHolder{tag: local, attrs: t.Attr})
			}
		case xml.CharData:
			if inText {
				curText.Write([]byte(t))
			}
		case xml.EndElement:
			local := t.Name.Local
			switch local {
			case "mi", "mn", "mo", "annotation", "mrow", "semantics", "mfrac", "msup":
				if len(stack) == 0 {
					continue
				}
				top := stack[len(stack)-1]
				stack = stack[:len(stack)-1]
				text := strings.TrimSpace(curText.String())
				curText.Reset()
				inText = false
				switch th := top.(type) {
				case *textHolder:
					el := leafFromTag(th.tag, text, th.attrs)
					if len(stack) == 0 {
						replaceLast(m, th, el)
					} else {
						replaceChild(stack[len(stack)-1], th, el)
						if sem, ok := stack[len(stack)-1].(*Semantics); ok && th.tag == "annotation" {
							enc := attrValue(th.attrs, "encoding")
							sem.AddAnnotation(enc, text)
							sem.Remove(el)
						}
					}
				}
			}
		}
	}
	if len(m.Elements) == 0 {
		return nil, fmt.Errorf("math: empty MathML")
	}
	return m, nil
}

type textHolder struct {
	elementBase
	tag   string
	attrs []xml.Attr
}

func leafFromTag(tag, text string, attrs []xml.Attr) Element {
	switch tag {
	case "mn":
		return NewNumeric(text)
	case "mo":
		return NewOperator(text)
	case "annotation":
		return NewIdentifier(text)
	default:
		return NewIdentifier(text)
	}
}

func attrValue(attrs []xml.Attr, name string) string {
	for _, a := range attrs {
		if a.Name.Local == name {
			return a.Value
		}
	}
	return ""
}

func attachChild(parent, child Element) {
	switch g := parent.(type) {
	case *Row:
		g.Add(child)
	case *Semantics:
		g.Add(child)
	case *Math:
		g.Add(child)
	case *Group:
		g.Add(child)
	case *Fraction:
		if g.Numerator == nil {
			g.Numerator = child
		} else if g.Denominator == nil {
			g.Denominator = child
		}
	case *Superscript:
		if g.Base == nil {
			g.Base = child
		} else if g.Sup == nil {
			g.Sup = child
		}
	}
}

func replaceLast(m *Math, old, neu Element) {
	for i, e := range m.Elements {
		if e == old {
			m.Elements[i] = neu
			return
		}
	}
	m.Add(neu)
}

func replaceChild(parent, old, neu Element) {
	switch g := parent.(type) {
	case *Row:
		for i, e := range g.Elements {
			if e == old {
				g.Elements[i] = neu
				return
			}
		}
	case *Semantics:
		for i, e := range g.Elements {
			if e == old {
				g.Elements[i] = neu
				return
			}
		}
	case *Math:
		replaceLast(g, old, neu)
	case *Group:
		for i, e := range g.Elements {
			if e == old {
				g.Elements[i] = neu
				return
			}
		}
	case *Fraction:
		if g.Numerator == old {
			g.Numerator = neu
		} else if g.Denominator == old {
			g.Denominator = neu
		}
	case *Superscript:
		if g.Base == old {
			g.Base = neu
		} else if g.Sup == old {
			g.Sup = neu
		}
	}
}
