# DOCX 转 HTML

从 **v0.12.0** 起提供，**v0.12.1** 修复 DOCX 样式还原，**v0.12.2** 修复纸张布局与超链接外观，**v0.13.0** 提升常见文档的版式保真度，**v0.14.0** 增强内容完整性，**v0.15.0** 增强复杂对象输出，**v0.16.0** 提升表格和字体保真度。使用 Go 标准库将已加载的文档或 DOCX 文件转换为 HTML 片段或完整页面，无需 Microsoft Word、LibreOffice、cgo 或外部转换进程。

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

**`RenderHTMLFile` 返回 HTML 字节，不接收输出文件名，也不负责写文件。** 由调用方保存结果。仓库的 [examples/docx_to_html](https://github.com/yunkeweb/go-word/tree/v0.13.0/examples/docx_to_html) 还提供了自动生成输入 DOCX 的可运行示例。

## 纸张预览与流式内容

从 **v0.12.2** 起，`Standalone: true` 配合 `IncludeCSS: true` 会按照各节的纸张尺寸、方向和页边距显示居中的白色页面。正文在 DOCX 的版心宽度内换行，不再铺满浏览器窗口；负段落缩进可以伸入页边距，避免在视口左侧被裁切。上方完整程序无需更改选项即可启用。

屏幕预览中的长分节会向下延伸，**不复现 Word 的自动分页边界**。窄屏通过横向滚动保留正文行宽。打印使用具名 CSS `@page` 设置纸张和页边距，实际效果取决于浏览器支持及打印设置。设备字体、文档网格、重复页眉页脚及 Word 分页规则仍可能带来视觉差异。

如果内容应随业务页面容器宽度排版，使用 `HTMLOptions{}` 返回片段；如果需要不带纸张预览的完整页面，使用 `HTMLOptions{Standalone: true, IncludeCSS: false}`，再自行提供 CSS。两种方式均保留支持的文字和段落内联格式。

DOCX 明确指定的链接颜色和下划线设置会直接应用到链接元素，包括自动颜色和取消下划线，避免被浏览器默认链接样式覆盖。

## v0.14.0 内容完整性

读取器会保留 Word 字段的缓存结果、脚注/尾注/批注引用、修订信息和结构化文档标签（SDT）。字段不会在 Go 中重新计算；没有缓存结果的字段会输出稳定占位符并产生诊断。

使用 `RevisionMode` 选择修订显示方式：

```go
word.HTMLOptions{RevisionMode: word.HTMLRevisionFinal}    // 显示插入，隐藏删除
word.HTMLOptions{RevisionMode: word.HTMLRevisionOriginal} // 显示删除内容
word.HTMLOptions{RevisionMode: word.HTMLRevisionMarkup}   // 使用 <ins>/<del> 同时显示
```

脚注、尾注和批注会生成文档内跳转及返回链接；悬空引用保留可见编号但不会生成坏链接。SDT 以只读语义容器输出，保留类型、alias、tag、日期格式、下拉选项和当前值；复选框使用 disabled input，不会生成表单提交、脚本或事件处理器。

## v0.15.0 复杂对象

DrawingML 形状和文本框会安全渲染为 HTML/CSS 容器。可以表达的几何、尺寸、填充、边框、旋转、翻转和内部文字会输出为 `data-*` 属性及内联样式；未知几何会保留为诊断，而不会静默当成图片。该输出面向可读布局，不模拟 Word 画布，也不承诺逐像素定位。

图表不依赖 JavaScript，输出为 `<figure class="goword-chart">` 和语义化 `<table class="goword-chart-data">`。表格包含标题、图例、系列名、类别、数值及 `scope` 无障碍属性，组合图保留第二坐标轴信息。图表没有缓存数值时，读取器会保留系列并产生诊断，渲染器仍输出可读的表格降级。

转换期间不会主动下载外链图片。只有显式提供 `ImageURL` 回调时才会发布安全 URL，现有协议校验和诊断仍然生效，从而避免不可预期的网络访问。

## v0.16.0 表格与字体保真

当前开发版会读取 DOCX 表格网格宽度、表格和单元格尺寸、固定布局、边框、间距、底色、行属性、单元格内边距及合并信息。HTML 以 `<colgroup>` 输出网格列宽，并保留支持的表格和单元格宽度，包括百分比宽度。`gridSpan` 和有效的 `vMerge` 组会映射为 `colspan` 和 `rowspan`；跨行合并中断或跨度不一致时，会保留受影响单元格的内容，避免吞掉文本。

Run 字体会分别保留 ASCII、High ANSI、东亚和复杂脚本字体及语言设置，同时保留现有 `style.Font.Name` 回退。DOCX 包含主题部件时，主题颜色、底色和边框会解析 `themeTint` 与 `themeShade`。这些改进提升了导入 DOM 和可读 HTML 的保真度，但不等于完整复刻 Word 表格排版或混合脚本的字体选择规则。

## v0.13.0 的保真度提升

v0.13.0 在保持现有片段和完整页面入口兼容的前提下，保留更多版式元数据：

- 段落分页提示映射为 `break-before`、`break-after`、`break-inside`、`widows` 和 `orphans`。这些是浏览器提示，渲染器不提供与 Word 等价的屏幕自动分页引擎。
- 完整纸张预览输出带分节信息的 `goword-header` 与 `goword-footer` 容器。只有单节且各有一个默认页眉、页脚时，打印才启用固定重复；首页、偶数页和多分节场景保留显式容器，不会被错误重复。
- 重复表头行输出 `<thead>`，正文行输出 `<tbody>`；`cantSplit` 映射为 `break-inside:avoid`，并在存在时输出固定表格布局、边框、间距、单元格内边距和行高。
- Drawing 图片保留 EMU 尺寸、对齐、边距、偏移，以及 inline、square、tight、behind、in-front 等常见环绕方式。未知环绕值仍输出图片并产生 `HTMLDiagnostic`；严格模式会返回错误。
- 文档存在 `word/theme/theme1.xml` 时解析主题字体和颜色引用。直接字体、颜色优先；主题缺失时省略无效 CSS。`FallbackFont`、字距、字体缩放、字符位置、空白保留、自定义制表位和禁止自动断词也会输出到 CSS。

这些能力改善屏幕阅读和浏览器打印，但不承诺逐像素复现 Word 分页。接入新的文档类型前，请使用目标浏览器和实际打印设置验证代表性样本。

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
| `RevisionMode` | `HTMLRevisionFinal` | 选择最终稿、原稿或修订标记模式 |

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

`HTMLDiagnostic` 包含兼容的 `ElementType` 和 `Message`，并在可用时提供 `Part`、`Path`、`ID`。`RenderHTMLWithDiagnostics` 在严格模式失败时仍保留诊断与已生成输出；返回字节的便捷 API 在此情况下只返回错误，不提供可用 HTML。

诊断只覆盖到达渲染器的元素，**不是完整 OOXML 保真审计**。读取器没有重建的内容无法在此阶段报告。ZIP、XML 和关系问题请另外使用[包诊断](./diagnostics)。

## 支持范围与限制

以下样式还原能力需要 v0.12.1 或更高版本；纸张布局和明确指定的超链接外观修复需要 v0.12.2。

| 内容 | 输出与边界 |
| --- | --- |
| 段落、标题、文本 run、链接 | 语义标签、行内换行、字体、字号、颜色、强调、下划线、背景色、对齐、缩进、段间距、行距、分页提示、自定义制表位和禁止自动断词。正文及表格单元格中的段落会解析文档默认样式、默认段落样式，以及具名段落/字符样式的 `basedOn` 继承；直接格式优先。数字样式 ID 可通过标题名称和大纲级别识别。存在主题部件时解析常见主题字体和颜色引用；复杂 Word 样式仍有限制 |
| 列表 | 嵌套列表、常见十进制/字母/罗马数字、中文计数编号、起始值及按编号 ID 续接；不完整复现自定义组合编号和所有 Word 重启规则 |
| 表格 | 嵌套表格及 DOM 合并属性映射为 HTML，包括 `rowspan` / `colspan`；存在时输出固定布局、边框、间距、单元格内边距、重复表头和不可拆分行；读取任意 DOCX 合并属性仍有限制 |
| 书签 | DOM 书签输出为锚点，内部链接指向锚点 |
| 节与分页 | 多节使用 `section` 和 `data-break-type`；显式分页使用 `goword-page-break`。分节元数据不等于 Word 分页引擎 |
| 页眉页脚 | 可选输出带 `data-type`（`default`、`first`、`even`）的分节容器。简单单节默认页眉页脚在打印中使用固定 CSS 重复；复杂按页选择场景保留显式 DOM，不自动模拟。导入页眉页脚图形的能力有限 |
| 图片 | 可用 drawing 图片及尺寸/替代文本；缺失的嵌入引用保留供诊断。外链 drawing 和 VML 支持不完整 |
| 字段 | 读取并渲染 DOCX 中保存的缓存结果；动态字段不计算，缺少缓存时诊断并输出占位符 |
| 脚注、尾注、批注 | 引用、正文容器和返回链接可导航；悬空引用保留编号并诊断，复杂批注界面不模拟 |
| 修订 | 最终稿、原稿和 markup 三种模式；显示模式使用 `<ins>` / `<del>`，不复现 Word 逐字修订界面 |
| 内容控件（SDT） | 只读输出类型、alias、tag、当前值及部分元数据；未知类型保留可见文本并诊断，不生成可提交表单 |
| 形状和文本框 | 支持的 DrawingML 几何、尺寸、填充、边框、旋转、翻转及内部文字映射为安全 HTML/CSS 容器；未知几何保留并产生诊断，不模拟 Word 画布和逐像素定位 |
| 图表 | 缓存的图表系列输出为不依赖 JavaScript 的语义化表格，包含标题、图例、类别、数值和组合图坐标轴信息；缺失缓存会产生诊断并保留可读行 |
| 公式 | DOM 容器可输出文字；公式排版不完整，保留的未支持元素输出占位符 |

转换目标是可读的 HTML，不保证与 Word 页面逐像素一致。严格模式成功也不代表无损转换；接入新类别文档前，请先验证有代表性的样本。

<a id="升级至-v0-12-0"></a>
<a id="升级至-v0-12-1"></a>

## 升级至 v0.16.0

```sh
go get github.com/yunkeweb/go-word@v0.16.0
```

`style.Font` 和 `style.Table` 新增了脚本字体与表格网格宽度字段。请将这两个公开结构体的按位置初始化字面量改为具名字段。HTML 快照可能变化：表格网格现在会输出 `<colgroup>`，表格和单元格百分比宽度会保留，格式不匹配的跨行合并会保留原单元格而不折叠。渲染器提升了可读性和保真度，但仍不复刻 Word 的完整分页与表格排版。

## 升级至 v0.15.0

```sh
go get github.com/yunkeweb/go-word@v0.15.0
```

v0.15.0 增加 DrawingML 形状/文本框 HTML 以及语义化图表表格。依赖 HTML 快照的应用需要更新图表占位符和形状标记。内存中创建的 `Chart` 不再被视为不支持元素；严格模式只会因缺失图表缓存或未知保留结构等真实诊断失败。外链图片仍需显式启用。

## 升级至 v0.13.0

```sh
go get github.com/yunkeweb/go-word@v0.13.0
```

v0.13.0 保留更多分页提示、表格布局元数据、图片环绕元数据、常见主题字体/颜色引用，并为简单默认页眉页脚增加打印重复。现有片段输出和零值 `HTMLOptions` 保持兼容。屏幕输出仍不实现 Word 自动分页；如果依赖表格边框、单元格内边距、主题字体或页眉页脚位置，请检查 HTML 快照和自定义 CSS。
## 升级至 v0.12.2

```sh
go get github.com/yunkeweb/go-word@v0.12.2
```

v0.12.2 在同时开启 `Standalone` 和 `IncludeCSS` 时增加纸张预览分节容器及打印 CSS，并将明确指定的超链接格式应用到链接元素。请更新受影响的 HTML 快照与自定义 CSS；如果应用自行控制页面布局，可使用片段或关闭 `IncludeCSS`。

从 v0.12.0 或更早版本升级时，还包含 v0.12.1 对 DOCX 样式还原及文字、段落内联 CSS 的修复。`style.Spacing` 新增了 `BeforeSet` 和 `AfterSet`：将对应标记设为 `true` 可保留显式零间距。请使用具名字段初始化，按位置初始化的字面量需要调整。

模块路径和纯 Go 运行要求不变。相较 v0.11.0，HTML API 为新增能力。如果使用过开发快照，请更新 HTML 快照测试和 CSS，适配 section 包裹、列表 `type`/`start` 属性、合并单元格属性以及正确闭合的分页元素。缺失嵌入图片引用现在可能导致严格模式失败。

`style.ListItem` 新增 `Start`（零值默认从 1 开始）。请使用具名字段初始化；旧版按位置初始化的结构体字面量需要调整。`element.Container.AppendElement` 追加已有元素并更新父节点；移动元素时请先将它从旧容器移除。

另见[性能基准](./benchmarks)和[更新日志](https://github.com/yunkeweb/go-word/blob/main/CHANGELOG.md)。
