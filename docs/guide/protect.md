# Watermark & Document Protection

Text watermarks are Word-native VML (`PowerPlusWaterMarkObject`) injected into every section header. Document protection writes `w:documentProtection` with an Office SHA-1 / 100000-spin hash when a password is set.

## SetTextWatermark / SetImageWatermark

### Signature

```go
func (d *Document) SetTextWatermark(text string)
func (d *Document) SetImageWatermark(imageBytes []byte)
```

### Parameters

| Name | Type | Description |
| --- | --- | --- |
| `text` | `string` | Diagonal word-art. Empty string clears a previously set text watermark. |
| `imageBytes` | `[]byte` | Encoded PNG/JPEG. Empty slice clears the image watermark. |

### Notes

- Watermarks are applied at write time to each section header. Creating headers yourself is compatible: the writer still injects the VML.
- Image watermarks add a media part referenced from the header relationships.

## Protect

### Signature

```go
func (d *Document) Protect(editing, password string) error
```

### Parameters

| Name | Type | Description |
| --- | --- | --- |
| `editing` | `string` | `ProtectTypeReadOnly`, `ProtectTypeComments`, `ProtectTypeTrackedChanges`, `ProtectTypeForms`. Empty → read-only. |
| `password` | `string` | Optional. Non-empty values use the ECMA-376 document-protection hash. |

### Notes

- This is **document editing restriction**, not ZIP encryption. Anyone can still unzip the package.
- Word prompts for the password when the user tries to edit. The sample password in demos is `goword`.

## Complete example

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

### What Word shows

Opening `protected.docx` displays a grey diagonal **CONFIDENTIAL** behind the body. The status bar reports read-only. Restrict Editing lists the password hash; entering `goword` unlocks editing.
