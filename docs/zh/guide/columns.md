# 多栏排版 (w:cols)

`SetColumns` 在**当前节**上写出符合 schema 的 `w:cols`。Word 把该节段落灌进等宽栏。其余内容若要回到单栏，请再开一节。

## SetColumns

### 签名

```go
func (s *Section) SetColumns(num int, space int, showLine bool)
```

### 参数

| 名称 | 类型 | 说明 |
| --- | --- | --- |
| `num` | `int` | 栏数（`w:num`）。 |
| `space` | `int` | 栏间距，单位 twip（`w:space`）。`720` 为 0.5 英寸。 |
| `showLine` | `bool` | 分隔线（`w:sep="1"`）。属性名是 `sep`，不是 `separator`。 |

### 注意

- 栏设置写在 `w:sectPr`。单栏与双栏混排需要 `AddSection` 并设置 `BreakType: "nextPage"`（或 `"continuous"`）。
- 横向加分栏：同一节传入 `style.OrientationLandscape`。
- 默认纸张为 A4（`style.DefaultPageWidth` = 11906 twip，`DefaultPageHeight` = 16838）。

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
	doc.SetDefaultAsianFontName("Microsoft YaHei")

	intro := doc.AddSection()
	intro.AddTitle("Single column", 1)
	intro.AddText("This section is a normal single-column page.")

	cols := doc.AddSection(style.Section{
		Orientation: style.OrientationPortrait,
		BreakType:   "nextPage",
		PageSizeW:   style.DefaultPageWidth,
		PageSizeH:   style.DefaultPageHeight,
		MarginTop:   1134, MarginBottom: 1134, MarginLeft: 1134, MarginRight: 1134,
	})
	cols.SetColumns(2, 720, true)
	cols.AddTitle("Two-column layout", 1)
	cols.AddText("The left column starts here. Word splits this section into two equal columns with a separator line (w:cols w:sep).")
	cols.AddText("The second paragraph continues the flow so the columns fill naturally. Formulas and pictures follow the same column stream.")
	cols.AddMath(`\frac{a}{b}`)

	if err := doc.Save("columns.docx"); err != nil {
		log.Fatal(err)
	}
}
```

### Word 中的效果

第 1 页单栏。第 2 页两栏等宽，中间有竖线。分数跟着栏流走，不会拉满整页宽度。打印版式下分栏最明显，草稿视图看不出来。
