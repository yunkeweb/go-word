// Command v0.5.0_demo builds enterprise_demo.docx: watermarks, document
// protection, mixed portrait/landscape sections, first/odd/even headers,
// PAGE/NUMPAGES fields, and an H1–H3 table of contents.
package main

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log"

	"github.com/yunkeweb/go-word"
	"github.com/yunkeweb/go-word/element"
	"github.com/yunkeweb/go-word/style"
)

const outFile = "enterprise_demo.docx"

func main() {
	doc := word.New()
	info := doc.GetDocInfo()
	info.Title = "GoWord v0.5.0 Enterprise Demo"
	info.Creator = "GoWord"
	info.Subject = "Watermarks, protection, multi-section page setup, TOC"
	info.Company = "yunkeweb"

	doc.SetDefaultFontName("Calibri")
	doc.SetDefaultAsianFontName("Microsoft YaHei")
	doc.SetDefaultFontSize(11)
	doc.AddTitleStyle(1, style.Font{Bold: true, Size: 16})
	doc.AddTitleStyle(2, style.Font{Bold: true, Size: 13})
	doc.AddTitleStyle(3, style.Font{Bold: true, Size: 12})

	if err := doc.Protect(word.ProtectTypeReadOnly, "goword"); err != nil {
		log.Fatal(err)
	}
	doc.SetTextWatermark("CONFIDENTIAL")
	doc.SetImageWatermark(watermarkPNG())

	portrait := doc.AddSection()
	portrait.AddHeader().AddText("GoWord Enterprise — odd pages")
	portrait.AddHeader(element.HeaderFirst).AddText("GoWord Enterprise — first page")
	portrait.AddHeader(element.HeaderEven).AddText("GoWord Enterprise — even pages")
	oddFooter := portrait.AddFooter()
	oddFooter.AddText("Page")
	oddFooter.AddPageNumber()
	oddFooter.AddText("of")
	oddFooter.AddNumPages()
	portrait.AddFooter(element.HeaderFirst).AddText("Cover footer")

	portrait.AddTitle("Table of Contents", 1)
	doc.AddTableOfContents()
	portrait.AddText("Right-click the directory field in Microsoft Word and choose Update Field.")

	portrait.AddTitle("Portrait Section", 1)
	portrait.AddText("This section is A4 portrait. The first page uses a distinct header and footer (w:titlePg). Odd and even pages use different headers (w:evenAndOddHeaders).")
	portrait.AddTitle("Document Protection", 2)
	portrait.AddText("The package is write-protected (w:documentProtection, edit=readOnly). The sample password is goword.")
	portrait.AddTitle("Watermarks", 3)
	portrait.AddText("A diagonal CONFIDENTIAL text watermark and an image watermark are stored in the headers as Word-native VML.")
	portrait.AddPageBreak()
	portrait.AddText("Second portrait page (odd). The PAGE field in the footer updates in Word.")
	portrait.AddPageBreak()
	portrait.AddText("Third portrait page (even). The even-page header is distinct from the odd-page header.")

	land := doc.AddSection(style.Section{
		Orientation:  style.OrientationLandscape,
		BreakType:    "nextPage",
		PageSizeW:    style.DefaultPageWidth,
		PageSizeH:    style.DefaultPageHeight,
		MarginTop:    style.DefaultMargin,
		MarginBottom: style.DefaultMargin,
		MarginLeft:   style.DefaultMargin,
		MarginRight:  style.DefaultMargin,
		HeaderHeight: style.DefaultHeaderHeight,
		FooterHeight: style.DefaultFooterHeight,
	})
	land.AddHeader().AddText("Landscape header")
	land.AddFooter().AddPageNumber()
	land.AddTitle("Landscape Section", 1)
	land.AddText("This section is A4 landscape (pgSz orient=landscape with width/height swapped). It shares the document watermark and protection settings.")
	land.AddTitle("Page Fields", 2)
	land.AddText("Footers use OpenXML PAGE and NUMPAGES fields (w:fldChar / w:instrText).")
	land.AddTitle("Section Breaks", 3)
	land.AddText("A next-page section break separates portrait and landscape content inside one document.")

	if err := doc.Save(outFile); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("wrote %s\n", outFile)
}

func watermarkPNG() []byte {
	img := image.NewRGBA(image.Rect(0, 0, 160, 80))
	for y := 0; y < 80; y++ {
		for x := 0; x < 160; x++ {
			img.Set(x, y, color.RGBA{R: 220, G: 220, B: 220, A: 80})
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		log.Fatal(err)
	}
	return buf.Bytes()
}
