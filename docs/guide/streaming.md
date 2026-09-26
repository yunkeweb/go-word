# Streaming Parser

`StreamExtractText` and `StreamExtractImages` scan a `.docx` ZIP with `encoding/xml.Decoder`. Paragraph buffers are discarded after each callback, so extra memory stays **O(1)** relative to document size.

`os.File` is a sized `io.ReaderAt`, so the extractor maps the ZIP without copying the whole package into memory. A plain `io.Reader` is buffered first. Rewind or reopen the file before a second pass.

## When to use which API

| API | Builds DOM | Memory |
| --- | --- | --- |
| `CreateReader("Word2007").Load` | Yes | Whole document |
| `StreamExtractText` / `StreamExtractImages` | No | O(1) extra |
| `NewStreamWriter` | Writes incrementally | Body XML is not fully buffered |

`ImageFile` fields: `Name` (ZIP path), `MIME`, `Data`. `StreamExtractImages` follows `a:blip r:embed` and VML `imagedata` through `word/_rels/document.xml.rels`.

## Complete example

Save as `main.go` and run `go run .`. The program writes `stream-src.docx`, prints every paragraph and image, then streams 1 000 paragraphs through `NewStreamWriter`.

```go
package main

import (
	"bytes"
	"fmt"
	"log"
	"os"

	"github.com/yunkeweb/go-word"
	"github.com/yunkeweb/go-word/style"
)

func main() {
	doc := word.New()
	sec := doc.AddSection()
	sec.AddTitle("Streaming source", 1)
	sec.AddText("First paragraph for StreamExtractText.")
	sec.AddText("Second paragraph.")
	sec.AddImageBytes("dot.png", tinyPNG(), style.Image{Width: 32, Height: 32, AltText: "swatch"})
	if err := doc.Save("stream-src.docx"); err != nil {
		log.Fatal(err)
	}

	f, err := os.Open("stream-src.docx")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	nPara := 0
	if err := word.StreamExtractText(f, func(paragraphText string) error {
		if paragraphText == "" {
			return nil
		}
		nPara++
		fmt.Println("p:", paragraphText)
		return nil
	}); err != nil {
		log.Fatal(err)
	}

	if _, err := f.Seek(0, 0); err != nil {
		log.Fatal(err)
	}
	nImg := 0
	if err := word.StreamExtractImages(f, func(img word.ImageFile) error {
		nImg++
		fmt.Printf("image %s %s %d bytes\n", img.Name, img.MIME, len(img.Data))
		return nil
	}); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("extracted %d paragraphs, %d images\n", nPara, nImg)

	var buf bytes.Buffer
	sw := word.NewStreamWriter(&buf)
	for i := 0; i < 1000; i++ {
		if err := sw.WriteParagraph(fmt.Sprintf("row %d", i+1)); err != nil {
			log.Fatal(err)
		}
	}
	if err := sw.Close(); err != nil {
		log.Fatal(err)
	}
	n := 0
	if err := word.StreamExtractText(bytes.NewReader(buf.Bytes()), func(s string) error {
		if s != "" {
			n++
		}
		return nil
	}); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("stream writer produced %d paragraphs, zip %d bytes\n", n, buf.Len())
}

func tinyPNG() []byte {
	return []byte{
		0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a, 0x00, 0x00, 0x00, 0x0d,
		0x49, 0x48, 0x44, 0x52, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
		0x08, 0x02, 0x00, 0x00, 0x00, 0x90, 0x77, 0x53, 0xde, 0x00, 0x00, 0x00,
		0x0c, 0x49, 0x44, 0x41, 0x54, 0x08, 0xd7, 0x63, 0xf8, 0xcf, 0xc0, 0x00,
		0x00, 0x00, 0x03, 0x00, 0x01, 0x00, 0x05, 0xfe, 0xd4, 0xef, 0x00, 0x00,
		0x00, 0x00, 0x49, 0x45, 0x4e, 0x44, 0xae, 0x42, 0x60, 0x82,
	}
}
```

See [`examples/v0.7.0_demo`](https://github.com/yunkeweb/go-word/tree/main/examples/v0.7.0_demo) for a 100 000-paragraph run.
