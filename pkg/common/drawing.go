package common

// Drawing ports PHPOffice Common\Drawing unit helpers used by images
// and shapes. Values are rounded the same way as the PHP original.

func AngleToDegrees(value int) float64 {
	if value == 0 {
		return 0
	}
	return float64(int(float64(value)/60000 + 0.5))
}

func CentimetersToPixels(cm float64) int {
	if cm == 0 {
		return 0
	}
	return int(cm/2.54*DPI96 + 0.5)
}

func CentimetersToPoints(cm float64) float64 {
	if cm == 0 {
		return 0
	}
	return cm / 2.54 * DPI96 * 0.75
}

func CentimetersToTwips(cm float64) float64 {
	if cm == 0 {
		return 0
	}
	return cm / 2.54 * 1440
}

func CentimetersToEMU(cm float64) int64 {
	return int64(float64(CentimetersToPixels(float64(int(cm*100+0.5))/100)) * EMUPerPixel)
}

func InchesToPoints(inch float64) float64 {
	return inch * 72
}

func InchesToTwips(inch float64) float64 {
	return inch * 1440
}

func InchesToPixels(inch float64) int {
	return int(inch*DPI96 + 0.5)
}

func InchesToEMU(inch float64) int64 {
	return int64(float64(InchesToPixels(inch)) * EMUPerPixel)
}

func PixelsToEMU(px float64) int64 {
	return int64(px * EMUPerPixel)
}

func EMUToPixelsInt(emu int64) int {
	if emu == 0 {
		return 0
	}
	return int(float64(emu)/EMUPerPixel + 0.5)
}

func PixelsToPoints(px float64) float64 {
	return px * 0.75
}

func PixelsToCentimeters(px float64) float64 {
	return px / DPI96 * 2.54
}

func PointsToPixels(pt float64) int {
	return int(pt/0.75 + 0.5)
}

func PointsToTwips(pt float64) float64 {
	return pt * 20
}

func TwipsToPixels(twip float64) int {
	return int(twip/15 + 0.5) // 1440 twip/inch / 96 px/inch = 15
}

func DegreesToAngle(deg float64) int {
	return int(deg*60000 + 0.5)
}

// EMUToPixels is the PHP Drawing::emuToPixels name for EMUToPixelsInt.
func EMUToPixels(emu int64) int { return EMUToPixelsInt(emu) }

func PicasToPoints(pica float64) float64 { return pica * 12 }

func PointsToCentimeters(pt float64) float64 {
	if pt == 0 {
		return 0
	}
	return (pt / 0.75) / DPI96 * 2.54
}

func PointsToEMU(pt float64) int64 {
	if pt == 0 {
		return 0
	}
	return int64(pt/0.75*EMUPerPixel + 0.5)
}

func TwipsToCentimeters(twip float64) float64 {
	if twip == 0 {
		return 0
	}
	return twip / 566.928
}

func TwipsToInches(twip float64) float64 {
	if twip == 0 {
		return 0
	}
	return twip / 1440
}

// HtmlToRGB parses "#RGB" / "#RRGGBB" into 8-bit components (PHP Drawing::htmlToRGB).
func HtmlToRGB(value string) (r, g, b int, ok bool) {
	if value == "" {
		return 0, 0, 0, false
	}
	if value[0] == '#' {
		value = value[1:]
	}
	var rs, gs, bs string
	switch len(value) {
	case 6:
		rs, gs, bs = value[0:2], value[2:4], value[4:6]
	case 3:
		rs = string([]byte{value[0], value[0]})
		gs = string([]byte{value[1], value[1]})
		bs = string([]byte{value[2], value[2]})
	default:
		return 0, 0, 0, false
	}
	ri, ok1 := parseHexByte(rs)
	gi, ok2 := parseHexByte(gs)
	bi, ok3 := parseHexByte(bs)
	if !ok1 || !ok2 || !ok3 {
		return 0, 0, 0, false
	}
	return ri, gi, bi, true
}

func parseHexByte(s string) (int, bool) {
	if len(s) != 2 {
		return 0, false
	}
	n := 0
	for i := 0; i < 2; i++ {
		c := s[i]
		var d int
		switch {
		case c >= '0' && c <= '9':
			d = int(c - '0')
		case c >= 'A' && c <= 'F':
			d = int(c-'A') + 10
		case c >= 'a' && c <= 'f':
			d = int(c-'a') + 10
		default:
			return 0, false
		}
		n = n<<4 | d
	}
	return n, true
}
