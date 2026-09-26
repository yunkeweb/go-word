# 快速开始

GoWord 写出原生 **OpenXML Word 2007（`.docx`）** 包。公开 API 沿用 PHPWord 命名（`AddSection`、`AddText`、`IOFactory`、`TemplateProcessor`），并使用惯用的 Go 类型与 `error` 返回。

需要 **Go 1.21+**。协议：[GNU LGPL v3](https://github.com/yunkeweb/go-word/blob/main/LICENSE)。

## 安装

```sh
go get github.com/yunkeweb/go-word@v0.8.0
```

## 第一份文档

创建文档、插入原生 OMML 公式、绘制 DrawingML 形状，并启用两栏排版：

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
	sec.AddTitle("GoWord v0.8.0", 1)
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

## 下一步

| 主题 | 页面 |
| --- | --- |
| 段落、表格、图片、页眉页脚 | [基础 DOM 操作](./basics) |
| LaTeX → Word 公式 | [Office Math](./math) |
| 形状与图表 | [DrawingML 图表与形状](./drawing) |
| \(O(1)\) ZIP 提取 | [流式提取器](./streaming) |
| `${block}` / `${if}` / 管道 | [模板引擎 v2](./template) |
| 拼接文档 | [文档无损合并](./merger) |
| 分栏、水印、保护、目录 | [高级排版与保护](./layout) |

可运行示例见 [`examples/`](https://github.com/yunkeweb/go-word/tree/main/examples)。完整 API：[pkg.go.dev/github.com/yunkeweb/go-word](https://pkg.go.dev/github.com/yunkeweb/go-word)。
