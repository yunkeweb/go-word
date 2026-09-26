package word

import (
	"strings"
	"testing"

	"github.com/yunkeweb/go-word/style"
)

func TestDLblPosForKind(t *testing.T) {
	bar := chartKinds["column"]
	if got := dLblPosForKind(bar, DataLabelPosTop); got != DataLabelPosOutEnd {
		t.Fatalf("column top -> %s", got)
	}
	if got := dLblPosForKind(bar, DataLabelPosCenter); got != DataLabelPosCenter {
		t.Fatalf("column ctr -> %s", got)
	}
	area := chartKinds["area"]
	if got := dLblPosForKind(area, DataLabelPosTop); got != "" {
		t.Fatalf("area must omit dLblPos, got %s", got)
	}
	if got := dLblPosForKind(area, DataLabelPosCenter); got != "" {
		t.Fatalf("area must omit dLblPos, got %s", got)
	}
}

func TestAreaChartXML(t *testing.T) {
	doc := New()
	ch := doc.AddChart(ChartTypeArea, []string{"Q1", "Q2", "Q3"}, []float64{1, 3, 2})
	ch.Series[0].Name = "Revenue"
	ch.SetLegendPosition(LegendBottom)
	ch.SetMajorGridlines(true)
	ch.SetDataLabels(ChartDataLabelOptions{
		ShowVal:     true,
		ShowCatName: true,
		Position:    DataLabelPosTop,
	})
	raw, err := doc.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	xml := readZipFile(t, raw, "word/charts/chart1.xml")
	assertWellFormedXML(t, "area", xml)
	if !strings.Contains(xml, "<c:areaChart>") {
		t.Fatal("missing c:areaChart")
	}
	if !strings.Contains(xml, `<c:legendPos val="b"`) {
		t.Fatal("legend bottom")
	}
	if strings.Contains(innerXML(xml, "c:plotArea"), "<c:legend") {
		t.Fatal("legend must not be inside plotArea")
	}
	assertSeq(t, innerChildNames(xml, "c:legend"), "c:legendPos", "c:layout", "c:overlay")
	if i, j := strings.Index(xml, "</c:plotArea>"), strings.Index(xml, "<c:legend"); i < 0 || j < 0 || i > j {
		t.Fatal("legend must follow plotArea under c:chart")
	}
	if i, j := strings.Index(xml, "</c:legend>"), strings.Index(xml, "<c:plotVisOnly"); i < 0 || j < 0 || i > j {
		t.Fatal("legend must precede plotVisOnly")
	}
	if !strings.Contains(xml, "<c:majorGridlines") {
		t.Fatal("major gridlines")
	}
	if !strings.Contains(xml, `<c:grouping val="standard"`) {
		t.Fatal("grouping standard")
	}
	areaKids := innerChildNames(xml, "c:areaChart")
	assertSeq(t, areaKids, "c:grouping", "c:ser", "c:dLbls", "c:axId")
	areaInner := innerXML(xml, "c:areaChart")
	if strings.Contains(areaInner, "<c:varyColors") {
		t.Fatal("areaChart must not emit c:varyColors")
	}
	if !strings.Contains(areaInner, "<c:dLbls") {
		t.Fatal("SetDataLabels must emit c:dLbls")
	}
	if strings.Contains(xml, "<c:dLblPos") {
		t.Fatal("area dLbls must omit c:dLblPos")
	}
	dLblKids := innerChildNames(xml, "c:dLbls")
	assertSeq(t, dLblKids, "c:showLegendKey", "c:showVal", "c:showCatName", "c:showSerName", "c:showPercent")
	dLblInner := innerXML(xml, "c:dLbls")
	if strings.Contains(dLblInner, "showBubbleSize") || strings.Contains(dLblInner, "showLeaderLines") {
		t.Fatal("area dLbls must omit showBubbleSize/showLeaderLines")
	}
	serKids := innerChildNames(xml, "c:ser")
	assertSeq(t, serKids, "c:idx", "c:order", "c:tx", "c:spPr", "c:cat", "c:val")
	if strings.Contains(xml, "<c:smooth") || strings.Contains(xml, "<c:marker") {
		t.Fatal("area ser must not contain c:smooth or c:marker")
	}
	plot := innerChildNames(xml, "c:plotArea")
	assertSeq(t, plot, "c:layout", "c:areaChart", "c:catAx", "c:valAx")
}

