package word

import (
	"archive/zip"
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/yunkeweb/go-word/element"
)

func TestOpenReadExtractTextAndMetadata(t *testing.T) {
	doc := New()
	info := doc.GetDocInfo()
	info.Title = "Parse Me"
	info.Creator = "Alice"
	info.Subject = "inspection"
	created := time.Date(2024, 3, 4, 5, 6, 7, 0, time.UTC)
	info.Created = created
	info.Modified = created.Add(time.Hour)
	info.LastModifiedBy = "Bob"

	sec := doc.AddSection()
	sec.AddTitle("Heading One", 1)
	sec.AddText("Hello paragraph")
	tbl := sec.AddTable()
	r := tbl.AddRow()
	r.AddCell(2000).AddText("A1")
	r.AddCell(2000).AddText("A2")
	r2 := tbl.AddRow()
	r2.AddCell(2000).AddText("B1")
	r2.AddCell(2000).AddText("B2")
	sec.AddText("After table")

	raw, err := doc.Bytes()
	if err != nil {
		t.Fatal(err)
	}

	got, err := Read(bytes.NewReader(raw))
	if err != nil {
		t.Fatal(err)
	}
	text := got.ExtractText()
	for _, want := range []string{"Heading One", "Hello paragraph", "A1", "A2", "B1", "B2", "After table"} {
		if !strings.Contains(text, want) {
			t.Fatalf("ExtractText missing %q in %q", want, text)
		}
	}
	if !strings.Contains(text, "A1\tA2") {
		t.Fatalf("table cells not tab-separated: %q", text)
	}
	meta := got.GetMetadata()
	if meta == nil || meta.Creator != "Alice" || meta.Title != "Parse Me" || meta.LastModifiedBy != "Bob" {
		t.Fatalf("metadata=%+v", meta)
	}
	if !meta.Created.Equal(created) {
		t.Fatalf("created=%v want %v", meta.Created, created)
	}

	dir := t.TempDir()
	path := filepath.Join(dir, "sample.docx")
	if err := os.WriteFile(path, raw, 0o644); err != nil {
		t.Fatal(err)
	}
	opened, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(opened.ExtractText(), "Hello paragraph") {
		t.Fatal("Open ExtractText")
	}
}

func TestExtractImagesFromMedia(t *testing.T) {
	png := pngBytes(t)
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	w, err := zw.Create("word/document.xml")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := w.Write([]byte(`<?xml version="1.0"?><w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main"><w:body><w:p><w:r><w:t>pic</w:t></w:r></w:p></w:body></w:document>`)); err != nil {
		t.Fatal(err)
	}
	mw, err := zw.Create("word/media/image1.png")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := mw.Write(png); err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}

	doc, err := LoadBytes(buf.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	imgs, err := doc.ExtractImages()
	if err != nil {
		t.Fatal(err)
	}
	if len(imgs) != 1 {
		t.Fatalf("images=%d", len(imgs))
	}
	if imgs[0].Name != "image1.png" || imgs[0].MIME != "image/png" {
		t.Fatalf("file=%+v", imgs[0])
	}
	if !bytes.Equal(imgs[0].Data, png) {
		t.Fatal("image bytes")
	}
}

func TestExtractImagesFromMemoryDocument(t *testing.T) {
	doc := New()
	png := pngBytes(t)
	doc.AddSection().AddImageBytes("dot.png", png)
	imgs, err := doc.ExtractImages()
	if err != nil {
		t.Fatal(err)
	}
	if len(imgs) != 1 || imgs[0].MIME != "image/png" || !bytes.Equal(imgs[0].Data, png) {
		t.Fatalf("imgs=%+v", imgs)
	}
}

func TestParseNestedTableAndBookmarkDOM(t *testing.T) {
	xml := `<?xml version="1.0"?><w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
	<w:body>
		<w:p><w:r><w:t>outer</w:t></w:r></w:p>
		<w:bookmarkStart w:id="1" w:name="here"/><w:bookmarkEnd w:id="1"/>
		<w:tbl>
			<w:tr>
				<w:tc>
					<w:tcPr><w:vAlign w:val="center"/><w:textDirection w:val="tbRl"/></w:tcPr>
					<w:p><w:r><w:t>parent</w:t></w:r></w:p>
					<w:tbl>
						<w:tr><w:tc><w:p><w:r><w:t>inner</w:t></w:r></w:p></w:tc></w:tr>
					</w:tbl>
					<w:p/>
				</w:tc>
			</w:tr>
		</w:tbl>
		<w:p><w:hyperlink w:anchor="here"><w:r><w:t>jump</w:t></w:r></w:hyperlink></w:p>
	</w:body></w:document>`
	doc, err := LoadBytes(zipDocXML(t, xml))
	if err != nil {
		t.Fatal(err)
	}
	text := doc.ExtractText()
	if !strings.Contains(text, "outer") || !strings.Contains(text, "parent") || !strings.Contains(text, "inner") {
		t.Fatalf("text=%q", text)
	}
	sec := doc.GetSection(0)
	var nested bool
	var bm int
	var internalLink bool
	walkDocument(doc, func(el element.Element) {
		if b, ok := el.(*element.Bookmark); ok && b.Name == "here" {
			bm++
		}
		if l, ok := el.(*element.Link); ok && l.Internal && l.Target == "here" {
			internalLink = true
		}
		if tbl, ok := el.(*element.Table); ok {
			for _, row := range tbl.Rows {
				for _, cell := range row.Cells {
					for _, child := range cell.Elements() {
						if _, ok := child.(*element.Table); ok {
							nested = true
						}
					}
				}
			}
		}
	})
	if !nested {
		t.Fatal("nested table not attached to cell")
	}
	if bm == 0 {
		t.Fatal("bookmark not parsed")
	}
	if !internalLink {
		t.Fatal("anchor hyperlink not parsed")
	}
	_ = sec
}

func TestGetMetadataNilAndExtractTextEmpty(t *testing.T) {
	var d *Document
	if d.ExtractText() != "" {
		t.Fatal("nil ExtractText")
	}
	if d.GetMetadata() != nil {
		t.Fatal("nil metadata")
	}
	imgs, err := d.ExtractImages()
	if err != nil || imgs != nil {
		t.Fatal("nil images")
	}
}
