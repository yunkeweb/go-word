package word

import (
	"strconv"

	"github.com/yunkeweb/go-word/element"
	"github.com/yunkeweb/go-word/ooxml"
	"github.com/yunkeweb/go-word/pkg/common"
	"github.com/yunkeweb/go-word/style"
)

type chartKind struct {
	typ      string
	colors   bool
	hole     int
	no3d     bool
	axes     bool
	bar      string
	grouping string
	radar    string
	scatter  string
}

var chartKinds = map[string]chartKind{
	"pie":                    {typ: "pie", colors: true},
	"doughnut":               {typ: "doughnut", colors: true, hole: 75, no3d: true},
	"bar":                    {typ: "bar", axes: true, bar: "bar", grouping: "clustered"},
	"stacked_bar":            {typ: "bar", axes: true, bar: "bar", grouping: "stacked"},
	"percent_stacked_bar":    {typ: "bar", axes: true, bar: "bar", grouping: "percentStacked"},
	"column":                 {typ: "bar", axes: true, bar: "col", grouping: "clustered"},
	"stacked_column":         {typ: "bar", axes: true, bar: "col", grouping: "stacked"},
	"percent_stacked_column": {typ: "bar", axes: true, bar: "col", grouping: "percentStacked"},
	"line":                   {typ: "line", axes: true},
	"area":                   {typ: "area", axes: true},
	"radar":                  {typ: "radar", axes: true, radar: "standard", no3d: true},
	"scatter":                {typ: "scatter", axes: true, scatter: "marker", no3d: true},
}

func chartPartXML(ch *element.Chart) []byte {
	xw := common.NewXMLWriter()
	xw.StartDocument()
	xw.Start("c:chartSpace",
		"xmlns:c", "http://schemas.openxmlformats.org/drawingml/2006/chart",
		"xmlns:a", ooxml.NSA,
		"xmlns:r", ooxml.NSR)
	writeChartBody(xw, ch)
	xw.Start("c:spPr")
	xw.Start("a:ln")
	xw.Empty("a:noFill")
	xw.End()
	xw.End()
	xw.End()
	return xw.Bytes()
}

func writeChartBody(xw *common.XMLWriter, ch *element.Chart) {
	kind, ok := chartKinds[ch.ChartType]
	if !ok {
		kind = chartKinds["pie"]
	}
	st := ch.Style
	xw.Start("c:chart")
	if st.Title != "" {
		xw.Start("c:title")
		xw.Start("c:tx")
		xw.Start("c:rich")
		xw.Empty("a:bodyPr")
		xw.Empty("a:lstStyle")
		xw.Start("a:p")
		xw.Start("a:pPr")
		xw.Empty("a:defRPr")
		xw.End()
		xw.Start("a:r")
		xw.Empty("a:rPr")
		xw.Element("a:t", st.Title)
		xw.End()
		xw.Empty("a:endParaRPr")
		xw.End()
		xw.End()
		xw.End()
		xw.End()
	} else {
		xw.Empty("c:autoTitleDeleted", "val", "1")
	}
	if st.ShowLegend {
		pos := st.LegendPosition
		if pos == "" {
			pos = "r"
		}
		xw.Start("c:legend")
		xw.Empty("c:legendPos", "val", pos)
		xw.End()
	}
	xw.Start("c:plotArea")
	xw.Empty("c:layout")
	chartType := kind.typ
	if st.ThreeD && !kind.no3d {
		chartType += "3D"
	}
	xw.Start("c:" + chartType + "Chart")
	if kind.colors {
		xw.Empty("c:varyColors", "val", "1")
	} else {
		xw.Empty("c:varyColors", "val", "0")
	}
	if ch.ChartType == "area" {
		xw.Empty("c:grouping", "val", "standard")
	}
	if kind.hole > 0 {
		xw.Empty("c:holeSize", "val", itoa(kind.hole))
	}
	if kind.bar != "" {
		xw.Empty("c:barDir", "val", kind.bar)
		xw.Empty("c:grouping", "val", kind.grouping)
	}
	if kind.radar != "" {
		xw.Empty("c:radarStyle", "val", kind.radar)
	}
	if kind.scatter != "" {
		xw.Empty("c:scatterStyle", "val", kind.scatter)
	}
	writeChartSeries(xw, ch, kind)
	if kind.grouping != "clustered" {
		xw.Empty("c:overlap", "val", "100")
	}
	if kind.axes {
		xw.Empty("c:axId", "val", "1")
		xw.Empty("c:axId", "val", "2")
	}
	xw.End()
	if kind.axes {
		writeChartAxis(xw, "c:catAx", 1, "b", 2, st, true, kind)
		writeChartAxis(xw, "c:valAx", 2, "l", 1, st, false, kind)
	}
	xw.End()
	xw.End()
}

