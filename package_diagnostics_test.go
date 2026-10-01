package word

import (
	"testing"
)

func TestValidatePackageCleanDocument(t *testing.T) {
	d := New()
	d.AddSection().AddText("ok")
	raw, err := d.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	for _, diag := range ValidatePackage(raw) {
		if diag.Severity == "error" {
			t.Fatalf("unexpected diagnostic: %+v", diag)
		}
	}
}

func TestValidatePackageFindsMalformedZipAndMissingDocument(t *testing.T) {
	if got := ValidatePackage([]byte("not zip")); len(got) != 1 || got[0].Code != "invalid_zip" {
		t.Fatalf("%+v", got)
	}
}

func TestValidatePackageWarnsForUnsupportedDrawings(t *testing.T) {
	d := New()
	d.AddSection().AddDMLShape("roundRect", 100, 50, "5B9BD5", "2E75B6", 12700)
	raw, err := d.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, diag := range ValidatePackage(raw) {
		if diag.Code == "unsupported_drawing" && diag.Severity == "warning" {
			found = true
		}
	}
	if !found {
		t.Fatal("expected unsupported drawing warning")
	}
}
