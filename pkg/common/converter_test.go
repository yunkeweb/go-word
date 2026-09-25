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

func TestAllUnitConversions(t *testing.T) {
	if CMToInch(2.54) != 1 {
		t.Fatalf("CMToInch=%v", CMToInch(2.54))
	}
	if CMToPixel(2.54) != PixelPerInch {
		t.Fatalf("CMToPixel=%v", CMToPixel(2.54))
	}
	if CMToPoint(2.54) != PointPerInch {
		t.Fatalf("CMToPoint=%v", CMToPoint(2.54))
	}
	if CMToEMU(2.54) != int64(PixelPerInch*EMUPerPixel) {
		t.Fatalf("CMToEMU=%d", CMToEMU(2.54))
	}
	if InchToTwip(1) != TwipPerInch {
		t.Fatalf("InchToTwip=%v", InchToTwip(1))
	}
	if InchToCM(1) != CMPerInch {
		t.Fatalf("InchToCM=%v", InchToCM(1))
	}
	if InchToPixel(1) != PixelPerInch {
		t.Fatalf("InchToPixel=%v", InchToPixel(1))
	}
	if InchToPoint(1) != PointPerInch {
		t.Fatalf("InchToPoint=%v", InchToPoint(1))
	}
	if InchToEMU(1) != int64(PixelPerInch*EMUPerPixel) {
		t.Fatalf("InchToEMU=%d", InchToEMU(1))
	}
	if PixelToTwip(PixelPerInch) != TwipPerInch {
		t.Fatalf("PixelToTwip=%v", PixelToTwip(PixelPerInch))
	}
	if PixelToCM(PixelPerInch) != CMPerInch {
		t.Fatalf("PixelToCM=%v", PixelToCM(PixelPerInch))
	}
	if PixelToPoint(PixelPerInch) != PointPerInch {
		t.Fatalf("PixelToPoint=%v", PixelToPoint(PixelPerInch))
	}
	if PointToTwip(PointPerInch) != TwipPerInch {
		t.Fatalf("PointToTwip=%v", PointToTwip(PointPerInch))
	}
	if PointToCM(PointPerInch) != CMPerInch {
		t.Fatalf("PointToCM=%v", PointToCM(PointPerInch))
	}
	if PointToInch(PointPerInch) != 1 {
		t.Fatalf("PointToInch=%v", PointToInch(PointPerInch))
	}
	if PointToPixel(PointPerInch) != PixelPerInch {
		t.Fatalf("PointToPixel=%v", PointToPixel(PointPerInch))
	}
	if PointToEMU(PointPerInch) != PixelToEMU(PixelPerInch) {
		t.Fatalf("PointToEMU=%d", PointToEMU(PointPerInch))
	}
	if TwipToCM(TwipPerInch) != CMPerInch {
		t.Fatalf("TwipToCM=%v", TwipToCM(TwipPerInch))
	}
	if TwipToInch(TwipPerInch) != 1 {
		t.Fatalf("TwipToInch=%v", TwipToInch(TwipPerInch))
	}
	if TwipToPixel(TwipPerInch) != PixelPerInch {
		t.Fatalf("TwipToPixel=%v", TwipToPixel(TwipPerInch))
	}
	if TwipToPoint(TwipPerInch) != PointPerInch {
		t.Fatalf("TwipToPoint=%v", TwipToPoint(TwipPerInch))
	}
	if EMUToPixel(9525) != 1 {
		t.Fatalf("EMUToPixel=%v", EMUToPixel(9525))
	}
	if EMUToCM(InchToEMU(1)) != CMPerInch {
		t.Fatalf("EMUToCM=%v", EMUToCM(InchToEMU(1)))
	}
	if DegreeToAngleVal(1) != 60000 {
		t.Fatalf("DegreeToAngleVal=%d", DegreeToAngleVal(1))
	}
	if AngleToDegree(60000) != 1 {
		t.Fatalf("AngleToDegree=%v", AngleToDegree(60000))
	}
	if PicaToPoint(1) != 12 {
		t.Fatalf("PicaToPoint=%v", PicaToPoint(1))
	}
	if PicaToTwip(1) != PointToTwip(12) {
		t.Fatalf("PicaToTwip=%v", PicaToTwip(1))
	}
}

