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
go get github.com/yunkeweb/go-word@v0.11.0
```

The command records a versioned require in your module and downloads the source into the module cache.

## Upgrade

```sh
go get -u github.com/yunkeweb/go-word@v0.11.0
```

Pin a tag in CI. Proxy indexes for this module live at [proxy.golang.org](https://proxy.golang.org/github.com/yunkeweb/go-word/@v/v0.11.0.info).

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

## Test and race detector

Normal builds and tests stay pure Go and do not require CGO:

```sh
go test ./...
go build ./...
```

The race detector is a separate test-time requirement. It enables CGO and needs a C compiler; it does not add a CGO dependency to applications that import GoWord. On Windows, Visual Studio provides MSVC (`cl.exe`) and optional LLVM/Clang, but it does not install GCC. For the most predictable Go race setup, install the MSYS2 UCRT64 MinGW-w64 GCC toolchain, add `C:\msys64\ucrt64\bin` to `PATH`, and run:

```powershell
$env:CGO_ENABLED = "1"
$env:CC = "gcc"
go test -race ./... -count=1
```

If you only need the library, keep `CGO_ENABLED=0` for regular builds and tests. GitHub Actions can run the same check on `ubuntu-latest`, where GCC is available by default.

## Next

[Quick Start](./getting-started) walks through formulas, shapes, and two-column layout. [Architecture](./architecture) explains how the ZIP package is assembled.
