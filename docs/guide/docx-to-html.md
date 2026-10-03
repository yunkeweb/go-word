# DOCX to HTML

Available since **v0.12.0**, with DOCX style restoration fixes in **v0.12.1**, paper layout and hyperlink appearance fixes in **v0.12.2**, and fidelity improvements in **v0.13.0**. Render a loaded document or a DOCX file as an HTML fragment or standalone page using the Go standard library. Microsoft Word, LibreOffice, cgo, and external conversion processes are not required.

## Convert a file

Save this complete program as `main.go`, place `input.docx` next to it, and run `go run main.go`:

```go
package main

import (
	"log"
	"os"

	word "github.com/yunkeweb/go-word"
)

func main() {
	html, err := word.RenderHTMLFile("input.docx", word.HTMLOptions{
		Standalone: true,
		IncludeCSS: true,
	})
	if err != nil {
		log.Fatal(err)
	}
	if err := os.WriteFile("output.html", html, 0o644); err != nil {
		log.Fatal(err)
	}
}
```

**`RenderHTMLFile` returns HTML bytes; it does not accept an output filename or write a file.** The caller saves the result. A self-contained sample that also creates the input DOCX is available in [examples/docx_to_html](https://github.com/yunkeweb/go-word/tree/v0.13.0/examples/docx_to_html).

## Paper preview and flowing content

In **v0.12.2**, `Standalone: true` together with `IncludeCSS: true` uses each section's paper size, orientation and margins for a centered white preview. Text wraps within the DOCX content width instead of filling the browser window. Negative paragraph indents can extend into the page margin without being clipped at the viewport edge. No option changes are required for the complete program above.

The screen preview grows vertically for a long section; it does **not** reproduce Word's automatic page boundaries. Narrow screens scroll horizontally to preserve the document's line width. Printing uses named CSS `@page` rules for paper size and margins, subject to browser support and print settings. Fonts installed on the device, document grids, repeated headers/footers and Word's pagination rules can still cause visual differences.

For content that should flow within your application's own container, use `HTMLOptions{}` to return a fragment. For a standalone page without the paper preview, use `HTMLOptions{Standalone: true, IncludeCSS: false}` and provide your own CSS. Text and paragraph inline formatting remains available in both cases.

Explicit DOCX hyperlink colors and underline settings apply directly to the anchor, including automatic color and no underline, so browser link defaults do not override them.

## Fidelity improvements in v0.13.0

The v0.13.0 renderer preserves more layout metadata while keeping the existing fragment and standalone entry points compatible:

- Paragraph pagination hints map to `break-before`, `break-after`, `break-inside`, `widows`, and `orphans`. These are browser hints; the renderer does not expose a Word-compatible screen pagination engine.
- Standalone paper previews emit section-aware `goword-header` and `goword-footer` containers. A single section with one default header and footer enables fixed repetition for printing. First-page, even-page, and multi-section variants remain explicit containers and are not incorrectly repeated.
- Repeating table header rows use `<thead>`, body rows use `<tbody>`, `cantSplit` maps to `break-inside:avoid`, and fixed table layout, borders, spacing, cell padding, and row heights are emitted when present.
- Drawing images preserve EMU dimensions, alignment, margins, offsets, and common inline/square/tight/behind/in-front wrapping modes. Unknown wrapping values remain visible and produce an `HTMLDiagnostic` (or a strict-mode error).
- Theme font and color references are resolved from `word/theme/theme1.xml` when available. Direct font and color values take precedence; missing theme values are omitted instead of producing invalid CSS. `FallbackFont`, letter spacing, font scaling, character position, whitespace, custom tab size, and hyphenation hints are also emitted.

These additions improve readable screen output and browser printing, but they do not promise pixel-identical Word pagination. Validate representative documents with the target browser and print settings.

## API reference

These signatures belong to package `word`; `io.Reader` and `io.Writer` are standard-library interfaces.

```go
func RenderHTML(r io.Reader, opts HTMLOptions) ([]byte, error)
func RenderHTMLFile(path string, opts HTMLOptions) ([]byte, error)
func RenderHTMLWithOptions(r io.Reader, readOpts ReadOptions, htmlOpts HTMLOptions) ([]byte, error)
func RenderHTMLFileWithOptions(path string, readOpts ReadOptions, htmlOpts HTMLOptions) ([]byte, error)
func (d *Document) RenderHTML(opts HTMLOptions) ([]byte, error)
func (d *Document) RenderHTMLWithDiagnostics(opts HTMLOptions) (HTMLRenderResult, error)
func (d *Document) WriteHTML(w io.Writer, opts HTMLOptions) error
```

| Parameter | Meaning |
| --- | --- |
| `path` | Input DOCX filename |
| `r` | Reader containing the DOCX package, not HTML |
| `d` | Loaded or programmatically created document |
| `w` | Destination writer; the caller owns flushing and closing it |
| `readOpts` | Per-call ZIP limits; zero means unlimited |
| `opts` / `htmlOpts` | HTML output options below |

For in-memory DOCX bytes, use `RenderHTML(bytes.NewReader(raw), opts)`, or load them with `LoadBytesWithOptions` before inspecting conversion diagnostics.

### HTML options

| Field | Default | Behavior |
| --- | --- | --- |
| `Standalone` | `false` | Adds doctype, head and body; otherwise returns a fragment |
| `Title` | Empty | Uses document metadata when available; applies to standalone output |
| `IncludeCSS` | `false` | Adds built-in CSS and the paper preview only with `Standalone: true`. Basic inline styles are still emitted otherwise |
| `IncludeHeadersFooters` | `false` | Includes header/footer content present in each section's DOM |
| `ImageMode` | `HTMLImageDataURI` | Embeds image bytes; `HTMLImageURL` uses existing source/target URLs |
| `ImageURL` | `nil` | Callback with signature `func(*element.Image) (string, error)`; overrides `ImageMode` |
| `Strict` | `false` | Returns `*HTMLUnsupportedError` when the renderer records diagnostics |

Repeated rendering of an unchanged document produces the same bytes when callbacks are deterministic. Keep documents and their resources immutable while rendering concurrently.

## Stream HTML with read limits

This alternative complete program loads a DOCX with explicit ZIP budgets, then writes HTML directly to a file:

```go
package main

import (
	"log"
	"os"

	word "github.com/yunkeweb/go-word"
)

func main() {
	doc, err := word.LoadWithOptions("input.docx", word.ReadOptions{
		MaxArchiveSize: 32 << 20,
		MaxPartSize:    16 << 20,
		MaxTotalSize:   128 << 20,
		MaxEntries:     2000,
	})
	if err != nil {
		log.Fatal(err)
	}
	out, err := os.Create("output.html")
	if err != nil {
		log.Fatal(err)
	}
	renderErr := doc.WriteHTML(out, word.HTMLOptions{
		Standalone:            true,
		IncludeCSS:            true,
		IncludeHeadersFooters: true,
		Strict:                true,
	})
	closeErr := out.Close()
	if renderErr != nil {
		log.Fatal(renderErr)
	}
	if closeErr != nil {
		log.Fatal(closeErr)
	}
}
```

The limits are examples: choose values appropriate to your documents. Negative limits are rejected. Use `errors.Is(err, word.ErrReadLimitExceeded)` to recognize a read-budget failure.

`WriteHTML` avoids building a full HTML byte buffer, but DOCX loading still builds an in-memory DOM, and image encoding uses memory. A write failure or strict-mode error may occur **after partial output has been written**. Stage output in a temporary file and publish it only on success when incomplete output must never become visible. For an HTTP response that must return an error before headers are sent, render to bytes first.

## Images and assets

Data URI mode embeds available image data and emits alt text and dimensions from the DOM. It does not retrieve linked remote images.

For separate assets, supply `ImageURL` (import `github.com/yunkeweb/go-word/element`). The callback must persist or upload the image data and return a browser-accessible URL; it can use `img.Data`, `img.Media.Data`, `img.GetName()`, and `img.GetAltText()`. Generate unique asset names, validate image formats, and propagate storage errors from the callback. The renderer does not create an assets directory for you.

`HTMLImageURL` without a callback only copies an existing safe source/target URL. A DOCX media filename is not automatically a URL hosted by your application. Callback errors, unsafe URLs, and unavailable embedded images produce diagnostics in normal mode and fail strict mode. A URL passing the renderer's scheme check is not a guarantee that it exists or belongs to an approved host.

## Diagnostics

For an existing `doc`, import `errors` and `log` and inspect the result even when strict mode fails:

```go
result, err := doc.RenderHTMLWithDiagnostics(word.HTMLOptions{Strict: true})
for _, diagnostic := range result.Diagnostics {
	log.Printf("%s: %s", diagnostic.ElementType, diagnostic.Message)
}
var unsupported *word.HTMLUnsupportedError
if errors.As(err, &unsupported) {
	log.Printf("conversion has %d unsupported elements", len(unsupported.Diagnostics))
} else if err != nil {
	log.Fatal(err)
}
```

`HTMLDiagnostic` contains `ElementType` and `Message`. `RenderHTMLWithDiagnostics` retains diagnostic information and generated output on a strict-mode error. The byte-returning convenience APIs return an error without usable HTML in that case.

Diagnostics cover elements that reach the renderer. They are **not a complete OOXML fidelity audit**: content the reader does not reconstruct cannot be reported by this stage. Use [package diagnostics](./diagnostics) separately for package/XML/relationship issues.

## Supported content and limits

The style restoration described below requires v0.12.1 or later; paper layout and explicit hyperlink appearance fixes require v0.12.2.

| Content | Output and boundary |
| --- | --- |
| Paragraphs, headings, text runs, links | Semantic tags, inline breaks, fonts, sizes, colors, emphasis, underlining, shading, alignment, indentation, paragraph spacing, line spacing, pagination hints, custom tab size, and hyphenation hints. Body and table-cell paragraphs resolve document defaults, default paragraph styles, and named paragraph/character styles through `basedOn`; direct formatting overrides inherited values. Heading names and outline levels work with numeric style IDs. Common theme font/color references are resolved when the theme part is present; complex Word style features remain incomplete |
| Lists | Nested lists, common decimal/letter/Roman formats, Chinese counting, start values and continuation by numbering ID; custom composite labels and all Word restart rules are not fully reproduced |
| Tables | Nested tables and DOM merge properties map to HTML, including `rowspan` / `colspan`; fixed layout, borders, spacing, cell padding, repeating header rows and unbreakable rows are emitted when present; reading arbitrary DOCX merge properties remains limited |
| Bookmarks | DOM bookmarks become anchors and internal links target them |
| Sections and page breaks | Multiple sections use `section` and `data-break-type`; explicit page breaks use `goword-page-break`. Section break metadata does not implement Word pagination |
| Headers and footers | Optional section containers with `data-type` (`default`, `first`, `even`). A simple single-section default pair is repeated in print with fixed CSS; complex page-dependent selection is retained as explicit DOM and is not automatically simulated. Imported header/footer graphics remain limited |
| Images | Available drawing images with size/alt metadata; missing embedded references are retained for diagnostics. External linked drawings and VML are not fully supported |
| Notes, comments, shapes, formulas | DOM containers can expose text; imported references, geometry and equation layout are incomplete. Unsupported retained elements produce placeholders |

Conversion aims at readable HTML, not pixel-identical Word pages. A successful strict conversion does not guarantee lossless conversion. Review representative documents before adopting it for a new document family.

<a id="upgrading-to-v0-12-0"></a>
<a id="upgrading-to-v0-12-1"></a>

## Upgrading to v0.13.0

```sh
go get github.com/yunkeweb/go-word@v0.13.0
```

v0.13.0 preserves additional pagination hints, table layout metadata, image wrapping metadata, common theme font/color references, and simple default header/footer print repetition. Existing fragment output and zero-value `HTMLOptions` remain compatible. Screen output still does not implement Word automatic pagination. Review HTML snapshots and custom CSS when you depend on table borders, cell padding, theme fonts, or header/footer placement.
## Upgrading to v0.12.2

```sh
go get github.com/yunkeweb/go-word@v0.12.2
```

v0.12.2 adds paper-preview section wrappers and print CSS when both `Standalone` and `IncludeCSS` are enabled, and applies explicit hyperlink formatting to anchors. Update affected HTML snapshots and custom CSS. Use fragments or disable `IncludeCSS` if your application controls page layout.

When upgrading from v0.12.0 or earlier, v0.12.1 also restores supported DOCX styles and emits additional inline CSS for text and paragraph formatting. `style.Spacing` gained `BeforeSet` and `AfterSet`: set the corresponding flag to `true` to preserve an explicit zero margin. Use keyed struct literals; positional literals need updating.

The module path and pure-Go runtime requirements are unchanged. HTML APIs are new since v0.11.0. If you used development snapshots, update HTML snapshots/CSS for section wrappers, list `type`/`start` attributes, merged-cell attributes and properly closed page-break elements. Missing embedded-image references can now cause strict mode to fail.

`style.ListItem` adds `Start` (zero defaults to 1). Use keyed struct literals; positional literals from older versions need updating. `element.Container.AppendElement` appends an existing element and updates its parent; remove it from its old container first when moving it.

See [benchmarks](./benchmarks) and the [changelog](https://github.com/yunkeweb/go-word/blob/main/CHANGELOG.md).
