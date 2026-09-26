package word

import (
	"strings"
	"testing"

	"github.com/yunkeweb/go-word/element"
	"github.com/yunkeweb/go-word/ooxml"
	"github.com/yunkeweb/go-word/style"
)

func sdtDocumentXML(t *testing.T, build func(*element.Section)) string {
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

func firstXMLBlock(s, tag string) string {
	open := "<" + tag
	i := strings.Index(s, open)
	if i < 0 {
		return ""
	}
	close := "</" + tag + ">"
	j := strings.Index(s[i:], close)
	if j < 0 {
		return s[i:]
	}
	return s[i : i+j+len(close)]
}

func TestSDTNamespacesAndBlockOrder(t *testing.T) {
	xml := sdtDocumentXML(t, func(sec *element.Section) {
		sec.AddSDTText("Name", "name", "Enter name")
		sec.AddSDTCheckbox("Done", "done", true)
	})
	for _, ns := range []string{
		`xmlns:w="` + ooxml.NSW + `"`,
		`xmlns:w14="` + ooxml.NSW14 + `"`,
		`mc:Ignorable="w14 wps"`,
	} {
		if !strings.Contains(xml, ns) {
			t.Fatalf("document.xml missing %s", ns)
		}
	}
	block := firstXMLBlock(xml, "w:sdt")
	if block == "" {
		t.Fatal("missing w:sdt")
	}
	seq := innerChildNames(block, "w:sdt")
	assertSeq(t, seq, "w:sdtPr", "w:sdtContent")
	if indexOf(seq, "w:sdtPr") < 0 || indexOf(seq, "w:sdtContent") < 0 {
		t.Fatalf("sdt children: %v", seq)
	}
}

func TestSDTPlainText(t *testing.T) {
	xml := sdtDocumentXML(t, func(sec *element.Section) {
		sec.AddSDTText("Full name", "full_name", "Enter your name")
	})
	block := firstXMLBlock(xml, "w:sdt")
	pr := firstXMLBlock(block, "w:sdtPr")
	seq := innerChildNames(pr, "w:sdtPr")
	assertSeq(t, seq, "w:alias", "w:tag", "w:id", "w:placeholder", "w:showingPlcHdr", "w:text")
	if !strings.Contains(pr, `<w:alias w:val="Full name"`) {
		t.Fatal("alias")
	}
	if !strings.Contains(pr, `<w:tag w:val="full_name"`) {
		t.Fatal("tag")
	}
	if !strings.Contains(pr, "<w:text") {
		t.Fatal("w:text control")
	}
	if !strings.Contains(block, "Enter your name") {
		t.Fatal("placeholder content")
	}
	if strings.Contains(pr, "w:dropDownList") || strings.Contains(pr, "w14:checkbox") {
		t.Fatal("plain text must not emit other controls")
	}
}

func TestSDTDropDownListItems(t *testing.T) {
	xml := sdtDocumentXML(t, func(sec *element.Section) {
		sec.AddSDTDropdown("Department", "dept", map[string]string{
			"eng": "Engineering",
			"hr":  "Human Resources",
			"fin": "Finance",
		})
	})
	block := firstXMLBlock(xml, "w:sdt")
	pr := firstXMLBlock(block, "w:sdtPr")
	seq := innerChildNames(pr, "w:sdtPr")
	assertSeq(t, seq, "w:alias", "w:tag", "w:id", "w:dropDownList")
	list := firstXMLBlock(pr, "w:dropDownList")
	if list == "" {
		t.Fatal("missing w:dropDownList")
	}
	items := innerChildNames(list, "w:dropDownList")
	if len(items) != 3 {
		t.Fatalf("listItem count: %v", items)
	}
	for _, want := range items {
		if want != "w:listItem" {
			t.Fatalf("dropDownList child %s", want)
		}
	}
	// Keys are sorted: eng, fin, hr.
	if !strings.Contains(list, `w:value="eng"`) || !strings.Contains(list, `w:displayText="Engineering"`) {
		t.Fatal("eng item")
	}
	if !strings.Contains(list, `w:value="fin"`) || !strings.Contains(list, `w:displayText="Finance"`) {
		t.Fatal("fin item")
	}
	if !strings.Contains(list, `w:value="hr"`) || !strings.Contains(list, `w:displayText="Human Resources"`) {
		t.Fatal("hr item")
	}
}

func TestSDTDateChildOrder(t *testing.T) {
	xml := sdtDocumentXML(t, func(sec *element.Section) {
		sec.AddSDTDate("Signed", "signed_at", "yyyy-MM-dd")
	})
	block := firstXMLBlock(xml, "w:sdt")
	pr := firstXMLBlock(block, "w:sdtPr")
	assertSeq(t, innerChildNames(pr, "w:sdtPr"), "w:alias", "w:tag", "w:id", "w:date")
	date := firstXMLBlock(pr, "w:date")
	seq := innerChildNames(date, "w:date")
	if len(seq) != 4 {
		t.Fatalf("w:date children %v", seq)
	}
	assertSeq(t, seq, "w:dateFormat", "w:lid", "w:storeMappedDataAs", "w:calendar")
	if seq[0] != "w:dateFormat" || seq[1] != "w:lid" || seq[2] != "w:storeMappedDataAs" || seq[3] != "w:calendar" {
		t.Fatalf("strict date order %v", seq)
	}
	if !strings.Contains(date, `w:dateFormat w:val="yyyy-MM-dd"`) {
		t.Fatal("dateFormat")
	}
	if !strings.Contains(date, `w:lid w:val="en-US"`) {
		t.Fatal("lid")
	}
	if !strings.Contains(date, `w:storeMappedDataAs w:val="dateTime"`) {
		t.Fatal("storeMappedDataAs")
	}
	if !strings.Contains(date, `w:calendar w:val="gregorian"`) {
		t.Fatal("calendar")
	}
}

func TestSDTDateFullDateAttribute(t *testing.T) {
	xml := sdtDocumentXML(t, func(sec *element.Section) {
		sec.AddSDTDate("Due", "due", "yyyy-MM-dd").SetValue("2026-09-26")
	})
	date := firstXMLBlock(xml, "w:date")
	if !strings.Contains(date, `w:fullDate="2026-09-26T00:00:00Z"`) {
		t.Fatalf("fullDate: %s", date)
	}
	if strings.Contains(firstXMLBlock(xml, "w:sdtPr"), "w:showingPlcHdr") {
		t.Fatal("filled date must not show placeholder")
	}
}

func TestSDTCheckboxW14(t *testing.T) {
	xml := sdtDocumentXML(t, func(sec *element.Section) {
		sec.AddSDTCheckbox("Accept", "accept", true)
		sec.AddSDTCheckbox("Reject", "reject", false)
	})
	first := firstXMLBlock(xml, "w:sdt")
	cb := firstXMLBlock(first, "w14:checkbox")
	if cb == "" {
		t.Fatal("missing w14:checkbox")
	}
	seq := innerChildNames(cb, "w14:checkbox")
	if len(seq) != 3 {
		t.Fatalf("checkbox children %v", seq)
	}
	assertSeq(t, seq, "w14:checked", "w14:checkedState", "w14:uncheckedState")
	if seq[0] != "w14:checked" || seq[1] != "w14:checkedState" || seq[2] != "w14:uncheckedState" {
		t.Fatalf("strict checkbox order %v", seq)
	}
	if !strings.Contains(cb, `w14:checked w14:val="1"`) {
		t.Fatal("checked val")
	}
	if !strings.Contains(cb, `w14:checkedState w14:val="2612"`) || !strings.Contains(cb, `w14:font="MS Gothic"`) {
		t.Fatal("checkedState")
	}
	if !strings.Contains(cb, `w14:uncheckedState w14:val="2610"`) {
		t.Fatal("uncheckedState")
	}
	if !strings.Contains(first, "\u2612") {
		t.Fatal("checked glyph")
	}
	rest := xml[strings.Index(xml, first)+len(first):]
	second := firstXMLBlock(rest, "w:sdt")
	if !strings.Contains(second, `w14:checked w14:val="0"`) {
		t.Fatal("unchecked val")
	}
	if !strings.Contains(second, "\u2610") {
		t.Fatal("unchecked glyph")
	}
}

func TestSDTUniqueIDs(t *testing.T) {
	xml := sdtDocumentXML(t, func(sec *element.Section) {
		sec.AddSDTText("A", "a", "a")
		sec.AddSDTText("B", "b", "b")
		sec.AddSDTText("C", "c", "c")
	})
	seen := map[string]bool{}
	for {
		i := strings.Index(xml, `<w:id w:val="`)
		if i < 0 {
			break
		}
		xml = xml[i+len(`<w:id w:val="`):]
		j := strings.Index(xml, `"`)
		id := xml[:j]
		if seen[id] {
			t.Fatalf("duplicate sdt id %s", id)
		}
		seen[id] = true
	}
	if len(seen) != 3 {
		t.Fatalf("id count %d", len(seen))
	}
}

func TestSDTOnCellAndTextRun(t *testing.T) {
	xml := sdtDocumentXML(t, func(sec *element.Section) {
		tbl := sec.AddTable(style.Table{Width: 5000})
		row := tbl.AddRow()
		cell := row.AddCell(2500)
		cell.AddSDTText("Cell", "cell_tag", "in cell")
		tr := sec.AddTextRun()
		tr.AddText("Inline ")
		tr.AddSDTCheckbox("Flag", "flag", false)
	})
	if !strings.Contains(xml, `<w:tag w:val="cell_tag"`) {
		t.Fatal("cell sdt")
	}
	if !strings.Contains(xml, `<w:tag w:val="flag"`) {
		t.Fatal("textrun sdt")
	}
	// Inline SDT lives inside the TextRun's w:p, not as a sibling block with its own wrapping only.
	if !strings.Contains(xml, "<w:sdt>") {
		t.Fatal("sdt present")
	}
}

func TestSDTPHPWordAddSDTStillWrites(t *testing.T) {
	xml := sdtDocumentXML(t, func(sec *element.Section) {
		s := sec.AddSDT("plainText")
		s.SetAlias("Old").SetTag("old").SetValue("hello")
		dd := sec.AddSDT("dropDownList")
		dd.SetListItems([]element.SDTListItem{{Value: "1", DisplayText: "One"}})
		dd.SetValue("One")
	})
	if !strings.Contains(xml, "<w:text") {
		t.Fatal("plainText")
	}
	if !strings.Contains(xml, "<w:dropDownList") {
		t.Fatal("dropDownList")
	}
	if !strings.Contains(xml, `w:displayText="One"`) {
		t.Fatal("list item")
	}
}
