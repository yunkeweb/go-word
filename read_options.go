package word

import (
	"archive/zip"
	"errors"
	"fmt"
	"math"

	"github.com/yunkeweb/go-word/pkg/common"
)

// ErrInvalidReadOptions reports an invalid negative read budget.
var ErrInvalidReadOptions = errors.New("word: invalid read options")

// ErrReadLimitExceeded reports that a configured ZIP read budget was exceeded.
// Use errors.Is to recognize it through contextual errors.
var ErrReadLimitExceeded = common.ErrZipLimitExceeded

// ReadOptions controls optional resource limits while reading a package.
// All limits are per call. Zero preserves the historical unlimited behavior.
type ReadOptions struct {
	MaxArchiveSize int64 // compressed package bytes
	MaxPartSize    int64 // one uncompressed ZIP member
	MaxTotalSize   int64 // sum of uncompressed ZIP members
	MaxEntries     int   // number of ZIP members
}

func (o ReadOptions) validate() error {
	if o.MaxArchiveSize < 0 || o.MaxPartSize < 0 || o.MaxTotalSize < 0 || o.MaxEntries < 0 {
		return fmt.Errorf("%w: limits must not be negative", ErrInvalidReadOptions)
	}
	return nil
}

func validateZipLimits(files []*zip.File, archiveSize int64, opts ReadOptions) error {
	if opts == (ReadOptions{}) {
		return nil
	}
	if opts.MaxArchiveSize > 0 && archiveSize >= 0 && archiveSize > opts.MaxArchiveSize {
		return fmt.Errorf("%w: MaxArchiveSize", ErrReadLimitExceeded)
	}
	if opts.MaxEntries > 0 && len(files) > opts.MaxEntries {
		return fmt.Errorf("%w: MaxEntries", ErrReadLimitExceeded)
	}
	var total uint64
	for _, f := range files {
		if opts.MaxPartSize > 0 && f.UncompressedSize64 > uint64(opts.MaxPartSize) {
			return fmt.Errorf("%w: MaxPartSize for %q", ErrReadLimitExceeded, f.Name)
		}
		if math.MaxUint64-total < f.UncompressedSize64 {
			return fmt.Errorf("%w: total size overflow", ErrReadLimitExceeded)
		}
		total += f.UncompressedSize64
	}
	if opts.MaxTotalSize > 0 && total > uint64(opts.MaxTotalSize) {
		return fmt.Errorf("%w: MaxTotalSize", ErrReadLimitExceeded)
	}
	return nil
}

func (o ReadOptions) partLimit() int64 {
	limit := o.MaxPartSize
	if o.MaxTotalSize > 0 && (limit == 0 || o.MaxTotalSize < limit) {
		limit = o.MaxTotalSize
	}
	return limit
}
