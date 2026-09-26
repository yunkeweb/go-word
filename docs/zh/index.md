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
      text: 实战案例
      link: /zh/guide/recipes
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

## 核心优势

生产环境的 Word 流水线通常卡在四件事：把二进制送进隔离网络、在数百页提取时把内存压平、写出 Word 能双击编辑的公式、以及在不引入第二套模板栈的前提下填充合同。

| 约束 | GoWord | 典型 PHP / Python 栈 | 其他 Go `.docx` 写出器 |
| --- | --- | --- | --- |
| 语言与依赖 | Go 1.21+ **仅标准库** | 运行时 + XML / ZIP 附加库 | 多为标准库，但表面是只写 |
| 提取内存 | `StreamExtractText` **O(1)** 额外堆 | 整包 / DOM 载入 | 整棵文档树 |
| 公式 | LaTeX → 原生 **OMML**（`m:oMathPara`） | 图片、OLE 或省略 | 省略 |
| 模板 | `${var \| pipe}` + `${block}` / `${if}` | PHPWord 宏或 Jinja→docx | 省略或字符串替换 |
| 合并隔离 | `AppendDocument` 重映射样式、书签、`rId` | ZIP 拷贝，`image1.png` 互相覆盖 | 省略 |
| 协议 | LGPL v3 | 混杂 | AGPL 或 MIT 只写 |

GoWord 把 PHPWord 的名字（`AddSection`、`AddText`、`IOFactory`、`TemplateProcessor`）落到纯 Go 的 OpenXML 写出器上。把这些优势拼在一起的三份可复制程序见 [企业级实战案例](/zh/guide/recipes)。

## 性能基准

数字来自 Intel Core i7-10870H、Windows amd64 上的 `go test -benchmem`。分配次数是较稳的信号；墙钟时间随磁盘与 CPU 波动。完整表格、硬件说明与流式提取占用示例见 [性能基准](/zh/guide/benchmarks)。

| 工作负载 | ns/op | B/op | allocs/op |
| --- | ---: | ---: | ---: |
| `BenchmarkSaveDocx` — 30 段 + 20×5 表 | 1.88e6 | 163 KiB | 949 |
| `BenchmarkTableRender` — 50×8 表 | 2.76e6 | 180 KiB | 1820 |
| `BenchmarkTemplateProcess` — 两处 `${}` 替换 | 1.43e6 | 238 KiB | 703 |
| `BenchmarkStreamWriter` — 增量 ZIP | 1.55e6 | 142 KiB | 627 |

`B/op` 是每次迭代的额外堆，不是 `.docx` 体积。流式提取在回调后丢弃段落缓冲，额外内存相对文件大小保持 O(1)。

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

接着阅读 [安装](/zh/guide/installation)、[快速开始](/zh/guide/getting-started)，以及三份 [企业级实战案例](/zh/guide/recipes)（合同 / 论文 / 无损拼接）。完整签名见 [pkg.go.dev](https://pkg.go.dev/github.com/yunkeweb/go-word)。
