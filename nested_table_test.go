package word

import (
	"strings"
	"testing"

	"github.com/yunkeweb/go-word/style"
)

func TestNestedTableTrailingParagraph(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	outer := sec.AddTable()
	cell := outer.AddRow().AddCell(4000)
	cell.AddText("parent cell")
	inner := cell.AddTable()
	inner.AddRow().AddCell(2000).AddText("nested cell")

	raw, err := doc.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	xml := readZipFile(t, raw, "word/document.xml")
	tc := innerTCWithNestedTable(xml)
	if tc == "" {
		t.Fatal("missing nested w:tbl inside w:tc")
	}
	if !strings.Contains(tc, "<w:tbl>") {
		t.Fatal("nested tbl")
	}
	if !strings.HasSuffix(strings.TrimSpace(tc), "</w:p>") && !strings.Contains(tc, "</w:tbl><w:p></w:p>") && !strings.Contains(tc, "</w:tbl><w:p/>") {
		if idx := strings.LastIndex(tc, "</w:tbl>"); idx < 0 {
			t.Fatal("no nested tbl close")
		} else {
			rest := tc[idx:]
			if !strings.Contains(rest, "<w:p") {
				t.Fatalf("tc must end with w:p after nested tbl: %s", tc)
			}
		}
	}

	loaded, err := LoadBytes(raw)
	if err != nil {
		t.Fatal(err)
	}
	text := loaded.ExtractText()
	if !strings.Contains(text, "parent cell") || !strings.Contains(text, "nested cell") {
		t.Fatalf("roundtrip text=%q", text)
	}
}

func TestCellBorderPaddingAlignAndTextDirection(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	tbl := sec.AddTable()
	cell := tbl.AddRow().AddCell(3000)
	cell.SetBorder("top", style.BorderSingle, 12, "FF0000")
	cell.SetBorder("left", style.BorderDashed, 8, "00FF00")
	cell.SetBorder("bottom", style.BorderDouble, 16, "0000FF")
	cell.SetBorder("right", style.BorderDotted, 4, "000000")
	cell.SetPadding(80, 100, 120, 140)
	cell.SetVerticalAlignment(VAlignCenter)
	cell.SetTextDirection(TextDirectionVertical)
	cell.AddText("styled")

	if cell.Style.Borders.Top.Style != style.BorderSingle || cell.Style.Borders.Top.Color != "FF0000" {
		t.Fatalf("top border=%+v", cell.Style.Borders.Top)
	}
	if cell.Style.PaddingTop != 80 || cell.Style.PaddingLeft != 100 || cell.Style.PaddingBottom != 120 || cell.Style.PaddingRight != 140 {
		t.Fatalf("padding=%+v", cell.Style)
	}
	if cell.Style.VAlign != style.VAlignCenter {
		t.Fatal("valign")
	}
	if cell.Style.TextDir != style.TextDirectionTbRl {
		t.Fatal("textdir")
	}

	raw, err := doc.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	xml := readZipFile(t, raw, "word/document.xml")
	block := tcPrBlock(xml)
	if block == "" {
		t.Fatal("missing w:tcPr")
	}
	order := innerChildNames("<w:tcPr>"+block+"</w:tcPr>", "w:tcPr")
	assertSeqSubset(t, order, []string{"w:tcW", "w:tcBorders", "w:tcMar", "w:textDirection", "w:vAlign"})
	if !strings.Contains(block, `w:val="center"`) {
		t.Fatal("vAlign")
	}
	if !strings.Contains(block, `w:val="tbRl"`) {
		t.Fatal("textDirection")
	}
	if !strings.Contains(block, `w:val="single"`) || !strings.Contains(block, "FF0000") {
		t.Fatal("borders")
	}
	if !strings.Contains(block, `w:w="80"`) || !strings.Contains(block, `w:w="140"`) {
		t.Fatal("tcMar padding")
	}

	all := tbl.AddRow().AddCell(1000)
	all.SetBorder("all", style.BorderSingle, 4, "111111")
	if all.Style.Borders.Top.Color != "111111" || all.Style.Borders.Right.Style != style.BorderSingle {
		t.Fatal("all borders")
	}
}

func TestCellAddTableIsNestedNotSectionSibling(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	outer := sec.AddTable()
	cell := outer.AddRow().AddCell(2000)
	inner := cell.AddTable()
	inner.AddRow().AddCell(1000).AddText("n")
	if sec.CountElements() != 1 {
		t.Fatalf("section elements=%d want 1 outer table", sec.CountElements())
	}
	found := false
	for _, el := range cell.Elements() {
		if el == inner {
			found = true
		}
	}
	if !found {
		t.Fatal("inner table not a cell child")
	}
}

func innerTCWithNestedTable(documentXML string) string {
	i := strings.Index(documentXML, "<w:tbl>")
	if i < 0 {
		return ""
	}
	rest := documentXML[i+7:]
	j := strings.Index(rest, "<w:tbl>")
	if j < 0 {
		return ""
	}
	// walk back to enclosing <w:tc>
	before := rest[:j]
	tc := strings.LastIndex(before, "<w:tc")
	if tc < 0 {
		return ""
	}
	start := i + 7 + tc
	end := strings.Index(documentXML[start:], "</w:tc>")
	if end < 0 {
		return ""
	}
	return documentXML[start : start+end]
}

func tcPrBlock(documentXML string) string {
	i := strings.Index(documentXML, "<w:tcPr>")
	if i < 0 {
		i = strings.Index(documentXML, "<w:tcPr ")
		if i < 0 {
			return ""
		}
	}
	gt := strings.Index(documentXML[i:], ">")
	if gt < 0 {
		return ""
	}
	start := i + gt + 1
	end := strings.Index(documentXML[start:], "</w:tcPr>")
	if end < 0 {
		return ""
	}
	return documentXML[start : start+end]
}

func assertSeqSubset(t *testing.T, got, want []string) {
	t.Helper()
	pos := 0
	for _, w := range want {
		found := -1
		for i := pos; i < len(got); i++ {
			if got[i] == w {
				found = i
				break
			}
		}
		if found < 0 {
			t.Fatalf("missing %s in order %v (want %v)", w, got, want)
		}
		pos = found + 1
	}
}
