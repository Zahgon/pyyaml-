// Port of the Python 3.12 value semantics PyYAML relies on: Objects/dictobject.c key lookup,
// Objects/longobject.c and Objects/floatobject.c comparison and hashing rules, Objects/listobject.c,
// Objects/tupleobject.c and Objects/setobject.c repr, and Modules/_datetimemodule.c (PSF License
// Version 2; Copyright (c) 2001-2024 Python Software Foundation).

package yaml

import (
	"fmt"
	"math"
	"math/big"
	"sort"
	"strconv"
	"strings"

	"github.com/yaml/pyyaml-go/internal/pyval"
)

// List is a Python list. A *List has identity: two pointers to one List are one object.
type List struct {
	Items []any
}

// Tuple is a Python tuple. A *Tuple has identity.
type Tuple struct {
	Items []any
}

// Complex is a Python complex. A *Complex has identity.
type Complex struct {
	Real float64
	Imag float64
}

// Date is a datetime.date. A *Date has identity.
type Date struct {
	Year  int
	Month int
	Day   int
}

// Timezone is a datetime.timezone with a fixed offset from UTC; a zero offset is
// datetime.timezone.utc.
type Timezone struct {
	OffsetMicroseconds int64
}

// DateTime is a datetime.datetime; TZ is nil for a naive one. A *DateTime has identity.
type DateTime struct {
	Year        int
	Month       int
	Day         int
	Hour        int
	Minute      int
	Second      int
	Microsecond int
	TZ          *Timezone
}

// Dict is a Python dict: insertion-ordered, with Python's key equality. The zero value is empty and
// ready to use. A *Dict has identity.
type Dict struct {
	keys   []any
	values []any
	index  map[string]int
}

// NewDict returns an empty dict.
func NewDict() *Dict { return &Dict{} }

// Set is d[key] = value: a key equal to an existing one keeps the existing key object and replaces
// its value. An unhashable key gives TypeError as a *PyError.
func (d *Dict) Set(key, value any) error {
	h, err := hashKey(key)
	if err != nil {
		return err
	}
	if d.index == nil {
		d.index = map[string]int{}
	}
	if i, ok := d.index[h]; ok {
		d.values[i] = value
		return nil
	}
	d.index[h] = len(d.keys)
	d.keys = append(d.keys, key)
	d.values = append(d.values, value)
	return nil
}

// Get is d.get(key): the value, whether the key is present, and TypeError for an unhashable key.
func (d *Dict) Get(key any) (any, bool, error) {
	h, err := hashKey(key)
	if err != nil {
		return nil, false, err
	}
	i, ok := d.index[h]
	if !ok {
		return nil, false, nil
	}
	return d.values[i], true, nil
}

// Len is len(d).
func (d *Dict) Len() int { return len(d.keys) }

// Keys returns the key objects in insertion order.
func (d *Dict) Keys() []any { return append([]any(nil), d.keys...) }

// Values returns the values in insertion order.
func (d *Dict) Values() []any { return append([]any(nil), d.values...) }

// PySet is a Python set: elements with Python's key equality, the first equal element kept. The zero
// value is empty and ready to use. A *PySet has identity.
type PySet struct {
	dict Dict
}

// NewSet returns an empty set.
func NewSet() *PySet { return &PySet{} }

// Add is s.add(item); an unhashable item gives TypeError as a *PyError.
func (s *PySet) Add(item any) error {
	h, err := hashKey(item)
	if err != nil {
		return err
	}
	if s.dict.index == nil {
		s.dict.index = map[string]int{}
	}
	if _, ok := s.dict.index[h]; ok {
		return nil
	}
	s.dict.index[h] = len(s.dict.keys)
	s.dict.keys = append(s.dict.keys, item)
	s.dict.values = append(s.dict.values, nil)
	return nil
}

// Contains is item in s.
func (s *PySet) Contains(item any) (bool, error) {
	_, ok, err := s.dict.Get(item)
	return ok, err
}

// Len is len(s).
func (s *PySet) Len() int { return s.dict.Len() }

// Items returns the elements in the order Python iterates them where that order is defined here: a
// set of integers ascending, any other set in insertion order.
func (s *PySet) Items() []any {
	items := s.dict.Keys()
	for _, it := range items {
		switch it.(type) {
		case bool, *big.Int:
		default:
			return items
		}
	}
	sort.SliceStable(items, func(i, j int) bool {
		return numericRat(items[i]).Cmp(numericRat(items[j])) < 0
	})
	return items
}

