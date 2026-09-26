# Template Engine v2

`TemplateProcessor` fills an existing `.docx` the same way PHPWord does: `${placeholders}` inside `word/document.xml` (and related parts).

```go
tpl, err := word.NewTemplateProcessor("letter.docx")
if err != nil {
	log.Fatal(err)
}
tpl.SetValue("name", "Ada")
tpl.SetValues(map[string]string{
	"city": "London",
	"date": "2026-09-26",
})
if err := tpl.Save("out.docx"); err != nil {
	log.Fatal(err)
}
```

## Pipe filters

Chained filters use `${var | filter}` or `${var | filter:arg}`:

| Filter | Effect |
| --- | --- |
| `upper` / `lower` | Case fold |
| `trim` | Strip surrounding space |
| `truncate:N` | Cut to N runes |
| `default:fallback` | Substitute when empty |
| `formatDate:2006-01-02` | Parse/format dates |
| `formatCurrency:USD` | Numeric currency |

```
${name | trim | upper}
${price | formatCurrency:CNY}
${when | formatDate:2006-01-02}
${note | default:n/a}
```

Register extra filters with `word.RegisterTemplateFilter`.

## Blocks and nested loops

```
${items}
  ${name} — ${qty}
${/items}
```

```go
_ = tpl.CloneBlock("items", 3)
_ = tpl.CloneBlockAndSetValues("items", []map[string]string{
	{"name": "A", "qty": "2"},
	{"name": "B", "qty": "1"},
})
_ = tpl.CloneNestedBlock("outer", []word.BlockData{
	{Values: map[string]string{"title": "Group 1"}},
})
```

`${block_a}` may contain `${block_b}`. `CloneRow` / `DeleteRow` keep `w:vMerge` restart/continue groups and `gridSpan` together.

## Conditions

```
${if paid}
Thank you for your payment.
${endif}
```

Binary comparisons are supported (`==`, `!=`, `>`, `<`, `>=`, `<=`). Set flags with `SetCondition` / `SetConditions`, or derive them from values via `ApplyConditionsFromValues`. Empty or false clips the covering paragraphs or table rows.

Images and charts can be injected with `SetImageValue`, `SetImageValueBytes`, and `SetChart`.
