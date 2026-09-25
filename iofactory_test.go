package word

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCreateWriterReaderAndLoad(t *testing.T) {
	doc := New()
	doc.AddSection().AddText("hi")
	w, err := CreateWriter(doc, "")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := CreateWriter(doc, "ODT"); err == nil {
		t.Fatal("invalid writer")
	}
	if _, err := CreateReader(""); err != nil {
		t.Fatal(err)
	}
	if _, err := CreateReader("ODT"); err == nil {
		t.Fatal("invalid reader")
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "a.docx")
	if err := w.Save(path); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(path); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(path, "Word2007"); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(path, ""); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(filepath.Join(dir, "missing.docx")); err == nil {
		t.Fatal("missing")
	}
	if _, err := Load(path, "ODT"); err == nil {
		t.Fatal("invalid reader name")
	}
	if err := doc.SaveAs(filepath.Join(dir, "b.docx"), "ODT"); err == nil {
		t.Fatal("bad format")
	}
	if err := doc.Save(""); err == nil {
		t.Fatal("empty filename")
	}
	varsPath := filepath.Join(dir, "v.docx")
	src := New()
	sec := src.AddSection()
	sec.AddText("Hello ${name}")
	sec.AddText("Again ${name}")
	sec.AddLink("https://x", "${title}")
	sec.AddTitle("${heading}", 1)
	sec.AddListItem("${item}", 0)
	if err := src.Save(varsPath); err != nil {
		t.Fatal(err)
	}
	vars, err := ExtractVariables(varsPath)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, v := range vars {
		if v == "name" {
			found = true
		}
	}
	if !found {
		t.Fatalf("vars=%v", vars)
	}
	if _, err := ExtractVariables(filepath.Join(dir, "no.docx")); err == nil {
		t.Fatal("extract missing")
	}
	_ = os.Remove
}

func TestPhpWordAlias(t *testing.T) {
	var d *PhpWord = New()
	if d == nil {
		t.Fatal("alias")
	}
}
