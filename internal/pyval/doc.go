// Port of CPython 3.12 value semantics: Objects/unicodeobject.c, Objects/bytesobject.c,
// Objects/floatobject.c, Objects/longobject.c, Objects/complexobject.c, Python/pystrtod.c,
// Python/dtoa.c and Modules/binascii.c (PSF License Version 2; Copyright (c) 2001-2024 Python
// Software Foundation).

// Package pyval implements the Python 3.12 value semantics the PyYAML port needs: repr() of str,
// bytes, int, float and complex; int(), float() and complex() of a str; str.lower(); code point
// access to strings; the utf-8 and utf-16 codecs with their error positions and reasons; and
// base64 decoding and encoding as binascii does it.
//
// A Python str is a Go string in generalized UTF-8: a lone surrogate U+D800..U+DFFF, which a str
// can hold, is its three-byte encoding. Go's own conversions turn those three bytes into U+FFFD,
// so strings go through Runes and FromRunes here.
//
// Python exceptions are returned as *Error values carrying the exception class name and
// str(exc). No exported function panics.
package pyval
