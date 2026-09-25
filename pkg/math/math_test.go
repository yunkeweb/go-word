package math

import (
	"bytes"
	"strings"
	"testing"
)

func TestConstructorsAndGroup(t *testing.T) {
	m := New()
	id := NewIdentifier("x")
	num := NewNumeric("1")
	op := NewOperator("+")
	m.Add(id)
	m.Add(num)
	m.Add(op)
	m.Remove(num)
	if len(m.Elements) != 2 {
		t.Fatalf("len=%d", len(m.Elements))
	}
	m.Remove(NewIdentifier("missing"))
	if len(m.Elements) != 2 {
		t.Fatal("remove missing")
	}
	sem := NewSemantics()
	if sem.Annotation("tex") != "" {
		t.Fatal("empty annotation")
	}
	sem.Annotations = nil
	if sem.Annotation("tex") != "" {
		t.Fatal("nil map")
	}
	sem.AddAnnotation("application/x-tex", "x")
	if sem.Annotation("application/x-tex") != "x" {
		t.Fatal("annotation")
	}
	frac := NewFraction(NewNumeric("1"), NewNumeric("2"))
	sup := NewSuperscript(NewIdentifier("x"), NewNumeric("2"))
	row := NewRow()
	row.Add(frac)
	row.Add(sup)
	id.isMathElement()
}

func TestWriteMathMLAllElements(t *testing.T) {
	m := New()
	row := NewRow()
	row.Add(NewIdentifier("x"))
	row.Add(NewOperator("+"))
	row.Add(NewNumeric("1"))
	m.Add(row)
	m.Add(NewFraction(NewNumeric("1"), nil))
	m.Add(NewFraction(nil, NewNumeric("2")))
	m.Add(NewSuperscript(NewIdentifier("x"), nil))
	m.Add(NewSuperscript(nil, NewNumeric("2")))
	inner := New()
	inner.Add(NewIdentifier("y"))
	m.Add(inner)
	g := &Group{}
	g.Add(NewIdentifier("g"))
	m.Add(g)
	sem := NewSemantics()
	sem.Add(NewIdentifier("s"))
	sem.AddAnnotation("tex", "s")
	m.Add(sem)
	b, err := WriteMathML(m)
	if err != nil {
		t.Fatal(err)
	}
	s := string(b)
	for _, need := range []string{"<math", "mrow", "mi", "mo", "mn", "mfrac", "msup", "semantics", "annotation"} {
		if !strings.Contains(s, need) {
			t.Fatalf("missing %q in %s", need, s)
		}
	}
}

func TestReadMathMLVariants(t *testing.T) {
	src := `<math xmlns="http://www.w3.org/1998/Math/MathML">
		<mrow><mi>x</mi><mo>+</mo><mn>1</mn></mrow>
		<mfrac><mn>1</mn><mn>2</mn></mfrac>
		<msup><mi>x</mi><mn>2</mn></msup>
		<semantics><mi>y</mi><annotation encoding="tex">y</annotation></semantics>
	</math>`
	m, err := ReadMathML(bytes.NewReader([]byte(src)))
	if err != nil {
		t.Fatal(err)
	}
	if len(m.Elements) < 3 {
		t.Fatalf("len=%d", len(m.Elements))
	}
	if _, err := ReadMathML(bytes.NewReader([]byte(`<math></math>`))); err == nil {
		t.Fatal("empty MathML")
	}
	if _, err := ReadMathML(bytes.NewReader([]byte(`<math><`))); err == nil {
		t.Fatal("invalid xml")
	}
	onlyLeaf := `<math><mn>9</mn></math>`
	m2, err := ReadMathML(bytes.NewReader([]byte(onlyLeaf)))
	if err != nil || len(m2.Elements) == 0 {
		t.Fatalf("leaf %v %v", m2, err)
	}
}

