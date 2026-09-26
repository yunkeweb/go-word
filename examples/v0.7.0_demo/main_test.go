package main

import (
	"os"
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
	for _, name := range []string{"v070_demo.docx", "template_v2.docx"} {
		if _, err := os.Stat(name); err != nil {
			t.Fatal(err)
		}
	}
}
