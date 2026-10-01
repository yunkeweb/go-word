# 读取、模板与诊断

当前开发线补齐了常用 Word 部件的往返读取，并提供两个安全辅助 API。

## 实例级默认值

使用 `NewWithOptions` 为不同文档设置隔离的默认字体：

```go
doc := word.NewWithOptions(word.DocumentOptions{
    DefaultFontName: "Aptos", DefaultAsianFontName: "Microsoft YaHei",
    DefaultFontSize: 11, DefaultFontColor: "202124",
})
```

## 包诊断

`ValidatePackage` 不修改 DOCX ZIP，返回带部件路径的 XML、关系目标、重复关系 ID、重复书签和无效内部锚点诊断。

读取器会恢复 section 页面属性、常用页眉页脚、脚注尾注、批注元信息、插入/删除修订、列表段落编号以及正文行内图片。图片会保留关系目标、原始字节、`wp:extent` 尺寸（EMU 与像素）和 `wp:docPr@descr` 替代文本。`ValidatePackage` 会把暂不支持的 chart、shape、textbox 绘图报告为 warning，不会静默伪装成图片。`TemplateProcessor.SetValue` 也能识别被 Word 拆到相邻 `w:r` / `w:t` 节点中的宏，并保留 run 属性。
