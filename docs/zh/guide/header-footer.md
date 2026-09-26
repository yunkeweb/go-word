# 页眉页脚与页码

页眉页脚是独立 ZIP 部件（`word/header1.xml`、`word/footer1.xml`），由 `w:sectPr` 通过 `r:id` 引用。类型与 PHPWord 一致：默认（奇数页）、首页、偶数页。

## AddHeader / AddFooter

### 签名

```go
func (s *Section) AddHeader(typ ...string) *Header
func (s *Section) AddFooter(typ ...string) *Footer
```

### 参数

| 名称 | 类型 | 说明 |
| --- | --- | --- |
| `typ` | `...string` | `element.HeaderAuto`（`"default"`）、`element.HeaderFirst`（`"first"`）、`element.HeaderEven`（`"even"`）。省略则为默认。 |

`Header` 与 `Footer` 嵌入 `Container`，因此 `AddText`、`AddImageBytes`、`AddPreserveText`、`AddPageNumber`、`AddNumPages` 均可使用。

## 页码域

### 签名

```go
func (c *Container) AddPreserveText(text string, styles ...any) *PreserveText
func (c *Container) AddPageNumber() *Field
func (c *Container) AddNumPages() *Field
func (d *Document) SetDifferentFirstPage(enable bool)
func (d *Document) SetEvenAndOddHeaders(enable bool)
```

| 调用 | OpenXML |
| --- | --- |
| `AddPreserveText("PAGE")` | 含 `PAGE` 的 `w:instrText` |
| `AddPageNumber()` | `w:fldChar` + `PAGE` 域 |
| `AddNumPages()` | `NUMPAGES` 域 |
| `SetDifferentFirstPage(true)` | `w:sectPr` 上的 `w:titlePg` |
| `SetEvenAndOddHeaders(true)` | settings 中的 `w:evenAndOddHeaders` |

### 注意

- 未调用 `SetDifferentFirstPage(true)` 时，首页页眉会被忽略。
- 未调用 `SetEvenAndOddHeaders(true)` 时，偶数页页眉会被忽略。
- 域显示缓存值，直到用户右键 → 更新域（或打印）。

## 完整示例

```go
package main

import (
	"log"

	"github.com/yunkeweb/go-word"
	"github.com/yunkeweb/go-word/element"
)

func main() {
	doc := word.New()
	doc.SetDefaultFontName("Calibri")
	doc.SetDifferentFirstPage(true)
	doc.SetEvenAndOddHeaders(true)

	sec := doc.AddSection()
	sec.AddHeader().AddText("GoWord — odd pages")
	sec.AddHeader(element.HeaderFirst).AddText("GoWord — cover")
	sec.AddHeader(element.HeaderEven).AddText("GoWord — even pages")

	f := sec.AddFooter()
	f.AddText("Page ")
	f.AddPreserveText("PAGE")
	f.AddText(" of ")
	f.AddNumPages()
	sec.AddFooter(element.HeaderFirst).AddText("Cover footer")

	sec.AddTitle("Headers and page numbers", 1)
	sec.AddText("This is the first page (cover header).")
	sec.AddPageBreak()
	sec.AddText("Second page uses the odd header.")
	sec.AddPageBreak()
	sec.AddText("Third page uses the even header.")

	if err := doc.Save("headers.docx"); err != nil {
		log.Fatal(err)
	}
}
```

Word 在三页上显示三种页眉，默认页脚含 `PAGE` / `NUMPAGES`。打印版式或打印预览会刷新域。
