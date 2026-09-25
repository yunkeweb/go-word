package common

import "testing"

func TestControlCharEdges(t *testing.T) {
	if ControlCharEncode("") != "" {
		t.Fatal("empty encode")
	}
	if ControlCharEncode("a\tb\nc\rd") != "a\tb\nc\rd" {
		t.Fatal("allowed controls")
	}
	if ControlCharDecode("plain") != "plain" {
		t.Fatal("no escape")
	}
	if ControlCharDecode("_xZZZZ_") != "_xZZZZ_" {
		t.Fatal("invalid hex kept")
	}
	if ControlCharDecode("_x000G_") != "_x000G_" {
		t.Fatal("bad nibble")
	}
	if _, ok := parseHex4("abc"); ok {
		t.Fatal("short hex4")
	}
	if _, ok := parseHex4("xxxx"); ok {
		t.Fatal("invalid hex4")
	}
	n, ok := parseHex4("00aF")
	if !ok || n != 0xAF {
		t.Fatalf("parseHex4=%d %v", n, ok)
	}
}

func TestTextHelpers(t *testing.T) {
	if ToUTF8("hi") != "hi" {
		t.Fatal("ToUTF8")
	}
	if Chr(-1) != "" {
		t.Fatal("Chr negative")
	}
	if IsUTF8(string([]byte{0xff, 0xfe})) {
		t.Fatal("invalid utf8")
	}
	u := UTF8ToUnicode("A中")
	if len(u) != 2 || u[0] != 'A' || u[1] != 0x4e2d {
		t.Fatalf("%v", u)
	}
	got := ToUnicode("A中\uFEFF")
	if got != `A\uc0{\u20013}` {
		t.Fatalf("ToUnicode=%q", got)
	}
	if RemoveUnderscorePrefix("x") != "x" {
		t.Fatal("no prefix")
	}
}
