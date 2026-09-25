package common

import (
	"strconv"
	"strings"
	"unicode/utf8"
)

// ControlCharEncode replaces XML-illegal C0 control characters with the
// OpenXML _xHHHH_ escape used by PHPOffice Common\Text.
func ControlCharEncode(s string) string {
	if s == "" {
		return s
	}
	var b strings.Builder
	for _, r := range s {
		if r < 0x20 && r != '\t' && r != '\n' && r != '\r' {
			b.WriteString("_x")
			b.WriteString(strings.ToUpper(itohex4(int(r))))
			b.WriteByte('_')
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

// ControlCharDecode reverses ControlCharEncode.
func ControlCharDecode(s string) string {
	if !strings.Contains(s, "_x") {
		return s
	}
	var b strings.Builder
	for i := 0; i < len(s); {
		if i+7 <= len(s) && s[i] == '_' && s[i+1] == 'x' && s[i+6] == '_' {
			if v, ok := parseHex4(s[i+2 : i+6]); ok {
				b.WriteRune(rune(v))
				i += 7
				continue
			}
		}
		r, size := utf8.DecodeRuneInString(s[i:])
		b.WriteRune(r)
		i += size
	}
	return b.String()
}

// ToUTF8 is a no-op in Go (strings are already UTF-8). Kept for PHP parity.
func ToUTF8(s string) string { return s }

// NumberFormat formats n with a fixed number of decimals using '.' (PHP Text::numberFormat).
func NumberFormat(n float64, decimals int) string {
	return strconv.FormatFloat(n, 'f', decimals, 64)
}

// Chr encodes Unicode code point dec as UTF-8 (PHP Text::chr).
func Chr(dec int) string {
	if dec < 0 {
		return ""
	}
	return string(rune(dec))
}

// IsUTF8 reports whether s is valid UTF-8.
func IsUTF8(s string) bool { return utf8.ValidString(s) }

// UTF8ToUnicode returns the Unicode code points of s.
func UTF8ToUnicode(s string) []int {
	out := make([]int, 0, len(s))
	for _, r := range s {
		out = append(out, int(r))
	}
	return out
}

// ToUnicode converts UTF-8 text to the RTF-style Unicode entities used by PHP Text::toUnicode.
func ToUnicode(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r == 0xFEFF {
			continue
		}
		if r > 127 {
			b.WriteString(`\uc0{\u`)
			b.WriteString(strconv.Itoa(int(r)))
			b.WriteByte('}')
			continue
		}
		b.WriteByte(byte(r))
	}
	return b.String()
}

// RemoveUnderscorePrefix strips a leading underscore (PHP Text::removeUnderscorePrefix).
func RemoveUnderscorePrefix(s string) string {
	return strings.TrimPrefix(s, "_")
}

func itohex4(v int) string {
	const hexdigits = "0123456789ABCDEF"
	return string([]byte{
		hexdigits[(v>>12)&0xF],
		hexdigits[(v>>8)&0xF],
		hexdigits[(v>>4)&0xF],
		hexdigits[v&0xF],
	})
}

func parseHex4(s string) (int, bool) {
	if len(s) != 4 {
		return 0, false
	}
	n := 0
	for i := 0; i < 4; i++ {
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
