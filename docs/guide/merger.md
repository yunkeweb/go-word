# Document Merger

`AppendDocument` clones every section of `src` onto `dst`, then remaps identifiers that would collide inside one OpenXML package.

```go
dst := word.New()
src := word.New()
// ... fill both documents ...

err := dst.AppendDocument(src, word.MergeOptions{
	StylePrefix:    "src_",
	BookmarkPrefix: "src_",
	SectionBreak:   "nextPage",
})
```

## What gets remapped

| Identifier | Behaviour |
| --- | --- |
| Paragraph / table **style IDs** | Colliding names receive `StylePrefix` (`Note` → `src_Note`) |
| **Bookmark** names (`w:bookmarkStart`) | Colliding names are prefixed; unique names stay |
| Internal hyperlink **anchors** | Updated to the new bookmark names |
| **Images / media** | `RelationID` cleared; writer assigns fresh `rId` and `word/media/imageN` |
| Section break | First cloned section uses `SectionBreak` (`nextPage` by default) |

`rId`s are allocated at write time (`word2007Writer.nextRel`), so two documents never share a relationship ID in the merged ZIP.

```go
if err := dst.AppendDocument(nil, word.MergeOptions{}); err != nil {
	// word: nil source document
}
```

See [`examples/v0.8.0_demo`](https://github.com/yunkeweb/go-word/tree/main/examples/v0.8.0_demo) for a merge that keeps both `Note` styles, `shared` / `src_shared` bookmarks, and two PNG parts.
