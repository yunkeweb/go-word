# DrawingML 图表与形状

GoWord 写出 Word 2010 **wordprocessingShape**（`wps:wsp`）图形以及 ECMA-376 **chart** 部件（`word/charts/chartN.xml`）。

## 形状与文本框

```go
doc.AddShape(word.ShapeRect, word.ShapeOptions{
	FillColor: "5B9BD5", LineColor: "2E75B6",
	Width: 1828800, Height: 914400, // EMU（1 英寸 = 914400）
})
doc.AddShape(word.ShapeRoundRect, word.ShapeOptions{
	FillColor: "ED7D31", Text: "Rounded",
	Font: style.Font{Bold: true, Color: "FFFFFF"},
})
doc.AddShape(word.ShapeArrow, word.ShapeOptions{FillColor: "70AD47"})
doc.AddShape(word.ShapeTextBox, word.ShapeOptions{
	FillColor: "FFF2CC", LineColor: "BF8F00",
	Text: "w:txbxContent", Font: style.Font{Size: 12},
})
```

| `ShapeType` | 预设几何 |
| --- | --- |
| `ShapeRect` | `rect` |
| `ShapeRoundRect` | `roundRect` |
| `ShapeArrow` | `rightArrow` |
| `ShapeTextBox` | `rect` + `wps:txbx` / `w:txbxContent` |

`ShapeOptions` 控制填充色、边框颜色/线宽、EMU 尺寸与内嵌文字。

## 图表

```go
cats := []string{"Q1", "Q2", "Q3", "Q4"}
doc.AddChart(word.ChartTypeColumn, cats, []float64{12, 18, 9, 15})
doc.AddChart(word.ChartTypeLine, cats, []float64{8, 11, 14, 10})
doc.AddChart(word.ChartTypePie, []string{"A", "B", "C"}, []float64{40, 35, 25})
doc.AddChart(word.ChartTypeArea, cats, []float64{4, 7, 6, 9})
doc.AddChart(word.ChartTypeStackedArea, cats, []float64{3, 5, 4, 6}, []float64{2, 2, 3, 1})
```

常量：`ChartTypeBar`、`ChartTypeColumn`、`ChartTypeLine`、`ChartTypePie`、`ChartTypeArea`、`ChartTypeStackedArea`、`ChartTypePercentStackedArea`、`ChartTypeCombo`。

### 组合图（双轴）

```go
ch := doc.AddChart(word.ChartTypeCombo, cats, []float64{12, 18, 9, 15})
ch.AddComboSeries("line", cats, []float64{8, 11, 14, 10}, "Trend", true)
ch.SetLegendPosition(word.LegendBottom)
ch.SetMajorGridlines(true)
ch.SetLineSmooth(true)
ch.SetLineMarker("circle")
```

次轴系列使用第二条数值轴。系列 XML 遵循 Word 图表 XSD 顺序（`idx` → `order` → `tx` → `spPr` → …），Word 打开时不会弹出修复对话框。

示例：[`examples/v0.7.0_demo`](https://github.com/yunkeweb/go-word/tree/main/examples/v0.7.0_demo)、[`examples/v0.8.0_demo`](https://github.com/yunkeweb/go-word/tree/main/examples/v0.8.0_demo)。
