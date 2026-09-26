# Structured Document Tags (SDT)

Content controls are Word-native form fields. GoWord writes `w:sdt` → `w:sdtPr` → `w:sdtContent` (CT_SdtBlock / CT_SdtRun). Unique `w:id` values start at `100000000`. Checkboxes use the Word 2010 `w14:checkbox` markup; the document root declares `mc:Ignorable="w14 wps"`.

`AddSDT*` methods live on `element.Container`, so they are available on a section, a table cell, a header, and a text run.

## AddSDTText / AddSDTDropdown / AddSDTDate / AddSDTCheckbox

### Signature

```go
func (c *Container) AddSDTText(alias, tag, placeholderText string) *SDT
func (c *Container) AddSDTDropdown(alias, tag string, options map[string]string) *SDT
func (c *Container) AddSDTDate(alias, tag, dateFormat string) *SDT
func (c *Container) AddSDTCheckbox(alias, tag string, checked bool) *SDT
func (c *Container) AddSDT(typ string) *SDT
```

### Parameters

| Name | Type | Description |
| --- | --- | --- |
| `alias` | `string` | `w:alias` — the title Word shows in the control chrome. |
| `tag` | `string` | `w:tag` — a stable machine name for later fill / extract. |
| `placeholderText` | `string` | Visible `w:sdtContent` text. Non-empty values also set `w:showingPlcHdr`. |
| `options` | `map[string]string` | Drop-down `value → displayText`. Keys are written in sorted order as `w:listItem`. |
| `dateFormat` | `string` | `w:dateFormat`. Empty becomes `yyyy-MM-dd`. Locale is `en-US`. |
| `checked` | `bool` | `w14:checked`. Content glyph is `☒` when true, `☐` when false (MS Gothic). |
| `typ` | `string` | PHPWord-style kind for `AddSDT`: `plainText`, `dropDownList`, `date`, `checkbox`, `comboBox`, `richText`. |

### Chainable setters

| Method | Effect |
| --- | --- |
| `SetAlias` / `SetTag` | Rewrite `w:alias` / `w:tag`. |
| `SetValue` | Replace visible content and clear the placeholder flag. |
| `SetListItems` | Replace drop-down / combo-box entries. |
| `SetDateFormat` | Rewrite `w:dateFormat`. |
| `SetChecked` | Toggle `w14:checkbox` and the content glyph. |

### Notes

- Node order is `w:sdt` → `w:sdtPr` → `w:sdtContent`. Do not hand-edit the XML.
- Empty drop-down maps still write a control whose placeholder is `Choose an item.`
- `AddSDT("plainText")` remains the PHPWord alias; prefer the typed helpers in new code.
- A Developer tab is **not** required to fill the controls in Microsoft Word.

## Complete example

Save as `main.go` and run `go run .`. Microsoft Word opens `form.docx` without a repair dialog.

```go
package main

import (
	"log"

	"github.com/yunkeweb/go-word"
	"github.com/yunkeweb/go-word/style"
)

func main() {
	doc := word.New()
	doc.SetDefaultFontName("Calibri")
	sec := doc.AddSection()
	sec.AddTitle("Employee onboarding form", 1)

	sec.AddSDTText("Full name", "full_name", "Enter full name")
	sec.AddSDTDropdown("Department", "department", map[string]string{
		"eng": "Engineering",
		"fin": "Finance",
		"hr":  "Human Resources",
	})
	sec.AddSDTDate("Start date", "start_date", "yyyy-MM-dd")

	p := sec.AddTextRun()
	p.AddText("I have read the handbook  ")
	p.AddSDTCheckbox("Handbook", "handbook_ack", false)

	tbl := sec.AddTable(style.Table{Width: 9000})
	hdr := tbl.AddRow()
	hdr.AddCell(3000, style.Cell{Shading: style.Shading{Fill: "1F4E79"}}).
		AddText("Field", style.Font{Bold: true, Color: "FFFFFF"})
	hdr.AddCell(6000, style.Cell{Shading: style.Shading{Fill: "1F4E79"}}).
		AddText("Value", style.Font{Bold: true, Color: "FFFFFF"})
	row := tbl.AddRow()
	row.AddCell(3000).AddText("Manager")
	row.AddCell(6000).AddSDTText("Manager", "manager", "Enter manager name")

	if err := doc.Save("form.docx"); err != nil {
		log.Fatal(err)
	}
}
```

### What Word shows

Four content controls: a plain-text box, a three-item drop-down, a date picker, and an unchecked checkbox next to the handbook sentence. The manager field sits inside a table cell (`w:sdt` as a child of `w:tc`). A longer sample lives in [`examples/v0.9.0_sdt`](https://github.com/yunkeweb/go-word/tree/main/examples/v0.9.0_sdt). Combine with [document protection](./protect) when the rest of the page must stay locked.
