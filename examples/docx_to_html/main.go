package main

import (
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

	if err := doc.Save("input.docx"); err != nil {
		log.Fatal(err)
	}
	html, err := word.RenderHTMLFile("input.docx", word.HTMLOptions{
		Standalone: true,
		Title:      "DOCX to HTML",
		IncludeCSS: true,
	})
	if err != nil {
		log.Fatal(err)
	}
	if err := writeFile("output.html", html); err != nil {
		log.Fatal(err)
	}
	log.Println("wrote output.html")
}