func TestWriteReadOMMLVariants(t *testing.T) {
	m := New()
	m.Add(NewRow())
	row := NewRow()
	row.Add(NewIdentifier("x"))
	m.Add(row)
	m.Add(NewFraction(NewNumeric("1"), NewNumeric("2")))
	m.Add(NewFraction(nil, nil))
	m.Add(NewSuperscript(NewIdentifier("x"), NewNumeric("2")))
	m.Add(NewSuperscript(nil, nil))
	m.Add(NewOperator("+"))
	sem := NewSemantics()
	sem.Add(NewIdentifier("s"))
	m.Add(sem)
	inner := New()
	inner.Add(NewNumeric("3"))
	m.Add(inner)
	b, err := WriteOMML(m)
	if err != nil {
		t.Fatal(err)
	}
	s := string(b)
	if !strings.Contains(s, "m:f") || !strings.Contains(s, "m:sSup") {
		t.Fatalf("%s", s)
	}

	src := `<m:oMathPara xmlns:m="http://schemas.openxmlformats.org/officeDocument/2006/math">
		<m:oMath>
			<m:f><m:num><m:r><m:t>1</m:t></m:r></m:num><m:den><m:r><m:t>2</m:t></m:r></m:den></m:f>
			<m:sSup><m:e><m:r><m:t>x</m:t></m:r></m:e><m:sup><m:r><m:t>2</m:t></m:r></m:sup></m:sSup>
			<m:r><m:t>abc</m:t></m:r>
			<m:r><m:t>-3.5</m:t></m:r>
			<m:unknown/>
		</m:oMath>
	</m:oMathPara>`
	got, err := ReadOMML(bytes.NewReader([]byte(src)))
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Elements) == 0 {
		t.Fatal("empty omml")
	}
	if _, err := ReadOMML(bytes.NewReader([]byte(`<m:oMath><`))); err == nil {
		t.Fatal("invalid omml")
	}
	if _, err := ReadOMML(bytes.NewReader([]byte(`<m:oMath><m:t>`))); err == nil {
		t.Fatal("truncated t")
	}
	nested := `<m:oMathPara xmlns:m="http://schemas.openxmlformats.org/officeDocument/2006/math">
		<m:oMath>
			<m:r><m:t>x</m:t></m:r>
			<m:f><m:num><m:f><m:num><m:r><m:t>1</m:t></m:r></m:num><m:den><m:r><m:t>2</m:t></m:r></m:den></m:f></m:num><m:den><m:r><m:t>3</m:t></m:r></m:den></m:f>
		</m:oMath>
	</m:oMathPara>`
	if _, err := ReadOMML(bytes.NewReader([]byte(nested))); err != nil {
		t.Fatal(err)
	}
	empty, err := ReadOMML(bytes.NewReader([]byte(`<m:oMathPara><m:oMath></m:oMath></m:oMathPara>`)))
	if err != nil || empty == nil {
		t.Fatalf("empty tree %v %v", empty, err)
	}
}

func TestLooksNumeric(t *testing.T) {
	if looksNumeric("") || looksNumeric("a1") || !looksNumeric("1.2") || !looksNumeric("-3") {
		t.Fatal("looksNumeric")
	}
}

func TestReplaceHelpers(t *testing.T) {
	m := New()
	old := NewIdentifier("old")
	m.Add(old)
	replaceLast(m, old, NewIdentifier("new"))
	replaceLast(m, NewIdentifier("missing"), NewIdentifier("added"))
	if len(m.Elements) != 2 {
		t.Fatalf("len=%d", len(m.Elements))
	}

	row := NewRow()
	a := NewIdentifier("a")
	row.Add(a)
	replaceChild(row, a, NewIdentifier("b"))
	replaceChild(row, NewIdentifier("x"), NewIdentifier("y"))

	sem := NewSemantics()
	c := NewIdentifier("c")
	sem.Add(c)
	replaceChild(sem, c, NewIdentifier("d"))

	g := &Group{}
	e := NewIdentifier("e")
	g.Add(e)
	replaceChild(g, e, NewIdentifier("f"))

	inner := New()
	z := NewIdentifier("z")
	inner.Add(z)
	replaceChild(inner, z, NewIdentifier("zz"))

	frac := NewFraction(NewNumeric("1"), NewNumeric("2"))
	replaceChild(frac, frac.Numerator, NewNumeric("3"))
	replaceChild(frac, frac.Denominator, NewNumeric("4"))
	sup := NewSuperscript(NewIdentifier("x"), NewNumeric("2"))
	replaceChild(sup, sup.Base, NewIdentifier("y"))
	replaceChild(sup, sup.Sup, NewNumeric("3"))

	attachChild(row, NewIdentifier("k"))
	attachChild(sem, NewIdentifier("k"))
	attachChild(inner, NewIdentifier("k"))
	attachChild(g, NewIdentifier("k"))
	f2 := NewFraction(nil, nil)
	attachChild(f2, NewNumeric("1"))
	attachChild(f2, NewNumeric("2"))
	attachChild(f2, NewNumeric("3"))
	s2 := NewSuperscript(nil, nil)
	attachChild(s2, NewIdentifier("a"))
	attachChild(s2, NewIdentifier("b"))
	attachChild(s2, NewIdentifier("c"))

	if leafFromTag("mn", "1", nil).(*Numeric).Value != "1" {
		t.Fatal("mn")
	}
	if leafFromTag("mo", "+", nil).(*Operator).Value != "+" {
		t.Fatal("mo")
	}
	if leafFromTag("annotation", "a", nil).(*Identifier).Value != "a" {
		t.Fatal("ann")
	}
	if attrValue(nil, "x") != "" {
		t.Fatal("attr")
	}
}
