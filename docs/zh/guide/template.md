# 模板引擎 v2

`TemplateProcessor` 与 PHPWord 一样填充已有 `.docx`：替换 `word/document.xml`（及相关部件）中的 `${占位符}`。

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

## 管道过滤器

链式过滤器写作 `${var | filter}` 或 `${var | filter:arg}`：

| 过滤器 | 作用 |
| --- | --- |
| `upper` / `lower` | 大小写 |
| `trim` | 去掉首尾空白 |
| `truncate:N` | 截到 N 个 rune |
| `default:fallback` | 空值回退 |
| `formatDate:2006-01-02` | 解析/格式化日期 |
| `formatCurrency:USD` | 货币数字 |

```
${name | trim | upper}
${price | formatCurrency:CNY}
${when | formatDate:2006-01-02}
${note | default:n/a}
```

可用 `word.RegisterTemplateFilter` 注册自定义过滤器。

## 块循环与嵌套

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

`${block_a}` 内可以再套 `${block_b}`。`CloneRow` / `DeleteRow` 会把 `w:vMerge` 的 restart/continue 组与 `gridSpan` 一并处理。

## 条件

```
${if paid}
感谢付款。
${endif}
```

支持二元比较（`==`、`!=`、`>`、`<`、`>=`、`<=`）。用 `SetCondition` / `SetConditions` 设标志，或通过 `ApplyConditionsFromValues` 从变量推导。条件为假或空时，覆盖的段落或表格行会被裁掉。

图片与图表可用 `SetImageValue`、`SetImageValueBytes`、`SetChart` 注入。
