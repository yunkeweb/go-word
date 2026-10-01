package main

import (
	"os"
	"strings"
	"testing"
)

func TestMainRuns(t *testing.T) {
	dir := t.TempDir()
	old, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(old)

	main()
	raw, err := os.ReadFile("output.html")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), "DOCX to HTML") || !strings.Contains(string(raw), "<h1>") {
		t.Fatalf("unexpected HTML output: %s", raw)
	}
}
