# 高级排版与保护

节属性、水印、文档保护与 TOC 域。

## 多栏排版

`SetColumns` 写出符合 XSD 的 `w:cols` 节点（`num`、`space`、`sep`）：

```go
sec.SetColumns(2, 720, true)
// <w:cols w:num="2" w:space="720" w:sep="1"/>
```

| 参数 | 含义 |
| --- | --- |
| `num` | 栏数（`w:num`） |
| `space` | 栏间距，twip（`w:space`，默认 720） |
| `showLine` | 分隔线（`w:sep="1"`） |

分栏作用在 **节** 上；正文要回到单栏时请再 `AddSection`。横纵向混排同样如此：新节使用 `style.OrientationLandscape`。

## 水印

```go
doc.SetTextWatermark("CONFIDENTIAL")
doc.SetImageWatermark(pngBytes)
```

文字水印是写在每节页眉里的 Word 原生 VML（`PowerPlusWaterMarkObject`）。

## 只读保护

```go
err := doc.Protect(word.ProtectTypeReadOnly, "secret")
```

模式：`ProtectTypeReadOnly`、`ProtectTypeComments`、`ProtectTypeTrackedChanges`、`ProtectTypeForms`。非空密码使用 Office SHA-1 / 100000 次迭代哈希（ECMA-376 `w:documentProtection`）。

## 目录

```go
sec.AddTOC(nil, nil, 1, 3)
```

写出 `TOC` 域（`TOC \o "1-3" \h \z \u`）。`AddTitle` 的标题带大纲级别，Word 可刷新该域。

## Markdown 与 HTML

```go
_ = doc.AddMarkdown("# 标题\n\n带 **加粗** 的段落。")
_ = doc.AddHTML("<h1>标题</h1><p>你好</p>")
```

水印、保护与 TOC 的合集见 [`examples/v0.5.0_demo`](https://github.com/yunkeweb/go-word/tree/main/examples/v0.5.0_demo)。
