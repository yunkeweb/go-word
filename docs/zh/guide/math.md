# Office Math 原生公式

`AddMath` 解析基础 LaTeX 并写出 **Office Math ML**（`m:oMathPara` / `m:oMath`）。Microsoft Word 将其作为原生公式打开，可双击进入公式编辑器。

`Document`、`Section` 与 `Paragraph`（`TextRun`）均可调用：

```go
sec.AddMath(`\frac{a}{b}`)                 // 独立成段的展示公式
p := sec.AddTextRun()
p.AddText("inline: ")
p.AddMath(`x^{2} + y^{2} = z^{2}`)         // 行内 m:oMath
```

展示公式包在段落内的 `m:oMathPara` 中；行内公式是当前段落里 `w:r` 的兄弟节点。

## 支持的 LaTeX

| 输入 | OMML |
| --- | --- |
| `\frac{a}{b}` | `m:f`（分子 / 分母） |
| `x^{2}` / `x_{n}` | `m:sSup` / `m:sSub` |
| `\sqrt{x}` | `m:rad` |
| `(a+b)` 或 `\left( ... \right)` | `m:d` |
| `\alpha` `\pi` `\times` `\leq` `\infty` `\sum` `\int` | Unicode 标识符 |

```go
sec.AddMath(`\sqrt{x_1} + \pi`)
sec.AddMath(`\left( x+1 \right)`)
```

如需手写表达式树，可使用 `pkg/math`（`ParseLaTeX`、`WriteOMML`、`WriteOMath`）。

示例：[`examples/v0.8.0_demo`](https://github.com/yunkeweb/go-word/tree/main/examples/v0.8.0_demo)。
