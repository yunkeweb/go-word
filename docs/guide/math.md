# LaTeX → OMML (AddMath)

`AddMath` parses a basic LaTeX string and writes **Office Math ML**. Microsoft Word opens the result as a native equation you can double-click to edit — the same editor as Insert → Equation.

Display math is wrapped in `m:oMathPara` inside a `w:p`. Inline math is a sibling of `w:r` inside the current paragraph.

## AddMath

### Signature

```go
func (d *Document) AddMath(formula string) *element.Formula
func (c *Container) AddMath(formula string) *Formula
func (p *TextRun) AddMath(formula string) *Formula
```

### Parameters

| Name | Type | Description |
| --- | --- | --- |
| `formula` | `string` | LaTeX subset. Unknown tokens become identifiers so Word still opens the file. |

### Supported LaTeX

| Input | OMML |
| --- | --- |
| `\frac{a}{b}` | `m:f` (numerator / denominator) |
| `x^{2}` / `x_{n}` | `m:sSup` / `m:sSub` |
| `\sqrt{x}` | `m:rad` |
| `(a+b)` or `\left( ... \right)` | `m:d` |
| `\alpha` `\pi` `\times` `\leq` `\infty` `\sum` `\int` | Unicode identifiers |

## pkg/math

### Signature

```go
func ParseLaTeX(src string) (*Math, error)
func WriteOMML(m *Math) ([]byte, error)  // display: m:oMathPara
func WriteOMath(m *Math) ([]byte, error) // inline: m:oMath
```

Use these when you need the OMML bytes without a `Document`.

## Complete example

```go
package main

import (
	"fmt"
	"log"

	"github.com/yunkeweb/go-word"
	gowordmath "github.com/yunkeweb/go-word/pkg/math"
)

func main() {
	doc := word.New()
	doc.SetDefaultFontName("Calibri")
	doc.SetDefaultAsianFontName("Microsoft YaHei")

	sec := doc.AddSection()
	sec.AddTitle("Office Math (OMML)", 1)
	sec.AddText("Display equations occupy their own paragraph (m:oMathPara).")

	sec.AddText("Fraction:")
	sec.AddMath(`\frac{a}{b}`)
	sec.AddText("Pythagoras:")
	sec.AddMath(`x^{2} + y^{2} = z^{2}`)
	sec.AddText("Nested subscript inside a radical:")
	sec.AddMath(`\sqrt{x_1} + \pi`)
	sec.AddText("Delimiter:")
	sec.AddMath(`\left( x+1 \right)`)

	p := sec.AddTextRun()
	p.AddText("The identity ")
	p.AddMath(`e^{i\pi} + 1 = 0`)
	p.AddText(" sits in the same w:p as the surrounding text.")

	tree, err := gowordmath.ParseLaTeX(`\frac{1}{2}`)
	if err != nil {
		log.Fatal(err)
	}
	omml, err := gowordmath.WriteOMML(tree)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("hand-serialized OMML bytes: %d\n", len(omml))

	if err := doc.Save("omml.docx"); err != nil {
		log.Fatal(err)
	}
}
```

### What Word shows

Each display formula is a centered, editable equation. Double-click `\frac{a}{b}` and Word opens the linear/professional equation editor with a stacked fraction. The Euler identity stays on one line with the words around it. There is no OLE object and no image fallback.

A two-column academic report that mixes these formulas with `SetColumns` lives in the [v0.8.0 demo](https://github.com/yunkeweb/go-word/tree/main/examples/v0.8.0_demo).
