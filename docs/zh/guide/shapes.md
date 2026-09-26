# DrawingML 矢量形状与文本框

GoWord 写出 Word 2010 **wordprocessingShape** 图形（`wps:wsp`），几何为 DrawingML 预设（`a:prstGeom`）。内部文字使用 `wps:txbx` / `w:txbxContent` —— 与“插入 → 文本框”相同。

## AddShape

### 签名

```go
func (d *Document) AddShape(shapeType ShapeType, opts ShapeOptions) *element.DMLShape
```

### 参数

| 名称 | 类型 | 说明 |
| --- | --- | --- |
| `shapeType` | `ShapeType` | `ShapeRect`、`ShapeRoundRect`、`ShapeArrow`、`ShapeTextBox`。 |
| `opts.Width` / `Height` | `int` | EMU。1 英寸 = 914400。零则使用写出器默认值。 |
| `opts.FillColor` | `string` | 不含 `#` 的十六进制（`srgbClr`）。 |
| `opts.LineColor` | `string` | 描边十六进制。 |
| `opts.LineWidth` | `int` | 描边宽度，单位 EMU。 |
| `opts.Text` | `string` | `w:txbxContent` 正文。 |
| `opts.Font` | `style.Font` | 内部 run 的粗体、字号、颜色。 |

### 几何

| 常量 | `a:prstGeom prst` | 额外 |
| --- | --- | --- |
| `ShapeRect` | `rect` | |
| `ShapeRoundRect` | `roundRect` | |
| `ShapeArrow` | `rightArrow` | |
| `ShapeTextBox` | `rect` | 总是写出 `wps:txbx`。空文本变成空格。 |

底层接口：`Section.AddDMLShape(prst, width, height, fill, line, lineWidth)`。

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
	sec := doc.AddSection()
	sec.AddTitle("DrawingML shapes", 1)

	doc.AddShape(word.ShapeRect, word.ShapeOptions{
		FillColor: "5B9BD5", LineColor: "2E75B6",
		Width: 1828800, Height: 914400,
	})
	doc.AddShape(word.ShapeRoundRect, word.ShapeOptions{
		FillColor: "ED7D31", LineColor: "C45911",
		Text: "Rounded", Font: style.Font{Bold: true, Color: "FFFFFF"},
	})
	doc.AddShape(word.ShapeArrow, word.ShapeOptions{
		FillColor: "70AD47", LineColor: "548235",
	})
	doc.AddShape(word.ShapeTextBox, word.ShapeOptions{
		FillColor: "FFF2CC", LineColor: "BF8F00",
		Text: "w:txbxContent", Font: style.Font{Size: 12, Color: "595959"},
	})

	if err := doc.Save("shapes.docx"); err != nil {
		log.Fatal(err)
	}
}
```

### Word 中的效果

四个行内图形：蓝色矩形（2 英寸 × 1 英寸）、带白色 **Rounded** 的橙色胶囊、绿色右箭头，以及可点击输入的米色文本框。选中形状会显示 Word 的绘图工具；这是 `wps:wsp`，不是 VML `w:pict`（水印仍是 VML）。
