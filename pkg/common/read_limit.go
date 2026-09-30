package common

import (
	"archive/zip"
	"fmt"
	"io"
)

// ReadAllWithLimit rejects input larger than maxBytes, without returning partial
// data. Zero means unlimited; negative values are invalid.
func ReadAllWithLimit(r io.Reader, maxBytes int64) ([]byte, error) {
	if maxBytes < 0 {
		return nil, fmt.Errorf("zip: negative read limit")
	}
	if r == nil {
		return nil, io.ErrUnexpectedEOF
	}
	data, err := io.ReadAll(limitedInput(r, maxBytes))
	if err != nil {
		return nil, err
	}
	return data, nil
}

// CopyWithLimit copies at most maxBytes and probes one extra byte to distinguish
// an exact fit from an oversized input. Zero means unlimited.
func CopyWithLimit(dst io.Writer, src io.Reader, maxBytes int64) (int64, error) {
	if maxBytes < 0 {
		return 0, fmt.Errorf("zip: negative read limit")
	}
	if src == nil {
		return 0, io.ErrUnexpectedEOF
	}
	return io.Copy(dst, limitedInput(src, maxBytes))
}

// OpenZipEntry checks the declared size and bounds actual decompressed reads.
// It retains the underlying ZIP checksum and format errors.
func OpenZipEntry(f *zip.File, maxBytes int64) (io.ReadCloser, error) {
	if maxBytes < 0 {
		return nil, fmt.Errorf("zip: negative read limit")
	}
	if maxBytes > 0 && f.UncompressedSize64 > uint64(maxBytes) {
		return nil, fmt.Errorf("%w: part %q", ErrZipLimitExceeded, f.Name)
	}
	rc, err := f.Open()
	if err != nil {
		return nil, err
	}
	return &limitedReadCloser{Reader: limitedInput(rc, maxBytes), Closer: rc}, nil
}

type limitedReadCloser struct {
	io.Reader
	io.Closer
}

func limitedInput(r io.Reader, limit int64) io.Reader {
	if limit == 0 {
		return r
	}
	return &sizeLimitedReader{r: r, remaining: limit}
}

type sizeLimitedReader struct {
	r         io.Reader
	remaining int64
	err       error
}

func (r *sizeLimitedReader) Read(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	if r.err != nil {
		return 0, r.err
	}
	if r.remaining == 0 {
		var probe [1]byte
		n, err := r.r.Read(probe[:])
		if n > 0 {
			err = ErrZipLimitExceeded
		}
		r.err = err
		return 0, err
	}
	if int64(len(p)) > r.remaining {
		p = p[:r.remaining]
	}
	n, err := r.r.Read(p)
	r.remaining -= int64(n)
	r.err = err
	return n, err
}
