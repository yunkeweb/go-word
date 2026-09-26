# FAQ

Microsoft Word shows a repair dialog when a part violates the schema it expects. GoWord’s `openxml_strict_test.go` encodes the rules that have already bitten production files. This page maps symptoms to those rules, then covers style remapping after a merge.

Related: [OpenXML Compatibility](./compatibility), [Document Merger](./merger), [Enterprise Recipes](./recipes).

## Word says the file is corrupt

| Symptom | Likely cause | Fix |
| --- | --- | --- |
| Repair dialog on open, charts missing | Series XML out of XSD order, or `c:legend` inside `c:plotArea` | Use `AddChart` / `AddComboSeries`. Do not hand-edit chart XML. See [Charts](./charts). |
| Repair dialog, area chart | `c:dLblPos` on an area series | `SetDataLabels` omits `dLblPos` for area charts. |
| Repair dialog, combo chart | Both series share one `c:axId` | `AddComboSeries(..., true)` allocates a second value axis. |
| Repair dialog, tables | A `w:tc` without a `w:p` | Always `AddText` (or leave the cell empty — GoWord inserts an empty paragraph). Nested tables still need that trailing paragraph; `Cell.AddTable` writes it. |
| Repair dialog, columns | `w:separator` instead of `w:sep` | Call `SetColumns`; do not emit a custom `w:cols`. |
| Repair dialog, shapes | DrawingML `wps:wsp` written without a `w:drawing` wrapper | Use `AddShape` / `AddTextBox`. |
| Repair dialog, content controls | `w:sdtContent` missing or out of order | Use `AddSDTText` / `AddSDTDropdown` / `AddSDTDate` / `AddSDTCheckbox`. See [SDT](./sdt). |
| Protection ignores a fill-in field | `AllowEdit` omitted on that paragraph or cell | Call `AllowEdit("Everyone")` after `Protect`. See [Protection](./protect). |
| Template output has `&#80;` / broken entities | Raw `&` in replaced `w:t` | Use `SetValue`. The processor XML-escapes replacements. |
| Need a full-feature smoke test | Combine every public module | `go run ./tests/matrix`, then the Go reader and Word COM validators. See [Compatibility](./compatibility) and [Recipe 5](./recipes#5-full-feature-matrix). |

If Word still repairs a file you built, dump the ZIP parts (example at the bottom) and compare `word/charts/chart1.xml` or `word/document.xml` with [`examples/v0.9.0_sdt`](https://github.com/yunkeweb/go-word/tree/main/examples/v0.9.0_sdt), [`examples/v0.8.0_demo`](https://github.com/yunkeweb/go-word/tree/main/examples/v0.8.0_demo) and [`openxml_strict_test.go`](https://github.com/yunkeweb/go-word/blob/main/openxml_strict_test.go).

## Style IDs collide after a merge

Two source documents that both define a paragraph style named `Note` (or `Heading1`) cannot share one `styles.xml`. Without remapping, the second definition overwrites the first and every `w:pStyle` that still says `Note` picks up the wrong spacing, outline level, or numbering.

`AppendDocument` prefixes **colliding** paragraph and table style IDs. Unique names stay. The clone’s `w:pStyle` / `w:tblStyle` values are rewritten to match.

```go
err := dst.AppendDocument(src, word.MergeOptions{
	StylePrefix:    "src_",
	BookmarkPrefix: "src_",
	SectionBreak:   "nextPage",
})
```

| Identifier | Default prefix | Example |
| --- | --- | --- |
| Paragraph / table style IDs | `src_` | `Note` on the source becomes `src_Note` when `dst` already has `Note` |
| Bookmark names | `src_` | `shared` → `src_shared`; unique names stay |
| Internal hyperlink anchors | (derived) | `#shared` in the clone is rewritten to the new bookmark name |
| Images / media | write-time `imageN` | `RelationID` cleared; the writer assigns a fresh `rId` |

Pass a distinct prefix per source when you splice more than two trees (`note_`, `annex_`, …). Recipe 3 in [Enterprise Recipes](./recipes) does this for a three-part dossier.

`rId`s are allocated at write time (`word2007Writer.nextRel`), so two documents never share a relationship ID in the merged ZIP. You do not assign `rId` by hand.

## Placeholders are not replaced

Word often splits `${name}` across several `w:r` nodes when the user edits a template in the GUI. Generate the template with GoWord (`AddText("${name}")`) or keep each placeholder in one run. `NewTemplateProcessorBytes` on a document you just built never hits this.

Pipes (`${amount | formatCurrency:¥}`) run only on tokens the processor still sees as a single `${...}`. A split `formatCurrency` becomes literal text.

`CloneRow` looks for the placeholder inside a `w:tr` (and keeps `w:vMerge` groups together). `CloneBlock` looks for `${name}` … `${/name}`. Mixing the two on the same token fails — pick one.

## Images disappear after a merge

Two source documents both created `word/media/image1.png`. Without remapping, the second part overwrites the first. `AppendDocument` clears `RelationID` and the writer assigns `imageN` + a fresh `rId`. See [Document Merger](./merger).

## TOC / PAGE fields show “Error! Bookmark not defined.”

Fields store a cached result. Right-click → Update Field, or print. `AddTitle` must be used so outline levels exist. This is Word behaviour, not a missing part. `SetUpdateFields(true)` on a [TemplateProcessor](./template) asks Word to refresh fields on open (`w:updateFields`).

## Protect does not encrypt the ZIP

`Protect` writes `w:documentProtection`. Anyone can unzip the package. Use OS-level encryption or a passworded archive if the bytes must stay secret. The sample password in demos is `goword`.

## Two-column layout only applies to part of the page

`SetColumns` is a section property. Content after a new `AddSection` is a new `w:sectPr`. Put the column body in that section; keep titles in the previous one if they should stay full width. The separator attribute is `w:sep`, not `w:separator`.

## Which Word versions are supported?

Word 2007 through Microsoft 365 open the packages. DrawingML `wps:wsp` is a Word 2010+ feature; Word 2007 still opens the file and may drop the shape. Charts need Word 2007+. LibreOffice opens paragraphs, tables, and images; treat charts and `wps` shapes as Word-first.

## Stream extract vs `LoadBytes`

`Load` / `LoadBytes` builds the full document tree. `StreamExtractText` / `StreamExtractImages` walk `word/document.xml` with `xml.Decoder` and discard each paragraph buffer after the callback — extra heap stays O(1) relative to file size. Prefer `*os.File` so the ZIP is mapped without copying. See [Streaming](./streaming) and [Benchmarks](./benchmarks).

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
