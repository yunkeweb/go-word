// Command generate_full_feature_docs writes 50–100 randomly combined
// GoWord feature documents into ./test_output_docs.
//
// Every public surface listed below is exercised through real exported
// APIs only (AddText, AddTable, SetTextWatermark, Protect, Save, …).
package main

import (
	"bytes"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strings"

	"github.com/yunkeweb/go-word"
	"github.com/yunkeweb/go-word/element"
	"github.com/yunkeweb/go-word/style"
)

const (
	featTypography uint16 = 1 << iota
	featSections
	featHeaderFooter
	featTables
	featImages
	featAdvanced
	featMath
	featCharts
	featSDT
	featWatermark
	featAll = (1 << iota) - 1
)

var featNames = []struct {
	bit  uint16
	tag  string
	name string
}{
	{featTypography, "typo", "typography"},
	{featSections, "sect", "multi-section"},
	{featHeaderFooter, "hf", "header-footer"},
	{featTables, "tbl", "tables"},
	{featImages, "img", "images-shapes"},
	{featAdvanced, "adv", "toc-bookmark-review"},
	{featMath, "omml", "omml-math"},
	{featCharts, "cht", "charts"},
	{featSDT, "sdt", "sdt-controls"},
	{featWatermark, "wm", "watermark-protect"},
}

func main() {
	outDir := flag.String("dir", "test_output_docs", "output directory for .docx files")
	n := flag.Int("n", 80, "number of documents to generate (50–100)")
	seed := flag.Int64("seed", 20260926, "base RNG seed")
	flag.Parse()
	if *n < 50 {
		*n = 50
	}
	if *n > 100 {
		*n = 100
	}

	if err := os.MkdirAll(*outDir, 0o755); err != nil {
		log.Fatal(err)
	}

	seen := uint16(0)
	indexPath := filepath.Join(*outDir, "matrix_index.tsv")
	idx, err := os.Create(indexPath)
	if err != nil {
		log.Fatal(err)
	}
	defer idx.Close()
	fmt.Fprintln(idx, "file\tmask\tfeatures")

	pngBytes := mustPNG(color.RGBA{R: 31, G: 78, B: 121, A: 255})
	jpegBytes := mustJPEG(color.RGBA{R: 237, G: 125, B: 49, A: 255})

	for i := 0; i < *n; i++ {
		rng := rand.New(rand.NewSource(*seed + int64(i)*997))
		mask := pickMask(i, *n, rng)
		seen |= mask
		name := filename(i, mask)
		path := filepath.Join(*outDir, name)
		if err := writeDoc(path, i, mask, rng, pngBytes, jpegBytes); err != nil {
			log.Fatalf("doc %d (%s): %v", i, name, err)
		}
		fmt.Fprintf(idx, "%s\t0x%03x\t%s\n", name, mask, featureList(mask))
		fmt.Printf("wrote %s\n", path)
	}

	if seen != featAll {
		log.Fatalf("coverage gap: got 0x%03x want 0x%03x", seen, featAll)
	}
	fmt.Printf("generated %d documents in %s (all 10 modules covered)\n", *n, *outDir)
}

func pickMask(i, n int, rng *rand.Rand) uint16 {
	if i == 0 {
		return featAll
	}
	// First 10 documents after the kitchen-sink each force one module.
	if i <= len(featNames) {
		mask := featNames[i-1].bit
		for _, f := range featNames {
			if f.bit != mask && rng.Float64() < 0.45 {
				mask |= f.bit
			}
		}
		return mask
	}
	var mask uint16
	for _, f := range featNames {
		if rng.Float64() < 0.55 {
			mask |= f.bit
		}
	}
	if mask == 0 {
		mask = featNames[rng.Intn(len(featNames))].bit
	}
	// Sprinkle a few extra kitchen-sink files across the batch.
	if i == n-1 || i%17 == 0 {
		mask = featAll
	}
	return mask
}

