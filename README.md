# GoWord

[![Go Reference](https://pkg.go.dev/badge/github.com/yunkeweb/go-word.svg)](https://pkg.go.dev/github.com/yunkeweb/go-word)
[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)](https://go.dev/)
[![License: LGPL v3](https://img.shields.io/badge/License-LGPL%20v3-blue.svg)](LICENSE)

纯 Go 实现的 Word 文档库，对应 [PHPWord](https://github.com/PHPOffice/PHPWord) 的文档模型与 API：用代码创建、读取、填充 Microsoft Word `.docx` 文件。

依赖仅为 Go 标准库（`archive/zip`、`encoding/xml`、`image` 等），`go.mod` 无外部 `require`。公开方法沿用 PHPWord 命名（`AddSection`、`AddText`、`IOFactory`），同时采用惯用的 Go 类型与 `error` 返回。

## 功能特性

- **读写 Word 2007 包**：生成与加载 `.docx` / `.docm`（OOXML / WordprocessingML）
- **流式输出**：`Save` / `SaveAs`、`WriteTo(io.Writer)`、`Bytes()`，以及从内存加载 `LoadBytes`
- **模板填充**：`${placeholder}` 替换、表格行克隆、块克隆、图片/图表/复选框、复杂元素插入
- **文档结构**：多节、页眉/页脚（默认 / 首页 / 偶数页）、水印、分页、分节
- **文本与样式**：段落、富文本 `TextRun`、标题 Heading 1–9、命名字符/段落/表格/编号样式
- **表格**：行高、列宽、合并（`GridSpan` / `VMerge`）、边框、底纹、对齐
- **媒体与图形**：本地或内存图片、超链接、书签、VML 线条/形状、文本框、OLE 对象
- **列表与域**：项目符号 / 多级编号、`PAGE`/`DATE` 等域、页眉 PreserveText、目录 TOC
- **批注与修订**：脚注、尾注、批注、修订标记
- **表单**：复选框、文本输入、下拉框、结构化文档标记（SDT）
- **图表**：饼图、柱状图等简单图表系列
- **公式**：Office Math（OMML）分数、上下标、标识符与运算符
- **HTML 导入**：将 HTML 片段转为段落、标题、列表、表格、链接与图片
- **文档属性**：核心/扩展属性、自定义属性、纸张尺寸、页边距、保护与校对设置
- **PHPWord 对照**：包路径与方法名可直接对照 PHP 侧实现

## 要求

- Go 1.21 或更高版本
- 生成与读取格式：Word 2007（`.docx`）

## 安装

```sh
go get github.com/yunkeweb/go-word
```

## 快速开始

```go
package main

import (
	"log"

	"github.com/yunkeweb/go-word"
	"github.com/yunkeweb/go-word/style"
)

func main() {
	doc := word.New()

	section := doc.AddSection()
	section.AddText(`"Learn from yesterday, live for today, hope for tomorrow."`)
	section.AddText(
		`"Great achievement is usually born of great sacrifice."`,
		style.Font{Name: "Tahoma", Size: 10},
	)

	if err := doc.Save("helloWorld.docx"); err != nil {
		log.Fatal(err)
	}
}
```

完整可运行示例见 [`examples/simple`](examples/simple)。

## 示例

### 样式、标题与行内格式

```go
doc := word.New()
doc.SetDefaultFontName("Calibri")
doc.SetDefaultFontSize(11)

doc.AddFontStyle("rStyle", style.Font{
	Bold: true, Italic: true, Size: 16, AllCaps: true,
})
doc.AddParagraphStyle("pStyle", style.Paragraph{
	Alignment: style.JcCenter,
	Spacing:   style.Spacing{After: 100},
})
doc.AddTitleStyle(1, style.Font{Bold: true, Size: 16}, style.Paragraph{
	Spacing: style.Spacing{After: 240},
})

section := doc.AddSection()
section.AddTitle("Welcome to GoWord", 1)
section.AddText("I am styled by a font style definition.", "rStyle")
section.AddText("I am styled by a paragraph style definition.", nil, "pStyle")

tr := section.AddTextRun()
tr.AddText("I am inline styled ", style.Font{Name: "Times New Roman", Size: 20})
tr.AddText("with ")
tr.AddText("color", style.Font{Color: "996699"})
tr.AddText(", ")
tr.AddText("bold", style.Font{Bold: true})
tr.AddText(", ")
tr.AddText("italic", style.Font{Italic: true})
tr.AddText(".")
```

### 表格、列表、链接与页眉页脚

```go
section := doc.AddSection()

tbl := section.AddTable(style.Table{
	Width: 5000,
	Borders: style.Borders{
		Top:    style.Border{Style: style.BorderSingle, Size: 4},
		Left:   style.Border{Style: style.BorderSingle, Size: 4},
		Right:  style.Border{Style: style.BorderSingle, Size: 4},
		Bottom: style.Border{Style: style.BorderSingle, Size: 4},
	},
})
row := tbl.AddRow(300)
row.AddCell(2500).AddText("A")
row.AddCell(2500).AddText("B")

section.AddListItem("one", 0, nil, nil, style.ListTypeBullet)
section.AddListItem("two", 0)
section.AddLink("https://github.com/yunkeweb/go-word", "GoWord on GitHub")

header := section.AddHeader()
header.AddText("Confidential")
footer := section.AddFooter()
footer.AddPreserveText("Page {PAGE}")
```

### 图片、水印与图表

```go
section.AddImage("logo.png", style.Image{Width: 120, Height: 40})
section.AddImageBytes("chart.png", pngBytes, style.Image{Width: 240})

first := section.AddHeader("first")
first.AddWatermark("mark.png")

section.AddChart("pie", []string{"A", "B", "C"}, []float64{3, 1, 2})
```

### HTML 导入

```go
html := `<p>Hello <strong>World</strong></p><h1>Title</h1><ul><li>one</li></ul>`
if err := word.AddHTML(section, html); err != nil {
	log.Fatal(err)
}
```

支持的标签包括 `p`、`h1`–`h6`、`br`、`hr`、`table`/`tr`/`td`、`ul`/`ol`/`li`、`a`、`img`、`strong`/`b`、`em`/`i`、`u`、`s`、`sup`/`sub` 以及常见内联 CSS（颜色、字号、对齐等）。

### 公式

```go
import "github.com/yunkeweb/go-word/pkg/math"

m := math.New()
m.Add(math.NewIdentifier("a"))
m.Add(math.NewOperator("="))
m.Add(math.NewFraction(math.NewNumeric("1"), math.NewNumeric("2")))
section.AddFormula(m)
```

### 文档属性与纸张

```go
info := doc.GetDocInfo()
info.SetCreator("Alice")
info.SetTitle("Quarterly Report")
info.SetCompany("Yunke")

section := doc.AddSection(style.Section{
	Orientation: style.OrientationPortrait,
	PaperSize:   "A4",
	MarginTop:   1440,
	MarginLeft:  1440,
})
```

纸张预设：`A3`、`A4`、`A5`、`B5`、`Folio`、`Legal`、`Letter`。度量单位可用 `word.SetMeasurementUnit(word.UnitTwip)`（亦支持 `cm`、`mm`、`inch`、`point`、`pica`）。

### 读取文档

```go
doc, err := word.Load("file.docx")
if err != nil {
	log.Fatal(err)
}

section := doc.GetSection(0)
for _, el := range section.Elements() {
	_ = el.Type()
}

names, err := word.ExtractVariables("template.docx")
```

也可通过 IOFactory：

```go
reader, err := word.CreateReader("Word2007")
doc, err = reader.Load("file.docx")

writer, err := word.CreateWriter(doc, "Word2007")
err = writer.Save("out.docx")
```

从内存读写：

```go
data, err := doc.Bytes()
doc, err = word.LoadBytes(data)
```

### 模板填充

Word 模板中使用 `${name}` 占位符（可用 `SetMacroChars` 自定义分隔符）：

```go
tp, err := word.NewTemplateProcessor("template.docx")
if err != nil {
	log.Fatal(err)
}

tp.SetValue("name", "World")
tp.SetValues(map[string]string{
	"title":   "Invoice",
	"company": "Yunke",
})

err = tp.CloneRowAndSetValues("item", []map[string]string{
	{"item": "Apple", "qty": "3"},
	{"item": "Pear", "qty": "1"},
})
err = tp.SetImageValue("logo", "logo.png")
err = tp.CloneBlock("blk", 2)

if err := tp.Save("out.docx"); err != nil {
	log.Fatal(err)
}
```

常用模板 API：

| 方法 | 作用 |
|------|------|
| `SetValue` / `SetValues` | 替换占位符 |
| `CloneRow` / `CloneRowAndSetValues` | 按 `${item}` 所在行克隆并编号为 `${item#1}` |
| `CloneBlock` / `CloneBlockAndSetValues` | 克隆 `${block}` … `${/block}` |
| `DeleteRow` / `DeleteBlock` | 删除行或块 |
| `SetImageValue` / `SetChart` | 插入图片或图表 |
| `SetComplexValue` / `SetComplexBlock` | 用元素替换 run / 段落 |
| `SetCheckbox` | 切换内容控件复选框 |
| `GetVariables` / `GetVariableCount` | 枚举占位符 |

### 与 PHPWord 对照

```php
$phpWord = new \PhpOffice\PhpWord\PhpWord();
$section = $phpWord->addSection();
$section->addText('Hello World');
$writer = \PhpOffice\PhpWord\IOFactory::createWriter($phpWord, 'Word2007');
$writer->save('helloWorld.docx');
```

```go
doc := word.New()
section := doc.AddSection()
section.AddText("Hello World")
if err := doc.Save("helloWorld.docx"); err != nil {
	log.Fatal(err)
}
```

## 包结构

| 路径 | 职责 | 对应 PHP |
|------|------|----------|
| `github.com/yunkeweb/go-word` | 文档、读写器、模板、HTML | `PhpWord`、`IOFactory`、`TemplateProcessor` |
| `.../style` | 字体、段落、表格、节、编号、纸张 | `PhpWord\Style\*` |
| `.../element` | 节、文本、表格、图片、图表等元素 | `PhpWord\Element\*` |
| `.../metadata` | 文档属性、设置、兼容性 | `PhpWord\Metadata\*` |
| `.../pkg/common` | ZIP、XML、单位换算、字体 | `phpoffice/common` |
| `.../pkg/math` | Office Math（OMML） | `phpoffice/math` |
| `.../ooxml` | WordprocessingML 部件与命名空间 | OOXML 部件 |

## 常用命令

| 命令 | 说明 |
|------|------|
| `go get github.com/yunkeweb/go-word` | 安装模块 |
| `go test ./...` | 运行测试 |
| `go run ./examples/simple` | 运行入门示例，生成 `helloWorld.docx` |
| `go doc github.com/yunkeweb/go-word` | 查看包文档 |

## 许可

GNU Lesser General Public License version 3，与 PHPWord 同族。详见 [LICENSE](LICENSE)。
