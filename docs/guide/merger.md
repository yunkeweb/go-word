# Document Merger

`AppendDocument` clones every section of `src` onto `dst`, then remaps identifiers that would collide inside one OpenXML package.

```go
err := dst.AppendDocument(src, word.MergeOptions{
	StylePrefix:    "src_",
	BookmarkPrefix: "src_",
	SectionBreak:   "nextPage",
})
```

## What gets remapped

| Identifier | Behaviour |
| --- | --- |
| Paragraph / table **style IDs** | Colliding names receive `StylePrefix` (`Note` → `src_Note`) |
| **Bookmark** names (`w:bookmarkStart`) | Colliding names are prefixed; unique names stay |
| Internal hyperlink **anchors** | Updated to the new bookmark names |
| **Images / media** | `RelationID` cleared; writer assigns fresh `rId` and `word/media/imageN` |
| Section break | First cloned section uses `SectionBreak` (`nextPage` by default) |

`rId`s are allocated at write time (`word2007Writer.nextRel`), so two documents never share a relationship ID in the merged ZIP. A `nil` source returns `word: nil source document`.

## Complete example

Save as `main.go` and run `go run .`. The file `merged.docx` keeps both `Note` styles, `shared` / `src_shared` bookmarks, and two PNG parts.

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

After the merge:

| Resource | Isolation |
| --- | --- |
| Style `Note` on A | Kept |
| Style `Note` on B | Rewritten to `src_Note` |
| Bookmark `shared` on A | Kept |
| Bookmark `shared` on B | Rewritten to `src_shared` |
| Images | `word/media/image1.png` and `image2.png` with distinct `rId` |

See [`examples/v0.8.0_demo`](https://github.com/yunkeweb/go-word/tree/main/examples/v0.8.0_demo) and [Examples & Recipes](./examples) Case B.
