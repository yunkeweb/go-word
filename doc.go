// Package word is a pure-Go port of PHPWord for creating, reading, and
// filling Microsoft Word documents.
//
// The public API follows PHPWord naming (AddSection, AddText, IOFactory)
// while using idiomatic Go types and error returns. The library depends
// only on the Go standard library.
//
//	doc := word.New()
//	section := doc.AddSection()
//	section.AddText("Hello World")
//	if err := doc.Save("hello.docx"); err != nil {
//	    log.Fatal(err)
//	}
package word
