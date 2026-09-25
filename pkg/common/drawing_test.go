package common

import "testing"

func TestDrawingZeroAndNonZero(t *testing.T) {
	if AngleToDegrees(0) != 0 {
		t.Fatal("AngleToDegrees 0")
	}
	if AngleToDegrees(60000) != 1 {
		t.Fatalf("AngleToDegrees=%v", AngleToDegrees(60000))
	}
	if CentimetersToPixels(0) != 0 {
		t.Fatal("CentimetersToPixels 0")
	}
	if CentimetersToPixels(2.54) != 96 {
		t.Fatalf("CentimetersToPixels=%d", CentimetersToPixels(2.54))
	}
	if CentimetersToPoints(0) != 0 {
		t.Fatal("CentimetersToPoints 0")
	}
	if CentimetersToPoints(2.54) != 72 {
		t.Fatalf("CentimetersToPoints=%v", CentimetersToPoints(2.54))
	}
	if CentimetersToTwips(0) != 0 {
		t.Fatal("CentimetersToTwips 0")
	}
	if CentimetersToTwips(2.54) != 1440 {
		t.Fatalf("CentimetersToTwips=%v", CentimetersToTwips(2.54))
	}
	if CentimetersToEMU(0) != 0 {
		t.Fatal("CentimetersToEMU 0")
	}
	if CentimetersToEMU(2.54) == 0 {
		t.Fatal("CentimetersToEMU nonzero")
	}
	if InchesToPoints(1) != 72 {
		t.Fatal("InchesToPoints")
	}
	if InchesToTwips(1) != 1440 {
		t.Fatal("InchesToTwips")
	}
	if InchesToPixels(1) != 96 {
		t.Fatal("InchesToPixels")
	}
	if InchesToEMU(1) != int64(96*EMUPerPixel) {
		t.Fatalf("InchesToEMU=%d", InchesToEMU(1))
	}
	if PixelsToEMU(1) != 9525 {
		t.Fatal("PixelsToEMU")
	}
	if EMUToPixelsInt(0) != 0 {
		t.Fatal("EMUToPixelsInt 0")
	}
	if EMUToPixelsInt(9525) != 1 {
		t.Fatalf("EMUToPixelsInt=%d", EMUToPixelsInt(9525))
	}
	if EMUToPixels(9525) != 1 {
		t.Fatal("EMUToPixels")
	}
	if PixelsToPoints(96) != 72 {
		t.Fatalf("PixelsToPoints=%v", PixelsToPoints(96))
	}
	if PixelsToCentimeters(96) != 2.54 {
		t.Fatalf("PixelsToCentimeters=%v", PixelsToCentimeters(96))
	}
	if PointsToPixels(72) != 96 {
		t.Fatalf("PointsToPixels=%d", PointsToPixels(72))
	}
	if PointsToTwips(1) != 20 {
		t.Fatal("PointsToTwips")
	}
	if TwipsToPixels(15) != 1 {
		t.Fatalf("TwipsToPixels=%d", TwipsToPixels(15))
	}
	if DegreesToAngle(1) != 60000 {
		t.Fatal("DegreesToAngle")
	}
	if PicasToPoints(1) != 12 {
		t.Fatal("PicasToPoints")
	}
	if PointsToCentimeters(0) != 0 {
		t.Fatal("PointsToCentimeters 0")
	}
	if PointsToCentimeters(72) != 2.54 {
		t.Fatalf("PointsToCentimeters=%v", PointsToCentimeters(72))
	}
	if PointsToEMU(0) != 0 {
		t.Fatal("PointsToEMU 0")
	}
	if PointsToEMU(0.75) != 9525 {
		t.Fatalf("PointsToEMU=%d", PointsToEMU(0.75))
	}
	if TwipsToCentimeters(0) != 0 {
		t.Fatal("TwipsToCentimeters 0")
	}
	if TwipsToCentimeters(566.928) != 1 {
		t.Fatalf("TwipsToCentimeters=%v", TwipsToCentimeters(566.928))
	}
	if TwipsToInches(0) != 0 {
		t.Fatal("TwipsToInches 0")
	}
	if TwipsToInches(1440) != 1 {
		t.Fatal("TwipsToInches")
	}
}

func TestHtmlToRGBEdges(t *testing.T) {
	if _, _, _, ok := HtmlToRGB(""); ok {
		t.Fatal("empty")
	}
	if _, _, _, ok := HtmlToRGB("#12"); ok {
		t.Fatal("short")
	}
	if _, _, _, ok := HtmlToRGB("gg0000"); ok {
		t.Fatal("bad hex")
	}
	r, g, b, ok := HtmlToRGB("ABCDEF")
	if !ok || r != 0xAB || g != 0xCD || b != 0xEF {
		t.Fatalf("upper=%d %d %d %v", r, g, b, ok)
	}
	if _, ok := parseHexByte("1"); ok {
		t.Fatal("parseHexByte short")
	}
	if _, ok := parseHexByte("GG"); ok {
		t.Fatal("parseHexByte invalid")
	}
	if _, ok := parseHexByte("0g"); ok {
		t.Fatal("parseHexByte invalid second")
	}
}
