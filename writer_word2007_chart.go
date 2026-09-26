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
	"line":                   {typ: "line", axes: true, grouping: "standard"},
	"area":                   {typ: "area", axes: true, grouping: "standard"},
	"radar":                  {typ: "radar", axes: true, radar: "standard", no3d: true},
	"scatter":                {typ: "scatter", axes: true, scatter: "marker", no3d: true},
}

func chartPartXML(ch *element.Chart) []byte {
	xw := common.GetXMLWriter()
	writeChartPart(xw, ch)
	return common.FinishXML(xw)
}

func writeChartPart(xw *common.XMLWriter, ch *element.Chart) {
	xw.StartDocument()
	xw.Start("c:chartSpace",
		"xmlns:c", ooxml.NSC,
		"xmlns:a", ooxml.NSA,
		"xmlns:r", ooxml.NSR)
	writeChartBody(xw, ch)
	xw.Start("c:spPr")
	xw.Start("a:ln")
	xw.Empty("a:noFill")
	xw.End()
	xw.End()
	xw.End()
}

func writeChartBody(xw *common.XMLWriter, ch *element.Chart) {
	kind, ok := chartKinds[ch.ChartType]
	if !ok {
		kind = chartKinds["pie"]
	}
	st := ch.Style
	xw.Start("c:chart")
	if st.Title != "" {
		writeChartTitle(xw, st.Title, false)
	} else {
		xw.Empty("c:autoTitleDeleted", "val", "1")
	}
	xw.Start("c:plotArea")
	xw.Empty("c:layout")
	writeChartType(xw, ch, kind)
	if kind.axes {
		writeChartAxis(xw, "c:catAx", 1, "b", 2, st, true, kind)
		writeChartAxis(xw, "c:valAx", 2, "l", 1, st, false, kind)
	}
	xw.End() // c:plotArea
	if st.ShowLegend {
		pos := st.LegendPosition
		if pos == "" {
			pos = "r"
		}
		xw.Start("c:legend")
		xw.Empty("c:legendPos", "val", pos)
		xw.Empty("c:overlay", "val", "0")
		xw.End()
	}
	xw.Empty("c:plotVisOnly", "val", "1")
	xw.End() // c:chart
}

func writeChartTitle(xw *common.XMLWriter, title string, overlay bool) {
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
	if overlay {
		xw.Empty("c:overlay", "val", "0")
	}
	xw.End()
}

