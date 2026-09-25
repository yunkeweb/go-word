package common

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// FileExists reports whether name exists on disk, or inside a zip when
// given the PHPOffice form zip://archive.zip#inner/path.
func FileExists(name string) bool {
	if strings.HasPrefix(strings.ToLower(name), "zip://") {
		rest := name[6:]
		hash := strings.IndexByte(rest, '#')
		if hash < 0 {
			return false
		}
		archive, inner := rest[:hash], rest[hash+1:]
		r, err := zip.OpenReader(archive)
		if err != nil {
			return false
		}
		defer r.Close()
		for _, f := range r.File {
			if f.Name == inner {
				return true
			}
		}
		return false
	}
	_, err := os.Stat(name)
	return err == nil
}

// ReadFile reads a regular file or a zip:// member.
func ReadFile(name string) ([]byte, error) {
	if strings.HasPrefix(strings.ToLower(name), "zip://") {
		rest := name[6:]
		hash := strings.IndexByte(rest, '#')
		if hash < 0 {
			return nil, os.ErrNotExist
		}
		archive, inner := rest[:hash], rest[hash+1:]
		r, err := zip.OpenReader(archive)
		if err != nil {
			return nil, err
		}
		defer r.Close()
		for _, f := range r.File {
			if f.Name != inner {
				continue
			}
			rc, err := f.Open()
			if err != nil {
				return nil, err
			}
			defer rc.Close()
			return io.ReadAll(rc)
		}
		return nil, os.ErrNotExist
	}
	return os.ReadFile(name)
}

// RealPath returns the canonical path, falling back to cleaning ".." segments
// when the file does not yet exist (PHP File::realpath).
func RealPath(name string) string {
	if abs, err := filepath.Abs(name); err == nil {
		if p, err := filepath.EvalSymlinks(abs); err == nil {
			return p
		}
		return abs
	}
	return filepath.Clean(name)
}
