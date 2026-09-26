# Office Math (OMML)

`AddMath` parses a basic LaTeX string and writes **Office Math ML** (`m:oMathPara` / `m:oMath`). Microsoft Word opens the result as a native equation you can double-click to edit.

Available on `Document`, `Section`, and `Paragraph` (`TextRun`):

- Display math is wrapped in `m:oMathPara` inside a `w:p`.
- Inline math is a sibling of `w:r` inside the current paragraph.

## Supported LaTeX

| Input | OMML |
| --- | --- |
| `\frac{a}{b}` | `m:f` (numerator / denominator) |
| `x^{2}` / `x_{n}` | `m:sSup` / `m:sSub` |
| `\sqrt{x}` | `m:rad` |
| `(a+b)` or `\left( ... \right)` | `m:d` |
| `\alpha` `\pi` `\times` `\leq` `\infty` `\sum` `\int` | Unicode identifiers |

Lower-level trees live in `github.com/yunkeweb/go-word/pkg/math` (`ParseLaTeX`, `WriteOMML`, `WriteOMath`) if you need to serialize OMML by hand.

## Complete example

Save as `main.go` and run `go run .`. The file `omml.docx` opens in Microsoft Word without a repair dialog.

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
	doc.SetDefaultFontSize(11)

	info := doc.GetDocInfo()
	info.Title = "Office Math (OMML)"
	info.Creator = "GoWord"

	sec := doc.AddSection()
	sec.AddTitle("Office Math (OMML)", 1)
	sec.AddText("Display equations occupy their own paragraph (m:oMathPara). Inline equations sit next to w:r runs.")

	sec.AddTitle("Fractions, roots, superscripts", 2)
	sec.AddText("Fraction:")
	sec.AddMath(`\frac{a}{b}`)
	sec.AddText("Pythagoras:")
	sec.AddMath(`x^{2} + y^{2} = z^{2}`)
	sec.AddText("Nested subscript inside a radical:")
	sec.AddMath(`\sqrt{x_1} + \pi`)
	sec.AddText("Delimiter:")
	sec.AddMath(`\left( x+1 \right)`)

	sec.AddTitle("Inline math", 2)
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

OpenXML produced by this program:

| Feature | Node |
| --- | --- |
| Display formula | `w:p` / `m:oMathPara` / `m:oMath` / `m:f` |
| Superscript | `m:sSup` |
| Radical | `m:rad` |
| Inline formula | `m:oMath` beside `w:r` |

See [`examples/v0.8.0_demo`](https://github.com/yunkeweb/go-word/tree/main/examples/v0.8.0_demo) and the academic report in [Examples & Recipes](./examples).
