# LaTeX 转 OMML 原生公式

`AddMath` 解析基础 LaTeX 并写出 **Office Math ML**。Microsoft Word 将其作为原生公式打开，可双击进入与“插入 → 公式”相同的编辑器。

展示公式包在段落内的 `m:oMathPara` 中。行内公式是当前段落里 `w:r` 的兄弟节点。

## AddMath

### 签名

```go
func (d *Document) AddMath(formula string) *element.Formula
func (c *Container) AddMath(formula string) *Formula
func (p *TextRun) AddMath(formula string) *Formula
```

### 参数

| 名称 | 类型 | 说明 |
| --- | --- | --- |
| `formula` | `string` | LaTeX 子集。未知 token 会变成标识符，Word 仍能打开文件。 |

### 支持的 LaTeX

| 输入 | OMML |
| --- | --- |
| `\frac{a}{b}` | `m:f`（分子 / 分母） |
| `x^{2}` / `x_{n}` | `m:sSup` / `m:sSub` |
| `\sqrt{x}` | `m:rad` |
| `(a+b)` 或 `\left( ... \right)` | `m:d` |
| `\alpha` `\pi` `\times` `\leq` `\infty` `\sum` `\int` | Unicode 标识符 |

## pkg/math

### 签名

```go
func ParseLaTeX(src string) (*Math, error)
func WriteOMML(m *Math) ([]byte, error)  // 展示：m:oMathPara
func WriteOMath(m *Math) ([]byte, error) // 行内：m:oMath
```

需要 OMML 字节、不必构造 `Document` 时使用这些函数。

## 完整示例

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

### Word 中的效果

每条展示公式都是可编辑的居中公式。双击 `\frac{a}{b}`，Word 打开线性/专业公式编辑器，显示堆叠分数。欧拉恒等式与周围文字同一行。没有 OLE 对象，也没有图片回退。

与 `SetColumns` 混排的学术报告见 [v0.8.0 demo](https://github.com/yunkeweb/go-word/tree/main/examples/v0.8.0_demo)。
