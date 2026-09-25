package style

import "sync"

// Pools for short-lived style values. Callers must Put after the value is
// no longer referenced; zeroing on Get/Put keeps reused structs clean.

var (
	fontPool = sync.Pool{New: func() any { return new(Font) }}
	paraPool = sync.Pool{New: func() any { return new(Paragraph) }}
	tblPool  = sync.Pool{New: func() any { return new(Table) }}
)

// GetFont returns a zero Font from the pool.
func GetFont() *Font {
	f := fontPool.Get().(*Font)
	*f = Font{}
	return f
}

// PutFont returns f to the pool.
func PutFont(f *Font) {
	if f == nil {
		return
	}
	*f = Font{}
	fontPool.Put(f)
}

// GetParagraph returns a zero Paragraph from the pool.
func GetParagraph() *Paragraph {
	p := paraPool.Get().(*Paragraph)
	*p = Paragraph{}
	return p
}

// PutParagraph returns p to the pool.
func PutParagraph(p *Paragraph) {
	if p == nil {
		return
	}
	*p = Paragraph{}
	paraPool.Put(p)
}

// GetTable returns a zero table style from the pool.
func GetTable() *Table {
	t := tblPool.Get().(*Table)
	*t = Table{}
	return t
}

// PutTable returns t to the pool.
func PutTable(t *Table) {
	if t == nil {
		return
	}
	*t = Table{}
	tblPool.Put(t)
}
