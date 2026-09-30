package word

import (
	"bytes"
	"image"
	"image/png"
	"io"
	"strconv"
	"strings"
	"testing"

	"github.com/yunkeweb/go-word/element"
	"github.com/yunkeweb/go-word/pkg/common"
)

func TestWriterRebuildsRelationshipLookups(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	img := sec.AddImageBytes("test.png", pngBytes(t))
	chart := sec.AddChart("pie", []string{"A"}, []float64{1})
	ole := sec.AddOLEObject("")
	ole.Media.Data = []byte("payload")
	link := sec.AddLink("https://example.com", "link")
	header, footer := sec.AddHeader(), sec.AddFooter()
	w := newWord2007Writer(doc)
	for pass := 0; pass < 2; pass++ {
		if _, err := w.WriteTo(io.Discard); err != nil {
			t.Fatal(err)
		}
		for _, el := range []element.Element{img, chart, ole, link, header, footer} {
			if w.relFor(el) == "" {
				t.Fatalf("pass %d: no relationship for %T", pass, el)
			}
		}
	}
	for _, el := range []element.Element{img, chart, ole, link} {
		sec.RemoveElement(el)
	}
	sec.Headers, sec.Footers = nil, nil
	if _, err := w.WriteTo(io.Discard); err != nil {
		t.Fatal(err)
	}
	for _, el := range []element.Element{img, chart, ole, link, header, footer, nil} {
		if id := w.relFor(el); id != "" {
			t.Fatalf("removed %T still has relationship %q", el, id)
		}
	}
}

func BenchmarkReadDocumentXML(b *testing.B) {
	paragraph := "<w:p><w:r><w:t>" + strings.Repeat("text ", 1024) + "</w:t></w:r></w:p>"
	data := []byte(`<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main"><w:body>` + strings.Repeat(paragraph, 100) + `</w:body></w:document>`)
	b.ReportAllocs()
	b.SetBytes(int64(len(data)))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := parseDocumentXML(data, New().AddSection()); err != nil {
			b.Fatal(err)
		}
	}
}

func benchmarkImageWriter(b *testing.B, count int) (*word2007Writer, []*element.Image) {
	b.Helper()
	var buf bytes.Buffer
	if err := png.Encode(&buf, image.NewRGBA(image.Rect(0, 0, 1, 1))); err != nil {
		b.Fatal(err)
	}
	doc := New()
	sec := doc.AddSection()
	images := make([]*element.Image, count)
	for i := range images {
		images[i] = sec.AddImageBytes("test.png", buf.Bytes())
	}
	w := newWord2007Writer(doc)
	if err := w.prepare(); err != nil {
		b.Fatal(err)
	}
	return w, images
}

func BenchmarkRelationshipLookup(b *testing.B) {
	for _, count := range []int{32, 1024, 4096} {
		b.Run(strconv.Itoa(count), func(b *testing.B) {
			w, images := benchmarkImageWriter(b, count)
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if w.relFor(images[i%len(images)]) == "" {
					b.Fatal("missing image relationship")
				}
			}
		})
	}
}

func BenchmarkRenderImageDocument(b *testing.B) {
	w, _ := benchmarkImageWriter(b, 2048)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := w.prepare(); err != nil {
			b.Fatal(err)
		}
		xw := common.NewXMLWriterTo(io.Discard)
		w.writeDocument(xw)
		if err := xw.Flush(); err != nil {
			b.Fatal(err)
		}
	}
}
