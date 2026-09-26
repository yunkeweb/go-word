# 柱 / 条 / 折 / 饼 / 面积 / 双轴组合图

`AddChart` 写出 ECMA-376 图表部件（`word/charts/chartN.xml`）以及正文中的 DrawingML 图框。系列 XML 遵循 Word 图表 XSD 顺序（`idx` → `order` → `tx` → `spPr` → …），Word 打开时不会弹出修复对话框。

## AddChart

### 签名

```go
func (d *Document) AddChart(chartType string, categories []string, seriesData ...[]float64) *element.Chart
func (d *Document) AddChartStyled(chartType string, categories []string, values []float64, st style.Chart) *element.Chart
```

### 参数

| 名称 | 类型 | 说明 |
| --- | --- | --- |
| `chartType` | `string` | 见下方常量。 |
| `categories` | `[]string` | X 轴标签（`c:cat` 字符串缓存）。 |
| `seriesData` | `...[]float64` | 与 `categories` 对齐的一个或多个数值系列。额外系列名为 `Series 2`、`Series 3`… |

### 图表类型

| 常量 | Word 图表 |
| --- | --- |
| `ChartTypeBar` | 条形图（`c:barChart` barDir `bar`） |
| `ChartTypeColumn` | 柱形图（`c:barChart` barDir `col`） |
| `ChartTypeLine` | 折线图（`c:lineChart`） |
| `ChartTypePie` | 饼图（`c:pieChart`） |
| `ChartTypeArea` | 面积图（`c:areaChart`，grouping `standard`） |
| `ChartTypeStackedArea` | 堆叠面积图 |
| `ChartTypePercentStackedArea` | 百分比堆叠面积图 |
| `ChartTypeCombo` | 柱 + 额外系列（常为次坐标轴上的折线） |

## 图表辅助方法

### 签名

```go
func (c *Chart) AddSeries(categories []string, values []float64, name string)
func (c *Chart) AddComboSeries(kind string, categories []string, values []float64, name string, secondary bool)
func (c *Chart) SetLegendPosition(pos string)
func (c *Chart) SetMajorGridlines(visible bool)
func (c *Chart) SetLineSmooth(on bool)
func (c *Chart) SetLineMarker(symbol string)
func (c *Chart) SetDataLabels(opts style.DataLabelOptions)
```

| 辅助方法 | OpenXML |
| --- | --- |
| `AddComboSeries(..., true)` | 第二条 `c:valAx` 上的额外 `c:ser` |
| `SetLegendPosition` | `c:legend` / `c:legendPos`（`t`/`b`/`l`/`r`）。图例在 `plotArea` **之后**。 |
| `SetLineMarker` | `c:marker`，含 `symbol` + `size` 5 |
| `SetDataLabels` | `c:dLbls`。面积图省略 `dLblPos`。 |

图例常量：`LegendTop`、`LegendBottom`、`LegendLeft`、`LegendRight`。

演示中用到的 `Chart.Style` 字段：`Title`、`Width`、`Height`（EMU）、`ValueNumFmt`、`SecondaryValueNumFmt`。

## 完整示例

```go
package main

import (
	"log"

	"github.com/yunkeweb/go-word"
)

func main() {
	doc := word.New()
	doc.SetDefaultFontName("Calibri")
	sec := doc.AddSection()
	sec.AddTitle("DrawingML charts", 1)

	cats := []string{"Q1", "Q2", "Q3", "Q4"}
	doc.AddChart(word.ChartTypeBar, cats, []float64{12, 18, 9, 15})
	doc.AddChart(word.ChartTypeColumn, cats, []float64{12, 18, 9, 15})
	line := doc.AddChart(word.ChartTypeLine, cats, []float64{8, 11, 14, 10})
	line.SetLineSmooth(true)
	line.SetLineMarker("circle")
	doc.AddChart(word.ChartTypePie, []string{"A", "B", "C"}, []float64{40, 35, 25})
	doc.AddChart(word.ChartTypeArea, cats, []float64{4, 7, 6, 9})

	area := doc.AddChart(word.ChartTypeStackedArea, cats,
		[]float64{12, 18, 15, 22}, []float64{8, 11, 13, 16})
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

	months := []string{"Jan", "Feb", "Mar", "Apr", "May", "Jun"}
	combo := doc.AddChart(word.ChartTypeCombo, months, []float64{120, 150, 140, 180, 210, 190})
	combo.Series[0].Name = "Sales"
	combo.Series[0].Kind = word.ChartTypeColumn
	combo.AddComboSeries(word.ChartTypeLine, months,
		[]float64{8.2, 9.1, 8.7, 10.4, 11.2, 10.8}, "Margin %", true)
	combo.Style.Title = "Sales vs Margin"
	combo.Style.ValueNumFmt = "0"
	combo.Style.SecondaryValueNumFmt = "0.0"
	combo.SetLegendPosition(word.LegendRight)
	combo.SetMajorGridlines(true)
	combo.SetLineSmooth(true)
	combo.SetLineMarker("circle")

	if err := doc.Save("charts.docx"); err != nil {
		log.Fatal(err)
	}
}
```

### Word 中的效果

六张可点击进入图表设计器的原生图表。堆叠面积图把 Hardware 叠在 Software 上，数值标在带中央。组合图左侧轴画 Sales 柱，右侧轴画带圆点标记的平滑 Margin % 折线。没有嵌入工作簿 OLE；分类与数值存在图表缓存里。
