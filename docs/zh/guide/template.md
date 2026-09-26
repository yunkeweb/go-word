# 模板引擎 v2（Pipe 过滤器）

`TemplateProcessor` 按 PHPWord 的方式填充已有 `.docx`：在 `word/document.xml`（及相关部件）中替换 `${placeholders}`。`NewTemplateProcessorBytes` 可以直接填充刚生成的文档，因此不必随仓库分发 `.docx` 模板。

## 构造与赋值

### 签名

```go
func NewTemplateProcessor(filename string) (*TemplateProcessor, error)
func NewTemplateProcessorBytes(data []byte) (*TemplateProcessor, error)
func (t *TemplateProcessor) SetValue(search, replace string)
func (t *TemplateProcessor) SetValues(values map[string]string)
func (t *TemplateProcessor) Save(filename string) error
func (t *TemplateProcessor) Bytes() ([]byte, error)
```

每个 `${...}` 应落在同一个 `w:t`。若 Word 把占位符拆到多个 run，处理器就看不到 token。

## Pipe 过滤器

链式过滤器写作 `${var | filter}` 或 `${var | filter:arg}`：

| 过滤器 | 效果 |
| --- | --- |
| `upper` / `lower` | 大小写转换 |
| `trim` | 去掉首尾空白 |
| `truncate:N` | 截到 N 个 rune |
| `default:fallback` | 空值时替换 |
| `formatDate:2006-01-02` | 解析/格式化日期 |
| `formatCurrency:¥` | 数值货币（参数作为前缀） |

### 签名

```go
func RegisterTemplateFilter(name string, fn func(in any, args ...string) string)
```

## 块与条件

### 签名

```go
func (t *TemplateProcessor) CloneBlock(blockName string, count int) error
func (t *TemplateProcessor) CloneBlockAndSetValues(blockName string, values []map[string]string) error
func (t *TemplateProcessor) CloneNestedBlock(blockName string, items []BlockData) error
func (t *TemplateProcessor) SetCondition(name string, keep bool) error
func (t *TemplateProcessor) SetConditions(conds map[string]bool) error
func (t *TemplateProcessor) ApplyConditionsFromValues(values map[string]string) error
```

`BlockData` 字段：`Values`、`Blocks`（嵌套命名克隆）、`If`（实例级条件）。

```
${items}
  ${sku} × ${qty}
  ${if overdue}OVERDUE${endif}
${/items}
${if paid}Thank you.${endif}
```

二元比较（`==`、`!=`、`>`、`<`、`>=`、`<=`）对照 `SetValue` 求值。为假的 `${if}` 会裁掉覆盖的段落或表格行。`ApplyConditionsFromValues` 把空 / `0` / `false` / `no` / `off` 视为假。

`CloneRow` / `DeleteRow` 会把 `w:vMerge` 组与 `gridSpan` 一起克隆。图片用 `SetImageValue` / `SetImageValueBytes` 注入；图表用 `SetChart`。

## 完整示例

```go
package main

import (
	"log"

	"github.com/yunkeweb/go-word"
)

func main() {
	src := word.New()
	sec := src.AddSection()
	sec.AddTitle("Template Engine v2", 1)
	sec.AddText("Hello ${name | upper}")
	sec.AddText("City ${city | lower}")
	sec.AddText("Title ${title | trim | upper}")
	sec.AddText("Blurb ${blurb | truncate:12}")
	sec.AddText("When ${created_at | formatDate:\"2006-01-02\"}")
	sec.AddText("Amount ${amount | formatCurrency:¥}")
	sec.AddText("Empty ${missing | default:N/A}")
	sec.AddText("${if age >= 18}Adult content is visible.${endif}")
	sec.AddText("${if age < 18}Minor content should be clipped.${endif}")
	sec.AddText("${items}")
	sec.AddText("${sku} × ${qty}")
	sec.AddText("${if overdue}OVERDUE${endif}")
	sec.AddText("${/items}")
	sec.AddText("${if paid}Thank you for your payment.${endif}")

	raw, err := src.Bytes()
	if err != nil {
		log.Fatal(err)
	}
	tp, err := word.NewTemplateProcessorBytes(raw)
	if err != nil {
		log.Fatal(err)
	}
	tp.SetValues(map[string]string{
		"name":       "alice",
		"city":       "Shanghai",
		"title":      "  go-word  ",
		"blurb":      "Streaming parser, charts and filters",
		"created_at": "2026-09-26T08:00:00Z",
		"amount":     "1999.5",
		"age":        "21",
	})
	if err := tp.CloneNestedBlock("items", []word.BlockData{
		{Values: map[string]string{"sku": "A-01", "qty": "2"}, If: map[string]bool{"overdue": true}},
		{Values: map[string]string{"sku": "B-02", "qty": "1"}, If: map[string]bool{"overdue": false}},
	}); err != nil {
		log.Fatal(err)
	}
	if err := tp.SetCondition("paid", true); err != nil {
		log.Fatal(err)
	}
	if err := tp.Save("template-v2.docx"); err != nil {
		log.Fatal(err)
	}
}
```

Word 中显示 `ALICE`、`shanghai`、`GO-WORD`、截断后的简介、`2026-09-26`、`¥1999.50`、`N/A`、成人行、两行明细（第一行带 OVERDUE）以及感谢句。未成年行已被裁掉。
