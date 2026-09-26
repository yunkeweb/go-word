# Office Math 原生公式

`AddMath` 解析基础 LaTeX 并写出 **Office Math ML**（`m:oMathPara` / `m:oMath`）。Microsoft Word 将其作为原生公式打开，可双击进入公式编辑器。

`Document`、`Section` 与 `Paragraph`（`TextRun`）均可调用：

- 展示公式包在段落内的 `m:oMathPara` 中。
- 行内公式是当前段落里 `w:r` 的兄弟节点。

## 支持的 LaTeX

| 输入 | OMML |
| --- | --- |
| `\frac{a}{b}` | `m:f`（分子 / 分母） |
| `x^{2}` / `x_{n}` | `m:sSup` / `m:sSub` |
| `\sqrt{x}` | `m:rad` |
| `(a+b)` 或 `\left( ... \right)` | `m:d` |
| `\alpha` `\pi` `\times` `\leq` `\infty` `\sum` `\int` | Unicode 标识符 |

如需手写表达式树，可使用 `github.com/yunkeweb/go-word/pkg/math`（`ParseLaTeX`、`WriteOMML`、`WriteOMath`）。

## 完整示例

保存为 `main.go` 后执行 `go run .`。生成的 `omml.docx` 可在 Microsoft Word 中直接打开，不会弹出修复对话框。

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

本程序写出的 OpenXML：

| 功能 | 节点 |
| --- | --- |
| 展示公式 | `w:p` / `m:oMathPara` / `m:oMath` / `m:f` |
| 上标 | `m:sSup` |
| 根号 | `m:rad` |
| 行内公式 | 与 `w:r` 并列的 `m:oMath` |

示例：[`examples/v0.8.0_demo`](https://github.com/yunkeweb/go-word/tree/main/examples/v0.8.0_demo)，以及 [实战案例库](./examples) 中的学术报告。
