# 基础 DOM 操作

`Document` 保存命名样式与一个或多个 `Section`。几乎所有正文元素都加在容器上（`Section`、`Header`、`Footer`、`Cell`、`TextRun`）。

## 文档与节

```go
doc := word.New()
doc.SetDefaultFontName("Calibri")
doc.SetDefaultFontSize(11)

info := doc.GetDocInfo()
info.Title = "季度报告"
info.Creator = "GoWord"

sec := doc.AddSection(style.Section{
	Orientation: style.OrientationPortrait,
	MarginTop:   1440, MarginBottom: 1440,
	MarginLeft:  1440, MarginRight:  1440,
})
```

`Save` / `WriteTo` 将 `word/document.xml` 流式写入 ZIP。`IOFactory` 名称与 PHPWord 兼容：

```go
w, err := word.CreateWriter(doc, "Word2007")
```

## 段落与富文本

```go
doc.AddFontStyle("strong", style.Font{Bold: true, Size: 14, Name: "Calibri"})
doc.AddParagraphStyle("center", style.Paragraph{Alignment: style.JcCenter})
doc.AddTitleStyle(1, style.Font{Bold: true, Size: 18}, style.Paragraph{
	Spacing: style.Spacing{After: 240},
})

sec.AddTitle("Welcome to GoWord", 1)
sec.AddText("Hello, Word 2007.", "strong", "center")

run := sec.AddTextRun()
run.AddText("Bold ", style.Font{Bold: true})
run.AddText("and italic.", style.Font{Italic: true})

sec.AddLink("https://github.com/yunkeweb/go-word", "GoWord on GitHub")
sec.AddBookmark("intro")
sec.AddPageBreak()
```

## 表格

```go
tbl := sec.AddTable(style.Table{Width: 9000})
hdr := tbl.AddRow()
hdr.AddCell(4500).AddText("项目")
hdr.AddCell(4500).AddText("金额")

row := tbl.AddRow()
row.AddCell(4500).AddText("纸张")
row.AddCell(4500).AddText("12")
```

单元格本身是容器：嵌套表、图片与公式都可以放进 `Cell`。模板行克隆时会把 `gridSpan` / `vMerge` 合并区域一起带走。

## 图片

```go
sec.AddImage("photo.png", style.Image{Width: 200, Height: 120})
sec.AddImageBytes("logo.png", pngBytes, style.Image{Width: 80, Height: 80})
```

写出时媒体部件落在 `word/media/imageN.ext`，并分配唯一 `rId`。

## 页眉与页脚

```go
h := sec.AddHeader()
h.AddText("GoWord — 内部资料")

f := sec.AddFooter()
f.AddPreserveText("第 {PAGE} 页 / 共 {NUMPAGES} 页")

sec.AddHeader(element.HeaderFirst) // 首页页眉
doc.SetDifferentFirstPage(true)
doc.SetEvenAndOddHeaders(true)
```

页眉类型：`element.HeaderAuto`（默认）、`HeaderFirst`、`HeaderEven`。

## 加载已有文件

```go
r, err := word.CreateReader("Word2007")
if err != nil {
	log.Fatal(err)
}
doc, err := r.Load("input.docx")
```

若只需抽取文本或图片、不想构建完整 DOM，见 [流式提取器](./streaming)。
