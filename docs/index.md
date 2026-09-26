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
      text: GitHub
      link: https://github.com/yunkeweb/go-word
    - theme: alt
      text: pkg.go.dev
      link: https://pkg.go.dev/github.com/yunkeweb/go-word

features:
  - title: Zero External Dependencies
    details: 100% Go standard library (encoding/xml, archive/zip, image, sync). go.mod has no third-party require.
  - title: Office Math (OMML)
    details: AddMath turns LaTeX such as \frac{a}{b} and x^{2} into Word-native equations you can double-click to edit.
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
