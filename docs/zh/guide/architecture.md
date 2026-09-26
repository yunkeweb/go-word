# 架构设计

GoWord 是套在严格 OpenXML Word 2007 写出器上的 PHPWord 风格门面。公开 API 的一切最终都落成 ZIP 中的部件：`word/document.xml`、`word/styles.xml`、`word/_rels/document.xml.rels`、媒体、页眉、图表和 `[Content_Types].xml`。

## 对象图

```
Document
 ├── DocInfo / Settings / Compatibility
 ├── 命名样式（字体、段落、表格、标题）
 └── Section[]                  // 每节拥有自己的 w:sectPr
      ├── Header[] / Footer[]  // 独立 ZIP 部件
      └── Container 子节点
           ├── Text / TextRun / Title / Link / Bookmark
           ├── Table → Row → Cell（Cell 也是 Container）
           ├── Image / DMLShape / Chart
           ├── Formula（OMML）
           └── TOC / Field / PreserveText
```

`Section`、`Header`、`Footer`、`Cell` 都嵌入 `element.Container`。嵌套表能工作，是因为单元格本身就是容器：`cell.AddTable(...)`。

## 写出路径

1. 应用代码修改内存中的 `Document`。
2. `Save` / `Bytes` / `WriteTo` 构造 `word2007Writer`。
3. 写出器用池化的 `XMLWriter` 流式写入 `word/document.xml`。
4. 关系 ID（`rIdN`）与媒体名（`word/media/imageN`）在**写出时**分配，而不是在搭树时分配。因此 [AppendDocument](./merger) 可以克隆两份都含 `image1.png` 的文档。
5. 其余部件（样式、编号、页眉、图表、内容类型）追加进 ZIP。

`NewStreamWriter` 跳过内存正文：段落和表格一边产生一边写入 `document.xml`。见 [流式解析器](./streaming)。

## 读取路径

| API | 构建 DOM | 额外内存 |
| --- | --- | --- |
| `Load` / `Open` / `CreateReader("Word2007").Load` | 是 | 整份文档 |
| `StreamExtractText` / `StreamExtractImages` | 否 | 相对文件大小 O(1) |
| `NewTemplateProcessor` / `NewTemplateProcessorBytes` | ZIP 部件的 XML 字符串 | 仅模板部件 |

DOM 读取器重建节、表格（含 `w:tc` 内的嵌套 `w:tbl`）、书签与图片。流式提取器使用 `encoding/xml.Decoder`，回调返回后丢弃该段缓冲。

## 公式与图表

- `AddMath` 在 `pkg/math` 中解析 LaTeX，把 `m:oMathPara` / `m:oMath` 嵌进正文。不增加额外 ZIP 部件。
- `AddChart` 追加 `word/charts/chartN.xml` 及一条关系。系列 XML 遵循 Word 图表 XSD 顺序（`idx` → `order` → `tx` → `spPr` → …）。

## 完整示例 — 查看包内部件

`Bytes` 返回 ZIP。`bytes.NewReader` 实现了 `io.ReaderAt`，因此不必落盘即可列出部件。

```go
package main

import (
	"archive/zip"
	"bytes"
	"fmt"
	"log"

	"github.com/yunkeweb/go-word"
)

func main() {
	doc := word.New()
	sec := doc.AddSection()
	sec.AddTitle("Architecture", 1)
	sec.AddMath(`\frac{a}{b}`)
	raw, err := doc.Bytes()
	if err != nil {
		log.Fatal(err)
	}
	zr, err := zip.NewReader(bytes.NewReader(raw), int64(len(raw)))
	if err != nil {
		log.Fatal(err)
	}
	for _, f := range zr.File {
		fmt.Println(f.Name)
	}
}
```

列表中总会出现 `[Content_Types].xml`、`_rels/.rels`、`word/document.xml`、`word/styles.xml` 和 `word/_rels/document.xml.rels`。加入图表或页眉会再增加部件。

## 相关

[OpenXML 兼容性](./compatibility) 列出 Word 实际接受的 schema。[FAQ](./faq) 覆盖部件违规时的修复对话框。
