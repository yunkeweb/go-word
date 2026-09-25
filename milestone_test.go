package word

import (
	"archive/zip"
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/yunkeweb/go-word/element"
	"github.com/yunkeweb/go-word/style"
)

func zipText(t *testing.T, raw []byte) string {
	t.Helper()
	zr, err := zip.NewReader(bytes.NewReader(raw), int64(len(raw)))
	if err != nil {
		t.Fatal(err)
	}
	var b strings.Builder
	for _, f := range zr.File {
		b.WriteString(f.Name)
		b.WriteByte('\n')
		rc, err := f.Open()
		if err != nil {
			t.Fatal(err)
		}
		_, _ = io.Copy(&b, rc)
		_ = rc.Close()
	}
	return b.String()
}

func TestDocumentAddChartBarLinePie(t *testing.T) {
	doc := New()
	cats := []string{"Q1", "Q2", "Q3"}
	bar := doc.AddChart(ChartTypeBar, cats, []float64{1, 2, 3}, []float64{4, 5, 6})
	if bar.ChartType != ChartTypeBar || len(bar.Series) != 2 {
		t.Fatalf("bar series=%d type=%s", len(bar.Series), bar.ChartType)
	}
	doc.AddChart(ChartTypeLine, cats, []float64{3, 1, 2})
	doc.AddChart(ChartTypePie, []string{"A", "B"}, []float64{70, 30})
	doc.AddChartStyled(ChartTypeColumn, cats, []float64{1, 1, 1}, style.Chart{Title: "Cols", ShowLegend: true})

	raw, err := doc.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	s := zipText(t, raw)
	for _, need := range []string{
		"word/charts/chart1.xml",
		"word/charts/chart2.xml",
		"word/charts/chart3.xml",
		"word/charts/chart4.xml",
		"c:barChart",
		"c:lineChart",
		"c:pieChart",
		"application/vnd.openxmlformats-officedocument.drawingml.chart+xml",
		"/word/charts/chart1.xml",
	} {
		if !strings.Contains(s, need) {
			t.Fatalf("missing %s", need)
		}
	}
	if !strings.Contains(s, "http://schemas.openxmlformats.org/officeDocument/2006/relationships/chart") {
		t.Fatal("chart rel missing")
	}
}

func TestDocumentAddMarkdown(t *testing.T) {
	doc := New()
	md := "# Title\n\nHello **bold** and *italic* and ++under++ and `code`.\n\n" +
		"- one\n- two\n\n1. first\n\n```\nfunc main() {}\n```\n\n" +
		"## Sub\n\n[link](https://example.com)\n\n---\n\n<u>u</u>\n"
	if err := doc.AddMarkdown(md); err != nil {
		t.Fatal(err)
	}
	sec := doc.GetSection(0)
	var titles, lists, texts int
	var sawCode, sawLink bool
	for _, e := range sec.Elements() {
		switch v := e.(type) {
		case *element.Title:
			titles++
			if v.Depth == 1 && v.Text != "Title" {
				t.Fatalf("h1=%q", v.Text)
			}
		case *element.ListItem:
			lists++
		case *element.Text:
			texts++
			if strings.Contains(v.Content, "func main") {
				sawCode = true
			}
		case *element.TextRun:
			if strings.Contains(v.GetText(), "bold") {
				texts++
			}
			for _, c := range v.Elements() {
				if _, ok := c.(*element.Link); ok {
					sawLink = true
				}
			}
		}
	}
	if titles < 2 || lists < 2 || !sawCode || !sawLink {
		t.Fatalf("titles=%d lists=%d code=%v link=%v", titles, lists, sawCode, sawLink)
	}
	raw, err := doc.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	out := zipText(t, raw)
	if !strings.Contains(out, "Title") || !strings.Contains(out, "bold") {
		t.Fatal("markdown not serialized")
	}
}

func TestDocumentAddHTMLWrapperAndCode(t *testing.T) {
	doc := New()
	if err := doc.AddHTML(`<h1>Hi</h1><p>A <code>x</code> B</p><pre>line1
line2</pre><blockquote>q</blockquote>`); err != nil {
		t.Fatal(err)
	}
	if doc.GetSection(0).CountElements() == 0 {
		t.Fatal("no html elements")
	}
	raw, err := doc.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	out := zipText(t, raw)
	if !strings.Contains(out, "Hi") || !strings.Contains(out, "line1") || !strings.Contains(out, "line2") {
		t.Fatal("html missing content")
	}
	if !strings.Contains(out, "Courier New") {
		t.Fatal("html code font")
	}
}

func TestCommentsAndTrackChanges(t *testing.T) {
	doc := New()
	doc.EnableTrackChanges(true)
	if !doc.TrackRevisions() {
		t.Fatal("track revisions flag")
	}
	tx := doc.CommentOn("visible", "please review", "Ann", "A", "2020-01-01T00:00:00Z")
	if tx.GetCommentRangeStart() == nil || tx.GetCommentRangeStart().Author != "Ann" {
		t.Fatal("comment range")
	}
	if len(doc.GetComments()) != 1 {
		t.Fatalf("comments=%d", len(doc.GetComments()))
	}
	doc.AddInsertion("added", "Bob", "2020-02-02T00:00:00Z")
	doc.AddDeletion("removed", "Bob", "2020-02-02T00:00:00Z")

	raw, err := doc.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	s := zipText(t, raw)
	for _, need := range []string{
		"word/comments.xml",
		"application/vnd.openxmlformats-officedocument.wordprocessingml.comments+xml",
		`w:author="Ann"`,
		"please review",
		"<w:ins",
		"<w:del",
		"<w:delText",
		"added",
		"removed",
		"w:trackRevisions",
		"w:commentRangeStart",
		"w:commentRangeEnd",
		"w:commentReference",
	} {
		if !strings.Contains(s, need) {
			t.Fatalf("missing %q", need)
		}
	}
}

func TestMarkdownHelpers(t *testing.T) {
	if d, title, ok := markdownHeading("### Hello"); !ok || d != 3 || title != "Hello" {
		t.Fatal("heading")
	}
	if _, _, ok := markdownHeading("####### x"); ok {
		t.Fatal("h7")
	}
	if !isMarkdownHR("---") || !isMarkdownHR("* * *") {
		t.Fatal("hr")
	}
	text, depth, ordered, ok := markdownList("  1. item")
	if !ok || text != "item" || depth != 1 || !ordered {
		t.Fatalf("list %q %d %v", text, depth, ordered)
	}
	if markdownPlain("**a**") != "a" {
		t.Fatal("plain")
	}
}

func TestAddMarkdownEmptyAndUnclosedFence(t *testing.T) {
	doc := New()
	if err := doc.AddMarkdown(""); err != nil {
		t.Fatal(err)
	}
	if err := doc.AddMarkdown("```\nstill open"); err != nil {
		t.Fatal(err)
	}
	raw, err := doc.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(zipText(t, raw), "still open") {
		t.Fatal("unclosed fence")
	}
}
