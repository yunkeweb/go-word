# 常见问题

部件违反 Word 期望的 schema 时，Microsoft Word 会弹出修复对话框。GoWord 的 `openxml_strict_test.go` 固化了已经在生产文件上踩过的规则。本页把症状映射到这些规则，并说明合并后的样式冲突如何重映射。

相关：[OpenXML 兼容性](./compatibility)、[文档合并](./merger)、[企业级实战案例](./recipes)。

## Word 提示文件已损坏

| 现象 | 可能原因 | 处理 |
| --- | --- | --- |
| 打开即修复，图表丢失 | 系列 XML 不按 XSD 顺序，或 `c:legend` 写在 `c:plotArea` 内部 | 使用 `AddChart` / `AddComboSeries`，不要手改图表 XML。见 [图表](./charts)。 |
| 修复对话框，面积图 | 面积系列上出现 `c:dLblPos` | `SetDataLabels` 对面积图会省略 `dLblPos`。 |
| 修复对话框，组合图 | 两个系列共用一个 `c:axId` | `AddComboSeries(..., true)` 会分配第二条数值轴。 |
| 修复对话框，表格 | 某个 `w:tc` 没有 `w:p` | 始终 `AddText`（或让单元格空着 —— GoWord 会补空段落）。嵌套表后面仍需要那段落；`Cell.AddTable` 会写上。 |
| 修复对话框，分栏 | 写成 `w:separator` 而不是 `w:sep` | 调用 `SetColumns`，不要自定义 `w:cols`。 |
| 修复对话框，形状 | DrawingML `wps:wsp` 缺少 `w:drawing` 包装 | 使用 `AddShape` / `AddTextBox`。 |
| 修复对话框，内容控件 | `w:sdtContent` 缺失或顺序错误 | 使用 `AddSDTText` / `AddSDTDropdown` / `AddSDTDate` / `AddSDTCheckbox`。见 [SDT](./sdt)。 |
| 保护后某字段仍无法填写 | 该段落或单元格未调用 `AllowEdit` | 在 `Protect` 之后调用 `AllowEdit("Everyone")`。见 [保护](./protect)。 |
| 模板输出出现 `&#80;` / 实体损坏 | 替换进 `w:t` 的裸 `&` | 使用 `SetValue`。处理器会 XML 转义替换值。 |
| 需要全功能冒烟测试 | 组合全部公开模块 | `go run ./tests/matrix`，再跑 Go Reader 与 Word COM 校验。见 [兼容性](./compatibility) 与 [案例 5](./recipes#5-全要素矩阵)。 |

若 Word 仍修复你生成的文件，列出 ZIP 部件（见文末示例），把 `word/charts/chart1.xml` 或 `word/document.xml` 与 [`examples/v0.9.0_sdt`](https://github.com/yunkeweb/go-word/tree/main/examples/v0.9.0_sdt)、[`examples/v0.8.0_demo`](https://github.com/yunkeweb/go-word/tree/main/examples/v0.8.0_demo) 以及 [`openxml_strict_test.go`](https://github.com/yunkeweb/go-word/blob/main/openxml_strict_test.go) 对照。

## 合并后样式 ID 冲突（重映射）

两份源文档都定义了名为 `Note`（或 `Heading1`）的段落样式时，不能共用一份 `styles.xml`。若不重映射，后写入的定义会覆盖先写入的，所有仍写着 `Note` 的 `w:pStyle` 会拿到错误的段距、大纲级别或编号。

`AppendDocument` 会给**发生冲突**的段落 / 表格样式 ID 加前缀。未冲突的名字保持不变。克隆树上的 `w:pStyle` / `w:tblStyle` 会改写成新 ID。

```go
err := dst.AppendDocument(src, word.MergeOptions{
	StylePrefix:    "src_",
	BookmarkPrefix: "src_",
	SectionBreak:   "nextPage",
})
```

| 标识符 | 默认前缀 | 例子 |
| --- | --- | --- |
| 段落 / 表格样式 ID | `src_` | 目标已有 `Note` 时，源上的 `Note` 变成 `src_Note` |
| 书签名 | `src_` | `shared` → `src_shared`；未冲突的名字保持 |
| 内部超链接锚点 | （派生） | 克隆树里的 `#shared` 改写到新书签名 |
| 图片 / 媒体 | 写出时的 `imageN` | 清空 `RelationID`，写出器分配新的 `rId` |

拼接超过两棵树时，为每个源传入不同前缀（`note_`、`annex_` …）。[企业级实战案例](./recipes) 的案例 3 用这种方式拼出三件套卷宗。

`rId` 在写出时分配（`word2007Writer.nextRel`），合并后的 ZIP 里两份文档不会共用关系 ID。不要手写 `rId`。

## 占位符没有被替换

用户在 GUI 里改模板时，Word 常把 `${name}` 拆到多个 `w:r`。用 GoWord 生成模板（`AddText("${name}")`），或保证每个占位符在同一个 run。对刚构建的文档调用 `NewTemplateProcessorBytes` 不会碰到这个问题。

过滤器（`${amount | formatCurrency:¥}`）只作用于处理器仍看成单个 `${...}` 的记号。被拆开的 `formatCurrency` 会变成字面文本。

`CloneRow` 在 `w:tr` 里找占位符（并保持 `w:vMerge` 组完整）。`CloneBlock` 找 `${name}` … `${/name}`。同一记号上混用两者会失败——只选一种。

## 合并后图片消失

两份源文档都创建了 `word/media/image1.png`。若不重映射，后写入的部件会覆盖先写入的。`AppendDocument` 会清空 `RelationID`，写出器再分配 `imageN` 和新的 `rId`。见 [文档合并](./merger)。

## TOC / PAGE 域显示“错误！未定义书签。”

域保存的是缓存结果。右键 → 更新域，或打印。必须使用 `AddTitle` 才会有大纲级别。这是 Word 行为，不是缺部件。在 [TemplateProcessor](./template) 上调用 `SetUpdateFields(true)` 会让 Word 打开时刷新域（`w:updateFields`）。

## Protect 并不加密 ZIP

`Protect` 写出 `w:documentProtection`。任何人仍可解压该包。字节必须保密时，请使用操作系统级加密或带密码的压缩包。演示里的样例密码是 `goword`。

## 双栏只作用在页面的一部分

`SetColumns` 是节属性。`AddSection` 之后的内容属于新的 `w:sectPr`。把分栏正文放在那一节；标题若要通栏，就留在上一节。分隔线属性是 `w:sep`，不是 `w:separator`。

## 支持哪些 Word 版本？

Word 2007 至 Microsoft 365 均可打开这些包。DrawingML `wps:wsp` 是 Word 2010+ 特性；Word 2007 仍能打开文件，但可能丢掉形状。图表需要 Word 2007+。LibreOffice 打开段落、表格与图片；图表和 `wps` 形状以 Word 为准。

## 流式提取 vs `LoadBytes`

`Load` / `LoadBytes` 构建整棵文档树。`StreamExtractText` / `StreamExtractImages` 用 `xml.Decoder` 遍历 `word/document.xml`，回调后丢弃段落缓冲——额外堆相对文件大小保持 O(1)。优先传入 `*os.File`，这样 ZIP 无需整包拷贝即可映射。见 [流式提取](./streaming) 与 [性能基准](./benchmarks)。

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
