# DrawingML 图表与形状

GoWord 写出 Word 2010 **wordprocessingShape**（`wps:wsp`）以及 ECMA-376 **chart** 部件（`word/charts/chartN.xml`）。

## 形状与文本框

| `ShapeType` | 预设几何 |
| --- | --- |
| `ShapeRect` | `rect` |
| `ShapeRoundRect` | `roundRect` |
| `ShapeArrow` | `rightArrow` |
| `ShapeTextBox` | `rect` + `wps:txbx` / `w:txbxContent` |

`ShapeOptions` 控制填充色、描边色/线宽、尺寸（EMU，1 英寸 = 914400）以及内部文字。`ShapeTextBox` 在文本为空时会写入一个空格，以便 Word 仍创建 `w:txbxContent`。

## 图表

常量：`ChartTypeBar`、`ChartTypeColumn`、`ChartTypeLine`、`ChartTypePie`、`ChartTypeArea`、`ChartTypeStackedArea`、`ChartTypePercentStackedArea`、`ChartTypeCombo`。

系列 XML 遵循 Word 图表 XSD 顺序（`idx` → `order` → `tx` → `spPr` → …），Word 打开时不会弹出修复对话框。组合图的次坐标系列使用第二条数值轴（带独立 `c:axId` 的 `c:valAx`）。

`*element.Chart` 上的辅助方法：

| 方法 | OpenXML |
| --- | --- |
| `AddComboSeries(kind, cats, vals, name, secondary)` | 组合图上的额外 `c:ser` |
| `SetLegendPosition` | `c:legend` / `c:legendPos`（`t`/`b`/`l`/`r`） |
| `SetMajorGridlines` | `c:majorGridlines` |
| `SetLineSmooth` | 折线系列上的 `c:smooth` |
| `SetLineMarker` | `c:marker`（`symbol` + `size` 5） |
| `SetDataLabels` | `c:dLbls`（面积图省略 `dLblPos`） |

## 完整示例

保存为 `main.go` 后执行 `go run .`。`drawing.docx` 包含全部图表类型与四种形状。

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
	doc.SetDefaultFontSize(11)

	info := doc.GetDocInfo()
	info.Title = "DrawingML charts and shapes"
	info.Creator = "GoWord"

	sec := doc.AddSection()
	sec.AddTitle("DrawingML shapes", 1)

	doc.AddShape(word.ShapeRect, word.ShapeOptions{
		FillColor: "5B9BD5", LineColor: "2E75B6",
		Width: 1828800, Height: 914400,
	})
	doc.AddShape(word.ShapeRoundRect, word.ShapeOptions{
		FillColor: "ED7D31", LineColor: "C45911",
		Text: "Rounded", Font: style.Font{Bold: true, Color: "FFFFFF"},
	})
	doc.AddShape(word.ShapeArrow, word.ShapeOptions{
		FillColor: "70AD47", LineColor: "548235",
	})
	doc.AddShape(word.ShapeTextBox, word.ShapeOptions{
		FillColor: "FFF2CC", LineColor: "BF8F00",
		Text: "w:txbxContent", Font: style.Font{Size: 12, Color: "595959"},
	})

	sec.AddTitle("Charts", 1)
	cats := []string{"Q1", "Q2", "Q3", "Q4"}

	sec.AddTitle("Bar / column / line / pie / area", 2)
	doc.AddChart(word.ChartTypeBar, cats, []float64{12, 18, 9, 15})
	doc.AddChart(word.ChartTypeColumn, cats, []float64{12, 18, 9, 15})
	line := doc.AddChart(word.ChartTypeLine, cats, []float64{8, 11, 14, 10})
	line.SetLineSmooth(true)
	line.SetLineMarker("circle")
	doc.AddChart(word.ChartTypePie, []string{"A", "B", "C"}, []float64{40, 35, 25})
	doc.AddChart(word.ChartTypeArea, cats, []float64{4, 7, 6, 9})

	sec.AddTitle("Stacked area", 2)
	area := doc.AddChart(word.ChartTypeStackedArea, cats, []float64{12, 18, 15, 22}, []float64{8, 11, 13, 16})
	area.Series[0].Name = "Hardware"
	area.Series[1].Name = "Software"
	area.Style.Title = "Quarterly Revenue Mix"
	area.Style.Width = 5486400
	area.Style.Height = 3200400
	area.SetLegendPosition(word.LegendBottom)
	area.SetMajorGridlines(true)
	area.SetDataLabels(word.ChartDataLabelOptions{
		ShowVal:  true,
		Position: word.DataLabelPosCenter,
	})

	sec.AddTitle("Combo (column + line, secondary axis)", 2)
	months := []string{"Jan", "Feb", "Mar", "Apr", "May", "Jun"}
	combo := doc.AddChart(word.ChartTypeCombo, months, []float64{120, 150, 140, 180, 210, 190})
	combo.Series[0].Name = "Sales"
	combo.Series[0].Kind = word.ChartTypeColumn
	combo.AddComboSeries(word.ChartTypeLine, months, []float64{8.2, 9.1, 8.7, 10.4, 11.2, 10.8}, "Margin %", true)
	combo.Style.Title = "Sales vs Margin"
	combo.Style.Width = 5486400
	combo.Style.Height = 3200400
	combo.Style.ValueNumFmt = "0"
	combo.Style.SecondaryValueNumFmt = "0.0"
	combo.SetLegendPosition(word.LegendRight)
	combo.SetMajorGridlines(true)
	combo.SetLineSmooth(true)
	combo.SetLineMarker("circle")
	combo.SetDataLabels(word.ChartDataLabelOptions{
		ShowVal:  true,
		Position: word.DataLabelPosTop,
	})

	if err := doc.Save("drawing.docx"); err != nil {
		log.Fatal(err)
	}
}
```

示例：[`examples/v0.7.0_demo`](https://github.com/yunkeweb/go-word/tree/main/examples/v0.7.0_demo) 与 [`examples/v0.8.0_demo`](https://github.com/yunkeweb/go-word/tree/main/examples/v0.8.0_demo)。
