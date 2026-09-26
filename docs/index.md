---
layout: home

hero:
  name: GoWord
  text: Pure Go OpenXML Word Library
  tagline: Create, read, fill, and merge Microsoft Word (.docx) files with the Go standard library — zero third-party dependencies.
  actions:
    - theme: brand
      text: Get Started
      link: /guide/getting-started
    - theme: alt
      text: Examples & Recipes
      link: /guide/examples
    - theme: alt
      text: GitHub
      link: https://github.com/yunkeweb/go-word

features:
  - title: Zero External Dependencies
    details: 100% Go standard library (encoding/xml, archive/zip, image, sync). go.mod has no third-party require.
  - title: Office Math (OMML)
    details: AddMath turns LaTeX such as \frac{a}{b} and x^{2} into Word-native m:oMathPara equations you can double-click to edit.
  - title: Document Merger
    details: AppendDocument remaps style IDs, bookmark names, and image rIds so several documents splice without clashes.
  - title: DrawingML & Charts
    details: Bar, line, pie, area, stacked, and dual-axis combo charts, plus vector shapes and text boxes (wps:wsp).
  - title: Streaming Parser
    details: StreamExtractText and StreamExtractImages walk a .docx ZIP with O(1) extra memory.
  - title: Template Engine v2
    details: Nested ${block} loops, ${if} comparisons, and chained ${var | pipe} filters (formatDate, trim, upper, …).
  - title: Advanced Layout
    details: Mixed portrait/landscape sections, multi-column w:cols, TOC, watermarks, and read-only protection.
---

<p align="center">
  <a href="https://pkg.go.dev/github.com/yunkeweb/go-word"><img src="https://pkg.go.dev/badge/github.com/yunkeweb/go-word.svg" alt="Go Reference" /></a>
  <a href="https://github.com/yunkeweb/go-word/actions/workflows/test.yml"><img src="https://github.com/yunkeweb/go-word/actions/workflows/test.yml/badge.svg" alt="CI" /></a>
  <a href="https://github.com/yunkeweb/go-word/blob/main/LICENSE"><img src="https://img.shields.io/badge/License-LGPL%20v3-blue.svg" alt="License: LGPL v3" /></a>
  <a href="https://go.dev/"><img src="https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go" alt="Go 1.21+" /></a>
  <a href="https://github.com/yunkeweb/go-word/releases/tag/v0.8.0"><img src="https://img.shields.io/badge/release-v0.8.0-green.svg" alt="v0.8.0" /></a>
</p>

## Quick links

- [Quick Start & Installation](/guide/getting-started) — `go get` and a first `.docx`
- [Core DOM](/guide/basics) — paragraph, nested table, image, header/footer
- [Office Math](/guide/math) — LaTeX → native OMML
- [DrawingML](/guide/drawing) — charts, shapes, text boxes
- [Examples & Recipes](/guide/examples) — academic report, merger, finance template

API reference: [pkg.go.dev/github.com/yunkeweb/go-word](https://pkg.go.dev/github.com/yunkeweb/go-word)
