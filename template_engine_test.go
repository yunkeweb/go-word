package word

import (
	"strings"
	"testing"

	"github.com/yunkeweb/go-word/element"
	"github.com/yunkeweb/go-word/style"
)

func mustTemplate(t *testing.T, build func(*element.Section)) *TemplateProcessor {
	t.Helper()
	doc := New()
	build(doc.AddSection())
	b, err := doc.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	tp, err := NewTemplateProcessorBytes(b)
	if err != nil {
		t.Fatal(err)
	}
	return tp
}

func TestCloneBlockNestedInnerIndexed(t *testing.T) {
	tp := mustTemplate(t, func(sec *element.Section) {
		sec.AddText("${orders}")
		sec.AddText("Order ${oid}")
		sec.AddText("${items}")
		sec.AddText("${item}")
		sec.AddText("${/items}")
		sec.AddText("${/orders}")
	})
	if err := tp.CloneBlock("orders", 2); err != nil {
		t.Fatal(err)
	}
	xml := string(tp.files[tp.mainPart()])
	for _, need := range []string{"${oid#1}", "${oid#2}", "${items#1}", "${items#2}", "${/items#1}", "${/items#2}", "${item#1}", "${item#2}"} {
		if !strings.Contains(xml, need) {
			t.Fatalf("missing %s in\n%s", need, xml)
		}
	}
	if err := tp.CloneBlock("items#1", 2); err != nil {
		t.Fatal(err)
	}
	if err := tp.CloneBlock("items#2", 1); err != nil {
		t.Fatal(err)
	}
	tp.SetValue("oid#1", "A")
	tp.SetValue("oid#2", "B")
	tp.SetValue("item#1#1", "apple")
	tp.SetValue("item#1#2", "pear")
	tp.SetValue("item#2#1", "tea")
	xml = string(tp.files[tp.mainPart()])
	for _, need := range []string{"Order A", "Order B", "apple", "pear", "tea"} {
		if !strings.Contains(xml, need) {
			t.Fatalf("missing %s in\n%s", need, xml)
		}
	}
	if strings.Contains(xml, "${") {
		t.Fatalf("leftover macros:\n%s", xml)
	}
}

