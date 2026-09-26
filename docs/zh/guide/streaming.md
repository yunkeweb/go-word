# O(1) 流式提取器

`StreamExtractText` 与 `StreamExtractImages` 使用 `encoding/xml.Decoder` 扫描 `.docx` ZIP。每个段落缓冲在回调后丢弃，因此相对文档体积的额外内存为 **O(1)**。

`os.File` 是带尺寸的 `io.ReaderAt`，提取器可以直接映射 ZIP。普通 `io.Reader` 会先缓冲。第二次遍历前请 `Seek` 回起点或重新打开文件。

## StreamExtractText / StreamExtractImages

### 签名

```go
func StreamExtractText(r io.Reader, fn func(paragraphText string) error) error
func StreamExtractImages(r io.Reader, fn func(img ImageFile) error) error
```

### 参数

| 名称 | 类型 | 说明 |
| --- | --- | --- |
| `r` | `io.Reader` | `.docx` 字节。优先 `*os.File`。 |
| `fn` | 回调 | 返回非 nil error 即停止扫描。 |

`ImageFile` 字段：`Name`（ZIP 路径）、`MIME`、`Data`。图片经 `a:blip r:embed` 与 VML `imagedata`，通过 `word/_rels/document.xml.rels` 解析。

## NewStreamWriter

### 签名

```go
func NewStreamWriter(dest io.Writer) *StreamWriter
func (s *StreamWriter) WriteParagraph(text string, styles ...any) error
func (s *StreamWriter) Close() error
```

第一次写入时打开 `document.xml`。`Close` 结束正文并写出样式、内容类型与关系。非并发安全。

## 选用哪套 API

| API | 构建 DOM | 内存 |
| --- | --- | --- |
| `CreateReader("Word2007").Load` | 是 | 整份文档 |
| `StreamExtractText` / `StreamExtractImages` | 否 | O(1) 额外 |
| `NewStreamWriter` | 增量写出 | 正文 XML 不全量缓冲 |

## 完整示例

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

[`examples/v0.7.0_demo`](https://github.com/yunkeweb/go-word/tree/main/examples/v0.7.0_demo) 会流式写出 10 万段。`BenchmarkStreamWriter` 的数字见 [基准测试](./benchmarks)。
