package word

import "github.com/yunkeweb/go-word/element"

// AddMath appends a display OMML equation parsed from a basic LaTeX string
// to the last (or a new) section.
func (d *Document) AddMath(formula string) *element.Formula {
	return d.lastOrNewSection().AddMath(formula)
}