func filename(i int, mask uint16) string {
	var tags []string
	for _, f := range featNames {
		if mask&f.bit != 0 {
			tags = append(tags, f.tag)
		}
	}
	if len(tags) > 4 {
		tags = tags[:4]
	}
	return fmt.Sprintf("full_%03d_%s.docx", i, strings.Join(tags, "-"))
}

func featureList(mask uint16) string {
	var names []string
	for _, f := range featNames {
		if mask&f.bit != 0 {
			names = append(names, f.name)
		}
	}
	return strings.Join(names, ",")
}

func writeDoc(path string, i int, mask uint16, rng *rand.Rand, pngBytes, jpegBytes []byte) error {
	doc := word.New()
	info := doc.GetDocInfo()
	info.Title = fmt.Sprintf("GoWord matrix %03d", i)
	info.Creator = "GoWord matrix"
	info.Subject = featureList(mask)
	info.Company = "yunkeweb"
	info.Keywords = "integration,matrix,v0.1.0-v0.9.0"

	doc.SetDefaultFontName("Calibri")
	doc.SetDefaultAsianFontName("Microsoft YaHei")
	doc.SetDefaultFontSize(11)
	doc.AddTitleStyle(1, style.Font{Bold: true, Size: 16, Color: "1F4E79"})
	doc.AddTitleStyle(2, style.Font{Bold: true, Size: 13, Color: "2E75B6"})
	doc.AddTitleStyle(3, style.Font{Bold: true, Size: 12, Color: "5B9BD5"})

	if mask&featWatermark != 0 {
		if err := applyWatermarkProtect(doc, rng, pngBytes, path); err != nil {
			return err
		}
	}
	if mask&featAdvanced != 0 {
		doc.EnableTrackChanges(true)
	}
	if mask&featHeaderFooter != 0 {
		doc.SetDifferentFirstPage(true)
		doc.SetEvenAndOddHeaders(true)
	}

	margin := 720 + rng.Intn(3)*360 // 0.5–1.25 inch
	cover := doc.AddSection(style.Section{
		Orientation:  style.OrientationPortrait,
		PageSizeW:    style.DefaultPageWidth,
		PageSizeH:    style.DefaultPageHeight,
		MarginTop:    margin,
		MarginBottom: margin,
		MarginLeft:   margin,
		MarginRight:  margin,
		HeaderHeight: style.DefaultHeaderHeight,
		FooterHeight: style.DefaultFooterHeight,
	})

	if mask&featHeaderFooter != 0 {
		addHeadersFooters(cover)
	}

	cover.AddTitle(fmt.Sprintf("Feature matrix document %03d", i), 1)
	cover.AddText("Enabled modules: "+featureList(mask), style.Font{Italic: true, Size: 11, Color: "44546A"})

	if mask&featAdvanced != 0 {
		cover.AddTitle("Contents", 2)
		cover.AddTOC(nil, nil, 1, 3)
		cover.AddText("Right-click the TOC field in Microsoft Word and choose Update Field.")
	}

	if mask&featTypography != 0 {
		addTypography(cover)
	}
	if mask&featAdvanced != 0 {
		addBookmarksComments(doc, cover, i)
	}
	if mask&featTables != 0 {
		addTables(cover, rng, mask&featWatermark != 0)
	}
	if mask&featSDT != 0 {
		addSDT(cover, i)
	}
	if mask&featImages != 0 {
		addImagesShapes(doc, cover, pngBytes, jpegBytes)
	}
	if mask&featMath != 0 {
		addMath(cover)
	}
	if mask&featCharts != 0 {
		addCharts(cover, rng)
	}

	if mask&featSections != 0 {
		addExtraSections(doc, rng, mask)
	}

	return doc.Save(path)
}

