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
| Watermark | VML `PowerPlusWaterMarkObject` in headers | Diagonal word-art |

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

`ExtractText` walks the reconstructed DOM. For files too large to load, use [StreamExtractText](./streaming).

## Related

When Word shows “the file is corrupt”, start at [FAQ](./faq). Chart XSD order is documented on [Charts](./charts).
