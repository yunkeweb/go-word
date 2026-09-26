# Template Engine v2

`TemplateProcessor` fills an existing `.docx` the same way PHPWord does: `${placeholders}` inside `word/document.xml` (and related parts). `NewTemplateProcessorBytes` fills a document you just generated, so you never have to ship a `.docx` template file.

## Pipe filters

Chained filters use `${var | filter}` or `${var | filter:arg}`:

| Filter | Effect |
| --- | --- |
| `upper` / `lower` | Case fold |
| `trim` | Strip surrounding space |
| `truncate:N` | Cut to N runes |
| `default:fallback` | Substitute when empty |
| `formatDate:2006-01-02` | Parse/format dates |
| `formatCurrency:USD` | Numeric currency (prefix is the argument) |

Register extra filters with `word.RegisterTemplateFilter`.

## Blocks, nested loops, conditions

```
${items}
  ${sku} × ${qty}
  ${if overdue}OVERDUE${endif}
${/items}
${if paid}Thank you.${endif}
```

- `${block}` / `${/block}` — `CloneBlock`, `CloneBlockAndSetValues`, `CloneNestedBlock`
- `${block_a}` may contain `${block_b}`. Nested markers are indexed (`${inner#1}`) per clone.
- `${if name}` / `${endif}` — `SetCondition` / `SetConditions`. Binary comparisons (`==`, `!=`, `>`, `<`, `>=`, `<=`) evaluate against `SetValue`.
- `CloneRow` / `DeleteRow` keep `w:vMerge` restart/continue groups and `gridSpan` together.

`BlockData` fields: `Values`, `Blocks` (nested named clones), `If` (per-instance conditions).

Images and charts inject with `SetImageValue`, `SetImageValueBytes`, and `SetChart`.

## Complete example

Save as `main.go` and run `go run .`. The program builds the template in memory, fills pipes / blocks / conditions, and writes `template-v2.docx`.

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

OpenXML notes:

- Placeholders live in `w:t`. Keep each `${...}` inside a single run so Word does not split the token across `w:r` nodes.
- `CloneBlock` copies the XML between the open and close macros, then suffixes nested macros with `#n`.
- False `${if}` clips the covering paragraphs (or table rows). Empty / `0` / `false` / `no` / `off` are false for `ApplyConditionsFromValues`.

See [Examples & Recipes](./examples) Case C for a finance / contract batch export.
