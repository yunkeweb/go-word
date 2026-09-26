# Office Math (OMML)

`AddMath` parses a basic LaTeX string and writes **Office Math ML** (`m:oMathPara` / `m:oMath`). Microsoft Word opens the result as a native equation you can double-click to edit.

Available on `Document`, `Section`, and `Paragraph` (`TextRun`):

```go
sec.AddMath(`\frac{a}{b}`)                 // display equation (own paragraph)
p := sec.AddTextRun()
p.AddText("inline: ")
p.AddMath(`x^{2} + y^{2} = z^{2}`)         // inline m:oMath
```

Display math is wrapped in `m:oMathPara` inside a `w:p`. Inline math is a sibling of `w:r` inside the current paragraph.

## Supported LaTeX

| Input | OMML |
| --- | --- |
| `\frac{a}{b}` | `m:f` (numerator / denominator) |
| `x^{2}` / `x_{n}` | `m:sSup` / `m:sSub` |
| `\sqrt{x}` | `m:rad` |
| `(a+b)` or `\left( ... \right)` | `m:d` |
| `\alpha` `\pi` `\times` `\leq` `\infty` `\sum` `\int` | Unicode identifiers |

```go
sec.AddMath(`\sqrt{x_1} + \pi`)
sec.AddMath(`\left( x+1 \right)`)
```

Lower-level trees live in `pkg/math` (`ParseLaTeX`, `WriteOMML`, `WriteOMath`) if you need to build expressions by hand.

See [`examples/v0.8.0_demo`](https://github.com/yunkeweb/go-word/tree/main/examples/v0.8.0_demo).
