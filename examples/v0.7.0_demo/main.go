// Command v0.7.0_demo builds v070_demo.docx (area + combo charts, image)
// and template_v2.docx (pipe filters and comparison ifs), then streams
// 100_000 paragraphs through StreamExtractText.
package main

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log"
	"os"

	"github.com/yunkeweb/go-word"
	"github.com/yunkeweb/go-word/style"
)

const (
	outFile     = "v070_demo.docx"
	templateOut = "template_v2.docx"
	streamCount = 100000
)

func main() {
	if err := writeChartsDoc(); err != nil {
		log.Fatal(err)
	}
	if err := writeTemplateDoc(); err != nil {
		log.Fatal(err)
	}
	if err := runStreamExtract(); err != nil {
		log.Fatal(err)
	}
}

func writeChartsDoc() error {
	doc := word.New()
	info := doc.GetDocInfo()
	info.Title = "GoWord v0.7.0 Streaming Parser, Charts & Filters"
	info.Creator = "GoWord"
	info.Subject = "Area chart, combo chart, data labels, streaming extract"
	info.Company = "yunkeweb"

	doc.SetDefaultFontName("Calibri")
	doc.SetDefaultAsianFontName("Microsoft YaHei")
	doc.SetDefaultFontSize(11)

	sec := doc.AddSection()
	sec.AddTitle("Streaming Parser, Extended Charts & Template Filters", 1)
	sec.AddText("This document exercises DrawingML area charts, column+line combo charts with a secondary axis, data labels, and an inline image for StreamExtractImages.")

	sec.AddTitle("Stacked Area Chart", 2)
	quarters := []string{"Q1", "Q2", "Q3", "Q4"}
	area := doc.AddChart(word.ChartTypeStackedArea, quarters, []float64{12, 18, 15, 22}, []float64{8, 11, 13, 16})
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

	sec.AddTitle("Combo Chart (column + line, secondary axis)", 2)
	cats := []string{"Jan", "Feb", "Mar", "Apr", "May", "Jun"}
	combo := doc.AddChart(word.ChartTypeCombo, cats, []float64{120, 150, 140, 180, 210, 190})
	combo.Series[0].Name = "Sales"
	combo.Series[0].Kind = word.ChartTypeColumn
	combo.AddComboSeries(word.ChartTypeLine, cats, []float64{8.2, 9.1, 8.7, 10.4, 11.2, 10.8}, "Margin %", true)
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

	sec.AddTitle("Inline Image", 2)
	sec.AddImageBytes("swatch.png", demoPNG(), style.Image{Width: 40, Height: 20, AltText: "demo swatch"})
	sec.AddText("StreamExtractText walks each w:p; StreamExtractImages follows a:blip r:embed through document.xml.rels.")

	if err := doc.Save(outFile); err != nil {
		return err
	}
	fmt.Printf("wrote %s\n", outFile)

	f, err := os.Open(outFile)
	if err != nil {
		return err
	}
	defer f.Close()
	nPara := 0
	if err := word.StreamExtractText(f, func(s string) error {
		if s != "" {
			nPara++
		}
		return nil
	}); err != nil {
		return err
	}
	if _, err := f.Seek(0, 0); err != nil {
		return err
	}
	nImg := 0
	if err := word.StreamExtractImages(f, func(img word.ImageFile) error {
		nImg++
		fmt.Printf("stream image %s %s %d bytes\n", img.Name, img.MIME, len(img.Data))
		return nil
	}); err != nil {
		return err
	}
	fmt.Printf("stream extract: %d non-empty paragraphs, %d images\n", nPara, nImg)
	return nil
}

func writeTemplateDoc() error {
	src := word.New()
	sec := src.AddSection()
	sec.AddTitle("Template Engine v2", 1)
	sec.AddText("Hello ${name | upper}")
	sec.AddText("City ${city | lower}")
	sec.AddText("Title ${title | trim | upper}")
	sec.AddText("Blurb ${blurb | truncate:12}")
	sec.AddText("When ${created_at | formatDate:\"2006-01-02\"}")
	sec.AddText("Amount ${amount | formatCurrency:¥}")
	sec.AddText("Empty ${missing | default:N/A}")
	sec.AddText("${if age >= 18}Adult content is visible.${endif}")
	sec.AddText("${if age < 18}Minor content should be clipped.${endif}")
	raw, err := src.Bytes()
	if err != nil {
		return err
	}
	tp, err := word.NewTemplateProcessorBytes(raw)
	if err != nil {
		return err
	}
	tp.SetValue("name", "alice")
	tp.SetValue("city", "Shanghai")
	tp.SetValue("title", "  go-word  ")
	tp.SetValue("blurb", "Streaming parser, charts and filters")
	tp.SetValue("created_at", "2026-09-26T08:00:00Z")
	tp.SetValue("amount", "1999.5")
	tp.SetValue("age", "21")
	if err := tp.Save(templateOut); err != nil {
		return err
	}
	fmt.Printf("wrote %s\n", templateOut)
	return nil
}

func runStreamExtract() error {
	var buf bytes.Buffer
	sw := word.NewStreamWriter(&buf)
	for i := 0; i < streamCount; i++ {
		if err := sw.WriteParagraph(fmt.Sprintf("row %d", i+1)); err != nil {
			return err
		}
	}
	if err := sw.Close(); err != nil {
		return err
	}
	n := 0
	first, last := "", ""
	if err := word.StreamExtractText(bytes.NewReader(buf.Bytes()), func(s string) error {
		if s == "" {
			return nil
		}
		if first == "" {
			first = s
		}
		last = s
		n++
		return nil
	}); err != nil {
		return err
	}
	fmt.Printf("streamed %d paragraphs (%s … %s), zip %d bytes\n", n, first, last, buf.Len())
	if n < streamCount {
		return fmt.Errorf("stream extract counted %d, want >= %d", n, streamCount)
	}
	return nil
}

func demoPNG() []byte {
	img := image.NewRGBA(image.Rect(0, 0, 40, 20))
	for y := 0; y < 20; y++ {
		for x := 0; x < 40; x++ {
			img.Set(x, y, color.RGBA{R: 31, G: 78, B: 121, A: 255})
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		log.Fatal(err)
	}
	return buf.Bytes()
}
