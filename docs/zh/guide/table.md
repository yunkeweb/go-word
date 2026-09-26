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
| `VAlign` | `w:vAlign` | `top` / `center` / `bottom`。优先使用 `Cell.SetVAlign`。 |
| `TextDir` | `w:textDirection` | `lrTb`、`tbRl`、`btLr` 等。优先使用 `Cell.SetTextDirection`。 |
| `Shading.Fill` | `w:shd w:fill` | 不含 `#` 的十六进制填充。 |

## SetHeader / SetCantSplit / SetVAlign / SetTextDirection {#setheader-setcantsplit-setvalign-settextdirection}

跨页重复页眉（`w:tblHeader`）与行禁止跨页断裂（`w:cantSplit`）写在 `w:trPr` 上。垂直对齐（`w:vAlign`）与文本方向（`w:textDirection`）写在 `w:tcPr` 上。行属性子节点顺序为 `cantSplit` → `trHeight` → `tblHeader`（CT_TrPrBase）。

### 签名

```go
func (t *Table) SetHeaderRow(row *Row) *Table
func (r *Row) SetHeader(v bool) *Row
func (r *Row) SetCantSplit(v bool) *Row
func (c *Cell) SetVAlign(v string) *Cell
func (c *Cell) SetTextDirection(v string) *Cell
```

### 参数

| 名称 | 类型 | 说明 |
| --- | --- | --- |
| `row` | `*Row` | 页眉行。`SetHeaderRow` 等于 `row.SetHeader(true)`。Word 会在每页表顶重复连续的页眉行。 |
| `v`（SetHeader） | `bool` | 写出 `w:tblHeader`。 |
| `v`（SetCantSplit） | `bool` | 写出 `w:cantSplit`，该行保持在同一页。 |
| `v`（SetVAlign） | `string` | `top`、`center` / `middle`、`bottom`、`both` / `justify`。 |
| `v`（SetTextDirection） | `string` | `lrTb` / `horizontal`、`tbRl` / `vertical`、`btLr`、`lrTbV`、`tbRlV`、`tbLrV`。 |

### 注意

- `SetHeader(true)` 与 `SetHeaderRow` 等价，任选其一。
- 页眉行常同时设置 `SetCantSplit`：重复页眉保持完整。
- `w:tcPr` 上 `w:textDirection` 写在 `w:vAlign` **之前**。

### 示例

```go
tbl := sec.AddTable(style.Table{Width: 9000})
hdr := tbl.AddRow(400)
tbl.SetHeaderRow(hdr)
hdr.SetCantSplit(true)
hdr.AddCell(3000).SetVAlign("center").AddText("Item", style.Font{Bold: true})
hdr.AddCell(6000).SetVAlign("center").AddText("Note", style.Font{Bold: true})

row := tbl.AddRow(1600)
row.SetCantSplit(true)
row.AddCell(1500).SetVAlign("center").SetTextDirection("tbRl").AddText("竖排")
row.AddCell(7500).SetVAlign("bottom").AddText("This row stays on one page.")
```

带 48 行重复页眉的示例见 [`examples/v0.9.0_table_advanced`](https://github.com/yunkeweb/go-word/tree/main/examples/v0.9.0_table_advanced)。

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
