package common

import (
	"bytes"
	"encoding/xml"
	"sync"
)

// maxPooledCap drops buffers larger than 1 MiB so the pool cannot retain
// a single huge document.xml for the rest of the process lifetime.
const maxPooledCap = 1 << 20

var (
	bufferPool = sync.Pool{
		New: func() any { return new(bytes.Buffer) },
	}
	xmlWriterPool = sync.Pool{
		New: func() any {
			return &XMLWriter{buf: new(bytes.Buffer)}
		},
	}
)

// GetBuffer returns a reset bytes.Buffer from the pool.
func GetBuffer() *bytes.Buffer {
	b := bufferPool.Get().(*bytes.Buffer)
	b.Reset()
	return b
}

// PutBuffer returns b to the pool. Buffers with Cap > 1 MiB are dropped.
func PutBuffer(b *bytes.Buffer) {
	if b == nil || b.Cap() > maxPooledCap {
		return
	}
	b.Reset()
	bufferPool.Put(b)
}

// GetXMLWriter returns a pooled XMLWriter bound to a reused bytes.Buffer.
// encoding/xml.Encoder has no Reset, so a new Encoder is created on the
// pooled buffer for each checkout.
func GetXMLWriter() *XMLWriter {
	w := xmlWriterPool.Get().(*XMLWriter)
	w.resetPooled()
	return w
}

// PutXMLWriter returns w to the pool. Destination-mode writers (NewXMLWriterTo)
// and writers whose buffer grew past 1 MiB are dropped.
func PutXMLWriter(w *XMLWriter) {
	if w == nil || w.dest != nil {
		return
	}
	if w.buf != nil && w.buf.Cap() > maxPooledCap {
		return
	}
	w.enc = nil
	w.stack = w.stack[:0]
	w.attrBuf = w.attrBuf[:0]
	if w.buf != nil {
		w.buf.Reset()
	}
	xmlWriterPool.Put(w)
}

func (w *XMLWriter) resetPooled() {
	if w.buf == nil {
		w.buf = new(bytes.Buffer)
	} else {
		w.buf.Reset()
	}
	w.dest = nil
	w.enc = xml.NewEncoder(w.buf)
	w.stack = w.stack[:0]
	w.attrBuf = w.attrBuf[:0]
}

// CloneBytes returns a copy of b that is safe to keep after PutBuffer/PutXMLWriter.
func CloneBytes(b []byte) []byte {
	if b == nil {
		return nil
	}
	out := make([]byte, len(b))
	copy(out, b)
	return out
}

// FinishXML flushes w, copies the document, and returns w to the pool.
func FinishXML(w *XMLWriter) []byte {
	if w == nil {
		return nil
	}
	out := CloneBytes(w.Bytes())
	PutXMLWriter(w)
	return out
}
