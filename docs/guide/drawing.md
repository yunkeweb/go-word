# DrawingML & Charts

GoWord writes Word 2010 **wordprocessingShape** (`wps:wsp`) drawings and ECMA-376 **chart** parts (`word/charts/chartN.xml`).

## Shapes and text boxes

```go
doc.AddShape(word.ShapeRect, word.ShapeOptions{
	FillColor: "5B9BD5", LineColor: "2E75B6",
	Width: 1828800, Height: 914400, // EMU (1 inch = 914400)
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

| `ShapeType` | Preset geometry |
| --- | --- |
| `ShapeRect` | `rect` |
| `ShapeRoundRect` | `roundRect` |
| `ShapeArrow` | `rightArrow` |
| `ShapeTextBox` | `rect` + `wps:txbx` / `w:txbxContent` |

`ShapeOptions` controls fill color, line color/width, size in EMU, and inner text.

## Charts

```go
cats := []string{"Q1", "Q2", "Q3", "Q4"}
doc.AddChart(word.ChartTypeColumn, cats, []float64{12, 18, 9, 15})
doc.AddChart(word.ChartTypeLine, cats, []float64{8, 11, 14, 10})
doc.AddChart(word.ChartTypePie, []string{"A", "B", "C"}, []float64{40, 35, 25})
doc.AddChart(word.ChartTypeArea, cats, []float64{4, 7, 6, 9})
doc.AddChart(word.ChartTypeStackedArea, cats, []float64{3, 5, 4, 6}, []float64{2, 2, 3, 1})
```

Constants: `ChartTypeBar`, `ChartTypeColumn`, `ChartTypeLine`, `ChartTypePie`, `ChartTypeArea`, `ChartTypeStackedArea`, `ChartTypePercentStackedArea`, `ChartTypeCombo`.

### Combo (dual axis)

```go
ch := doc.AddChart(word.ChartTypeCombo, cats, []float64{12, 18, 9, 15})
ch.AddComboSeries("line", cats, []float64{8, 11, 14, 10}, "Trend", true)
ch.SetLegendPosition(word.LegendBottom)
ch.SetMajorGridlines(true)
ch.SetLineSmooth(true)
ch.SetLineMarker("circle")
```

The secondary series uses a second value axis. Series XML follows the Word chart XSD order (`idx` → `order` → `tx` → `spPr` → …) so Word opens the file without a repair dialog.

See [`examples/v0.7.0_demo`](https://github.com/yunkeweb/go-word/tree/main/examples/v0.7.0_demo) and [`examples/v0.8.0_demo`](https://github.com/yunkeweb/go-word/tree/main/examples/v0.8.0_demo).
