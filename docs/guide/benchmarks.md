# Benchmarks

All numbers come from `go test -bench` on the library’s own `bench_test.go`. They measure **writer / template / stream** cost, not Microsoft Word.

## Environment

| Item | Value |
| --- | --- |
| CPU | Intel Core i7-10870H @ 2.20 GHz (16 threads reported by Go) |
| OS | Windows amd64 |
| Go | 1.21+ (module line) |
| Command | `go test -bench="BenchmarkSaveDocx\|BenchmarkTableRender\|BenchmarkTemplateProcess\|BenchmarkStreamWriter" -benchtime=30x -count=1` |

Re-run on your hardware before quoting the figures in a capacity plan. Allocations are the stable signal; wall time moves with disk and CPU.

## Results

| Benchmark | Workload | ns/op | B/op | allocs/op |
| --- | --- | ---: | ---: | ---: |
| `BenchmarkSaveDocx` | 30 paragraphs + 20×5 table, `WriteTo` | 1.88e6 | 163 KiB | 949 |
| `BenchmarkTableRender` | 50×8 table, `WriteTo` | 2.76e6 | 180 KiB | 1820 |
| `BenchmarkTemplateProcess` | two `${}` replacements via `NewTemplateProcessorBytes` | 1.43e6 | 238 KiB | 703 |
| `BenchmarkStreamWriter` | incremental paragraphs into a ZIP | 1.55e6 | 142 KiB | 627 |

`B/op` is extra heap per iteration, not the size of the `.docx`. Stream writer allocations stay flat as you add paragraphs because `document.xml` is flushed as it is produced.

## How to reproduce

```sh
go test -bench=BenchmarkSaveDocx -benchmem -count=3
go test -bench=BenchmarkStreamWriter -benchmem -count=3
```

The fixtures live in `bench_test.go`:

- `benchDocument` builds 30 styled paragraphs and a 20×5 table.
- `BenchmarkTableRender` builds a 50×8 table only.
- `BenchmarkTemplateProcess` compiles a two-placeholder template each iteration.
- `BenchmarkStreamWriter` uses `NewStreamWriter`.

## Complete example — time a Save

```go
package main

import (
	"fmt"
	"log"
	"time"

	"github.com/yunkeweb/go-word"
	"github.com/yunkeweb/go-word/style"
)

func main() {
	doc := word.New()
	sec := doc.AddSection()
	sec.AddTitle("Benchmark", 1)
	for i := 0; i < 30; i++ {
		sec.AddText("The quick brown fox jumps over the lazy dog.", style.Font{Size: 11})
	}
	tbl := sec.AddTable(style.Table{Width: 9000})
	for r := 0; r < 20; r++ {
		row := tbl.AddRow()
		for c := 0; c < 5; c++ {
			row.AddCell(1800).AddText("cell")
		}
	}
	start := time.Now()
	raw, err := doc.Bytes()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("docx %d bytes in %s\n", len(raw), time.Since(start))
}
```

For files that must not sit in RAM as a DOM, write with [NewStreamWriter](./streaming) and extract with `StreamExtractText`.
