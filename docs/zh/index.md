---
layout: home

hero:
  name: GoWord
  text: 纯 Go 的 OpenXML Word 库
  tagline: 用 Go 标准库创建、读取、填充与合并 Microsoft Word（.docx）—— 零第三方依赖。
  actions:
    - theme: brand
      text: 快速开始
      link: /zh/guide/getting-started
    - theme: alt
      text: 实战案例
      link: /zh/guide/examples
    - theme: alt
      text: GitHub
      link: https://github.com/yunkeweb/go-word

features:
  - title: 零第三方依赖
    details: 100% Go 标准库（encoding/xml、archive/zip、image、sync）。go.mod 不含任何外部 require。
  - title: Office Math (OMML)
    details: AddMath 将 \frac{a}{b}、x^{2} 等 LaTeX 转成 Word 原生 m:oMathPara 公式，可双击编辑。
  - title: 无损文档合并
    details: AppendDocument 自动重映射样式 ID、书签名与图片 rId，多份文档拼接后资源彼此隔离。
  - title: DrawingML 与图表
    details: 柱状/折线/饼图/面积图/堆叠图/双轴组合图，以及矢量形状与文本框（wps:wsp）。
  - title: 流式解析器
    details: StreamExtractText / StreamExtractImages 以 O(1) 额外内存遍历 .docx ZIP。
  - title: 模板引擎 v2
    details: 嵌套 ${block} 循环、${if} 条件，以及 ${var | pipe} 链式过滤器（formatDate、trim、upper 等）。
  - title: 高级排版
    details: 多节横纵向混排、多栏 w:cols、TOC、水印与只读文档保护。
---

<p align="center">
  <a href="https://pkg.go.dev/github.com/yunkeweb/go-word"><img src="https://pkg.go.dev/badge/github.com/yunkeweb/go-word.svg" alt="Go Reference" /></a>
  <a href="https://github.com/yunkeweb/go-word/actions/workflows/test.yml"><img src="https://github.com/yunkeweb/go-word/actions/workflows/test.yml/badge.svg" alt="CI" /></a>
  <a href="https://github.com/yunkeweb/go-word/blob/main/LICENSE"><img src="https://img.shields.io/badge/License-LGPL%20v3-blue.svg" alt="License: LGPL v3" /></a>
  <a href="https://go.dev/"><img src="https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go" alt="Go 1.21+" /></a>
  <a href="https://github.com/yunkeweb/go-word/releases/tag/v0.8.0"><img src="https://img.shields.io/badge/release-v0.8.0-green.svg" alt="v0.8.0" /></a>
</p>

## 快速入口

- [快速开始与安装](/zh/guide/getting-started) — `go get` 与第一份 `.docx`
- [核心 DOM](/zh/guide/basics) — 段落、嵌套表、图片、页眉页脚
- [Office Math](/zh/guide/math) — LaTeX → 原生 OMML
- [DrawingML](/zh/guide/drawing) — 图表、形状、文本框
- [实战案例库](/zh/guide/examples) — 学术报告、文档拼接、财务报表模板

包文档：[pkg.go.dev/github.com/yunkeweb/go-word](https://pkg.go.dev/github.com/yunkeweb/go-word)
