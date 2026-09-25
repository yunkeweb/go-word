package common

import (
	"archive/zip"
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestZipWriterReader(t *testing.T) {
	var buf bytes.Buffer
	zw := NewZipWriter(&buf)
	if err := zw.AddFile(`word\media\a.png`, []byte("img")); err != nil {
		t.Fatal(err)
	}
	ew, err := zw.Create("word/document.xml")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ew.Write([]byte("<w:document/>")); err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	zr, err := OpenZipBytes(buf.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	defer zr.Close()
	if !zr.Has("word/media/a.png") {
		t.Fatal("Has")
	}
	if zr.Has("missing") {
		t.Fatal("Has missing")
	}
	b, err := zr.ReadFile("word/media/a.png")
	if err != nil || string(b) != "img" {
		t.Fatalf("ReadFile=%q %v", b, err)
	}
	if _, err := zr.ReadFile("nope"); err == nil {
		t.Fatal("missing file")
	}
	if len(zr.Files()) != 2 {
		t.Fatalf("files=%d", len(zr.Files()))
	}
	doc, err := zr.ReadFile("word/document.xml")
	if err != nil || string(doc) != "<w:document/>" {
		t.Fatalf("document.xml=%q %v", doc, err)
	}
}

func TestOpenZipFileAndErrors(t *testing.T) {
	dir := t.TempDir()
	var buf bytes.Buffer
	zw := NewZipWriter(&buf)
	if err := zw.AddFile("a.xml", []byte("<a/>")); err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "x.zip")
	if err := os.WriteFile(path, buf.Bytes(), 0o644); err != nil {
		t.Fatal(err)
	}
	zr, err := OpenZipFile(path)
	if err != nil {
		t.Fatal(err)
	}
	b, err := zr.ReadFile("a.xml")
	if err != nil || string(b) != "<a/>" {
		t.Fatalf("%q %v", b, err)
	}
	if err := zr.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := OpenZipFile(filepath.Join(dir, "missing.zip")); err == nil {
		t.Fatal("missing zip")
	}
	if _, err := OpenZipBytes([]byte("not a zip")); err == nil {
		t.Fatal("bad bytes")
	}
	notZip := filepath.Join(dir, "empty.zip")
	if err := os.WriteFile(notZip, []byte("xxxx"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := OpenZipFile(notZip); err == nil {
		t.Fatal("invalid zip file")
	}

	var buf2 bytes.Buffer
	zw2 := NewZipWriter(&buf2)
	_ = zw2.Close()
	_ = zw2.AddFile("x", []byte("y"))

	zw3 := NewZipWriter(errWriter{})
	_ = zw3.AddFile("x", []byte("y"))
}

type errWriter struct{}

func (errWriter) Write([]byte) (int, error) { return 0, os.ErrClosed }

func TestZipReadFileOpenError(t *testing.T) {
	raw := makeEncryptedZip(t, "a.txt", []byte("hello"))
	zr, err := OpenZipBytes(raw)
	if err != nil {
		t.Fatal(err)
	}
	defer zr.Close()
	if _, err := zr.ReadFile("a.txt"); err == nil {
		t.Fatal("expected open error")
	}

	zw3 := NewZipWriter(errWriter{})
	_ = zw3.AddFile("big.bin", bytes.Repeat([]byte("x"), 1<<20))

	var buf bytes.Buffer
	zw4 := NewZipWriter(&buf)
	if err := zw4.AddFile(strings.Repeat("a", 70000), []byte("x")); err == nil {
		t.Fatal("expected long name error")
	}
}

func makeEncryptedZip(t *testing.T, name string, data []byte) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	w, err := zw.Create(name)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := w.Write(data); err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	return setZipUnknownMethod(buf.Bytes())
}

func setZipUnknownMethod(b []byte) []byte {
	out := append([]byte(nil), b...)
	for i := 0; i+12 < len(out); i++ {
		if out[i] == 'P' && out[i+1] == 'K' && out[i+2] == 3 && out[i+3] == 4 {
			out[i+8] = 99
			out[i+9] = 0
		}
		if out[i] == 'P' && out[i+1] == 'K' && out[i+2] == 1 && out[i+3] == 2 {
			out[i+10] = 99
			out[i+11] = 0
		}
	}
	return out
}
