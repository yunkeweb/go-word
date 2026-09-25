package word

import (
	"bytes"
	"testing"

	"github.com/yunkeweb/go-word/element"
	"github.com/yunkeweb/go-word/style"
)

func benchDocument() *Document {
	doc := New()
	sec := doc.AddSection()
	sec.AddTitle("Benchmark", 1)
	for i := 0; i < 30; i++ {
		sec.AddText("The quick brown fox jumps over the lazy dog.", style.Font{Size: 11})
	}
	tbl := sec.AddTable(style.Table{Width: 9000})
	for r := 0; r < 20; r++ {
		row := tbl.AddRow()
		for c := 0; c < 5; c++ {
			row.AddCell(1800).AddText("cell")
		}
	}
	return doc
}

func BenchmarkSaveDocx(b *testing.B) {
	doc := benchDocument()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		if _, err := doc.WriteTo(&buf); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkTableRender(b *testing.B) {
	doc := New()
	sec := doc.AddSection()
	tbl := sec.AddTable()
	for r := 0; r < 50; r++ {
		row := tbl.AddRow()
		for c := 0; c < 8; c++ {
			row.AddCell(1000).AddText("cell")
		}
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		if _, err := doc.WriteTo(&buf); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkTemplateProcess(b *testing.B) {
	src := New()
	sec := src.AddSection()
	sec.AddText("Hello ${name}")
	sec.AddText("Company ${company}")
	raw, err := src.Bytes()
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tp, err := NewTemplateProcessorBytes(raw)
		if err != nil {
			b.Fatal(err)
		}
		tp.SetValue("name", "World")
		tp.SetValue("company", "GoWord")
		if _, err := tp.Bytes(); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkStreamWriter(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		sw := NewStreamWriter(&buf)
		for p := 0; p < 20; p++ {
			if err := sw.WriteParagraph("hello"); err != nil {
				b.Fatal(err)
			}
		}
		tbl := element.NewTable(nil)
		for r := 0; r < 10; r++ {
			row := tbl.AddRow()
			row.AddCell(1000).AddText("a")
			row.AddCell(1000).AddText("b")
		}
		if err := sw.WriteTable(tbl); err != nil {
			b.Fatal(err)
		}
		if err := sw.Close(); err != nil {
			b.Fatal(err)
		}
	}
}
