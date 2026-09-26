package math

// Element is a node in an Office Math expression tree.
type Element interface {
	isMathElement()
}

type elementBase struct{}

func (elementBase) isMathElement() {}

// Group is a sequence of child elements.
type Group struct {
	elementBase
	Elements []Element
}

// Add appends a child and returns it.
func (g *Group) Add(el Element) Element {
	g.Elements = append(g.Elements, el)
	return el
}

// Remove deletes the first occurrence of el (PHP AbstractGroupElement::remove).
func (g *Group) Remove(el Element) {
	for i, c := range g.Elements {
		if c == el {
			g.Elements = append(g.Elements[:i], g.Elements[i+1:]...)
			return
		}
	}
}

// Identifier is an OMML identifier (variable name).
type Identifier struct {
	elementBase
	Value string
}

func NewIdentifier(v string) *Identifier { return &Identifier{Value: v} }

// Numeric is a numeric literal.
type Numeric struct {
	elementBase
	Value string
}

func NewNumeric(v string) *Numeric { return &Numeric{Value: v} }

// Operator is a math operator such as +, −, =.
type Operator struct {
	elementBase
	Value string
}

func NewOperator(v string) *Operator { return &Operator{Value: v} }

// Fraction is m:f (numerator / denominator).
type Fraction struct {
	elementBase
	Numerator   Element
	Denominator Element
}

func NewFraction(num, den Element) *Fraction {
	return &Fraction{Numerator: num, Denominator: den}
}

// Superscript is m:sSup.
type Superscript struct {
	elementBase
	Base Element
	Sup  Element
}

func NewSuperscript(base, sup Element) *Superscript {
	return &Superscript{Base: base, Sup: sup}
}

// Subscript is m:sSub.
type Subscript struct {
	elementBase
	Base Element
	Sub  Element
}

func NewSubscript(base, sub Element) *Subscript {
	return &Subscript{Base: base, Sub: sub}
}

// Radical is m:rad (square root).
type Radical struct {
	elementBase
	Deg  Element
	Base Element
}

func NewRadical(base Element) *Radical {
	return &Radical{Base: base}
}

// Delimiter is m:d (matched fences).
type Delimiter struct {
	elementBase
	Beg     string
	End     string
	Content Element
}

func NewDelimiter(beg, end string, content Element) *Delimiter {
	return &Delimiter{Beg: beg, End: end, Content: content}
}

// Row is a horizontal group of elements.
type Row struct {
	Group
}

func NewRow() *Row { return &Row{} }

// Semantics wraps a MathML semantics node.
type Semantics struct {
	Group
	Annotations map[string]string
}

func NewSemantics() *Semantics { return &Semantics{Annotations: map[string]string{}} }

// AddAnnotation stores a MathML annotation by encoding (PHP Semantics::addAnnotation).
func (s *Semantics) AddAnnotation(encoding, annotation string) *Semantics {
	if s.Annotations == nil {
		s.Annotations = map[string]string{}
	}
	s.Annotations[encoding] = annotation
	return s
}

// Annotation returns the annotation for encoding, or "".
func (s *Semantics) Annotation(encoding string) string {
	if s.Annotations == nil {
		return ""
	}
	return s.Annotations[encoding]
}
