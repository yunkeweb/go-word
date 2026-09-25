package common

import (
	"archive/zip"
	"bytes"
	"io"
	"os"
	"path"
	"strings"
)

// ZipWriter is a thin archive/zip helper for OOXML packages.
type ZipWriter struct {
	zw *zip.Writer
}

// NewZipWriter wraps w.
func NewZipWriter(w io.Writer) *ZipWriter {
	return &ZipWriter{zw: zip.NewWriter(w)}
}

// AddFile writes name with DEFLATE compression.
func (z *ZipWriter) AddFile(name string, data []byte) error {
	w, err := z.Create(name)
	if err != nil {
		return err
	}
	_, err = w.Write(data)
	return err
}

// Create opens a streaming zip entry. Only one entry may be open at a time.
func (z *ZipWriter) Create(name string) (io.Writer, error) {
	name = path.Clean(strings.ReplaceAll(name, "\\", "/"))
	return z.zw.Create(name)
}

// Close flushes the zip central directory.
func (z *ZipWriter) Close() error { return z.zw.Close() }

// ZipReader opens an OOXML package from a file or in-memory buffer.
type ZipReader struct {
	zr *zip.Reader
	f  *os.File
}

// OpenZipFile opens a zip from disk.
func OpenZipFile(filename string) (*ZipReader, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	st, err := f.Stat()
	if err != nil {
		f.Close()
		return nil, err
	}
	zr, err := zip.NewReader(f, st.Size())
	if err != nil {
		f.Close()
		return nil, err
	}
	return &ZipReader{zr: zr, f: f}, nil
}

// OpenZipBytes opens a zip from memory.
func OpenZipBytes(data []byte) (*ZipReader, error) {
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, err
	}
	return &ZipReader{zr: zr}, nil
}

// Close closes the underlying file, if any.
func (z *ZipReader) Close() error {
	if z.f != nil {
		return z.f.Close()
	}
	return nil
}

// Files returns the zip directory.
func (z *ZipReader) Files() []*zip.File { return z.zr.File }

// ReadFile returns the contents of name, or nil if missing.
func (z *ZipReader) ReadFile(name string) ([]byte, error) {
	name = strings.ReplaceAll(name, "\\", "/")
	for _, f := range z.zr.File {
		if f.Name == name {
			rc, err := f.Open()
			if err != nil {
				return nil, err
			}
			defer rc.Close()
			return io.ReadAll(rc)
		}
	}
	return nil, os.ErrNotExist
}

// Has reports whether name exists in the archive.
func (z *ZipReader) Has(name string) bool {
	name = strings.ReplaceAll(name, "\\", "/")
	for _, f := range z.zr.File {
		if f.Name == name {
			return true
		}
	}
	return false
}
