package word

import (
	"fmt"
	"strings"
	"testing"

	"github.com/yunkeweb/go-word/element"
	"github.com/yunkeweb/go-word/style"
)

func tableAdvDocumentXML(t *testing.T, build func(*element.Section)) string {
	t.Helper()
	doc := New()
	sec := doc.AddSection()
	build(sec)
	raw, err := doc.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	document := readZipFile(t, raw, "word/document.xml")
	assertWellFormedXML(t, "word/document.xml", document)
	return document
}

func TestTableHeaderAndCantSplitXML(t *testing.T) {
	xml := tableAdvDocumentXML(t, func(sec *element.Section) {
		tbl := sec.AddTable(style.Table{Width: 8000})
		hdr := tbl.AddRow(360)
		tbl.SetHeaderRow(hdr)
		hdr.SetCantSplit(true)
		hdr.AddCell(4000).AddText("Name")
		hdr.AddCell(4000).AddText("Dept")
		body := tbl.AddRow(280)
		body.SetCantSplit(true)
		body.AddCell(4000).AddText("Ada")
		body.AddCell(4000).AddText("Eng")
	})
	if !strings.Contains(xml, "<w:tblHeader") {
		t.Fatal("missing w:tblHeader")
	}
	if !strings.Contains(xml, "<w:cantSplit") {
		t.Fatal("missing w:cantSplit")
	}
	trPr := firstXMLBlock(xml, "w:trPr")
	if trPr == "" {
		t.Fatal("missing w:trPr")
	}
	seq := innerChildNames(trPr, "w:trPr")
	if len(seq) != 3 {
		t.Fatalf("trPr children %v", seq)
	}
	assertSeq(t, seq, "w:cantSplit", "w:trHeight", "w:tblHeader")
	if seq[0] != "w:cantSplit" || seq[1] != "w:trHeight" || seq[2] != "w:tblHeader" {
		t.Fatalf("CT_TrPrBase order %v", seq)
	}
}

func TestTableSetHeaderFalseOmitsTblHeader(t *testing.T) {
	xml := tableAdvDocumentXML(t, func(sec *element.Section) {
		tbl := sec.AddTable(style.Table{Width: 4000})
		row := tbl.AddRow(200)
		row.SetHeader(true).SetHeader(false)
		row.AddCell(4000).AddText("plain")
	})
	if strings.Contains(xml, "w:tblHeader") {
		t.Fatal("cleared header must not emit w:tblHeader")
	}
	if !strings.Contains(xml, "w:trHeight") {
		t.Fatal("height still written")
	}
}

func TestCellVAlignAndTextDirectionXML(t *testing.T) {
	xml := tableAdvDocumentXML(t, func(sec *element.Section) {
		tbl := sec.AddTable(style.Table{Width: 6000})
		row := tbl.AddRow(800)
		row.AddCell(2000).SetVAlign("top").AddText("top")
		row.AddCell(2000).SetVAlign("center").AddText("center")
		row.AddCell(2000).SetVAlign("bottom").SetTextDirection("tbRl").AddText("竖排")
	})
	if !strings.Contains(xml, `w:vAlign w:val="top"`) {
		t.Fatal("vAlign top")
	}
	if !strings.Contains(xml, `w:vAlign w:val="center"`) {
		t.Fatal("vAlign center")
	}
	if !strings.Contains(xml, `w:vAlign w:val="bottom"`) {
		t.Fatal("vAlign bottom")
	}
	if !strings.Contains(xml, `w:textDirection w:val="tbRl"`) {
		t.Fatal("textDirection tbRl")
	}
	tcPr := lastXMLBlockContaining(xml, "w:tcPr", `w:textDirection w:val="tbRl"`)
	if tcPr == "" {
		t.Fatal("missing tcPr with textDirection")
	}
	seq := innerChildNames(tcPr, "w:tcPr")
	assertSeq(t, seq, "w:tcW", "w:textDirection", "w:vAlign")
	if indexOf(seq, "w:textDirection") > indexOf(seq, "w:vAlign") {
		t.Fatalf("CT_TcPr order %v", seq)
	}
}

