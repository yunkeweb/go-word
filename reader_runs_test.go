package word

import (
	"testing"

	"github.com/yunkeweb/go-word/style"
)

func TestReaderPreservesCellRunBoundaries(t *testing.T) {
	d := New()
	cell := d.AddSection().AddTable().AddRow().AddCell(1000)
	tr := cell.AddTextRun()
	tr.AddText("hel")
	tr.AddText("lo", style.Font{Bold: true})
	tr.AddLink("https://example.com", "world")
	cell.AddText("next paragraph")
	raw, err := d.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	loaded, err := LoadBytes(raw)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := loaded.ExtractText(), "helloworld\nnext paragraph"; got != want {
		t.Fatalf("got %q want %q", got, want)
	}
	again, err := loaded.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	reloaded, err := LoadBytes(again)
	if err != nil {
		t.Fatal(err)
	}
	if reloaded.ExtractText() != loaded.ExtractText() {
		t.Fatal("second round-trip changed text")
	}
}
