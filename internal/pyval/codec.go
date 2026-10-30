// Port of CPython 3.12's codecs and binascii: Objects/stringlib/codecs.h, Objects/unicodeobject.c
// and Modules/binascii.c (PSF License Version 2; Copyright (c) 2001-2024 Python Software
// Foundation).

package pyval

import (
	"fmt"
	"strings"
)

// DecodeError is a UnicodeDecodeError: Encoding, Reason, Start and End are its attributes (End is
// exclusive), and Data the object it was decoding.
type DecodeError struct {
	Encoding string
	Start    int
	End      int
	Reason   string
	Data     []byte
}

// Error is str(exc).
func (e *DecodeError) Error() string {
	if e.End == e.Start+1 {
		return fmt.Sprintf("'%s' codec can't decode byte 0x%02x in position %d: %s", e.Encoding, e.Data[e.Start], e.Start, e.Reason)
	}
	return fmt.Sprintf("'%s' codec can't decode bytes in position %d-%d: %s", e.Encoding, e.Start, e.End-1, e.Reason)
}

// DecodeUTF8 is codecs.utf_8_decode(data, 'strict', True).
func DecodeUTF8(data []byte) ([]rune, *DecodeError) {
	out := make([]rune, 0, len(data))
	fail := func(i, n int, reason string) ([]rune, *DecodeError) {
		return nil, &DecodeError{Encoding: "utf-8", Start: i, End: i + n, Reason: reason, Data: data}
	}
	cont := func(c byte) bool { return c&0xC0 == 0x80 }
	const end = "unexpected end of data"
	const bad = "invalid continuation byte"
	for i := 0; i < len(data); {
		c := data[i]
		left := len(data) - i
		switch {
		case c < 0x80:
			out = append(out, rune(c))
			i++
		case c < 0xC2 || c > 0xF4:
			return fail(i, 1, "invalid start byte")
		case c < 0xE0:
			if left < 2 {
				return fail(i, left, end)
			}
			if !cont(data[i+1]) {
				return fail(i, 1, bad)
			}
			out = append(out, rune(c&0x1F)<<6|rune(data[i+1]&0x3F))
			i += 2
		case c < 0xF0:
			if left < 2 {
				return fail(i, left, end)
			}
			d := data[i+1]
			if !cont(d) || (c == 0xE0 && d < 0xA0) || (c == 0xED && d >= 0xA0) {
				return fail(i, 1, bad)
			}
			if left < 3 {
				return fail(i, left, end)
			}
			if !cont(data[i+2]) {
				return fail(i, 2, bad)
			}
			out = append(out, rune(c&0x0F)<<12|rune(d&0x3F)<<6|rune(data[i+2]&0x3F))
			i += 3
		default:
			if left < 2 {
				return fail(i, left, end)
			}
			d := data[i+1]
			if !cont(d) || (c == 0xF0 && d < 0x90) || (c == 0xF4 && d >= 0x90) {
				return fail(i, 1, bad)
			}
			if left < 3 {
				return fail(i, left, end)
			}
			if !cont(data[i+2]) {
				return fail(i, 2, bad)
			}
			if left < 4 {
				return fail(i, left, end)
			}
			if !cont(data[i+3]) {
				return fail(i, 3, bad)
			}
			out = append(out, rune(c&0x07)<<18|rune(d&0x3F)<<12|rune(data[i+2]&0x3F)<<6|rune(data[i+3]&0x3F))
			i += 4
		}
	}
	return out, nil
}

// DecodeUTF16 is codecs.utf_16_le_decode or utf_16_be_decode(data, 'strict', True); a byte order
// mark stays in the result.
func DecodeUTF16(data []byte, bigEndian bool) ([]rune, *DecodeError) {
	encoding := "utf-16-le"
	if bigEndian {
		encoding = "utf-16-be"
	}
	unit := func(i int) rune {
		if bigEndian {
			return rune(data[i])<<8 | rune(data[i+1])
		}
		return rune(data[i+1])<<8 | rune(data[i])
	}
	fail := func(i int, reason string) ([]rune, *DecodeError) {
		return nil, &DecodeError{Encoding: encoding, Start: i, End: i + 2, Reason: reason, Data: data}
	}
	out := make([]rune, 0, len(data)/2)
	i := 0
	for i+1 < len(data) {
		u := unit(i)
		switch {
		case u < 0xD800 || u > 0xDFFF:
			out = append(out, u)
			i += 2
		case u >= 0xDC00:
			return fail(i, "illegal encoding")
		default:
			if i+3 >= len(data) {
				return fail(i, "unexpected end of data")
			}
			v := unit(i + 2)
			if v < 0xDC00 || v > 0xDFFF {
				return fail(i, "illegal UTF-16 surrogate")
			}
			out = append(out, 0x10000+(u-0xD800)<<10+(v-0xDC00))
			i += 4
		}
	}
	if i < len(data) {
		return fail(i, "truncated data")
	}
	return out, nil
}

// EncodeError is a UnicodeEncodeError; Msg is its str().
type EncodeError struct{ Msg string }

func (e *EncodeError) Error() string { return e.Msg }

