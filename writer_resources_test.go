package word

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yunkeweb/go-word/element"
	"github.com/yunkeweb/go-word/pkg/common"
)

func TestMissingResourcesReturnErrors(t *testing.T) {
	for _, kind := range []string{"body image", "header image", "footer image", "OLE"} {
		t.Run(kind, func(t *testing.T) {
			d := New()
			sec := d.AddSection()
			missing := filepath.Join(t.TempDir(), "missing.bin")
			switch kind {
			case "body image":
				sec.AddImage(missing)
			case "header image":
				sec.AddHeader().AddImage(missing)
			case "footer image":
				sec.AddFooter().AddImage(missing)
		case "OLE":
			sec.AddOLEObject(missing)
			}
			if _, err := d.Bytes(); !errors.Is(err, os.ErrNotExist) {
				t.Fatalf("Bytes() error = %v, want missing file", err)
			}
		})
	}
}

func TestStreamNestedResources(t *testing.T) {
	for _, mode := range []string{"table", "element"} {
		t.Run(mode, func(t *testing.T) {
			table := element.NewTable(nil)
			cell := table.AddRow().AddCell(1000)
			tr := cell.AddTextRun()
			tr.AddImageBytes("x.png", pngBytes(t))
			tr.AddLink("https://example.com/nested", "link")
			cell.AddChart("pie", []string{"A"}, []float64{1})
			cell.AddOLEObject("").Media.Data = []byte("ole-data")
			var out bytes.Buffer
			sw := NewStreamWriter(&out)
			var err error
			if mode == "table" {
				err = sw.WriteTable(table)
			} else {
				err = sw.WriteElement(table)
			}
			if err != nil {
				t.Fatal(err)
			}
			if err := sw.Close(); err != nil {
				t.Fatal(err)
			}
			zr, err := common.OpenZipBytes(out.Bytes())
			if err != nil {
				t.Fatal(err)
			}
			defer zr.Close()
			for _, name := range []string{"word/media/image1.png", "word/charts/chart1.xml", "word/embeddings/oleObject1.bin"} {
				if !zr.Has(name) {
					t.Errorf("missing %s", name)
				}
			}
			doc, _ := zr.ReadFile("word/document.xml")
			rels, _ := zr.ReadFile("word/_rels/document.xml.rels")
			for _, token := range []string{"w:drawing", "w:hyperlink", "c:chart"} {
				if !strings.Contains(string(doc), token) {
					t.Errorf("missing %s", token)
				}
			}
			if !bytes.Contains(rels, []byte("https://example.com/nested")) {
				t.Error("missing nested hyperlink relationship")
			}
		})
	}
}

func TestStreamResourceErrorRemainsVisibleOnClose(t *testing.T) {
	for _, kind := range []string{"image", "OLE", "table"} {
		t.Run(kind, func(t *testing.T) {
			var out bytes.Buffer
			sw := NewStreamWriter(&out)
			missing := filepath.Join(t.TempDir(), "missing.bin")
			var err error
			switch kind {
			case "image":
				err = sw.WriteElement(element.NewImage(missing, nil))
			case "OLE":
				err = sw.WriteElement(element.NewOLEObject(missing, nil))
			case "table":
				tb := element.NewTable(nil)
				tb.AddRow().AddCell(1000).AddImage(missing)
				err = sw.WriteTable(tb)
			}
			if !errors.Is(err, os.ErrNotExist) {
				t.Fatalf("write error = %v", err)
			}
			if !errors.Is(sw.Close(), os.ErrNotExist) {
				t.Fatal("Close lost resource error")
			}
		})
	}
}
