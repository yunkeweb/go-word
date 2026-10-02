# DOCX 转 HTML

从 **v0.12.0** 起提供，**v0.12.1** 修复 DOCX 样式还原，**v0.12.2** 修复纸张布局与超链接外观。使用 Go 标准库将已加载的文档或 DOCX 文件转换为 HTML 片段或完整页面，无需 Microsoft Word、LibreOffice、cgo 或外部转换进程。

## 转换文件

将以下完整程序保存为 `main.go`，在同一目录放入 `input.docx`，然后执行 `go run main.go`：

```go
package main

import (
	"log"
	"os"

	word "github.com/yunkeweb/go-word"
)

func main() {
	html, err := word.RenderHTMLFile("input.docx", word.HTMLOptions{
		Standalone: true,
		IncludeCSS: true,
	})
	if err != nil {
		log.Fatal(err)
	}
	if err := os.WriteFile("output.html", html, 0o644); err != nil {
		log.Fatal(err)
	}
}
```

**`RenderHTMLFile` 返回 HTML 字节，不接收输出文件名，也不负责写文件。** 由调用方保存结果。仓库的 [examples/docx_to_html](https://github.com/yunkeweb/go-word/tree/v0.12.2/examples/docx_to_html) 还提供了自动生成输入 DOCX 的可运行示例。

## 纸张预览与流式内容

从 **v0.12.2** 起，`Standalone: true` 配合 `IncludeCSS: true` 会按照各节的纸张尺寸、方向和页边距显示居中的白色页面。正文在 DOCX 的版心宽度内换行，不再铺满浏览器窗口；负段落缩进可以伸入页边距，避免在视口左侧被裁切。上方完整程序无需更改选项即可启用。

屏幕预览中的长分节会向下延伸，**不复现 Word 的自动分页边界**。窄屏通过横向滚动保留正文行宽。打印使用具名 CSS `@page` 设置纸张和页边距，实际效果取决于浏览器支持及打印设置。设备字体、文档网格、重复页眉页脚及 Word 分页规则仍可能带来视觉差异。

如果内容应随业务页面容器宽度排版，使用 `HTMLOptions{}` 返回片段；如果需要不带纸张预览的完整页面，使用 `HTMLOptions{Standalone: true, IncludeCSS: false}`，再自行提供 CSS。两种方式均保留支持的文字和段落内联格式。

DOCX 明确指定的链接颜色和下划线设置会直接应用到链接元素，包括自动颜色和取消下划线，避免被浏览器默认链接样式覆盖。

## API 参考

以下签名属于 `word` 包；`io.Reader` 和 `io.Writer` 是标准库接口。

```go
func RenderHTML(r io.Reader, opts HTMLOptions) ([]byte, error)
func RenderHTMLFile(path string, opts HTMLOptions) ([]byte, error)
func RenderHTMLWithOptions(r io.Reader, readOpts ReadOptions, htmlOpts HTMLOptions) ([]byte, error)
func RenderHTMLFileWithOptions(path string, readOpts ReadOptions, htmlOpts HTMLOptions) ([]byte, error)
func (d *Document) RenderHTML(opts HTMLOptions) ([]byte, error)
func (d *Document) RenderHTMLWithDiagnostics(opts HTMLOptions) (HTMLRenderResult, error)
func (d *Document) WriteHTML(w io.Writer, opts HTMLOptions) error
```

| 参数 | 含义 |
| --- | --- |
| `path` | 输入 DOCX 的文件路径 |
| `r` | DOCX 压缩包的 Reader，不是 HTML |
| `d` | 已加载或通过代码创建的文档 |
| `w` | 输出 Writer，由调用方负责刷新缓冲区及关闭 |
| `readOpts` | 本次读取的 ZIP 限制，零值表示无限制 |
| `opts` / `htmlOpts` | 下表中的 HTML 输出选项 |

内存中的 DOCX 字节可使用 `RenderHTML(bytes.NewReader(raw), opts)`；需要查看转换诊断时，先用 `LoadBytesWithOptions` 加载文档。

### HTML 选项

| 字段 | 默认值 | 行为 |
| --- | --- | --- |
| `Standalone` | `false` | 开启时添加 doctype、head 和 body，否则返回片段 |
| `Title` | 空 | 可用时回退到文档元数据中的标题，仅用于完整页面 |
| `IncludeCSS` | `false` | 仅配合 `Standalone: true` 添加内置 CSS 与纸张预览；关闭时仍会输出支持的内联样式 |
| `IncludeHeadersFooters` | `false` | 输出各节 DOM 中已有的页眉页脚内容 |
| `ImageMode` | `HTMLImageDataURI` | 嵌入图片字节；`HTMLImageURL` 使用现有来源或目标 URL |
| `ImageURL` | `nil` | 签名为 `func(*element.Image) (string, error)`，优先于 `ImageMode` |
| `Strict` | `false` | 渲染器产生诊断时返回 `*HTMLUnsupportedError` |

文档保持不变且回调结果确定时，重复渲染会产生相同字节。并发渲染期间不要修改文档或其资源。

## 带读取限制的流式输出

以下是另一份完整程序：先按 ZIP 预算加载文档，再将 HTML 直接写入文件。

```go
package main

import (
	"log"
	"os"

	word "github.com/yunkeweb/go-word"
)

func main() {
	doc, err := word.LoadWithOptions("input.docx", word.ReadOptions{
		MaxArchiveSize: 32 << 20,
		MaxPartSize:    16 << 20,
		MaxTotalSize:   128 << 20,
		MaxEntries:     2000,
	})
	if err != nil {
		log.Fatal(err)
	}
	out, err := os.Create("output.html")
	if err != nil {
		log.Fatal(err)
	}
	renderErr := doc.WriteHTML(out, word.HTMLOptions{
		Standalone:            true,
		IncludeCSS:            true,
		IncludeHeadersFooters: true,
		Strict:                true,
	})
	closeErr := out.Close()
	if renderErr != nil {
		log.Fatal(renderErr)
	}
	if closeErr != nil {
		log.Fatal(closeErr)
	}
}
```

限制数值仅为示例，请根据文档规模调整。负数限制会被拒绝；可以用 `errors.Is(err, word.ErrReadLimitExceeded)` 判断读取预算超限。

`WriteHTML` 不构造完整 HTML 字节缓冲区，但 DOCX 读取仍构建内存 DOM，图片编码也会使用内存。写入失败或严格模式错误可能发生在**部分输出已经写出之后**。不允许暴露不完整结果时，先写临时文件，成功后再发布；HTTP 响应需要在发送响应头前确定是否成功时，先渲染为字节。

## 图片与资源

data URI 模式嵌入可用的图片数据，并输出 DOM 中的替代文本和尺寸；不会主动下载远程链接图片。

需要独立资源时，提供 `ImageURL` 回调，并导入 `github.com/yunkeweb/go-word/element`。回调负责保存或上传图片数据，返回浏览器可访问的 URL；可使用 `img.Data`、`img.Media.Data`、`img.GetName()` 和 `img.GetAltText()`。请生成唯一资源名、校验图片格式，并将存储错误从回调返回。渲染器不会自动创建 assets 目录。

未提供回调时，`HTMLImageURL` 只复用已有的安全来源或目标 URL。DOCX 中的媒体文件名不会自动变成应用托管的地址。回调错误、不安全 URL、无法获取的嵌入图片在普通模式下产生诊断，严格模式下导致失败。通过 URL 协议检查不代表地址真实存在，也不代表主机属于应用允许的范围。

## 转换诊断

已有 `doc` 时，导入 `errors` 与 `log`，即使严格模式返回错误也可以读取诊断：

```go
result, err := doc.RenderHTMLWithDiagnostics(word.HTMLOptions{Strict: true})
for _, diagnostic := range result.Diagnostics {
	log.Printf("%s: %s", diagnostic.ElementType, diagnostic.Message)
}
var unsupported *word.HTMLUnsupportedError
if errors.As(err, &unsupported) {
	log.Printf("conversion has %d unsupported elements", len(unsupported.Diagnostics))
} else if err != nil {
	log.Fatal(err)
}
```

`HTMLDiagnostic` 包含 `ElementType` 和 `Message`。`RenderHTMLWithDiagnostics` 在严格模式失败时仍保留诊断与已生成输出；返回字节的便捷 API 在此情况下只返回错误，不提供可用 HTML。

诊断只覆盖到达渲染器的元素，**不是完整 OOXML 保真审计**。读取器没有重建的内容无法在此阶段报告。ZIP、XML 和关系问题请另外使用[包诊断](./diagnostics)。

## 支持范围与限制

以下样式还原能力需要 v0.12.1 或更高版本；纸张布局和明确指定的超链接外观修复需要 v0.12.2。

| 内容 | 输出与边界 |
| --- | --- |
| 段落、标题、文本 run、链接 | 语义标签、行内换行、字体、字号、颜色、强调、下划线、背景色、对齐、缩进、段间距和行距。正文及表格单元格中的段落会解析文档默认样式、默认段落样式，以及具名段落/字符样式的 `basedOn` 继承；直接格式优先。数字样式 ID 可通过标题名称和大纲级别识别。主题字体/颜色及完整 Word 样式级联仍未完全覆盖 |
| 列表 | 嵌套列表、常见十进制/字母/罗马数字、中文计数编号、起始值及按编号 ID 续接；不完整复现自定义组合编号和所有 Word 重启规则 |
| 表格 | 嵌套表格及 DOM 合并属性映射为 HTML，包括 `rowspan` / `colspan`；读取任意 DOCX 合并属性仍有限制 |
| 书签 | DOM 书签输出为锚点，内部链接指向锚点 |
| 节与分页 | 多节使用 `section` 和 `data-break-type`；显式分页使用 `goword-page-break`。分节元数据不等于 Word 分页引擎 |
| 页眉页脚 | 可选输出，保留 `data-type`（`default`、`first`、`even`）；不按实际页码选择或重复，导入页眉页脚图形的能力有限 |
| 图片 | 可用 drawing 图片及尺寸/替代文本；缺失的嵌入引用保留供诊断。外链 drawing 和 VML 支持不完整 |
| 脚注、批注、形状、公式 | DOM 容器可输出文字；导入引用关联、几何及公式排版不完整。保留的未支持元素输出占位符 |

转换目标是可读的 HTML，不保证与 Word 页面逐像素一致。严格模式成功也不代表无损转换；接入新类别文档前，请先验证有代表性的样本。

<a id="升级至-v0-12-0"></a>
<a id="升级至-v0-12-1"></a>

## 升级至 v0.12.2

```sh
go get github.com/yunkeweb/go-word@v0.12.2
```

v0.12.2 在同时开启 `Standalone` 和 `IncludeCSS` 时增加纸张预览分节容器及打印 CSS，并将明确指定的超链接格式应用到链接元素。请更新受影响的 HTML 快照与自定义 CSS；如果应用自行控制页面布局，可使用片段或关闭 `IncludeCSS`。

从 v0.12.0 或更早版本升级时，还包含 v0.12.1 对 DOCX 样式还原及文字、段落内联 CSS 的修复。`style.Spacing` 新增了 `BeforeSet` 和 `AfterSet`：将对应标记设为 `true` 可保留显式零间距。请使用具名字段初始化，按位置初始化的字面量需要调整。

模块路径和纯 Go 运行要求不变。相较 v0.11.0，HTML API 为新增能力。如果使用过开发快照，请更新 HTML 快照测试和 CSS，适配 section 包裹、列表 `type`/`start` 属性、合并单元格属性以及正确闭合的分页元素。缺失嵌入图片引用现在可能导致严格模式失败。

`style.ListItem` 新增 `Start`（零值默认从 1 开始）。请使用具名字段初始化；旧版按位置初始化的结构体字面量需要调整。`element.Container.AppendElement` 追加已有元素并更新父节点；移动元素时请先将它从旧容器移除。

另见[性能基准](./benchmarks)和[更新日志](https://github.com/yunkeweb/go-word/blob/main/CHANGELOG.md)。