func TestCellVAlignAliases(t *testing.T) {
	xml := tableAdvDocumentXML(t, func(sec *element.Section) {
		tbl := sec.AddTable(style.Table{Width: 2000})
		tbl.AddRow().AddCell(2000).SetVAlign("middle").SetTextDirection("vertical").AddText("x")
	})
	if !strings.Contains(xml, `w:vAlign w:val="center"`) {
		t.Fatal("middle -> center")
	}
	if !strings.Contains(xml, `w:textDirection w:val="tbRl"`) {
		t.Fatal("vertical -> tbRl")
	}
}

func TestTableAdvancedRoundTrip(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	tbl := sec.AddTable(style.Table{Width: 5000})
	hdr := tbl.AddRow(400)
	tbl.SetHeaderRow(hdr)
	hdr.SetCantSplit(true)
	hdr.AddCell(2500).SetVAlign("center").AddText("A")
	hdr.AddCell(2500).SetTextDirection("btLr").SetVAlign("bottom").AddText("B")
	body := tbl.AddRow(240)
	body.AddCell(2500).AddText("1")
	body.AddCell(2500).AddText("2")

	raw, err := doc.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	loaded, err := LoadBytes(raw)
	if err != nil {
		t.Fatal(err)
	}
	got := firstLoadedTable(t, loaded)
	if len(got.Rows) != 2 {
		t.Fatalf("rows %d", len(got.Rows))
	}
	if !got.Rows[0].IsHeader() {
		t.Fatal("header round-trip")
	}
	if !got.Rows[0].IsCantSplit() {
		t.Fatal("cantSplit round-trip")
	}
	if got.Rows[0].Style.Height != 400 {
		t.Fatalf("height %d", got.Rows[0].Style.Height)
	}
	if got.Rows[0].Cells[0].Style.VAlign != style.VAlignCenter {
		t.Fatal("vAlign round-trip")
	}
	if got.Rows[0].Cells[1].Style.TextDir != style.TextDirectionBtLr {
		t.Fatal("textDirection round-trip")
	}
}

func TestLongTableHeaderRepeatsInXML(t *testing.T) {
	xml := tableAdvDocumentXML(t, func(sec *element.Section) {
		tbl := sec.AddTable(style.Table{Width: 9000})
		hdr := tbl.AddRow(300)
		tbl.SetHeaderRow(hdr)
		hdr.AddCell(3000).SetVAlign("center").AddText("Item")
		hdr.AddCell(3000).SetVAlign("center").AddText("Qty")
		hdr.AddCell(3000).SetVAlign("center").AddText("Note")
		for i := 1; i <= 40; i++ {
			row := tbl.AddRow(280)
			if i%5 == 0 {
				row.SetCantSplit(true)
			}
			row.AddCell(3000).AddText(fmt.Sprintf("Item %d", i))
			row.AddCell(3000).AddText("1")
			row.AddCell(3000).AddText("ok")
		}
	})
	nHeader := strings.Count(xml, "<w:tblHeader")
	if nHeader != 1 {
		t.Fatalf("tblHeader count %d", nHeader)
	}
	nSplit := strings.Count(xml, "<w:cantSplit")
	if nSplit != 8 {
		t.Fatalf("cantSplit count %d", nSplit)
	}
}

func lastXMLBlockContaining(s, tag, needle string) string {
	var last string
	rest := s
	for {
		block := firstXMLBlock(rest, tag)
		if block == "" {
			return last
		}
		if strings.Contains(block, needle) {
			last = block
		}
		i := strings.Index(rest, block)
		if i < 0 {
			return last
		}
		rest = rest[i+len(block):]
	}
}

func firstLoadedTable(t *testing.T, doc *Document) *element.Table {
	t.Helper()
	for _, sec := range doc.GetSections() {
		for _, el := range sec.Elements() {
			if tbl, ok := el.(*element.Table); ok {
				return tbl
			}
		}
	}
	t.Fatal("no table")
	return nil
}
