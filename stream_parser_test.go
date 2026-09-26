package word

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/yunkeweb/go-word/style"
)

func TestStreamExtractTextParagraphs(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	sec.AddText("Hello")
	sec.AddText("World")
	sec.AddTitle("Heading", 1)
	raw, err := doc.Bytes()
	if err != nil {
		t.Fatal(err)
	}

	var paras []string
	if err := StreamExtractText(bytes.NewReader(raw), func(s string) error {
		paras = append(paras, s)
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	joined := strings.Join(paras, "\n")
	for _, want := range []string{"Hello", "World", "Heading"} {
		if !strings.Contains(joined, want) {
			t.Fatalf("missing %q in %#v", want, paras)
		}
	}

	var viaDoc []string
	if err := New().StreamExtractText(bytes.NewReader(raw), func(s string) error {
		viaDoc = append(viaDoc, s)
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if strings.Join(viaDoc, "\n") != joined {
		t.Fatal("Document.StreamExtractText mismatch")
	}
}

func TestStreamExtractTextStopsOnCallbackError(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	sec.AddText("one")
	sec.AddText("two")
	sec.AddText("three")
	raw, err := doc.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	boom := errors.New("stop")
	n := 0
	err = StreamExtractText(bytes.NewReader(raw), func(s string) error {
		if s == "" {
			return nil
		}
		n++
		if n >= 2 {
			return boom
		}
		return nil
	})
	if !errors.Is(err, boom) {
		t.Fatalf("err=%v", err)
	}
}

func TestStreamExtractTextFromPlainReader(t *testing.T) {
	doc := New()
	doc.AddSection().AddText("piped")
	raw, err := doc.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	pr, pw := io.Pipe()
	go func() {
		_, _ = pw.Write(raw)
		_ = pw.Close()
	}()
	found := false
	if err := StreamExtractText(pr, func(s string) error {
		if s == "piped" {
			found = true
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if !found {
		t.Fatal("piped paragraph missing")
	}
}

func TestStreamExtractImages(t *testing.T) {
	png := pngBytes(t)
	doc := New()
	sec := doc.AddSection()
	sec.AddText("before")
	sec.AddImageBytes("dot.png", png, style.Image{Width: 40, Height: 20})
	sec.AddText("after")
	raw, err := doc.Bytes()
	if err != nil {
		t.Fatal(err)
	}

	var imgs []ImageFile
	if err := StreamExtractImages(bytes.NewReader(raw), func(img ImageFile) error {
		imgs = append(imgs, img)
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if len(imgs) != 1 {
		t.Fatalf("images=%d", len(imgs))
	}
	if imgs[0].MIME != "image/png" || !bytes.Equal(imgs[0].Data, png) {
		t.Fatalf("img=%+v", imgs[0])
	}
	if imgs[0].Name == "" {
		t.Fatal("empty image name")
	}

	n := 0
	if err := New().StreamExtractImages(bytes.NewReader(raw), func(img ImageFile) error {
		n++
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("doc method images=%d", n)
	}
}

func TestStreamExtractManyParagraphs(t *testing.T) {
	const n = 2000
	var buf bytes.Buffer
	sw := NewStreamWriter(&buf)
	for i := 0; i < n; i++ {
		if err := sw.WriteParagraph("p" + itoa(i)); err != nil {
			t.Fatal(err)
		}
	}
	if err := sw.Close(); err != nil {
		t.Fatal(err)
	}
	count := 0
	seenFirst, seenLast := false, false
	if err := StreamExtractText(bytes.NewReader(buf.Bytes()), func(s string) error {
		if s == "p0" {
			seenFirst = true
		}
		if s == "p"+itoa(n-1) {
			seenLast = true
		}
		if s != "" {
			count++
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if count < n || !seenFirst || !seenLast {
		t.Fatalf("count=%d first=%v last=%v", count, seenFirst, seenLast)
	}
}