func TestStringToRGB(t *testing.T) {
	cases := map[string]string{
		"yellow": "FFFF00", "green": "90EE90", "lightGreen": "90EE90",
		"cyan": "00FFFF", "magenta": "FF00FF", "blue": "0000FF", "red": "FF0000",
		"darkBlue": "00008B", "darkCyan": "008B8B", "darkGreen": "006400",
		"darkMagenta": "8B008B", "darkRed": "8B0000", "darkYellow": "8B8B00",
		"darkGray": "A9A9A9", "lightGray": "D3D3D3", "black": "000000",
		"custom": "custom",
	}
	for in, want := range cases {
		if got := StringToRGB(in); got != want {
			t.Fatalf("StringToRGB(%q)=%q want %q", in, got, want)
		}
	}
}

func TestHtmlToRGBColor(t *testing.T) {
	if _, _, _, ok := HtmlToRGBColor(""); ok {
		t.Fatal("empty should fail")
	}
	r, g, b, ok := HtmlToRGBColor("red")
	if !ok || r != 0xff || g != 0 || b != 0 {
		t.Fatalf("named red = %d %d %d %v", r, g, b, ok)
	}
	r, g, b, ok = HtmlToRGBColor("#00FF00")
	if !ok || r != 0 || g != 0xff || b != 0 {
		t.Fatalf("hex = %d %d %d %v", r, g, b, ok)
	}
}

func TestCSSLengthVariants(t *testing.T) {
	if n, ok := CSSToPoint("0"); !ok || n != 0 {
		t.Fatalf("zero=%v %v", n, ok)
	}
	if n, ok := CSSToPoint("  10px "); !ok || n != PixelToPoint(10) {
		t.Fatalf("px=%v %v", n, ok)
	}
	if n, ok := CSSToPoint("10mm"); !ok || n != CMToPoint(1) {
		t.Fatalf("mm=%v %v", n, ok)
	}
	if n, ok := CSSToPoint("1in"); !ok || n != 72 {
		t.Fatalf("in=%v %v", n, ok)
	}
	if n, ok := CSSToPoint("2pc"); !ok || n != 24 {
		t.Fatalf("pc=%v %v", n, ok)
	}
	if _, ok := CSSToPoint("10zz"); ok {
		t.Fatal("unknown unit")
	}
	if _, ok := CSSToPoint("pt"); ok {
		t.Fatal("missing number")
	}
	if _, ok := CSSToPoint("10"); ok {
		t.Fatal("missing unit")
	}
	if _, ok := CSSToPoint(""); ok {
		t.Fatal("empty")
	}
	if n, ok := CSSToPoint("-2pt"); !ok || n != -2 {
		t.Fatalf("neg=%v %v", n, ok)
	}
	if n, ok := CSSToPoint("+3.5pt"); !ok || n != 3.5 {
		t.Fatalf("plus=%v %v", n, ok)
	}
	if n, ok := CSSToTwip("1pt"); !ok || n != 20 {
		t.Fatalf("CSSToTwip=%v %v", n, ok)
	}
	if _, ok := CSSToTwip("bad"); ok {
		t.Fatal("CSSToTwip bad")
	}
	if n, ok := CSSToPixel("72pt"); !ok || n != 96 {
		t.Fatalf("CSSToPixel=%v %v", n, ok)
	}
	if _, ok := CSSToPixel("bad"); ok {
		t.Fatal("CSSToPixel bad")
	}
	if n, ok := CSSToCM("72pt"); !ok || n != CMPerInch {
		t.Fatalf("CSSToCM=%v %v", n, ok)
	}
	if _, ok := CSSToCM("bad"); ok {
		t.Fatal("CSSToCM bad")
	}
	if n, ok := CSSToEMU("72pt"); !ok || n != PointToEMU(72) {
		t.Fatalf("CSSToEMU=%d %v", n, ok)
	}
	if _, ok := CSSToEMU("bad"); ok {
		t.Fatal("CSSToEMU bad")
	}
}
