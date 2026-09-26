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

var defaultChartColors = []string{"5B9BD5", "ED7D31", "A9D08E", "FFC000", "4472C4", "70AD47"}

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
	"stacked_area":           {typ: "area", axes: true, grouping: "stacked"},
	"percent_stacked_area":   {typ: "area", axes: true, grouping: "percentStacked"},
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
	st := ch.Style
	xw.Start("c:chart")
	if st.Title != "" {
		writeChartTitle(xw, st.Title, false)
	} else {
		xw.Empty("c:autoTitleDeleted", "val", "1")
	}
	xw.Start("c:plotArea")
	xw.Empty("c:layout")
	if isComboChart(ch) {
		writeComboPlotArea(xw, ch)
	} else {
		kind, ok := chartKinds[ch.ChartType]
		if !ok {
			kind = chartKinds["pie"]
		}
		writeChartType(xw, ch, kind)
		if kind.axes {
			writeChartAxis(xw, "c:catAx", 1, "b", 2, st, true, kind, "")
			writeChartAxis(xw, "c:valAx", 2, "l", 1, st, false, kind, "")
		}
	}
	xw.End() // c:plotArea
	if st.ShowLegend {
		writeChartLegend(xw, st.LegendPosition)
	}
	xw.Empty("c:plotVisOnly", "val", "1")
	xw.End() // c:chart
}

// writeChartLegend emits CT_Legend as a sibling of plotArea (never inside it):
// legendPos -> layout -> overlay.
func writeChartLegend(xw *common.XMLWriter, pos string) {
	if pos == "" {
		pos = "r"
	}
	xw.Start("c:legend")
	xw.Empty("c:legendPos", "val", pos)
	xw.Empty("c:layout")
	xw.Empty("c:overlay", "val", "0")
	xw.End()
}

func isComboChart(ch *element.Chart) bool {
	if ch == nil {
		return false
	}
	if ch.ChartType == ChartTypeCombo {
		return true
	}
	seen := ""
	for _, s := range chartAllSeries(ch) {
		if s.SecondaryAxis {
			return true
		}
		k := s.Kind
		if k == "" {
			continue
		}
		if seen == "" {
			seen = k
			continue
		}
		if k != seen {
			return true
		}
	}
	return false
}

func chartAllSeries(ch *element.Chart) []element.ChartSeries {
	if len(ch.Series) > 0 {
		return ch.Series
	}
	if len(ch.Categories) > 0 || len(ch.Values) > 0 {
		return []element.ChartSeries{{
			Categories: ch.Categories,
			Values:     ch.Values,
			Name:       ch.SeriesName,
		}}
	}
	return nil
}

func seriesKindName(ch *element.Chart, s element.ChartSeries) string {
	if s.Kind != "" {
		return s.Kind
	}
	if ch.ChartType == ChartTypeCombo {
		if s.SecondaryAxis {
			return ChartTypeLine
		}
		return ChartTypeColumn
	}
	if ch.ChartType != "" {
		return ch.ChartType
	}
	return ChartTypeColumn
}

func writeComboPlotArea(xw *common.XMLWriter, ch *element.Chart) {
	all := chartAllSeries(ch)
	hasSecondary := false
	for _, s := range all {
		if s.SecondaryAxis {
			hasSecondary = true
			break
		}
	}
	type grouped struct {
		kind      chartKind
		kindName  string
		secondary bool
		series    []element.ChartSeries
		idxs      []int
	}
	var groups []grouped
	indexOf := func(kindName string, secondary bool) int {
		for i := range groups {
			if groups[i].kindName == kindName && groups[i].secondary == secondary {
				return i
			}
		}
		return -1
	}
	for i, s := range all {
		kn := seriesKindName(ch, s)
		kind, ok := chartKinds[kn]
		if !ok {
			kind = chartKinds["column"]
			kn = "column"
		}
		g := indexOf(kn, s.SecondaryAxis)
		if g < 0 {
			groups = append(groups, grouped{kind: kind, kindName: kn, secondary: s.SecondaryAxis})
			g = len(groups) - 1
		}
		groups[g].series = append(groups[g].series, s)
		groups[g].idxs = append(groups[g].idxs, i)
	}
	order := []string{"column", "bar", "stacked_bar", "percent_stacked_bar", "stacked_column", "percent_stacked_column", "line", "area", "stacked_area", "percent_stacked_area"}
	writeGroup := func(g grouped) {
		axVal := 2
		if g.secondary {
			axVal = 3
		}
		writeChartTypeSeries(xw, ch, g.kind, g.series, g.idxs, 1, axVal)
	}
	for _, name := range order {
		for _, g := range groups {
			if g.kindName == name && !g.secondary {
				writeGroup(g)
			}
		}
		for _, g := range groups {
			if g.kindName == name && g.secondary {
				writeGroup(g)
			}
		}
	}
	st := ch.Style
	kind := chartKinds["column"]
	writeChartAxis(xw, "c:catAx", 1, "b", 2, st, true, kind, "")
	writeChartAxis(xw, "c:valAx", 2, "l", 1, st, false, kind, "")
	if hasSecondary {
		sec := st
		sec.ShowGridX = false
		sec.ShowGridY = false
		writeChartAxis(xw, "c:valAx", 3, "r", 1, sec, false, kind, "max")
	}
}

