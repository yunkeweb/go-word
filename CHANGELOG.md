# Changelog

## v0.16.0 — 2026-11-26

### Table and typography fidelity

- Read and round-trip DOCX table grids, dimensions, layout, borders, spacing, shading, row properties, cell padding, and merge metadata; render grid widths as HTML `<colgroup>` columns.
- Preserve cell content when vertical merges are interrupted or merge spans disagree, and retain percentage table/cell widths.
- Preserve script-specific run fonts and language properties, and resolve theme tint/shade for text colors, shading, and borders.
- Add OOXML reader, DOCX round-trip, HTML structure, and fallback regression coverage for tables, nested tables, merges, fonts, and theme colors.
- Upgrade note: `style.Font` and `style.Table` gain fields; use keyed struct literals because unkeyed literals need to be updated.
- 读取并往返保留 DOCX 表格网格、尺寸、布局、边框、间距、底色、行属性、单元格内边距及合并信息；HTML 以 `<colgroup>` 输出网格列宽。
- 跨行合并中断或合并跨度不匹配时保留单元格内容，并保留表格和单元格的百分比宽度。
- 分别保留脚本字体与语言属性，并解析文本颜色、底色和边框的主题色 tint/shade。
- 增加覆盖表格、嵌套表格、单元格合并、字体和主题色的 OOXML 读取、DOCX 往返、HTML 结构及降级回归测试。
- 升级提示：`style.Font` 和 `style.Table` 新增字段；请使用具名字段初始化，原有按位置初始化的字面量需要调整。

## v0.15.0 — 2026-10-29

### DOCX to HTML complex objects / DOCX 转 HTML 复杂对象

- Render DrawingML shapes and text boxes as safe HTML/CSS containers with supported geometry, dimensions, fill, outline, rotation, flipping, and inner text.
- Render charts as JavaScript-free semantic tables with captions, legends, series names, category/value caches, accessibility scopes, and combination-chart axis metadata.
- Preserve missing chart caches and unknown DrawingML structures through source-aware diagnostics and readable fallback output.
- Keep external image URLs opt-in; conversion does not download remote resources by default.
- Add complex-object reader, renderer, and strict-mode regression coverage while retaining pure-Go, no-cgo operation.
- 将 DrawingML 形状和文本框安全渲染为 HTML/CSS 容器，保留支持的几何、尺寸、填充、边框、旋转、翻转和内部文字。
- 将图表渲染为不依赖 JavaScript 的语义化表格，包含标题、图例、系列名、类别/数值缓存、无障碍 scope 及组合图坐标轴信息。
- 通过带来源信息的诊断和可读降级保留缺失图表缓存及未知 DrawingML 结构；外链图片默认不下载。

## v0.14.0 — 2026-10-18

### DOCX to HTML content integrity
- Preserve cached field results and diagnose fields without cached values or malformed field structure.
- Resolve footnote, endnote, and comment references with stable return links, visible dangling-reference markers, and unique HTML IDs.
- Add final, original, and markup rendering modes for tracked revisions.
- Read and render structured document tags as safe, read-only HTML, including rich-text run formatting, dates, lists, placeholders, and disabled checkboxes.
- Add content-integrity regression fixtures and source-aware conversion diagnostics while retaining pure-Go, no-cgo operation.
- 增强 DOCX 转 HTML 内容完整性：保留字段缓存结果，诊断缺失缓存和结构异常；解析脚注、尾注、批注引用并生成稳定返回链接；支持最终稿、原稿和修订标记模式；安全只读渲染 SDT 内容控件及富文本 run 格式、日期、下拉选项、占位符和禁用复选框。

## v0.13.0 — 2026-10-03

### DOCX to HTML fidelity / DOCX 转 HTML 保真度

