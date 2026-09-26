# Benchmarks

All numbers come from `go test -bench` on the library’s own `bench_test.go`. They measure **writer / template / stream** cost, not Microsoft Word. Streaming extract is shown with a `runtime.MemStats` occupancy program because extra heap stays O(1) relative to file size — a single `ns/op` would hide that.

Related: [O(1) Streaming Extractor](./streaming), [Enterprise Recipes](./recipes).

## Environment

| Item | Value |
| --- | --- |
| CPU | Intel Core i7-10870H @ 2.20 GHz (16 threads reported by Go) |
| OS | Windows amd64 |
| Go | 1.21+ (module line) |
| Command | `go test -bench="BenchmarkSaveDocx\|BenchmarkTableRender\|BenchmarkTemplateProcess\|BenchmarkStreamWriter" -benchtime=30x -count=1` |

Re-run on your hardware before quoting the figures in a capacity plan. Allocations are the stable signal; wall time moves with disk and CPU.

## Writer / template results

| Benchmark | Workload | ns/op | B/op | allocs/op |
| --- | --- | ---: | ---: | ---: |
| `BenchmarkSaveDocx` | 30 paragraphs + 20×5 table, `WriteTo` | 1.88e6 | 163 KiB | 949 |
| `BenchmarkTableRender` | 50×8 table, `WriteTo` | 2.76e6 | 180 KiB | 1820 |
| `BenchmarkTemplateProcess` | two `${}` replacements via `NewTemplateProcessorBytes` | 1.43e6 | 238 KiB | 703 |
| `BenchmarkStreamWriter` | incremental paragraphs into a ZIP | 1.55e6 | 142 KiB | 627 |

`B/op` is extra heap per iteration, not the size of the `.docx`. Stream writer allocations stay flat as you add paragraphs because `document.xml` is flushed as it is produced.

## Streaming parse and memory occupancy

`StreamExtractText` / `StreamExtractImages` scan the ZIP with `encoding/xml.Decoder`. Each paragraph string is handed to the callback and then dropped. Extra heap does **not** grow with page count the way `Load` / `LoadBytes` does.

| API | Builds DOM | Extra memory vs file size |
| --- | --- | --- |
| `Load` / `LoadBytes` | Yes | Whole tree |
| `StreamExtractText` / `StreamExtractImages` | No | O(1) |
| `NewStreamWriter` | Writes incrementally | Body XML is not fully buffered |

Prefer `*os.File` so `archive/zip` maps the package without copying it. A plain `io.Reader` is buffered first. Rewind or reopen before a second pass.

### Complete example — occupancy of a stream extract

Save as `main.go` and run `go run .`. The program builds a 200-paragraph document, then walks it twice: once with `StreamExtractText`, once with `LoadBytes`. Heap delta after `runtime.GC()` is the occupancy signal.

```go
package main

import (
	"bytes"
	"fmt"
	"log"
	"runtime"

	"github.com/yunkeweb/go-word"
	"github.com/yunkeweb/go-word/style"
)

func heap() uint64 {
	runtime.GC()
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return m.HeapAlloc
}

func main() {
	doc := word.New()
	sec := doc.AddSection()
	sec.AddTitle("Occupancy", 1)
	for i := 0; i < 200; i++ {
		sec.AddText("The quick brown fox jumps over the lazy dog.", style.Font{Size: 11})
	}
	raw, err := doc.Bytes()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("docx %d bytes\n", len(raw))

	before := heap()
	n := 0
	if err := word.StreamExtractText(bytes.NewReader(raw), func(p string) error {
		if p != "" {
			n++
		}
		return nil
	}); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("stream  paragraphs=%d  heapΔ=%d KiB\n", n, int64(heap()-before)/1024)

	before = heap()
	loaded, err := word.LoadBytes(raw)
	if err != nil {
		log.Fatal(err)
	}
	_ = loaded
	fmt.Printf("dom     heapΔ=%d KiB (tree retained)\n", int64(heap()-before)/1024)
}
```

On the same i7-10870H host the stream pass reports a heap delta of a few hundred KiB regardless of stretching the loop from 200 to a few thousand paragraphs. The DOM pass grows with the tree. Treat the printed Δ as a trend on *your* hardware — `HeapAlloc` is not a laboratory instrument.

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
