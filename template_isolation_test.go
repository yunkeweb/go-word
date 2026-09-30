package word

import (
	"strings"
	"sync"
	"testing"
)

func TestTemplateDelimiterIsolation(t *testing.T) {
	doc := New()
	doc.AddSection().AddText("{{block}}{{name | upper}}{{/block}}{{if active}}YES{{endif}}")
	raw, err := doc.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	first, err := NewTemplateProcessorBytes(raw)
	if err != nil {
		t.Fatal(err)
	}
	second, err := NewTemplateProcessorBytes(raw)
	if err != nil {
		t.Fatal(err)
	}
	first.SetMacroChars("{{", "}}")
	second.SetMacroChars("[[", "]]")
	if err := first.CloneBlock("block", 2); err != nil {
		t.Fatal(err)
	}
	first.SetValue("name#1", "one")
	first.SetValue("name#2", "two")
	if err := first.SetCondition("active", true); err != nil {
		t.Fatal(err)
	}
	if _, err := first.Bytes(); err != nil {
		t.Fatal(err)
	}
	xml := string(first.files[first.mainPart()])
	if !strings.Contains(xml, "ONETWOYES") || strings.Contains(xml, "{{") {
		t.Fatalf("incorrect custom delimiter output: %s", xml)
	}
}

func TestTemplateGlobalDefaultsAreSnapshots(t *testing.T) {
	defer SetMacroChars("${", "}")
	SetMacroChars("{{", "}}")
	doc := New()
	doc.AddSection().AddText("{{name}}")
	raw, err := doc.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	before, err := NewTemplateProcessorBytes(raw)
	if err != nil {
		t.Fatal(err)
	}
	SetMacroChars("[[", "]]")
	before.SetValue("name", "Alice")
	if strings.Contains(string(before.files[before.mainPart()]), "{{name}}") {
		t.Fatal("changing defaults affected existing processor")
	}
}

func TestTemplateSeparateInstancesConcurrent(t *testing.T) {
	doc := New()
	doc.AddSection().AddText("{{name}} [[name]]")
	raw, err := doc.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	a, err := NewTemplateProcessorBytes(raw)
	if err != nil {
		t.Fatal(err)
	}
	b, err := NewTemplateProcessorBytes(raw)
	if err != nil {
		t.Fatal(err)
	}
	a.SetMacroChars("{{", "}}")
	b.SetMacroChars("[[", "]]")
	var wg sync.WaitGroup
	for _, tp := range []*TemplateProcessor{a, b} {
		wg.Add(1)
		go func(tp *TemplateProcessor) {
			defer wg.Done()
			for i := 0; i < 50; i++ {
				tp.GetVariables()
				tp.GetVariableCount()
			}
			tp.SetValue("name", "filled")
		}(tp)
	}
	wg.Wait()
	if strings.Contains(string(a.files[a.mainPart()]), "{{name}}") || strings.Contains(string(b.files[b.mainPart()]), "[[name]]") {
		t.Fatal("instance delimiter conflict")
	}
}
