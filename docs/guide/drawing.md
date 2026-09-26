# DrawingML Charts & Shapes

GoWord writes Word 2010 **wordprocessingShape** (`wps:wsp`) drawings and ECMA-376 **chart** parts (`word/charts/chartN.xml`).

## Shapes and text boxes

| `ShapeType` | Preset geometry |
| --- | --- |
| `ShapeRect` | `rect` |
| `ShapeRoundRect` | `roundRect` |
| `ShapeArrow` | `rightArrow` |
| `ShapeTextBox` | `rect` + `wps:txbx` / `w:txbxContent` |

`ShapeOptions` controls fill color, line color/width, size in EMU (1 inch = 914400), and inner text. Empty text on `ShapeTextBox` is replaced with a space so Word still creates `w:txbxContent`.

## Charts

Constants: `ChartTypeBar`, `ChartTypeColumn`, `ChartTypeLine`, `ChartTypePie`, `ChartTypeArea`, `ChartTypeStackedArea`, `ChartTypePercentStackedArea`, `ChartTypeCombo`.

Series XML follows the Word chart XSD order (`idx` → `order` → `tx` → `spPr` → …) so Word opens the file without a repair dialog. Combo secondary series use a second value axis (`c:valAx` with a distinct `c:axId`).

Helpers on `*element.Chart`:

| Method | OpenXML |
| --- | --- |
| `AddComboSeries(kind, cats, vals, name, secondary)` | Extra `c:ser` on a combo chart |
| `SetLegendPosition` | `c:legend` / `c:legendPos` (`t`/`b`/`l`/`r`) |
| `SetMajorGridlines` | `c:majorGridlines` |
| `SetLineSmooth` | `c:smooth` on line series |
| `SetLineMarker` | `c:marker` (`symbol` + `size` 5) |
| `SetDataLabels` | `c:dLbls` (area charts omit `dLblPos`) |

## Complete example

Save as `main.go` and run `go run .`. The file `drawing.docx` contains every chart type plus four shapes.

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

See [`examples/v0.7.0_demo`](https://github.com/yunkeweb/go-word/tree/main/examples/v0.7.0_demo) and [`examples/v0.8.0_demo`](https://github.com/yunkeweb/go-word/tree/main/examples/v0.8.0_demo).