func writeChartTitle(xw *common.XMLWriter, title string, overlay bool) {
	_ = overlay
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

func writeChartType(xw *common.XMLWriter, ch *element.Chart, kind chartKind) {
	series := chartAllSeries(ch)
	idxs := make([]int, len(series))
	for i := range series {
		idxs[i] = i
	}
	writeChartTypeSeries(xw, ch, kind, series, idxs, 1, 2)
}

func writeChartTypeSeries(xw *common.XMLWriter, ch *element.Chart, kind chartKind, series []element.ChartSeries, idxs []int, axCat, axVal int) {
	st := ch.Style
	chartType := kind.typ
	if st.ThreeD && !kind.no3d {
		chartType += "3D"
	}
	xw.Start("c:" + chartType + "Chart")
	if kind.bar != "" {
		xw.Empty("c:barDir", "val", kind.bar)
	}
	if kind.typ == "area" {
		xw.Empty("c:grouping", "val", areaGrouping(kind.grouping))
	} else if kind.grouping != "" && kind.scatter == "" && kind.radar == "" && kind.typ != "pie" && kind.typ != "doughnut" {
		xw.Empty("c:grouping", "val", kind.grouping)
	}
	if kind.radar != "" {
		xw.Empty("c:radarStyle", "val", kind.radar)
	}
	if kind.scatter != "" {
		xw.Empty("c:scatterStyle", "val", kind.scatter)
	}
	if kind.typ != "area" {
		if kind.colors {
			xw.Empty("c:varyColors", "val", "1")
		} else {
			xw.Empty("c:varyColors", "val", "0")
		}
	}
	writeChartSeriesList(xw, ch, kind, series, idxs)
	writeChartDLbls(xw, ch, kind)
	if kind.hole > 0 {
		xw.Empty("c:holeSize", "val", itoa(kind.hole))
	}
	if kind.bar != "" && !st.ThreeD && kind.grouping != "" && kind.grouping != "clustered" {
		xw.Empty("c:overlap", "val", "100")
	}
	if kind.axes {
		xw.Empty("c:axId", "val", itoa(axCat))
		xw.Empty("c:axId", "val", itoa(axVal))
	}
	xw.End()
}

func writeChartSeries(xw *common.XMLWriter, ch *element.Chart, kind chartKind) {
	series := chartAllSeries(ch)
	idxs := make([]int, len(series))
	for i := range series {
		idxs[i] = i
	}
	writeChartSeriesList(xw, ch, kind, series, idxs)
}

func dataLabelsActive(labels style.DataLabelOptions) bool {
	return labels.ShowVal || labels.ShowCatName || labels.ShowLegendKey ||
		labels.ShowSerName || labels.ShowPercent || labels.ShowBubbleSize ||
		labels.ShowLeaderLines || labels.Position != ""
}

// dLblPosForKind returns a Word-safe ST_DLblPos for the chart type.
// Clustered bar/column only accept ctr/inBase/inEnd/outEnd; t/b map to outEnd.
func dLblPosForKind(kind chartKind, pos string) string {
	if pos == "" {
		return ""
	}
	switch kind.typ {
	case "bar":
		switch pos {
		case "ctr", "inBase", "inEnd", "outEnd":
			return pos
		case "t", "b":
			return "outEnd"
		default:
			return ""
		}
	case "pie", "doughnut":
		switch pos {
		case "ctr", "inEnd", "outEnd", "bestFit":
			return pos
		default:
			return ""
		}
	case "area":
		// Area charts reject c:dLblPos in Word's strict schema; omit it.
		return ""
	case "line", "radar", "scatter":
		switch pos {
		case "ctr", "l", "r", "t", "b":
			return pos
		default:
			return ""
		}
	default:
		return pos
	}
}

func writeChartDLbls(xw *common.XMLWriter, ch *element.Chart, kind chartKind) {
	if ch == nil || !ch.Style.DataLabelsSet {
		return
	}
	labels := ch.Style.DataLabels
	if !dataLabelsActive(labels) {
		return
	}
	xw.Start("c:dLbls")
	if kind.typ != "area" {
		if pos := dLblPosForKind(kind, labels.Position); pos != "" {
			xw.Empty("c:dLblPos", "val", pos)
		}
	}
	writeBoolVal(xw, "c:showLegendKey", labels.ShowLegendKey)
	writeBoolVal(xw, "c:showVal", labels.ShowVal)
	writeBoolVal(xw, "c:showCatName", labels.ShowCatName)
	writeBoolVal(xw, "c:showSerName", labels.ShowSerName)
	writeBoolVal(xw, "c:showPercent", labels.ShowPercent)
	if kind.typ != "area" {
		writeBoolVal(xw, "c:showBubbleSize", labels.ShowBubbleSize)
		writeBoolVal(xw, "c:showLeaderLines", labels.ShowLeaderLines)
	}
	xw.End()
}

func areaGrouping(v string) string {
	switch v {
	case "standard", "stacked", "percentStacked":
		return v
	default:
		return "standard"
	}
}

func writeLineMarker(xw *common.XMLWriter, ser element.ChartSeries, st style.Chart) {
	symbol := ser.Marker
	if symbol == "" {
		symbol = st.LineMarker
	}
	if symbol == "" {
		return
	}
	switch symbol {
	case "circle", "dash", "diamond", "dot", "none", "plus", "square", "star", "triangle", "x":
	default:
		symbol = "circle"
	}
	xw.Start("c:marker")
	xw.Empty("c:symbol", "val", symbol)
	if symbol != "none" {
		xw.Empty("c:size", "val", "5")
	}
	xw.End()
}

func lineMarkerSymbol(ser element.ChartSeries, st style.Chart) string {
	symbol := ser.Marker
	if symbol == "" {
		symbol = st.LineMarker
	}
	return symbol
}

func writeLineSerSpPr(xw *common.XMLWriter, color string) {
	if color == "" {
		color = defaultChartColors[0]
	}
	xw.Start("c:spPr")
	xw.Start("a:ln", "w", "25400")
	xw.Start("a:solidFill")
	xw.Empty("a:srgbClr", "val", color)
	xw.End()
	xw.End()
	xw.End()
}

func writeSerSpPr(xw *common.XMLWriter, color string) {
	if color == "" {
		color = defaultChartColors[0]
	}
	xw.Start("c:spPr")
	xw.Start("a:solidFill")
	xw.Empty("a:srgbClr", "val", color)
	xw.End()
	xw.Start("a:ln")
	xw.Empty("a:noFill")
	xw.End()
	xw.End()
}

func writeChartSeriesList(xw *common.XMLWriter, ch *element.Chart, kind chartKind, series []element.ChartSeries, idxs []int) {
	st := ch.Style
	colors := st.Colors
	colorIdx := 0
	for n, ser := range series {
		i := n
		if n < len(idxs) {
			i = idxs[n]
		}
		xw.Start("c:ser")
		xw.Empty("c:idx", "val", itoa(i))
		xw.Empty("c:order", "val", itoa(i))
		if ser.Name != "" {
			// CT_SerTx is strRef (requires c:f) or a literal c:v.
			xw.Start("c:tx")
			xw.Element("c:v", ser.Name)
			xw.End()
		}
		if kind.typ == "area" {
			color := defaultChartColors[n%len(defaultChartColors)]
			if len(colors) > 0 {
				color = colors[colorIdx%len(colors)]
				colorIdx++
			}
			writeSerSpPr(xw, color)
		} else if kind.typ == "line" {
			if lineMarkerSymbol(ser, st) != "" {
				color := defaultChartColors[n%len(defaultChartColors)]
				if len(colors) > 0 {
					color = colors[colorIdx%len(colors)]
					colorIdx++
				}
				writeLineSerSpPr(xw, color)
				writeLineMarker(xw, ser, st)
			}
		} else if kind.scatter == "" && len(colors) > 0 {
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
		if kind.scatter != "" {
			writeChartSeriesItem(xw, "c:xVal", "c:strLit", stringSlice(ser.Categories))
			writeChartSeriesItem(xw, "c:yVal", "c:numLit", floatSlice(ser.Values))
		} else {
			writeChartSeriesItem(xw, "c:cat", "c:strLit", stringSlice(ser.Categories))
			writeChartSeriesItem(xw, "c:val", "c:numLit", floatSlice(ser.Values))
		}
		if kind.typ == "line" && (ser.Smooth || st.LineSmooth) {
			xw.Empty("c:smooth", "val", "1")
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

func writeChartAxisTxPr(xw *common.XMLWriter) {
	xw.Start("c:txPr")
	xw.Empty("a:bodyPr")
	xw.Empty("a:lstStyle")
	xw.Start("a:p")
	xw.Start("a:pPr")
	xw.Empty("a:defRPr")
	xw.End()
	xw.Empty("a:endParaRPr", "lang", "en-US")
	xw.End()
	xw.End()
}

func writeChartAxis(xw *common.XMLWriter, tag string, id int, pos string, cross int, st style.Chart, cat bool, kind chartKind, crosses string) {
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
	code := "General"
	linked := "1"
	if !cat {
		if pos == "r" {
			linked = "0"
			if st.SecondaryValueNumFmt != "" {
				code = st.SecondaryValueNumFmt
			}
		} else if st.ValueNumFmt != "" {
			code = st.ValueNumFmt
		}
	}
	xw.Empty("c:numFmt", "formatCode", code, "sourceLinked", linked)
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
	writeChartAxisTxPr(xw)
	xw.Empty("c:crossAx", "val", itoa(cross))
	if crosses == "" {
		crosses = "autoZero"
	}
	xw.Empty("c:crosses", "val", crosses)
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
