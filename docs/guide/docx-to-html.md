# DOCX to HTML

GoWord can read a `.docx` package and render the reconstructed document as an HTML fragment or a standalone page. The conversion uses the standard library only and does not require Microsoft Word, LibreOffice, cgo, or an external runtime.

## File conversion

Use `RenderHTMLFile` when the source is a path:

```go
package main

import (
    "log"

    "github.com/yunkeweb/go-word/element"
    word "github.com/yunkeweb/go-word"
)

func main() {
    err := word.RenderHTMLFile("input.docx", "output.html", word.HTMLOptions{
        Standalone: true,
        IncludeCSS: true,
    })
    if err != nil {
        log.Fatal(err)
    }
}
```

For an already loaded document, call `RenderHTML`. For an `io.Reader`, use `RenderHTMLWithOptions`:

```go
html, err := doc.RenderHTML(word.HTMLOptions{
    Standalone: true,
    IncludeCSS: true,
    IncludeHeadersFooters: true,
})
```

`Standalone` adds the document shell and `IncludeCSS` adds the built-in layout rules. Leave both disabled when embedding the returned fragment into an existing page.

## Images and assets

Images are embedded as data URIs by default, which makes the returned HTML self-contained. To serve image files separately, use `HTMLImageURL` or provide an `ImageURL` callback:

```go
html, err := doc.RenderHTML(word.HTMLOptions{
    ImageMode: word.HTMLImageURL,
    ImageURL: func(img *element.Image) (string, error) {
        return "/assets/" + img.GetName(), nil
    },
})
```

Callback URLs are validated before they are written into HTML. Keep the callback output under your application's asset policy.

## Diagnostics and strict mode

Some Word features do not have a lossless HTML equivalent. Use diagnostics to inspect those elements:

```go
result, err := doc.RenderHTMLWithDiagnostics(word.HTMLOptions{})
if err != nil {
    log.Fatal(err)
}
for _, diagnostic := range result.Diagnostics {
    log.Printf("%s: %s", diagnostic.ElementType, diagnostic.Message)
}
```

Set `Strict: true` when an unsupported or missing resource should fail the conversion instead of producing a diagnostic placeholder. The renderer preserves paragraphs, rich text, headings, links, lists, tables, merged cells, images, section boundaries, headers, footers, page breaks, bookmarks, notes, and comments. Complex charts and shapes remain diagnostic-driven.

## Resource limits

When loading untrusted files, use `ReadOptions` with `RenderHTMLWithOptions` to bound archive size, part size, expansion, and entry count. This keeps DOCX ingestion predictable before HTML rendering begins.
