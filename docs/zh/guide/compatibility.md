# OpenXML 兼容性

GoWord 写出 **Office Open XML WordprocessingML 2007** 包（`ECMA-376` / ISO/IEC 29500 transitional）。Microsoft Word 2007 至 Microsoft 365 均可打开。WPS 与 LibreOffice 打开常见子集（段落、表格、图片）；DrawingML 图表与 `wps:wsp` 形状以 Word 为准。

## 包布局

| 部件 | 必需 | 说明 |
| --- | --- | --- |
| `[Content_Types].xml` | 是 | `document.xml`、页眉、图表、媒体的 Override |
| `_rels/.rels` | 是 | 指向 `word/document.xml` |
| `word/document.xml` | 是 | 正文 + 最后一段 `w:sectPr` |
| `word/_rels/document.xml.rels` | 是 | 图片、页眉页脚、图表、编号 |
| `word/styles.xml` | 是 | 命名样式与默认 `Normal` |
| `word/media/imageN.*` | 有图时 | 名称在写出时分配 |
| `word/charts/chartN.xml` | 有图时 | 每个 `AddChart` 一个部件 |
| `word/headerN.xml` / `footerN.xml` | 有页眉时 | 由 `w:sectPr` 经 `r:id` 引用 |

## 写出器对准的 schema

| 功能 | 命名空间 / 节点 | Word 行为 |
| --- | --- | --- |
| 正文 | `w:` WordprocessingML | 段落、表格、sectPr |
| 原生公式 | `m:` Office Math（`m:oMathPara`、`m:oMath`） | 双击打开公式编辑器 |
| 图表 | `c:` ECMA-376 Chart | 系列顺序为 `idx` → `order` → `tx` → `spPr` |
| Word 2010 形状 | `wps:wsp` / `a:prstGeom` | 矩形、圆角矩形、箭头、文本框 |
| 域 | `w:instrText`（`PAGE`、`TOC`、`NUMPAGES`） | 用“更新域”刷新 |
| 保护 | `w:documentProtection` | 只读 / 批注 / 修订 / 窗体 |
| 编辑例外 | `w:permStart` / `w:permEnd` | `AllowEdit` 区域保持可写 |
| 水印 | 页眉中的 VML `PowerPlusWaterMarkObject` | 斜向或平铺艺术字；图片用水印为 `WordPictureWatermark` |
| SDT | `w:sdt` / `w:sdtPr` / `w:sdtContent` | 纯文本、下拉、日期、`w14:checkbox` |
| 表格行 / 单元格 | `w:tblHeader`、`w:cantSplit`、`w:vAlign`、`w:textDirection` | 跨页表头、禁止断行、竖排文字 |

## 测试套件强制的严格规则

`openxml_strict_test.go` 会拒绝 Word 将要提示修复的输出：

- 每个 `w:tc` 至少包含一个 `w:p`。
- 模板替换不会在 `w:t` 中写出裸 `&`（普通 ASCII 不用 `&#80;` 这类数字实体）。
- 组合图使用两个不同的 `c:axId`；次系列指向第二条数值轴。
- 面积图省略 `c:dLblPos`（面积图 XSD 不允许）。
- `c:legend` 在 `c:plotArea` **之后**，绝不在其内部。
- 折线标记同时写出 `c:symbol` 与 `c:size` 5。
- `w:cols` 使用 `w:sep`，不是 `w:separator`。

## 完整示例 — 经读取器往返

```go
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/yunkeweb/go-word"
)

func main() {
	doc := word.New()
	sec := doc.AddSection()
	sec.AddTitle("Compatibility", 1)
	sec.AddText("Round-trip through CreateReader.")
	sec.AddMath(`x^{2}`)
	if err := doc.Save("compat.docx"); err != nil {
		log.Fatal(err)
	}

	r, err := word.CreateReader("Word2007")
	if err != nil {
		log.Fatal(err)
	}
	loaded, err := r.Load("compat.docx")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(loaded.ExtractText())

	_ = os.Remove("compat.docx")
}
```

`ExtractText` 遍历重建后的 DOM。文件太大不宜整树加载时，使用 [StreamExtractText](./streaming)。库中没有 `word.ReadDOM`；DOM 加载入口是 `word.Open`、`word.Read` 与 `word.Load`。

## 全要素矩阵（v0.9.0）

[`tests/matrix`](https://github.com/yunkeweb/go-word/tree/main/tests/matrix) 会写出 80 份随机组合文档，覆盖 v0.1.0 至 v0.9.0 的全部导出模块：排版、多节、页眉页脚、表格、图片/形状、TOC/书签/批注、OMML、图表、SDT、水印/保护。产物落在 `./test_output_docs`（已 gitignore）。

```sh
go run ./tests/matrix
go run tests/matrix/validate_reader.go
powershell -NoProfile -ExecutionPolicy Bypass -File tests/matrix/validate_docs.ps1
```

| 引擎 | 检查内容 | v0.9.0 结果 |
| --- | --- | --- |
| `word.Open` / `word.Read` + `ExtractText` / `ExtractImages` | DOM 逆向解析，无 panic，`error == nil` | 80 PASS |
| `word.StreamExtractText` / `word.StreamExtractImages` | 流式提取，无 panic，`error == nil` | 80 PASS |
| Microsoft Word COM（`DisplayAlerts=0`、`OpenNoRepairDialog`） | OpenXML 修复弹窗、节点顺序、解析异常 | 80 PASS |

`validate_docs.ps1` 启动无头 `Word.Application`，只读打开每份文件。Word 本会弹出的修复对话框会变成异常。

## 相关

Word 提示“文件损坏”时，从 [FAQ](./faq) 开始。图表 XSD 顺序见 [图表](./charts)。[企业级实战案例](./recipes) 第 4 则组合了 SDT、跨页表头、平铺水印与 `AllowEdit`。
