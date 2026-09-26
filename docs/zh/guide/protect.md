# 水印与只读保护

文字水印是注入每个节页眉的 Word 原生 VML（`PowerPlusWaterMarkObject`）。文档保护写出 `w:documentProtection`；设置密码时使用 Office SHA-1 / 100000 次迭代哈希。

## SetTextWatermark / SetImageWatermark

### 签名

```go
func (d *Document) SetTextWatermark(text string)
func (d *Document) SetImageWatermark(imageBytes []byte)
```

### 参数

| 名称 | 类型 | 说明 |
| --- | --- | --- |
| `text` | `string` | 斜向艺术字。空字符串清除先前的文字水印。 |
| `imageBytes` | `[]byte` | 已编码的 PNG/JPEG。空切片清除图片水印。 |

### 注意

- 水印在写出时应用到每个节页眉。你自己创建页眉也可以：写出器仍会注入 VML。
- 图片水印增加一个由页眉关系引用的媒体部件。

## Protect

### 签名

```go
func (d *Document) Protect(editing, password string) error
```

### 参数

| 名称 | 类型 | 说明 |
| --- | --- | --- |
| `editing` | `string` | `ProtectTypeReadOnly`、`ProtectTypeComments`、`ProtectTypeTrackedChanges`、`ProtectTypeForms`。空则只读。 |
| `password` | `string` | 可选。非空时使用 ECMA-376 文档保护哈希。 |

### 注意

- 这是**编辑限制**，不是 ZIP 加密。任何人仍可解压该包。
- 用户尝试编辑时 Word 会要密码。示例密码为 `goword`。

## 完整示例

```go
package main

import (
	"log"

	"github.com/yunkeweb/go-word"
)

func main() {
	doc := word.New()
	doc.SetDefaultFontName("Calibri")
	if err := doc.Protect(word.ProtectTypeReadOnly, "goword"); err != nil {
		log.Fatal(err)
	}
	doc.SetTextWatermark("CONFIDENTIAL")

	sec := doc.AddSection()
	sec.AddTitle("Protected report", 1)
	sec.AddText("The package is write-protected (w:documentProtection edit=readOnly). Sample password: goword.")
	sec.AddText("A diagonal CONFIDENTIAL watermark is stored as VML in the section header.")

	if err := doc.Save("protected.docx"); err != nil {
		log.Fatal(err)
	}
}
```

### Word 中的效果

打开 `protected.docx` 后，正文后方有灰色斜向 **CONFIDENTIAL**。状态栏显示只读。限制编辑列出密码哈希；输入 `goword` 即可解锁。
