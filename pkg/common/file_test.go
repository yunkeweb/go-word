package common

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"
)

func TestFileExistsAndReadFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "a.txt")
	if err := os.WriteFile(path, []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}
	if !FileExists(path) {
		t.Fatal("exists")
	}
	if FileExists(filepath.Join(dir, "missing.txt")) {
		t.Fatal("missing")
	}
	b, err := ReadFile(path)
	if err != nil || string(b) != "hello" {
		t.Fatalf("ReadFile=%q %v", b, err)
	}
	if _, err := ReadFile(filepath.Join(dir, "missing.txt")); err == nil {
		t.Fatal("ReadFile missing")
	}
}

func TestFileZipProtocol(t *testing.T) {
	dir := t.TempDir()
	zipPath := filepath.Join(dir, "a.zip")
	f, err := os.Create(zipPath)
	if err != nil {
		t.Fatal(err)
	}
	zw := zip.NewWriter(f)
	w, err := zw.Create("inner/x.txt")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := w.Write([]byte("zipdata")); err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}

	uri := "zip://" + zipPath + "#inner/x.txt"
	if !FileExists(uri) {
		t.Fatal("zip exists")
	}
	if FileExists("zip://" + zipPath + "#nope") {
		t.Fatal("missing inner")
	}
	if FileExists("zip://" + zipPath) {
		t.Fatal("no hash")
	}
	if FileExists("zip://no-such.zip#x") {
		t.Fatal("missing archive")
	}
	b, err := ReadFile(uri)
	if err != nil || string(b) != "zipdata" {
		t.Fatalf("ReadFile zip=%q %v", b, err)
	}
	if _, err := ReadFile("zip://" + zipPath); err == nil {
		t.Fatal("ReadFile no hash")
	}
	if _, err := ReadFile("zip://no-such.zip#x"); err == nil {
		t.Fatal("ReadFile missing archive")
	}
	if _, err := ReadFile("zip://" + zipPath + "#nope"); err == nil {
		t.Fatal("ReadFile missing inner")
	}

	badZip := filepath.Join(dir, "bad.zip")
	if err := os.WriteFile(badZip, makeEncryptedZip(t, "bad.txt", []byte("hello")), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := ReadFile("zip://" + badZip + "#bad.txt"); err == nil {
		t.Fatal("expected zip member open error")
	}
}

func TestRealPath(t *testing.T) {
	dir := t.TempDir()
	p := RealPath(filepath.Join(dir, "x", "..", "y.txt"))
	if p == "" {
		t.Fatal("empty")
	}
	got := RealPath(filepath.Join(".", "no-such-file-xyz"))
	if got == "" {
		t.Fatal("fallback empty")
	}
	existing := filepath.Join(dir, "exists.txt")
	if err := os.WriteFile(existing, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if RealPath(existing) == "" {
		t.Fatal("existing empty")
	}
	_ = RealPath(string([]byte{0}))
}