func applyWatermarkProtect(doc *word.Document, rng *rand.Rand, pngBytes []byte, path string) error {
	kind := rng.Intn(3)
	if kind == 0 || kind == 2 {
		opts := word.WatermarkOptions{
			Angle:    []float64{-45, -30, 15, 0}[rng.Intn(4)],
			Color:    "C0C0C0",
			FontSize: 28 + rng.Intn(12),
			Opacity:  0.22 + rng.Float64()*0.2,
			Tile:     rng.Float64() < 0.7,
			Rows:     3,
			Cols:     3,
		}
		doc.SetTextWatermark("CONFIDENTIAL", opts)
	}
	if kind == 1 || kind == 2 {
		mark := filepath.Join(filepath.Dir(path), "_matrix_mark.png")
		if err := os.WriteFile(mark, pngBytes, 0o644); err != nil {
			return err
		}
		if err := doc.SetImageWatermarkFile(mark, word.ImageWatermarkOptions{
			Washout: true,
			Scale:   1.0 + rng.Float64()*0.4,
			Opacity: 0.3,
		}); err != nil {
			return err
		}
	}
	modes := []string{
		word.ProtectTypeReadOnly,
		word.ProtectTypeComments,
		word.ProtectTypeForms,
	}
	return doc.Protect(modes[rng.Intn(len(modes))], "goword")
}

func addHeadersFooters(sec *element.Section) {
	sec.AddHeader().AddText("GoWord matrix — odd pages", style.Font{Size: 9, Color: "666666"})
	sec.AddHeader(element.HeaderFirst).AddText("GoWord matrix — first page", style.Font{Size: 9, Color: "666666"})
	sec.AddHeader(element.HeaderEven).AddText("GoWord matrix — even pages", style.Font{Size: 9, Color: "666666"})

	odd := sec.AddFooter()
	odd.AddText("Page ", style.Font{Size: 9})
	odd.AddPageNumber()
	odd.AddText(" of ", style.Font{Size: 9})
	odd.AddNumPages()

	first := sec.AddFooter(element.HeaderFirst)
	first.AddText("Cover footer", style.Font{Size: 9})

	even := sec.AddFooter(element.HeaderEven)
	even.AddText("Even  ", style.Font{Size: 9})
	even.AddPageNumber()
}

func addTypography(sec *element.Section) {
	sec.AddTitle("Typography", 2)
	sec.AddText("Left aligned body with 11 pt Calibri.",
		style.Font{Size: 11},
		style.Paragraph{Alignment: style.JcLeft})
	sec.AddText("Centered heading-colored line.",
		style.Font{Size: 14, Color: "1F4E79", Bold: true},
		style.Paragraph{Alignment: style.JcCenter})
	sec.AddText("Right aligned italic note.",
		style.Font{Size: 11, Italic: true, Color: "C45911"},
		style.Paragraph{Alignment: style.JcRight})
	sec.AddText("Justified paragraph with first-line indent, hanging is unused, line spacing 360 twips (1.5×). The sentence is long enough for Word to wrap and show both alignment and spacing together.",
		style.Font{Size: 11},
		style.Paragraph{
			Alignment:   style.JcBoth,
			Indentation: style.Indentation{Left: 420, FirstLine: 420},
			Spacing:     style.Spacing{Before: 120, After: 160, Line: 360, Rule: "auto"},
		})

	sec.AddText("Bold run.", style.Font{Bold: true, Size: 12})
	sec.AddText("Italic run.", style.Font{Italic: true, Size: 12, Color: "548235"})
	sec.AddText("Single underline.", style.Font{Underline: style.UnderlineSingle, Size: 12, Color: "2E75B6"})
	sec.AddText("Strikethrough.", style.Font{Strikethrough: true, Size: 12, Color: "C00000"})
	sec.AddText("Yellow highlight on 16 pt text.", style.Font{Size: 16, Color: "1F4E79", FgColor: style.FgYellow, Bold: true})

	tr := sec.AddTextRun(style.Paragraph{Alignment: style.JcLeft})
	tr.AddText("Inline mix: ")
	tr.AddText("red-bold", style.Font{Color: "C00000", Bold: true})
	tr.AddText(" / ")
	tr.AddText("green-italic", style.Font{Color: "548235", Italic: true})
	tr.AddText(" / ")
	tr.AddText("blue-underline", style.Font{Color: "2E75B6", Underline: style.UnderlineSingle})
	tr.AddText(" / ")
	tr.AddText("strike", style.Font{Strikethrough: true})
	tr.AddText(".")
}

