package word

import (
	"strconv"

	"github.com/yunkeweb/go-word/element"
	"github.com/yunkeweb/go-word/style"
)

// Chart type names used by AddChart and Container.AddChart.
const (
	ChartTypeBar                = "bar"
	ChartTypeColumn             = "column"
	ChartTypeLine               = "line"
	ChartTypePie                = "pie"
	ChartTypeArea               = "area"
	ChartTypeStackedArea        = "stacked_area"
	ChartTypePercentStackedArea = "percent_stacked_area"
	ChartTypeCombo              = "combo"
)

// Legend positions (c:legendPos).
const (
	LegendTop    = "t"
	LegendBottom = "b"
	LegendLeft   = "l"
	LegendRight  = "r"
)

// Data label positions (c:dLblPos).
const (
	DataLabelPosBestFit = "bestFit"
	DataLabelPosBottom  = "b"
	DataLabelPosCenter  = "ctr"
	DataLabelPosInBase  = "inBase"
	DataLabelPosInEnd   = "inEnd"
	DataLabelPosLeft    = "l"
	DataLabelPosOutEnd  = "outEnd"
	DataLabelPosRight   = "r"
	DataLabelPosTop     = "t"
)

// ChartDataLabelOptions configures c:dLbls on a chart.
type ChartDataLabelOptions = style.DataLabelOptions

// AddChart appends a native OpenXML chart to the last (or a new) section.
// seriesData is one or more value series aligned with categories.
func (d *Document) AddChart(chartType string, categories []string, seriesData ...[]float64) *element.Chart {
	sec := d.lastOrNewSection()
	var first []float64
	if len(seriesData) > 0 {
		first = seriesData[0]
	}
	ch := sec.AddChart(chartType, categories, first)
	for i := 1; i < len(seriesData); i++ {
		ch.AddSeries(categories, seriesData[i], "Series "+strconv.Itoa(i+1))
	}
	return ch
}

// AddChartStyled is AddChart with DrawingML chart style options.
func (d *Document) AddChartStyled(chartType string, categories []string, values []float64, st style.Chart) *element.Chart {
	return d.lastOrNewSection().AddChart(chartType, categories, values, st)
}