func writeChartSeries(xw *common.XMLWriter, ch *element.Chart, kind chartKind) {
	series := ch.Series
	if len(series) == 0 && (len(ch.Categories) > 0 || len(ch.Values) > 0) {
		series = []element.ChartSeries{{
			Categories: ch.Categories,
			Values:     ch.Values,
			Name:       ch.SeriesName,
		}}
	}
	st := ch.Style
	labels := st.DataLabels
	if !st.DataLabelsSet {
		labels = style.DefaultDataLabelOptions()
	}
	colors := st.Colors
	colorIdx := 0
	for i, ser := range series {
		xw.Start("c:ser")
		xw.Empty("c:idx", "val", itoa(i))
		xw.Empty("c:order", "val", itoa(i))
		if ser.Name != "" {
			xw.Start("c:tx")
			xw.Start("c:strRef")
			xw.Start("c:strCache")
			xw.Empty("c:ptCount", "val", "1")
			xw.Start("c:pt", "idx", "0")
			xw.Element("c:v", ser.Name)
			xw.End()
			xw.End()
			xw.End()
			xw.End()
		}
		xw.Start("c:dLbls")
		writeBoolVal(xw, "c:showVal", labels.ShowVal)
		writeBoolVal(xw, "c:showCatName", labels.ShowCatName)
		writeBoolVal(xw, "c:showLegendKey", labels.ShowLegendKey)
		writeBoolVal(xw, "c:showSerName", labels.ShowSerName)
		writeBoolVal(xw, "c:showPercent", labels.ShowPercent)
		writeBoolVal(xw, "c:showLeaderLines", labels.ShowLeaderLines)
		writeBoolVal(xw, "c:showBubbleSize", labels.ShowBubbleSize)
		xw.End()
		if kind.scatter != "" {
			writeChartSeriesItem(xw, "c:xVal", "c:strLit", stringSlice(ser.Categories))
			writeChartSeriesItem(xw, "c:yVal", "c:numLit", floatSlice(ser.Values))
		} else {
			writeChartSeriesItem(xw, "c:cat", "c:strLit", stringSlice(ser.Categories))
			writeChartSeriesItem(xw, "c:val", "c:numLit", floatSlice(ser.Values))
			if len(colors) > 0 {
				for vi := range ser.Values {
					xw.Start("c:dPt")
					xw.Empty("c:idx", "val", itoa(vi))
					xw.Start("c:spPr")
					xw.Start("a:solidFill")
					xw.Empty("a:srgbClr", "val", colors[colorIdx%len(colors)])
					xw.End()
					xw.End()
					xw.End()
					colorIdx++
				}
			}
		}
		xw.End()
	}
}

func writeBoolVal(xw *common.XMLWriter, name string, v bool) {
	if v {
		xw.Empty(name, "val", "1")
	} else {
		xw.Empty(name, "val", "0")
	}
}

func stringSlice(v []string) []string { return v }

func floatSlice(v []float64) []string {
	out := make([]string, len(v))
	for i, n := range v {
		out[i] = strconv.FormatFloat(n, 'f', -1, 64)
	}
	return out
}

