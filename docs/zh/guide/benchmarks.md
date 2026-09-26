# 内存基准测试

下列数字来自库内 `bench_test.go` 的 `go test -bench`，衡量的是**写出器 / 模板 / 流式写入**成本，不是 Microsoft Word 的打开速度。

## 环境

| 项 | 值 |
| --- | --- |
| CPU | Intel Core i7-10870H @ 2.20 GHz（Go 报告 16 线程） |
| OS | Windows amd64 |
| Go | 1.21+（模块行） |
| 命令 | `go test -bench="BenchmarkSaveDocx\|BenchmarkTableRender\|BenchmarkTemplateProcess\|BenchmarkStreamWriter" -benchtime=30x -count=1` |

写入容量规划前请在目标硬件上重跑。分配次数是较稳的信号；墙钟时间随磁盘与 CPU 波动。

## 结果

| 基准 | 工作负载 | ns/op | B/op | allocs/op |
| --- | --- | ---: | ---: | ---: |
| `BenchmarkSaveDocx` | 30 段 + 20×5 表，`WriteTo` | 1.88e6 | 163 KiB | 949 |
| `BenchmarkTableRender` | 50×8 表，`WriteTo` | 2.76e6 | 180 KiB | 1820 |
| `BenchmarkTemplateProcess` | `NewTemplateProcessorBytes` 做两处 `${}` 替换 | 1.43e6 | 238 KiB | 703 |
| `BenchmarkStreamWriter` | 增量段落写入 ZIP | 1.55e6 | 142 KiB | 627 |

`B/op` 是每次迭代的额外堆，不是 `.docx` 体积。随着段落增加，流式写出器的分配保持平稳，因为 `document.xml` 边写边刷。

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
