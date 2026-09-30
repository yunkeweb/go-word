package word

import (
	"bytes"
	"encoding/xml"
	"io"
	"strings"
	"testing"
)

func TestTemplateImagesPreserveRunsAndPartRelationships(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	sec.AddImageBytes("existing.png", pngBytes(t))
	sec.AddText("before ${image} middle ${image} after")
	sec.AddHeader().AddText("header ${image}")
	sec.AddFooter().AddText("footer ${image}")
	b, err := doc.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	tp, err := NewTemplateProcessorBytes(b)
	if err != nil {
		t.Fatal(err)
	}
	if err := tp.SetImageValueBytes("image", "picture.png", pngBytes(t)); err != nil {
		t.Fatal(err)
	}
	for _, part := range []string{"word/document.xml", "word/header1.xml", "word/footer2.xml"} {
		t.Run(part, func(t *testing.T) {
			dec := xml.NewDecoder(bytes.NewReader(tp.files[part]))
			var stack []string
			text := ""
			drawings := 0
			ids := map[string]bool{}
			var imageRefs []string
			for {
				tok, err := dec.Token()
				if err == io.EOF {
					break
				}
				if err != nil {
					t.Fatal(err)
				}
				switch v := tok.(type) {
				case xml.StartElement:
					if v.Name.Local == "drawing" {
						drawings++
						if len(stack) == 0 || stack[len(stack)-1] != "r" {
							t.Fatalf("drawing has parent %v, want run", stack)
						}
					}
					if v.Name.Local == "docPr" {
						id := attr(v, "id")
						if ids[id] {
							t.Errorf("duplicate drawing ID %s", id)
						}
						ids[id] = true
					}
					if v.Name.Local == "blip" {
						imageRefs = append(imageRefs, attr(v, "embed"))
					}
					stack = append(stack, v.Name.Local)
				case xml.EndElement:
					stack = stack[:len(stack)-1]
				case xml.CharData:
					if len(stack) > 0 && stack[len(stack)-1] == "t" {
						text += string(v)
					}
				}
			}
			if drawings == 0 {
				t.Fatal("no drawings")
			}
			if strings.Contains(text, "${image}") {
				t.Fatal("placeholder remains")
			}
			if part == "word/document.xml" && text != "before  middle  after" {
				t.Fatalf("surrounding text=%q", text)
			}
			relPart := strings.Replace(part, "word/", "word/_rels/", 1) + ".rels"
			rels := parseRelationshipTargets(tp.files[relPart])
			for _, id := range imageRefs {
				if rels[id] == "" {
					t.Errorf("unresolved %s in %s", id, relPart)
				}
			}
		})
	}
}
