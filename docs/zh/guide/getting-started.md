# 快速开始

本页生成一份 `.docx`，覆盖多数团队第一天就会用到的四项能力：原生 Office Math 公式、DrawingML 形状、双栏节，以及 `Save`。

公开 API 沿用 PHPWord 命名（`AddSection`、`AddText`、`IOFactory`），并使用惯用的 Go 类型与 `error` 返回。

## New / AddSection / Save

### 签名

```go
func New() *Document
func (d *Document) AddSection(style ...any) *element.Section
func (d *Document) Save(filename string) error
```

### 参数

| 名称 | 类型 | 说明 |
| --- | --- | --- |
| `style` | `...any` | 可选 `style.Section`（纸张、边距、方向、分节符）。 |
| `filename` | `string` | 目标路径。父目录必须已存在。 |

### 注意

- `New` 创建空文档，默认 run 字体为 Calibri 11 pt。
- `AddSection` 追加一块 `w:sectPr`。正文元素挂在返回的 `*element.Section` 上。
- `Save` 等价于 `CreateWriter(doc, "Word2007")` 再写文件。关系 ID（`rIdN`）与媒体名（`word/media/imageN`）在写出时分配。

## 完整示例

保存为 `main.go` 后执行 `go run .`。Microsoft Word 打开 `hello.docx` 时不会弹出修复对话框。

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
	doc.SetDefaultFontSize(11)

	info := doc.GetDocInfo()
	info.Title = "GoWord v0.9.0"
	info.Creator = "GoWord"

	sec := doc.AddSection()
	sec.AddTitle("GoWord v0.9.0", 1)
	sec.AddText("Native Office Math:")
	sec.AddMath(`\frac{a}{b}`)

	p := sec.AddTextRun()
	p.AddText("Pythagoras: ")
	p.AddMath(`x^{2} + y^{2} = z^{2}`)

	doc.AddShape(word.ShapeRoundRect, word.ShapeOptions{
		FillColor: "5B9BD5",
		LineColor: "2E75B6",
		Text:      "DrawingML",
		Font:      style.Font{Bold: true, Color: "FFFFFF"},
	})

	cols := doc.AddSection()
	cols.SetColumns(2, 720, true)
	cols.AddText("The left column starts here. Word flows this section into two equal columns.")
	cols.AddText("A separator line is emitted as w:cols w:sep.")

	if err := doc.Save("hello.docx"); err != nil {
		log.Fatal(err)
	}
}
```

### Word 中的效果

| 功能 | 屏幕上 | OpenXML |
| --- | --- | --- |
| 展示公式 | 可双击的居中公式 | `w:p` / `m:oMathPara` / `m:oMath` / `m:f` |
| 行内公式 | 勾股定理与标签同一行 | 与 `w:r` 并列的 `m:oMath` |
| 圆角矩形 | 蓝色胶囊，白色 “DrawingML” | `wps:wsp` / `a:prstGeom prst="roundRect"` |
| 两栏 | 等宽栏 + 竖向分隔线 | `w:cols w:num="2" w:space="720" w:sep="1"` |

## IOFactory 别名

### 签名

```go
func CreateWriter(doc *Document, name string) (Writer, error)
func CreateReader(name string) (Reader, error)
func Load(filename string, readerName ...string) (*Document, error)
func Open(filePath string) (*Document, error)
```

`name` 为 `"Word2007"`（默认）。`Load` / `Open` 构建完整 DOM；O(1) 抽文本见 [流式解析器](./streaming)。

## 下一步

| 主题 | 页面 |
| --- | --- |
| ZIP 如何组装 | [架构设计](./architecture) |
| 段落、表格、图片 | [段落与 Run](./paragraph) |
| SDT 表单控件 | [结构化文档标签](./sdt) |
| 跨页表头 / cantSplit | [表格](./table#setheader-setcantsplit-setvalign-settextdirection) |
| 平铺水印与 AllowEdit | [水印与保护](./protect) |
| LaTeX → Word 公式 | [Office Math](./math) |
| 图表 | [DrawingML 图表](./charts) |
