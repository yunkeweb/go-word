# Changelog

## Unreleased / 开发中

## v0.12.0 — 2026-10-02

### Added / 新增

- Pure-Go DOCX to HTML conversion through `RenderHTML`, `RenderHTMLFile`, bounded-read `WithOptions` variants, and `Document.RenderHTML`. Supports fragments, standalone pages, built-in CSS, embedded images, and custom image URL callbacks without cgo or an external office suite.
- 新增纯 Go DOCX 转 HTML：`RenderHTML`、`RenderHTMLFile`、带读取限制的 `WithOptions` 变体及 `Document.RenderHTML`。支持片段、完整页面、内置 CSS、嵌入图片及图片 URL 回调，无需 cgo 或外部 Office 软件。
- `Document.WriteHTML` writes progressively to `io.Writer`; `RenderHTMLWithDiagnostics`, `HTMLDiagnostic`, and `HTMLUnsupportedError` expose conversion diagnostics and strict-mode errors.
- `Document.WriteHTML` 向 `io.Writer` 逐步写出；`RenderHTMLWithDiagnostics`、`HTMLDiagnostic` 和 `HTMLUnsupportedError` 提供转换诊断与严格模式错误。
- Render nested lists, common numbering formats and start values, DOM table merges, and bookmark anchors. DOCX round trips retain tested list numbering, section boundaries, header/footer types, and explicit page breaks.
- 渲染嵌套列表、常见编号格式与起始值、DOM 表格合并属性及书签锚点；DOCX 往返保留已覆盖场景中的列表编号、分节、页眉页脚类型及显式分页。
- `element.Container.AppendElement` appends existing DOM elements and updates their parent. Bilingual conversion guides and runnable examples are available at `https://go-word.yunkeweb.com/`.
- `element.Container.AppendElement` 可追加已有 DOM 元素并更新父节点。中英文转换指南与可运行示例已同步到 `https://go-word.yunkeweb.com/`。

### Fixed / 修复

- Missing embedded DOCX image references remain visible to the HTML renderer for diagnostics instead of being silently dropped. Page-break HTML elements are properly closed.
- 缺失的 DOCX 嵌入图片引用保留供 HTML 渲染器诊断，不再静默丢弃；修复分页 HTML 元素的闭合。
- Correct the conversion documentation: `RenderHTMLFile` returns bytes and accepts no output filename; applications must save the bytes themselves.
- 修正文档中的转换示例：`RenderHTMLFile` 返回字节，不接受输出文件名，需由应用自行保存。
- Pin the documentation build's Vite dependency to 6.4.3 to address development-server security advisories; the Go module still has no third-party dependencies.
- 将文档构建的 Vite 依赖固定为 6.4.3，修复开发服务器安全告警；Go 模块仍不包含第三方依赖。

### Tests / 测试

- Add DOM and DOCX round-trip tests, entry-point determinism checks, writer error propagation tests, missing-image XML fixtures, HTML escaping/structure fuzz seeds, and large-document rendering benchmarks.
- 新增 DOM 与 DOCX 往返测试、各入口确定性检查、写入错误传播测试、缺图 XML 用例、HTML 转义/结构模糊测试种子及大文档渲染基准。

### Upgrade notes / 升级说明

- `style.ListItem` adds `Start` (zero defaults to 1). Use keyed struct literals; positional literals from older releases need updating.
- `style.ListItem` 新增 `Start`（零值默认从 1 开始）；请使用具名字段初始化，旧版按位置初始化的字面量需要调整。
- Development snapshot users should update HTML snapshots/CSS for section wrappers, list `type`/`start`, merged-cell attributes, and page-break markup. Missing embedded images may now fail strict mode.
- 开发快照用户需更新 HTML 快照测试和 CSS，适配 section 包裹、列表 `type`/`start`、合并单元格属性和分页标记；缺失嵌入图片现在可能导致严格模式失败。
- Strict mode reports unsupported elements that reach the renderer; it cannot diagnose content omitted by the reader. HTML output is not a complete Word layout engine. `WriteHTML` can return an error after partial output; stage output before publishing when atomic delivery is required.
- 严格模式只诊断到达渲染器的未支持元素，无法检测读取器遗漏的内容；HTML 输出不等于完整 Word 排版引擎。`WriteHTML` 可能在部分写出后返回错误，需要原子交付时请暂存后再发布。

## v0.11.0 — 2026-10-01

- Read section properties, common header/footer parts, notes, comments, tracked changes, and list numbering.
- Support template macros split across adjacent Word text runs.
- Add NewWithOptions and ValidatePackage for instance defaults and package diagnostics.

## v0.10.0 — 2026-10-01

### Added / 新增

- Optional `ReadOptions` budgets for archive bytes, uncompressed part size, total declared expansion, and ZIP entry count across DOM, template, and streaming readers. Existing APIs and zero limits retain unlimited behavior.
- DOM、模板和流式读取器新增可选 `ReadOptions`，限制压缩包字节数、单部件解压大小、声明的解压总量及 ZIP 条目数；旧 API 和零值配置保持无限制行为。

### Fixed / 修复

- Propagate XML output and media registration errors; register nested resources in streaming output.
- Preserve comments across repeated saves, isolate header/footer relationships, and produce valid template image XML.
- Isolate template delimiters per processor; preserve consecutive cell text runs and remap bookmark links across merged sections.
- 修复 XML 写出与媒体注册错误丢失、流式输出嵌套资源遗漏、重复保存批注丢失、页眉页脚关系冲突、模板图片 XML 层级、模板分隔符实例隔离、表格连续文本保真和跨节合并书签链接。

### Performance / 性能

- Avoid an XML input copy and cache image, header/footer, chart, and OLE relationship lookups.
- 消除 XML 输入的多余复制，缓存图片、页眉页脚、图表和 OLE 的关系查找。
