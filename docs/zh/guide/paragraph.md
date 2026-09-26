# 段落与 Run

段落是 `w:p`。Run 是带 `w:t` 的 `w:r`。GoWord 沿用 PHPWord 命名：`AddText`（一段落一 run）、`AddTextRun`（一段落多 run）、`AddTitle`、`AddLink`、`AddBookmark`、`AddListItem`。

## AddText

### 签名

```go
func (c *Container) AddText(text string, styles ...any) *Text
```

### 参数

| 名称 | 类型 | 说明 |
| --- | --- | --- |
| `text` | `string` | 段落内容。写出时转义 XML 特殊字符。 |
| `styles` | `...any` | 可选 `style.Font`，然后是段落样式（`style.Paragraph` 或已注册样式名）。 |

### 注意

- 可用于 `Section`、`Header`、`Footer`、`Cell`。
- `style.Font` 字段：`Name`、`Size`（磅）、`Bold`、`Italic`、`Color`（不含 `#` 的十六进制）、`Underline`、`Highlight`。

## AddTextRun

### 签名

```go
func (c *Container) AddTextRun(styles ...any) *TextRun
func (p *TextRun) AddText(text string, styles ...any) *Text
func (p *TextRun) AddMath(formula string) *Formula
```

粗体与斜体（或文字与公式）必须共处同一 `w:p` 时，使用 `AddTextRun`。

## AddTitle / AddLink / AddBookmark / AddListItem

### 签名

```go
func (c *Container) AddTitle(text string, depth int, page ...int) *Title
func (c *Container) AddLink(target, text string, styles ...any) *Link
func (c *Container) AddBookmark(name string) *Bookmark
func (c *Container) AddListItem(text string, depth int, styles ...any) *ListItem
func (c *Container) AddPageBreak() *PageBreak
func (c *Container) AddTextBreak(count ...int)
```

| 方法 | OpenXML |
| --- | --- |
| `AddTitle` | 带大纲级别 `w:outlineLvl` 的段落。深度 1–9。 |
| `AddLink` | `w:hyperlink r:id`（外部）或 `w:anchor`（内部，第三个 style 传 `true`）。 |
| `AddBookmark` | `w:bookmarkStart` / `w:bookmarkEnd` |
| `AddListItem` | `w:pPr` 上的编号实例 |
| `AddPageBreak` | `w:br w:type="page"` |

在文档上注册可复用样式：

```go
doc.AddFontStyle("strong", style.Font{Bold: true, Size: 14})
doc.AddParagraphStyle("center", style.Paragraph{Alignment: style.JcCenter})
doc.AddTitleStyle(1, style.Font{Bold: true, Size: 18}, style.Paragraph{
	Spacing: style.Spacing{After: 240},
})
```

## 完整示例

```go
package main

import (
	"log"

	"github.com/yunkeweb/go-word"
	"github.com/yunkeweb/go-word/style"
)

func main() {
	doc := word.New()
	doc.SetDefaultFontName("Calibri")
	doc.SetDefaultFontSize(11)
	doc.AddFontStyle("strong", style.Font{Bold: true, Size: 14, Name: "Calibri"})
	doc.AddParagraphStyle("center", style.Paragraph{Alignment: style.JcCenter})
	doc.AddTitleStyle(1, style.Font{Bold: true, Size: 18}, style.Paragraph{
		Spacing: style.Spacing{After: 240},
	})

	sec := doc.AddSection()
	sec.AddTitle("Paragraphs and runs", 1)
	sec.AddText("Centered heading style.", "strong", "center")

	run := sec.AddTextRun()
	run.AddText("Bold ", style.Font{Bold: true})
	run.AddText("and italic.", style.Font{Italic: true, Color: "C00000"})

	sec.AddBookmark("intro")
	sec.AddLink("https://github.com/yunkeweb/go-word", "GoWord on GitHub")
	sec.AddListItem("First item", 0)
	sec.AddListItem("Nested item", 1)
	sec.AddPageBreak()
	sec.AddText("After the page break.")

	if err := doc.Save("paragraphs.docx"); err != nil {
		log.Fatal(err)
	}
}
```

Word 中会看到 Heading 1、居中粗体行、混排段落、蓝色超链接、两级列表，然后是第二页。`AddTitle` 的大纲级别供 [TOC](./toc) 收集。