// typeName is type(v).__name__ for the values a dict may meet.
func typeName(v any) string {
	switch v.(type) {
	case nil:
		return "NoneType"
	case bool:
		return "bool"
	case *big.Int:
		return "int"
	case float64:
		return "float"
	case *Complex:
		return "complex"
	case string:
		return "str"
	case []byte:
		return "bytes"
	case *List:
		return "list"
	case *Tuple:
		return "tuple"
	case *Dict:
		return "dict"
	case *PySet:
		return "set"
	case *Date:
		return "date"
	case *DateTime:
		return "datetime"
	case *Timezone:
		return "timezone"
	}
	return fmt.Sprintf("%T", v)
}

func numericRat(v any) *big.Rat {
	switch x := v.(type) {
	case bool:
		if x {
			return big.NewRat(1, 1)
		}
		return new(big.Rat)
	case *big.Int:
		return new(big.Rat).SetInt(x)
	case float64:
		return new(big.Rat).SetFloat64(x)
	}
	return nil
}

// hashKey returns a string equal for exactly the values Python's dict treats as one key: identical
// objects and equal values. A float NaN equals only itself, which here means the same bits (the
// constructor gives each new NaN object bits of its own).
func hashKey(v any) (string, error) {
	switch x := v.(type) {
	case nil:
		return "N", nil
	case bool:
		if x {
			return "n1", nil
		}
		return "n0", nil
	case *big.Int:
		return "n" + x.String(), nil
	case float64:
		return floatKey(x), nil
	case *Complex:
		if x.Imag == 0 {
			return floatKey(x.Real), nil
		}
		return "c" + floatKey(x.Real) + "," + floatKey(x.Imag), nil
	case string:
		return "s" + x, nil
	case []byte:
		return "b" + string(x), nil
	case *Tuple:
		var b strings.Builder
		b.WriteString("t")
		b.WriteString(strconv.Itoa(len(x.Items)))
		for _, it := range x.Items {
			h, err := hashKey(it)
			if err != nil {
				return "", err
			}
			b.WriteString(":")
			b.WriteString(strconv.Itoa(len(h)))
			b.WriteString(":")
			b.WriteString(h)
		}
		return b.String(), nil
	case *Date:
		return fmt.Sprintf("d%d-%d-%d", x.Year, x.Month, x.Day), nil
	case *DateTime:
		if x.TZ == nil {
			return fmt.Sprintf("D%d", x.localMicroseconds()), nil
		}
		return fmt.Sprintf("A%d", x.localMicroseconds()-x.TZ.OffsetMicroseconds), nil
	case *Timezone:
		return fmt.Sprintf("z%d", x.OffsetMicroseconds), nil
	}
	return "", &PyError{Class: "TypeError", Msg: "unhashable type: '" + typeName(v) + "'"}
}

func floatKey(f float64) string {
	switch {
	case math.IsNaN(f):
		return "x" + strconv.FormatUint(math.Float64bits(f), 16)
	case math.IsInf(f, 0):
		return "i" + strconv.FormatFloat(f, 'f', -1, 64)
	case f == math.Trunc(f):
		z, _ := new(big.Float).SetFloat64(f).Int(nil)
		return "n" + z.String()
	}
	return "f" + strconv.FormatUint(math.Float64bits(f), 16)
}

func daysBeforeYear(year int) int64 {
	y := int64(year - 1)
	return y*365 + y/4 - y/100 + y/400
}

var daysBefore = [...]int64{0, 31, 59, 90, 120, 151, 181, 212, 243, 273, 304, 334}

func isLeap(year int) bool { return year%4 == 0 && (year%100 != 0 || year%400 == 0) }

func (t *DateTime) localMicroseconds() int64 {
	days := daysBeforeYear(t.Year) + daysBefore[t.Month-1] + int64(t.Day)
	if t.Month > 2 && isLeap(t.Year) {
		days++
	}
	secs := days*86400 + int64(t.Hour)*3600 + int64(t.Minute)*60 + int64(t.Second)
	return secs*1000000 + int64(t.Microsecond)
}

