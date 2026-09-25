package common

import (
	"bytes"
	"io"
	"strings"
	"sync"
	"testing"
)

func TestBufferPoolRoundTrip(t *testing.T) {
	b := GetBuffer()
	b.WriteString("hello")
	if b.String() != "hello" {
		t.Fatalf("%q", b.String())
	}
	PutBuffer(b)
	b2 := GetBuffer()
	if b2.Len() != 0 {
		t.Fatalf("reset failed: %q", b2.String())
	}
	PutBuffer(b2)
	PutBuffer(nil)

	huge := GetBuffer()
	huge.Grow(maxPooledCap + 64)
	huge.Write(make([]byte, maxPooledCap+1))
	PutBuffer(huge) // dropped
}

func TestXMLWriterPoolReuse(t *testing.T) {
	w := GetXMLWriter()
	w.StartDocument()
	w.Start("root")
	w.Text("one")
	w.End()
	first := string(w.Bytes())
	if !strings.Contains(first, "one") {
		t.Fatal(first)
	}
	PutXMLWriter(w)

	w = GetXMLWriter()
	w.Start("root")
	w.Text("two")
	w.End()
	second := string(FinishXML(w))
	if strings.Contains(second, "one") {
		t.Fatalf("pool leak: %s", second)
	}
	if !strings.Contains(second, "two") {
		t.Fatal(second)
	}
	if FinishXML(nil) != nil {
		t.Fatal("nil FinishXML")
	}
	PutXMLWriter(nil)
}

func TestXMLWriterDestMode(t *testing.T) {
	var buf bytes.Buffer
	w := NewXMLWriterTo(&buf)
	w.StartDocument()
	w.Start("root", "a", "1", "skip", "")
	w.Raw("<!--raw-->")
	w.Element("t", "hi")
	w.End()
	if err := w.Flush(); err != nil {
		t.Fatal(err)
	}
	s := buf.String()
	if !strings.Contains(s, "<?xml") || !strings.Contains(s, "<!--raw-->") || !strings.Contains(s, "hi") {
		t.Fatal(s)
	}
	if w.Bytes() != nil {
		t.Fatal("dest Bytes should be nil")
	}
	PutXMLWriter(w) // dest-mode dropped
	var empty XMLWriter
	if err := empty.Flush(); err != nil {
		t.Fatal(err)
	}
}

func TestXMLWriterPoolDropsHugeAndDest(t *testing.T) {
	w := GetXMLWriter()
	w.buf.Grow(maxPooledCap + 8)
	w.buf.Write(make([]byte, maxPooledCap+1))
	PutXMLWriter(w)

	dest := NewXMLWriterTo(io.Discard)
	PutXMLWriter(dest)
}

func TestCloneBytes(t *testing.T) {
	if CloneBytes(nil) != nil {
		t.Fatal("nil")
	}
	src := []byte("ab")
	out := CloneBytes(src)
	src[0] = 'x'
	if string(out) != "ab" {
		t.Fatal(out)
	}
}

func TestXMLWriterPoolConcurrent(t *testing.T) {
	var wg sync.WaitGroup
	for i := 0; i < 32; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 20; j++ {
				w := GetXMLWriter()
				w.Start("n")
				w.Text("x")
				w.End()
				_ = w.Bytes()
				PutXMLWriter(w)
				b := GetBuffer()
				b.WriteByte('a')
				PutBuffer(b)
			}
		}()
	}
	wg.Wait()
}
