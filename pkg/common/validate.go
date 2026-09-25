package common

// CSS white-space values accepted by PHPWord Shared\Validate.
var cssWhiteSpace = map[string]struct{}{
	"pre-wrap": {},
	"normal":   {},
	"nowrap":   {},
	"pre":      {},
	"pre-line": {},
	"initial":  {},
	"inherit":  {},
}

// CSS generic font families accepted by PHPWord Shared\Validate.
var cssGenericFont = map[string]struct{}{
	"serif":      {},
	"sans-serif": {},
	"monospace":  {},
	"cursive":    {},
	"fantasy":    {},
	"system-ui":  {},
	"math":       {},
	"emoji":      {},
	"fangsong":   {},
}

// ValidateCSSWhiteSpace returns value if it is a known CSS white-space, else "".
func ValidateCSSWhiteSpace(value string) string {
	if _, ok := cssWhiteSpace[value]; ok {
		return value
	}
	return ""
}

// ValidateCSSGenericFont returns value if it is a generic CSS font family, else "".
func ValidateCSSGenericFont(value string) string {
	if _, ok := cssGenericFont[value]; ok {
		return value
	}
	return ""
}
