# Basic DOM

A `Document` holds named styles and one or more `Section`s. Almost every body element is added on a container (`Section`, `Header`, `Footer`, `Cell`, `TextRun`).

## Document and section

```go
doc := word.New()
doc.SetDefaultFontName("Calibri")
doc.SetDefaultFontSize(11)

info := doc.GetDocInfo()
info.Title = "Quarterly report"
info.Creator = "GoWord"

sec := doc.AddSection(style.Section{
	Orientation: style.OrientationPortrait,
	MarginTop:   1440, MarginBottom: 1440,
	MarginLeft:  1440, MarginRight:  1440,
})
```

`Save` / `WriteTo` stream `word/document.xml` into the ZIP. `IOFactory` names stay PHPWord-compatible:

```go
w, err := word.CreateWriter(doc, "Word2007")
```

## Paragraph and rich text

```go
doc.AddFontStyle("strong", style.Font{Bold: true, Size: 14, Name: "Calibri"})
doc.AddParagraphStyle("center", style.Paragraph{Alignment: style.JcCenter})
doc.AddTitleStyle(1, style.Font{Bold: true, Size: 18}, style.Paragraph{
	Spacing: style.Spacing{After: 240},
})

sec.AddTitle("Welcome to GoWord", 1)
sec.AddText("Hello, Word 2007.", "strong", "center")

run := sec.AddTextRun()
run.AddText("Bold ", style.Font{Bold: true})
run.AddText("and italic.", style.Font{Italic: true})

sec.AddLink("https://github.com/yunkeweb/go-word", "GoWord on GitHub")
sec.AddBookmark("intro")
sec.AddPageBreak()
```

## Table

```go
tbl := sec.AddTable(style.Table{Width: 9000})
hdr := tbl.AddRow()
hdr.AddCell(4500).AddText("Item")
hdr.AddCell(4500).AddText("Amount")

row := tbl.AddRow()
row.AddCell(4500).AddText("Paper")
row.AddCell(4500).AddText("12")
```

Cells are containers: nested tables, images, and formulas can live inside a `Cell`. `gridSpan` / `vMerge` keep merged regions together when cloning template rows.

## Image

```go
sec.AddImage("photo.png", style.Image{Width: 200, Height: 120})
sec.AddImageBytes("logo.png", pngBytes, style.Image{Width: 80, Height: 80})
```

Media parts are written as `word/media/imageN.ext` with unique `rId`s at save time.

## Header and footer

```go
h := sec.AddHeader()
h.AddText("GoWord — confidential")

f := sec.AddFooter()
f.AddPreserveText("Page {PAGE} of {NUMPAGES}")

sec.AddHeader(element.HeaderFirst) // first-page header
doc.SetDifferentFirstPage(true)
doc.SetEvenAndOddHeaders(true)
```

Header types: `element.HeaderAuto` (default), `HeaderFirst`, `HeaderEven`.

## Load an existing file

```go
r, err := word.CreateReader("Word2007")
if err != nil {
	log.Fatal(err)
}
doc, err := r.Load("input.docx")
```

See also [Streaming Parser](./streaming) for \(O(1)\) text and image extraction without building a full DOM.
