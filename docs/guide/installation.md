# Installation

GoWord is a single Go module. The writer, reader, template engine, merger, and math compiler all live in `github.com/yunkeweb/go-word`. There is no CGO, no native Word installation, and no third-party `require` in `go.mod`.

## Requirements

| Item | Value |
| --- | --- |
| Go | 1.21 or later |
| Operating system | Any GOOS that the Go toolchain supports |
| Microsoft Word | Optional. Needed only to open the generated `.docx` |
| License | [GNU LGPL v3](https://github.com/yunkeweb/go-word/blob/main/LICENSE) |

## New

```sh
go get github.com/yunkeweb/go-word@v0.9.0
```

The command records a versioned require in your module and downloads the source into the module cache.

## Upgrade

```sh
go get -u github.com/yunkeweb/go-word@v0.9.0
```

Pin a tag in CI. Proxy indexes for this module live at [proxy.golang.org](https://proxy.golang.org/github.com/yunkeweb/go-word/@v/v0.9.0.info).

## Import

```go
import "github.com/yunkeweb/go-word"
```

Sub-packages you will see in signatures:

| Package | Role |
| --- | --- |
| `github.com/yunkeweb/go-word` | `Document`, `TemplateProcessor`, `AppendDocument`, charts, math, streaming |
| `github.com/yunkeweb/go-word/style` | Font, paragraph, table, section, image, chart styles |
| `github.com/yunkeweb/go-word/element` | DOM nodes (`Section`, `Table`, `Cell`, `Header`, `Chart`) |
| `github.com/yunkeweb/go-word/pkg/math` | `ParseLaTeX`, `WriteOMML`, `WriteOMath` |

## Verify

Save as `main.go` and run `go run .`. The file `install-check.docx` is a valid Word 2007 package.

```go
package main

import (
	"fmt"
	"log"

	"github.com/yunkeweb/go-word"
)

func main() {
	doc := word.New()
	sec := doc.AddSection()
	sec.AddText("GoWord installation check.")
	if err := doc.Save("install-check.docx"); err != nil {
		log.Fatal(err)
	}
	fmt.Println("wrote install-check.docx")
}
```

## Next

[Quick Start](./getting-started) walks through formulas, shapes, and two-column layout. [Architecture](./architecture) explains how the ZIP package is assembled.
