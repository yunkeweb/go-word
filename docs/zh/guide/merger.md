# 文档无损合并

`AppendDocument` 将 `src` 的每一个节克隆到 `dst`，并重映射在同一个 OpenXML 包内会冲突的标识符。

```go
err := dst.AppendDocument(src, word.MergeOptions{
	StylePrefix:    "src_",
	BookmarkPrefix: "src_",
	SectionBreak:   "nextPage",
})
```

## 重映射范围

| 标识符 | 行为 |
| --- | --- |
| 段落 / 表格 **样式 ID** | 冲突名称加上 `StylePrefix`（`Note` → `src_Note`） |
| **书签** 名（`w:bookmarkStart`） | 冲突名称加前缀；唯一名称保持不变 |
| 内部超链接 **锚点** | 更新为新的书签名 |
| **图片 / 媒体** | 清空 `RelationID`；写出时分配新的 `rId` 与 `word/media/imageN` |
| 分节符 | 克隆过来的第一节使用 `SectionBreak`（默认 `nextPage`） |

`rId` 在写出时分配（`word2007Writer.nextRel`），合并后的 ZIP 中两份文档不会共用关系 ID。`src` 为 `nil` 时返回 `word: nil source document`。

## 完整示例

保存为 `main.go` 后执行 `go run .`。`merged.docx` 同时保留两套 `Note` 样式、`shared` / `src_shared` 书签，以及两个 PNG 部件。

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

合并之后：

| 资源 | 隔离结果 |
| --- | --- |
| A 的样式 `Note` | 保留 |
| B 的样式 `Note` | 改写为 `src_Note` |
| A 的书签 `shared` | 保留 |
| B 的书签 `shared` | 改写为 `src_shared` |
| 图片 | `word/media/image1.png` 与 `image2.png`，各有独立 `rId` |

示例：[`examples/v0.8.0_demo`](https://github.com/yunkeweb/go-word/tree/main/examples/v0.8.0_demo) 以及 [实战案例库](./examples) 案例 B。