func writeChartType(xw *common.XMLWriter, ch *element.Chart, kind chartKind) {
	st := ch.Style
	chartType := kind.typ
	if st.ThreeD && !kind.no3d {
		chartType += "3D"
	}
	xw.Start("c:" + chartType + "Chart")
	if kind.bar != "" {
		xw.Empty("c:barDir", "val", kind.bar)
	}
	if kind.grouping != "" && kind.scatter == "" && kind.radar == "" && kind.typ != "pie" && kind.typ != "doughnut" {
		xw.Empty("c:grouping", "val", kind.grouping)
	}
	if kind.radar != "" {
		xw.Empty("c:radarStyle", "val", kind.radar)
	}
	if kind.scatter != "" {
		xw.Empty("c:scatterStyle", "val", kind.scatter)
	}
	if kind.colors {
		xw.Empty("c:varyColors", "val", "1")
	} else {
		xw.Empty("c:varyColors", "val", "0")
	}
	writeChartSeries(xw, ch, kind)
	if kind.hole > 0 {
		xw.Empty("c:holeSize", "val", itoa(kind.hole))
	}
	if kind.bar != "" && !st.ThreeD && kind.grouping != "" && kind.grouping != "clustered" {
		xw.Empty("c:overlap", "val", "100")
	}
	if kind.axes {
		xw.Empty("c:axId", "val", "1")
		xw.Empty("c:axId", "val", "2")
	}
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
			// CT_SerTx is strRef (requires c:f) or a literal c:v.
			xw.Start("c:tx")
			xw.Element("c:v", ser.Name)
			xw.End()
		}
		if kind.scatter == "" && len(colors) > 0 {
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
		xw.Start("c:dLbls")
		writeBoolVal(xw, "c:showLegendKey", labels.ShowLegendKey)
		writeBoolVal(xw, "c:showVal", labels.ShowVal)
		writeBoolVal(xw, "c:showCatName", labels.ShowCatName)
		writeBoolVal(xw, "c:showSerName", labels.ShowSerName)
		writeBoolVal(xw, "c:showPercent", labels.ShowPercent)
		writeBoolVal(xw, "c:showBubbleSize", labels.ShowBubbleSize)
		writeBoolVal(xw, "c:showLeaderLines", labels.ShowLeaderLines)
		xw.End()
		if kind.scatter != "" {
			writeChartSeriesItem(xw, "c:xVal", "c:strLit", stringSlice(ser.Categories))
			writeChartSeriesItem(xw, "c:yVal", "c:numLit", floatSlice(ser.Values))
		} else {
			writeChartSeriesItem(xw, "c:cat", "c:strLit", stringSlice(ser.Categories))
			writeChartSeriesItem(xw, "c:val", "c:numLit", floatSlice(ser.Values))
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
	if lit == "c:numLit" {
		xw.Element("c:formatCode", "General")
	}
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
	xw.Start("c:scaling")
	xw.Empty("c:orientation", "val", "minMax")
	xw.End()
	xw.Empty("c:delete", "val", "0")
	xw.Empty("c:axPos", "val", pos)
	if kind.radar != "" || (cat && st.ShowGridX) || (!cat && st.ShowGridY) {
		xw.Empty("c:majorGridlines")
	}
	title := st.ValueAxisTitle
	if cat {
		title = st.CategoryAxisTitle
	}
	if title != "" {
		writeChartTitle(xw, title, true)
	}
	tick := st.MajorTickPosition
	if tick == "" {
		tick = "out"
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
	xw.Start("c:spPr")
	xw.Start("a:ln", "w", "9525")
	xw.Start("a:solidFill")
	xw.Empty("a:srgbClr", "val", "000000")
	xw.End()
	xw.End()
	xw.End()
	xw.Empty("c:crossAx", "val", itoa(cross))
	xw.Empty("c:crosses", "val", "autoZero")
	if cat {
		xw.Empty("c:auto", "val", "1")
		xw.Empty("c:lblAlgn", "val", "ctr")
		xw.Empty("c:lblOffset", "val", "100")
	} else {
		xw.Empty("c:crossBetween", "val", "between")
	}
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
	docPrID := rid
	if len(rid) > 3 && rid[:3] == "rId" {
		docPrID = rid[3:]
	}
	xw := common.GetXMLWriter()
	if wrapP {
		xw.Start("w:p")
	}
	xw.Start("w:r")
	xw.Start("w:drawing")
	xw.Start("wp:inline", "distT", "0", "distB", "0", "distL", "0", "distR", "0")
	xw.Empty("wp:extent", "cx", itoa(cx), "cy", itoa(cy))
	xw.Empty("wp:effectExtent", "l", "0", "t", "0", "r", "0", "b", "0")
	xw.Empty("wp:docPr", "id", docPrID, "name", name)
	xw.Start("wp:cNvGraphicFramePr")
	xw.Empty("a:graphicFrameLocks", "xmlns:a", ooxml.NSA, "noChangeAspect", "1")
	xw.End()
	xw.Start("a:graphic", "xmlns:a", ooxml.NSA)
	xw.Start("a:graphicData", "uri", ooxml.NSC)
	xw.Empty("c:chart",
		"xmlns:c", ooxml.NSC,
		"xmlns:r", ooxml.NSR,
		"r:id", rid)
	xw.End()
	xw.End()
	xw.End()
	xw.End()
	xw.End()
	if wrapP {
		xw.End()
	}
	s := xw.String()
	common.PutXMLWriter(xw)
	return s
}
