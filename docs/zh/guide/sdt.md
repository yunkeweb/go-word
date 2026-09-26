# 结构化文档标签（SDT）

内容控件是 Word 原生表单域。GoWord 写出 `w:sdt` → `w:sdtPr` → `w:sdtContent`（CT_SdtBlock / CT_SdtRun）。唯一 `w:id` 从 `100000000` 起编。复选框使用 Word 2010 `w14:checkbox`；文档根声明 `mc:Ignorable="w14 wps"`。

`AddSDT*` 定义在 `element.Container` 上，因此节、单元格、页眉和文本 Run 都可以调用。

## AddSDTText / AddSDTDropdown / AddSDTDate / AddSDTCheckbox

### 签名

```go
func (c *Container) AddSDTText(alias, tag, placeholderText string) *SDT
func (c *Container) AddSDTDropdown(alias, tag string, options map[string]string) *SDT
func (c *Container) AddSDTDate(alias, tag, dateFormat string) *SDT
func (c *Container) AddSDTCheckbox(alias, tag string, checked bool) *SDT
func (c *Container) AddSDT(typ string) *SDT
```

### 参数

| 名称 | 类型 | 说明 |
| --- | --- | --- |
| `alias` | `string` | `w:alias` — Word 控件标题栏显示的名称。 |
| `tag` | `string` | `w:tag` — 后续填充 / 提取用的稳定机器名。 |
| `placeholderText` | `string` | 可见的 `w:sdtContent` 文本。非空时同时设置 `w:showingPlcHdr`。 |
| `options` | `map[string]string` | 下拉 `value → displayText`。键按排序顺序写成 `w:listItem`。 |
| `dateFormat` | `string` | `w:dateFormat`。空则使用 `yyyy-MM-dd`。区域为 `en-US`。 |
| `checked` | `bool` | `w14:checked`。选中写 `☒`，未选写 `☐`（字体 MS Gothic）。 |
| `typ` | `string` | `AddSDT` 的 PHPWord 风格类型：`plainText`、`dropDownList`、`date`、`checkbox`、`comboBox`、`richText`。 |

### 可链式调用的设置器

| 方法 | 效果 |
| --- | --- |
| `SetAlias` / `SetTag` | 改写 `w:alias` / `w:tag`。 |
| `SetValue` | 替换可见内容并清除占位标记。 |
| `SetListItems` | 替换下拉 / 组合框条目。 |
| `SetDateFormat` | 改写 `w:dateFormat`。 |
| `SetChecked` | 切换 `w14:checkbox` 与内容字形。 |

### 注意

- 节点顺序是 `w:sdt` → `w:sdtPr` → `w:sdtContent`。不要手改这段 XML。
- 空的下拉 map 仍会写出控件，占位文本为 `Choose an item.`
- `AddSDT("plainText")` 仍是 PHPWord 别名；新代码请使用带类型的辅助方法。
- 在 Microsoft Word 中填写控件**不需要**打开“开发工具”选项卡。

## 完整示例

保存为 `main.go` 后执行 `go run .`。Microsoft Word 打开 `form.docx` 时不会出现修复对话框。

```go
package main

import (
	"log"

	"github.com/yunkeweb/go-word"
	"github.com/yunkeweb/go-word/style"
)

func main() {
	doc := word.New()
	doc.SetDefaultFontName("Calibri")
	sec := doc.AddSection()
	sec.AddTitle("Employee onboarding form", 1)

	sec.AddSDTText("Full name", "full_name", "Enter full name")
	sec.AddSDTDropdown("Department", "department", map[string]string{
		"eng": "Engineering",
		"fin": "Finance",
		"hr":  "Human Resources",
	})
	sec.AddSDTDate("Start date", "start_date", "yyyy-MM-dd")

	p := sec.AddTextRun()
	p.AddText("I have read the handbook  ")
	p.AddSDTCheckbox("Handbook", "handbook_ack", false)

	tbl := sec.AddTable(style.Table{Width: 9000})
	hdr := tbl.AddRow()
	hdr.AddCell(3000, style.Cell{Shading: style.Shading{Fill: "1F4E79"}}).
		AddText("Field", style.Font{Bold: true, Color: "FFFFFF"})
	hdr.AddCell(6000, style.Cell{Shading: style.Shading{Fill: "1F4E79"}}).
		AddText("Value", style.Font{Bold: true, Color: "FFFFFF"})
	row := tbl.AddRow()
	row.AddCell(3000).AddText("Manager")
	row.AddCell(6000).AddSDTText("Manager", "manager", "Enter manager name")

	if err := doc.Save("form.docx"); err != nil {
		log.Fatal(err)
	}
}
```

### Word 中的效果

四个内容控件：纯文本框、三项下拉、日期选择器，以及手册句旁未勾选的复选框。经理字段位于单元格内（`w:sdt` 作为 `w:tc` 的子节点）。更长的示例见 [`examples/v0.9.0_sdt`](https://github.com/yunkeweb/go-word/tree/main/examples/v0.9.0_sdt)。需要锁定页面其余部分时，与 [文档保护](./protect) 一起使用。
