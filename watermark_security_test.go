package word

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yunkeweb/go-word/style"
)

func TestTextWatermarkAngleColorOpacity(t *testing.T) {
	doc := New()
	doc.SetTextWatermark("DRAFT", WatermarkOptions{
		Angle:    -30,
		Color:    "C0C0C0",
		FontSize: 48,
		FontName: "Calibri",
		Opacity:  0.25,
	})
	doc.AddSection().AddText("body")
	raw, err := doc.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	hdr := readZipFile(t, raw, "word/header1.xml")
	assertWellFormedXML(t, "header1.xml", hdr)
	if !strings.Contains(hdr, `rotation:330`) {
		t.Fatalf("angle -30 -> rotation 330: %s", hdr)
	}
	if !strings.Contains(hdr, `fillcolor="#C0C0C0"`) {
		t.Fatal("fillcolor hex")
	}
	if !strings.Contains(hdr, `opacity=".25"`) {
		t.Fatal("opacity")
	}
	if !strings.Contains(hdr, `font-size:48pt`) {
		t.Fatal("font size")
	}
	if !strings.Contains(hdr, `string="DRAFT"`) {
		t.Fatal("text")
	}
}

func TestTextWatermarkTileGrid(t *testing.T) {
	doc := New()
	doc.SetTextWatermark("CONFIDENTIAL", WatermarkOptions{Tile: true, Rows: 2, Cols: 3, Angle: -45})
	doc.AddSection().AddText("body")
	raw, err := doc.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	hdr := readZipFile(t, raw, "word/header1.xml")
	assertWellFormedXML(t, "header1.xml", hdr)
	if strings.Count(hdr, "PowerPlusWaterMarkObject") != 6 {
		t.Fatalf("tile count: %d", strings.Count(hdr, "PowerPlusWaterMarkObject"))
	}
	if !strings.Contains(hdr, "mso-position-horizontal-relative:page") {
		t.Fatal("page-relative tiles")
	}
	if !strings.Contains(hdr, `rotation:315`) {
		t.Fatal("default diagonal on tiles")
	}
}

func TestImageWatermarkWashoutScaleOpacity(t *testing.T) {
	doc := New()
	doc.SetImageWatermark(pngBytes(t), ImageWatermarkOptions{
		Washout: true,
		Scale:   0.5,
		Opacity: 0.4,
	})
	doc.AddSection().AddText("body")
	raw, err := doc.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	hdr := readZipFile(t, raw, "word/header1.xml")
	assertWellFormedXML(t, "header1.xml", hdr)
	if !strings.Contains(hdr, `v:imagedata`) {
		t.Fatal("imagedata")
	}
	if !strings.Contains(hdr, `gain="19661f"`) || !strings.Contains(hdr, `blacklevel="22938f"`) {
		t.Fatal("washout")
	}
	if !strings.Contains(hdr, `opacity=".4"`) {
		t.Fatal("image opacity")
	}
	if !strings.Contains(hdr, "WordPictureWatermark") {
		t.Fatal("shape id")
	}
}

func TestImageWatermarkFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "mark.png")
	if err := os.WriteFile(path, pngBytes(t), 0o644); err != nil {
		t.Fatal(err)
	}
	doc := New()
	if err := doc.SetImageWatermarkFile(path, ImageWatermarkOptions{Washout: true}); err != nil {
		t.Fatal(err)
	}
	doc.AddSection().AddText("body")
	raw, err := doc.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	hdr := readZipFile(t, raw, "word/header1.xml")
	if !strings.Contains(hdr, "v:imagedata") {
		t.Fatal("file watermark")
	}
	rels := readZipFile(t, raw, "word/_rels/header1.xml.rels")
	if !strings.Contains(rels, "media/") {
		t.Fatal("header media rel")
	}
}

