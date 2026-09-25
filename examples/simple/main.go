package main

import (
	"log"

	"github.com/yunkeweb/go-word"
	"github.com/yunkeweb/go-word/style"
)

func main() {
	doc := word.New()
	doc.GetSettings().ThemeFontLang = "en-GB"

	doc.AddFontStyle("rStyle", style.Font{
		Bold: true, Italic: true, Size: 16,
		AllCaps: true, DoubleStrikethrough: true,
	})
	doc.AddParagraphStyle("pStyle", style.Paragraph{
		Alignment: style.JcCenter,
		Spacing:   style.Spacing{After: 100},
	})
	doc.AddTitleStyle(1, style.Font{Bold: true}, style.Paragraph{Spacing: style.Spacing{After: 240}})

	section := doc.AddSection()
	section.AddTitle("Welcome to GoWord", 1)
	section.AddText("Hello World!")
	section.AddTextBreak(2)
	section.AddText("I am styled by a font style definition.", "rStyle")
	section.AddText("I am styled by a paragraph style definition.", nil, "pStyle")
	section.AddText("I am styled by both font and paragraph style.", "rStyle", "pStyle")
	section.AddTextBreak()

	tr := section.AddTextRun()
	tr.AddText("I am inline styled ", style.Font{Name: "Times New Roman", Size: 20})
	tr.AddText("with ")
	tr.AddText("color", style.Font{Color: "996699"})
	tr.AddText(", ")
	tr.AddText("bold", style.Font{Bold: true})
	tr.AddText(", ")
	tr.AddText("italic", style.Font{Italic: true})
	tr.AddText(".")

	section.AddLink("https://github.com/PHPOffice/PHPWord", "PHPWord on GitHub")

	if err := doc.Save("helloWorld.docx"); err != nil {
		log.Fatal(err)
	}
}
