package word

import (
	"strings"
	"testing"

	"github.com/yunkeweb/go-word/style"
)

func TestSetColumnsWritesCols(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	sec.SetColumns(2, 720, true)
	sec.AddText("left and right")
	raw, err := doc.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	xml := readZipFile(t, raw, "word/document.xml")
	assertWellFormedXML(t, "cols", xml)
	if !strings.Contains(xml, `<w:cols w:num="2" w:space="720" w:sep="1"`) &&
		!strings.Contains(xml, `w:num="2"`) {
		t.Fatalf("cols attrs: %s", xml)
	}
	if !strings.Contains(xml, `w:sep="1"`) {
		t.Fatal("separator")
	}
	kids := innerChildNames(xml, "w:sectPr")
	assertSeq(t, kids, "w:pgSz", "w:pgMar", "w:cols", "w:docGrid")
}

func TestDrawingMLShapesAndTextBox(t *testing.T) {
	doc := New()
	doc.AddSection()
	doc.AddShape(ShapeRect, ShapeOptions{FillColor: "5B9BD5", LineColor: "2E75B6", Width: 1828800, Height: 914400})
	doc.AddShape(ShapeRoundRect, ShapeOptions{FillColor: "ED7D31", Text: "Round"})
	doc.AddShape(ShapeArrow, ShapeOptions{FillColor: "70AD47"})
	doc.AddShape(ShapeTextBox, ShapeOptions{
		FillColor: "FFF2CC",
		LineColor: "BF8F00",
		Text:      "Text box",
		Font:      style.Font{Bold: true, Size: 12},
	})
	raw, err := doc.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	xml := readZipFile(t, raw, "word/document.xml")
	assertWellFormedXML(t, "shapes", xml)
	if !strings.Contains(xml, `xmlns:wps="`+`http://schemas.microsoft.com/office/word/2010/wordprocessingShape"`) {
		t.Fatal("wps ns")
	}
	if !strings.Contains(xml, `<a:prstGeom prst="rect"`) {
		t.Fatal("rect geom")
	}
	if !strings.Contains(xml, `<a:prstGeom prst="roundRect"`) {
		t.Fatal("roundRect geom")
	}
	if !strings.Contains(xml, `<a:prstGeom prst="rightArrow"`) {
		t.Fatal("arrow geom")
	}
	if !strings.Contains(xml, "<w:txbxContent>") {
		t.Fatal("txbxContent")
	}
	if !strings.Contains(xml, "Text box") {
		t.Fatal("textbox text")
	}
	if !strings.Contains(xml, "<wps:wsp>") {
		t.Fatal("wps:wsp")
	}
}
