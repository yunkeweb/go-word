# Advanced Layout & Security

Section properties, watermarks, document protection, and a TOC field.

## Multi-column layout

`SetColumns` writes a schema-valid `w:cols` node (`num`, `space`, `sep`):

| Argument | Meaning |
| --- | --- |
| `num` | Column count (`w:num`) |
| `space` | Gap in twips (`w:space`, default 720) |
| `showLine` | Separator line (`w:sep="1"`) |

Columns apply to the **section**, so start a new section when the rest of the document should go back to one column. Mixed portrait / landscape is the same idea: `AddSection` with `style.OrientationLandscape`.

## Watermark

Text watermarks are Word-native VML (`PowerPlusWaterMarkObject`) in every section header. Image watermarks store a media part referenced from that VML.

## Read-only protection

Modes: `ProtectTypeReadOnly`, `ProtectTypeComments`, `ProtectTypeTrackedChanges`, `ProtectTypeForms`. A non-empty password uses the Office SHA-1 / 100000-spin hash (ECMA-376 `w:documentProtection`).

## Table of contents

`AddTOC` / `AddTableOfContents` write a `TOC` field (`TOC \o "1-3" \h \z \u`). Heading paragraphs from `AddTitle` carry outline levels so Word can refresh the field.

## Complete example

Save as `main.go` and run `go run .`. The file `layout.docx` is read-only (`password` = `goword`) with a confidential watermark, a TOC, a two-column section, and a landscape section.

```go
package main

import (
	"log"

	"github.com/yunkeweb/go-word"
	"github.com/yunkeweb/go-word/element"
	"github.com/yunkeweb/go-word/style"
)

func main() {
	doc := word.New()
	info := doc.GetDocInfo()
	info.Title = "Advanced layout and protection"
	info.Creator = "GoWord"

	doc.SetDefaultFontName("Calibri")
	doc.SetDefaultAsianFontName("Microsoft YaHei")
	doc.SetDefaultFontSize(11)
	doc.AddTitleStyle(1, style.Font{Bold: true, Size: 16})
	doc.AddTitleStyle(2, style.Font{Bold: true, Size: 13})

	if err := doc.Protect(word.ProtectTypeReadOnly, "goword"); err != nil {
		log.Fatal(err)
	}
	doc.SetTextWatermark("CONFIDENTIAL")

	portrait := doc.AddSection()
	portrait.AddHeader().AddText("GoWord — odd pages")
	portrait.AddHeader(element.HeaderFirst).AddText("GoWord — first page")
	portrait.AddHeader(element.HeaderEven).AddText("GoWord — even pages")
	oddFooter := portrait.AddFooter()
	oddFooter.AddText("Page ")
	oddFooter.AddPreserveText("PAGE")
	portrait.AddFooter(element.HeaderFirst).AddText("Cover footer")
	doc.SetDifferentFirstPage(true)

	portrait.AddTitle("Table of Contents", 1)
	portrait.AddTOC(nil, nil, 1, 3)
	portrait.AddText("Right-click the TOC field in Microsoft Word and choose Update Field.")

	portrait.AddTitle("Portrait section", 1)
	portrait.AddText("This section is A4 portrait. w:titlePg selects the first-page header.")
	portrait.AddTitle("Document protection", 2)
	portrait.AddText("w:documentProtection edit=readOnly. Sample password: goword.")
	portrait.AddTitle("Watermark", 2)
	portrait.AddText("A diagonal CONFIDENTIAL watermark is stored as VML in every section header.")

	cols := doc.AddSection(style.Section{
		Orientation: style.OrientationPortrait,
		BreakType:   "nextPage",
		PageSizeW:   style.DefaultPageWidth,
		PageSizeH:   style.DefaultPageHeight,
		MarginTop:   1134, MarginBottom: 1134, MarginLeft: 1134, MarginRight: 1134,
	})
	cols.SetColumns(2, 720, true)
	cols.AddTitle("Two-column layout", 1)
	cols.AddText("The left column starts here. Word splits this section into two equal columns with a separator line (w:cols w:sep).")
	cols.AddText("The second paragraph continues the flow so the columns fill naturally.")

	land := doc.AddSection(style.Section{
		Orientation: style.OrientationLandscape,
		BreakType:   "nextPage",
		PageSizeW:   style.DefaultPageWidth,
		PageSizeH:   style.DefaultPageHeight,
		MarginTop:   style.DefaultMargin, MarginBottom: style.DefaultMargin,
		MarginLeft:  style.DefaultMargin, MarginRight:  style.DefaultMargin,
	})
	land.AddHeader().AddText("Landscape header")
	land.AddTitle("Landscape section", 1)
	land.AddText("w:orient=\"landscape\" on w:pgSz. Page width and height are swapped by Word.")

	if err := doc.AddMarkdown("# Markdown\n\nParagraph with **bold**."); err != nil {
		log.Fatal(err)
	}
	if err := doc.AddHTML("<h2>HTML</h2><p>Hello from AddHTML.</p>"); err != nil {
		log.Fatal(err)
	}

	if err := doc.Save("layout.docx"); err != nil {
		log.Fatal(err)
	}
}
```

OpenXML produced by this program:

| Feature | Node |
| --- | --- |
| Two columns | `w:cols w:num="2" w:space="720" w:sep="1"` |
| First-page header | `w:titlePg` + `w:headerReference w:type="first"` |
| Watermark | VML `PowerPlusWaterMarkObject` in header |
| Protection | `w:documentProtection w:edit="readOnly"` |
| TOC | `w:instrText` `TOC \o "1-3" \h \z \u` |

See [`examples/v0.5.0_demo`](https://github.com/yunkeweb/go-word/tree/main/examples/v0.5.0_demo) and [Examples & Recipes](./examples) Case A.
