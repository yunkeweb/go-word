package math

import (
	"bytes"
	"strings"
	"testing"
)

func TestWriteOMMLFraction(t *testing.T) {
	m := New()
	m.Add(NewFraction(NewNumeric("1"), NewNumeric("2")))
	b, err := WriteOMML(m)
	if err != nil {
		t.Fatal(err)
	}
	s := string(b)
	for _, need := range []string{"m:oMathPara", "m:f", "m:num", "m:den", ">1<", ">2<"} {
		if !strings.Contains(s, need) {
			t.Fatalf("missing %q in %s", need, s)
		}
	}
}

func TestReadOMMLRoundTrip(t *testing.T) {
	src := `<m:oMathPara xmlns:m="http://schemas.openxmlformats.org/officeDocument/2006/math"><m:oMath><m:r><m:t>x</m:t></m:r></m:oMath></m:oMathPara>`
	m, err := ReadOMML(bytes.NewReader([]byte(src)))
	if err != nil {
		t.Fatal(err)
	}
	if len(m.Elements) == 0 {
		t.Fatal("no elements")
	}
}

func TestWriteReadMathML(t *testing.T) {
	m := New()
	sem := NewSemantics()
	sem.Add(NewFraction(NewNumeric("1"), NewNumeric("2")))
	sem.AddAnnotation("application/x-tex", "1/2")
	m.Add(sem)
	b, err := WriteMathML(m)
	if err != nil {
		t.Fatal(err)
	}
	s := string(b)
	for _, need := range []string{"<math", "mfrac", "mn", "semantics"} {
		if !strings.Contains(s, need) {
			t.Fatalf("missing %q in %s", need, s)
		}
	}
	got, err := ReadMathML(bytes.NewReader(b))
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Elements) == 0 {
		t.Fatal("empty MathML tree")
	}
}

func TestGroupRemove(t *testing.T) {
	r := NewRow()
	a := NewIdentifier("a")
	b := NewIdentifier("b")
	r.Add(a)
	r.Add(b)
	r.Remove(a)
	if len(r.Elements) != 1 {
		t.Fatalf("len=%d", len(r.Elements))
	}
}
