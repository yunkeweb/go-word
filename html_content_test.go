package word

import (
	"archive/zip"
	"bytes"
	"errors"
	"regexp"
	"strings"
	"testing"

	"github.com/yunkeweb/go-word/element"
	"github.com/yunkeweb/go-word/style"
)

const contentWordNS = `http://schemas.openxmlformats.org/wordprocessingml/2006/main`

func contentDOCXFixture(t *testing.T, body, relationships string, parts map[string]string) []byte {
	t.Helper()
	files := map[string]string{
		"[Content_Types].xml":          `<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types"><Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/><Default Extension="xml" ContentType="application/xml"/><Override PartName="/word/document.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"/></Types>`,
		"_rels/.rels":                  `<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="word/document.xml"/></Relationships>`,
		"word/document.xml":            `<w:document xmlns:w="` + contentWordNS + `" xmlns:w14="http://schemas.microsoft.com/office/word/2010/wordml"><w:body>` + body + `</w:body></w:document>`,
		"word/_rels/document.xml.rels": `<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">` + relationships + `</Relationships>`,
	}
	for name, content := range parts {
		files[name] = content
	}
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	for name, content := range files {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := w.Write([]byte(content)); err != nil {
			t.Fatal(err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func TestDOCXHTMLContentBaseline(t *testing.T) {
	t.Run("fields", func(t *testing.T) {
		body := `<w:p><w:r><w:t>before </w:t></w:r>` +
			`<w:fldSimple w:instr=" DATE \\@ &quot;yyyy&quot; "><w:r><w:t>2026</w:t></w:r></w:fldSimple>` +
			`<w:r><w:t> page </w:t></w:r>` +
			`<w:r><w:fldChar w:fldCharType="begin"/></w:r>` +
			`<w:r><w:instrText xml:space="preserve"> PA</w:instrText></w:r>` +
			`<w:r><w:instrText>GE </w:instrText></w:r>` +
			`<w:r><w:fldChar w:fldCharType="separate"/></w:r>` +
			`<w:r><w:t>7</w:t></w:r>` +
			`<w:r><w:fldChar w:fldCharType="end"/></w:r>` +
			`<w:r><w:t> after</w:t></w:r></w:p>`
		doc, err := LoadBytes(contentDOCXFixture(t, body, "", nil))
		if err != nil {
			t.Fatal(err)
		}
		var fields []*element.Field
		walkDocument(doc, func(el element.Element) {
			if f, ok := el.(*element.Field); ok {
				fields = append(fields, f)
			}
		})
		if len(fields) != 2 {
			t.Fatalf("fields=%d", len(fields))
		}
		if fields[0].FieldType != "DATE" || fields[0].Text != "2026" || !fields[0].HasResult {
			t.Fatalf("simple field=%+v", fields[0])
		}
		if fields[1].FieldType != "PAGE" || strings.TrimSpace(fields[1].Instruction) != "PAGE" || fields[1].Text != "7" || !fields[1].HasResult {
			t.Fatalf("complex field=%+v", fields[1])
		}
		html, err := doc.RenderHTML(HTMLOptions{})
		if err != nil {
			t.Fatal(err)
		}
		got := string(html)
		for _, want := range []string{"before ", `data-field="DATE"`, ">2026</span>", " page ", `data-field="PAGE"`, ">7</span>", " after"} {
			if !strings.Contains(got, want) {
				t.Fatalf("HTML missing %q: %s", want, got)
			}
		}
	})

	t.Run("annotations", func(t *testing.T) {
		const relBase = `http://schemas.openxmlformats.org/officeDocument/2006/relationships/`
		rels := `<Relationship Id="rFoot" Type="` + relBase + `footnotes" Target="footnotes.xml"/>` +
			`<Relationship Id="rEnd" Type="` + relBase + `endnotes" Target="endnotes.xml"/>` +
			`<Relationship Id="rComments" Type="` + relBase + `comments" Target="comments.xml"/>`
		body := `<w:p><w:r><w:t>Text</w:t></w:r><w:commentRangeStart w:id="4"/><w:r><w:t>reviewed</w:t></w:r><w:commentRangeEnd w:id="4"/><w:r><w:commentReference w:id="4"/></w:r><w:r><w:footnoteReference w:id="2"/></w:r><w:r><w:endnoteReference w:id="3"/></w:r></w:p>`
		parts := map[string]string{
			"word/footnotes.xml": `<w:footnotes xmlns:w="` + contentWordNS + `"><w:footnote w:id="-1"><w:p><w:r><w:t>separator</w:t></w:r></w:p></w:footnote><w:footnote w:id="2"><w:p><w:r><w:t>foot body</w:t></w:r></w:p></w:footnote></w:footnotes>`,
			"word/endnotes.xml":  `<w:endnotes xmlns:w="` + contentWordNS + `"><w:endnote w:id="3"><w:p><w:r><w:t>end body</w:t></w:r></w:p></w:endnote></w:endnotes>`,
			"word/comments.xml":  `<w:comments xmlns:w="` + contentWordNS + `"><w:comment w:id="4" w:author="Alice" w:initials="A" w:date="2026-10-17T00:00:00Z"><w:p><w:r><w:t>comment body</w:t></w:r></w:p></w:comment></w:comments>`,
		}
		doc, err := LoadBytes(contentDOCXFixture(t, body, rels, parts))
		if err != nil {
			t.Fatal(err)
		}
		result, err := doc.RenderHTMLWithDiagnostics(HTMLOptions{})
		if err != nil {
			t.Fatal(err)
		}
		got := string(result.HTML)
		for _, want := range []string{
			`href="#goword-comment-4"`, `id="goword-comment-4"`, "comment body", "Alice",
			`href="#goword-footnote-2"`, `id="goword-footnote-2"`, "foot body",
			`href="#goword-endnote-3"`, `id="goword-endnote-3"`, "end body",
			`href="#goword-footnote-ref-2"`, `href="#goword-endnote-ref-3"`, `href="#goword-comment-ref-4"`,
		} {
			if !strings.Contains(got, want) {
				t.Fatalf("HTML missing %q: %s", want, got)
			}
		}
	})

	t.Run("content_controls", func(t *testing.T) {
		body := `<w:sdt><w:sdtPr><w:alias w:val="Customer"/><w:tag w:val="customer"/><w:id w:val="12"/><w:text w:multiLine="1"/></w:sdtPr><w:sdtContent><w:p><w:r><w:t>Ada</w:t></w:r></w:p></w:sdtContent></w:sdt>` +
			`<w:p><w:r><w:t>Accepted </w:t></w:r><w:sdt><w:sdtPr><w:alias w:val="Accepted"/><w:tag w:val="accepted"/><w:id w:val="13"/><w14:checkbox><w14:checked w14:val="1"/></w14:checkbox></w:sdtPr><w:sdtContent><w:r><w:t>☒</w:t></w:r></w:sdtContent></w:sdt></w:p>`
		doc, err := LoadBytes(contentDOCXFixture(t, body, "", nil))
		if err != nil {
			t.Fatal(err)
		}
		var controls []*element.SDT
		walkDocument(doc, func(el element.Element) {
			if s, ok := el.(*element.SDT); ok {
				controls = append(controls, s)
			}
		})
		if len(controls) != 2 || !controls[0].IsBlock || controls[0].Alias != "Customer" || controls[0].Value != "Ada" || controls[1].SDTType != element.SDTTypeCheckbox || !controls[1].Checked {
			t.Fatalf("controls=%+v", controls)
		}
		html, err := doc.RenderHTML(HTMLOptions{})
		if err != nil {
			t.Fatal(err)
		}
		got := string(html)
		for _, want := range []string{`class="goword-sdt"`, `data-sdt-type="plainText"`, `data-tag="customer"`, "Ada", `type="checkbox"`, "checked", "disabled"} {
			if !strings.Contains(got, want) {
				t.Fatalf("HTML missing %q: %s", want, got)
			}
		}
	})
	t.Run("rich_text_keeps_run_formatting", func(t *testing.T) {
		body := `<w:p><w:sdt><w:sdtPr><w:richText/></w:sdtPr><w:sdtContent><w:r><w:rPr><w:b/></w:rPr><w:t>bold</w:t></w:r><w:r><w:t> plain</w:t></w:r></w:sdtContent></w:sdt></w:p>`
		doc, err := LoadBytes(contentDOCXFixture(t, body, "", nil))
		if err != nil {
			t.Fatal(err)
		}
		html, err := doc.RenderHTML(HTMLOptions{})
		if err != nil {
			t.Fatal(err)
		}
		if (!strings.Contains(string(html), `font-weight:bold`) && !strings.Contains(string(html), `font-weight:700`)) || !strings.Contains(string(html), "plain") {
			t.Fatalf("SDT rich structure was flattened: %s", html)
		}
	})
	t.Run("unknown_content_control_diagnostic", func(t *testing.T) {
		body := `<w:p><w:sdt><w:sdtPr><w:alias w:val="Future"/></w:sdtPr><w:sdtContent><w:r><w:t>kept</w:t></w:r></w:sdtContent></w:sdt></w:p>`
		doc, err := LoadBytes(contentDOCXFixture(t, body, "", nil))
		if err != nil {
			t.Fatal(err)
		}
		result, err := doc.RenderHTMLWithDiagnostics(HTMLOptions{})
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(result.HTML), "kept") || len(result.Diagnostics) != 1 || result.Diagnostics[0].ElementType != "SDT" || result.Diagnostics[0].Part != "word/document.xml" || result.Diagnostics[0].ID == "" || result.Diagnostics[0].Path == "" {
			t.Fatalf("HTML=%s diagnostics=%+v", result.HTML, result.Diagnostics)
		}
	})
	t.Run("table_sdt_and_revision_combination", func(t *testing.T) {
		body := `<w:tbl><w:tr><w:tc><w:p><w:ins w:author="Alice" w:date="2026-10-18"><w:sdt><w:sdtPr><w:text/></w:sdtPr><w:sdtContent><w:r><w:t>table value</w:t></w:r></w:sdtContent></w:sdt></w:ins></w:p></w:tc></w:tr></w:tbl>`
		doc, err := LoadBytes(contentDOCXFixture(t, body, "", nil))
		if err != nil {
			t.Fatal(err)
		}
		final, err := doc.RenderHTML(HTMLOptions{RevisionMode: HTMLRevisionFinal})
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(final), "table value") || !strings.Contains(string(final), `data-sdt-type="plainText"`) {
			t.Fatalf("table SDT missing from final HTML: %s", final)
		}
		original, err := doc.RenderHTML(HTMLOptions{RevisionMode: HTMLRevisionOriginal})
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(string(original), "table value") {
			t.Fatalf("insertion leaked into original HTML: %s", original)
		}
		markup, err := doc.RenderHTML(HTMLOptions{RevisionMode: HTMLRevisionMarkup})
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(markup), `<ins data-author="Alice"`) || !strings.Contains(string(markup), "table value") {
			t.Fatalf("markup did not retain SDT revision: %s", markup)
		}
	})
}

func TestHTMLContentBudget(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	for i := 0; i < 128; i++ {
		run := sec.AddTextRun()
		run.AddText("content-")
		run.AddField("PAGE", nil, nil, "1")
		run.AddSDTText("item", "item", "value")
	}
	result, err := doc.RenderHTMLWithDiagnostics(HTMLOptions{})
	if err != nil {
		t.Fatal(err)
	}
	got := string(result.HTML)
	if strings.Count(got, "content-") != 128 || strings.Count(got, `class="goword-sdt"`) != 128 {
		t.Fatalf("content counts changed: text=%d sdt=%d", strings.Count(got, "content-"), strings.Count(got, `class="goword-sdt"`))
	}
	if len(result.Diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %+v", result.Diagnostics)
	}
}

func TestRenderHTMLRevisionModes(t *testing.T) {
	doc := New()
	run := doc.AddSection().AddTextRun()
	run.AddText("before ")
	run.AddInsertion("new", "Alice", "2026-10-17")
	run.AddDeletion("old", "Bob", "2026-10-16")
	run.AddText(" after")

	tests := []struct {
		name string
		mode HTMLRevisionMode
		want []string
		drop []string
	}{
		{name: "final", mode: HTMLRevisionFinal, want: []string{"new"}, drop: []string{"old", "<ins", "<del"}},
		{name: "original", mode: HTMLRevisionOriginal, want: []string{"old"}, drop: []string{"new", "<ins", "<del"}},
		{name: "markup", mode: HTMLRevisionMarkup, want: []string{`<ins data-author="Alice"`, "new</ins>", `<del data-author="Bob"`, "old</del>"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			html, err := doc.RenderHTML(HTMLOptions{RevisionMode: tc.mode})
			if err != nil {
				t.Fatal(err)
			}
			got := string(html)
			for _, want := range tc.want {
				if !strings.Contains(got, want) {
					t.Fatalf("HTML missing %q: %s", want, got)
				}
			}
			for _, drop := range tc.drop {
				if strings.Contains(got, drop) {
					t.Fatalf("HTML unexpectedly contains %q: %s", drop, got)
				}
			}
		})
	}
}

func TestRenderHTMLFieldWithoutCachedResultIsDiagnostic(t *testing.T) {
	doc := New()
	doc.AddSection().AddField("PAGE", nil, nil, "")
	result, err := doc.RenderHTMLWithDiagnostics(HTMLOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Diagnostics) != 1 || result.Diagnostics[0].ElementType != "Field" || result.Diagnostics[0].Part != "word/document.xml" {
		t.Fatalf("diagnostics=%+v", result.Diagnostics)
	}
	if !strings.Contains(string(result.HTML), "[PAGE]") {
		t.Fatalf("fallback missing: %s", result.HTML)
	}
	_, err = doc.RenderHTML(HTMLOptions{Strict: true})
	var unsupported *HTMLUnsupportedError
	if !errors.As(err, &unsupported) {
		t.Fatalf("strict error=%v", err)
	}
}

func TestRenderHTMLContentIntegrityEdges(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	item := sec.AddListItem("deleted item", 0, style.ListItem{NumId: 1, Depth: 0, ListType: style.ListTypeNumber})
	item.SetTrackChange(&element.TrackChange{ChangeType: "del", Author: "Bob"})
	footnote := element.NewFootnote(nil)
	footnote.NoteID = 2
	footnote.AddText("foot")
	doc.footnotes = append(doc.footnotes, footnote)
	endnote := element.NewEndnote(nil)
	endnote.NoteID = 3
	endnote.AddText("end")
	doc.endnotes = append(doc.endnotes, endnote)
	comment := &element.Comment{CommentID: 4, Author: "Alice"}
	comment.AddText("comment")
	doc.comments = append(doc.comments, comment)
	sec.AddTextRun().AppendElement(&element.NoteReference{NoteID: 2, Kind: "footnote"})
	sec.AddTextRun().AppendElement(&element.NoteReference{NoteID: 99, Kind: "footnote"})
	sec.AddTextRun().AppendElement(&element.CommentReference{CommentID: 404})

	original, err := doc.RenderHTML(HTMLOptions{RevisionMode: HTMLRevisionOriginal})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(original), "deleted item") {
		t.Fatalf("deleted list item missing from original view: %s", original)
	}
	final, err := doc.RenderHTML(HTMLOptions{RevisionMode: HTMLRevisionFinal})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(final), "deleted item") {
		t.Fatalf("deleted list item unexpectedly visible: %s", final)
	}

	result, err := doc.RenderHTMLWithDiagnostics(HTMLOptions{})
	if err != nil {
		t.Fatal(err)
	}
	got := string(result.HTML)
	for _, want := range []string{
		`href="#goword-footnote-2"`, `href="#goword-footnote-ref-2"`,
		`class="goword-missing-reference"`, `data-note-id="99"`, `data-comment-id="404"`,
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("HTML missing %q: %s", want, got)
		}
	}
	if strings.Contains(got, `href="#goword-footnote-99"`) || strings.Contains(got, `href="#goword-comment-404"`) {
		t.Fatalf("HTML contains dangling link: %s", got)
	}
	if len(result.Diagnostics) != 2 {
		t.Fatalf("diagnostics=%+v", result.Diagnostics)
	}
}