// pyLess is a < b. ok is false where Python raises TypeError.
func pyLess(a, b any) (less, ok bool) {
	switch x := a.(type) {
	case bool, *big.Int, float64:
		switch b.(type) {
		case bool, *big.Int, float64:
			return numberLess(a, b), true
		}
	case string:
		if y, isStr := b.(string); isStr {
			return x < y, true
		}
	case []byte:
		if y, isBytes := b.([]byte); isBytes {
			return string(x) < string(y), true
		}
	case *Tuple:
		y, isTuple := b.(*Tuple)
		if !isTuple {
			return false, false
		}
		for i := 0; i < len(x.Items) && i < len(y.Items); i++ {
			h1, err1 := hashKey(x.Items[i])
			h2, err2 := hashKey(y.Items[i])
			if err1 != nil || err2 != nil {
				return false, false
			}
			if h1 != h2 {
				return pyLess(x.Items[i], y.Items[i])
			}
		}
		return len(x.Items) < len(y.Items), true
	case *Date:
		if y, isDate := b.(*Date); isDate {
			if x.Year != y.Year {
				return x.Year < y.Year, true
			}
			if x.Month != y.Month {
				return x.Month < y.Month, true
			}
			return x.Day < y.Day, true
		}
	case *DateTime:
		y, isDT := b.(*DateTime)
		if !isDT || (x.TZ == nil) != (y.TZ == nil) {
			return false, false
		}
		m, n := x.localMicroseconds(), y.localMicroseconds()
		if x.TZ != nil {
			m -= x.TZ.OffsetMicroseconds
			n -= y.TZ.OffsetMicroseconds
		}
		return m < n, true
	}
	return false, false
}

func numberLess(a, b any) bool {
	if f, isFloat := a.(float64); isFloat && (math.IsNaN(f) || math.IsInf(f, 0)) {
		if g, isFloat := b.(float64); isFloat {
			return f < g
		}
		return math.IsInf(f, -1)
	}
	if g, isFloat := b.(float64); isFloat && (math.IsNaN(g) || math.IsInf(g, 0)) {
		return math.IsInf(g, 1)
	}
	return numericRat(a).Cmp(numericRat(b)) < 0
}

type reprState struct {
	active map[any]bool
}

// Repr renders v as Python 3.12's repr() does. It handles the Go values that stand for Python values
// and *Token, *Event and *Node. The error is the exception repr() raises, as for an int of more than
// 4300 decimal digits.
func Repr(v any) (string, error) {
	st := &reprState{active: map[any]bool{}}
	var b strings.Builder
	if err := st.write(&b, v); err != nil {
		return "", err
	}
	return b.String(), nil
}

func (st *reprState) writeItems(b *strings.Builder, items []any) error {
	for i, it := range items {
		if i > 0 {
			b.WriteString(", ")
		}
		if err := st.write(b, it); err != nil {
			return err
		}
	}
	return nil
}

