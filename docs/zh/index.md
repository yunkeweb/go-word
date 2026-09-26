---
layout: home

hero:
  name: GoWord
  text: 工业级 Word（.docx）引擎
  tagline: 100% 纯 Go 标准库、零第三方依赖的 Microsoft Word (.docx) 工业级处理引擎。
  actions:
    - theme: brand
      text: 快速开始
      link: /zh/guide/getting-started
    - theme: alt
      text: 安装
      link: /zh/guide/installation
    - theme: alt
      text: GitHub
      link: https://github.com/yunkeweb/go-word

features:
  - title: 零依赖
    details: 只用 encoding/xml、archive/zip、image、sync。go.mod 没有任何第三方 require，可直接进入隔离与合规环境。
  - title: OMML 公式引擎
    details: AddMath 将 \frac{a}{b}、x^{2} 等 LaTeX 编译成 Word 原生 m:oMathPara，可在公式编辑器中双击编辑。
  - title: 无损文档合并
    details: AppendDocument 自动重映射冲突的样式 ID、书签名与图片 rId，多棵 .docx 树拼接后资源彼此隔离。
  - title: DrawingML 复杂图表
    details: 柱状 / 条形 / 折线 / 饼图 / 面积 / 堆叠 / 双轴组合图，以及矢量形状与文本框（wps:wsp / w:txbxContent）。
  - title: O(1) 流式解析
    details: StreamExtractText / StreamExtractImages 用 xml.Decoder 遍历 .docx ZIP，段落缓冲在回调后立即丢弃。
  - title: 模板 Engine v2
    details: 嵌套 ${block} 循环、${if} 条件，以及 ${var | pipe} 链式过滤器（formatDate、formatCurrency、trim、upper、truncate、default）。
---

<p align="center">
  <a href="https://pkg.go.dev/github.com/yunkeweb/go-word"><img src="https://pkg.go.dev/badge/github.com/yunkeweb/go-word.svg" alt="Go Reference" /></a>
  <a href="https://github.com/yunkeweb/go-word/actions/workflows/test.yml"><img src="https://github.com/yunkeweb/go-word/actions/workflows/test.yml/badge.svg" alt="CI" /></a>
  <a href="https://github.com/yunkeweb/go-word/blob/main/LICENSE"><img src="https://img.shields.io/badge/License-LGPL%20v3-blue.svg" alt="License: LGPL v3" /></a>
  <a href="https://go.dev/"><img src="https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go" alt="Go 1.21+" /></a>
  <a href="https://github.com/yunkeweb/go-word/releases/tag/v0.8.0"><img src="https://img.shields.io/badge/release-v0.8.0-green.svg" alt="v0.8.0" /></a>
</p>

## 特性对比

GoWord 把 PHPWord 的公开 API（`AddSection`、`AddText`、`IOFactory`、`TemplateProcessor`）落到纯 Go 的 OpenXML 写出器上。下表覆盖生产环境 Word 流水线最常问到的能力。

| 能力 | GoWord | PHPWord | unioffice | fumiama/go-docx | gingfrederik/docx |
| --- | :---: | :---: | :---: | :---: | :---: |
| 语言 | Go 1.21+ | PHP | Go（商业） | Go | Go |
| 协议 | LGPL v3 | LGPL v3 | 商业授权 | AGPL-3.0 | MIT |
| Go 第三方依赖 | **无** | 不适用 | 无（付费 SDK） | 无 | 无 |
| 创建 `.docx` | ✓ | ✓ | ✓ | ✓ | ✓ |
| 读取 / 往返 `.docx` | ✓ | ✓ | ✓ | ✓ | |
| LaTeX 生成原生 OMML | ✓ | | ✓ | | |
| DrawingML 图表（柱 / 折 / 饼 / 面积 / 组合） | ✓ | ✓ | ✓ | | |
| 矢量形状 `wps:wsp` | ✓ | ✓ | ✓ | ✓ | |
| 多栏 `w:cols` | ✓ | ✓ | ✓ | | |
| 嵌套表 + `vMerge` / `gridSpan` | ✓ | ✓ | ✓ | ✓ | |
| 模板 `${var \| pipe}` + `${block}` / `${if}` | ✓ | ✓ | 模板 | | |
| O(1) 流式提取 | ✓ | | | | |
| 样式 / 书签 / `rId` 隔离合并 | ✓ | | ✓ | | |
| 水印 + `w:documentProtection` | ✓ | ✓ | ✓ | | |

PHPWord 是 API 祖先。unioffice 是付费的多格式 Office SDK。两款轻量 Go 写出器覆盖段落（go-docx 另含图片与表格），没有 OMML、图表、流式提取与标识符安全合并。

## 快速开始

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
	sec.AddTitle("GoWord v0.8.0", 1)
	sec.AddMath(`\frac{a}{b}`)
	doc.AddShape(word.ShapeRoundRect, word.ShapeOptions{
		FillColor: "5B9BD5", Text: "DrawingML",
		Font: style.Font{Bold: true, Color: "FFFFFF"},
	})
	if err := doc.Save("hello.docx"); err != nil {
		log.Fatal(err)
	}
}
```

```sh
go get github.com/yunkeweb/go-word@v0.8.0
```

接着阅读 [安装](/zh/guide/installation) 与 [快速开始](/zh/guide/getting-started)。完整签名见 [pkg.go.dev](https://pkg.go.dev/github.com/yunkeweb/go-word)。
