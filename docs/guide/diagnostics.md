# Reader, Templates, and Diagnostics

The development line adds round-trip support for common Word parts and two small safety APIs.

## Instance defaults

Use `NewWithOptions` when different documents need different default fonts in the same process:

```go
doc := word.NewWithOptions(word.DocumentOptions{
    DefaultFontName: "Aptos", DefaultAsianFontName: "Microsoft YaHei",
    DefaultFontSize: 11, DefaultFontColor: "202124",
})
```

## Package diagnostics

`ValidatePackage` checks a DOCX ZIP without changing it and returns part-scoped findings for malformed XML, missing relationship targets, duplicate relationship IDs, duplicate bookmarks, and missing internal anchors.

```go
raw, _ := os.ReadFile("input.docx")
for _, finding := range word.ValidatePackage(raw) {
    log.Printf("%s %s %s: %s", finding.Severity, finding.Code, finding.Part, finding.Message)
}
```

The reader restores section geometry and common header/footer parts, note bodies, comment metadata, tracked insertions/deletions, list paragraph numbering, and inline pictures. For pictures it retains the relationship target, encoded bytes, `wp:extent` dimensions (EMU and pixels), and `wp:docPr@descr` alt text. `ValidatePackage` reports unsupported chart, shape, and text-box drawings as warnings; those drawing types are not silently treated as pictures. `TemplateProcessor.SetValue` also recognizes macros split across adjacent `w:r` / `w:t` nodes while retaining run properties.
