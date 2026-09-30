package common

import (
	"archive/zip"
	"bytes"
	"errors"
	"io"
	"math"
	"strings"
	"testing"
)

func TestReadAllWithLimitBoundaries(t *testing.T) {
	for _, tc := range []struct {
		input    string
		limit    int64
		tooLarge bool
	}{
		{"", 1, false}, {"abc", 3, false}, {"abcd", 3, true},
		{"abc", 0, false}, {"abc", math.MaxInt64, false},
	} {
		data, err := ReadAllWithLimit(strings.NewReader(tc.input), tc.limit)
		if tc.tooLarge {
			if !errors.Is(err, ErrZipLimitExceeded) || data != nil {
				t.Errorf("got %q, %v", data, err)
			}
		} else if err != nil || string(data) != tc.input {
			t.Errorf("got %q, %v", data, err)
		}
	}
	if _, err := ReadAllWithLimit(strings.NewReader("abc"), -1); err == nil {
		t.Fatal("negative limit accepted")
	}
}

func TestCopyWithLimitAndStickyError(t *testing.T) {
	var dst bytes.Buffer
	n, err := CopyWithLimit(&dst, strings.NewReader("abcd"), 3)
	if n != 3 || dst.String() != "abc" || !errors.Is(err, ErrZipLimitExceeded) {
		t.Fatalf("copied %d, %q: %v", n, dst.String(), err)
	}
	r := limitedInput(strings.NewReader("ab"), 1)
	if _, err := io.ReadAll(r); !errors.Is(err, ErrZipLimitExceeded) {
		t.Fatal(err)
	}
	if _, err := r.Read(make([]byte, 1)); !errors.Is(err, ErrZipLimitExceeded) {
		t.Fatal("lost limit error", err)
	}
}

func TestZipEntryActualSizeCannotBypassBudget(t *testing.T) {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	// Deliberately understate the uncompressed size in the central directory.
	data := bytes.Repeat([]byte{1}, 1024)
	w, err := zw.CreateRaw(&zip.FileHeader{Name: "bad", Method: zip.Store, CompressedSize64: uint64(len(data)), UncompressedSize64: 1})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := w.Write(data); err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	zr, err := OpenZipBytes(buf.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	if data, err := zr.ReadFileWithLimit("bad", 16); err == nil || data != nil {
		t.Fatalf("forged size accepted: bytes=%d err=%v", len(data), err)
	}
}
