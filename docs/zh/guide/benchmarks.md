# 性能基准

下列数字来自库内 `bench_test.go` 的 `go test -bench`，衡量的是**写出器 / 模板 / 流式写入**成本，不是 Microsoft Word 的打开速度。流式提取用 `runtime.MemStats` 占用程序展示：额外堆相对文件大小保持 O(1)，单次 `ns/op` 会把这一点淹没。

相关：[O(1) 流式提取器](./streaming)、[企业级实战案例](./recipes)。

## 环境

| 项 | 值 |
| --- | --- |
| CPU | Intel Core i7-10870H @ 2.20 GHz（Go 报告 16 线程） |
| OS | Windows amd64 |
| Go | 1.21+（模块行） |
| 命令 | `go test -bench="BenchmarkSaveDocx\|BenchmarkTableRender\|BenchmarkTemplateProcess\|BenchmarkStreamWriter" -benchtime=30x -count=1` |

写入容量规划前请在目标硬件上重跑。分配次数是较稳的信号；墙钟时间随磁盘与 CPU 波动。

## 写出器 / 模板结果

| 基准 | 工作负载 | ns/op | B/op | allocs/op |
| --- | --- | ---: | ---: | ---: |
| `BenchmarkSaveDocx` | 30 段 + 20×5 表，`WriteTo` | 1.88e6 | 163 KiB | 949 |
| `BenchmarkTableRender` | 50×8 表，`WriteTo` | 2.76e6 | 180 KiB | 1820 |
| `BenchmarkTemplateProcess` | `NewTemplateProcessorBytes` 做两处 `${}` 替换 | 1.43e6 | 238 KiB | 703 |
| `BenchmarkStreamWriter` | 增量段落写入 ZIP | 1.55e6 | 142 KiB | 627 |

`B/op` 是每次迭代的额外堆，不是 `.docx` 体积。随着段落增加，流式写出器的分配保持平稳，因为 `document.xml` 边写边刷。

## 流式解析与内存占用

`StreamExtractText` / `StreamExtractImages` 用 `encoding/xml.Decoder` 扫描 ZIP。每个段落字符串交给回调后即丢弃。额外堆**不会**像 `Load` / `LoadBytes` 那样随页数增长。

| API | 构建 DOM | 相对文件大小的额外内存 |
| --- | --- | --- |
| `Load` / `LoadBytes` | 是 | 整棵树 |
| `StreamExtractText` / `StreamExtractImages` | 否 | O(1) |
| `NewStreamWriter` | 增量写出 | 正文 XML 不会整段缓冲 |

优先传入 `*os.File`，这样 `archive/zip` 无需拷贝即可映射。普通 `io.Reader` 会先被缓冲。第二次遍历前请回绕或重新打开。

### 完整示例 — 流式提取的占用

保存为 `main.go`，执行 `go run .`。程序先构建 200 段文档，再走两遍：一遍 `StreamExtractText`，一遍 `LoadBytes`。`runtime.GC()` 之后的堆增量就是占用信号。

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

在同一台 i7-10870H 上，把循环从 200 拉到几千段，流式那一趟的堆增量仍是数百 KiB 量级。DOM 那一趟随树增长。把打印出的 Δ 当成*你这台机器*上的趋势——`HeapAlloc` 不是实验室仪器。

## 如何复现

```sh
go test -bench=BenchmarkSaveDocx -benchmem -count=3
go test -bench=BenchmarkStreamWriter -benchmem -count=3
```

夹具在 `bench_test.go`：

- `benchDocument` 构建 30 个带样式段落和一张 20×5 表。
- `BenchmarkTableRender` 只构建 50×8 表。
- `BenchmarkTemplateProcess` 每次迭代编译含两个占位符的模板。
- `BenchmarkStreamWriter` 使用 `NewStreamWriter`。

## 完整示例 — 给一次 Save 计时

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

不能把整棵 DOM 放进内存时，用 [NewStreamWriter](./streaming) 写出，用 `StreamExtractText` 提取。
