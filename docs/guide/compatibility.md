# OpenXML Compatibility

GoWord writes **Office Open XML WordprocessingML 2007** packages (`ECMA-376` / ISO/IEC 29500 transitional). Microsoft Word 2007 through Microsoft 365 open the files. WPS Office and LibreOffice open the common subset (paragraphs, tables, images); DrawingML charts and `wps:wsp` shapes are Word-oriented.

## Package layout

| Part | Required | Notes |
| --- | --- | --- |
| `[Content_Types].xml` | Yes | Overrides for `document.xml`, headers, charts, media |
| `_rels/.rels` | Yes | Points at `word/document.xml` |
| `word/document.xml` | Yes | Body + last `w:sectPr` |
| `word/_rels/document.xml.rels` | Yes | Images, headers, footers, charts, numbering |
| `word/styles.xml` | Yes | Named styles plus default `Normal` |
| `word/media/imageN.*` | If images | Names assigned at write time |
| `word/charts/chartN.xml` | If charts | One part per `AddChart` |
| `word/headerN.xml` / `footerN.xml` | If headers | Referenced from `w:sectPr` via `r:id` |

## Schemas the writer targets

| Feature | Namespace / node | Word behaviour |
| --- | --- | --- |
| Body text | `w:` WordprocessingML | Paragraphs, tables, sectPr |
| Native equations | `m:` Office Math (`m:oMathPara`, `m:oMath`) | Double-click opens the equation editor |
| Charts | `c:` ECMA-376 Chart | Series order is `idx` → `order` → `tx` → `spPr` |
| Word 2010 shapes | `wps:wsp` / `a:prstGeom` | Rect, roundRect, rightArrow, text box |
| Fields | `w:instrText` (`PAGE`, `TOC`, `NUMPAGES`) | Refresh with Update Field |
| Protection | `w:documentProtection` | Read-only / comments / tracked changes / forms |
| Edit exceptions | `w:permStart` / `w:permEnd` | `AllowEdit` ranges stay writable |
| Watermark | VML `PowerPlusWaterMarkObject` in headers | Diagonal or tiled word-art; `WordPictureWatermark` for images |
| SDT | `w:sdt` / `w:sdtPr` / `w:sdtContent` | Plain text, drop-down, date, `w14:checkbox` |
| Table row / cell | `w:tblHeader`, `w:cantSplit`, `w:vAlign`, `w:textDirection` | Repeating headers, unbreakable rows, vertical text |

## Strict rules the test suite enforces

The package `openxml_strict_test.go` rejects output that Word would offer to repair:

- Every `w:tc` contains at least one `w:p`.
- Template replacements never emit raw `&` inside `w:t` (no `&#80;` style numeric entities for ordinary ASCII).
- Combo charts use two distinct `c:axId` values; the secondary series points at the second value axis.
- Area charts omit `c:dLblPos` (the area XSD does not allow it).
- `c:legend` sits after `c:plotArea`, never inside it.
- Line markers emit `c:symbol` plus `c:size` 5.
- `w:cols` uses `w:sep`, not `w:separator`.

## Complete example — round-trip through the reader

```go
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/yunkeweb/go-word"
)

func main() {
	doc := word.New()
	sec := doc.AddSection()
	sec.AddTitle("Compatibility", 1)
	sec.AddText("Round-trip through CreateReader.")
	sec.AddMath(`x^{2}`)
	if err := doc.Save("compat.docx"); err != nil {
		log.Fatal(err)
	}

	r, err := word.CreateReader("Word2007")
	if err != nil {
		log.Fatal(err)
	}
	loaded, err := r.Load("compat.docx")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(loaded.ExtractText())

	_ = os.Remove("compat.docx")
}
```

`ExtractText` walks the reconstructed DOM. For files too large to load, use [StreamExtractText](./streaming). There is no `word.ReadDOM`; the DOM loaders are `word.Open`, `word.Read`, and `word.Load`.

## Full-feature matrix (v0.9.0)

[`tests/matrix`](https://github.com/yunkeweb/go-word/tree/main/tests/matrix) writes 80 randomly combined documents covering every exported module from v0.1.0 through v0.9.0: typography, multi-section layout, headers/footers, tables, images/shapes, TOC/bookmarks/comments, OMML, charts, SDT, and watermark/protection. Output stays in `./test_output_docs` (gitignored).

```sh
go run ./tests/matrix
go run tests/matrix/validate_reader.go
powershell -NoProfile -ExecutionPolicy Bypass -File tests/matrix/validate_docs.ps1
```

| Engine | What it checks | v0.9.0 result |
| --- | --- | --- |
| `word.Open` / `word.Read` + `ExtractText` / `ExtractImages` | DOM reverse-parse, no panic, `error == nil` | 80 PASS |
| `word.StreamExtractText` / `word.StreamExtractImages` | Streaming extract, no panic, `error == nil` | 80 PASS |
| Microsoft Word COM (`DisplayAlerts=0`, `OpenNoRepairDialog`) | OpenXML repair dialogs, node order, parse exceptions | 80 PASS |

`validate_docs.ps1` starts a headless `Word.Application` and opens each file read-only. A repair that Word would normally dialog becomes an exception.

## Related

When Word shows “the file is corrupt”, start at [FAQ](./faq). Chart XSD order is documented on [Charts](./charts). Recipe 4 on [Enterprise Recipes](./recipes) combines SDT, repeating headers, tiled watermarks, and `AllowEdit`.