func TestProtectModes(t *testing.T) {
	modes := []string{ProtectTypeReadOnly, ProtectTypeComments, ProtectTypeTrackedChanges, ProtectTypeForms}
	for _, mode := range modes {
		doc := New()
		if err := doc.Protect(mode, "secret"); err != nil {
			t.Fatal(err)
		}
		doc.AddSection().AddText("x")
		raw, err := doc.Bytes()
		if err != nil {
			t.Fatal(err)
		}
		settings := readZipFile(t, raw, "word/settings.xml")
		assertWellFormedXML(t, "settings.xml", settings)
		if !strings.Contains(settings, `w:documentProtection`) {
			t.Fatal("documentProtection")
		}
		if !strings.Contains(settings, `w:edit="`+mode+`"`) {
			t.Fatalf("edit=%s", mode)
		}
		if !strings.Contains(settings, `w:enforcement="1"`) {
			t.Fatal("enforcement")
		}
		if !strings.Contains(settings, `w:hash="`) {
			t.Fatal("hash")
		}
	}
}

func TestAllowEditPermStartEnd(t *testing.T) {
	doc := New()
	if err := doc.Protect(ProtectTypeReadOnly, "secret"); err != nil {
		t.Fatal(err)
	}
	sec := doc.AddSection()
	sec.AddText("locked")
	sec.AddText("editable paragraph").AllowEdit("Everyone")
	p := sec.AddTextRun()
	p.AddText("rich editable")
	p.AllowEdit("editors")
	sec.AddText("user range").AllowEdit("alice@example.com")
	tbl := sec.AddTable(style.Table{Width: 4000})
	row := tbl.AddRow()
	row.AddCell(2000).AllowEdit("Everyone").AddText("cell edit")
	row.AddCell(2000).AddText("locked cell")

	raw, err := doc.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	xml := readZipFile(t, raw, "word/document.xml")
	assertWellFormedXML(t, "document.xml", xml)
	if strings.Count(xml, "<w:permStart") != 4 {
		t.Fatalf("permStart count %d", strings.Count(xml, "<w:permStart"))
	}
	if strings.Count(xml, "<w:permEnd") != 4 {
		t.Fatalf("permEnd count %d", strings.Count(xml, "<w:permEnd"))
	}
	if !strings.Contains(xml, `w:edGrp="everyone"`) {
		t.Fatal("everyone")
	}
	if !strings.Contains(xml, `w:edGrp="editors"`) {
		t.Fatal("editors")
	}
	if !strings.Contains(xml, `w:ed="alice@example.com"`) {
		t.Fatal("user ed")
	}
	start := firstXMLBlock(xml, "w:permStart")
	if start == "" {
		t.Fatal("permStart block")
	}
	id := permIDFrom(start)
	if id == "" {
		t.Fatal("perm id")
	}
	if !strings.Contains(xml, `<w:permEnd w:id="`+id+`"`) {
		t.Fatalf("matching permEnd id=%s", id)
	}
}

func TestDefaultTextWatermarkStillDiagonal(t *testing.T) {
	doc := New()
	doc.SetTextWatermark("CONFIDENTIAL")
	doc.AddSection().AddText("body")
	raw, err := doc.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	hdr := readZipFile(t, raw, "word/header1.xml")
	if !strings.Contains(hdr, `rotation:315`) {
		t.Fatal("default -45")
	}
	if !strings.Contains(hdr, `id="PowerPlusWaterMarkObject"`) {
		t.Fatal("single shape id")
	}
}

func permIDFrom(block string) string {
	const key = `w:id="`
	i := strings.Index(block, key)
	if i < 0 {
		return ""
	}
	rest := block[i+len(key):]
	j := strings.Index(rest, `"`)
	if j < 0 {
		return ""
	}
	return rest[:j]
}

func TestAllowEditOnTable(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	tbl := sec.AddTable(style.Table{Width: 2000})
	tbl.AllowEdit("Everyone")
	tbl.AddRow().AddCell(2000).AddText("x")
	raw, err := doc.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	xml := readZipFile(t, raw, "word/document.xml")
	seq := innerChildNames(xml, "w:body")
	assertSeq(t, seq, "w:permStart", "w:tbl", "w:permEnd")
}
