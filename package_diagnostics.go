package word

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"path"
	"strings"
)

// PackageDiagnostic describes a recoverable package validation finding.
type PackageDiagnostic struct {
	Severity string // error, warning
	Code     string
	Part     string
	Message  string
}

// ValidatePackage validates a DOCX ZIP package without mutating it.
func ValidatePackage(data []byte) []PackageDiagnostic {
	var out []PackageDiagnostic
	add := func(sev, code, part, msg string) {
		out = append(out, PackageDiagnostic{Severity: sev, Code: code, Part: part, Message: msg})
	}
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		add("error", "invalid_zip", "", err.Error())
		return out
	}
	parts := map[string]bool{}
	for _, f := range zr.File {
		parts[path.Clean(strings.ReplaceAll(f.Name, "\\", "/"))] = true
	}
	for _, f := range zr.File {
		name := path.Clean(strings.ReplaceAll(f.Name, "\\", "/"))
		if strings.HasSuffix(strings.ToLower(name), ".xml") || strings.HasSuffix(strings.ToLower(name), ".rels") {
			r, e := f.Open()
			if e != nil {
				add("error", "part_open", name, e.Error())
				continue
			}
			if e = validateXML(r); e != nil {
				add("error", "invalid_xml", name, e.Error())
			}
			_ = r.Close()
		}
		if strings.HasSuffix(name, ".rels") {
			r, e := f.Open()
			if e != nil {
				continue
			}
			raw, _ := io.ReadAll(r)
			_ = r.Close()
			validateRelationships(raw, name, parts, add)
		}
	}
	if !parts["word/document.xml"] {
		add("error", "missing_document", "word/document.xml", "document part is missing")
	} else {
		for _, f := range zr.File {
			if path.Clean(f.Name) != "word/document.xml" {
				continue
			}
			r, _ := f.Open()
			raw, _ := io.ReadAll(r)
			_ = r.Close()
			validateDocumentAnchors(raw, add)
		}
	}
	return out
}

func validateXML(r io.Reader) error {
	dec := xml.NewDecoder(r)
	for {
		_, err := dec.Token()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
	}
}

func validateRelationships(data []byte, relPart string, parts map[string]bool, add func(string, string, string, string)) {
	dec := xml.NewDecoder(bytes.NewReader(data))
	seen := map[string]bool{}
	base := path.Dir(path.Dir(relPart))
	if strings.HasSuffix(relPart, "/_rels/.rels") {
		base = ""
	}
	for {
		tok, err := dec.Token()
		if err != nil {
			return
		}
		se, ok := tok.(xml.StartElement)
		if !ok || localName(se.Name) != "Relationship" {
			continue
		}
		id, target, mode := attr(se, "Id"), attr(se, "Target"), attr(se, "TargetMode")
		if id == "" {
			continue
		}
		if seen[id] {
			add("error", "duplicate_relationship_id", relPart, fmt.Sprintf("duplicate relationship id %s", id))
		}
		seen[id] = true
		if mode == "External" || target == "" {
			continue
		}
		resolved := path.Clean(path.Join(base, target))
		if !parts[resolved] {
			add("error", "missing_relationship_target", relPart, fmt.Sprintf("target %s is missing", target))
		}
	}
}

func validateDocumentAnchors(data []byte, add func(string, string, string, string)) {
	dec := xml.NewDecoder(bytes.NewReader(data))
	bookmarks := map[string]bool{}
	var anchors []string
	for {
		tok, err := dec.Token()
		if err != nil {
			break
		}
		se, ok := tok.(xml.StartElement)
		if !ok {
			continue
		}
		switch localName(se.Name) {
		case "bookmarkStart":
			if n := attr(se, "name"); n != "" {
				if bookmarks[n] {
					add("error", "duplicate_bookmark", "word/document.xml", fmt.Sprintf("bookmark %s appears more than once", n))
				}
				bookmarks[n] = true
			}
		case "hyperlink":
			if a := attr(se, "anchor"); a != "" {
				anchors = append(anchors, a)
			}
		}
	}
	for _, a := range anchors {
		if !bookmarks[a] {
			add("warning", "missing_internal_anchor", "word/document.xml", fmt.Sprintf("internal link %s has no bookmark", a))
		}
	}
}
