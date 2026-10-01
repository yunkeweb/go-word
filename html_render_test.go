package word

import (
	"bytes"
	"strings"
	"sync"
	"testing"

	"github.com/yunkeweb/go-word/element"
	"github.com/yunkeweb/go-word/style"
)

func TestRenderHTMLBasicDocument(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	sec.AddTitle("A <title>", 2)
	run := sec.AddTextRun(style.Paragraph{Alignment: style.JcCenter})
	run.AddText("Hello & <world>", style.Font{Bold: true, Color: "#123456"})
	run.AddLink("https://example.test/?a=1&b=2", "link & text")
	got, err := doc.RenderHTML(HTMLOptions{Standalone: true, IncludeCSS: true})
	if err != nil {
		t.Fatal(err)
	}
	s := string(got)
	for _, want := range []string{"<!doctype html>", "<h2>A &lt;title&gt;</h2>", "Hello &amp; &lt;world&gt;", "font-weight:700", "href=\"https://example.test/?a=1&amp;b=2\"", "<style>"} {
		if !strings.Contains(s, want) {
			t.Fatalf("HTML missing %q: %s", want, s)
		}
	}
}

func TestRenderHTMLListsTablesAndImage(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	sec.AddListItem("one", 0, style.Font{}, nil, style.ListItem{ListType: style.ListTypeBullet})
	sec.AddListItem("two", 0, style.Font{}, nil, style.ListItem{ListType: style.ListTypeBullet})
	tbl := sec.AddTable(style.Table{Width: 1500})
	row := tbl.AddRow()
	row.AddCell(750).AddText("A")
	row.AddCell(750).AddText("B")
	sec.AddImageBytes("pixel.png", []byte{0x89, 0x50, 0x4e, 0x47}, style.Image{AltText: "pixel", Width: 20})
	got, err := doc.RenderHTML(HTMLOptions{})
	if err != nil {
		t.Fatal(err)
	}
	s := string(got)
	for _, want := range []string{"<ul>", "<li>one", "<table", "<td", "data:image/png;base64,", "alt=\"pixel\"", "width=\"20\""} {
		if !strings.Contains(s, want) {
			t.Fatalf("HTML missing %q: %s", want, s)
		}
	}
}

func TestRenderHTMLStrictDiagnostics(t *testing.T) {
	doc := New()
	doc.AddSection().AddChart("bar", []string{"A"}, []float64{1})
	res, err := doc.RenderHTMLWithDiagnostics(HTMLOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Diagnostics) != 1 || res.Diagnostics[0].ElementType != "Chart" {
		t.Fatalf("diagnostics=%+v", res.Diagnostics)
	}
	if !strings.Contains(string(res.HTML), "data-element=\"Chart\"") {
		t.Fatalf("fallback missing: %s", res.HTML)
	}
	if _, err := doc.RenderHTML(HTMLOptions{Strict: true}); err == nil {
		t.Fatal("strict mode accepted unsupported chart")
	}
}

func TestRenderHTMLReaderEntrypoints(t *testing.T) {
	doc := New()
	doc.AddSection().AddText("reader entry")
	raw, err := doc.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	got, err := RenderHTML(bytes.NewReader(raw), HTMLOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(got), "reader entry") {
		t.Fatalf("reader output=%s", got)
	}
}

func TestRenderHTMLRoundTripFromDocx(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	sec.AddTitle("Roundtrip", 1)
	sec.AddText("paragraph")
	sec.AddListItem("item", 0, style.Font{}, nil, style.ListItem{ListType: style.ListTypeBullet})
	tbl := sec.AddTable()
	row := tbl.AddRow()
	row.AddCell(900).AddText("cell")
	sec.AddImageBytes("pixel.png", []byte{1, 2, 3}, style.Image{AltText: "roundtrip"})
	raw, err := doc.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	loaded, err := LoadBytes(raw)
	if err != nil {
		t.Fatal(err)
	}
	got, err := loaded.RenderHTML(HTMLOptions{})
	if err != nil {
		t.Fatal(err)
	}
	s := string(got)
	for _, want := range []string{"Roundtrip", "paragraph", "<ul>", "cell", "data:image/"} {
		if !strings.Contains(s, want) {
			t.Fatalf("roundtrip HTML missing %q: %s", want, s)
		}
	}
}

func TestRenderHTMLImageURLCallback(t *testing.T) {
	doc := New()
	img := doc.AddSection().AddImageBytes("pixel.png", []byte{1, 2, 3}, nil)
	got, err := doc.RenderHTML(HTMLOptions{ImageURL: func(got *element.Image) (string, error) {
		if got != img {
			t.Fatal("unexpected image callback")
		}
		return "/assets/pixel.png", nil
	}})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(got), "src=\"/assets/pixel.png\"") {
		t.Fatalf("callback URL missing: %s", got)
	}
}

func TestRenderHTMLImageURLCallbackRejectsUnsafeURL(t *testing.T) {
	doc := New()
	doc.AddSection().AddImageBytes("pixel.png", []byte{1, 2, 3}, nil)
	res, err := doc.RenderHTMLWithDiagnostics(HTMLOptions{ImageURL: func(*element.Image) (string, error) {
		return "javascript:alert(1)", nil
	}})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Diagnostics) != 1 || !strings.Contains(res.Diagnostics[0].Message, "unsafe URL") {
		t.Fatalf("diagnostics=%+v", res.Diagnostics)
	}
	if strings.Contains(string(res.HTML), "javascript:") {
		t.Fatalf("unsafe callback URL leaked: %s", res.HTML)
	}
}

func TestRenderHTMLSanitizesURLsAndSupportsHeaders(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	sec.AddLink("javascript:alert(1)", "unsafe")
	sec.AddHeader().AddText("Header")
	sec.AddFooter().AddText("Footer")
	got, err := doc.RenderHTML(HTMLOptions{IncludeHeadersFooters: true})
	if err != nil {
		t.Fatal(err)
	}
	s := string(got)
	if strings.Contains(s, "javascript:") || !strings.Contains(s, "Header") || !strings.Contains(s, "Footer") {
		t.Fatalf("unsafe or header output: %s", s)
	}
}

func TestRenderHTMLConcurrentIsolation(t *testing.T) {
	doc := New()
	doc.AddSection().AddText("concurrent")
	var wg sync.WaitGroup
	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			got, err := doc.RenderHTML(HTMLOptions{Standalone: true})
			if err != nil || !strings.Contains(string(got), "concurrent") {
				t.Errorf("render failed: %v %s", err, got)
			}
		}()
	}
	wg.Wait()
}

func TestRenderHTMLWithOptionsAndWriteHTML(t *testing.T) {
	doc := New()
	doc.AddSection().AddText("limited")
	raw, err := doc.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	got, err := RenderHTMLWithOptions(bytes.NewReader(raw), ReadOptions{MaxArchiveSize: int64(len(raw) + 1)}, HTMLOptions{})
	if err != nil || !strings.Contains(string(got), "limited") {
		t.Fatalf("limited render failed: %v %s", err, got)
	}
	var out bytes.Buffer
	if err := doc.WriteHTML(&out, HTMLOptions{Standalone: true}); err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(out.String(), "<!doctype html>") {
		t.Fatalf("stream output=%s", out.String())
	}
}
