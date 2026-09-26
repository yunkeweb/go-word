# Watermark & Document Protection

Text watermarks are Word-native VML (`PowerPlusWaterMarkObject`) injected into every section header. Image watermarks use `WordPictureWatermark` (`v:imagedata`). Document protection writes `w:documentProtection` with an Office SHA-1 / 100000-spin hash when a password is set. Editable exception ranges wrap `w:permStart` / `w:permEnd` around a paragraph, cell, or table.

## SetTextWatermark

### Signature

```go
func (d *Document) SetTextWatermark(text string, opts ...WatermarkOptions)
```

```go
type WatermarkOptions struct {
	Angle    float64 // degrees; 0 defaults to -45 (Word rotation:315)
	Color    string  // VML fillcolor, e.g. silver or C0C0C0
	FontSize int     // points; 0 uses Word's 1pt + fitshape
	FontName string  // default Calibri
	Opacity  float64 // 0–1; 0 defaults to 0.5. Values > 1 are treated as percent
	Tile     bool    // full-page  grid on a 612×792 pt page
	Rows     int     // tile rows, default 3
	Cols     int     // tile columns, default 3
}
```

### Parameters

| Name | Type | Description |
| --- | --- | --- |
| `text` | `string` | Watermark caption. Empty string clears a previously set text watermark. |
| `opts` | `...WatermarkOptions` | Optional. Omitted → one diagonal silver mark at opacity 0.5 (same as v0.8.0). |

### Notes

- `Angle: 0` is treated as `-45°` (`rotation:315`). Pass a non-zero angle to override.
- `Tile: true` paints a `Rows × Cols` grid (default 3×3) across the page.
- Watermarks are applied at write time to each section header. Creating headers yourself is compatible: the writer still injects the VML.

## SetImageWatermark / SetImageWatermarkFile

### Signature

```go
func (d *Document) SetImageWatermark(imageBytes []byte, opts ...ImageWatermarkOptions)
func (d *Document) SetImageWatermarkFile(path string, opts ...ImageWatermarkOptions) error
```

```go
type ImageWatermarkOptions struct {
	Washout bool    // Word washout (gain="19661f" blacklevel="22938f")
	Scale   float64 // size multiplier applied at write time; 0 leaves the default
	Opacity float64 // 0–1; 0 omits v:fill opacity
}
```

### Parameters

| Name | Type | Description |
| --- | --- | --- |
| `imageBytes` | `[]byte` | Encoded PNG/JPEG. Empty slice clears the image watermark. |
| `path` | `string` | File path for `SetImageWatermarkFile` (PNG or JPEG). |
| `opts` | `...ImageWatermarkOptions` | Optional. Omitted → no washout (same as v0.8.0). |

### Notes

- `SetImageWatermark` keeps the `[]byte` signature. Use `SetImageWatermarkFile` when the source is on disk.
- Image watermarks add a media part referenced from the header relationships.
- Washout is Word’s grey-out filter, independent of `Opacity`.

## Protect

### Signature

```go
func (d *Document) Protect(editing, password string) error
```

### Parameters

| Name | Type | Description |
| --- | --- | --- |
| `editing` | `string` | `ProtectTypeReadOnly`, `ProtectTypeComments`, `ProtectTypeTrackedChanges`, `ProtectTypeForms`. Empty → read-only. |
| `password` | `string` | Optional. Non-empty values use the ECMA-376 document-protection hash. |

### Notes

- This is **document editing restriction**, not ZIP encryption. Anyone can still unzip the package.
- Word prompts for the password when the user tries to edit. The sample password in demos is `goword`.

## AllowEdit {#allowedit}

### Signature

```go
func (t *Text) AllowEdit(groupOrUser string) *Text
func (t *TextRun) AllowEdit(groupOrUser string) *TextRun
func (c *Cell) AllowEdit(groupOrUser string) *Cell
func (t *Table) AllowEdit(groupOrUser string) *Table
```

### Parameters

| Name | Type | Description |
| --- | --- | --- |
| `groupOrUser` | `string` | `Everyone` / `everybody` / `all` → `w:edGrp="everyone"`. Also `none`, `administrators`, `contributors`, `editors`, `owners`, `current`. Any other string is written as `w:ed` (a named user). Empty defaults to everyone. |

### Notes

- Exception ranges only take effect when `Protect` is also called.
- The writer emits matching `w:permStart` / `w:permEnd` around the element’s XML.

## Complete example

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
	doc.SetTextWatermark("CONFIDENTIAL", word.WatermarkOptions{
		Angle: -45, Color: "C0C0C0", FontSize: 36, Opacity: 0.28,
		Tile: true, Rows: 3, Cols: 3,
	})
	if err := doc.Protect(word.ProtectTypeReadOnly, "goword"); err != nil {
		log.Fatal(err)
	}

	sec := doc.AddSection()
	sec.AddTitle("Protected contract", 1)
	sec.AddText("Standard clauses stay locked.")
	sec.AddText("Party A: ________________", style.Font{Bold: true}).AllowEdit("Everyone")

	tbl := sec.AddTable(style.Table{Width: 9000})
	hdr := tbl.AddRow()
	tbl.SetHeaderRow(hdr)
	hdr.AddCell(3000).AddText("Field", style.Font{Bold: true})
	hdr.AddCell(6000).AddText("Value", style.Font{Bold: true})
	row := tbl.AddRow()
	row.AddCell(3000).AddText("Contract no.")
	row.AddCell(6000, style.Cell{Shading: style.Shading{Fill: "E2F0D9"}}).
		AllowEdit("Everyone").
		AddText("CN-2026-001")

	if err := doc.Save("protected.docx"); err != nil {
		log.Fatal(err)
	}
}
```

Picture watermark from a file:

```go
err := doc.SetImageWatermarkFile("mark.png", word.ImageWatermarkOptions{
	Washout: true, Scale: 1.2, Opacity: 0.35,
})
```

### What Word shows

Opening `protected.docx` displays a 3×3 grey **CONFIDENTIAL** grid behind the body. The status bar reports read-only. Restrict Editing lists the password hash; entering `goword` unlocks the whole file. Without the password, only the Party A paragraph and the green contract-number cell accept input. A longer sample lives in [`examples/v0.9.0_watermark_security`](https://github.com/yunkeweb/go-word/tree/main/examples/v0.9.0_watermark_security).
