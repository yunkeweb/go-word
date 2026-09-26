# 高级排版与安全

节属性、水印、文档保护与目录域。

## 多栏排版

`SetColumns` 写出符合 schema 的 `w:cols` 节点（`num`、`space`、`sep`）：

| 参数 | 含义 |
| --- | --- |
| `num` | 栏数（`w:num`） |
| `space` | 栏间距，单位 twip（`w:space`，默认 720） |
| `showLine` | 分隔线（`w:sep="1"`） |

栏设置作用于**当前节**，其余内容若要回到单栏，请再开一节。横纵向混排同理：`AddSection` 时传入 `style.OrientationLandscape`。

## 水印

文字水印是每个节页眉中的 Word 原生 VML（`PowerPlusWaterMarkObject`）。图片水印把媒体部件挂到同一段 VML 上。

## 只读保护

模式：`ProtectTypeReadOnly`、`ProtectTypeComments`、`ProtectTypeTrackedChanges`、`ProtectTypeForms`。非空密码使用 Office SHA-1 / 100000 次迭代哈希（ECMA-376 `w:documentProtection`）。

## 目录

`AddTOC` / `AddTableOfContents` 写出 `TOC` 域（`TOC \o "1-3" \h \z \u`）。`AddTitle` 生成的标题段落带大纲级别，Word 刷新域时即可收集。

## 完整示例

保存为 `main.go` 后执行 `go run .`。`layout.docx` 为只读（密码 `goword`），带机密水印、目录、双栏节与横向节。

```go
package main

import (
	"log"

	"github.com/yunkeweb/go-word"
	"github.com/yunkeweb/go-word/element"
	"github.com/yunkeweb/go-word/style"
)

func main() {
	doc := word.New()
	info := doc.GetDocInfo()
	info.Title = "Advanced layout and protection"
	info.Creator = "GoWord"

	doc.SetDefaultFontName("Calibri")
	doc.SetDefaultAsianFontName("Microsoft YaHei")
	doc.SetDefaultFontSize(11)
	doc.AddTitleStyle(1, style.Font{Bold: true, Size: 16})
	doc.AddTitleStyle(2, style.Font{Bold: true, Size: 13})

	if err := doc.Protect(word.ProtectTypeReadOnly, "goword"); err != nil {
		log.Fatal(err)
	}
	doc.SetTextWatermark("CONFIDENTIAL")

	portrait := doc.AddSection()
	portrait.AddHeader().AddText("GoWord — odd pages")
	portrait.AddHeader(element.HeaderFirst).AddText("GoWord — first page")
	portrait.AddHeader(element.HeaderEven).AddText("GoWord — even pages")
	oddFooter := portrait.AddFooter()
	oddFooter.AddText("Page ")
	oddFooter.AddPreserveText("PAGE")
	portrait.AddFooter(element.HeaderFirst).AddText("Cover footer")
	doc.SetDifferentFirstPage(true)

	portrait.AddTitle("Table of Contents", 1)
	portrait.AddTOC(nil, nil, 1, 3)
	portrait.AddText("Right-click the TOC field in Microsoft Word and choose Update Field.")

	portrait.AddTitle("Portrait section", 1)
	portrait.AddText("This section is A4 portrait. w:titlePg selects the first-page header.")
	portrait.AddTitle("Document protection", 2)
	portrait.AddText("w:documentProtection edit=readOnly. Sample password: goword.")
	portrait.AddTitle("Watermark", 2)
	portrait.AddText("A diagonal CONFIDENTIAL watermark is stored as VML in every section header.")

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
	cols.AddText("The second paragraph continues the flow so the columns fill naturally.")

	land := doc.AddSection(style.Section{
		Orientation: style.OrientationLandscape,
		BreakType:   "nextPage",
		PageSizeW:   style.DefaultPageWidth,
		PageSizeH:   style.DefaultPageHeight,
		MarginTop:   style.DefaultMargin, MarginBottom: style.DefaultMargin,
		MarginLeft:  style.DefaultMargin, MarginRight:  style.DefaultMargin,
	})
	land.AddHeader().AddText("Landscape header")
	land.AddTitle("Landscape section", 1)
	land.AddText("w:orient=\"landscape\" on w:pgSz. Page width and height are swapped by Word.")

	if err := doc.AddMarkdown("# Markdown\n\nParagraph with **bold**."); err != nil {
		log.Fatal(err)
	}
	if err := doc.AddHTML("<h2>HTML</h2><p>Hello from AddHTML.</p>"); err != nil {
		log.Fatal(err)
	}

	if err := doc.Save("layout.docx"); err != nil {
		log.Fatal(err)
	}
}
```

本程序写出的 OpenXML：

| 功能 | 节点 |
| --- | --- |
| 两栏 | `w:cols w:num="2" w:space="720" w:sep="1"` |
| 首页页眉 | `w:titlePg` + `w:headerReference w:type="first"` |
| 水印 | 页眉中的 VML `PowerPlusWaterMarkObject` |
| 保护 | `w:documentProtection w:edit="readOnly"` |
| 目录 | `w:instrText` `TOC \o "1-3" \h \z \u` |

示例：[`examples/v0.5.0_demo`](https://github.com/yunkeweb/go-word/tree/main/examples/v0.5.0_demo) 以及 [实战案例库](./examples) 案例 A。
