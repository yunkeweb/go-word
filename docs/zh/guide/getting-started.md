# 快速开始与安装

GoWord 写出原生 **OpenXML Word 2007（`.docx`）** 包。公开 API 沿用 PHPWord 命名（`AddSection`、`AddText`、`IOFactory`、`TemplateProcessor`），并使用惯用的 Go 类型与 `error` 返回。

需要 **Go 1.21+**。协议：[GNU LGPL v3](https://github.com/yunkeweb/go-word/blob/main/LICENSE)。

## 安装

```sh
go get github.com/yunkeweb/go-word@v0.8.0
```

`go.mod` 没有任何第三方 `require`。序列化走 `encoding/xml`，打包走 `archive/zip`。

## 第一份文档

保存为 `main.go` 后执行 `go run .`。生成的 `hello.docx` 可在 Microsoft Word 中直接打开，不会弹出修复对话框。

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
	info.Title = "GoWord v0.8.0"
	info.Creator = "GoWord"

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

本程序写出的 OpenXML：

| 能力 | 节点 |
| --- | --- |
| 展示公式 | `w:p` / `m:oMathPara` / `m:oMath` / `m:f` |
| 行内公式 | 与 `w:r` 并列的 `m:oMath` |
| 圆角矩形 | `wps:wsp` / `a:prstGeom prst="roundRect"` |
| 两栏 | `w:cols w:num="2" w:space="720" w:sep="1"` |

`Save` 将 `word/document.xml` 流式写入 ZIP。`CreateWriter(doc, "Word2007")` 是 PHPWord 兼容入口。

## 下一步

| 主题 | 页面 |
| --- | --- |
| 段落、嵌套表、图片、页眉 | [核心 DOM](./basics) |
| LaTeX → Word 公式 | [Office Math](./math) |
| 图表与形状 | [DrawingML](./drawing) |
| 学术报告、拼接、财务报表 | [实战案例库](./examples) |