func encodeFailure(encoding, reason string, rs []rune, start, end int) *EncodeError {
	if end == start+1 {
		var char string
		switch r := rs[start]; {
		case r <= 0xff:
			char = fmt.Sprintf(`\x%02x`, r)
		case r <= 0xffff:
			char = fmt.Sprintf(`\u%04x`, r)
		default:
			char = fmt.Sprintf(`\U%08x`, r)
		}
		return &EncodeError{Msg: fmt.Sprintf("'%s' codec can't encode character '%s' in position %d: %s",
			encoding, char, start, reason)}
	}
	return &EncodeError{Msg: fmt.Sprintf("'%s' codec can't encode characters in position %d-%d: %s",
		encoding, start, end-1, reason)}
}

// Encode is s.encode(encoding) for utf-8, utf-16, utf-16-le, utf-16-be, ascii and latin-1. Each
// call with utf-16 starts with its own byte order mark, little-endian as on the recording hosts.
func Encode(s, encoding string) ([]byte, *EncodeError) {
	rs := Runes(s)
	isSurrogate := func(r rune) bool { return r >= 0xD800 && r <= 0xDFFF }
	var out []byte
	switch encoding {
	case "utf-8":
		for i := 0; i < len(rs); i++ {
			if isSurrogate(rs[i]) {
				j := i + 1
				for j < len(rs) && isSurrogate(rs[j]) {
					j++
				}
				return nil, encodeFailure(encoding, "surrogates not allowed", rs, i, j)
			}
			out = AppendRune(out, rs[i])
		}
	case "utf-16", "utf-16-le", "utf-16-be":
		big := encoding == "utf-16-be"
		put := func(u rune) {
			if big {
				out = append(out, byte(u>>8), byte(u))
			} else {
				out = append(out, byte(u), byte(u>>8))
			}
		}
		if encoding == "utf-16" {
			put(0xFEFF)
		}
		for i := 0; i < len(rs); i++ {
			r := rs[i]
			switch {
			case isSurrogate(r):
				j := i + 1
				for j < len(rs) && isSurrogate(rs[j]) {
					j++
				}
				return nil, encodeFailure(encoding, "surrogates not allowed", rs, i, j)
			case r >= 0x10000:
				r -= 0x10000
				put(0xD800 + r>>10)
				put(0xDC00 + r&0x3FF)
			default:
				put(r)
			}
		}
	case "ascii", "latin-1":
		limit, reason := rune(0x80), "ordinal not in range(128)"
		if encoding == "latin-1" {
			limit, reason = 0x100, "ordinal not in range(256)"
		}
		for i := 0; i < len(rs); i++ {
			if rs[i] >= limit {
				j := i + 1
				for j < len(rs) && rs[j] >= limit {
					j++
				}
				return nil, encodeFailure(encoding, reason, rs, i, j)
			}
			out = append(out, byte(rs[i]))
		}
	default:
		return nil, &EncodeError{Msg: "unknown encoding: " + encoding}
	}
	return out, nil
}

const base64Alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"

// DecodeBase64 is binascii.a2b_base64(data), the non-strict mode base64.decodebytes uses. The
// error is binascii.Error's message.
func DecodeBase64(data []byte) ([]byte, string) {
	var out []byte
	quadPos, pads, dataChars := 0, 0, 0
	var left byte
	for _, c := range data {
		if c == '=' {
			if quadPos >= 2 {
				pads++
				if quadPos+pads >= 4 {
					return out, ""
				}
			}
			continue
		}
		v := strings.IndexByte(base64Alphabet, c)
		if v < 0 {
			continue
		}
		pads = 0
		switch quadPos {
		case 0:
			quadPos = 1
			left = byte(v)
		case 1:
			quadPos = 2
			out = append(out, left<<2|byte(v)>>4)
			left = byte(v) & 0x0f
		case 2:
			quadPos = 3
			out = append(out, left<<4|byte(v)>>2)
			left = byte(v) & 0x03
		case 3:
			quadPos = 0
			out = append(out, left<<6|byte(v))
			left = 0
		}
		dataChars++
	}
	if quadPos != 0 {
		if quadPos == 1 {
			return nil, fmt.Sprintf("Invalid base64-encoded string: number of data characters (%d) cannot be 1 more than a multiple of 4", dataChars)
		}
		return nil, "Incorrect padding"
	}
	return out, ""
}

// EncodeBase64Lines is base64.encodebytes(data): lines of 76 characters, each ended by a newline,
// and nothing at all for empty data.
func EncodeBase64Lines(data []byte) string {
	var b strings.Builder
	for i := 0; i < len(data); i += 57 {
		end := i + 57
		if end > len(data) {
			end = len(data)
		}
		chunk := data[i:end]
		for j := 0; j < len(chunk); j += 3 {
			switch len(chunk) - j {
			case 1:
				b.WriteByte(base64Alphabet[chunk[j]>>2])
				b.WriteByte(base64Alphabet[chunk[j]&3<<4])
				b.WriteString("==")
			case 2:
				b.WriteByte(base64Alphabet[chunk[j]>>2])
				b.WriteByte(base64Alphabet[chunk[j]&3<<4|chunk[j+1]>>4])
				b.WriteByte(base64Alphabet[chunk[j+1]&0xf<<2])
				b.WriteByte('=')
			default:
				b.WriteByte(base64Alphabet[chunk[j]>>2])
				b.WriteByte(base64Alphabet[chunk[j]&3<<4|chunk[j+1]>>4])
				b.WriteByte(base64Alphabet[chunk[j+1]&0xf<<2|chunk[j+2]>>6])
				b.WriteByte(base64Alphabet[chunk[j+2]&0x3f])
			}
		}
		b.WriteByte('\n')
	}
	return b.String()
}
