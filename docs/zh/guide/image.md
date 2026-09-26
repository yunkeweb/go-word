# 图片与 Shape

图片会变成 `word/media/imageN` 加上 DrawingML `a:blip r:embed`。关系 ID 在写出时分配，因此两份都添加 `logo.png` 的文档仍能干净合并。矢量形状（`wps:wsp`）在 [DrawingML 形状](./shapes) 有专页；本页覆盖图片以及一行式的 `Document.AddShape` 辅助方法。

## AddImage / AddImageBytes

### 签名

```go
func (c *Container) AddImage(source string, styles ...any) *Image
func (c *Container) AddImageBytes(name string, data []byte, styles ...any) *Image
```

### 参数

| 名称 | 类型 | 说明 |
| --- | --- | --- |
| `source` | `string` | 文件系统路径。扩展名决定内容类型（`png`、`jpeg`、`gif`、`emf`）。 |
| `name` | `string` | 写出器改写为 `imageN` 之前使用的部件文件名。 |
| `data` | `[]byte` | 已编码的图片字节（不是原始像素）。 |
| `styles` | `...any` | 可选 `style.Image`。 |

### style.Image

| 字段 | 单位 | 说明 |
| --- | --- | --- |
| `Width` / `Height` | CSS 像素 | 写出时转 EMU。`0` 保持固有尺寸。 |
| `WidthEMU` / `HeightEMU` | EMU | 1 英寸 = 914400 EMU。设置后覆盖像素尺寸。 |
| `AltText` | string | `wp:docPr descr`。 |
| `WrappingStyle` | string | `inline`（默认）、`square`、`tight`、`behind`、`infront`。 |
| `Alignment` | string | 行内图片所在段落的对齐。 |

### 注意

- 测试与服务端优先 `AddImageBytes`：不必在磁盘上放额外文件。
- PNG / JPEG / GIF 从字节检测。非法字节仍会创建部件，Word 随后显示红叉。
- 页眉接受同样的调用（`header.AddImageBytes(...)`），logo 可以放在 `word/header1.xml`。

## Document.AddShape（DrawingML 快捷方式）

### 签名

```go
func (d *Document) AddShape(shapeType ShapeType, opts ShapeOptions) *element.DMLShape
```

| `ShapeType` | 几何 |
| --- | --- |
| `ShapeRect` | `rect` |
| `ShapeRoundRect` | `roundRect` |
| `ShapeArrow` | `rightArrow` |
| `ShapeTextBox` | `rect` + `wps:txbx` / `w:txbxContent` |

`ShapeOptions`：`Width` / `Height`（EMU）、`FillColor`、`LineColor`、`LineWidth`、`Text`、`Font`。`ShapeTextBox` 在文本为空时写入空格，以便 Word 仍创建 `w:txbxContent`。

完整预设、填充与文本框示例见 [形状与文本框](./shapes)。

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
	sec := doc.AddSection()
	sec.AddTitle("Images and shapes", 1)
	sec.AddImageBytes("dot.png", tinyPNG(), style.Image{
		Width: 48, Height: 48, AltText: "red swatch",
	})
	sec.AddText("The PNG is stored as word/media/image1.png with a unique rId.")

	doc.AddShape(word.ShapeRoundRect, word.ShapeOptions{
		FillColor: "5B9BD5",
		LineColor: "2E75B6",
		Text:      "DrawingML",
		Font:      style.Font{Bold: true, Color: "FFFFFF"},
	})

	if err := doc.Save("image.docx"); err != nil {
		log.Fatal(err)
	}
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

Word 中先看到行内小红方块，然后是带白字的蓝色圆角矩形。之后可用 [StreamExtractImages](./streaming) 抽出图片。
