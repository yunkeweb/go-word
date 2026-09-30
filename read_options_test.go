package word

import (
	"archive/zip"
	"bytes"
	"errors"
	"hash/crc32"
	"io"
	"math"
	"os"
	"path/filepath"
	"testing"
)

func TestReadOptionsAcrossEntryPoints(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	sec.AddText("budgeted input")
	sec.AddImageBytes("test.png", pngBytes(t))
	raw, err := doc.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	filename := filepath.Join(t.TempDir(), "input.docx")
	if err := os.WriteFile(filename, raw, 0600); err != nil {
		t.Fatal(err)
	}
	zr, err := zip.NewReader(bytes.NewReader(raw), int64(len(raw)))
	if err != nil {
		t.Fatal(err)
	}
	exact := ReadOptions{MaxArchiveSize: int64(len(raw)), MaxEntries: len(zr.File)}
	for _, f := range zr.File {
		size := int64(f.UncompressedSize64)
		exact.MaxTotalSize += size
		if size > exact.MaxPartSize {
			exact.MaxPartSize = size
		}
	}
	textFn := func(s string) error { return nil }
	imageFn := func(img ImageFile) error { return nil }
	readers := map[string]func(ReadOptions) error{
		"bytes":          func(o ReadOptions) error { _, err := LoadBytesWithOptions(raw, o); return err },
		"reader":         func(o ReadOptions) error { _, err := ReadWithOptions(bytes.NewReader(raw), o); return err },
		"load":           func(o ReadOptions) error { _, err := LoadWithOptions(filename, o); return err },
		"open":           func(o ReadOptions) error { _, err := OpenWithOptions(filename, o); return err },
		"template bytes": func(o ReadOptions) error { _, err := NewTemplateProcessorBytesWithOptions(raw, o); return err },
		"template file":  func(o ReadOptions) error { _, err := NewTemplateProcessorWithOptions(filename, o); return err },
		"stream text":    func(o ReadOptions) error { return StreamExtractTextWithOptions(bytes.NewReader(raw), textFn, o) },
		"stream images":  func(o ReadOptions) error { return StreamExtractImagesWithOptions(bytes.NewReader(raw), imageFn, o) },
		"stream plain": func(o ReadOptions) error {
			return StreamExtractTextWithOptions(struct{ io.Reader }{bytes.NewReader(raw)}, textFn, o)
		},
		"stream file": func(o ReadOptions) error {
			f, err := os.Open(filename)
			if err != nil {
				return err
			}
			defer f.Close()
			err = StreamExtractImagesWithOptions(f, imageFn, o)
			if _, statErr := f.Stat(); statErr != nil {
				t.Error("closed caller-owned file", statErr)
			}
			return err
		},
		"method text":   func(o ReadOptions) error { return doc.StreamExtractTextWithOptions(bytes.NewReader(raw), textFn, o) },
		"method images": func(o ReadOptions) error { return doc.StreamExtractImagesWithOptions(bytes.NewReader(raw), imageFn, o) },
	}
	budgets := map[string]ReadOptions{
		"archive": {MaxArchiveSize: exact.MaxArchiveSize - 1},
		"part":    {MaxPartSize: exact.MaxPartSize - 1},
		"total":   {MaxPartSize: exact.MaxPartSize, MaxTotalSize: exact.MaxTotalSize - 1},
		"entries": {MaxEntries: exact.MaxEntries - 1},
	}
	for name, read := range readers {
		t.Run(name, func(t *testing.T) {
			for _, o := range []ReadOptions{{}, exact, {MaxArchiveSize: math.MaxInt64, MaxPartSize: math.MaxInt64, MaxTotalSize: math.MaxInt64}} {
				if err := read(o); err != nil {
					t.Fatalf("allowed input: %v", err)
				}
			}
			for budget, o := range budgets {
				if err := read(o); !errors.Is(err, ErrReadLimitExceeded) {
					t.Errorf("%s: got %v", budget, err)
				}
			}
			for _, o := range []ReadOptions{{MaxArchiveSize: -1}, {MaxPartSize: -1}, {MaxTotalSize: -1}, {MaxEntries: -1}} {
				if err := read(o); !errors.Is(err, ErrInvalidReadOptions) {
					t.Errorf("negative limit: got %v", err)
				}
			}
		})
	}
}

type countedInput struct {
	r io.Reader
	n int
}

func (r *countedInput) Read(p []byte) (int, error) {
	n, err := r.r.Read(p)
	r.n += n
	return n, err
}

