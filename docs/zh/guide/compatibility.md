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
| 水印 | 页眉中的 VML `PowerPlusWaterMarkObject` | 斜向艺术字 |

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

`ExtractText` 遍历重建后的 DOM。文件太大不宜整树加载时，使用 [StreamExtractText](./streaming)。

## 相关

Word 提示“文件损坏”时，从 [FAQ](./faq) 开始。图表 XSD 顺序见 [图表](./charts)。
