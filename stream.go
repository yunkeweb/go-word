package word

import (
	"fmt"
	"io"

	"github.com/yunkeweb/go-word/element"
	"github.com/yunkeweb/go-word/ooxml"
	"github.com/yunkeweb/go-word/pkg/common"
)

// StreamWriter writes paragraphs and tables incrementally into a Word2007
// (.docx) ZIP stream. document.xml is opened on the first write so large
// tables never accumulate as a single in-memory XML tree.
//
// Close finishes document.xml and emits the remaining OOXML parts
// (content types, relationships, styles, media). StreamWriter is not
// safe for concurrent use.
type StreamWriter struct {
	dest   io.Writer
	cw     *countingWriter
	zw     *common.ZipWriter
	xw     *common.XMLWriter
	inner  *word2007Writer
	sec    *element.Section
	closed bool
	err    error
}

// NewStreamWriter starts a Word2007 package on dest.
func NewStreamWriter(dest io.Writer) *StreamWriter {
	doc := New()
	sec := doc.AddSection()
	return &StreamWriter{
		dest:  dest,
		inner: newWord2007Writer(doc),
		sec:   sec,
	}
}

// WriteParagraph writes a paragraph of text into the open document.xml stream.
// styles follows Section.AddText: optional font then paragraph style.
func (s *StreamWriter) WriteParagraph(text string, styles ...any) error {
	if err := s.check(); err != nil {
		return err
	}
	var font, para any
	if len(styles) > 0 {
		font = styles[0]
	}
	if len(styles) > 1 {
		para = styles[1]
	}
	t := element.NewText(text, font, para)
	s.inner.writeText(s.xw, t, false)
	return s.xw.Flush()
}

// WriteTable writes a table, flushing XML after each row.
func (s *StreamWriter) WriteTable(tbl *element.Table) error {
	if err := s.check(); err != nil {
		return err
	}
	if tbl == nil {
		return fmt.Errorf("word: nil table")
	}
	s.inner.writeTable(s.xw, tbl)
	return s.xw.Flush()
}

// WriteElement writes any document body element into the stream.
func (s *StreamWriter) WriteElement(el element.Element) error {
	if err := s.check(); err != nil {
		return err
	}
	if el == nil {
		return fmt.Errorf("word: nil element")
	}
	switch v := el.(type) {
	case *element.Image:
		_ = s.inner.registerImage(v)
	case *element.Chart:
		s.inner.chartIndex++
		name := fmt.Sprintf("charts/chart%d.xml", s.inner.chartIndex)
		id := s.inner.addRel(ooxml.NSOfficeRelChart, name, "")
		v.RelationID = relIDNum(id)
		s.inner.charts = append(s.inner.charts, pkgChart{RelID: id, Name: "word/" + name, El: v})
	case *element.OLEObject:
		s.inner.registerOLE(v)
	case *element.Link:
		if !v.Internal && v.Target != "" {
			s.inner.addRel(ooxml.NSOfficeRelHyperlink, v.Target, "External")
		}
	}
	s.inner.writeElement(s.xw, el, false)
	return s.xw.Flush()
}

// BytesWritten returns the number of bytes written to dest so far.
func (s *StreamWriter) BytesWritten() int64 {
	if s.cw == nil {
		return 0
	}
	return s.cw.n
}

// Close finishes document.xml and writes the remaining package parts.
func (s *StreamWriter) Close() error {
	if s.closed {
		return s.err
	}
	s.closed = true
	if err := s.begin(); err != nil {
		s.err = err
		return err
	}
	s.inner.writeSectPr(s.xw, s.sec)
	s.inner.writeDocumentEnd(s.xw)
	if err := s.xw.Flush(); err != nil {
		s.err = err
		_ = s.zw.Close()
		return err
	}
	s.xw = nil
	if err := s.inner.writeSupportingParts(s.zw); err != nil {
		s.err = err
		_ = s.zw.Close()
		return err
	}
	s.err = s.zw.Close()
	return s.err
}

func (s *StreamWriter) check() error {
	if s.err != nil {
		return s.err
	}
	if s.closed {
		return fmt.Errorf("word: stream writer is closed")
	}
	return s.begin()
}

func (s *StreamWriter) begin() error {
	if s.xw != nil || s.err != nil {
		return s.err
	}
	if err := s.inner.prepare(); err != nil {
		s.err = err
		return err
	}
	s.cw = &countingWriter{w: s.dest}
	s.zw = common.NewZipWriter(s.cw)
	fw, err := s.zw.Create("word/document.xml")
	if err != nil {
		s.err = err
		return err
	}
	s.xw = common.NewXMLWriterTo(fw)
	s.inner.writeDocumentStart(s.xw)
	return nil
}
