package word

import (
	"testing"

	"github.com/yunkeweb/go-word/element"
	"github.com/yunkeweb/go-word/style"
)

func TestProcessSettings(t *testing.T) {
	origFont := DefaultFontName()
	origAsian := DefaultAsianFontName()
	origSize := DefaultFontSize()
	origColor := DefaultFontColor()
	origPaper := DefaultPaper()
	origUnit := MeasurementUnit()
	origCompat := HasCompatibility()
	origEsc := IsOutputEscapingEnabled()
	origZip := ZipClass()
	origTemp := TempDir()
	origRtl := IsDefaultRtl()
	defer func() {
		SetDefaultFontName(origFont)
		SetDefaultAsianFontName(origAsian)
		SetDefaultFontSize(origSize)
		SetDefaultFontColor(origColor)
		SetDefaultPaper(origPaper)
		SetMeasurementUnit(origUnit)
		SetCompatibility(origCompat)
		SetOutputEscapingEnabled(origEsc)
		SetZipClass(origZip)
		SetTempDir(origTemp)
		SetDefaultRtl(origRtl)
	}()

	SetDefaultFontName("Calibri")
	if DefaultFontName() != "Calibri" {
		t.Fatal("font")
	}
	SetDefaultAsianFontName("SimSun")
	if DefaultAsianFontName() != "SimSun" {
		t.Fatal("asian")
	}
	SetDefaultFontSize(12)
	if DefaultFontSize() != 12 {
		t.Fatal("size")
	}
	SetDefaultFontColor("111111")
	if DefaultFontColor() != "111111" {
		t.Fatal("color")
	}
	SetOutputEscapingEnabled(false)
	if IsOutputEscapingEnabled() {
		t.Fatal("esc")
	}
	SetOutputEscapingEnabled(true)
	v := true
	SetDefaultRtl(&v)
	if IsDefaultRtl() == nil || !*IsDefaultRtl() {
		t.Fatal("rtl")
	}
	SetTempDir("tmp")
	if TempDir() != "tmp" {
		t.Fatal("temp")
	}
	SetZipClass("ZipArchive")
	if ZipClass() != "ZipArchive" {
		t.Fatal("zip")
	}
}

func TestDocumentDefaultsAndSections(t *testing.T) {
	doc := New()
	if doc.DocInfo() == nil || doc.GetDocInfo() == nil {
		t.Fatal("info")
	}
	if doc.Settings() == nil || doc.GetSettings() == nil {
		t.Fatal("settings")
	}
	if doc.Compatibility() == nil || doc.GetCompatibility() == nil {
		t.Fatal("compat")
	}
	doc.SetDefaultFontName("X")
	doc.SetDefaultAsianFontName("Y")
	doc.SetDefaultFontSize(14)
	doc.SetDefaultFontColor("abc")
	if doc.GetDefaultFontName() != "X" || doc.GetDefaultAsianFontName() != "Y" {
		t.Fatal("font names")
	}
	if doc.GetDefaultFontSize() != 14 || doc.GetDefaultFontColor() != "abc" {
		t.Fatal("size/color")
	}
	if doc.GetSection(0) != nil || doc.GetSection(-1) != nil {
		t.Fatal("empty section")
	}
	s1 := doc.AddSection()
	s2 := doc.AddSection(style.Section{Orientation: style.OrientationLandscape})
	_ = s1
	if doc.GetSection(0) == nil || doc.GetSection(1) != s2 || doc.GetSection(9) != nil {
		t.Fatal("get section")
	}
	if len(doc.GetSections()) != 2 || len(doc.Sections()) != 2 {
		t.Fatal("sections")
	}
	doc.sections[0].SectionID = 2
	doc.sections[1].SectionID = 1
	doc.SortSections(func(a, b *element.Section) bool { return a.SectionID < b.SectionID })
	if doc.sections[0].SectionID != 1 {
		t.Fatal("sort")
	}
	doc.GetSection(0).AddEndnote()
	doc.GetSection(0).AddComment("A", "A", "")
	if len(doc.GetEndnotes()) != 1 || len(doc.GetComments()) != 1 {
		t.Fatal("notes/comments")
	}
}
