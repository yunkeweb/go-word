package common

import "testing"

func TestCMToTwip(t *testing.T) {
	got := CMToTwip(2.54)
	if got != 1440 {
		t.Fatalf("CMToTwip(2.54)=%v want 1440", got)
	}
}

func TestPixelToEMU(t *testing.T) {
	if PixelToEMU(1) != 9525 {
		t.Fatalf("PixelToEMU(1)=%d want 9525", PixelToEMU(1))
	}
}

func TestHtmlToRGB(t *testing.T) {
	r, g, b, ok := HtmlToRGB("#1a2b3c")
	if !ok || r != 0x1a || g != 0x2b || b != 0x3c {
		t.Fatalf("HtmlToRGB #1a2b3c = %d %d %d %v", r, g, b, ok)
	}
	r, g, b, ok = HtmlToRGB("#abc")
	if !ok || r != 0xaa || g != 0xbb || b != 0xcc {
		t.Fatalf("HtmlToRGB #abc = %d %d %d %v", r, g, b, ok)
	}
}

func TestCSSToPoint(t *testing.T) {
	pt, ok := CSSToPoint("12pt")
	if !ok || pt != 12 {
		t.Fatalf("CSSToPoint(12pt)=%v %v", pt, ok)
	}
	pt, ok = CSSToPoint("2.54cm")
	if !ok || pt != 72 {
		t.Fatalf("CSSToPoint(2.54cm)=%v %v", pt, ok)
	}
}

func TestNumberFormatAndUnicode(t *testing.T) {
	if NumberFormat(1.2, 2) != "1.20" {
		t.Fatalf("NumberFormat=%q", NumberFormat(1.2, 2))
	}
	if Chr(0x4e2d) != "中" {
		t.Fatalf("Chr=%q", Chr(0x4e2d))
	}
	if RemoveUnderscorePrefix("_x") != "x" {
		t.Fatal("RemoveUnderscorePrefix")
	}
	if !IsUTF8("ok") {
		t.Fatal("IsUTF8")
	}
}

func TestAlgorithmID(t *testing.T) {
	if AlgorithmID(AlgorithmSHA1) != 4 {
		t.Fatalf("SHA-1 id=%d", AlgorithmID(AlgorithmSHA1))
	}
	if AlgorithmID(AlgorithmSHA256) != 12 {
		t.Fatalf("SHA-256 id=%d", AlgorithmID(AlgorithmSHA256))
	}
}

func TestValidateCSS(t *testing.T) {
	if ValidateCSSWhiteSpace("pre-wrap") != "pre-wrap" {
		t.Fatal("white-space")
	}
	if ValidateCSSGenericFont("serif") != "serif" {
		t.Fatal("generic font")
	}
	if ValidateCSSWhiteSpace("nope") != "" {
		t.Fatal("invalid white-space")
	}
}

func TestControlCharRoundTrip(t *testing.T) {
	in := "a\x08b"
	enc := ControlCharEncode(in)
	if enc != "a_x0008_b" {
		t.Fatalf("encode=%q", enc)
	}
	if ControlCharDecode(enc) != in {
		t.Fatalf("decode=%q", ControlCharDecode(enc))
	}
}

func TestXMLWriterEscape(t *testing.T) {
	w := NewXMLWriter()
	w.StartDocument()
	w.Start("w:t")
	w.Text(`a&b<c>"d`)
	w.End()
	s := w.String()
	if !containsAll(s, "a&amp;b", "&lt;c&gt;", "&#34;d") && !containsAll(s, "a&amp;b", "&lt;c&gt;", "&quot;d") {
		t.Fatalf("escape failed: %s", s)
	}
}

func containsAll(s string, parts ...string) bool {
	for _, p := range parts {
		if !contains(s, p) {
			return false
		}
	}
	return true
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 || indexOf(s, sub) >= 0)
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
