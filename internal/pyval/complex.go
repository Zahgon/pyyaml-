// Port of CPython 3.12's complex semantics: Objects/complexobject.c complex_repr and
// complex_from_string_inner, and Python/pystrtod.c (PSF License Version 2; Copyright (c)
// 2001-2024 Python Software Foundation).

package pyval

import (
	"math"
	"strings"
)

// floatDigits is repr(f) with neither a ".0" nor an added exponent, as complex_repr formats its
// parts with PyOS_double_to_string(x, 'r', 0, 0, NULL).
func floatDigits(f float64) string {
	s := FloatRepr(f)
	return strings.TrimSuffix(s, ".0")
}

// ComplexRepr is repr(complex(re, im)).
func ComplexRepr(re, im float64) string {
	if re == 0 && !math.Signbit(re) {
		return floatDigits(im) + "j"
	}
	sign := ""
	if !math.Signbit(im) || math.IsNaN(im) {
		sign = "+"
	}
	return "(" + floatDigits(re) + sign + floatDigits(im) + "j)"
}

// ParseComplex is complex(s) for a str argument.
func ParseComplex(s string) (float64, float64, error) {
	fail := func() (float64, float64, error) {
		return 0, 0, valueError("complex() arg is a malformed string")
	}
	buf, ok := transformDecimalAndSpace(s)
	if !ok {
		buf = append(buf, '?')
	}
	t := string(buf)
	if strings.IndexByte(t, '_') >= 0 {
		clean, ok := removeUnderscores(buf)
		if !ok {
			return 0, 0, valueError("could not convert string to complex: " + StrRepr(s))
		}
		t = string(clean)
	}
	if strings.IndexByte(t, 0) >= 0 {
		return fail()
	}
	p := 0
	skip := func() {
		for p < len(t) && isPySpaceASCII(t[p]) {
			p++
		}
	}
	skip()
	paren := false
	if p < len(t) && t[p] == '(' {
		paren = true
		p++
		skip()
	}
	var re, im float64
	v, n, ok := floatPrefix(t[p:])
	if ok {
		p += n
		switch {
		case p < len(t) && (t[p] == '+' || t[p] == '-'):
			// <float><signed-float-or-sign>j
			re = v
			w, m, ok := floatPrefix(t[p:])
			if ok {
				p += m
				im = w
			} else {
				if t[p] == '+' {
					im = 1
				} else {
					im = -1
				}
				p++
			}
			if p >= len(t) || (t[p] != 'j' && t[p] != 'J') {
				return fail()
			}
			p++
		case p < len(t) && (t[p] == 'j' || t[p] == 'J'):
			im = v
			p++
		default:
			re = v
		}
	} else {
		// no float: [sign]j
		im = 1
		if p < len(t) && t[p] == '+' {
			p++
		} else if p < len(t) && t[p] == '-' {
			im = -1
			p++
		}
		if p >= len(t) || (t[p] != 'j' && t[p] != 'J') {
			return fail()
		}
		p++
	}
	skip()
	if paren {
		if p >= len(t) || t[p] != ')' {
			return fail()
		}
		p++
		skip()
	}
	if p != len(t) {
		return fail()
	}
	return re, im, nil
}

// floatPrefix is PyOS_string_to_double with an end pointer: the longest float at the start of s,
// without skipping whitespace.
func floatPrefix(s string) (float64, int, bool) {
	p := 0
	if p < len(s) && (s[p] == '+' || s[p] == '-') {
		p++
	}
	nd := 0
	q := p
	for q < len(s) && s[q] >= '0' && s[q] <= '9' {
		q++
		nd++
	}
	if q < len(s) && s[q] == '.' {
		q++
		for q < len(s) && s[q] >= '0' && s[q] <= '9' {
			q++
			nd++
		}
	}
	if nd == 0 {
		rest := strings.ToLower(s[p:])
		neg := p > 0 && s[0] == '-'
		var v float64
		var n int
		switch {
		case strings.HasPrefix(rest, "infinity"):
			v, n = math.Inf(1), 8
		case strings.HasPrefix(rest, "inf"):
			v, n = math.Inf(1), 3
		case strings.HasPrefix(rest, "nan"):
			v, n = math.NaN(), 3
		default:
			return 0, 0, false
		}
		if neg {
			v = math.Copysign(v, -1)
		}
		return v, p + n, true
	}
	if q < len(s) && (s[q] == 'e' || s[q] == 'E') {
		r := q + 1
		if r < len(s) && (s[r] == '+' || s[r] == '-') {
			r++
		}
		ds := r
		for r < len(s) && s[r] >= '0' && s[r] <= '9' {
			r++
		}
		if r > ds {
			q = r
		}
	}
	f, ok := parseFloatLiteral(s[:q])
	if !ok {
		return 0, 0, false
	}
	return f, q, true
}
