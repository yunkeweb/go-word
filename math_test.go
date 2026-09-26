package word

import (
	"strings"
	"testing"
)

func TestAddMathWritesNativeOMML(t *testing.T) {
	doc := New()
	doc.AddMath(`\frac{a}{b}`)
	sec := doc.AddSection()
	sec.AddMath(`x^{2} + y^{2} = z^{2}`)
	p := sec.AddTextRun()
	p.AddText("inline ")
	p.AddMath(`\sqrt{x_1} + \pi`)
	raw, err := doc.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	xml := readZipFile(t, raw, "word/document.xml")
	assertWellFormedXML(t, "math-doc", xml)
	if !strings.Contains(xml, "<m:oMathPara") {
		t.Fatal("display oMathPara")
	}
	if !strings.Contains(xml, "<m:f>") || !strings.Contains(xml, "<m:num>") {
		t.Fatal("fraction")
	}
	if !strings.Contains(xml, "<m:sSup>") {
		t.Fatal("superscript")
	}
	if !strings.Contains(xml, "<m:rad>") || !strings.Contains(xml, "<m:sSub>") {
		t.Fatal("radical/subscript")
	}
	if strings.Count(xml, "<m:oMath>") < 3 {
		t.Fatal("display + inline oMath")
	}
}

func TestAddMathEmptyFallsBack(t *testing.T) {
	doc := New()
	f := doc.AddMath("")
	if f == nil || f.Math == nil {
		t.Fatal("formula")
	}
	raw, err := doc.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	xml := readZipFile(t, raw, "word/document.xml")
	if !strings.Contains(xml, "<m:oMath>") {
		t.Fatal("empty math still writes oMath")
	}
}
