package pyuni

import (
	"crypto/sha256"
	"fmt"
	"os"
	"strings"
	"testing"
)

// TestCodePointDigest checks every code point's str.isprintable() and str.lower() against the digest
// CPython 3.12 produces for the same walk (testdata/codepoints.sha256, recorded from the interpreter).
// A single wrong property changes the digest.
func TestCodePointDigest(t *testing.T) {
	want, err := os.ReadFile("testdata/codepoints.sha256")
	if err != nil {
		t.Fatalf("reading the recorded digest: %v", err)
	}
	h := sha256.New()
	for cp := rune(0); cp <= MaxRune; cp++ {
		printable := 0
		if IsPrintable(cp) {
			printable = 1
		}
		lower := string(LowerRune(cp))
		if cp >= 0xD800 && cp <= 0xDFFF {
			// A Go string cannot hold a surrogate; CPython maps each one to itself.
			lower = string([]byte{0xED, byte(0x80 | (cp>>6)&0x3F), byte(0x80 | cp&0x3F)})
		}
		fmt.Fprintf(h, "%d:%d:%s;", cp, printable, lower)
	}
	if got := fmt.Sprintf("%x", h.Sum(nil)); got != strings.TrimSpace(string(want)) {
		t.Errorf("the code point digest is %s, want %s", got, strings.TrimSpace(string(want)))
	}
}

// TestFinalSigma checks the context-sensitive lowercase mapping of U+03A3.
func TestFinalSigma(t *testing.T) {
	for _, row := range [][2]string{
		{"Σ", "σ"},
		{"ΑΣ", "ας"},
		{"ΑΣΑ", "ασα"},
		{"ΑΣ'", "ας'"},
		{"ΣΑ", "σα"},
	} {
		if got := Lower(row[0]); got != row[1] {
			t.Errorf("Lower(%q) = %q, want %q", row[0], got, row[1])
		}
	}
}
