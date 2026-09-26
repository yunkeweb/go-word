# Template Engine v2

`TemplateProcessor` fills an existing `.docx` the same way PHPWord does: `${placeholders}` inside `word/document.xml` (and related parts). `NewTemplateProcessorBytes` fills a document you just generated, so you never have to ship a `.docx` template file.

## Constructors and values

### Signature

```go
func NewTemplateProcessor(filename string) (*TemplateProcessor, error)
func NewTemplateProcessorBytes(data []byte) (*TemplateProcessor, error)
func (t *TemplateProcessor) SetValue(search, replace string)
func (t *TemplateProcessor) SetValues(values map[string]string)
func (t *TemplateProcessor) Save(filename string) error
func (t *TemplateProcessor) Bytes() ([]byte, error)
```

Keep each `${...}` inside a single `w:t`. If Word splits a placeholder across runs, the processor will not see the token.

## Pipe filters

Chained filters use `${var | filter}` or `${var | filter:arg}`:

| Filter | Effect |
| --- | --- |
| `upper` / `lower` | Case fold |
| `trim` | Strip surrounding space |
| `truncate:N` | Cut to N runes |
| `default:fallback` | Substitute when empty |
| `formatDate:2006-01-02` | Parse/format dates |
| `formatCurrency:¥` | Numeric currency (argument is the prefix) |

### Signature

```go
func RegisterTemplateFilter(name string, fn func(in any, args ...string) string)
```

## Blocks and conditions

### Signature

```go
func (t *TemplateProcessor) CloneBlock(blockName string, count int) error
func (t *TemplateProcessor) CloneBlockAndSetValues(blockName string, values []map[string]string) error
func (t *TemplateProcessor) CloneNestedBlock(blockName string, items []BlockData) error
func (t *TemplateProcessor) SetCondition(name string, keep bool) error
func (t *TemplateProcessor) SetConditions(conds map[string]bool) error
func (t *TemplateProcessor) ApplyConditionsFromValues(values map[string]string) error
```

`BlockData` fields: `Values`, `Blocks` (nested named clones), `If` (per-instance conditions).

```
${items}
  ${sku} × ${qty}
  ${if overdue}OVERDUE${endif}
${/items}
${if paid}Thank you.${endif}
```

Binary comparisons (`==`, `!=`, `>`, `<`, `>=`, `<=`) evaluate against `SetValue`. False `${if}` clips covering paragraphs or table rows. Empty / `0` / `false` / `no` / `off` are false for `ApplyConditionsFromValues`.

`CloneRow` / `DeleteRow` keep `w:vMerge` groups and `gridSpan` together. Images inject with `SetImageValue` / `SetImageValueBytes`; charts with `SetChart`.

## Complete example

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

Word shows `ALICE`, `shanghai`, `GO-WORD`, a truncated blurb, `2026-09-26`, `¥1999.50`, `N/A`, the adult line, two item rows (the first tagged OVERDUE), and the thank-you sentence. The minor line is gone.