func (st *reprState) write(b *strings.Builder, v any) error {
	switch x := v.(type) {
	case nil:
		b.WriteString("None")
	case bool:
		if x {
			b.WriteString("True")
		} else {
			b.WriteString("False")
		}
	case *big.Int:
		s, err := pyval.IntReprChecked(x)
		if err != nil {
			return pyError(err)
		}
		b.WriteString(s)
	case float64:
		b.WriteString(pyval.FloatRepr(x))
	case *Complex:
		b.WriteString(pyval.ComplexRepr(x.Real, x.Imag))
	case string:
		b.WriteString(pyval.StrRepr(x))
	case []byte:
		b.WriteString(pyval.BytesRepr(x))
	case *List:
		if len(x.Items) == 0 {
			b.WriteString("[]")
			return nil
		}
		if st.active[x] {
			b.WriteString("[...]")
			return nil
		}
		st.active[x] = true
		defer delete(st.active, x)
		b.WriteByte('[')
		if err := st.writeItems(b, x.Items); err != nil {
			return err
		}
		b.WriteByte(']')
	case *Tuple:
		if len(x.Items) == 0 {
			b.WriteString("()")
			return nil
		}
		if st.active[x] {
			b.WriteString("(...)")
			return nil
		}
		st.active[x] = true
		defer delete(st.active, x)
		b.WriteByte('(')
		if err := st.writeItems(b, x.Items); err != nil {
			return err
		}
		if len(x.Items) == 1 {
			b.WriteByte(',')
		}
		b.WriteByte(')')
	case *Dict:
		if x.Len() == 0 {
			b.WriteString("{}")
			return nil
		}
		if st.active[x] {
			b.WriteString("{...}")
			return nil
		}
		st.active[x] = true
		defer delete(st.active, x)
		b.WriteByte('{')
		for i, k := range x.keys {
			if i > 0 {
				b.WriteString(", ")
			}
			if err := st.write(b, k); err != nil {
				return err
			}
			b.WriteString(": ")
			if err := st.write(b, x.values[i]); err != nil {
				return err
			}
		}
		b.WriteByte('}')
	case *PySet:
		if x.Len() == 0 {
			b.WriteString("set()")
			return nil
		}
		if st.active[x] {
			b.WriteString("set(...)")
			return nil
		}
		st.active[x] = true
		defer delete(st.active, x)
		b.WriteByte('{')
		if err := st.writeItems(b, x.Items()); err != nil {
			return err
		}
		b.WriteByte('}')
	case *Date:
		fmt.Fprintf(b, "datetime.date(%d, %d, %d)", x.Year, x.Month, x.Day)
	case *DateTime:
		fmt.Fprintf(b, "datetime.datetime(%d, %d, %d, %d, %d", x.Year, x.Month, x.Day, x.Hour, x.Minute)
		switch {
		case x.Microsecond != 0:
			fmt.Fprintf(b, ", %d, %d", x.Second, x.Microsecond)
		case x.Second != 0:
			fmt.Fprintf(b, ", %d", x.Second)
		}
		if x.TZ != nil {
			b.WriteString(", tzinfo=")
			b.WriteString(x.TZ.repr())
		}
		b.WriteByte(')')
	case *Timezone:
		b.WriteString(x.repr())
	case *Token:
		return st.writeToken(b, x)
	case *Event:
		return st.writeEvent(b, x)
	case *Node:
		return st.writeNode(b, x)
	default:
		fmt.Fprintf(b, "<%T>", v)
	}
	return nil
}

func (z *Timezone) repr() string {
	if z.OffsetMicroseconds == 0 {
		return "datetime.timezone.utc"
	}
	return "datetime.timezone(" + timedeltaRepr(z.OffsetMicroseconds) + ")"
}

// timedeltaRepr is repr(datetime.timedelta(microseconds=us)).
func timedeltaRepr(us int64) string {
	const dayUS = 86400 * 1000000
	days := us / dayUS
	rest := us % dayUS
	if rest < 0 {
		days--
		rest += dayUS
	}
	seconds, micro := rest/1000000, rest%1000000
	var parts []string
	if days != 0 {
		parts = append(parts, fmt.Sprintf("days=%d", days))
	}
	if seconds != 0 {
		parts = append(parts, fmt.Sprintf("seconds=%d", seconds))
	}
	if micro != 0 {
		parts = append(parts, fmt.Sprintf("microseconds=%d", micro))
	}
	if len(parts) == 0 {
		return "datetime.timedelta(0)"
	}
	return "datetime.timedelta(" + strings.Join(parts, ", ") + ")"
}

// isoformat is d.isoformat().
func (d *Date) isoformat() string { return fmt.Sprintf("%04d-%02d-%02d", d.Year, d.Month, d.Day) }

// isoformat is t.isoformat(' ').
func (t *DateTime) isoformat() string {
	s := fmt.Sprintf("%04d-%02d-%02d %02d:%02d:%02d", t.Year, t.Month, t.Day, t.Hour, t.Minute, t.Second)
	if t.Microsecond != 0 {
		s += fmt.Sprintf(".%06d", t.Microsecond)
	}
	if t.TZ != nil {
		off := t.TZ.OffsetMicroseconds
		sign := "+"
		if off < 0 {
			sign = "-"
			off = -off
		}
		micro := off % 1000000
		secs := off / 1000000
		s += fmt.Sprintf("%s%02d:%02d", sign, secs/3600, secs/60%60)
		if secs%60 != 0 || micro != 0 {
			s += fmt.Sprintf(":%02d", secs%60)
			if micro != 0 {
				s += fmt.Sprintf(".%06d", micro)
			}
		}
	}
	return s
}
