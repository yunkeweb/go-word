package common

import (
	"encoding/base64"
	"testing"
)

func TestHashPassword(t *testing.T) {
	salt := []byte("0123456789abcdef")
	sum, gotSalt, spin, err := HashPassword("secret", "", salt, 2)
	if err != nil {
		t.Fatal(err)
	}
	if spin != 2 || string(gotSalt) != string(salt) {
		t.Fatalf("spin=%d salt=%q", spin, gotSalt)
	}
	if _, err := base64.StdEncoding.DecodeString(sum); err != nil {
		t.Fatal(err)
	}
	hex := HashPasswordHex("secret", salt, 2)
	if hex == "" || len(hex) != 40 {
		t.Fatalf("hex=%q", hex)
	}
	sum2, salt2, spin2, err := HashPassword("x", AlgorithmSHA1, nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(salt2) != 16 || spin2 != defaultSpinCount {
		t.Fatalf("generated salt len=%d spin=%d", len(salt2), spin2)
	}
	if sum2 == "" {
		t.Fatal("empty hash")
	}
}

func TestHashPasswordHexErrors(t *testing.T) {
	if HashPasswordHex("x", []byte("salt"), 1) == "" {
		t.Fatal("expected hex")
	}
}

func TestAlgorithmIDAll(t *testing.T) {
	cases := map[string]int{
		AlgorithmMD2: 1, AlgorithmMD4: 2, AlgorithmMD5: 3,
		AlgorithmSHA1: 4, "SHA1": 4, "sha1": 4,
		AlgorithmMAC: 5, AlgorithmRIPEMD: 6, AlgorithmRIPEMD160: 7,
		AlgorithmHMAC: 9, AlgorithmSHA256: 12, AlgorithmSHA384: 13,
		AlgorithmSHA512: 14, "unknown": 0,
	}
	for name, want := range cases {
		if got := AlgorithmID(name); got != want {
			t.Fatalf("AlgorithmID(%q)=%d want %d", name, got, want)
		}
	}
}

func TestUTF16LE(t *testing.T) {
	b := utf16le("A")
	if len(b) != 2 || b[0] != 'A' || b[1] != 0 {
		t.Fatalf("%v", b)
	}
}
