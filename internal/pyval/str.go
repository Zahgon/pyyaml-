// Port of CPython 3.12's str behaviour: Objects/unicodeobject.c unicode_repr and
// Objects/unicodectype.c through internal/pyuni (PSF License Version 2; Copyright (c) 2001-2024
// Python Software Foundation).

package pyval

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/yaml/pyyaml-go/internal/pyuni"
)

// Runes returns the code points of s, keeping lone surrogates. An invalid byte, which no str
// produced by this library contains, becomes U+FFFD.
func Runes(s string) []rune {
	out := make([]rune, 0, len(s))
	for i := 0; i < len(s); {
		r, size := DecodeRune(s[i:])
		out = append(out, r)
		i += size
	}
	return out
}

// DecodeRune decodes the first code point of s, accepting the three-byte form of a surrogate.
func DecodeRune(s string) (rune, int) {
	r, size := utf8.DecodeRuneInString(s)
	if r == utf8.RuneError && size == 1 && len(s) >= 3 && s[0] == 0xED && s[1] >= 0xA0 && s[1] <= 0xBF &&
		s[2] >= 0x80 && s[2] <= 0xBF {
		return 0xD000 | rune(s[1]&0x3F)<<6 | rune(s[2]&0x3F), 3
	}
	return r, size
}

// AppendRune appends r in generalized UTF-8.
func AppendRune(b []byte, r rune) []byte {
	if r >= 0xD800 && r <= 0xDFFF {
		return append(b, 0xED, byte(0x80|(r>>6)&0x3F), byte(0x80|r&0x3F))
	}
	return utf8.AppendRune(b, r)
}

// FromRunes is the inverse of Runes.
func FromRunes(rs []rune) string {
	b := make([]byte, 0, len(rs))
	for _, r := range rs {
		b = AppendRune(b, r)
	}
	return string(b)
}

// Len is len(s).
func Len(s string) int {
	n := 0
	for i := 0; i < len(s); n++ {
		_, size := DecodeRune(s[i:])
		i += size
	}
	return n
}

// StrRepr is repr(s): single quotes unless the string contains a single quote and no double
// quote; backslash, the chosen quote, \t, \n and \r escaped; every other character that is not
// printable in Unicode 15.0.0 escaped as \xhh, \uhhhh or \Uhhhhhhhh.
func StrRepr(s string) string {
	quote := byte('\'')
	if strings.IndexByte(s, '\'') >= 0 && strings.IndexByte(s, '"') < 0 {
		quote = '"'
	}
	var b strings.Builder
	b.Grow(len(s) + 2)
	b.WriteByte(quote)
	for i := 0; i < len(s); {
		r, size := DecodeRune(s[i:])
		i += size
		switch {
		case r == rune(quote) || r == '\\':
			b.WriteByte('\\')
			b.WriteRune(r)
		case r == '\t':
			b.WriteString(`\t`)
		case r == '\n':
			b.WriteString(`\n`)
		case r == '\r':
			b.WriteString(`\r`)
		case r < ' ' || r == 0x7f:
			fmt.Fprintf(&b, "\\x%02x", r)
		case r < 0x7f:
			b.WriteRune(r)
		case !pyuni.IsPrintable(r):
			switch {
			case r <= 0xff:
				fmt.Fprintf(&b, "\\x%02x", r)
			case r <= 0xffff:
				fmt.Fprintf(&b, "\\u%04x", r)
			default:
				fmt.Fprintf(&b, "\\U%08x", r)
			}
		default:
			b.WriteRune(r)
		}
	}
	b.WriteByte(quote)
	return b.String()
}

// BytesRepr is repr(b): single quotes unless b holds a single quote and no double quote; \t, \n,
// \r, backslash and the chosen quote escaped; other bytes outside 0x20..0x7E as \xhh.
func BytesRepr(data []byte) string {
	quote := byte('\'')
	if strings.IndexByte(string(data), '\'') >= 0 && strings.IndexByte(string(data), '"') < 0 {
		quote = '"'
	}
	var b strings.Builder
	b.Grow(len(data) + 3)
	b.WriteByte('b')
	b.WriteByte(quote)
	for _, c := range data {
		switch {
		case c == quote || c == '\\':
			b.WriteByte('\\')
			b.WriteByte(c)
		case c == '\t':
			b.WriteString(`\t`)
		case c == '\n':
			b.WriteString(`\n`)
		case c == '\r':
			b.WriteString(`\r`)
		case c < ' ' || c >= 0x7f:
			fmt.Fprintf(&b, "\\x%02x", c)
		default:
			b.WriteByte(c)
		}
	}
	b.WriteByte(quote)
	return b.String()
}

// Lower is str.lower(), keeping lone surrogates.
func Lower(s string) string { return FromRunes(pyuni.LowerRunes(Runes(s))) }
