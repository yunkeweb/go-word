# 文档无损合并

`AppendDocument` 把 `src` 的每一节克隆到 `dst` 上，并重映射在同一个 OpenXML 包里会冲突的标识符。

```go
dst := word.New()
src := word.New()
// ... 分别填充两份文档 ...

err := dst.AppendDocument(src, word.MergeOptions{
	StylePrefix:    "src_",
	BookmarkPrefix: "src_",
	SectionBreak:   "nextPage",
})
```

## 重映射内容

| 标识符 | 行为 |
| --- | --- |
| 段落 / 表格 **样式 ID** | 冲突名称加上 `StylePrefix`（`Note` → `src_Note`） |
| **书签** 名（`w:bookmarkStart`） | 冲突名称加前缀；本来唯一的名称保留 |
| 内部超链接 **锚点** | 改写为新的书签名 |
| **图片 / 媒体** | 清空 `RelationID`；写出时分配新的 `rId` 与 `word/media/imageN` |
| 分节符 | 克隆后的第一节使用 `SectionBreak`（默认 `nextPage`） |

`rId` 在写出时分配（`word2007Writer.nextRel`），合并后的 ZIP 里两份文档不会共用关系 ID。

```go
if err := dst.AppendDocument(nil, word.MergeOptions{}); err != nil {
	// word: nil source document
}
```

[`examples/v0.8.0_demo`](https://github.com/yunkeweb/go-word/tree/main/examples/v0.8.0_demo) 演示了同时保留两套 `Note` 样式、`shared` / `src_shared` 书签以及两张 PNG。
