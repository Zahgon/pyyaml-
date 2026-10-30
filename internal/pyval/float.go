// Port of CPython 3.12's float semantics: Objects/floatobject.c (float_repr, PyFloat_FromString)
// and Python/pystrtod.c / Python/dtoa.c (PSF License Version 2; Copyright (c) 2001-2024 Python
// Software Foundation).

package pyval

import (
	"math"
	"strconv"
	"strings"
)

// FloatRepr returns Python's repr(f) (equal to str(f)): the shortest string that round-trips,
// in positional notation when the decimal exponent is in [-4, 16) ("0.0001",
// "123456789012345.0", "1234567890123456.0") and in exponent notation otherwise ("1e-05",
// "1e+16", "1.2345678901234568e+16"); "-0.0", "inf", "-inf" and "nan" (the sign of a NaN is
// not shown).
func FloatRepr(f float64) string {
	switch {
	case math.IsNaN(f):
		return "nan"
	case math.IsInf(f, 1):
		return "inf"
	case math.IsInf(f, -1):
		return "-inf"
	}
	// Shortest round-trip digits; Go and David Gay's dtoa mode 0 both pick the digit string
	// closest to the exact value among the shortest ones.
	e := strconv.FormatFloat(f, 'e', -1, 64) // "-d.ddde±XX"
	neg := false
	if e[0] == '-' {
		neg = true
		e = e[1:]
	}
	epos := strings.IndexByte(e, 'e')
	mant, exps := e[:epos], e[epos+1:]
	exp, _ := strconv.Atoi(exps)
	digits := strings.Replace(mant, ".", "", 1)
	// Python's decpt: value = 0.DIGITS * 10**decpt.
	decpt := exp + 1
	if digits == "0" {
		decpt = 1
	}
	var b strings.Builder
	if neg {
		b.WriteByte('-')
	}
	if decpt <= -4 || decpt > 16 {
		b.WriteByte(digits[0])
		if len(digits) > 1 {
			b.WriteByte('.')
			b.WriteString(digits[1:])
		}
		x := decpt - 1
		b.WriteByte('e')
		if x < 0 {
			b.WriteByte('-')
			x = -x
		} else {
			b.WriteByte('+')
		}
		if x < 10 {
			b.WriteByte('0')
		}
		b.WriteString(strconv.Itoa(x))
		return b.String()
	}
	switch {
	case decpt <= 0:
		b.WriteString("0.")
		b.WriteString(strings.Repeat("0", -decpt))
		b.WriteString(digits)
	case decpt >= len(digits):
		b.WriteString(digits)
		b.WriteString(strings.Repeat("0", decpt-len(digits)))
		b.WriteString(".0")
	default:
		b.WriteString(digits[:decpt])
		b.WriteByte('.')
		b.WriteString(digits[decpt:])
	}
	return b.String()
}

// FloatStr returns Python's str(f), which equals repr(f).
func FloatStr(f float64) string { return FloatRepr(f) }

// isPySpaceASCII is Py_ISSPACE: the C-locale-free ASCII whitespace test used by the number
// parsers (it does not include \x1c-\x1f, unlike str.isspace).
func isPySpaceASCII(c byte) bool {
	switch c {
	case ' ', '\t', '\n', '\v', '\f', '\r':
		return true
	}
	return false
}

// ParseFloat is Python's float(s) for a str argument. It strips leading and trailing
// whitespace (Unicode whitespace, but of ASCII only " \t\n\v\f\r"), accepts Unicode decimal
// digits, single underscores between digits, an optional sign, "inf", "infinity" and "nan"
// in any case, and a decimal mantissa with an optional exponent. Overflow gives ±inf and
// underflow ±0.0, as in Python. Any other input gives ValueError
// "could not convert string to float: <repr(s)>".
func ParseFloat(s string) (float64, error) {
	fail := func() (float64, error) {
		return 0, valueError("could not convert string to float: " + StrRepr(s))
	}
	buf, ok := transformDecimalAndSpace(s)
	if !ok {
		return fail()
	}
	i, j := 0, len(buf)
	for i < j && isPySpaceASCII(buf[i]) {
		i++
	}
	if i == j {
		return fail()
	}
	for j-1 > i && isPySpaceASCII(buf[j-1]) {
		j--
	}
	t := buf[i:j]
	if strings.IndexByte(string(t), '_') >= 0 {
		var ok bool
		t, ok = removeUnderscores(t)
		if !ok {
			return fail()
		}
	}
	f, ok := parseFloatLiteral(string(t))
	if !ok {
		return fail()
	}
	return f, nil
}

// transformDecimalAndSpace is _PyUnicode_TransformDecimalAndSpaceToASCII: code points below
// 127 are kept, other whitespace becomes ' ', other decimal digits become their ASCII digit;
// any other code point makes the conversion fail (CPython writes '?', which no number parser
// accepts).
func transformDecimalAndSpace(s string) ([]byte, bool) {
	out := make([]byte, 0, len(s))
	for _, r := range s {
		switch {
		case r < 127:
			out = append(out, byte(r))
		case isSpaceRune(r):
			out = append(out, ' ')
		default:
			d, ok := decimalValue(r)
			if !ok {
				return out, false
			}
			out = append(out, byte('0'+d))
		}
	}
	return out, true
}

// removeUnderscores is _Py_string_to_number_with_underscores: an underscore is allowed only
// between two ASCII digits.
func removeUnderscores(s []byte) ([]byte, bool) {
	out := make([]byte, 0, len(s))
	var prev byte
	for _, c := range s {
		if c == '_' {
			if !(prev >= '0' && prev <= '9') {
				return nil, false
			}
		} else {
			out = append(out, c)
			if prev == '_' && !(c >= '0' && c <= '9') {
				return nil, false
			}
		}
		prev = c
	}
	if prev == '_' {
		return nil, false
	}
	return out, true
}

// parseFloatLiteral is PyOS_string_to_double followed by the whole-string check: _Py_dg_strtod
// syntax ([sign] digits [. digits] [e [sign] digits], at least one mantissa digit), falling
// back to _Py_parse_inf_or_nan when no number is recognised.
func parseFloatLiteral(s string) (float64, bool) {
	n := len(s)
	p := 0
	if p < n && (s[p] == '+' || s[p] == '-') {
		p++
	}
	start := p
	nd := 0
	for p < n && s[p] >= '0' && s[p] <= '9' {
		p++
		nd++
	}
	if p < n && s[p] == '.' {
		p++
		for p < n && s[p] >= '0' && s[p] <= '9' {
			p++
			nd++
		}
	}
	if nd == 0 {
		// _Py_dg_strtod consumed nothing; try inf/nan.
		q := start
		neg := s[0] == '-'
		rest := strings.ToLower(s[q:])
		var v float64
		switch {
		case rest == "inf" || rest == "infinity":
			v = math.Inf(1)
		case rest == "nan":
			v = math.NaN()
		default:
			return 0, false
		}
		if neg {
			v = math.Copysign(v, -1)
		}
		return v, true
	}
	if p < n && (s[p] == 'e' || s[p] == 'E') {
		q := p + 1
		if q < n && (s[q] == '+' || s[q] == '-') {
			q++
		}
		ds := q
		for q < n && s[q] >= '0' && s[q] <= '9' {
			q++
		}
		if q > ds {
			p = q
		}
	}
	if p != n {
		return 0, false
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		if ne, ok := err.(*strconv.NumError); !ok || ne.Err != strconv.ErrRange {
			return 0, false
		}
	}
	return f, true
}
