package common

import "testing"

func TestValidateCSSAll(t *testing.T) {
	ws := []string{"pre-wrap", "normal", "nowrap", "pre", "pre-line", "initial", "inherit"}
	for _, v := range ws {
		if ValidateCSSWhiteSpace(v) != v {
			t.Fatalf("white-space %q", v)
		}
	}
	fonts := []string{"serif", "sans-serif", "monospace", "cursive", "fantasy", "system-ui", "math", "emoji", "fangsong"}
	for _, v := range fonts {
		if ValidateCSSGenericFont(v) != v {
			t.Fatalf("font %q", v)
		}
	}
	if ValidateCSSGenericFont("Comic Sans") != "" {
		t.Fatal("invalid font")
	}
}
