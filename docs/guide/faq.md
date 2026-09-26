# FAQ & Word Compatibility

Microsoft Word shows a repair dialog when a part violates the schema it expects. GoWord’s `openxml_strict_test.go` encodes the rules that have already bitten production files. This page maps symptoms to those rules.

## Word says the file is corrupt

| Symptom | Likely cause | Fix |
| --- | --- | --- |
| Repair dialog on open, charts missing | Series XML out of XSD order, or `c:legend` inside `c:plotArea` | Use `AddChart` / `AddComboSeries`. Do not hand-edit chart XML. See [Charts](./charts). |
| Repair dialog, area chart | `c:dLblPos` on an area series | `SetDataLabels` omits `dLblPos` for area charts. |
| Repair dialog, combo chart | Both series share one `c:axId` | `AddComboSeries(..., true)` allocates a second value axis. |
| Repair dialog, tables | A `w:tc` without a `w:p` | Always `AddText` (or leave the cell empty — GoWord inserts an empty paragraph). |
| Repair dialog, columns | `w:separator` instead of `w:sep` | Call `SetColumns`; do not emit a custom `w:cols`. |
| Template output has `&#80;` / broken entities | Raw `&` in replaced `w:t` | Use `SetValue`. The processor XML-escapes replacements. |

## Placeholders are not replaced

Word often splits `${name}` across several `w:r` nodes when the user edits a template in the GUI. Generate the template with GoWord (`AddText("${name}")`) or keep each placeholder in one run. `NewTemplateProcessorBytes` on a document you just built never hits this.

## Images disappear after a merge

Two source documents both created `word/media/image1.png`. Without remapping, the second part overwrites the first. `AppendDocument` clears `RelationID` and the writer assigns `imageN` + a fresh `rId`. See [Document Merger](./merger).

## TOC / PAGE fields show “Error! Bookmark not defined.”

Fields store a cached result. Right-click → Update Field, or print. `AddTitle` must be used so outline levels exist. This is Word behaviour, not a missing part.

## Protect does not encrypt the ZIP

`Protect` writes `w:documentProtection`. Anyone can unzip the package. Use OS-level encryption or a passworded archive if the bytes must stay secret.

## Two-column layout only applies to part of the page

`SetColumns` is a section property. Content after a new `AddSection` is a new `w:sectPr`. Put the column body in that section; keep titles in the previous one if they should stay full width.

## Which Word versions are supported?

Word 2007 through Microsoft 365 open the packages. DrawingML `wps:wsp` is a Word 2010+ feature; Word 2007 still opens the file and may drop the shape. Charts need Word 2007+. LibreOffice opens paragraphs, tables, and images; treat charts and `wps` shapes as Word-first.

## Complete example — dump parts after a failure

```go
package main

import (
	"archive/zip"
	"bytes"
	"fmt"
	"log"

	"github.com/yunkeweb/go-word"
)

func main() {
	doc := word.New()
	sec := doc.AddSection()
	sec.AddText("debug")
	sec.AddMath(`\frac{a}{b}`)
	raw, err := doc.Bytes()
	if err != nil {
		log.Fatal(err)
	}
	zr, err := zip.NewReader(bytes.NewReader(raw), int64(len(raw)))
	if err != nil {
		log.Fatal(err)
	}
	for _, f := range zr.File {
		fmt.Printf("%s %d\n", f.Name, f.UncompressedSize64)
	}
}
```

If Word still repairs a file you built, open the corresponding part (`word/charts/chart1.xml`, `word/document.xml`) and compare it with [`examples/v0.8.0_demo`](https://github.com/yunkeweb/go-word/tree/main/examples/v0.8.0_demo) and [`openxml_strict_test.go`](https://github.com/yunkeweb/go-word/blob/main/openxml_strict_test.go).