func TestComboChartSecondaryAxisXML(t *testing.T) {
	doc := New()
	cats := []string{"Jan", "Feb", "Mar"}
	ch := doc.AddChart(ChartTypeCombo, cats, []float64{10, 20, 15})
	ch.Series[0].Name = "Sales"
	ch.Series[0].Kind = ChartTypeColumn
	ch.AddComboSeries(ChartTypeLine, cats, []float64{1.1, 2.2, 1.8}, "Rate", true)
	ch.SetLegendPosition(LegendRight)
	ch.SetMajorGridlines(true)
	ch.SetDataLabels(ChartDataLabelOptions{ShowVal: true, Position: DataLabelPosTop})
	ch.Style.Title = "Sales vs Rate"

	raw, err := doc.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	xml := readZipFile(t, raw, "word/charts/chart1.xml")
	assertWellFormedXML(t, "combo", xml)
	if !strings.Contains(xml, "<c:barChart>") {
		t.Fatal("missing c:barChart")
	}
	if !strings.Contains(xml, "<c:lineChart>") {
		t.Fatal("missing c:lineChart")
	}
	if !strings.Contains(xml, `<c:barDir val="col"`) {
		t.Fatal("column barDir")
	}
	if !strings.Contains(xml, `<c:axId val="3"`) {
		t.Fatal("secondary axId 3")
	}
	if !strings.Contains(xml, `<c:axPos val="r"`) {
		t.Fatal("secondary axPos r")
	}
	if !strings.Contains(xml, `<c:crosses val="max"`) {
		t.Fatal("secondary crosses max")
	}
	plot := innerChildNames(xml, "c:plotArea")
	assertSeq(t, plot, "c:layout", "c:barChart", "c:lineChart", "c:catAx", "c:valAx")
	barKids := innerChildNames(xml, "c:barChart")
	assertSeq(t, barKids, "c:barDir", "c:grouping", "c:varyColors", "c:ser", "c:dLbls", "c:axId")
	lineKids := innerChildNames(xml, "c:lineChart")
	assertSeq(t, lineKids, "c:grouping", "c:varyColors", "c:ser", "c:dLbls", "c:axId")
	if strings.Count(xml, "<c:valAx>") != 2 {
		t.Fatalf("want 2 valAx, xml=%s", xml)
	}
	if !strings.Contains(xml, "<c:v>Sales</c:v>") || !strings.Contains(xml, "<c:v>Rate</c:v>") {
		t.Fatal("series names")
	}
}

func TestChartDataLabelsAndLegendAPI(t *testing.T) {
	doc := New()
	ch := doc.AddChartStyled(ChartTypeBar, []string{"A", "B"}, []float64{1, 2}, style.Chart{Title: "T"})
	ch.SetDataLabels(ChartDataLabelOptions{ShowVal: true, ShowPercent: true, ShowCatName: false, Position: DataLabelPosCenter})
	ch.SetLegendPosition(LegendTop)
	ch.SetMajorGridlines(false)
	raw, err := doc.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	xml := readZipFile(t, raw, "word/charts/chart1.xml")
	if !strings.Contains(xml, `<c:legendPos val="t"`) {
		t.Fatal("legend top")
	}
	assertSeq(t, innerChildNames(xml, "c:legend"), "c:legendPos", "c:layout", "c:overlay")
	if strings.Contains(innerXML(xml, "c:plotArea"), "<c:legend") {
		t.Fatal("legend must not be inside plotArea")
	}
	if !strings.Contains(xml, `<c:showPercent val="1"`) {
		t.Fatal("showPercent")
	}
	if !strings.Contains(xml, `<c:showCatName val="0"`) {
		t.Fatal("showCatName off")
	}
	if strings.Contains(xml, "<c:majorGridlines") {
		t.Fatal("gridlines should be off")
	}
}

