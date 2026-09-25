package common

import (
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"strings"
	"unicode/utf16"
)

const defaultSpinCount = 100000

// Algorithm identifiers from ECMA-376 write-protection (PHP PasswordEncoder).
const (
	AlgorithmMD2        = "MD2"
	AlgorithmMD4        = "MD4"
	AlgorithmMD5        = "MD5"
	AlgorithmSHA1       = "SHA-1"
	AlgorithmSHA256     = "SHA-256"
	AlgorithmSHA384     = "SHA-384"
	AlgorithmSHA512     = "SHA-512"
	AlgorithmRIPEMD     = "RIPEMD"
	AlgorithmRIPEMD160  = "RIPEMD-160"
	AlgorithmMAC        = "MAC"
	AlgorithmHMAC       = "HMAC"
)

// HashPassword implements the Office document-protection algorithm used by
// PHPOffice Common\Microsoft\PasswordEncoder (SHA1, 100000 spins).
func HashPassword(password string, algorithm string, salt []byte, spinCount int) (string, []byte, int, error) {
	if algorithm == "" {
		algorithm = "SHA-1"
	}
	if spinCount <= 0 {
		spinCount = defaultSpinCount
	}
	if salt == nil {
		salt = make([]byte, 16)
		if _, err := rand.Read(salt); err != nil {
			return "", nil, 0, err
		}
	}
	encoded := utf16le(password)
	h := sha1.New()
	h.Write(salt)
	h.Write(encoded)
	hash := h.Sum(nil)
	buf := make([]byte, 4)
	for i := 0; i < spinCount; i++ {
		binary.LittleEndian.PutUint32(buf, uint32(i))
		h.Reset()
		h.Write(hash)
		h.Write(buf)
		hash = h.Sum(nil)
	}
	return base64.StdEncoding.EncodeToString(hash), salt, spinCount, nil
}

// HashPasswordHex is a test helper that returns the SHA1 digest as hex.
func HashPasswordHex(password string, salt []byte, spinCount int) string {
	sum, _, _, err := HashPassword(password, "SHA-1", salt, spinCount)
	if err != nil {
		return ""
	}
	raw, err := base64.StdEncoding.DecodeString(sum)
	if err != nil {
		return ""
	}
	return strings.ToUpper(hex.EncodeToString(raw))
}

// AlgorithmID returns the OOXML cryptAlgorithmSid for a named algorithm.
func AlgorithmID(algorithm string) int {
	switch algorithm {
	case AlgorithmMD2:
		return 1
	case AlgorithmMD4:
		return 2
	case AlgorithmMD5:
		return 3
	case AlgorithmSHA1, "SHA1", "sha1":
		return 4
	case AlgorithmMAC:
		return 5
	case AlgorithmRIPEMD:
		return 6
	case AlgorithmRIPEMD160:
		return 7
	case AlgorithmHMAC:
		return 9
	case AlgorithmSHA256:
		return 12
	case AlgorithmSHA384:
		return 13
	case AlgorithmSHA512:
		return 14
	default:
		return 0
	}
}

func utf16le(s string) []byte {
	u := utf16.Encode([]rune(s))
	out := make([]byte, len(u)*2)
	for i, v := range u {
		binary.LittleEndian.PutUint16(out[i*2:], v)
	}
	return out
}