func TestCloneNestedBlockHierarchy(t *testing.T) {
	tp := mustTemplate(t, func(sec *element.Section) {
		sec.AddText("${orders}")
		sec.AddText("Order ${oid}")
		sec.AddText("${items}")
		sec.AddText("${item}:${qty}")
		sec.AddText("${/items}")
		sec.AddText("${/orders}")
	})
	err := tp.CloneNestedBlock("orders", []BlockData{
		{
			Values: map[string]string{"oid": "A"},
			Blocks: map[string][]BlockData{
				"items": {
					{Values: map[string]string{"item": "apple", "qty": "2"}},
					{Values: map[string]string{"item": "pear", "qty": "1"}},
				},
			},
		},
		{
			Values: map[string]string{"oid": "B"},
			Blocks: map[string][]BlockData{
				"items": {
					{Values: map[string]string{"item": "tea", "qty": "9"}},
				},
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	xml := string(tp.files[tp.mainPart()])
	for _, need := range []string{"Order A", "Order B", "apple:2", "pear:1", "tea:9"} {
		if !strings.Contains(xml, need) {
			t.Fatalf("missing %s in\n%s", need, xml)
		}
	}
}

func TestCloneNestedBlockEmptyDeletes(t *testing.T) {
	tp := mustTemplate(t, func(sec *element.Section) {
		sec.AddText("before")
		sec.AddText("${blk}gone${/blk}")
		sec.AddText("after")
	})
	if err := tp.CloneNestedBlock("blk", nil); err != nil {
		t.Fatal(err)
	}
	xml := string(tp.files[tp.mainPart()])
	if strings.Contains(xml, "gone") || strings.Contains(xml, "${blk}") {
		t.Fatalf("block not deleted: %s", xml)
	}
	if !strings.Contains(xml, "before") || !strings.Contains(xml, "after") {
		t.Fatalf("neighbors lost: %s", xml)
	}
}

func TestSetConditionFalseRemovesParagraphs(t *testing.T) {
	tp := mustTemplate(t, func(sec *element.Section) {
		sec.AddText("keep")
		sec.AddText("${if show}")
		sec.AddText("secret")
		sec.AddText("${endif}")
		sec.AddText("tail")
	})
	if err := tp.SetCondition("show", false); err != nil {
		t.Fatal(err)
	}
	xml := string(tp.files[tp.mainPart()])
	if strings.Contains(xml, "secret") || strings.Contains(xml, "${if") || strings.Contains(xml, "${endif}") {
		t.Fatalf("false block not clipped: %s", xml)
	}
	if !strings.Contains(xml, "keep") || !strings.Contains(xml, "tail") {
		t.Fatalf("neighbors lost: %s", xml)
	}
}

func TestSetConditionTrueKeepsInnerStripsMarkers(t *testing.T) {
	tp := mustTemplate(t, func(sec *element.Section) {
		sec.AddText("${if show}")
		sec.AddText("visible")
		sec.AddText("${endif}")
	})
	if err := tp.SetCondition("show", true); err != nil {
		t.Fatal(err)
	}
	xml := string(tp.files[tp.mainPart()])
	if !strings.Contains(xml, "visible") {
		t.Fatalf("content lost: %s", xml)
	}
	if strings.Contains(xml, "${if") || strings.Contains(xml, "${endif}") {
		t.Fatalf("markers remain: %s", xml)
	}
}

func TestSetConditionSameParagraph(t *testing.T) {
	tp := mustTemplate(t, func(sec *element.Section) {
		sec.AddText("A ${if x}hidden${endif} B")
	})
	if err := tp.SetCondition("x", false); err != nil {
		t.Fatal(err)
	}
	xml := string(tp.files[tp.mainPart()])
	if strings.Contains(xml, "hidden") {
		t.Fatalf("inline not removed: %s", xml)
	}
	if !strings.Contains(xml, "A ") || !strings.Contains(xml, " B") {
		t.Fatalf("neighbors lost: %s", xml)
	}

	tp2 := mustTemplate(t, func(sec *element.Section) {
		sec.AddText("A ${if x}shown${endif} B")
	})
	if err := tp2.SetCondition("x", true); err != nil {
		t.Fatal(err)
	}
	xml2 := string(tp2.files[tp2.mainPart()])
	if !strings.Contains(xml2, "A shown B") {
		t.Fatalf("inline keep failed: %s", xml2)
	}
}

func TestSetConditionTableRows(t *testing.T) {
	tp := mustTemplate(t, func(sec *element.Section) {
		tbl := sec.AddTable()
		tbl.AddRow().AddCell(1000).AddText("head")
		tbl.AddRow().AddCell(1000).AddText("${if extra}")
		tbl.AddRow().AddCell(1000).AddText("bonus")
		tbl.AddRow().AddCell(1000).AddText("${endif}")
		tbl.AddRow().AddCell(1000).AddText("foot")
	})
	if err := tp.SetCondition("extra", false); err != nil {
		t.Fatal(err)
	}
	xml := string(tp.files[tp.mainPart()])
	if strings.Contains(xml, "bonus") {
		t.Fatalf("row block not clipped: %s", xml)
	}
	if !strings.Contains(xml, "head") || !strings.Contains(xml, "foot") {
		t.Fatalf("table neighbors lost: %s", xml)
	}
}

func TestNestedIfBlocks(t *testing.T) {
	tp := mustTemplate(t, func(sec *element.Section) {
		sec.AddText("${if outer}")
		sec.AddText("A")
		sec.AddText("${if inner}")
		sec.AddText("B")
		sec.AddText("${endif}")
		sec.AddText("C")
		sec.AddText("${endif}")
	})
	if err := tp.SetCondition("inner", false); err != nil {
		t.Fatal(err)
	}
	if err := tp.SetCondition("outer", true); err != nil {
		t.Fatal(err)
	}
	xml := string(tp.files[tp.mainPart()])
	if strings.Contains(xml, ">B</w:t>") || strings.Contains(xml, ">B<") {
		t.Fatalf("inner not clipped: %s", xml)
	}
	if !strings.Contains(xml, "A") || !strings.Contains(xml, "C") {
		t.Fatalf("outer content lost: %s", xml)
	}
}

func TestSetConditionsMapAndMissing(t *testing.T) {
	tp := mustTemplate(t, func(sec *element.Section) {
		sec.AddText("${if a}AA${endif}")
		sec.AddText("${if b}BB${endif}")
	})
	if err := tp.SetConditions(map[string]bool{"a": true, "b": false}); err != nil {
		t.Fatal(err)
	}
	xml := string(tp.files[tp.mainPart()])
	if !strings.Contains(xml, "AA") || strings.Contains(xml, "BB") {
		t.Fatalf("map apply: %s", xml)
	}
	if err := tp.SetCondition("missing", true); err == nil {
		t.Fatal("expected missing if")
	}
}

func TestApplyConditionsFromValues(t *testing.T) {
	tp := mustTemplate(t, func(sec *element.Section) {
		sec.AddText("${if one}ONE${endif}")
		sec.AddText("${if zero}ZERO${endif}")
		sec.AddText("${if no}NO${endif}")
		sec.AddText("${if empty}EMPTY${endif}")
		sec.AddText("${if yes}YES${endif}")
	})
	if err := tp.ApplyConditionsFromValues(map[string]string{
		"one":   "1",
		"zero":  "0",
		"no":    "false",
		"empty": "",
		"yes":   "true",
	}); err != nil {
		t.Fatal(err)
	}
	xml := string(tp.files[tp.mainPart()])
	if !strings.Contains(xml, "ONE") || !strings.Contains(xml, "YES") {
		t.Fatalf("truthy dropped: %s", xml)
	}
	for _, ban := range []string{"ZERO", "NO", "EMPTY"} {
		if strings.Contains(xml, ban) {
			t.Fatalf("%s should be clipped: %s", ban, xml)
		}
	}
}

func TestRemainingIfBlocksRemovedOnBytes(t *testing.T) {
	tp := mustTemplate(t, func(sec *element.Section) {
		sec.AddText("x${if leftover}HIDE${endif}y")
	})
	out, err := tp.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(out), "HIDE") || strings.Contains(string(out), "${if") {
		t.Fatal("leftover if not auto-clipped")
	}
}

func TestCloneRowVerticalMerge(t *testing.T) {
	tp := mustTemplate(t, func(sec *element.Section) {
		tbl := sec.AddTable()
		r1 := tbl.AddRow()
		r1.AddCell(2000, style.Cell{VMerge: "restart"}).AddText("${item}")
		r1.AddCell(2000).AddText("${qty}")
		r2 := tbl.AddRow()
		r2.AddCell(2000, style.Cell{VMerge: "continue"})
		r2.AddCell(2000).AddText("note")
	})
	if err := tp.CloneRow("item", 2); err != nil {
		t.Fatal(err)
	}
	xml := string(tp.files[tp.mainPart()])
	if strings.Count(xml, "<w:tr>") != 4 {
		t.Fatalf("want 4 rows, xml=\n%s", xml)
	}
	if strings.Count(xml, `w:val="restart"`) != 2 {
		t.Fatalf("want 2 restarts: %s", xml)
	}
	if !strings.Contains(xml, "${item#1}") || !strings.Contains(xml, "${item#2}") {
		t.Fatalf("indexed macros missing: %s", xml)
	}
	if strings.Count(xml, "${qty#1}") != 1 || strings.Count(xml, "note") < 2 {
		t.Fatalf("merged group not fully cloned: %s", xml)
	}
}

func TestCloneRowGridSpan(t *testing.T) {
	tp := mustTemplate(t, func(sec *element.Section) {
		tbl := sec.AddTable()
		r := tbl.AddRow()
		r.AddCell(4000, style.Cell{GridSpan: 2}).AddText("${item}")
		r.AddCell(2000).AddText("${qty}")
	})
	if err := tp.CloneRow("item", 3); err != nil {
		t.Fatal(err)
	}
	xml := string(tp.files[tp.mainPart()])
	if strings.Count(xml, `w:val="2"`) < 3 {
		t.Fatalf("gridSpan not cloned: %s", xml)
	}
	if !strings.Contains(xml, "${item#3}") {
		t.Fatalf("third clone missing: %s", xml)
	}
}

func TestDeleteRowVerticalMergeGroup(t *testing.T) {
	tp := mustTemplate(t, func(sec *element.Section) {
		tbl := sec.AddTable()
		tbl.AddRow().AddCell(1000).AddText("keep")
		r1 := tbl.AddRow()
		r1.AddCell(1000, style.Cell{VMerge: "restart"}).AddText("${item}")
		r2 := tbl.AddRow()
		r2.AddCell(1000, style.Cell{VMerge: "continue"}).AddText("span")
		tbl.AddRow().AddCell(1000).AddText("tail")
	})
	if err := tp.DeleteRow("item"); err != nil {
		t.Fatal(err)
	}
	xml := string(tp.files[tp.mainPart()])
	if strings.Contains(xml, "${item}") || strings.Contains(xml, "span") {
		t.Fatalf("merged group not deleted: %s", xml)
	}
	if !strings.Contains(xml, "keep") || !strings.Contains(xml, "tail") {
		t.Fatalf("neighbors lost: %s", xml)
	}
}

func TestCloneBlockSameNameNestedMatches(t *testing.T) {
	tp := mustTemplate(t, func(sec *element.Section) {
		sec.AddText("${a}outer-start ${a}inner${/a} outer-end${/a}")
	})
	if err := tp.CloneBlock("a", 1); err != nil {
		t.Fatal(err)
	}
	xml := string(tp.files[tp.mainPart()])
	if !strings.Contains(xml, "outer-start") || !strings.Contains(xml, "outer-end") {
		t.Fatalf("outer inner lost: %s", xml)
	}
}

func TestGetVariablesSkipsControlMacros(t *testing.T) {
	tp := mustTemplate(t, func(sec *element.Section) {
		sec.AddText("${name}")
		sec.AddText("${if show}x${endif}")
		sec.AddText("${blk}y${/blk}")
	})
	vars := tp.GetVariables()
	for _, v := range vars {
		if strings.HasPrefix(v, "if ") || v == "endif" || strings.HasPrefix(v, "/") {
			t.Fatalf("control macro leaked: %v", vars)
		}
	}
	if !containsStr(vars, "name") || !containsStr(vars, "blk") {
		t.Fatalf("vars=%v", vars)
	}
}

func containsStr(ss []string, want string) bool {
	for _, s := range ss {
		if s == want {
			return true
		}
	}
	return false
}
