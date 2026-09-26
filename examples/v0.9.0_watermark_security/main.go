// Command v0.9.0_watermark_security builds a protected document with a
// tiled diagonal text watermark, an image-watermark companion file, and
// exception ranges (w:permStart / w:permEnd) that remain editable.
package main

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"log"
	"os"
	"path/filepath"

	"github.com/yunkeweb/go-word"
	"github.com/yunkeweb/go-word/style"
)

const (
	outProtected = "v090_watermark_security.docx"
	outImage     = "v090_image_watermark.docx"
)

func main() {
	if err := writeProtected(); err != nil {
		log.Fatal(err)
	}
	if err := writeImageWatermark(); err != nil {
		log.Fatal(err)
	}
}

func writeProtected() error {
	doc := word.New()
	info := doc.GetDocInfo()
	info.Title = "GoWord v0.9.0 watermark and protection"
	info.Creator = "GoWord"
	info.Subject = "Tiled text watermark, documentProtection, permStart/permEnd"
	info.Company = "yunkeweb"

	doc.SetDefaultFontName("Calibri")
	doc.SetDefaultAsianFontName("Microsoft YaHei")
	doc.SetDefaultFontSize(11)

	doc.SetTextWatermark("CONFIDENTIAL", word.WatermarkOptions{
		Angle:    -45,
		Color:    "C0C0C0",
		FontSize: 36,
		Opacity:  0.28,
		Tile:     true,
		Rows:     3,
		Cols:     3,
	})
	if err := doc.Protect(word.ProtectTypeReadOnly, "goword"); err != nil {
		return err
	}

	sec := doc.AddSection()
	sec.AddTitle("Protected contract", 1)
	sec.AddText("This document is read-only (password: goword). Grey ranges stay locked; highlighted ranges are w:permStart exceptions.")

	sec.AddTitle("Locked body", 2)
	sec.AddText("Standard clauses cannot be edited while protection is enforced.")

	sec.AddTitle("Editable party name", 2)
	sec.AddText("Party A: ________________________", style.Font{Bold: true}, style.Paragraph{}).
		AllowEdit("Everyone")

	p := sec.AddTextRun()
	p.AddText("Party B: ")
	p.AddText("________________________", style.Font{Underline: style.UnderlineSingle})
	p.AllowEdit("Everyone")

	sec.AddTitle("Editable table cell", 2)
	tbl := sec.AddTable(style.Table{
		Width: 9000,
		Borders: style.Borders{
			Top:     style.Border{Style: style.BorderSingle, Size: 4, Color: "1F4E79"},
			Left:    style.Border{Style: style.BorderSingle, Size: 4, Color: "1F4E79"},
			Bottom:  style.Border{Style: style.BorderSingle, Size: 4, Color: "1F4E79"},
			Right:   style.Border{Style: style.BorderSingle, Size: 4, Color: "1F4E79"},
			InsideH: style.Border{Style: style.BorderSingle, Size: 4, Color: "BDD7EE"},
			InsideV: style.Border{Style: style.BorderSingle, Size: 4, Color: "BDD7EE"},
		},
	})
	hdr := tbl.AddRow(360)
	tbl.SetHeaderRow(hdr)
	hdr.AddCell(3000, style.Cell{Width: 3000, BgColor: "1F4E79"}).
		SetVAlign("center").
		AddText("Field", style.Font{Bold: true, Color: "FFFFFF"})
	hdr.AddCell(6000, style.Cell{Width: 6000, BgColor: "1F4E79"}).
		SetVAlign("center").
		AddText("Value", style.Font{Bold: true, Color: "FFFFFF"})
	row := tbl.AddRow(320)
	row.AddCell(3000).AddText("Contract no.")
	row.AddCell(6000, style.Cell{BgColor: "E2F0D9"}).
		AllowEdit("Everyone").
		AddText("CN-2026-001")
	row2 := tbl.AddRow(320)
	row2.AddCell(3000).AddText("Status")
	row2.AddCell(6000).AddText("Locked — not an exception range")

	return doc.Save(outProtected)
}

func writeImageWatermark() error {
	pngData := watermarkPNG()
	tmp := filepath.Join(os.TempDir(), "goword-v090-mark.png")
	if err := os.WriteFile(tmp, pngData, 0o644); err != nil {
		return err
	}
	defer os.Remove(tmp)

	doc := word.New()
	info := doc.GetDocInfo()
	info.Title = "GoWord v0.9.0 image watermark"
	info.Creator = "GoWord"
	if err := doc.SetImageWatermarkFile(tmp, word.ImageWatermarkOptions{
		Washout: true,
		Scale:   1.2,
		Opacity: 0.35,
	}); err != nil {
		return err
	}
	sec := doc.AddSection()
	sec.AddTitle("Picture watermark", 1)
	sec.AddText("The header contains v:shape + v:imagedata with Word washout (gain/blacklevel) and fill opacity.")
	return doc.Save(outImage)
}

func watermarkPNG() []byte {
	img := image.NewRGBA(image.Rect(0, 0, 240, 80))
	for y := 0; y < 80; y++ {
		for x := 0; x < 240; x++ {
			img.Set(x, y, color.RGBA{R: 192, G: 192, B: 192, A: 255})
		}
	}
	var buf bytes.Buffer
	_ = png.Encode(&buf, img)
	return buf.Bytes()
}
