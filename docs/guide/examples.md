# Examples & Recipes

Three complete programs you can save as `main.go` and run with `go run .`. Each one uses public APIs from v0.8.0.

| Case | Output | Features |
| --- | --- | --- |
| A | `academic-report.docx` | OMML formulas, two-column body, TOC, header/footer |
| B | `spliced.docx` | `AppendDocument` with colliding styles, bookmarks, and PNG parts |
| C | `invoice-INV-1042.docx`, `invoice-INV-1043.docx` | In-memory template, pipe filters, `${block}` / `${if}` |

## Case A — Academic / engineering report

Formulas as native Office Math, a two-column methods section (`w:cols w:num="2" w:space="720" w:sep="1"`), and a TOC field.

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
	info.Title = "Stress analysis of a simply supported beam"
	info.Creator = "GoWord"
	info.Subject = "Academic report with OMML and two-column layout"

	doc.SetDefaultFontName("Calibri")
	doc.SetDefaultAsianFontName("Microsoft YaHei")
	doc.SetDefaultFontSize(11)
	doc.AddTitleStyle(1, style.Font{Bold: true, Size: 16})
	doc.AddTitleStyle(2, style.Font{Bold: true, Size: 13})

	cover := doc.AddSection()
	h := cover.AddHeader()
	h.AddText("GoWord academic report")
	f := cover.AddFooter()
	f.AddText("Page ")
	f.AddPreserveText("PAGE")
	cover.AddHeader(element.HeaderFirst).AddText("Cover")
	doc.SetDifferentFirstPage(true)

	cover.AddTitle("Stress analysis of a simply supported beam", 1)
	cover.AddText("An engineering note that mixes native Office Math with a two-column methods section.")
	cover.AddTitle("Contents", 2)
	cover.AddTOC(nil, nil, 1, 2)
	cover.AddText("Right-click the TOC field in Microsoft Word and choose Update Field.")

	cover.AddTitle("Governing equations", 1)
	cover.AddText("Maximum bending stress:")
	cover.AddMath(`\sigma = \frac{M y}{I}`)
	p := cover.AddTextRun()
	p.AddText("Second moment of area for a rectangle: ")
	p.AddMath(`I = \frac{b h^{3}}{12}`)
	cover.AddText("Shear formula:")
	cover.AddMath(`\tau = \frac{V Q}{I t}`)

	methods := doc.AddSection(style.Section{
		Orientation: style.OrientationPortrait,
		BreakType:   "nextPage",
		PageSizeW:   style.DefaultPageWidth,
		PageSizeH:   style.DefaultPageHeight,
		MarginTop:   1134, MarginBottom: 1134, MarginLeft: 1134, MarginRight: 1134,
	})
	methods.SetColumns(2, 720, true)
	methods.AddTitle("Methods", 1)
	methods.AddText("The left column records the loading: a concentrated force at mid-span on a simply supported beam of length L.")
	methods.AddText("The right column continues with the section properties. Word flows these paragraphs into two equal columns with a separator line.")
	run := methods.AddTextRun()
	run.AddText("Check: ")
	run.AddMath(`\sqrt{x_1} + \pi`)

	if err := doc.Save("academic-report.docx"); err != nil {
		log.Fatal(err)
	}
}
```

OpenXML: display math is `m:oMathPara`; inline math is `m:oMath` beside `w:r`; columns are `w:cols` on the second section's `w:sectPr`.

## Case B — Lossless multi-document splice

Two source documents share a paragraph style named `Note` and a bookmark named `shared`. `AppendDocument` prefixes the colliding identifiers and assigns unique image `rId`s at write time.

```go
package main

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"log"

	"github.com/yunkeweb/go-word"
	"github.com/yunkeweb/go-word/style"
)

func main() {
	chapter1 := word.New()
	chapter1.AddParagraphStyle("Note", style.Paragraph{Spacing: style.Spacing{After: 120}})
	s1 := chapter1.AddSection()
	s1.AddTitle("Chapter 1 — Specimens", 1)
	s1.AddText("Macrograph from the first lab.", nil, "Note")
	s1.AddBookmark("shared")
	s1.AddImageBytes("lab-a.png", swatch(color.RGBA{R: 31, G: 78, B: 121, A: 255}), style.Image{Width: 80, Height: 40, AltText: "lab A"})

	chapter2 := word.New()
	chapter2.AddParagraphStyle("Note", style.Paragraph{Spacing: style.Spacing{After: 240}})
	s2 := chapter2.AddSection()
	s2.AddTitle("Chapter 2 — Fracture surface", 1)
	s2.AddText("Macrograph from the second lab, with a different Note style.", nil, "Note")
	s2.AddBookmark("shared")
	s2.AddImageBytes("lab-b.png", swatch(color.RGBA{R: 237, G: 125, B: 49, A: 255}), style.Image{Width: 80, Height: 40, AltText: "lab B"})
	s2.AddMath(`\frac{\sigma}{E}`)

	if err := chapter1.AppendDocument(chapter2, word.MergeOptions{
		StylePrefix:    "ch2_",
		BookmarkPrefix: "ch2_",
		SectionBreak:   "nextPage",
	}); err != nil {
		log.Fatal(err)
	}
	if err := chapter1.Save("spliced.docx"); err != nil {
		log.Fatal(err)
	}
}

