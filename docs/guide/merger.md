# Lossless Document Merger

`AppendDocument` clones every section of `src` onto `dst`, then remaps identifiers that would collide inside one OpenXML package. The result is a single ZIP whose styles, bookmarks, and media do not overwrite each other.

## AppendDocument

### Signature

```go
func (d *Document) AppendDocument(src *Document, opts MergeOptions) error

type MergeOptions struct {
	StylePrefix    string
	BookmarkPrefix string
	SectionBreak   string
}
```

### Parameters

| Name | Type | Description |
| --- | --- | --- |
| `src` | `*Document` | Source tree. `nil` returns `word: nil source document`. |
| `StylePrefix` | `string` | Prefix for colliding paragraph/table style IDs. Default `src_`. |
| `BookmarkPrefix` | `string` | Prefix for colliding `w:bookmarkStart` names. Default `src_`. |
| `SectionBreak` | `string` | Break on the first cloned section. Default `nextPage`. |

### What gets remapped

| Identifier | Behaviour |
| --- | --- |
| Paragraph / table style IDs | Colliding names receive the prefix (`Note` → `src_Note`) |
| Bookmark names | Colliding names are prefixed; unique names stay |
| Internal hyperlink anchors | Updated to the new bookmark names |
| Images / media | `RelationID` cleared; writer assigns fresh `rId` and `word/media/imageN` |

`rId`s are allocated at write time (`word2007Writer.nextRel`), so two documents never share a relationship ID in the merged ZIP.

## Complete example

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
	dst := word.New()
	dst.AddParagraphStyle("Note", style.Paragraph{Spacing: style.Spacing{After: 120}})
	a := dst.AddSection()
	a.AddTitle("Document A", 1)
	a.AddText("Bookmark and image from A.", nil, "Note")
	a.AddBookmark("shared")
	a.AddImageBytes("a.png", swatch(color.RGBA{R: 31, G: 78, B: 121, A: 255}), style.Image{Width: 40, Height: 20})

	src := word.New()
	src.AddParagraphStyle("Note", style.Paragraph{Spacing: style.Spacing{After: 200}})
	b := src.AddSection()
	b.AddTitle("Document B", 1)
	b.AddText("Bookmark, image and formula from B.", nil, "Note")
	b.AddBookmark("shared")
	b.AddImageBytes("b.png", swatch(color.RGBA{R: 237, G: 125, B: 49, A: 255}), style.Image{Width: 40, Height: 20})
	b.AddMath(`\frac{1}{2}`)

	if err := dst.AppendDocument(src, word.MergeOptions{SectionBreak: "nextPage"}); err != nil {
		log.Fatal(err)
	}
	if err := dst.Save("merged.docx"); err != nil {
		log.Fatal(err)
	}
}

func swatch(c color.RGBA) []byte {
	img := image.NewRGBA(image.Rect(0, 0, 40, 20))
	for y := 0; y < 20; y++ {
		for x := 0; x < 40; x++ {
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

### What Word shows

Two pages. Page 1 is navy-blue swatch + Document A. Page 2 starts after a next-page section break, shows the orange swatch, and contains an editable `1/2` fraction. Styles.xml contains both `Note` and `src_Note` (different after-spacing). Bookmarks `shared` and `src_shared` both exist. The ZIP has `word/media/image1.png` and `image2.png` with distinct `rId`s — unzipping never overwrites a picture.
