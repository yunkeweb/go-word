# O(1) 流式提取器

`StreamExtractText` 与 `StreamExtractImages` 使用 `encoding/xml.Decoder` 扫描 `.docx` ZIP。每个段落缓冲在回调后丢弃，因此相对文档体积的额外内存为 **O(1)**。

`os.File` 是带尺寸的 `io.ReaderAt`，提取器可以直接映射 ZIP。普通 `io.Reader` 会先缓冲。第二次遍历前请 `Seek` 回起点或重新打开文件。

## StreamExtractText / StreamExtractImages

### 签名

```go
func StreamExtractText(r io.Reader, fn func(paragraphText string) error) error
func StreamExtractImages(r io.Reader, fn func(img ImageFile) error) error
func StreamExtractTextWithOptions(r io.Reader, fn func(string) error, opts ReadOptions) error
func StreamExtractImagesWithOptions(r io.Reader, fn func(ImageFile) error, opts ReadOptions) error
func (d *Document) StreamExtractTextWithOptions(r io.Reader, fn func(string) error, opts ReadOptions) error
func (d *Document) StreamExtractImagesWithOptions(r io.Reader, fn func(ImageFile) error, opts ReadOptions) error
```

### 参数

| 名称 | 类型 | 说明 |
| --- | --- | --- |
| `r` | `io.Reader` | `.docx` 字节。优先 `*os.File`。 |
| `fn` | 回调 | 返回非 nil error 即停止扫描。 |
| `opts` | `ReadOptions` | `WithOptions` 入口的单次 ZIP 预算。 |

`ImageFile` 字段：`Name`（ZIP 路径）、`MIME`、`Data`。图片经 `a:blip r:embed` 与 VML `imagedata`，通过 `word/_rels/document.xml.rels` 解析。

### 可选 ZIP 预算

处理不可信文档时，可使用新增的 `WithOptions` 入口：

`opts` 为单次调用的 `ReadOptions`，各字段含义如下：

| 字段 | 类型 | 限制 |
| --- | --- | --- |
| `MaxArchiveSize` | `int64` | 压缩包字节数。 |
| `MaxPartSize` | `int64` | 单个 ZIP 部件的解压字节数。 |
| `MaxTotalSize` | `int64` | 所有部件声明的解压大小之和，包含未使用的部件。 |
| `MaxEntries` | `int` | ZIP 条目数，包含目录条目。 |

```go
opts := word.ReadOptions{
    MaxArchiveSize: 32 << 20,
    MaxPartSize:    16 << 20,
    MaxTotalSize:   64 << 20,
    MaxEntries:     512,
}
doc, err := word.ReadWithOptions(input, opts)
if errors.Is(err, word.ErrReadLimitExceeded) {
    // 按输入策略拒绝文档。
}
```

同一选项也适用于 `LoadWithOptions`、`OpenWithOptions`、
`LoadBytesWithOptions`、两个模板构造函数和两个流式提取入口。
单项值为 0 表示关闭限制；负数会在读取前返回 `ErrInvalidReadOptions`，恰好等于
上限的输入允许通过。可用 `errors.Is` 判断 `ErrInvalidReadOptions` 与
`ErrReadLimitExceeded`。解压前检查声明大小，实际部件读取也受上限约束，
并保留 ZIP 校验和及格式错误。预算限制 ZIP 输入和声明的
解压部件大小，不会限制 DOM 内存或 CPU；流式回调可能已经收到前面的段落，
之后才发现部件超限。

完整的预算读取示例：

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

在仓库根目录执行 `go run ./examples/read_limits` 可运行示例。示例中的数值仅供
演示，处理不可信输入时请根据业务文档大小同时配置四项预算。

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
