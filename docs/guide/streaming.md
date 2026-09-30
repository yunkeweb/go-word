# O(1) Streaming Extractor

`StreamExtractText` and `StreamExtractImages` scan a `.docx` ZIP with `encoding/xml.Decoder`. Paragraph buffers are discarded after each callback, so extra memory stays **O(1)** relative to document size.

`os.File` is a sized `io.ReaderAt`, so the extractor maps the ZIP without copying the whole package. A plain `io.Reader` is buffered first. Rewind or reopen the file before a second pass.

## StreamExtractText / StreamExtractImages

### Signature

```go
func StreamExtractText(r io.Reader, fn func(paragraphText string) error) error
func StreamExtractImages(r io.Reader, fn func(img ImageFile) error) error
func StreamExtractTextWithOptions(r io.Reader, fn func(string) error, opts ReadOptions) error
func StreamExtractImagesWithOptions(r io.Reader, fn func(ImageFile) error, opts ReadOptions) error
func (d *Document) StreamExtractTextWithOptions(r io.Reader, fn func(string) error, opts ReadOptions) error
func (d *Document) StreamExtractImagesWithOptions(r io.Reader, fn func(ImageFile) error, opts ReadOptions) error
```

### Parameters

| Name | Type | Description |
| --- | --- | --- |
| `r` | `io.Reader` | `.docx` bytes. `*os.File` is preferred. |
| `fn` | callback | Return a non-nil error to stop the scan. |
| `opts` | `ReadOptions` | Per-call ZIP budgets for the `WithOptions` variants. |

`ImageFile` fields: `Name` (ZIP path), `MIME`, `Data`. Images are resolved through `a:blip r:embed` and VML `imagedata` via `word/_rels/document.xml.rels`.

### Optional ZIP budgets

Use additive `WithOptions` APIs when the package is not trusted:

| Field | Type | Limit |
| --- | --- | --- |
| `MaxArchiveSize` | `int64` | Compressed archive bytes. |
| `MaxPartSize` | `int64` | Uncompressed bytes in a single ZIP member. |
| `MaxTotalSize` | `int64` | Sum of declared uncompressed sizes, including unused members. |
| `MaxEntries` | `int` | Number of ZIP members, including directory entries. |

```go
opts := word.ReadOptions{
    MaxArchiveSize: 32 << 20,
    MaxPartSize:    16 << 20,
    MaxTotalSize:   64 << 20,
    MaxEntries:     512,
}
doc, err := word.ReadWithOptions(input, opts)
if errors.Is(err, word.ErrReadLimitExceeded) {
    // Reject the package according to your input policy.
}
```

The same options are accepted by `LoadWithOptions`, `OpenWithOptions`,
`LoadBytesWithOptions`, both template constructors, and both stream extractors.
Zero disables an individual limit; negative values return `ErrInvalidReadOptions`
before reading. Exact limits are accepted. Use `errors.Is` to recognize either
`ErrInvalidReadOptions` or `ErrReadLimitExceeded`. Declared sizes are checked
before decompression; actual member reads are bounded as well, and ZIP checksum
and format errors are preserved.
The budgets protect ZIP input and declared decompressed parts, not DOM memory or
CPU. Streaming callbacks may have observed earlier paragraphs before a later
part fails.

Complete bounded-read example:

```go
package main

import (
    "bytes"
    "errors"
    "fmt"
    "log"

    "github.com/yunkeweb/go-word"
)

func main() {
    source := word.New()
    source.AddSection().AddText("bounded input")
    raw, err := source.Bytes()
    if err != nil {
        log.Fatal(err)
    }

    opts := word.ReadOptions{MaxArchiveSize: 8 << 20, MaxPartSize: 4 << 20, MaxTotalSize: 16 << 20, MaxEntries: 256}
    if _, err := word.ReadWithOptions(bytes.NewReader(raw), opts); err != nil {
        if errors.Is(err, word.ErrReadLimitExceeded) {
            log.Fatal("document exceeds the configured ZIP budget")
        }
        log.Fatal(err)
    }

    paragraphs := 0
    err = word.StreamExtractTextWithOptions(bytes.NewReader(raw), func(text string) error {
        if text != "" {
            paragraphs++
        }
        return nil
    }, opts)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("read %d paragraph(s) under the configured budget\n", paragraphs)
}
```

Run the repository example with `go run ./examples/read_limits`. Its limits are
illustrative; choose all four budgets for your expected inputs.

## NewStreamWriter

### Signature

```go
func NewStreamWriter(dest io.Writer) *StreamWriter
func (s *StreamWriter) WriteParagraph(text string, styles ...any) error
func (s *StreamWriter) Close() error
```

`document.xml` opens on the first write. `Close` finishes the body and emits styles, content types, and relationships. Not safe for concurrent use.

## When to use which API

| API | Builds DOM | Memory |
| --- | --- | --- |
| `CreateReader("Word2007").Load` | Yes | Whole document |
| `StreamExtractText` / `StreamExtractImages` | No | O(1) extra |
| `NewStreamWriter` | Writes incrementally | Body XML is not fully buffered |

## Complete example

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

[`examples/v0.7.0_demo`](https://github.com/yunkeweb/go-word/tree/main/examples/v0.7.0_demo) streams 100 000 paragraphs. Numbers for `BenchmarkStreamWriter` are on [Benchmarks](./benchmarks).
