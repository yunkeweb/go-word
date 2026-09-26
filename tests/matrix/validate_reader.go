//go:build ignore

// Temporary validator: reverse-parse every test_output_docs/*.docx through
// the exported DOM loaders (word.Open / word.Read — there is no ReadDOM)
// and the streaming extractors.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"
	"sort"
	"strings"

	"github.com/yunkeweb/go-word"
)

type row struct {
	file   string
	dom    string
	text   string
	images string
	detail string
}

func main() {
	dir := "test_output_docs"
	if len(os.Args) > 1 {
		dir = os.Args[1]
	}
	matches, err := filepath.Glob(filepath.Join(dir, "*.docx"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "glob: %v\n", err)
		os.Exit(2)
	}
	sort.Strings(matches)
	if len(matches) == 0 {
		fmt.Fprintf(os.Stderr, "no .docx in %s\n", dir)
		os.Exit(2)
	}

	pass, fail := 0, 0

	fmt.Printf("Go reader self-check: %d files in %s\n", len(matches), dir)
	fmt.Printf("%-42s  %-6s %-6s %-6s  %s\n", "FILE", "DOM", "STREAM", "IMAGES", "DETAIL")

	for _, path := range matches {
		r := validateOne(path)
		ok := r.dom == "PASS" && r.text == "PASS" && r.images == "PASS"
		if ok {
			pass++
		} else {
			fail++
		}
		fmt.Printf("%-42s  %-6s %-6s %-6s  %s\n", filepath.Base(path), r.dom, r.text, r.images, r.detail)
	}

	fmt.Printf("\nGo reader totals: %d PASS  %d FAIL  (%d files)\n", pass, fail, len(matches))
	if fail > 0 {
		os.Exit(1)
	}
}

func validateOne(path string) (r row) {
	r.file = filepath.Base(path)
	r.dom, r.text, r.images = "FAIL", "FAIL", "FAIL"
	var parts []string

	if err := withPanic(func() error {
		doc, err := word.Open(path)
		if err != nil {
			return fmt.Errorf("Open: %w", err)
		}
		if doc == nil {
			return fmt.Errorf("Open: nil document")
		}
		_ = doc.ExtractText()
		if _, err := doc.ExtractImages(); err != nil {
			return fmt.Errorf("ExtractImages: %w", err)
		}
		secs := doc.GetSections()
		if len(secs) == 0 {
			return fmt.Errorf("DOM has 0 sections")
		}

		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()
		viaRead, err := word.Read(f)
		if err != nil {
			return fmt.Errorf("Read: %w", err)
		}
		if viaRead == nil || len(viaRead.GetSections()) == 0 {
			return fmt.Errorf("Read: empty DOM")
		}
		parts = append(parts, fmt.Sprintf("sections=%d", len(secs)))
		return nil
	}); err != nil {
		r.detail = err.Error()
		return r
	}
	r.dom = "PASS"

	var nParas int
	if err := withPanic(func() error {
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()
		return word.StreamExtractText(f, func(string) error {
			nParas++
			return nil
		})
	}); err != nil {
		r.detail = err.Error()
		return r
	}
	r.text = "PASS"
	parts = append(parts, fmt.Sprintf("paras=%d", nParas))

	var nImgs int
	if err := withPanic(func() error {
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()
		return word.StreamExtractImages(f, func(word.ImageFile) error {
			nImgs++
			return nil
		})
	}); err != nil {
		r.detail = err.Error()
		return r
	}
	r.images = "PASS"
	parts = append(parts, fmt.Sprintf("streamImgs=%d", nImgs))
	r.detail = strings.Join(parts, " ")
	return r
}

func withPanic(fn func() error) (err error) {
	defer func() {
		if rec := recover(); rec != nil {
			err = fmt.Errorf("panic: %v\n%s", rec, debug.Stack())
		}
	}()
	return fn()
}
