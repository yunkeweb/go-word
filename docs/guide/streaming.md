# Streaming Parser

`StreamExtractText` and `StreamExtractImages` scan a `.docx` ZIP with `encoding/xml.Decoder`. Paragraph buffers are discarded after each callback, so extra memory stays **\(O(1)\)** relative to document size.

Package-level helpers and `Document` methods share the same implementation:

```go
f, err := os.Open("report.docx")
if err != nil {
	log.Fatal(err)
}
defer f.Close()

err = word.StreamExtractText(f, func(paragraphText string) error {
	fmt.Println(paragraphText)
	return nil
})
```

Rewind or reopen the file before a second pass:

```go
f, _ = os.Open("report.docx")
defer f.Close()

err = word.StreamExtractImages(f, func(img word.ImageFile) error {
	fmt.Println(img.Name, len(img.Data))
	return nil
})
```

`os.File` is a sized `io.ReaderAt`, so the extractor maps the ZIP without copying the whole package into memory. A plain `io.Reader` is buffered first.

## When to use which API

| API | Builds DOM | Memory |
| --- | --- | --- |
| `CreateReader("Word2007").Load` | Yes | Whole document |
| `StreamExtractText` / `StreamExtractImages` | No | \(O(1)\) extra |
| `NewStreamWriter` | Writes incrementally | Body XML is not fully buffered |

`NewStreamWriter` is the write-side counterpart: paragraphs and tables go into the ZIP as they are produced.

See [`examples/v0.7.0_demo`](https://github.com/yunkeweb/go-word/tree/main/examples/v0.7.0_demo).
