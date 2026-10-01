package word

import (
	"sync"
	"testing"
)

func TestNewWithOptionsIsolatesDefaults(t *testing.T) {
	a := NewWithOptions(DocumentOptions{DefaultFontName: "Aptos", DefaultAsianFontName: "Noto Sans CJK", DefaultFontSize: 12, DefaultFontColor: "112233"})
	b := NewWithOptions(DocumentOptions{DefaultFontName: "Calibri", DefaultFontSize: 9})
	if a.GetDefaultFontName() != "Aptos" || a.GetDefaultAsianFontName() != "Noto Sans CJK" || a.GetDefaultFontSize() != 12 || a.GetDefaultFontColor() != "112233" {
		t.Fatalf("options not applied: %+v", a)
	}
	if b.GetDefaultFontName() != "Calibri" || b.GetDefaultFontSize() != 9 {
		t.Fatalf("second options not applied: %+v", b)
	}
	if New().GetDefaultFontName() == "Aptos" || New().GetDefaultFontSize() == 12 {
		t.Fatal("instance options leaked into New")
	}
}

func TestNewWithOptionsConcurrentIsolation(t *testing.T) {
	var wg sync.WaitGroup
	for i := 0; i < 32; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			name := "font-" + string(rune('A'+i))
			d := NewWithOptions(DocumentOptions{DefaultFontName: name, DefaultFontSize: float64(i + 1)})
			if d.GetDefaultFontName() != name || d.GetDefaultFontSize() != float64(i+1) {
				t.Errorf("goroutine %d got %+v", i, d)
			}
		}(i)
	}
	wg.Wait()
}