func TestStackedMultiSeriesAreaChartXML(t *testing.T) {
	doc := New()
	cats := []string{"Q1", "Q2", "Q3"}
	ch := doc.AddChart(ChartTypeStackedArea, cats, []float64{1, 2, 3}, []float64{4, 5, 6})
	ch.Series[0].Name = "Hardware"
	ch.Series[1].Name = "Software"
	ch.SetLegendPosition(LegendTop)
	raw, err := doc.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	xml := readZipFile(t, raw, "word/charts/chart1.xml")
	assertWellFormedXML(t, "stacked-area", xml)
	if !strings.Contains(xml, `<c:grouping val="stacked"`) {
		t.Fatal("stacked grouping")
	}
	if strings.Count(xml, "<c:ser>") != 2 {
		t.Fatalf("want 2 series, xml=%s", xml)
	}
	if !strings.Contains(xml, `<c:legendPos val="t"`) {
		t.Fatal("legend top")
	}
	if strings.Contains(innerXML(xml, "c:plotArea"), "<c:legend") {
		t.Fatal("legend must not be inside plotArea")
	}
	if strings.Contains(innerXML(xml, "c:areaChart"), "<c:dLbls") {
		t.Fatal("area dLbls only when SetDataLabels")
	}
	if strings.Contains(innerXML(xml, "c:areaChart"), "<c:varyColors") {
		t.Fatal("areaChart must not emit c:varyColors")
	}
	if strings.Contains(xml, "<c:smooth") || strings.Contains(xml, "<c:marker") {
		t.Fatal("area ser must not contain c:smooth or c:marker")
	}

	pct := New()
	pch := pct.AddChart(ChartTypePercentStackedArea, cats, []float64{1, 2, 3}, []float64{4, 5, 6})
	pch.SetLegendPosition(LegendRight)
	praw, err := pct.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	pxml := readZipFile(t, praw, "word/charts/chart1.xml")
	if !strings.Contains(pxml, `<c:grouping val="percentStacked"`) {
		t.Fatal("percentStacked grouping")
	}
	if !strings.Contains(pxml, `<c:legendPos val="r"`) {
		t.Fatal("legend right")
	}
}

func TestComboLineSmoothMarkerAndAxisFmt(t *testing.T) {
	doc := New()
	cats := []string{"Jan", "Feb", "Mar"}
	ch := doc.AddChart(ChartTypeCombo, cats, []float64{10, 20, 15})
	ch.Series[0].Name = "Sales"
	ch.Series[0].Kind = ChartTypeColumn
	ch.AddComboSeries(ChartTypeLine, cats, []float64{1.1, 2.2, 1.8}, "Rate", true)
	ch.SetLineSmooth(true)
	ch.SetLineMarker("circle")
	ch.SetMajorGridlines(true)
	ch.Style.ValueNumFmt = "0"
	ch.Style.SecondaryValueNumFmt = "0.00"
	raw, err := doc.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	xml := readZipFile(t, raw, "word/charts/chart1.xml")
	assertWellFormedXML(t, "combo-smooth", xml)
	line := innerXML(xml, "c:lineChart")
	if !strings.Contains(line, `<c:symbol val="circle"`) {
		t.Fatal("line marker circle")
	}
	if !strings.Contains(line, `<c:size val="5"`) {
		t.Fatal("line marker size 5")
	}
	if !strings.Contains(line, `<c:smooth val="1"`) {
		t.Fatal("line smooth")
	}
	serKids := innerChildNames(line, "c:ser")
	assertSeq(t, serKids, "c:idx", "c:order", "c:tx", "c:spPr", "c:marker", "c:cat", "c:val", "c:smooth")
	prim := axisBlockByID(xml, "c:valAx", "2")
	if !strings.Contains(prim, "<c:majorGridlines") {
		t.Fatal("primary valAx gridlines")
	}
	if !strings.Contains(prim, `formatCode="0"`) {
		t.Fatal("primary valAx numFmt")
	}
	sec := axisBlockByID(xml, "c:valAx", "3")
	if strings.Contains(sec, "<c:majorGridlines") {
		t.Fatal("secondary valAx must not emit gridlines")
	}
	if !strings.Contains(sec, `formatCode="0.00"`) || !strings.Contains(sec, `sourceLinked="0"`) {
		t.Fatalf("secondary valAx numFmt isolation: %s", sec)
	}
}
