# DOCX 转 HTML

GoWord 可以读取 `.docx` 压缩包，并把重建后的文档渲染为 HTML 片段或完整页面。转换只使用 Go 标准库，不需要 Microsoft Word、LibreOffice、cgo 或外部运行时。

## 文件转换

源文件是路径时使用 `RenderHTMLFile`：

```go
package main

import (
    "log"

    "github.com/yunkeweb/go-word/element"
    word "github.com/yunkeweb/go-word"
)

func main() {
    err := word.RenderHTMLFile("input.docx", "output.html", word.HTMLOptions{
        Standalone: true,
        IncludeCSS: true,
    })
    if err != nil {
        log.Fatal(err)
    }
}
```

已经加载到内存的文档使用 `RenderHTML`；输入是 `io.Reader` 时使用 `RenderHTMLWithOptions`：

```go
html, err := doc.RenderHTML(word.HTMLOptions{
    Standalone: true,
    IncludeCSS: true,
    IncludeHeadersFooters: true,
})
```

`Standalone` 会输出 HTML 文档外壳，`IncludeCSS` 会加入内置布局样式。需要把结果嵌入已有页面时，可以关闭这两个选项。

## 图片与资源

默认将图片嵌入 data URI，使 HTML 可以独立传输。需要由 Web 服务单独提供图片时，使用 `HTMLImageURL` 或 `ImageURL` 回调：

```go
html, err := doc.RenderHTML(word.HTMLOptions{
    ImageMode: word.HTMLImageURL,
    ImageURL: func(img *element.Image) (string, error) {
        return "/assets/" + img.GetName(), nil
    },
})
```

回调返回的 URL 在写入 HTML 前会经过安全校验，请让回调只返回应用允许的资源路径。

## 诊断与严格模式

部分 Word 特性无法无损映射到 HTML，可以用诊断结果定位：

```go
result, err := doc.RenderHTMLWithDiagnostics(word.HTMLOptions{})
if err != nil {
    log.Fatal(err)
}
for _, diagnostic := range result.Diagnostics {
    log.Printf("%s: %s", diagnostic.ElementType, diagnostic.Message)
}
```

启用 `Strict: true` 后，遇到不支持的元素或缺失资源会直接失败，而不是输出诊断占位符。当前渲染器保留段落、富文本、标题、链接、列表、表格、合并单元格、图片、多节边界、页眉、页脚、分页、书签、脚注和批注；复杂图表与形状仍通过诊断处理。

## 资源限制

处理不可信文件时，使用 `ReadOptions` 配合 `RenderHTMLWithOptions`，限制压缩包大小、单部件大小、解压总量和条目数。这样可以在 HTML 渲染前控制 DOCX 输入成本。
