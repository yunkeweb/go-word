# 常见报错与 Word 兼容排查

部件违反 Word 期望的 schema 时，Microsoft Word 会弹出修复对话框。GoWord 的 `openxml_strict_test.go` 固化了已经在生产文件上踩过的规则。本页把症状映射到这些规则。

## Word 提示文件已损坏

| 现象 | 可能原因 | 处理 |
| --- | --- | --- |
| 打开即修复，图表丢失 | 系列 XML 不按 XSD 顺序，或 `c:legend` 写在 `c:plotArea` 内部 | 使用 `AddChart` / `AddComboSeries`，不要手改图表 XML。见 [图表](./charts)。 |
| 修复对话框，面积图 | 面积系列上出现 `c:dLblPos` | `SetDataLabels` 对面积图会省略 `dLblPos`。 |
| 修复对话框，组合图 | 两个系列共用一个 `c:axId` | `AddComboSeries(..., true)` 会分配第二条数值轴。 |
| 修复对话框，表格 | 某个 `w:tc` 没有 `w:p` | 始终 `AddText`（或让单元格空着 —— GoWord 会补空段落）。 |
| 修复对话框，分栏 | 写成 `w:separator` 而不是 `w:sep` | 调用 `SetColumns`，不要自定义 `w:cols`。 |
| 模板输出出现 `&#80;` / 实体损坏 | 替换进 `w:t` 的裸 `&` | 使用 `SetValue`。处理器会 XML 转义替换值。 |

## 占位符没有被替换

用户在 GUI 里改模板时，Word 常把 `${name}` 拆到多个 `w:r`。用 GoWord 生成模板（`AddText("${name}")`），或保证每个占位符在同一个 run。对刚构建的文档调用 `NewTemplateProcessorBytes` 不会碰到这个问题。

## 合并后图片消失

两份源文档都创建了 `word/media/image1.png`。若不重映射，后写入的部件会覆盖先写入的。`AppendDocument` 会清空 `RelationID`，写出器再分配 `imageN` 和新的 `rId`。见 [文档合并](./merger)。

## TOC / PAGE 域显示“错误！未定义书签。”

域保存的是缓存结果。右键 → 更新域，或打印。必须使用 `AddTitle` 才会有大纲级别。这是 Word 行为，不是缺部件。

## Protect 并不加密 ZIP

`Protect` 写出 `w:documentProtection`。任何人仍可解压该包。字节必须保密时，请使用操作系统级加密或带密码的压缩包。

## 双栏只作用在页面的一部分

`SetColumns` 是节属性。`AddSection` 之后的内容属于新的 `w:sectPr`。把分栏正文放在那一节；标题若要通栏，就留在上一节。

## 支持哪些 Word 版本？

Word 2007 至 Microsoft 365 均可打开这些包。DrawingML `wps:wsp` 是 Word 2010+ 特性；Word 2007 仍能打开文件，但可能丢掉形状。图表需要 Word 2007+。LibreOffice 打开段落、表格与图片；图表和 `wps` 形状以 Word 为准。

## 完整示例 — 失败后列出部件

```go
package main

import (
	"archive/zip"
	"bytes"
	"fmt"
	"log"

	"github.com/yunkeweb/go-word"
)

func main() {
	doc := word.New()
	sec := doc.AddSection()
	sec.AddText("debug")
	sec.AddMath(`\frac{a}{b}`)
	raw, err := doc.Bytes()
	if err != nil {
		log.Fatal(err)
	}
	zr, err := zip.NewReader(bytes.NewReader(raw), int64(len(raw)))
	if err != nil {
		log.Fatal(err)
	}
	for _, f := range zr.File {
		fmt.Printf("%s %d\n", f.Name, f.UncompressedSize64)
	}
}
```

若 Word 仍修复你生成的文件，打开对应部件（`word/charts/chart1.xml`、`word/document.xml`），与 [`examples/v0.8.0_demo`](https://github.com/yunkeweb/go-word/tree/main/examples/v0.8.0_demo) 以及 [`openxml_strict_test.go`](https://github.com/yunkeweb/go-word/blob/main/openxml_strict_test.go) 对照。
