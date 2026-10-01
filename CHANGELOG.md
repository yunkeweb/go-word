# Changelog

## Unreleased / 开发中

- Read section properties, common header/footer parts, notes, comments, tracked changes, and list numbering.
- Support template macros split across adjacent Word text runs.
- Add NewWithOptions and ValidatePackage for instance defaults and package diagnostics.`r`n`r`n## v0.10.0 — 2026-10-01

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