func TestReadOptionsLimitInputConsumption(t *testing.T) {
	for _, stream := range []bool{false, true} {
		r := &countedInput{r: bytes.NewReader(bytes.Repeat([]byte{1}, 4096))}
		o := ReadOptions{MaxArchiveSize: 64}
		var err error
		if stream {
			err = StreamExtractTextWithOptions(r, func(string) error { return nil }, o)
		} else {
			_, err = ReadWithOptions(r, o)
		}
		if !errors.Is(err, ErrReadLimitExceeded) || r.n != 65 {
			t.Fatalf("stream=%v: error=%v bytes=%d", stream, err, r.n)
		}
	}
}

func TestReadOptionsInvalidBeforeIO(t *testing.T) {
	r := &countedInput{r: bytes.NewReader([]byte("do not consume"))}
	_, err := ReadWithOptions(r, ReadOptions{MaxPartSize: -1})
	if !errors.Is(err, ErrInvalidReadOptions) || r.n != 0 {
		t.Fatalf("error=%v bytes=%d", err, r.n)
	}
	if err := StreamExtractTextWithOptions(nil, nil, ReadOptions{MaxEntries: -1}); !errors.Is(err, ErrInvalidReadOptions) {
		t.Errorf("text nil callback: %v", err)
	}
	if err := StreamExtractImagesWithOptions(nil, nil, ReadOptions{MaxEntries: -1}); !errors.Is(err, ErrInvalidReadOptions) {
		t.Errorf("image nil callback: %v", err)
	}
}

func TestReadOptionsTempCleanup(t *testing.T) {
	dir := t.TempDir()
	for _, key := range []string{"TMP", "TEMP", "TMPDIR"} {
		t.Setenv(key, dir)
	}
	doc := New()
	doc.AddSection().AddText("text")
	raw, err := doc.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	callbackErr := errors.New("stop")
	for _, tc := range []struct {
		opts     ReadOptions
		callback func(string) error
		want     error
	}{
		{ReadOptions{}, func(string) error { return nil }, nil},
		{ReadOptions{MaxArchiveSize: 64}, func(string) error { return nil }, ErrReadLimitExceeded},
		{ReadOptions{MaxEntries: 1}, func(string) error { return nil }, ErrReadLimitExceeded},
		{ReadOptions{}, func(string) error { return callbackErr }, callbackErr},
	} {
		err := StreamExtractTextWithOptions(struct{ io.Reader }{bytes.NewReader(raw)}, tc.callback, tc.opts)
		if !errors.Is(err, tc.want) {
			t.Errorf("got %v want %v", err, tc.want)
		}
		files, err := os.ReadDir(dir)
		if err != nil || len(files) != 0 {
			t.Fatalf("temporary files leaked: %v %v", files, err)
		}
	}
}

func TestReadOptionsKeepCorruptPartErrors(t *testing.T) {
	for _, bad := range []string{"word/document.xml", "word/_rels/document.xml.rels", "docProps/core.xml"} {
		t.Run(bad, func(t *testing.T) {
			var buf bytes.Buffer
			zw := zip.NewWriter(&buf)
			parts := map[string]string{
				"word/document.xml":            `<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main"><w:body><w:p><w:r><w:t>text</w:t></w:r></w:p></w:body></w:document>`,
				"word/_rels/document.xml.rels": `<Relationships/>`,
				"docProps/core.xml":            `<coreProperties/>`,
			}
			for name, s := range parts {
				data := []byte(s)
				crc := crc32.ChecksumIEEE(data)
				if name == bad {
					crc++
				}
				w, err := zw.CreateRaw(&zip.FileHeader{Name: name, Method: zip.Store, CRC32: crc, CompressedSize64: uint64(len(data)), UncompressedSize64: uint64(len(data))})
				if err != nil {
					t.Fatal(err)
				}
				if _, err := w.Write(data); err != nil {
					t.Fatal(err)
				}
			}
			if err := zw.Close(); err != nil {
				t.Fatal(err)
			}
			o := ReadOptions{MaxPartSize: 4096}
			if _, err := LoadBytesWithOptions(buf.Bytes(), o); !errors.Is(err, zip.ErrChecksum) {
				t.Errorf("DOM lost CRC error: %v", err)
			}
			if _, err := NewTemplateProcessorBytesWithOptions(buf.Bytes(), o); !errors.Is(err, zip.ErrChecksum) {
				t.Errorf("template lost CRC error: %v", err)
			}
		})
	}
}

func TestReadOptionsTotalOverflow(t *testing.T) {
	files := []*zip.File{{FileHeader: zip.FileHeader{UncompressedSize64: math.MaxUint64}}, {FileHeader: zip.FileHeader{UncompressedSize64: 1}}}
	if err := validateZipLimits(files, 0, ReadOptions{MaxTotalSize: math.MaxInt64}); !errors.Is(err, ErrReadLimitExceeded) {
		t.Fatal(err)
	}
}
