# 水印与文档保护

文字水印是注入每个节页眉的 Word 原生 VML（`PowerPlusWaterMarkObject`）。图片水印使用 `WordPictureWatermark`（`v:imagedata`）。文档保护写出 `w:documentProtection`；设置密码时使用 Office SHA-1 / 100000 次迭代哈希。可编辑例外区域用 `w:permStart` / `w:permEnd` 包裹段落、单元格或整表。

## SetTextWatermark

### 签名

```go
func (d *Document) SetTextWatermark(text string, opts ...WatermarkOptions)
```

```go
type WatermarkOptions struct {
	Angle    float64 // 角度；0 视为 -45（Word rotation:315）
	Color    string  // VML fillcolor，如 silver 或 C0C0C0
	FontSize int     // 磅值；0 使用 Word 的 1pt + fitshape
	FontName string  // 默认 Calibri
	Opacity  float64 // 0–1；0 默认 0.5。大于 1 视为百分数
	Tile     bool    // 在 612×792 pt 页面上铺满网格
	Rows     int     // 行数，默认 3
	Cols     int     // 列数，默认 3
}
```

### 参数

| 名称 | 类型 | 说明 |
| --- | --- | --- |
| `text` | `string` | 水印文案。空字符串清除先前的文字水印。 |
| `opts` | `...WatermarkOptions` | 可选。省略时写出一条斜向银色、透明度 0.5 的水印（与 v0.8.0 相同）。 |

### 注意

- `Angle: 0` 视为 `-45°`（`rotation:315`）。传入非零角度才会覆盖。
- `Tile: true` 在整页绘制 `Rows × Cols` 网格（默认 3×3）。
- 水印在写出时应用到每个节页眉。你自己创建页眉也可以：写出器仍会注入 VML。

## SetImageWatermark / SetImageWatermarkFile

### 签名

```go
func (d *Document) SetImageWatermark(imageBytes []byte, opts ...ImageWatermarkOptions)
func (d *Document) SetImageWatermarkFile(path string, opts ...ImageWatermarkOptions) error
```

```go
type ImageWatermarkOptions struct {
	Washout bool    // Word 洗白（gain="19661f" blacklevel="22938f"）
	Scale   float64 // 写出时的尺寸倍率；0 保持默认大小
	Opacity float64 // 0–1；0 省略 v:fill opacity
}
```

### 参数

| 名称 | 类型 | 说明 |
| --- | --- | --- |
| `imageBytes` | `[]byte` | 已编码的 PNG/JPEG。空切片清除图片水印。 |
| `path` | `string` | `SetImageWatermarkFile` 的文件路径（PNG 或 JPEG）。 |
| `opts` | `...ImageWatermarkOptions` | 可选。省略则不洗白（与 v0.8.0 相同）。 |

### 注意

- `SetImageWatermark` 保持 `[]byte` 签名。源文件在磁盘上时使用 `SetImageWatermarkFile`。
- 图片水印增加一个由页眉关系引用的媒体部件。
- 洗白是 Word 的变灰滤镜，与 `Opacity` 彼此独立。

## Protect

### 签名

```go
func (d *Document) Protect(editing, password string) error
```

### 参数

| 名称 | 类型 | 说明 |
| --- | --- | --- |
| `editing` | `string` | `ProtectTypeReadOnly`、`ProtectTypeComments`、`ProtectTypeTrackedChanges`、`ProtectTypeForms`。空则只读。 |
| `password` | `string` | 可选。非空时使用 ECMA-376 文档保护哈希。 |

### 注意

- 这是**编辑限制**，不是 ZIP 加密。任何人仍可解压该包。
- 用户尝试编辑时 Word 会要密码。示例密码为 `goword`。

## AllowEdit {#allowedit}

### 签名

```go
func (t *Text) AllowEdit(groupOrUser string) *Text
func (t *TextRun) AllowEdit(groupOrUser string) *TextRun
func (c *Cell) AllowEdit(groupOrUser string) *Cell
func (t *Table) AllowEdit(groupOrUser string) *Table
```

### 参数

| 名称 | 类型 | 说明 |
| --- | --- | --- |
| `groupOrUser` | `string` | `Everyone` / `everybody` / `all` → `w:edGrp="everyone"`。也可为 `none`、`administrators`、`contributors`、`editors`、`owners`、`current`。其他字符串写成 `w:ed`（指定用户）。空则默认 everyone。 |

### 注意

- 例外区域只有在同时调用 `Protect` 时才会生效。
- 写出器在该元素的 XML 外围发出配对的 `w:permStart` / `w:permEnd`。

## 完整示例

```go
package main

import (
	"log"

	"github.com/yunkeweb/go-word"
	"github.com/yunkeweb/go-word/style"
)

func main() {
	doc := word.New()
	doc.SetDefaultFontName("Calibri")
	doc.SetTextWatermark("CONFIDENTIAL", word.WatermarkOptions{
		Angle: -45, Color: "C0C0C0", FontSize: 36, Opacity: 0.28,
		Tile: true, Rows: 3, Cols: 3,
	})
	if err := doc.Protect(word.ProtectTypeReadOnly, "goword"); err != nil {
		log.Fatal(err)
	}

	sec := doc.AddSection()
	sec.AddTitle("Protected contract", 1)
	sec.AddText("Standard clauses stay locked.")
	sec.AddText("Party A: ________________", style.Font{Bold: true}).AllowEdit("Everyone")

	tbl := sec.AddTable(style.Table{Width: 9000})
	hdr := tbl.AddRow()
	tbl.SetHeaderRow(hdr)
	hdr.AddCell(3000).AddText("Field", style.Font{Bold: true})
	hdr.AddCell(6000).AddText("Value", style.Font{Bold: true})
	row := tbl.AddRow()
	row.AddCell(3000).AddText("Contract no.")
	row.AddCell(6000, style.Cell{Shading: style.Shading{Fill: "E2F0D9"}}).
		AllowEdit("Everyone").
		AddText("CN-2026-001")

	if err := doc.Save("protected.docx"); err != nil {
		log.Fatal(err)
	}
}
```

从文件写入图片水印：

```go
err := doc.SetImageWatermarkFile("mark.png", word.ImageWatermarkOptions{
	Washout: true, Scale: 1.2, Opacity: 0.35,
})
```

### Word 中的效果

打开 `protected.docx` 后，正文后方是 3×3 灰色 **CONFIDENTIAL** 网格。状态栏显示只读。限制编辑列出密码哈希；输入 `goword` 即可解锁整份文件。不输入密码时，只有甲方段落与绿色合同编号单元格可以填写。更长的示例见 [`examples/v0.9.0_watermark_security`](https://github.com/yunkeweb/go-word/tree/main/examples/v0.9.0_watermark_security)。
