// Command v0.8.0_demo builds v080_demo.docx (OMML math, DrawingML shapes,
// multi-column layout) and v080_merged.docx (document merger).
package main

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log"

	"github.com/yunkeweb/go-word"
	"github.com/yunkeweb/go-word/style"
)

const (
	outFile   = "v080_demo.docx"
	mergedOut = "v080_merged.docx"
)

func main() {
	if err := writeFeatureDoc(); err != nil {
		log.Fatal(err)
	}
	if err := writeMergedDoc(); err != nil {
		log.Fatal(err)
	}
}

func writeFeatureDoc() error {
	doc := word.New()
	info := doc.GetDocInfo()
	info.Title = "GoWord v0.8.0 OMML, Shapes, Columns"
	info.Creator = "GoWord"
	info.Subject = "Office Math, DrawingML shapes, multi-column layout"
	info.Company = "yunkeweb"

	doc.SetDefaultFontName("Calibri")
	doc.SetDefaultAsianFontName("Microsoft YaHei")
	doc.SetDefaultFontSize(11)

	sec := doc.AddSection()
	sec.AddTitle("Office Math, DrawingML Shapes & Multi-Column Layout", 1)

	sec.AddTitle("Office Math (OMML)", 2)
	sec.AddText("Display fraction:")
	sec.AddMath(`\frac{a}{b}`)
	sec.AddText("Pythagorean identity:")
	sec.AddMath(`x^{2} + y^{2} = z^{2}`)
	tr := sec.AddTextRun()
	tr.AddText("Inline math: ")
	tr.AddMath(`\sqrt{x_1} + \pi`)

	sec.AddTitle("DrawingML Shapes", 2)
	doc.AddShape(word.ShapeRect, word.ShapeOptions{
		FillColor: "5B9BD5", LineColor: "2E75B6",
		Width: 1828800, Height: 914400,
	})
	doc.AddShape(word.ShapeRoundRect, word.ShapeOptions{
		FillColor: "ED7D31", LineColor: "C45911",
		Text: "Rounded", Font: style.Font{Bold: true, Color: "FFFFFF"},
	})
	doc.AddShape(word.ShapeArrow, word.ShapeOptions{
		FillColor: "70AD47", LineColor: "548235",
	})
	doc.AddShape(word.ShapeTextBox, word.ShapeOptions{
		FillColor: "FFF2CC", LineColor: "BF8F00",
		Text: "w:txbxContent", Font: style.Font{Size: 12, Color: "595959"},
	})

	col := doc.AddSection(style.Section{
		Orientation: style.OrientationPortrait,
		PageSizeW:   style.DefaultPageWidth,
		PageSizeH:   style.DefaultPageHeight,
		MarginTop:   1134, MarginBottom: 1134, MarginLeft: 1134, MarginRight: 1134,
	})
	col.SetColumns(2, 720, true)
	col.AddTitle("Two-column layout", 2)
	col.AddText("The left column starts here. Word splits this section into two equal columns with a separator line between them.")
	col.AddText("The second paragraph continues the flow so the columns fill naturally.")

	if err := doc.Save(outFile); err != nil {
		return err
	}
	fmt.Printf("wrote %s\n", outFile)
	return nil
}

func writeMergedDoc() error {
	dst := word.New()
	dst.AddParagraphStyle("Note", style.Paragraph{Spacing: style.Spacing{After: 120}})
	a := dst.AddSection()
	a.AddTitle("Document A", 1)
	a.AddText("Bookmark and image from A.", nil, "Note")
	a.AddBookmark("shared")
	a.AddImageBytes("a.png", demoPNG(color.RGBA{R: 31, G: 78, B: 121, A: 255}), style.Image{Width: 40, Height: 20})

	src := word.New()
	src.AddParagraphStyle("Note", style.Paragraph{Spacing: style.Spacing{After: 200}})
	b := src.AddSection()
	b.AddTitle("Document B", 1)
	b.AddText("Bookmark, image and formula from B.", nil, "Note")
	b.AddBookmark("shared")
	b.AddImageBytes("b.png", demoPNG(color.RGBA{R: 237, G: 125, B: 49, A: 255}), style.Image{Width: 40, Height: 20})
	b.AddMath(`\frac{1}{2}`)

	if err := dst.AppendDocument(src, word.MergeOptions{SectionBreak: "nextPage"}); err != nil {
		return err
	}
	if err := dst.Save(mergedOut); err != nil {
		return err
	}
	fmt.Printf("wrote %s\n", mergedOut)
	return nil
}

func demoPNG(c color.RGBA) []byte {
	img := image.NewRGBA(image.Rect(0, 0, 40, 20))
	for y := 0; y < 20; y++ {
		for x := 0; x < 40; x++ {
			img.Set(x, y, c)
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		log.Fatal(err)
	}
	return buf.Bytes()
}
