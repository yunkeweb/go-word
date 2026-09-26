package word

import (
	"strings"
	"testing"

	"github.com/yunkeweb/go-word/element"
)

func unzipMainXML(t *testing.T, tp *TemplateProcessor) string {
	t.Helper()
	raw, err := tp.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	return readZipFile(t, raw, "word/document.xml")
}

func TestTemplatePipeFilters(t *testing.T) {
	tp := mustTemplate(t, func(sec *element.Section) {
		sec.AddText("name=${name | upper}")
		sec.AddText("low=${city | lower}")
		sec.AddText("none=${missing | default:N/A}")
		sec.AddText("quoted=${blank | default:\"N/A\"}")
		sec.AddText("date=${created_at | formatDate:\"2006-01-02\"}")
		sec.AddText("money=${amount | formatCurrency:¥}")
		sec.AddText("chain=${name | lower | upper}")
		sec.AddText("trim=${title | trim | upper}")
		sec.AddText("cut=${blurb | truncate:8}")
	})
	tp.SetValue("name", "alice")
	tp.SetValue("city", "Paris")
	tp.SetValue("created_at", "2026-09-26T10:11:12Z")
	tp.SetValue("amount", "1234.5")
	tp.SetValue("blank", "")
	tp.SetValue("title", "  hello world  ")
	tp.SetValue("blurb", "abcdefghijk")
	xml := unzipMainXML(t, tp)
	for _, want := range []string{"ALICE", "paris", "N/A", "2026-09-26", "¥1234.50", "HELLO WORLD", "abcdefgh"} {
		if !strings.Contains(xml, want) {
			t.Fatalf("missing %q in %s", want, xml)
		}
	}
	if strings.Contains(xml, "${") {
		t.Fatalf("unresolved macro: %s", xml)
	}
	if strings.Contains(xml, "&#") {
		t.Fatalf("numeric entity in filter output: %s", xml)
	}
}

func TestRegisterTemplateFilter(t *testing.T) {
	RegisterTemplateFilter("surround", func(in any, args ...string) string {
		s := stringifyFilterIn(in)
		l, r := "[", "]"
		if len(args) > 0 && args[0] != "" {
			l = args[0]
			r = args[0]
		}
		return l + s + r
	})
	tp := mustTemplate(t, func(sec *element.Section) {
		sec.AddText("${title | surround:*}")
	})
	tp.SetValue("title", "Go")
	xml := unzipMainXML(t, tp)
	if !strings.Contains(xml, "*Go*") {
		t.Fatalf("custom filter: %s", xml)
	}
}

func TestTemplateIfComparisons(t *testing.T) {
	tp := mustTemplate(t, func(sec *element.Section) {
		sec.AddText("x${if age >= 18}ADULT${endif}y")
		sec.AddText("x${if age < 18}MINOR${endif}y")
		sec.AddText("x${if status == active}ON${endif}y")
		sec.AddText("x${if status != active}OFF${endif}y")
		sec.AddText("x${if leftover}HIDE${endif}y")
	})
	tp.SetValue("age", "21")
	tp.SetValue("status", "active")
	xml := unzipMainXML(t, tp)
	if !strings.Contains(xml, "ADULT") {
		t.Fatalf("age>=18 kept: %s", xml)
	}
	if strings.Contains(xml, "MINOR") {
		t.Fatalf("age<18 clipped: %s", xml)
	}
	if !strings.Contains(xml, "ON") || strings.Contains(xml, "OFF") {
		t.Fatalf("status: %s", xml)
	}
	if strings.Contains(xml, "HIDE") || strings.Contains(xml, "${if") {
		t.Fatalf("leftover if: %s", xml)
	}
}

func TestTemplateIfComparisonXMLEntities(t *testing.T) {
	tp := mustTemplate(t, func(sec *element.Section) {
		sec.AddText("${if score >= 60}PASS${endif}")
	})
	main := string(tp.files[tp.mainPart()])
	main = strings.ReplaceAll(main, ">=", "&gt;=")
	tp.files[tp.mainPart()] = []byte(main)
	tp.SetValue("score", "75")
	xml := unzipMainXML(t, tp)
	if !strings.Contains(xml, "PASS") {
		t.Fatalf("entity >= not evaluated: %s", xml)
	}
}

func TestTemplateCloneIndexesPipesAndComparisons(t *testing.T) {
	tp := mustTemplate(t, func(sec *element.Section) {
		sec.AddText("${row}")
		sec.AddText("${name | upper}")
		sec.AddText("${if age >= 18}ok${endif}")
		sec.AddText("${/row}")
	})
	if err := tp.CloneBlock("row", 2); err != nil {
		t.Fatal(err)
	}
	xml := string(tp.files[tp.mainPart()])
	if !strings.Contains(xml, "${name#1 | upper}") || !strings.Contains(xml, "${name#2 | upper}") {
		t.Fatalf("pipe index: %s", xml)
	}
	if !strings.Contains(xml, "${if age#1 >= 18}") || !strings.Contains(xml, "${if age#2 >= 18}") {
		t.Fatalf("if index: %s", xml)
	}
	tp.SetValue("name#1", "ann")
	tp.SetValue("name#2", "bob")
	tp.SetValue("age#1", "20")
	tp.SetValue("age#2", "16")
	out := unzipMainXML(t, tp)
	if !strings.Contains(out, "ANN") || !strings.Contains(out, "BOB") {
		t.Fatalf("piped values: %s", out)
	}
	if strings.Count(out, "ok") != 1 {
		t.Fatalf("want one ok: %s", out)
	}
}

func TestTemplateTrimTruncateFilters(t *testing.T) {
	if got := filterTrim("  Hi  "); got != "Hi" {
		t.Fatalf("trim: %q", got)
	}
	if got := filterTruncate("abcdefgh", "3"); got != "abc" {
		t.Fatalf("truncate: %q", got)
	}
	if got := filterTruncate("ab", "8"); got != "ab" {
		t.Fatalf("truncate short: %q", got)
	}
	if got := filterTruncate("abcdef", "-1"); got != "abcdef" {
		t.Fatalf("truncate invalid: %q", got)
	}
	if got := filterTruncate("你好世界", "2"); got != "你好" {
		t.Fatalf("truncate runes: %q", got)
	}
	if got := applyFilters("  title  ", []pipeCall{{name: "trim"}, {name: "upper"}}); got != "TITLE" {
		t.Fatalf("chain trim|upper: %q", got)
	}
}

func TestTemplateDefaultFilterWithoutSetValue(t *testing.T) {
	tp := mustTemplate(t, func(sec *element.Section) {
		sec.AddText("${nobody | default:N/A}")
	})
	xml := unzipMainXML(t, tp)
	if !strings.Contains(xml, "N/A") {
		t.Fatalf("default: %s", xml)
	}
}
