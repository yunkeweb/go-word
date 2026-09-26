# 表格与嵌套单元格

表格遵循 `w:tbl` → `w:tr` → `w:tc`。每个 `w:tc` 必须至少包含一个 `w:p`。单元格没有子节点时，GoWord 会补一个空段落。`Cell` 嵌入 `Container`，因此单元格里可以再放一张表。

## AddTable / AddRow / AddCell

### 签名

```go
func (c *Container) AddTable(styles ...any) *Table
func (t *Table) AddRow(height ...int) *Row
func (r *Row) AddCell(width int, st ...any) *Cell
func (t *Table) SetWidth(w int)
```

### 参数

| 名称 | 类型 | 说明 |
| --- | --- | --- |
| `styles` | `...any` | 可选 `style.Table`（`Width` 为 twip、对齐、边框、底纹）或样式名。 |
| `height` | `...int` | 可选行高，单位 twip（`w:trHeight`）。 |
| `width` | `int` | 首选单元格宽度，单位 twip（`w:tcW`）。 |
| `st` | `...any` | 可选 `style.Cell`。 |

### 单元格样式

| 字段 | OpenXML | 含义 |
| --- | --- | --- |
| `GridSpan` | `w:gridSpan` | 横向合并。跨度计入 `CountColumns`。 |
| `VMerge` | `w:vMerge` | 首行 `"restart"`，后续行 `"continue"`。 |
| `VAlign` | `w:vAlign` | `top` / `center` / `bottom`。 |
| `Shading.Fill` | `w:shd w:fill` | 不含 `#` 的十六进制填充。 |

## 嵌套表

在单元格上调用 `AddTable`：

```go
host := row.AddCell(6000)
inner := host.AddTable(style.Table{Width: 5800})
```

内层 `w:tbl` 是 `w:tc` 的子节点，嵌套表之后仍会有一个 `w:p`。

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
	sec := doc.AddSection()
	sec.AddTitle("Tables with nested cells", 1)

	tbl := sec.AddTable(style.Table{Width: 9000})
	hdr := tbl.AddRow()
	hdr.AddCell(3000, style.Cell{Shading: style.Shading{Fill: "1F4E79"}}).
		AddText("Region", style.Font{Bold: true, Color: "FFFFFF"})
	hdr.AddCell(6000, style.Cell{Shading: style.Shading{Fill: "1F4E79"}}).
		AddText("Breakdown", style.Font{Bold: true, Color: "FFFFFF"})

	row := tbl.AddRow()
	merged := row.AddCell(3000, style.Cell{VMerge: "restart", VAlign: "center"})
	merged.AddText("APAC")
	host := row.AddCell(6000)
	inner := host.AddTable(style.Table{Width: 5800})
	ir := inner.AddRow()
	ir.AddCell(2900).AddText("Hardware")
	ir.AddCell(2900).AddText("120")
	ir2 := inner.AddRow()
	ir2.AddCell(2900).AddText("Software")
	ir2.AddCell(2900).AddText("80")

	cont := tbl.AddRow()
	cont.AddCell(3000, style.Cell{VMerge: "continue"})
	cont.AddCell(6000).AddText("Continued APAC row uses w:vMerge continue.")

	span := tbl.AddRow()
	span.AddCell(9000, style.Cell{GridSpan: 2}).
		AddText("Footer spans both columns (w:gridSpan=2).")

	if err := doc.Save("table.docx"); err != nil {
		log.Fatal(err)
	}
}
```

### Word 中的效果

海军蓝双列表头，APAC 标签跨两行，右侧嵌套 2×2 网格，底行单元格横跨整表。[模板](./template) 的 `CloneRow` 会把 `vMerge` 组与 `gridSpan` 一起克隆。
