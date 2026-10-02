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
go get github.com/yunkeweb/go-word@v0.12.1
```

该命令会在你的模块中写入带版本的 require，并把源码下载到模块缓存。

## 升级

```sh
go get github.com/yunkeweb/go-word@v0.12.1
```

CI 中请钉死 tag。模块代理索引见 [proxy.golang.org](https://proxy.golang.org/github.com/yunkeweb/go-word/@v/v0.12.1.info)。请阅读 [v0.12.1 升级说明](./docx-to-html#升级至-v0-12-1)，留意 `style.ListItem` / `style.Spacing` 的具名字段初始化及 HTML 保真限制。

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

## 测试与 race 检测器

普通构建和测试保持纯 Go，不需要启用 CGO：

```sh
go test ./...
go build ./...
```

race 检测器只在测试时需要额外的 C 编译器。它不会给依赖 GoWord 的应用增加运行时 CGO 依赖。Windows 下的 Visual Studio 提供 MSVC（`cl.exe`）和可选的 LLVM/Clang，但不会安装 GCC。为了更稳定地运行 Go race，建议安装 MSYS2 的 UCRT64 MinGW-w64 GCC 工具链，将 `C:\msys64\ucrt64\bin` 加入 `PATH`，然后执行：

```powershell
$env:CGO_ENABLED = "1"
$env:CC = "gcc"
go test -race ./... -count=1
```

如果只使用 GoWord 库，普通构建和测试保持 `CGO_ENABLED=0` 即可。GitHub Actions 可以在 `ubuntu-latest` 上执行同样的检查，该 runner 默认提供 GCC。

## 下一步

[快速开始](./getting-started) 会走一遍公式、形状与双栏。[架构设计](./architecture) 说明 ZIP 包如何组装。
