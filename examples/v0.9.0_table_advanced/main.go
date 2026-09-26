// Command v0.9.0_table_advanced builds v090_table_advanced.docx with
// repeating table headers, unbreakable rows, vertical cell alignment,
// and vertical text direction.
package main

import (
	"fmt"
	"log"

	"github.com/yunkeweb/go-word"
	"github.com/yunkeweb/go-word/element"
	"github.com/yunkeweb/go-word/style"
)

const outFile = "v090_table_advanced.docx"

func main() {
	if err := writeDoc(); err != nil {
		log.Fatal(err)
	}
}

func writeDoc() error {
	doc := word.New()
	info := doc.GetDocInfo()
	info.Title = "GoWord v0.9.0 table mechanics"
	info.Creator = "GoWord"
	info.Subject = "tblHeader, cantSplit, vAlign, textDirection"
	info.Company = "yunkeweb"

	doc.SetDefaultFontName("Calibri")
	doc.SetDefaultAsianFontName("Microsoft YaHei")
	doc.SetDefaultFontSize(11)

	sec := doc.AddSection()
	sec.AddTitle("Table Mechanics Plus", 1)
	sec.AddText("The header row repeats on every page. Highlighted rows stay together (w:cantSplit).")

	tbl := sec.AddTable(style.Table{
		Width: 9360, Layout: "fixed", Alignment: style.JcTableCenter,
		Borders: style.Borders{
			Top:     style.Border{Style: style.BorderSingle, Size: 4, Color: "1F4E79"},
			Left:    style.Border{Style: style.BorderSingle, Size: 4, Color: "1F4E79"},
			Bottom:  style.Border{Style: style.BorderSingle, Size: 4, Color: "1F4E79"},
			Right:   style.Border{Style: style.BorderSingle, Size: 4, Color: "1F4E79"},
			InsideH: style.Border{Style: style.BorderSingle, Size: 4, Color: "BDD7EE"},
			InsideV: style.Border{Style: style.BorderSingle, Size: 4, Color: "BDD7EE"},
		},
	})
	hdr := tbl.AddRow(400)
	tbl.SetHeaderRow(hdr)
	hdr.SetCantSplit(true)
	addHeaderCell(hdr, 1200, "No.")
	addHeaderCell(hdr, 2800, "Item")
	addHeaderCell(hdr, 1600, "Qty")
	addHeaderCell(hdr, 3760, "Note")

	for i := 1; i <= 48; i++ {
		row := tbl.AddRow(320)
		if i%6 == 0 {
			row.SetCantSplit(true)
		}
		bg := "FFFFFF"
		if i%2 == 0 {
			bg = "F2F2F2"
		}
		if i%6 == 0 {
			bg = "FFF2CC"
		}
		addBodyCell(row, 1200, bg, fmt.Sprintf("%d", i), "center")
		addBodyCell(row, 2800, bg, fmt.Sprintf("Widget %02d", i), "center")
		addBodyCell(row, 1600, bg, "1", "center")
		note := "Fits on one line."
		if i%6 == 0 {
			note = "This row is marked w:cantSplit so Word keeps it on a single page."
		}
		addBodyCell(row, 3760, bg, note, "top")
	}

	sec.AddTitle("Vertical alignment and text direction", 2)
	sec.AddText("Left cell is tbRl vertical text; the others show top / center / bottom vAlign.")

	demo := sec.AddTable(style.Table{
		Width: 9360, Layout: "fixed",
		Borders: style.Borders{
			Top:     style.Border{Style: style.BorderSingle, Size: 4, Color: "1F4E79"},
			Left:    style.Border{Style: style.BorderSingle, Size: 4, Color: "1F4E79"},
			Bottom:  style.Border{Style: style.BorderSingle, Size: 4, Color: "1F4E79"},
			Right:   style.Border{Style: style.BorderSingle, Size: 4, Color: "1F4E79"},
			InsideH: style.Border{Style: style.BorderSingle, Size: 4, Color: "BDD7EE"},
			InsideV: style.Border{Style: style.BorderSingle, Size: 4, Color: "BDD7EE"},
		},
	})
	row := demo.AddRow(1600)
	row.SetCantSplit(true)
	row.AddCell(1600, style.Cell{Width: 1600, BgColor: "1F4E79"}).
		SetVAlign("center").
		SetTextDirection("tbRl").
		AddText("竖排标题", style.Font{Bold: true, Color: "FFFFFF", Size: 12})
	row.AddCell(2586, style.Cell{Width: 2586, BgColor: "DEEBF7"}).
		SetVAlign("top").
		AddText("top")
	row.AddCell(2587, style.Cell{Width: 2587, BgColor: "FFF2CC"}).
		SetVAlign("center").
		AddText("center")
	row.AddCell(2587, style.Cell{Width: 2587, BgColor: "E2F0D9"}).
		SetVAlign("bottom").
		AddText("bottom")

	return doc.Save(outFile)
}

func addHeaderCell(row *element.Row, width int, text string) {
	row.AddCell(width, style.Cell{Width: width, BgColor: "1F4E79"}).
		SetVAlign("center").
		AddText(text, style.Font{Bold: true, Color: "FFFFFF"})
}

func addBodyCell(row *element.Row, width int, bg, text, valign string) {
	row.AddCell(width, style.Cell{Width: width, BgColor: bg}).
		SetVAlign(valign).
		AddText(text)
}
