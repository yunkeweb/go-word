package main

import (
	"errors"
	"log"

	word "github.com/yunkeweb/go-word"
	"github.com/yunkeweb/go-word/style"
)

func main() {
	doc := word.New()
	sec := doc.AddSection()
	sec.AddTitle("DOCX to HTML", 1)
	sec.AddText("This document is rendered back to HTML by GoWord.")
	run := sec.AddTextRun()
	run.AddText("Bold text", style.Font{Bold: true})
	run.AddText(" and a ")
	run.AddLink("https://example.com", "link")
	run.AddInsertion(" tracked", "Example", "2026-10-18")

	if err := doc.Save("input.docx"); err != nil {
		log.Fatal(err)
	}
	limits := word.ReadOptions{
		MaxArchiveSize: 32 << 20,
		MaxPartSize:    16 << 20,
		MaxTotalSize:   128 << 20,
		MaxEntries:     2000,
	}
	loaded, err := word.LoadWithOptions("input.docx", limits)
	if err != nil {
		if errors.Is(err, word.ErrReadLimitExceeded) {
			log.Fatal("input.docx exceeds the configured ZIP budget")
		}
		log.Fatal(err)
	}
	result, err := loaded.RenderHTMLWithDiagnostics(word.HTMLOptions{
		Standalone:   true,
		Title:        "DOCX to HTML",
		IncludeCSS:   true,
		Strict:       true,
		RevisionMode: word.HTMLRevisionMarkup,
	})
	if err != nil {
		log.Fatal(err)
	}
	if err := writeFile("output.html", result.HTML); err != nil {
		log.Fatal(err)
	}
	log.Println("wrote output.html")
}
