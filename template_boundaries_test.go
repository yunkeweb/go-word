package word

import (
	"strings"
	"testing"

	"github.com/yunkeweb/go-word/element"
)

func TestTemplateBlockMismatchReturnsError(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{name: "missing close", body: "${items}row"},
		{name: "mismatched close", body: "${items}row${/other}"},
		{name: "extra close", body: "row${/items}"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tp := mustTemplate(t, func(sec *element.Section) { sec.AddText(tc.body) })
			if err := tp.CloneBlock("items", 1); err == nil {
				t.Fatal("expected malformed block error")
			}
		})
	}
}

func TestTemplateReplacesHeaderFooterAndKeepsXMLValid(t *testing.T) {
	d := New()
	sec := d.AddSection()
	sec.AddHeader().AddText("Header ${name}")
	sec.AddFooter().AddText("Footer ${name}")
	sec.AddText("Body ${name}")
	raw, err := d.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	tp, err := NewTemplateProcessorBytes(raw)
	if err != nil {
		t.Fatal(err)
	}
	tp.SetValue("name", "Ada & Co")
	out, err := tp.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	for _, diag := range ValidatePackage(out) {
		if diag.Severity == "error" {
			t.Fatalf("invalid XML after replacement: %+v", diag)
		}
	}
	parts := []string{"word/document.xml"}
	for _, rel := range parseRelationshipDetails([]byte(readZipFile(t, out, "word/_rels/document.xml.rels"))) {
		if strings.HasSuffix(rel.Type, "/header") || strings.HasSuffix(rel.Type, "/footer") {
			parts = append(parts, "word/"+rel.Target)
		}
	}
	if len(parts) != 3 {
		t.Fatalf("expected body, header and footer, got %v", parts)
	}
	for _, part := range parts {
		if !strings.Contains(string(readZipFile(t, out, part)), "Ada &amp; Co") {
			t.Fatalf("replacement missing from %s", part)
		}
	}
	loaded, err := LoadBytes(out)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(loaded.ExtractText(), "Ada & Co") {
		t.Fatalf("body replacement missing: %q", loaded.ExtractText())
	}
}
