package word

import (
	"strings"
	"testing"

	"github.com/yunkeweb/go-word/element"
)

func TestTemplateSetValueAndClone(t *testing.T) {
	src := New()
	sec := src.AddSection()
	sec.AddText("Hello ${name}")
	tbl := sec.AddTable()
	row := tbl.AddRow()
	row.AddCell(2000).AddText("${item}")
	row.AddCell(2000).AddText("${qty}")
	sec.AddText("${blk}X${/blk}")
	b, err := src.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	tp, err := NewTemplateProcessorBytes(b)
	if err != nil {
		t.Fatal(err)
	}
	tp.SetValue("name", "World")
	if err := tp.CloneRow("item", 2); err != nil {
		t.Fatal(err)
	}
	tp.SetValue("item#1", "Apple")
	tp.SetValue("qty#1", "3")
	tp.SetValue("item#2", "Pear")
	tp.SetValue("qty#2", "1")
	if err := tp.CloneBlock("blk", 2); err != nil {
		t.Fatal(err)
	}
	out, err := tp.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	loaded, err := LoadBytes(out)
	if err != nil {
		t.Fatal(err)
	}
	var all strings.Builder
	for _, s := range loaded.Sections() {
		for _, e := range s.Elements() {
			all.WriteString(elementText(e))
			all.WriteByte(' ')
		}
	}
	got := all.String()
	if !strings.Contains(got, "Hello World") {
		t.Fatalf("placeholder not replaced: %q", got)
	}
}

func TestTemplateVariablesCloneAndDelete(t *testing.T) {
	src := New()
	sec := src.AddSection()
	sec.AddText("Hi ${name}")
	tbl := sec.AddTable()
	row := tbl.AddRow()
	row.AddCell(2000).AddText("${item}")
	b, err := src.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	tp, err := NewTemplateProcessorBytes(b)
	if err != nil {
		t.Fatal(err)
	}
	vars := tp.GetVariables()
	if len(vars) < 2 {
		t.Fatalf("vars=%v", vars)
	}
	counts := tp.GetVariableCount()
	if counts["name"] < 1 {
		t.Fatalf("counts=%v", counts)
	}
	if err := tp.CloneRowAndSetValues("item", []map[string]string{
		{"item": "Apple"},
		{"item": "Pear"},
	}); err != nil {
		t.Fatal(err)
	}
	out, err := tp.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	loaded, err := LoadBytes(out)
	if err != nil {
		t.Fatal(err)
	}
	var got strings.Builder
	walkDocument(loaded, func(el element.Element) {
		got.WriteString(elementText(el))
		got.WriteByte(' ')
	})
	if !strings.Contains(got.String(), "Apple") || !strings.Contains(got.String(), "Pear") {
		t.Fatalf("clone values missing: %q", got.String())
	}

	tp2, err := NewTemplateProcessorBytes(b)
	if err != nil {
		t.Fatal(err)
	}
	if err := tp2.DeleteRow("item"); err != nil {
		t.Fatal(err)
	}
	out2, err := tp2.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	loaded2, err := LoadBytes(out2)
	if err != nil {
		t.Fatal(err)
	}
	var got2 strings.Builder
	walkDocument(loaded2, func(el element.Element) {
		got2.WriteString(elementText(el))
	})
	if strings.Contains(got2.String(), "${item}") {
		t.Fatalf("row not deleted: %q", got2.String())
	}
}

func TestTemplateSetChart(t *testing.T) {
	src := New()
	sec := src.AddSection()
	sec.AddText("See ${chart}")
	b, err := src.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	tp, err := NewTemplateProcessorBytes(b)
	if err != nil {
		t.Fatal(err)
	}
	ch := &element.Chart{ChartType: "pie", Categories: []string{"A"}, Values: []float64{1}}
	if err := tp.SetChart("chart", ch); err != nil {
		t.Fatal(err)
	}
	out, err := tp.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(out), "word/charts/chart") {
		t.Fatal("chart part not embedded")
	}
	cr := tp.ReplaceCarriageReturns("a\nb")
	if !strings.Contains(cr, "<w:br/>") {
		t.Fatal("carriage returns")
	}
}
