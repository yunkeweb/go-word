package word

import (
	"bytes"
	"errors"
	"io"
	"testing"
)

func TestReadOptionsLimitDocumentPart(t *testing.T) {
	doc := New()
	doc.AddSection().AddText("a document part large enough to exceed the configured budget")
	raw, err := doc.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := LoadBytesWithOptions(raw, ReadOptions{MaxPartSize: 32}); !errors.Is(err, ErrReadLimitExceeded) {
		t.Fatalf("LoadBytesWithOptions error = %v, want ErrReadLimitExceeded", err)
	}
	if _, err := LoadBytesWithOptions(raw, ReadOptions{}); err != nil {
		t.Fatalf("unlimited load failed: %v", err)
	}
}

func TestReadOptionsLimitTemplatePart(t *testing.T) {
	doc := New()
	doc.AddSection().AddText("template content large enough to exceed the configured budget")
	raw, err := doc.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := NewTemplateProcessorBytesWithOptions(raw, ReadOptions{MaxPartSize: 32}); !errors.Is(err, ErrReadLimitExceeded) {
		t.Fatalf("template error = %v, want ErrReadLimitExceeded", err)
	}
}

func TestReadOptionsLimitStreamImage(t *testing.T) {
	doc := New()
	doc.AddSection().AddImageBytes("large.png", pngBytes(t))
	raw, err := doc.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	if err := StreamExtractImagesWithOptions(bytes.NewReader(raw), func(ImageFile) error { return nil }, ReadOptions{MaxPartSize: 1}); !errors.Is(err, ErrReadLimitExceeded) {
		t.Fatalf("stream image error = %v, want ErrReadLimitExceeded", err)
	}
	if err := StreamExtractImagesWithOptions(bytes.NewReader(raw), func(ImageFile) error { return io.EOF }, ReadOptions{}); !errors.Is(err, io.EOF) {
		t.Fatalf("callback error = %v, want io.EOF", err)
	}
}
