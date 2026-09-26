# Architecture

GoWord is a PHPWord-shaped façade over a strict OpenXML Word 2007 writer. Everything the public API does ends up as parts inside a ZIP: `word/document.xml`, `word/styles.xml`, `word/_rels/document.xml.rels`, media, headers, charts, and `[Content_Types].xml`.

## Object graph

```
Document
 ├── DocInfo / Settings / Compatibility
 ├── named styles (font, paragraph, table, title)
 └── Section[]                  // each section owns w:sectPr
      ├── Header[] / Footer[]  // separate ZIP parts
      └── Container children
           ├── Text / TextRun / Title / Link / Bookmark
           ├── Table → Row → Cell (Cell is a Container)
           ├── Image / DMLShape / Chart
           ├── Formula (OMML)
           └── TOC / Field / PreserveText
```

`Section`, `Header`, `Footer`, and `Cell` all embed `element.Container`. Nested tables work because a cell is a container: `cell.AddTable(...)`.

## Write path

1. Application code mutates the in-memory `Document`.
2. `Save` / `Bytes` / `WriteTo` construct a `word2007Writer`.
3. The writer streams `word/document.xml` with a pooled `XMLWriter`.
4. Relationship IDs (`rIdN`) and media names (`word/media/imageN`) are allocated **at write time**, never while you are still building the tree. That is why [AppendDocument](./merger) can clone two documents that both contain `image1.png`.
5. Remaining parts (styles, numbering, headers, charts, content types) are appended to the ZIP.

`NewStreamWriter` skips the in-memory body: paragraphs and tables go into `document.xml` as they are produced. See [Streaming Parser](./streaming).

## Read path

| API | Builds DOM | Extra memory |
| --- | --- | --- |
| `Load` / `Open` / `CreateReader("Word2007").Load` | Yes | Whole document |
| `StreamExtractText` / `StreamExtractImages` | No | O(1) relative to file size |
| `NewTemplateProcessor` / `NewTemplateProcessorBytes` | XML strings of ZIP parts | Template parts only |

The DOM reader reconstructs sections, tables (including nested `w:tbl` inside `w:tc`), bookmarks, and images. The stream extractors use `encoding/xml.Decoder` and drop each paragraph buffer after the callback returns.

## Math and charts as extra parts

- `AddMath` parses LaTeX in `pkg/math` and embeds `m:oMathPara` / `m:oMath` in the body. No extra ZIP part.
- `AddChart` appends a `word/charts/chartN.xml` part plus a relationship. Series XML follows the Word chart XSD order (`idx` → `order` → `tx` → `spPr` → …).

## Complete example — inspect the package

`Bytes` returns the ZIP. `bytes.NewReader` implements `io.ReaderAt`, so `zip.NewReader` can list parts without a temp file.

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
	sec.AddTitle("Architecture", 1)
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
		fmt.Println(f.Name)
	}
}
```

The listing always includes `[Content_Types].xml`, `_rels/.rels`, `word/document.xml`, `word/styles.xml`, and `word/_rels/document.xml.rels`. Adding a chart or header adds further parts.

## Related

[OpenXML Compatibility](./compatibility) lists the schemas Word actually accepts. [FAQ](./faq) covers repair dialogs when a part violates those schemas.
