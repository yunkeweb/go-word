# 安装

GoWord 是单一 Go 模块。写出器、读取器、模板引擎、合并器与公式编译器都在 `github.com/yunkeweb/go-word`。无 CGO、无需安装本机 Word，`go.mod` 也没有任何第三方 `require`。

## 环境要求

| 项 | 值 |
| --- | --- |
| Go | 1.21 或更高 |
| 操作系统 | Go 工具链支持的任意 GOOS |
| Microsoft Word | 可选。仅在打开生成的 `.docx` 时需要 |
| 协议 | [GNU LGPL v3](https://github.com/yunkeweb/go-word/blob/main/LICENSE) |

## 安装

```sh
go get github.com/yunkeweb/go-word@v0.8.0
```

该命令会在你的模块中写入带版本的 require，并把源码下载到模块缓存。

## 升级

```sh
go get -u github.com/yunkeweb/go-word@v0.8.0
```

CI 中请钉死 tag。模块代理索引见 [proxy.golang.org](https://proxy.golang.org/github.com/yunkeweb/go-word/@v/v0.8.0.info)。

## 导入

```go
import "github.com/yunkeweb/go-word"
```

签名里会出现的子包：

| 包 | 职责 |
| --- | --- |
| `github.com/yunkeweb/go-word` | `Document`、`TemplateProcessor`、`AppendDocument`、图表、公式、流式 API |
| `github.com/yunkeweb/go-word/style` | 字体、段落、表格、节、图片、图表样式 |
| `github.com/yunkeweb/go-word/element` | DOM 节点（`Section`、`Table`、`Cell`、`Header`、`Chart`） |
| `github.com/yunkeweb/go-word/pkg/math` | `ParseLaTeX`、`WriteOMML`、`WriteOMath` |

## 验证

保存为 `main.go` 后执行 `go run .`。`install-check.docx` 是一份合法的 Word 2007 包。

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

## 下一步

[快速开始](./getting-started) 会走一遍公式、形状与双栏。[架构设计](./architecture) 说明 ZIP 包如何组装。
