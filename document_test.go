package word

import (
	"archive/zip"
	"bytes"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yunkeweb/go-word/style"
)

func TestCreateSaveAndParts(t *testing.T) {
	doc := New()
	doc.AddFontStyle("rStyle", style.Font{Bold: true, Italic: true, Size: 16})
	doc.AddParagraphStyle("pStyle", style.Paragraph{Alignment: style.JcCenter, Spacing: style.Spacing{After: 100}})
	doc.AddTitleStyle(1, style.Font{Bold: true}, style.Paragraph{Spacing: style.Spacing{After: 240}})

	sec := doc.AddSection()
	sec.AddTitle("Welcome to GoWord", 1)
	sec.AddText("Hello World!")
	sec.AddTextBreak(1)
	sec.AddText("I am styled by a font style definition.", "rStyle")
	sec.AddText("I am styled by a paragraph style definition.", nil, "pStyle")

	tr := sec.AddTextRun()
	tr.AddText("I am inline styled ", style.Font{Name: "Times New Roman", Size: 20})
	tr.AddText("with ")
	tr.AddText("color", style.Font{Color: "996699"})
	tr.AddText(", ")
	tr.AddText("bold", style.Font{Bold: true})

	sec.AddLink("https://github.com/PHPOffice/PHPWord", "PHPWord on GitHub")

	tbl := sec.AddTable(style.Table{Width: 5000, Borders: style.Borders{
		Top: style.Border{Style: "single", Size: 4}, Left: style.Border{Style: "single", Size: 4},
		Right: style.Border{Style: "single", Size: 4}, Bottom: style.Border{Style: "single", Size: 4},
	}})
	row := tbl.AddRow()
	row.AddCell(2500).AddText("A")
	row.AddCell(2500).AddText("B")

	hdr := sec.AddHeader()
	hdr.AddText("header")
	ftr := sec.AddFooter()
	ftr.AddText("footer")

	sec.AddListItem("one", 0, nil, nil, style.ListTypeBullet)
	sec.AddListItem("two", 0)

	dir := t.TempDir()
	path := filepath.Join(dir, "hello.docx")
	if err := doc.Save(path); err != nil {
		t.Fatal(err)
	}
	st, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if st.Size() == 0 {
		t.Fatal("empty docx")
	}

	z, err := zip.OpenReader(path)
	if err != nil {
		t.Fatal(err)
	}
	defer z.Close()
	need := map[string]bool{
		"[Content_Types].xml":          false,
		"word/document.xml":            false,
		"_rels/.rels":                  false,
		"word/_rels/document.xml.rels": false,
		"word/styles.xml":              false,
	}
	for _, f := range z.File {
		if _, ok := need[f.Name]; ok {
			need[f.Name] = true
		}
	}
	for name, ok := range need {
		if !ok {
			t.Fatalf("missing part %s", name)
		}
	}

	raw := readZip(t, z, "word/document.xml")
	for _, s := range []string{"Hello World!", "PHPWord on GitHub", "Welcome to GoWord", "<w:tbl"} {
		if !strings.Contains(raw, s) {
			t.Fatalf("document.xml missing %q", s)
		}
	}
	if strings.Contains(raw, "<script>") {
		t.Fatal("unexpected raw markup")
	}
}

func TestRoundTripText(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	sec.AddText("Alpha")
	sec.AddText("Beta & Gamma")
	b, err := doc.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	got, err := LoadBytes(b)
	if err != nil {
		t.Fatal(err)
	}
	var texts []string
	for _, s := range got.Sections() {
		for _, e := range s.Elements() {
			if tx := elementText(e); tx != "" {
				texts = append(texts, tx)
			}
		}
	}
	joined := strings.Join(texts, " ")
	if !strings.Contains(joined, "Alpha") || !strings.Contains(joined, "Beta & Gamma") {
		t.Fatalf("round trip text = %q", joined)
	}
}

func TestXMLEscaping(t *testing.T) {
	doc := New()
	doc.AddSection().AddText(`a < b & c > "d"`)
	b, err := doc.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	z, err := zip.NewReader(bytes.NewReader(b), int64(len(b)))
	if err != nil {
		t.Fatal(err)
	}
	raw := readZipReader(t, z, "word/document.xml")
	if strings.Contains(raw, "a < b") {
		t.Fatal("unescaped < in XML")
	}
	if !strings.Contains(raw, "&lt;") || !strings.Contains(raw, "&amp;") {
		t.Fatalf("missing escapes: %s", raw)
	}
}

func TestImagePart(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 8, 8))
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}
	doc := New()
	sec := doc.AddSection()
	sec.AddImageBytes("dot.png", buf.Bytes(), style.Image{Width: 8, Height: 8})
	b, err := doc.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	z, err := zip.NewReader(bytes.NewReader(b), int64(len(b)))
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, f := range z.File {
		if strings.HasPrefix(f.Name, "word/media/") {
			found = true
		}
	}
	if !found {
		t.Fatal("image part missing")
	}
}

func readZip(t *testing.T, z *zip.ReadCloser, name string) string {
	t.Helper()
	for _, f := range z.File {
		if f.Name == name {
			rc, err := f.Open()
			if err != nil {
				t.Fatal(err)
			}
			defer rc.Close()
			var b bytes.Buffer
			if _, err := b.ReadFrom(rc); err != nil {
				t.Fatal(err)
			}
			return b.String()
		}
	}
	t.Fatalf("missing %s", name)
	return ""
}

func readZipReader(t *testing.T, z *zip.Reader, name string) string {
	t.Helper()
	for _, f := range z.File {
		if f.Name == name {
			rc, err := f.Open()
			if err != nil {
				t.Fatal(err)
			}
			defer rc.Close()
			var b bytes.Buffer
			if _, err := b.ReadFrom(rc); err != nil {
				t.Fatal(err)
			}
			return b.String()
		}
	}
	t.Fatalf("missing %s", name)
	return ""
}
