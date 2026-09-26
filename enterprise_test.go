package word

import (
	"encoding/base64"
	"strings"
	"testing"

	"github.com/yunkeweb/go-word/element"
	"github.com/yunkeweb/go-word/ooxml"
	"github.com/yunkeweb/go-word/style"
)

func TestTextWatermarkAndProtection(t *testing.T) {
	doc := New()
	doc.SetTextWatermark("CONFIDENTIAL")
	if err := doc.Protect(ProtectTypeReadOnly, "secret"); err != nil {
		t.Fatal(err)
	}
	sec := doc.AddSection()
	sec.AddText("protected body")

	raw, err := doc.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	settings := readZipFile(t, raw, "word/settings.xml")
	assertWellFormedXML(t, "settings.xml", settings)
	if !strings.Contains(settings, `w:documentProtection`) {
		t.Fatal("missing documentProtection")
	}
	if !strings.Contains(settings, `w:edit="readOnly"`) {
		t.Fatal("edit=readOnly")
	}
	if !strings.Contains(settings, `w:enforcement="1"`) {
		t.Fatal("enforcement")
	}
	if !strings.Contains(settings, `w:cryptProviderType="rsaFull"`) {
		t.Fatal("cryptProviderType")
	}
	if !strings.Contains(settings, `w:cryptAlgorithmSid="4"`) {
		t.Fatal("SHA-1 sid")
	}
	if !strings.Contains(settings, `w:cryptSpinCount="100000"`) {
		t.Fatal("spin count")
	}
	if !strings.Contains(settings, `w:hash="`) || !strings.Contains(settings, `w:salt="`) {
		t.Fatal("hash/salt")
	}
	p := doc.Settings().DocumentProtection
	if p == nil || p.Hash == "" || len(p.Salt) != 16 {
		t.Fatal("protection hash state")
	}
	if _, err := base64.StdEncoding.DecodeString(p.Hash); err != nil {
		t.Fatal(err)
	}

	hdr := readZipFile(t, raw, "word/header1.xml")
	assertWellFormedXML(t, "header1.xml", hdr)
	if !strings.Contains(hdr, `xmlns:v="`+ooxml.NSV+`"`) {
		t.Fatal("header vml ns")
	}
	if !strings.Contains(hdr, `xmlns:w10="`+ooxml.NSW10+`"`) {
		t.Fatal("header w10 ns")
	}
	if !strings.Contains(hdr, `PowerPlusWaterMarkObject`) {
		t.Fatal("text watermark shape")
	}
	if !strings.Contains(hdr, `string="CONFIDENTIAL"`) {
		t.Fatal("watermark text")
	}
	if !strings.Contains(hdr, `v:textpath`) {
		t.Fatal("v:textpath")
	}
}

func TestImageWatermarkHeaderRels(t *testing.T) {
	doc := New()
	doc.SetImageWatermark(pngBytes(t))
	doc.AddSection().AddText("body")

	raw, err := doc.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	hdr := readZipFile(t, raw, "word/header1.xml")
	assertWellFormedXML(t, "header1.xml", hdr)
	if !strings.Contains(hdr, `v:imagedata`) {
		t.Fatal("image watermark vml")
	}
	if !strings.Contains(hdr, `WordPictureWatermark`) {
		t.Fatal("image watermark shape id")
	}
	if !strings.Contains(hdr, `r:id="rId1"`) {
		t.Fatal("header-local rId")
	}
	rels := readZipFile(t, raw, "word/_rels/header1.xml.rels")
	assertWellFormedXML(t, "header1.xml.rels", rels)
	if !strings.Contains(rels, ooxml.NSOfficeRelImage) {
		t.Fatal("header image rel type")
	}
	if !strings.Contains(rels, `Target="media/image1.png"`) && !strings.Contains(rels, `Target="media/image1.jpeg"`) {
		t.Fatalf("header image target: %s", rels)
	}
	ct := readZipFile(t, raw, "[Content_Types].xml")
	if !strings.Contains(ct, `PartName="/word/header1.xml"`) {
		t.Fatal("header content type")
	}
}

func TestProtectWithoutPassword(t *testing.T) {
	doc := New()
	if err := doc.Protect("", ""); err != nil {
		t.Fatal(err)
	}
	raw, err := doc.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	settings := readZipFile(t, raw, "word/settings.xml")
	if !strings.Contains(settings, `w:edit="readOnly"`) {
		t.Fatal("default readOnly")
	}
	if strings.Contains(settings, `w:cryptProviderType`) {
		t.Fatal("crypto attrs without password")
	}
}

