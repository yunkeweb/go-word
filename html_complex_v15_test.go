package word

import (
	"bytes"
	"strings"
	"testing"

	"github.com/yunkeweb/go-word/element"
	"github.com/yunkeweb/go-word/style"
)

func TestHTMLComplexObjectBaseline(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	shape := sec.AddDMLShape("roundRect", 1200, 800, "D9EAF7", "205493", 12)
	shape.AddText("shape text")
	box := sec.AddTextBox(style.TextBox{})
	box.AddText("box text")
	chart := sec.AddChart("column", []string{"Q1", "Q2"}, []float64{10, 20})
	chart.SeriesName = "Sales"
	img := element.NewImage("https://example.com/image.png", style.Image{AltText: "remote"})
	sec.AddTextRun().AppendElement(img)

	result, err := doc.RenderHTMLWithDiagnostics(HTMLOptions{})
	if err != nil {
		t.Fatal(err)
	}
	got := string(result.HTML)
	if !strings.Contains(got, "shape text") || !strings.Contains(got, "box text") {
		t.Fatalf("complex object text was lost: %s", got)
	}
	if !strings.Contains(got, "Sales") || !strings.Contains(got, "Q1") {
		t.Fatalf("chart is not readable in HTML: %s", got)
	}
	if strings.Contains(got, "https://example.com/image.png") {
		t.Fatalf("external image was emitted without an explicit callback: %s", got)
	}
}

func TestHTMLComplexObjectModel(t *testing.T) {
	doc := New()
	shape := doc.AddSection().AddDMLShape("triangle", 100, 200, "fff", "000", 1)
	shape.AddText("content")
	if shape.PrstGeom != "triangle" || shape.Width != 100 || len(shape.Elements()) != 1 {
		t.Fatalf("shape model=%+v elements=%d", shape, len(shape.Elements()))
	}
	chart := doc.AddSection().AddChart("line", []string{"A"}, []float64{1})
	chart.AddComboSeries("line", []string{"A"}, []float64{2}, "Margin", true)
	if len(chart.Series) != 2 || !chart.Series[1].SecondaryAxis {
		t.Fatalf("chart model=%+v", chart)
	}
}

func TestRenderHTMLChartAccessibility(t *testing.T) {
	doc := New()
	chart := doc.AddSection().AddChart("bar", []string{"A", "B"}, []float64{3, 4})
	chart.SeriesName = "Revenue"
	html, err := doc.RenderHTML(HTMLOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(html, []byte("<figure")) || !bytes.Contains(html, []byte("Revenue")) || !bytes.Contains(html, []byte("<table")) {
		t.Fatalf("chart output is not accessible: %s", html)
	}
}