func swatch(c color.RGBA) []byte {
	img := image.NewRGBA(image.Rect(0, 0, 80, 40))
	for y := 0; y < 40; y++ {
		for x := 0; x < 80; x++ {
			img.Set(x, y, c)
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		log.Fatal(err)
	}
	return buf.Bytes()
}
```

After the splice: chapter 1 keeps `Note` and `shared`; chapter 2 becomes `ch2_Note` and `ch2_shared`; media lands in `word/media/image1.png` and `image2.png`.

## Case C — Finance / contract batch export

Build the template with GoWord, fill it through `NewTemplateProcessorBytes`, and emit one invoice per record. Pipe filters format dates and currency; `${block}` clones line items; `${if}` clips overdue / paid copy.

```go
package main

import (
	"fmt"
	"log"

	"github.com/yunkeweb/go-word"
)

func main() {
	raw, err := invoiceTemplate()
	if err != nil {
		log.Fatal(err)
	}

	type line struct{ SKU, Qty, Amount string }
	type invoice struct {
		ID, Client, When, Total string
		Paid, Overdue           bool
		Lines                   []line
	}
	batch := []invoice{
		{
			ID: "INV-1042", Client: "  acme labs  ", When: "2026-09-01T00:00:00Z",
			Total: "279.50", Paid: true, Overdue: false,
			Lines: []line{{"A-01", "2", "199.50"}, {"B-02", "1", "80.00"}},
		},
		{
			ID: "INV-1043", Client: "beta works", When: "2026-08-15T00:00:00Z",
			Total: "50.00", Paid: false, Overdue: true,
			Lines: []line{{"C-09", "1", "50.00"}},
		},
	}

	for _, inv := range batch {
		tp, err := word.NewTemplateProcessorBytes(raw)
		if err != nil {
			log.Fatal(err)
		}
		tp.SetValues(map[string]string{
			"invoice": inv.ID,
			"client":  inv.Client,
			"when":    inv.When,
			"total":   inv.Total,
		})
		items := make([]word.BlockData, 0, len(inv.Lines))
		for _, ln := range inv.Lines {
			items = append(items, word.BlockData{
				Values: map[string]string{"sku": ln.SKU, "qty": ln.Qty, "amount": ln.Amount},
			})
		}
		if err := tp.CloneNestedBlock("items", items); err != nil {
			log.Fatal(err)
		}
		if err := tp.SetCondition("paid", inv.Paid); err != nil {
			log.Fatal(err)
		}
		if err := tp.SetCondition("overdue", inv.Overdue); err != nil {
			log.Fatal(err)
		}
		out := fmt.Sprintf("invoice-%s.docx", inv.ID)
		if err := tp.Save(out); err != nil {
			log.Fatal(err)
		}
		fmt.Println("wrote", out)
	}
}

func invoiceTemplate() ([]byte, error) {
	src := word.New()
	sec := src.AddSection()
	sec.AddTitle("Invoice ${invoice | trim | upper}", 1)
	sec.AddText("Client: ${client | trim | upper}")
	sec.AddText("Date: ${when | formatDate:\"2006-01-02\"}")
	sec.AddText("${items}")
	sec.AddText("${sku} × ${qty} = ${amount | formatCurrency:¥}")
	sec.AddText("${/items}")
	sec.AddText("Total ${total | formatCurrency:¥}")
	sec.AddText("${if paid}Thank you for your payment. This invoice is closed.${endif}")
	sec.AddText("${if overdue}OVERDUE — please settle this contract immediately.${endif}")
	return src.Bytes()
}
```

Each `${...}` stays in a single `w:t` run. `CloneNestedBlock` suffixes nested macros with `#n` so line-item values do not leak across invoices.

## Related guides

- [Office Math](./math) — LaTeX → OMML used in Case A
- [Document Merger](./merger) — style / bookmark / media isolation used in Case B
- [Template Engine v2](./template) — pipes, blocks, and `${if}` used in Case C
- [Advanced Layout](./layout) — columns, TOC, headers
