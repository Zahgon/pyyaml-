// Port of PyYAML's lib/yaml/reader.py (MIT License; Copyright (c) 2017-2021 Ingy döt Net, Copyright
// (c) 2006-2016 Kirill Simonov; see LICENSE).

package yaml

import (
	"bytes"
	"fmt"

	"github.com/yaml/pyyaml-go/internal/pyval"
)

// reader decodes a stream, checks its characters and tracks the position.
type reader struct {
	name     string
	buffer   []rune
	pointer  int
	encoding *string
	index    int
	line     int
	column   int
}

func newReader(stream any) *reader {
	r := &reader{}
	switch s := stream.(type) {
	case string:
		r.name = "<unicode string>"
		rs := pyval.Runes(s)
		r.checkPrintable(rs)
		r.buffer = append(rs, 0)
	case []byte:
		r.name = "<byte string>"
		r.decode(s)
	default:
		raise(&PyError{Class: "AttributeError", Msg: fmt.Sprintf("'%s' object has no attribute 'read'", typeName(stream))})
	}
	return r
}

func (r *reader) decode(data []byte) {
	var rs []rune
	var derr *pyval.DecodeError
	switch {
	case bytes.HasPrefix(data, []byte{0xFF, 0xFE}):
		r.encoding = ptr("utf-16-le")
		rs, derr = pyval.DecodeUTF16(data, false)
	case bytes.HasPrefix(data, []byte{0xFE, 0xFF}):
		r.encoding = ptr("utf-16-be")
		rs, derr = pyval.DecodeUTF16(data, true)
	default:
		r.encoding = ptr("utf-8")
		rs, derr = pyval.DecodeUTF8(data)
	}
	if derr != nil {
		raise(&ReaderError{Name: r.name, Position: derr.Start, Character: int(data[derr.Start]),
			Encoding: derr.Encoding, Reason: derr.Reason})
	}
	r.checkPrintable(rs)
	r.buffer = append(rs, 0)
}

func isPrintable(c rune) bool {
	return c == 0x09 || c == 0x0A || c == 0x0D || (c >= 0x20 && c <= 0x7E) || c == 0x85 ||
		(c >= 0xA0 && c <= 0xD7FF) || (c >= 0xE000 && c <= 0xFFFD) || (c >= 0x10000 && c <= 0x10FFFF)
}

func (r *reader) checkPrintable(data []rune) {
	for i, c := range data {
		if !isPrintable(c) {
			raise(&ReaderError{Name: r.name, Position: i, Character: int(c), Encoding: "unicode",
				Reason: "special characters are not allowed"})
		}
	}
}

func (r *reader) peek(index int) rune {
	if r.pointer+index < len(r.buffer) {
		return r.buffer[r.pointer+index]
	}
	return 0
}

func (r *reader) prefix(length int) string {
	end := r.pointer + length
	if end > len(r.buffer) {
		end = len(r.buffer)
	}
	return pyval.FromRunes(r.buffer[r.pointer:end])
}

func (r *reader) forward(length int) {
	for ; length > 0; length-- {
		ch := r.buffer[r.pointer]
		r.pointer++
		r.index++
		if ch == '\n' || ch == 0x85 || ch == 0x2028 || ch == 0x2029 || (ch == '\r' && r.buffer[r.pointer] != '\n') {
			r.line++
			r.column = 0
		} else if ch != 0xFEFF {
			r.column++
		}
	}
}

func (r *reader) mark() *Mark {
	return &Mark{Name: r.name, Index: r.index, Line: r.line, Column: r.column, Buffer: r.buffer, Pointer: r.pointer}
}