func writeChartSeriesItem(xw *common.XMLWriter, outer, lit string, values []string) {
	xw.Start(outer)
	xw.Start(lit)
	xw.Empty("c:ptCount", "val", itoa(len(values)))
	for i, v := range values {
		xw.Start("c:pt", "idx", itoa(i))
		xw.Element("c:v", v)
		xw.End()
	}
	xw.End()
	xw.End()
}

func writeChartAxis(xw *common.XMLWriter, tag string, id int, pos string, cross int, st style.Chart, cat bool, kind chartKind) {
	xw.Start(tag)
	xw.Empty("c:axId", "val", itoa(id))
	xw.Empty("c:axPos", "val", pos)
	title := st.ValueAxisTitle
	if cat {
		title = st.CategoryAxisTitle
	}
	if title != "" {
		xw.Start("c:title")
		xw.Start("c:tx")
		xw.Start("c:rich")
		xw.Empty("a:bodyPr")
		xw.Empty("a:lstStyle")
		xw.Start("a:p")
		xw.Start("a:pPr")
		xw.Empty("a:defRPr")
		xw.End()
		xw.Start("a:r")
		xw.Empty("a:rPr", "lang", "en-US")
		xw.Element("a:t", title)
		xw.End()
		xw.End()
		xw.End()
		xw.End()
		xw.Empty("c:overlay", "val", "0")
		xw.End()
	}
	xw.Empty("c:crossAx", "val", itoa(cross))
	xw.Empty("c:auto", "val", "1")
	xw.Empty("c:delete", "val", "0")
	tick := st.MajorTickPosition
	if tick == "" {
		tick = "none"
	}
	xw.Empty("c:majorTickMark", "val", tick)
	xw.Empty("c:minorTickMark", "val", "none")
	if st.ShowAxisLabels {
		lbl := st.ValueLabelPosition
		if cat {
			lbl = st.CategoryLabelPosition
		}
		if lbl == "" {
			lbl = "nextTo"
		}
		xw.Empty("c:tickLblPos", "val", lbl)
	} else {
		xw.Empty("c:tickLblPos", "val", "none")
	}
	xw.Empty("c:crosses", "val", "autoZero")
	if kind.radar != "" || (cat && st.ShowGridX) || (!cat && st.ShowGridY) {
		xw.Empty("c:majorGridlines")
	}
	xw.Start("c:scaling")
	xw.Empty("c:orientation", "val", "minMax")
	xw.End()
	xw.Start("c:spPr")
	xw.Start("a:ln")
	xw.Empty("a:solidFill")
	xw.End()
	xw.End()
	xw.End()
}

func chartDrawingXML(rid string, ch *element.Chart, wrapP bool) string {
	st := ch.Style
	cx := st.Width
	cy := st.Height
	if cx == 0 {
		cx = 1000000
	}
	if cy == 0 {
		cy = 1000000
	}
	name := "Chart" + rid
	xw := common.NewXMLWriter()
	if wrapP {
		xw.Start("w:p")
	}
	xw.Start("w:r")
	xw.Start("w:drawing")
	xw.Start("wp:inline")
	xw.Empty("wp:extent", "cx", itoa(cx), "cy", itoa(cy))
	xw.Empty("wp:docPr", "id", rid[3:], "name", name)
	xw.Start("a:graphic", "xmlns:a", ooxml.NSA)
	xw.Start("a:graphicData", "uri", "http://schemas.openxmlformats.org/drawingml/2006/chart")
	xw.Empty("c:chart",
		"r:id", rid,
		"xmlns:c", "http://schemas.openxmlformats.org/drawingml/2006/chart",
		"xmlns:r", ooxml.NSR)
	xw.End()
	xw.End()
	xw.End()
	xw.End()
	xw.End()
	if wrapP {
		xw.End()
	}
	return xw.String()
}
