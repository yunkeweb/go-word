# 模板引擎 v2

`TemplateProcessor` 按 PHPWord 的方式填充已有 `.docx`：在 `word/document.xml`（及相关部件）中替换 `${placeholders}`。`NewTemplateProcessorBytes` 可以直接填充刚生成的文档，因此不必随仓库分发 `.docx` 模板文件。

## Pipe 过滤器

链式过滤器写作 `${var | filter}` 或 `${var | filter:arg}`：

| 过滤器 | 效果 |
| --- | --- |
| `upper` / `lower` | 大小写转换 |
| `trim` | 去掉首尾空白 |
| `truncate:N` | 截到 N 个 rune |
| `default:fallback` | 空值时替换 |
| `formatDate:2006-01-02` | 解析/格式化日期 |
| `formatCurrency:USD` | 数值货币（参数作为前缀） |

额外过滤器通过 `word.RegisterTemplateFilter` 注册。

## 块、嵌套循环、条件

```
${items}
  ${sku} × ${qty}
  ${if overdue}OVERDUE${endif}
${/items}
${if paid}Thank you.${endif}
```

- `${block}` / `${/block}` — `CloneBlock`、`CloneBlockAndSetValues`、`CloneNestedBlock`
- `${block_a}` 可以包含 `${block_b}`。嵌套标记会按克隆编号索引（`${inner#1}`）。
- `${if name}` / `${endif}` — `SetCondition` / `SetConditions`。二元比较（`==`、`!=`、`>`、`<`、`>=`、`<=`）对照 `SetValue` 求值。
- `CloneRow` / `DeleteRow` 会把 `w:vMerge` 的 restart/continue 组以及 `gridSpan` 一起克隆。

`BlockData` 字段：`Values`、`Blocks`（嵌套命名克隆）、`If`（实例级条件）。

图片与图表通过 `SetImageValue`、`SetImageValueBytes`、`SetChart` 注入。

## 完整示例

保存为 `main.go` 后执行 `go run .`。程序在内存中构建模板，填充管道 / 块 / 条件，写出 `template-v2.docx`。

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

OpenXML 注意：

- 占位符写在 `w:t` 中。每个 `${...}` 应落在同一个 run 内，避免 Word 把 token 拆到多个 `w:r`。
- `CloneBlock` 复制开闭宏之间的 XML，再给嵌套宏加上 `#n` 后缀。
- 为假的 `${if}` 会裁掉覆盖的段落（或表格行）。`ApplyConditionsFromValues` 把空 / `0` / `false` / `no` / `off` 视为假。

财务 / 合同批量导出见 [实战案例库](./examples) 案例 C。