func addBookmarksComments(doc *word.Document, sec *element.Section, i int) {
	sec.AddTitle("Bookmarks, hyperlinks, comments, revisions", 2)
	name := fmt.Sprintf("matrix_target_%d", i)
	sec.AddBookmark(name)
	sec.AddText("Bookmark destination paragraph.", style.Font{Bold: true})
	sec.AddLink(name, "Jump to bookmark", style.Font{Color: "0563C1", Underline: style.UnderlineSingle}, nil, true)
	sec.AddLink("https://pkg.go.dev/github.com/yunkeweb/go-word", "GoWord on pkg.go.dev",
		style.Font{Color: "0563C1", Underline: style.UnderlineSingle})

	sec.CommentOn("This sentence is annotated.", "Please confirm the figure.", "Alice", "AL", "2026-09-26T08:00:00Z")
	sec.AddInsertion("Tracked insertion added in review.", "Bob", "2026-09-26T09:00:00Z", style.Font{Color: "548235"})
	sec.AddDeletion("Tracked deletion of obsolete clause.", "Bob", "2026-09-26T09:00:00Z", style.Font{Color: "C00000"})
	doc.CommentOn("Document-level CommentOn helper.", "Seen from Document.CommentOn.", "Carol", "CA")
}

func addTables(sec *element.Section, rng *rand.Rand, protect bool) {
	sec.AddTitle("Tables", 2)
	sec.AddText("Repeating header (w:tblHeader), unbreakable rows (w:cantSplit), vMerge, gridSpan, nested table, vAlign, textDirection.")

	line := style.Border{Style: style.BorderSingle, Size: 4, Color: "1F4E79"}
	tbl := sec.AddTable(style.Table{
		Width: 9360, Layout: "fixed", Alignment: style.JcTableCenter,
		Borders: style.Borders{Top: line, Left: line, Bottom: line, Right: line, InsideH: line, InsideV: line},
	})

	hdr := tbl.AddRow(360)
	tbl.SetHeaderRow(hdr)
	hdr.SetCantSplit(true)
	hdr.AddCell(2340, style.Cell{Width: 2340, BgColor: "1F4E79"}).
		SetVAlign("center").
		AddText("Region", style.Font{Bold: true, Color: "FFFFFF"})
	hdr.AddCell(2340, style.Cell{Width: 2340, BgColor: "1F4E79"}).
		SetVAlign("center").
		AddText("Item", style.Font{Bold: true, Color: "FFFFFF"})
	hdr.AddCell(2340, style.Cell{Width: 2340, BgColor: "1F4E79"}).
		SetVAlign("center").
		AddText("Qty", style.Font{Bold: true, Color: "FFFFFF"})
	hdr.AddCell(2340, style.Cell{Width: 2340, BgColor: "1F4E79"}).
		SetVAlign("center").
		AddText("Note", style.Font{Bold: true, Color: "FFFFFF"})

	r1 := tbl.AddRow(480)
	r1.SetCantSplit(true)
	merged := r1.AddCell(2340, style.Cell{VMerge: "restart", Width: 2340, BgColor: "DEEBF7"})
	merged.SetVAlign("center").AddText("APAC")
	r1.AddCell(2340).AddText("Hardware")
	r1.AddCell(2340).SetVAlign("center").AddText("12")
	host := r1.AddCell(2340)
	inner := host.AddTable(style.Table{Width: 2200})
	ir := inner.AddRow()
	ir.AddCell(1100).AddText("SKU")
	ir.AddCell(1100).AddText("A-01")

	r2 := tbl.AddRow(320)
	r2.AddCell(2340, style.Cell{VMerge: "continue"})
	r2.AddCell(2340).AddText("Software")
	r2.AddCell(2340).SetVAlign("bottom").AddText("8")
	note := r2.AddCell(2340, style.Cell{BgColor: "E2F0D9"})
	note.SetVAlign("center")
	if protect {
		note.AllowEdit("Everyone")
	}
	note.AddText(fmt.Sprintf("editable-%d", rng.Intn(99)))

	span := tbl.AddRow(400)
	span.SetCantSplit(true)
	vert := span.AddCell(2340, style.Cell{Width: 2340, BgColor: "1F4E79"})
	vert.SetVAlign("center").SetTextDirection("tbRl").
		AddText("竖排", style.Font{Bold: true, Color: "FFFFFF"})
	span.AddCell(7020, style.Cell{GridSpan: 3, Shading: style.Shading{Fill: "D6DCE4"}}).
		SetVAlign("center").
		AddText("Footer spans three columns (w:gridSpan=3).", style.Font{Bold: true})
}

