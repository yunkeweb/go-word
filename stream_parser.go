package word

import (
	"archive/zip"
	"encoding/xml"
	"io"
	"os"
	"path"
	"strings"

	"github.com/yunkeweb/go-word/ooxml"
)

// StreamExtractText scans word/document.xml with xml.Decoder and invokes fn
// once for every complete w:p. Paragraph buffers are discarded after each
// callback so memory stays O(1) relative to document size.
func StreamExtractText(r io.Reader, fn func(paragraphText string) error) error {
	if fn == nil {
		return nil
	}
	zr, cleanup, err := openZipStream(r)
	if err != nil {
		return err
	}
	defer cleanup()
	f := zipFile(zr, "word/document.xml")
	if f == nil {
		return os.ErrNotExist
	}
	rc, err := f.Open()
	if err != nil {
		return err
	}
	defer rc.Close()
	return streamParagraphs(rc, fn)
}

// StreamExtractImages scans word/document.xml for DrawingML blips and VML
// imagedata, then streams each referenced media part from the ZIP into fn.
func StreamExtractImages(r io.Reader, fn func(img ImageFile) error) error {
	if fn == nil {
		return nil
	}
	zr, cleanup, err := openZipStream(r)
	if err != nil {
		return err
	}
	defer cleanup()
	rels, err := streamDocumentRels(zr)
	if err != nil {
		return err
	}
	f := zipFile(zr, "word/document.xml")
	if f == nil {
		return os.ErrNotExist
	}
	rc, err := f.Open()
	if err != nil {
		return err
	}
	defer rc.Close()
	return streamDrawings(rc, zr, rels, fn)
}

// StreamExtractText is the Document-method form of the package-level extractor.
func (d *Document) StreamExtractText(r io.Reader, fn func(paragraphText string) error) error {
	return StreamExtractText(r, fn)
}

// StreamExtractImages is the Document-method form of the package-level extractor.
func (d *Document) StreamExtractImages(r io.Reader, fn func(img ImageFile) error) error {
	return StreamExtractImages(r, fn)
}

type sizedReaderAt interface {
	io.ReaderAt
	Size() int64
}

func openZipStream(r io.Reader) (*zip.Reader, func(), error) {
	if r == nil {
		return nil, func() {}, io.ErrUnexpectedEOF
	}
	if sr, ok := r.(sizedReaderAt); ok {
		zr, err := zip.NewReader(sr, sr.Size())
		return zr, func() {}, err
	}
	if f, ok := r.(*os.File); ok {
		st, err := f.Stat()
		if err != nil {
			return nil, func() {}, err
		}
		zr, err := zip.NewReader(f, st.Size())
		return zr, func() {}, err
	}
	tmp, err := os.CreateTemp("", "goword-stream-*.docx")
	if err != nil {
		return nil, func() {}, err
	}
	if _, err := io.Copy(tmp, r); err != nil {
		tmp.Close()
		os.Remove(tmp.Name())
		return nil, func() {}, err
	}
	if _, err := tmp.Seek(0, io.SeekStart); err != nil {
		tmp.Close()
		os.Remove(tmp.Name())
		return nil, func() {}, err
	}
	st, err := tmp.Stat()
	if err != nil {
		tmp.Close()
		os.Remove(tmp.Name())
		return nil, func() {}, err
	}
	zr, err := zip.NewReader(tmp, st.Size())
	if err != nil {
		tmp.Close()
		os.Remove(tmp.Name())
		return nil, func() {}, err
	}
	return zr, func() {
		tmp.Close()
		os.Remove(tmp.Name())
	}, nil
}

func zipFile(zr *zip.Reader, name string) *zip.File {
	name = strings.ReplaceAll(name, "\\", "/")
	for _, f := range zr.File {
		if strings.ReplaceAll(f.Name, "\\", "/") == name {
			return f
		}
	}
	return nil
}

