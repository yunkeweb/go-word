# 流式提取器

`StreamExtractText` 与 `StreamExtractImages` 用 `encoding/xml.Decoder` 扫描 `.docx` ZIP。每个段落缓冲在回调后丢弃，相对文档体积的额外内存为 **\(O(1)\)**。

包级函数与 `Document` 方法是同一套实现：

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

第二次扫描前请重新打开（或 Seek 回文件头）：

```go
f, _ = os.Open("report.docx")
defer f.Close()

err = word.StreamExtractImages(f, func(img word.ImageFile) error {
	fmt.Println(img.Name, len(img.Data))
	return nil
})
```

`os.File` 实现了带尺寸的 `io.ReaderAt`，提取器可以直接映射 ZIP，不必把整个包拷进内存。普通 `io.Reader` 会先被缓冲。

## 选用哪套 API

| API | 构建 DOM | 内存 |
| --- | --- | --- |
| `CreateReader("Word2007").Load` | 是 | 整份文档 |
| `StreamExtractText` / `StreamExtractImages` | 否 | \(O(1)\) 额外内存 |
| `NewStreamWriter` | 增量写出 | 正文 XML 不全量缓冲 |

`NewStreamWriter` 是写出侧对应物：段落与表格边生成边写入 ZIP。

示例：[`examples/v0.7.0_demo`](https://github.com/yunkeweb/go-word/tree/main/examples/v0.7.0_demo)。