func TestMixedOrientationAndHeaderModes(t *testing.T) {
	doc := New()
	portrait := doc.AddSection()
	portrait.AddHeader().AddText("odd")
	portrait.AddHeader(element.HeaderFirst).AddText("first")
	portrait.AddHeader(element.HeaderEven).AddText("even")
	portrait.AddFooter().AddPageNumber()
	portrait.AddFooter(element.HeaderFirst).AddText("first footer")
	portrait.AddText("portrait")

	land := doc.AddSection(style.Section{
		Orientation: style.OrientationLandscape,
		BreakType:   "nextPage",
	})
	land.AddText("landscape")
	land.AddFooter().AddNumPages()

	doc.SetEvenAndOddHeaders(true)

	raw, err := doc.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	document := readZipFile(t, raw, "word/document.xml")
	assertWellFormedXML(t, "document.xml", document)
	if !strings.Contains(document, `w:orient="landscape"`) {
		t.Fatal("landscape orient")
	}
	if !strings.Contains(document, `w:w="16838"`) || !strings.Contains(document, `w:h="11906"`) {
		t.Fatal("A4 landscape swapped dimensions")
	}
	if !strings.Contains(document, `w:titlePg`) {
		t.Fatal("titlePg")
	}
	all := zipText(t, raw)
	if !strings.Contains(all, `w:instrText`) || !strings.Contains(all, " PAGE ") {
		t.Fatal("PAGE field")
	}
	if !strings.Contains(all, " NUMPAGES ") {
		t.Fatal("NUMPAGES field")
	}

	sectKids := innerChildNames(document, "w:sectPr")
	assertSeq(t, sectKids,
		"w:headerReference", "w:footerReference", "w:type", "w:pgSz", "w:pgMar",
		"w:pgNumType", "w:cols", "w:titlePg", "w:docGrid")

	settings := readZipFile(t, raw, "word/settings.xml")
	if !strings.Contains(settings, `w:evenAndOddHeaders`) {
		t.Fatal("evenAndOddHeaders")
	}
}

func TestTableOfContentsFieldsAndBookmarks(t *testing.T) {
	doc := New()
	doc.AddTitleStyle(1, style.Font{Bold: true, Size: 16})
	doc.AddTitleStyle(2, style.Font{Bold: true, Size: 14})
	doc.AddTitleStyle(3, style.Font{Bold: true, Size: 12})
	sec := doc.AddSection()
	toc := doc.AddTableOfContents()
	if toc.MinDepth != 1 || toc.MaxDepth != 3 {
		t.Fatalf("toc depth %d-%d", toc.MinDepth, toc.MaxDepth)
	}
	sec.AddTitle("Chapter One", 1)
	sec.AddTitle("Section 1.1", 2)
	sec.AddTitle("Detail", 3)
	sec.AddTitle("Appendix", 4)

	raw, err := doc.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	document := readZipFile(t, raw, "word/document.xml")
	assertWellFormedXML(t, "document.xml", document)
	if !strings.Contains(document, `TOC \o`) || !strings.Contains(document, `1-3`) || !strings.Contains(document, `\h \z \u`) {
		t.Fatal("TOC instruction")
	}
	if !strings.Contains(document, `w:instrText`) {
		t.Fatal("instrText")
	}
	if !strings.Contains(document, `w:fldSimple`) {
		t.Fatal("fldSimple")
	}
	if !strings.Contains(document, `w:fldChar`) {
		t.Fatal("fldChar")
	}
	if !strings.Contains(document, `w:name="_Toc1"`) {
		t.Fatal("heading bookmark")
	}
	if !strings.Contains(document, `w:anchor="_Toc1"`) {
		t.Fatal("toc hyperlink")
	}
	if !strings.Contains(document, `PAGEREF _Toc1`) {
		t.Fatal("PAGEREF")
	}
	if !strings.Contains(document, "Chapter One") || !strings.Contains(document, "Section 1.1") {
		t.Fatal("toc preview entries")
	}
	if strings.Contains(document, `w:anchor="_Toc4"`) {
		t.Fatal("H4 should be excluded from default TOC")
	}
}

func TestWatermarkHelpersAndPageFields(t *testing.T) {
	doc := New()
	sec := doc.AddSection()
	sec.SetOrientation(style.OrientationLandscape)
	sec.SetDifferentFirstPage(true)
	if !sec.HasDifferentFirstPage() {
		t.Fatal("first page")
	}
	h := sec.AddHeader()
	h.AddTextWatermark("DRAFT")
	h.EnsureTextWatermark("DRAFT")
	img := h.AddWatermarkBytes("mark.png", pngBytes(t))
	if !img.IsWatermark {
		t.Fatal("image watermark flag")
	}
	h.EnsureImageWatermark(pngBytes(t))
	if !sec.HasDifferentEvenPage() {
		doc.SetEvenAndOddHeaders(true)
	}
	if doc.AddPageNumber() == nil || doc.AddNumPages() == nil {
		t.Fatal("page fields")
	}
	doc.SetDifferentFirstPage(true)
	raw, err := doc.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	if len(raw) < 1000 {
		t.Fatal("tiny package")
	}
}

func TestAddTableOfContentsDepthArgs(t *testing.T) {
	doc := New()
	if toc := doc.AddTableOfContents(2); toc.MaxDepth != 2 || toc.MinDepth != 1 {
		t.Fatalf("max-only %+v", toc)
	}
	doc2 := New()
	if toc := doc2.AddTableOfContents(2, 3); toc.MinDepth != 2 || toc.MaxDepth != 3 {
		t.Fatalf("min-max %+v", toc)
	}
}

func TestSetImageWatermarkEmptyClears(t *testing.T) {
	doc := New()
	doc.SetImageWatermark(pngBytes(t))
	doc.SetImageWatermark(nil)
	doc.AddSection().AddText("x")
	raw, err := doc.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	all := zipText(t, raw)
	if strings.Contains(all, "WordPictureWatermark") || strings.Contains(all, "v:imagedata") {
		t.Fatal("cleared image watermark still present")
	}
}
