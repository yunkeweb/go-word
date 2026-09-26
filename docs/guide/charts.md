# Charts (Bar / Line / Pie / Area / Combo)

`AddChart` writes an ECMA-376 chart part (`word/charts/chartN.xml`) and a DrawingML frame in the body. Series XML follows the Word chart XSD order (`idx` → `order` → `tx` → `spPr` → …) so Word opens the file without a repair dialog.

## AddChart

### Signature

```go
func (d *Document) AddChart(chartType string, categories []string, seriesData ...[]float64) *element.Chart
func (d *Document) AddChartStyled(chartType string, categories []string, values []float64, st style.Chart) *element.Chart
```

### Parameters

| Name | Type | Description |
| --- | --- | --- |
| `chartType` | `string` | See constants below. |
| `categories` | `[]string` | X-axis labels (`c:cat` string cache). |
| `seriesData` | `...[]float64` | One or more value series aligned with `categories`. Extra series become `Series 2`, `Series 3`, … |

### Chart types

| Constant | Word chart |
| --- | --- |
| `ChartTypeBar` | Bar (`c:barChart` barDir `bar`) |
| `ChartTypeColumn` | Column (`c:barChart` barDir `col`) |
| `ChartTypeLine` | Line (`c:lineChart`) |
| `ChartTypePie` | Pie (`c:pieChart`) |
| `ChartTypeArea` | Area (`c:areaChart`, grouping `standard`) |
| `ChartTypeStackedArea` | Stacked area |
| `ChartTypePercentStackedArea` | 100% stacked area |
| `ChartTypeCombo` | Column + extra series (often a line on a secondary axis) |

## Chart helpers

### Signature

```go
func (c *Chart) AddSeries(categories []string, values []float64, name string)
func (c *Chart) AddComboSeries(kind string, categories []string, values []float64, name string, secondary bool)
func (c *Chart) SetLegendPosition(pos string)
func (c *Chart) SetMajorGridlines(visible bool)
func (c *Chart) SetLineSmooth(on bool)
func (c *Chart) SetLineMarker(symbol string)
func (c *Chart) SetDataLabels(opts style.DataLabelOptions)
```

| Helper | OpenXML |
| --- | --- |
| `AddComboSeries(..., true)` | Extra `c:ser` on a second `c:valAx` |
| `SetLegendPosition` | `c:legend` / `c:legendPos` (`t`/`b`/`l`/`r`). Legend is **after** `plotArea`. |
| `SetLineMarker` | `c:marker` with `symbol` + `size` 5 |
| `SetDataLabels` | `c:dLbls`. Area charts omit `dLblPos`. |

Legend constants: `LegendTop`, `LegendBottom`, `LegendLeft`, `LegendRight`.

`Chart.Style` fields used in demos: `Title`, `Width`, `Height` (EMU), `ValueNumFmt`, `SecondaryValueNumFmt`.

## Complete example

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

### What Word shows

Six native charts you can click to edit in the Word chart designer. The stacked area stacks Hardware on Software with values in the center of each band. The combo chart plots Sales as columns on the left axis and Margin % as a smoothed line with circle markers on the right axis. There is no embedded workbook OLE; categories and values live in the chart caches.
