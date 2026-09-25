package common

// Font ports PHPOffice Common\Font approximate size conversions.

func FontSizeToPixels(points float64) float64 {
	return (16.0 / 12.0) * points
}

func InchSizeToPixels(inch float64) int {
	return int(inch * 96)
}

func CentimeterSizeToPixels(cm float64) float64 {
	return cm * 37.795275591
}

func CentimeterSizeToTwips(cm float64) float64 {
	return cm / 2.54 * 1440
}

func InchSizeToTwips(inch float64) float64 {
	return inch * 1440
}

func PixelSizeToTwips(px float64) float64 {
	return px / 96 * 1440
}

func PointSizeToTwips(pt float64) float64 {
	return pt * 20
}
