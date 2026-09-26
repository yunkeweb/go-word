package word

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"strings"
	"testing"

	"github.com/yunkeweb/go-word/style"
)

func testPNG(c color.RGBA) []byte {
	img := image.NewRGBA(image.Rect(0, 0, 8, 8))
	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			img.Set(x, y, c)
		}
	}
	var buf bytes.Buffer
	_ = png.Encode(&buf, img)
	return buf.Bytes()
}

func TestAppendDocumentRemapsStylesBookmarksAndMedia(t *testing.T) {
	dst := New()
	dst.AddParagraphStyle("Note", style.Paragraph{Alignment: style.JcLeft})
	sec := dst.AddSection()
	sec.AddText("Dest body", nil, "Note")
	sec.AddBookmark("shared")
	sec.AddImageBytes("a.png", testPNG(color.RGBA{R: 255, A: 255}), style.Image{Width: 8, Height: 8})

	src := New()
	src.AddParagraphStyle("Note", style.Paragraph{Alignment: style.JcCenter})
	ss := src.AddSection()
	ss.AddText("Src body", nil, "Note")
	ss.AddBookmark("shared")
	ss.AddBookmark("only-src")
	ss.AddImageBytes("b.png", testPNG(color.RGBA{G: 255, A: 255}), style.Image{Width: 8, Height: 8})
	ss.AddMath(`\frac{1}{2}`)

	if err := dst.AppendDocument(src, MergeOptions{StylePrefix: "src_", BookmarkPrefix: "src_", SectionBreak: "nextPage"}); err != nil {
		t.Fatal(err)
	}
	if len(dst.Sections()) != 2 {
		t.Fatalf("sections=%d", len(dst.Sections()))
	}
	raw, err := dst.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	docxml := readZipFile(t, raw, "word/document.xml")
	assertWellFormedXML(t, "merged", docxml)
	if !strings.Contains(docxml, `w:name="shared"`) {
		t.Fatal("kept dest bookmark")
	}
	if !strings.Contains(docxml, `w:name="src_shared"`) {
		t.Fatal("remapped src bookmark")
	}
	if !strings.Contains(docxml, `w:name="only-src"`) {
		t.Fatal("unique src bookmark kept")
	}
	if !strings.Contains(docxml, "Dest body") || !strings.Contains(docxml, "Src body") {
		t.Fatal("both bodies")
	}
	if !strings.Contains(docxml, "<m:f>") {
		t.Fatal("merged math")
	}
	styles := readZipFile(t, raw, "word/styles.xml")
	if !strings.Contains(styles, `w:styleId="Note"`) {
		t.Fatal("dest style")
	}
	if !strings.Contains(styles, `w:styleId="src_Note"`) {
		t.Fatal("remapped style")
	}
	_ = readZipFile(t, raw, "word/media/image1.png")
	_ = readZipFile(t, raw, "word/media/image2.png")
	if dst.GetStyle("src_Note") == nil {
		t.Fatal("src_Note registered")
	}
}

func TestAppendDocumentNilSource(t *testing.T) {
	doc := New()
	if err := doc.AppendDocument(nil, MergeOptions{}); err == nil {
		t.Fatal("expected error")
	}
}
