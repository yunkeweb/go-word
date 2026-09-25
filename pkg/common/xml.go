package common

import (
	"bytes"
	"encoding/xml"
	"io"
	"strings"
)

// XMLWriter writes OpenXML using encoding/xml.Encoder.
type XMLWriter struct {
	buf   *bytes.Buffer
	enc   *xml.Encoder
	stack []string
}

// NewXMLWriter returns a writer that emits UTF-8 XML.
func NewXMLWriter() *XMLWriter {
	buf := &bytes.Buffer{}
	enc := xml.NewEncoder(buf)
	return &XMLWriter{buf: buf, enc: enc}
}

// StartDocument writes the XML declaration used by Office Open XML.
func (w *XMLWriter) StartDocument() {
	w.buf.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`)
}

// Start opens an element. attrs is a sequence of name, value pairs.
func (w *XMLWriter) Start(name string, attrs ...string) {
	el := xml.StartElement{Name: xml.Name{Local: name}}
	for i := 0; i+1 < len(attrs); i += 2 {
		if attrs[i+1] == "" {
			continue
		}
		el.Attr = append(el.Attr, xml.Attr{
			Name:  xml.Name{Local: attrs[i]},
			Value: attrs[i+1],
		})
	}
	_ = w.enc.EncodeToken(el)
	w.stack = append(w.stack, name)
}

// End closes the most recently opened element.
func (w *XMLWriter) End() {
	if len(w.stack) == 0 {
		return
	}
	name := w.stack[len(w.stack)-1]
	w.stack = w.stack[:len(w.stack)-1]
	_ = w.enc.EncodeToken(xml.EndElement{Name: xml.Name{Local: name}})
}

// Empty writes a start/end pair with optional attributes and no content.
func (w *XMLWriter) Empty(name string, attrs ...string) {
	w.Start(name, attrs...)
	w.End()
}

// Element writes an element with character data.
func (w *XMLWriter) Element(name, text string, attrs ...string) {
	w.Start(name, attrs...)
	if text != "" {
		w.Text(text)
	}
	w.End()
}

// Text writes character data. encoding/xml escapes the five XML entities.
func (w *XMLWriter) Text(s string) {
	if s == "" {
		return
	}
	_ = w.enc.EncodeToken(xml.CharData(s))
}

// WT writes a w:t run of text, setting xml:space="preserve" when needed.
func (w *XMLWriter) WT(s string) {
	s = ControlCharEncode(s)
	attrs := []string{}
	if s != strings.TrimSpace(s) || strings.ContainsAny(s, "\t\n") {
		attrs = []string{"xml:space", "preserve"}
	}
	w.Start("w:t", attrs...)
	w.Text(s)
	w.End()
}

// Raw writes already-encoded XML after flushing the encoder.
func (w *XMLWriter) Raw(s string) {
	_ = w.enc.Flush()
	w.buf.WriteString(s)
}

// Bytes flushes the encoder and returns the document.
func (w *XMLWriter) Bytes() []byte {
	_ = w.enc.Flush()
	return w.buf.Bytes()
}

// String flushes the encoder and returns the document as a string.
func (w *XMLWriter) String() string {
	return string(w.Bytes())
}

// WriteTo flushes and copies the document to dest.
func (w *XMLWriter) WriteTo(dest io.Writer) (int64, error) {
	n, err := dest.Write(w.Bytes())
	return int64(n), err
}

// WriteElementIf writes name only when cond is true (PHP XMLWriter::writeElementIf).
func (w *XMLWriter) WriteElementIf(cond bool, name, text string, attrs ...string) {
	if !cond {
		return
	}
	w.Element(name, text, attrs...)
}

// WriteAttributeIf is a no-op helper kept for PHP XMLWriter::writeAttributeIf parity;
// callers pass attrs into Start/Empty which already skip empty values.
func (w *XMLWriter) WriteAttributeIf(cond bool, name, value string) (string, string) {
	if !cond || value == "" {
		return "", ""
	}
	return name, value
}

// WriteElementBlock writes a parent wrapping one or more children (PHP writeElementBlock).
func (w *XMLWriter) WriteElementBlock(name string, attrs []string, inner func()) {
	w.Start(name, attrs...)
	if inner != nil {
		inner()
	}
	w.End()
}

// MarshalXML marshals v with the Office XML declaration.
func MarshalXML(v any) ([]byte, error) {
	body, err := xml.Marshal(v)
	if err != nil {
		return nil, err
	}
	out := make([]byte, 0, len(xmlDecl)+len(body))
	out = append(out, xmlDecl...)
	out = append(out, body...)
	return out, nil
}

const xmlDecl = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`