func streamParagraphs(r io.Reader, fn func(string) error) error {
	dec := xml.NewDecoder(r)
	var buf strings.Builder
	depth := 0
	inT := 0
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			if isWML(t.Name, "p") {
				if depth == 0 {
					buf.Reset()
				}
				depth++
			} else if depth > 0 && isWML(t.Name, "t") {
				inT++
			}
		case xml.EndElement:
			if isWML(t.Name, "t") && inT > 0 {
				inT--
			} else if isWML(t.Name, "p") {
				depth--
				if depth == 0 {
					if err := fn(buf.String()); err != nil {
						return err
					}
					buf.Reset()
					inT = 0
				}
			}
		case xml.CharData:
			if inT > 0 {
				buf.Write(t)
			}
		}
	}
}

func streamDrawings(r io.Reader, zr *zip.Reader, rels map[string]string, fn func(ImageFile) error) error {
	dec := xml.NewDecoder(r)
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		se, ok := tok.(xml.StartElement)
		if !ok {
			continue
		}
		rid := drawingRid(se)
		if rid == "" {
			continue
		}
		target, ok := rels[rid]
		if !ok {
			continue
		}
		part := resolveMediaTarget(target)
		if part == "" {
			continue
		}
		img, err := readZipImage(zr, part)
		if err != nil {
			if err == os.ErrNotExist {
				continue
			}
			return err
		}
		if err := fn(img); err != nil {
			return err
		}
	}
}

func drawingRid(se xml.StartElement) string {
	switch xmlLocal(se.Name) {
	case "blip":
		return attrLocal(se, "embed")
	case "imagedata":
		return attrLocal(se, "id")
	default:
		return ""
	}
}

func attrLocal(se xml.StartElement, local string) string {
	for _, a := range se.Attr {
		if a.Name.Local == local {
			return a.Value
		}
	}
	return ""
}

func xmlLocal(n xml.Name) string {
	if i := strings.IndexByte(n.Local, ':'); i >= 0 {
		return n.Local[i+1:]
	}
	return n.Local
}

func isWML(n xml.Name, local string) bool {
	if xmlLocal(n) != local {
		return false
	}
	if n.Space == ooxml.NSW || n.Space == "" {
		return true
	}
	return strings.HasPrefix(n.Local, "w:")
}

func streamDocumentRels(zr *zip.Reader) (map[string]string, error) {
	f := zipFile(zr, "word/_rels/document.xml.rels")
	if f == nil {
		return map[string]string{}, nil
	}
	rc, err := f.Open()
	if err != nil {
		return nil, err
	}
	defer rc.Close()
	out := map[string]string{}
	dec := xml.NewDecoder(rc)
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			return out, nil
		}
		if err != nil {
			return nil, err
		}
		se, ok := tok.(xml.StartElement)
		if !ok || se.Name.Local != "Relationship" {
			continue
		}
		var id, target string
		for _, a := range se.Attr {
			switch a.Name.Local {
			case "Id":
				id = a.Value
			case "Target":
				target = a.Value
			}
		}
		if id != "" && target != "" {
			out[id] = target
		}
	}
}

func resolveMediaTarget(target string) string {
	target = strings.ReplaceAll(strings.TrimSpace(target), "\\", "/")
	if target == "" || strings.Contains(target, "://") {
		return ""
	}
	if strings.HasPrefix(target, "/") {
		return strings.TrimPrefix(target, "/")
	}
	return path.Clean("word/" + target)
}

func readZipImage(zr *zip.Reader, name string) (ImageFile, error) {
	f := zipFile(zr, name)
	if f == nil {
		return ImageFile{}, os.ErrNotExist
	}
	rc, err := f.Open()
	if err != nil {
		return ImageFile{}, err
	}
	defer rc.Close()
	data, err := io.ReadAll(rc)
	if err != nil {
		return ImageFile{}, err
	}
	base := path.Base(name)
	ext := strings.ToLower(strings.TrimPrefix(path.Ext(base), "."))
	return ImageFile{Name: base, MIME: mimeForExt(ext), Data: data}, nil
}