func addSDT(sec *element.Section, i int) {
	sec.AddTitle("SDT form controls", 2)
	sec.AddSDTText("Full name", fmt.Sprintf("full_name_%d", i), "Enter full name")
	sec.AddSDTDropdown("Department", fmt.Sprintf("dept_%d", i), map[string]string{
		"eng": "Engineering",
		"fin": "Finance",
		"hr":  "Human Resources",
	})
	sec.AddSDTDate("Start date", fmt.Sprintf("start_%d", i), "yyyy-MM-dd")
	p := sec.AddTextRun()
	p.AddText("I have read the handbook  ")
	p.AddSDTCheckbox("Handbook", fmt.Sprintf("ack_%d", i), i%2 == 0)

	tbl := sec.AddTable(style.Table{Width: 9000})
	row := tbl.AddRow()
	row.AddCell(3000).AddText("Manager")
	row.AddCell(6000).AddSDTText("Manager", fmt.Sprintf("mgr_%d", i), "Enter manager name")
}

func addImagesShapes(doc *word.Document, sec *element.Section, pngBytes, jpegBytes []byte) {
	sec.AddTitle("Images and DrawingML shapes", 2)
	sec.AddImageBytes("matrix.png", pngBytes, style.Image{Width: 120, Height: 40, AltText: "PNG swatch"})
	sec.AddImageBytes("matrix.jpeg", jpegBytes, style.Image{Width: 120, Height: 40, AltText: "JPEG swatch"})

	sec.AddDMLShape("roundRect", 1828800, 914400, "5B9BD5", "2E75B6", 12700)
	tb := sec.AddTextBox(style.TextBox{
		Width: 240, Height: 80, BgColor: "FFF2CC", BorderColor: "BF8F00", BorderSize: 8,
	})
	tb.AddText("w:txbxContent", style.Font{Size: 11, Color: "595959"})

	doc.AddShape(word.ShapeRoundRect, word.ShapeOptions{
		FillColor: "ED7D31",
		LineColor: "C45911",
		Text:      "DrawingML",
		Font:      style.Font{Bold: true, Color: "FFFFFF"},
	})
	doc.AddShape(word.ShapeArrow, word.ShapeOptions{
		FillColor: "70AD47", LineColor: "548235",
	})
}

func addMath(sec *element.Section) {
	sec.AddTitle("Office Math (OMML)", 2)
	sec.AddText("Display fraction:")
	sec.AddMath(`\frac{a}{b}`)
	sec.AddMath(`x^{2} + y^{2} = z^{2}`)
	p := sec.AddTextRun()
	p.AddText("Inline: ")
	p.AddMath(`\sqrt{x_1} + \pi`)
}

