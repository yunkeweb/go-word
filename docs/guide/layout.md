# Advanced Layout & Protect

Section properties, watermarks, document protection, and a TOC field.

## Multi-column layout

`SetColumns` writes a schema-valid `w:cols` node (`num`, `space`, `sep`):

```go
sec.SetColumns(2, 720, true)
// <w:cols w:num="2" w:space="720" w:sep="1"/>
```

| Argument | Meaning |
| --- | --- |
| `num` | Column count (`w:num`) |
| `space` | Gap in twips (`w:space`, default 720) |
| `showLine` | Separator line (`w:sep="1"`) |

Columns apply to the **section**, so start a new section when the rest of the document should go back to one column. Mixed portrait / landscape is the same idea: `AddSection` with `style.OrientationLandscape`.

## Watermark

```go
doc.SetTextWatermark("CONFIDENTIAL")
doc.SetImageWatermark(pngBytes)
```

Text watermarks are Word-native VML (`PowerPlusWaterMarkObject`) in every section header.

## Read-only protection

```go
err := doc.Protect(word.ProtectTypeReadOnly, "secret")
```

Modes: `ProtectTypeReadOnly`, `ProtectTypeComments`, `ProtectTypeTrackedChanges`, `ProtectTypeForms`. A non-empty password uses the Office SHA-1 / 100000-spin hash (ECMA-376 `w:documentProtection`).

## Table of contents

```go
sec.AddTOC(nil, nil, 1, 3)
```

Writes a `TOC` field (`TOC \o "1-3" \h \z \u`). Heading paragraphs from `AddTitle` carry outline levels so Word can refresh the field.

## Markdown and HTML

```go
_ = doc.AddMarkdown("# Heading\n\nParagraph with **bold**.")
_ = doc.AddHTML("<h1>Heading</h1><p>Hello</p>")
```

See [`examples/v0.5.0_demo`](https://github.com/yunkeweb/go-word/tree/main/examples/v0.5.0_demo) for watermarks, protection, and TOC together.
