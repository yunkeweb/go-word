// Command v0.9.0_sdt builds v090_sdt.docx with four Word content controls
// (plain text, drop-down, date picker, checkbox) that open in Microsoft Word.
package main

import (
	"log"

	"github.com/yunkeweb/go-word"
	"github.com/yunkeweb/go-word/style"
)

const outFile = "v090_sdt.docx"

func main() {
	if err := writeForm(); err != nil {
		log.Fatal(err)
	}
}

func writeForm() error {
	doc := word.New()
	info := doc.GetDocInfo()
	info.Title = "GoWord v0.9.0 SDT form controls"
	info.Creator = "GoWord"
	info.Subject = "Structured Document Tags (w:sdt)"
	info.Company = "yunkeweb"

	doc.SetDefaultFontName("Calibri")
	doc.SetDefaultAsianFontName("Microsoft YaHei")
	doc.SetDefaultFontSize(11)

	sec := doc.AddSection()
	sec.AddTitle("Employee onboarding form", 1)
	sec.AddText("Fill the content controls in Microsoft Word (Developer tab is not required).")

	sec.AddTitle("Plain text", 2)
	sec.AddSDTText("Full name", "full_name", "Enter full name")

	sec.AddTitle("Drop-down list", 2)
	sec.AddSDTDropdown("Department", "department", map[string]string{
		"eng": "Engineering",
		"fin": "Finance",
		"hr":  "Human Resources",
	})

	sec.AddTitle("Date picker", 2)
	sec.AddSDTDate("Start date", "start_date", "yyyy-MM-dd")

	sec.AddTitle("Checkbox", 2)
	p := sec.AddTextRun()
	p.AddText("I have read the handbook  ")
	p.AddSDTCheckbox("Handbook", "handbook_ack", false)

	sec.AddTitle("Nested in a table cell", 2)
	tbl := sec.AddTable(style.Table{Width: 9000})
	hdr := tbl.AddRow()
	hdr.AddCell(3000, style.Cell{Shading: style.Shading{Fill: "1F4E79"}}).
		AddText("Field", style.Font{Bold: true, Color: "FFFFFF"})
	hdr.AddCell(6000, style.Cell{Shading: style.Shading{Fill: "1F4E79"}}).
		AddText("Value", style.Font{Bold: true, Color: "FFFFFF"})
	row := tbl.AddRow()
	row.AddCell(3000).AddText("Manager")
	row.AddCell(6000).AddSDTText("Manager", "manager", "Enter manager name")

	return doc.Save(outFile)
}