func addCharts(sec *element.Section, rng *rand.Rand) {
	sec.AddTitle("Charts", 2)
	cats := []string{"Q1", "Q2", "Q3", "Q4"}
	v1 := []float64{12, 18, 15, 22}
	v2 := []float64{8, 11, 13, 16}
	st := style.Chart{Width: 5486400, Height: 2800400, ShowLegend: true, ShowGridY: true}

	kinds := []string{word.ChartTypeBar, word.ChartTypeLine, word.ChartTypePie, word.ChartTypeArea, word.ChartTypeStackedArea}
	// Always emit all five chart kinds so the module is complete in every
	// document that opts into charts; order is shuffled.
	rng.Shuffle(len(kinds), func(i, j int) { kinds[i], kinds[j] = kinds[j], kinds[i] })
	for _, k := range kinds {
		cs := st
		cs.Title = k
		switch k {
		case word.ChartTypePie:
			sec.AddChart(k, []string{"A", "B", "C"}, []float64{40, 35, 25}, cs)
		case word.ChartTypeStackedArea:
			ch := sec.AddChart(k, cats, v1, cs)
			ch.AddSeries(cats, v2, "Series B")
		default:
			sec.AddChart(k, cats, v1, cs)
		}
	}
}

func addExtraSections(doc *word.Document, rng *rand.Rand, mask uint16) {
	land := doc.AddSection(style.Section{
		Orientation:  style.OrientationLandscape,
		BreakType:    "nextPage",
		PageSizeW:    style.DefaultPageWidth,
		PageSizeH:    style.DefaultPageHeight,
		MarginTop:    1134,
		MarginBottom: 1134,
		MarginLeft:   1134,
		MarginRight:  1134,
	})
	land.SetOrientation(style.OrientationLandscape)
	if mask&featHeaderFooter != 0 {
		land.AddHeader().AddText("Landscape header", style.Font{Size: 9})
		f := land.AddFooter()
		f.AddText("L ", style.Font{Size: 9})
		f.AddPageNumber()
	}
	land.AddTitle("Landscape section", 1)
	land.AddText("This section is landscape (w:orient). Custom 2 cm margins.")
	if mask&featWatermark != 0 {
		land.AddText("Party A: ________________", style.Font{Bold: true}).AllowEdit("Everyone")
	}

	cols := 2 + rng.Intn(2)
	col := doc.AddSection(style.Section{
		Orientation: style.OrientationPortrait,
		BreakType:   "nextPage",
		PageSizeW:   style.DefaultPageWidth,
		PageSizeH:   style.DefaultPageHeight,
		MarginTop:   1134, MarginBottom: 1134, MarginLeft: 1134, MarginRight: 1134,
	})
	col.SetColumns(cols, 720, true)
	col.AddTitle(fmt.Sprintf("%d-column layout", cols), 1)
	col.AddText("Word flows this section into equal columns with a separator (w:cols w:sep). The first column starts here and continues until the column is full.")
	col.AddText("The second column receives overflow. Multi-section documents mix portrait, landscape, and w:cols in one package.")
	if mask&featMath != 0 {
		col.AddMath(`\frac{P L}{4}`)
	}
}

func mustPNG(c color.RGBA) []byte {
	img := image.NewRGBA(image.Rect(0, 0, 120, 40))
	for y := 0; y < 40; y++ {
		for x := 0; x < 120; x++ {
			img.Set(x, y, c)
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		log.Fatal(err)
	}
	return buf.Bytes()
}

func mustJPEG(c color.RGBA) []byte {
	img := image.NewRGBA(image.Rect(0, 0, 120, 40))
	for y := 0; y < 40; y++ {
		for x := 0; x < 120; x++ {
			img.Set(x, y, c)
		}
	}
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 85}); err != nil {
		log.Fatal(err)
	}
	return buf.Bytes()
}
