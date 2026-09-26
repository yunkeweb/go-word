package word

import (
	"fmt"
	"io"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/yunkeweb/go-word/element"
)

func variableRegexp() *regexp.Regexp {
	return regexp.MustCompile(regexp.QuoteMeta(macroOpen) + `([^` + regexp.QuoteMeta(macroClose) + `]+)` + regexp.QuoteMeta(macroClose))
}

// CreateWriter returns a writer for the named format (PHPWord IOFactory::createWriter).
func CreateWriter(doc *Document, name string) (Writer, error) {
	if name == "" {
		name = "Word2007"
	}
	switch name {
	case "Word2007":
		return newWord2007Writer(doc), nil
	default:
		return nil, fmt.Errorf("%q is not a valid writer", name)
	}
}

// CreateReader returns a reader for the named format (PHPWord IOFactory::createReader).
func CreateReader(name string) (Reader, error) {
	if name == "" {
		name = "Word2007"
	}
	switch name {
	case "Word2007":
		return word2007Reader{}, nil
	default:
		return nil, fmt.Errorf("%q is not a valid reader", name)
	}
}

// Open loads a .docx from a filesystem path into an in-memory DOM.
func Open(filePath string) (*Document, error) {
	return Load(filePath)
}

// Read loads a .docx from r into an in-memory DOM.
func Read(r io.Reader) (*Document, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return LoadBytes(data)
}

// Load loads a document (PHPWord IOFactory::load).
// If readerName is omitted, the format is inferred from the file extension.
func Load(filename string, readerName ...string) (*Document, error) {
	name := "Word2007"
	if len(readerName) > 0 && readerName[0] != "" {
		name = readerName[0]
	} else {
		switch strings.ToLower(filepath.Ext(filename)) {
		case ".docx", ".docm":
			name = "Word2007"
		}
	}
	r, err := CreateReader(name)
	if err != nil {
		return nil, err
	}
	return r.Load(filename)
}

// ExtractVariables returns ${placeholder} names from a template file.
func ExtractVariables(filename string, readerName ...string) ([]string, error) {
	doc, err := Load(filename, readerName...)
	if err != nil {
		return nil, err
	}
	seen := map[string]struct{}{}
	var out []string
	walkDocument(doc, func(el element.Element) {
		for _, m := range variableRegexp().FindAllStringSubmatch(elementText(el), -1) {
			name := strings.TrimSpace(m[1])
			if _, ok := seen[name]; ok {
				continue
			}
			seen[name] = struct{}{}
			out = append(out, name)
		}
	})
	return out, nil
}
