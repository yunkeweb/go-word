// Command v0.6.0_demo builds parsing_demo.docx: nested tables, cell styling,
// bookmarks with internal hyperlinks, then re-opens the package and prints
// ExtractText / ExtractImages / GetMetadata results.
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

const outFile = "parsing_demo.docx"

func main() {
	doc := word.New()
	info := doc.GetDocInfo()
	info.Title = "GoWord v0.6.0 Parsing Demo"
	info.Creator = "GoWord"
	info.Subject = "Nested tables, cell styles, bookmarks, extraction"
	info.Company = "yunkeweb"

	doc.SetDefaultFontName("Calibri")
	doc.SetDefaultAsianFontName("Microsoft YaHei")
	doc.SetDefaultFontSize(11)

	sec := doc.AddSection()
	sec.AddTitle("Document Parsing & Advanced Media", 1)

	jump := sec.AddTextRun()
	jump.AddText("Jump to nested tables: ")
	doc.AddHyperlinkToBookmark(jump, "Nested Tables", "nested_tables")

	sec.AddTitle("Paragraphs", 2)
	sec.AddText("This document is written with go-word, saved as OOXML, then opened again with word.Open / ExtractText / ExtractImages / GetMetadata.")

	sec.AddTitle("Inline Image", 2)
	sec.AddImageBytes("demo.png", demoPNG(), style.Image{Width: 40, Height: 20, AltText: "demo swatch"})

	target := sec.AddTextRun()
	doc.AddBookmark(target, "nested_tables")
	target.AddText("Nested Tables")
	sec.AddTitle("Nested Tables", 2)

	outer := sec.AddTable(style.Table{
		Width:     9000,
		Alignment: style.JcTableCenter,
		Borders: style.Borders{
			Top:     style.Border{Style: style.BorderSingle, Size: 4, Color: "1F4E79"},
			Left:    style.Border{Style: style.BorderSingle, Size: 4, Color: "1F4E79"},
			Bottom:  style.Border{Style: style.BorderSingle, Size: 4, Color: "1F4E79"},
			Right:   style.Border{Style: style.BorderSingle, Size: 4, Color: "1F4E79"},
			InsideH: style.Border{Style: style.BorderSingle, Size: 4, Color: "1F4E79"},
			InsideV: style.Border{Style: style.BorderSingle, Size: 4, Color: "1F4E79"},
		},
	})
	header := outer.AddRow()
	h1 := header.AddCell(4500)
	h1.SetVerticalAlignment(word.VAlignCenter)
	h1.SetPadding(80, 80, 80, 80)
	h1.SetBorder("all", style.BorderSingle, 8, "1F4E79")
	h1.AddText("Outer cell (with nested table)", style.Font{Bold: true, Color: "FFFFFF", Size: 11}, style.Paragraph{Alignment: style.JcCenter})
	h1.Style.BgColor = "1F4E79"
	h2 := header.AddCell(4500)
	h2.SetVerticalAlignment(word.VAlignCenter)
	h2.SetPadding(80, 80, 80, 80)
	h2.AddText("Notes", style.Font{Bold: true, Color: "FFFFFF"}, style.Paragraph{Alignment: style.JcCenter})
	h2.Style.BgColor = "1F4E79"

	body := outer.AddRow()
	parent := body.AddCell(4500)
	parent.SetPadding(60, 80, 60, 80)
	parent.SetVerticalAlignment(word.VAlignTop)
	parent.AddText("Parent cell contains a nested w:tbl. The cell ends with an empty w:p as required by CT_Tc.")
	nested := parent.AddTable(style.Table{Width: 4000})
	nr := nested.AddRow()
	c11 := nr.AddCell(2000)
	c11.SetBorder("top", style.BorderSingle, 12, "C00000")
	c11.SetBorder("left", style.BorderDashed, 8, "00B050")
	c11.SetBorder("bottom", style.BorderDouble, 16, "0070C0")
	c11.SetBorder("right", style.BorderDotted, 8, "7030A0")
	c11.SetPadding(40, 60, 40, 60)
	c11.SetVerticalAlignment(word.VAlignCenter)
	c11.AddText("Nested A")
	c12 := nr.AddCell(2000)
	c12.SetTextDirection(word.TextDirectionVertical)
	c12.SetVerticalAlignment(word.VAlignCenter)
	c12.SetPadding(40, 40, 40, 40)
	c12.AddText("Vert")

	nr2 := nested.AddRow()
	nr2.AddCell(2000).AddText("Nested B1")
	nr2.AddCell(2000).AddText("Nested B2")

	note := body.AddCell(4500)
	note.SetVerticalAlignment(word.VAlignCenter)
	note.SetPadding(80, 100, 80, 100)
	note.AddText("Independent cell borders, padding (w:tcMar), vertical alignment, and text direction are set per cell.")

	sec.AddTitle("Bookmarks", 2)
	sec.AddText("The heading area above is bookmarked as nested_tables. The first paragraph uses w:hyperlink w:anchor to jump there.")

	if err := doc.Save(outFile); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("wrote %s\n", outFile)

	opened, err := word.Open(outFile)
	if err != nil {
		log.Fatal(err)
	}
	meta := opened.GetMetadata()
	fmt.Printf("metadata title=%q creator=%q created=%s\n", meta.Title, meta.Creator, meta.Created.UTC().Format("2006-01-02T15:04:05Z"))
	fmt.Println("--- ExtractText ---")
	fmt.Println(opened.ExtractText())
	imgs, err := opened.ExtractImages()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("--- ExtractImages (%d) ---\n", len(imgs))
	for _, img := range imgs {
		fmt.Printf("  %s %s %d bytes\n", img.Name, img.MIME, len(img.Data))
	}
}

func demoPNG() []byte {
	img := image.NewRGBA(image.Rect(0, 0, 40, 20))
	for y := 0; y < 20; y++ {
		for x := 0; x < 40; x++ {
			img.Set(x, y, color.RGBA{R: 31, G: 78, B: 121, A: 255})
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		log.Fatal(err)
	}
	return buf.Bytes()
}