- Preserve paragraph pagination hints, section-aware header/footer containers, safe print repetition for a simple default header/footer pair, repeating table headers, fixed table layout, table borders/spacing/cell padding, row heights, EMU image dimensions and common wrapping metadata.
- Resolve common theme font/color references, add font fallback and additional font/paragraph metrics, and report unknown image wrapping through `HTMLDiagnostic` instead of silently dropping the image.
- 保留段落分页提示、分节页眉页脚容器、简单默认页眉页脚的安全打印重复、重复表头、固定表格布局、表格边框/间距/单元格内边距、行高、EMU 图片尺寸及常见环绕元数据。
- 解析常见主题字体/颜色引用，增加字体回退及更多字体/段落度量；未知图片环绕通过 `HTMLDiagnostic` 报告，不再静默丢弃图片。

## v0.12.2 — 2026-10-03

### Fixed / 修复

- Match standalone HTML screen and print layout to DOCX section paper geometry, including page size, orientation, margins, centered paper previews, narrow-screen overflow behavior and named print-page rules. Preserve flowing fragments and standalone output without built-in CSS. Screen previews preserve content width but do not implement Word's automatic pagination.
- 让完整 HTML 的屏幕及打印布局匹配 DOCX 分节纸张属性，包括纸张尺寸、方向、页边距、居中纸张预览、窄屏溢出行为及具名打印页面规则；片段和不带内置 CSS 的完整页面保持流式布局。屏幕预览保留版心宽度，但不实现 Word 自动分页。
- Apply explicit hyperlink color and underline settings to anchor elements so browser defaults cannot change DOCX link appearance.
- 将超链接明确的颜色和下划线设置应用到链接元素，避免浏览器默认样式改变 DOCX 外观。

### Tests / 测试

- Add synthetic section-layout tests for page geometry, margins, orientation, print CSS, multiple sections, option boundaries and entry-point parity. Validate the supplied DOCX in Chrome, print-to-PDF, narrow-screen layout and text-order checks.
- 新增分节布局合成测试，覆盖纸张尺寸、页边距、方向、打印 CSS、多分节、选项边界和入口一致性；在 Chrome、打印 PDF、窄屏布局中验证用户提供的 DOCX，并核对文本顺序。

## v0.12.1 — 2026-10-03

### Fixed / 修复

- Restore common DOCX body and table-cell formatting from document defaults, default paragraph styles, named paragraph/character styles and `basedOn` inheritance. Preserve direct overrides, including explicit false/none/zero values, and safely handle missing or cyclic parent styles.
- 从文档默认样式、默认段落样式、具名段落/字符样式及 `basedOn` 继承恢复正文和表格单元格的常用格式；保留直接格式覆盖（包括显式 false/none/零值），并安全处理缺失或循环父样式。
- Preserve fonts, sizes, underlining, shading, paragraph alignment, indentation, spacing and line-spacing rules in HTML. Recognize headings with numeric style IDs and retain their formatted text runs; normalize OOXML RGB colors for CSS and render Chinese counting lists with their start values.
- HTML 保留字体、字号、下划线、背景色、段落对齐、缩进、段间距及行距规则；识别数字样式 ID 的标题并保留其文本 run 格式；将 OOXML RGB 颜色转换为有效 CSS，并保留中文计数编号及起始值。
- Retain inline `w:br` / `w:cr` breaks and hyperlink run formatting in their original text order. DOCX round trips retain inline breaks and formatted heading runs.
- 按原始文本顺序保留行内 `w:br` / `w:cr` 换行及超链接 run 格式；DOCX 往返保留行内换行和标题文本 run 格式。

### Tests / 测试

- Add synthetic OOXML regression fixtures for style inheritance, direct overrides, headings, numbering, colors, line spacing, hyperlinks and DOCX round trips, without including user documents.
- 新增独立构造的 OOXML 回归用例，覆盖样式继承、直接格式覆盖、标题、编号、颜色、行距、超链接和 DOCX 往返，不包含用户原始文档。

### Upgrade notes / 升级说明

- `style.Spacing` adds `BeforeSet` and `AfterSet` to preserve explicit zero spacing. Use keyed struct literals; positional literals need updating. Imported HTML now includes more inline formatting, so update affected snapshots and custom CSS.
- `style.Spacing` 新增 `BeforeSet` 和 `AfterSet`，用于保留显式零间距。请使用具名字段初始化，按位置初始化的字面量需要调整。导入后输出的 HTML 会包含更多内联格式，请相应更新快照测试与自定义 CSS。

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