func TestReaderMalformedFieldPreservesFollowingText(t *testing.T) {
	body := `<w:p><w:r><w:fldChar w:fldCharType="begin"/></w:r><w:r><w:instrText>PAGE</w:instrText></w:r><w:r><w:t>kept text</w:t></w:r></w:p>`
	doc, err := LoadBytes(contentDOCXFixture(t, body, "", nil))
	if err != nil {
		t.Fatal(err)
	}
	result, err := doc.RenderHTMLWithDiagnostics(HTMLOptions{})
	if err != nil {
		t.Fatal(err)
	}
	got := string(result.HTML)
	if !strings.Contains(got, "kept text") {
		t.Fatalf("later text was lost: %s", got)
	}
	malformed := false
	for _, diagnostic := range result.Diagnostics {
		if strings.Contains(diagnostic.Message, "malformed complex field") {
			malformed = true
			break
		}
	}
	if !malformed {
		t.Fatalf("diagnostics=%+v", result.Diagnostics)
	}
}

func TestRenderHTMLAnnotationAnchorIntegrity(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	note := element.NewFootnote(nil)
	note.NoteID = 2
	note.AddText("note")
	doc.footnotes = append(doc.footnotes, note, note)
	orphan := element.NewEndnote(nil)
	orphan.NoteID = 3
	orphan.AddText("unreferenced")
	doc.endnotes = append(doc.endnotes, orphan)
	comment := &element.Comment{CommentID: 0, Author: "Alice"}
	comment.AddText("zero is a valid comment ID")
	doc.comments = append(doc.comments, comment)
	for i := 0; i < 2; i++ {
		run := sec.AddTextRun()
		run.AppendElement(&element.NoteReference{NoteID: 2, Kind: "footnote"})
		run.AppendElement(&element.CommentReference{CommentID: 0})
	}
	result, err := doc.RenderHTMLWithDiagnostics(HTMLOptions{})
	if err != nil {
		t.Fatal(err)
	}
	got := string(result.HTML)
	ids := map[string]bool{}
	for _, match := range regexp.MustCompile(`(?:^|\s)id="([^"]+)"`).FindAllStringSubmatch(got, -1) {
		if ids[match[1]] {
			t.Fatalf("duplicate HTML ID %q: %s", match[1], got)
		}
		ids[match[1]] = true
	}
	for _, match := range regexp.MustCompile(`href="#([^"]+)"`).FindAllStringSubmatch(got, -1) {
		if !ids[match[1]] {
			t.Fatalf("dangling HTML link %q: %s", match[1], got)
		}
	}
	if !strings.Contains(got, `href="#goword-comment-0"`) {
		t.Fatalf("comment ID zero was lost: %s", got)
	}
}
