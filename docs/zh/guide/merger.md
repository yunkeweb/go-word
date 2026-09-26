# 多文档无损合并

`AppendDocument` 将 `src` 的每一个节克隆到 `dst`，并重映射在同一个 OpenXML 包内会冲突的标识符。结果是一份 ZIP，其中的样式、书签与媒体不会互相覆盖。

## AppendDocument

### 签名

```go
func (d *Document) AppendDocument(src *Document, opts MergeOptions) error

type MergeOptions struct {
	StylePrefix    string
	BookmarkPrefix string
	SectionBreak   string
}
```

### 参数

| 名称 | 类型 | 说明 |
| --- | --- | --- |
| `src` | `*Document` | 源树。`nil` 返回 `word: nil source document`。 |
| `StylePrefix` | `string` | 冲突段落/表格样式 ID 的前缀。默认 `src_`。 |
| `BookmarkPrefix` | `string` | 冲突 `w:bookmarkStart` 名的前缀。默认 `src_`。 |
| `SectionBreak` | `string` | 克隆过来的第一节的分节符。默认 `nextPage`。 |

### 重映射范围

| 标识符 | 行为 |
| --- | --- |
| 段落 / 表格样式 ID | 冲突名称加上前缀（`Note` → `src_Note`） |
| 书签名 | 冲突名称加前缀；唯一名称保持不变 |
| 内部超链接锚点 | 更新为新的书签名 |
| 图片 / 媒体 | 清空 `RelationID`；写出时分配新的 `rId` 与 `word/media/imageN` |

`rId` 在写出时分配（`word2007Writer.nextRel`），合并后的 ZIP 中两份文档不会共用关系 ID。

## 完整示例

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

### Word 中的效果

两页。第 1 页是藏青色色块 + Document A。第 2 页在下一页分节符之后，显示橙色色块，并含可编辑的 `1/2` 分数。styles.xml 同时有 `Note` 与 `src_Note`（段后间距不同）。书签 `shared` 与 `src_shared` 都在。ZIP 中有 `word/media/image1.png` 与 `image2.png`，各有独立 `rId` —— 解压时图片不会互相覆盖。
