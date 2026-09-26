package word

import (
	"github.com/yunkeweb/go-word/element"
	"github.com/yunkeweb/go-word/metadata"
	"github.com/yunkeweb/go-word/pkg/common"
	"github.com/yunkeweb/go-word/style"
)

// Document protection modes (OOXML ST_DocProtect).
const (
	ProtectTypeNone           = style.DocProtectNone
	ProtectTypeReadOnly       = style.DocProtectReadOnly
	ProtectTypeComments       = style.DocProtectComments
	ProtectTypeTrackedChanges = style.DocProtectTrackedChanges
	ProtectTypeForms          = style.DocProtectForms
)

// Protect enables document protection (w:documentProtection). An empty password
// writes an unenforced-hash protection node; a non-empty password uses the
// Office SHA-1 / 100000-spin algorithm (ECMA-376).
func (d *Document) Protect(editing, password string) error {
	if editing == "" {
		editing = ProtectTypeReadOnly
	}
	p := &metadata.Protection{
		Editing:   editing,
		Password:  password,
		Algorithm: common.AlgorithmSHA1,
		SpinCount: 100000,
	}
	if password != "" {
		hash, salt, spin, err := common.HashPassword(password, p.Algorithm, nil, p.SpinCount)
		if err != nil {
			return err
		}
		p.Hash = hash
		p.Salt = salt
		p.SpinCount = spin
	}
	d.settings.DocumentProtection = p
	return nil
}

// SetEvenAndOddHeaders toggles odd/even page headers and footers
// (w:evenAndOddHeaders plus HeaderEven / FooterEven parts).
func (d *Document) SetEvenAndOddHeaders(enable bool) {
	d.settings.EvenAndOddHeaders = enable
	if !enable {
		return
	}
	sec := d.lastOrNewSection()
	if !sec.HasDifferentEvenPage() {
		sec.AddHeader(element.HeaderEven)
	}
}

// SetDifferentFirstPage enables a first-page header/footer on the current
// section (w:titlePg plus HeaderFirst / FooterFirst).
func (d *Document) SetDifferentFirstPage(enable bool) {
	if !enable {
		return
	}
	sec := d.lastOrNewSection()
	if !sec.HasDifferentFirstPage() {
		sec.AddHeader(element.HeaderFirst)
	}
}

// AddTableOfContents inserts an OpenXML TOC field covering Heading 1–3 by
// default (TOC \o "1-3" \h \z \u). Optional depth is max, or min,max.
func (d *Document) AddTableOfContents(depth ...int) *element.TOC {
	min, max := 1, 3
	if len(depth) == 1 && depth[0] > 0 {
		max = depth[0]
	}
	if len(depth) >= 2 {
		if depth[0] > 0 {
			min = depth[0]
		}
		if depth[1] > 0 {
			max = depth[1]
		}
	}
	if max < min {
		max = min
	}
	return d.lastOrNewSection().AddTOC(nil, nil, min, max)
}

// AddPageNumber inserts a PAGE field into the current section.
func (d *Document) AddPageNumber() *element.Field {
	return d.lastOrNewSection().AddPageNumber()
}

// AddNumPages inserts a NUMPAGES field into the current section.
func (d *Document) AddNumPages() *element.Field {
	return d.lastOrNewSection().AddNumPages()
}

func (d *Document) applyWatermarks() {
	if d.textWatermark == "" && len(d.imageWatermark) == 0 {
		return
	}
	if len(d.sections) == 0 {
		d.AddSection()
	}
	for _, sec := range d.sections {
		d.applySectionWatermarks(sec)
	}
}

func (d *Document) applySectionWatermarks(sec *element.Section) {
	headers := sec.GetHeaders()
	if len(headers) == 0 {
		headers = []*element.Header{sec.AddHeader()}
	}
	for _, h := range headers {
		if d.textWatermark != "" {
			tw := h.EnsureTextWatermark(d.textWatermark)
			applyTextWatermarkOptions(tw, d.textWatermarkOpts)
		}
		if len(d.imageWatermark) > 0 {
			img := h.EnsureImageWatermark(d.imageWatermark, d.imageWatermarkName)
			applyImageWatermarkOptions(img, d.imageWatermarkOpts)
		}
	}
}

func (d *Document) syncEvenAndOddHeaders() {
	if d.settings.EvenAndOddHeaders {
		return
	}
	for _, sec := range d.sections {
		if sec.HasDifferentEvenPage() {
			d.settings.EvenAndOddHeaders = true
			return
		}
	}
}
