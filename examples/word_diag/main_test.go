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
	files := []string{
		"diag_a_paragraphs_tables.docx",
		"diag_b_streamwriter.docx",
		"diag_c_charts.docx",
		"diag_d_markdown_html.docx",
		"diag_e_template.docx",
		"diag_f_comments_revisions.docx",
	}
	for _, name := range files {
		if _, err := os.Stat(name); err != nil {
			t.Fatal(err)
		}
	}
}
