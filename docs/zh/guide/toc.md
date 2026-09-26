# TOC 自动目录

`AddTOC` / `AddTableOfContents` 写出 Word `TOC` 域。`AddTitle` 生成的标题段落带 `w:outlineLvl`，用户更新域时 Word 即可收集。

## AddTOC / AddTableOfContents

### 签名

```go
func (c *Container) AddTOC(font any, tocStyle any, minDepth, maxDepth int) *TOC
func (d *Document) AddTableOfContents(depth ...int) *element.TOC
```

### 参数

| 名称 | 类型 | 说明 |
| --- | --- | --- |
| `font` | `any` | 域段落的可选 `style.Font`。`nil` 使用默认。 |
| `tocStyle` | `any` | 可选段落样式。`nil` 即可。 |
| `minDepth` / `maxDepth` | `int` | 大纲范围。`0` 会变成 1 和 9。 |
| `depth` | `...int` | `AddTableOfContents(3)` 收集 1–3 级。 |

指令文本为 `TOC \o "1-3" \h \z \u`（范围随 min/max 变化）。

### 注意

- 首次打开时 Word 显示“更新域”。右键目录 → 更新域 → 更新整个目录。
- [TemplateProcessor](./template) 上的 `SetUpdateFields(true)` 会让 Word 打开时刷新域（`w:updateFields`）。
- 标题必须用 `AddTitle`，不能只用带样式的 `AddText`，否则没有大纲级别。

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
	doc.AddTitleStyle(1, style.Font{Bold: true, Size: 16})
	doc.AddTitleStyle(2, style.Font{Bold: true, Size: 13})

	sec := doc.AddSection()
	sec.AddTitle("Table of Contents", 1)
	sec.AddTOC(nil, nil, 1, 2)
	sec.AddText("Right-click the TOC field in Microsoft Word and choose Update Field.")

	sec.AddTitle("Governing equations", 1)
	sec.AddText("Maximum bending stress.")
	sec.AddTitle("Methods", 2)
	sec.AddText("A concentrated force at mid-span.")
	sec.AddTitle("Results", 1)
	sec.AddText("Peak stress stays below yield.")

	if err := doc.Save("toc.docx"); err != nil {
		log.Fatal(err)
	}
}
```

### Word 中的效果

更新域之后，目录列出 “Governing equations”、“Methods”、“Results”，带页码与超链接（`\h`）。级别跟随 `AddTitle` 的深度。
