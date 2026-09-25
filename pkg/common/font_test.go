package common

import "testing"

func TestFontSizeHelpers(t *testing.T) {
	if FontSizeToPixels(12) != 16 {
		t.Fatalf("FontSizeToPixels=%v", FontSizeToPixels(12))
	}
	if InchSizeToPixels(1) != 96 {
		t.Fatalf("InchSizeToPixels=%d", InchSizeToPixels(1))
	}
	if CentimeterSizeToPixels(1) < 37 || CentimeterSizeToPixels(1) > 38 {
		t.Fatalf("CentimeterSizeToPixels=%v", CentimeterSizeToPixels(1))
	}
	if CentimeterSizeToTwips(2.54) != 1440 {
		t.Fatalf("CentimeterSizeToTwips=%v", CentimeterSizeToTwips(2.54))
	}
	if InchSizeToTwips(1) != 1440 {
		t.Fatal("InchSizeToTwips")
	}
	if PixelSizeToTwips(96) != 1440 {
		t.Fatalf("PixelSizeToTwips=%v", PixelSizeToTwips(96))
	}
	if PointSizeToTwips(1) != 20 {
		t.Fatal("PointSizeToTwips")
	}
}
