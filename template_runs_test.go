package word

import (
	"archive/zip"
	"bytes"
	"strings"
	"testing"
)

func TestTemplateSetValueAcrossRuns(t *testing.T) {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	part := `<?xml version="1.0"?><w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main"><w:body><w:p><w:r><w:rPr><w:b/></w:rPr><w:t>${na</w:t></w:r><w:r><w:rPr><w:i/></w:rPr><w:t>me}</w:t></w:r></w:p></w:body></w:document>`
	w, err := zw.Create("word/document.xml")
	if err != nil {
		t.Fatal(err)
	}
	_, _ = w.Write([]byte(part))
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	tpl, err := NewTemplateProcessorBytes(buf.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	tpl.SetValue("name", "Ada")
	got := string(tpl.files["word/document.xml"])
	if !strings.Contains(got, "Ada") || strings.Contains(got, "${na") || strings.Contains(got, "me}") {
		t.Fatalf("cross-run replacement failed: %s", got)
	}
	if !strings.Contains(got, "<w:rPr><w:b/></w:rPr>") || !strings.Contains(got, "<w:rPr><w:i/></w:rPr>") {
		t.Fatal("run properties were not preserved")
	}
}
