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
      text: GitHub
      link: https://github.com/yunkeweb/go-word
    - theme: alt
      text: pkg.go.dev
      link: https://pkg.go.dev/github.com/yunkeweb/go-word

features:
  - title: 零第三方依赖
    details: 100% Go 标准库（encoding/xml、archive/zip、image、sync）。go.mod 不含任何外部 require。
  - title: Office Math (OMML)
    details: AddMath 将 \frac{a}{b}、x^{2} 等 LaTeX 转成 Word 原生公式，可在公式编辑器中双击编辑。
  - title: 无损文档合并
    details: AppendDocument 自动重映射样式 ID、书签名与图片 rId，多份文档拼接后资源彼此隔离。
  - title: DrawingML 与图表
    details: 柱状/折线/饼图/面积图/堆叠图/双轴组合图，以及矢量形状与文本框（wps:wsp）。
  - title: 流式提取器
    details: StreamExtractText / StreamExtractImages 以 O(1) 额外内存遍历 .docx ZIP。
  - title: 模板引擎 v2
    details: 嵌套 ${block} 循环、${if} 条件，以及 ${var | pipe} 链式过滤器（formatDate、trim、upper 等）。
  - title: 高级排版
    details: 多节横纵向混排、多栏 w:cols、TOC、水印与只读文档保护。
---
