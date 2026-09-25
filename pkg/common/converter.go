package common

// Unit conversion constants matching PHPWord Shared\Converter
// and PHPOffice Common\Drawing.
const (
	CMPerInch      = 2.54
	TwipPerInch    = 1440.0
	PixelPerInch   = 96.0
	PointPerInch   = 72.0
	PicaPerInch    = 6.0
	EMUPerPixel    = 9525.0
	AnglePerDegree = 60000.0
	DPI96          = 96.0
)

func CMToTwip(cm float64) float64  { return cm / CMPerInch * TwipPerInch }
func CMToInch(cm float64) float64  { return cm / CMPerInch }
func CMToPixel(cm float64) float64 { return cm / CMPerInch * PixelPerInch }
func CMToPoint(cm float64) float64 { return cm / CMPerInch * PointPerInch }
func CMToEMU(cm float64) int64     { return int64(CMToPixel(cm) * EMUPerPixel) }

func InchToTwip(inch float64) float64  { return inch * TwipPerInch }
func InchToCM(inch float64) float64    { return inch * CMPerInch }
func InchToPixel(inch float64) float64 { return inch * PixelPerInch }
func InchToPoint(inch float64) float64 { return inch * PointPerInch }
func InchToEMU(inch float64) int64     { return int64(InchToPixel(inch) * EMUPerPixel) }

func PixelToTwip(px float64) float64  { return px / PixelPerInch * TwipPerInch }
func PixelToCM(px float64) float64    { return px / PixelPerInch * CMPerInch }
func PixelToPoint(px float64) float64 { return px / PixelPerInch * PointPerInch }
func PixelToEMU(px float64) int64     { return int64(px * EMUPerPixel) }

func PointToTwip(pt float64) float64  { return pt / PointPerInch * TwipPerInch }
func PointToCM(pt float64) float64    { return pt / PointPerInch * CMPerInch }
func PointToInch(pt float64) float64  { return pt / PointPerInch }
func PointToPixel(pt float64) float64 { return pt / PointPerInch * PixelPerInch }
func PointToEMU(pt float64) int64     { return PixelToEMU(PointToPixel(pt)) }

func TwipToCM(twip float64) float64    { return twip / TwipPerInch * CMPerInch }
func TwipToInch(twip float64) float64  { return twip / TwipPerInch }
func TwipToPixel(twip float64) float64 { return twip / TwipPerInch * PixelPerInch }
func TwipToPoint(twip float64) float64 { return twip / TwipPerInch * PointPerInch }

func EMUToPixel(emu int64) float64 { return float64(emu) / EMUPerPixel }
func EMUToCM(emu int64) float64    { return PixelToCM(EMUToPixel(emu)) }

func DegreeToAngleVal(deg float64) int64 { return int64(deg * AnglePerDegree) }
func AngleToDegree(angle int64) float64  { return float64(angle) / AnglePerDegree }

func PicaToPoint(pica float64) float64 { return pica * 12 }
func PicaToTwip(pica float64) float64  { return PointToTwip(PicaToPoint(pica)) }

// StringToRGB maps a PHPWord highlight color name to RRGGBB (PHP Converter::stringToRgb).
func StringToRGB(value string) string {
	switch value {
	case "yellow":
		return "FFFF00"
	case "green", "lightGreen":
		return "90EE90"
	case "cyan":
		return "00FFFF"
	case "magenta":
		return "FF00FF"
	case "blue":
		return "0000FF"
	case "red":
		return "FF0000"
	case "darkBlue":
		return "00008B"
	case "darkCyan":
		return "008B8B"
	case "darkGreen":
		return "006400"
	case "darkMagenta":
		return "8B008B"
	case "darkRed":
		return "8B0000"
	case "darkYellow":
		return "8B8B00"
	case "darkGray":
		return "A9A9A9"
	case "lightGray":
		return "D3D3D3"
	case "black":
		return "000000"
	default:
		return value
	}
}

// HtmlToRGBColor parses a CSS/HTML color into RRGGBB components.
func HtmlToRGBColor(value string) (r, g, b int, ok bool) {
	if value == "" {
		return 0, 0, 0, false
	}
	if value[0] != '#' {
		value = StringToRGB(value)
	}
	return HtmlToRGB(value)
}

// CSSToPoint converts a CSS length (10px, 2cm, 12pt, ...) to points.
func CSSToPoint(value string) (float64, bool) {
	value = trimSpace(value)
	if value == "0" {
		return 0, true
	}
	n, unit, ok := splitCSSLength(value)
	if !ok {
		return 0, false
	}
	switch unit {
	case "pt":
		return n, true
	case "px":
		return PixelToPoint(n), true
	case "cm":
		return CMToPoint(n), true
	case "mm":
		return CMToPoint(n / 10), true
	case "in":
		return InchToPoint(n), true
	case "pc":
		return PicaToPoint(n), true
	default:
		return 0, false
	}
}

func CSSToTwip(value string) (float64, bool) {
	pt, ok := CSSToPoint(value)
	if !ok {
		return 0, false
	}
	return PointToTwip(pt), true
}

func CSSToPixel(value string) (float64, bool) {
	pt, ok := CSSToPoint(value)
	if !ok {
		return 0, false
	}
	return PointToPixel(pt), true
}

func CSSToCM(value string) (float64, bool) {
	pt, ok := CSSToPoint(value)
	if !ok {
		return 0, false
	}
	return PointToCM(pt), true
}

func CSSToEMU(value string) (int64, bool) {
	pt, ok := CSSToPoint(value)
	if !ok {
		return 0, false
	}
	return PointToEMU(pt), true
}

func trimSpace(s string) string {
	i, j := 0, len(s)
	for i < j && (s[i] == ' ' || s[i] == '\t' || s[i] == '\n' || s[i] == '\r') {
		i++
	}
	for j > i && (s[j-1] == ' ' || s[j-1] == '\t' || s[j-1] == '\n' || s[j-1] == '\r') {
		j--
	}
	return s[i:j]
}

func splitCSSLength(value string) (float64, string, bool) {
	i := 0
	if i < len(value) && (value[i] == '+' || value[i] == '-') {
		i++
	}
	dot := false
	start := i
	for i < len(value) {
		c := value[i]
		if c >= '0' && c <= '9' {
			i++
			continue
		}
		if c == '.' && !dot {
			dot = true
			i++
			continue
		}
		break
	}
	if i == start {
		return 0, "", false
	}
	n := 0.0
	frac := 0.0
	fracDiv := 1.0
	seenDot := false
	sign := 1.0
	p := 0
	if value[0] == '-' {
		sign = -1
		p = 1
	} else if value[0] == '+' {
		p = 1
	}
	for ; p < i; p++ {
		c := value[p]
		if c == '.' {
			seenDot = true
			continue
		}
		d := float64(c - '0')
		if seenDot {
			fracDiv *= 10
			frac += d / fracDiv
		} else {
			n = n*10 + d
		}
	}
	unit := value[i:]
	if unit == "" {
		return 0, "", false
	}
	return sign * (n + frac), unit, true
}
