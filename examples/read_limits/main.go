package main

import (
	"bytes"
	"errors"
	"fmt"
	"log"

	word "github.com/yunkeweb/go-word"
)

func main() {
	doc := word.New()
	doc.AddSection().AddText("A document read with explicit ZIP budgets.")
	raw, err := doc.Bytes()
	if err != nil {
		log.Fatal(err)
	}

	limits := word.ReadOptions{
		MaxArchiveSize: 8 << 20,
		MaxPartSize:    4 << 20,
		MaxTotalSize:   16 << 20,
		MaxEntries:     256,
	}
	loaded, err := word.ReadWithOptions(bytes.NewReader(raw), limits)
	if err != nil {
		if errors.Is(err, word.ErrReadLimitExceeded) {
			log.Fatal("document exceeds the configured ZIP budget")
		}
		log.Fatal(err)
	}
	found := 0
	if err := loaded.StreamExtractTextWithOptions(bytes.NewReader(raw), func(text string) error {
		if text != "" {
			found++
		}
		return nil
	}, limits); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("loaded %d paragraph(s) under the configured ZIP budget\n", found)
}
