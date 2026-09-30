package common

import (
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
)

// ErrZipLimitExceeded reports that a ZIP package exceeds a caller-supplied limit.
var ErrZipLimitExceeded = errors.New("zip: configured size limit exceeded")

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
	return OpenZipFileWithLimit(filename, 0)
}

// OpenZipFileWithLimit rejects archives larger than maxBytes before parsing.
// Zero means unlimited; negative values are invalid.
func OpenZipFileWithLimit(filename string, maxBytes int64) (*ZipReader, error) {
	if maxBytes < 0 {
		return nil, fmt.Errorf("zip: negative archive limit")
	}
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	st, err := f.Stat()
	if err != nil {
		f.Close()
		return nil, err
	}
	if maxBytes > 0 && st.Size() > maxBytes {
		f.Close()
		return nil, ErrZipLimitExceeded
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

// ReadFile returns the contents of name, or os.ErrNotExist if missing.
func (z *ZipReader) ReadFile(name string) ([]byte, error) {
	return z.ReadFileWithLimit(name, 0)
}

// ReadFileWithLimit reads name and rejects entries larger than maxBytes.
// Zero means no limit; negative values are invalid.
func (z *ZipReader) ReadFileWithLimit(name string, maxBytes int64) ([]byte, error) {
	name = strings.ReplaceAll(name, "\\", "/")
	for _, f := range z.zr.File {
		if f.Name == name {
			rc, err := OpenZipEntry(f, maxBytes)
			if err != nil {
				return nil, err
			}
			defer rc.Close()
			data, err := io.ReadAll(rc)
			if err != nil {
				return nil, err
			}
			return data, nil
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
