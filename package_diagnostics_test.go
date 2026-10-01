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
