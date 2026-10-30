package yaml

import (
	"encoding/base64"
	"encoding/json"
	"math/big"
	"strconv"
	"testing"
)

// The helpers the generated tests share.

func mustSet(d *Dict, key string, value any) {
	if err := d.Set(key, value); err != nil {
		panic(err)
	}
}

func mustSetKey(d *Dict, key, value any) {
	if err := d.Set(key, value); err != nil {
		panic(err)
	}
}

func mustFloat(lit string) float64 {
	f, err := strconv.ParseFloat(lit, 64)
	if err != nil {
		panic(err)
	}
	return f
}

func mustInt(lit string) *big.Int {
	z, ok := new(big.Int).SetString(lit, 10)
	if !ok {
		panic("not an integer: " + lit)
	}
	return z
}

func mustBytes(lit string) []byte {
	data, err := base64.StdEncoding.DecodeString(lit)
	if err != nil {
		panic(err)
	}
	return data
}

func itoa(n int) string { return strconv.Itoa(n) }

func errOrEmpty(err error) string {
	if err == nil {
		return ""
	}
	return errText(err)
}

// dumpOptionsFromJSON reads the keyword arguments mktests.py recorded for a dump.
func dumpOptionsFromJSON(text string) (DumpOptions, error) {
	opts := NewDumpOptions()
	var raw map[string]any
	if err := json.Unmarshal([]byte(text), &raw); err != nil {
		return opts, err
	}
	for key, value := range raw {
		switch key {
		case "allow_unicode":
			opts.AllowUnicode = value.(bool)
		case "canonical":
			opts.Canonical = value.(bool)
		case "explicit_start":
			opts.ExplicitStart = value.(bool)
		case "explicit_end":
			opts.ExplicitEnd = value.(bool)
		case "sort_keys":
			opts.SortKeys = value.(bool)
		case "default_style":
			opts.DefaultStyle = value.(string)
		case "line_break":
			opts.LineBreak = value.(string)
		case "encoding":
			opts.Encoding = value.(string)
		case "indent":
			opts.Indent = int(value.(float64))
		case "width":
			opts.Width = int(value.(float64))
		case "default_flow_style":
			if value == nil {
				opts.DefaultFlowStyle = nil
			} else {
				flow := value.(bool)
				opts.DefaultFlowStyle = &flow
			}
		case "version":
			pair := value.([]any)
			opts.Version = &Tuple{Items: []any{big.NewInt(int64(pair[0].(float64))), big.NewInt(int64(pair[1].(float64)))}}
		case "tags":
			opts.Tags = map[string]string{}
			for handle, prefix := range value.(map[string]any) {
				opts.Tags[handle] = prefix.(string)
			}
		default:
			return opts, &PyError{Class: "TypeError", Msg: "unexpected keyword argument " + key}
		}
	}
	return opts, nil
}

// The library's own tests, beyond the recorded ones.

func TestDictKeysCompareLikePython(t *testing.T) {
	d := NewDict()
	mustSetKey(d, big.NewInt(1), "int")
	mustSetKey(d, 1.0, "float")
	mustSetKey(d, true, "bool")
	if d.Len() != 1 {
		t.Fatalf("1, 1.0 and True are one key, got %d keys", d.Len())
	}
	keys := d.Keys()
	if _, ok := keys[0].(*big.Int); !ok {
		t.Errorf("the first key object is kept, got %T", keys[0])
	}
	if value, _, _ := d.Get(big.NewInt(1)); value != "bool" {
		t.Errorf("the last value wins, got %v", value)
	}
	if err := d.Set(&List{}, "x"); err == nil {
		t.Error("a list is not hashable, want TypeError")
	} else if pe, ok := err.(*PyError); !ok || pe.Class != "TypeError" {
		t.Errorf("Set with a list = %v, want a TypeError", err)
	}
}

func TestReprOfRecursiveValues(t *testing.T) {
	list := &List{Items: []any{big.NewInt(1)}}
	list.Items = append(list.Items, list)
	got, err := Repr(list)
	if err != nil {
		t.Fatalf("Repr: %v", err)
	}
	if want := "[1, [...]]"; got != want {
		t.Errorf("Repr of a self-referencing list = %s, want %s", got, want)
	}
	d := NewDict()
	mustSet(d, "self", d)
	got, err = Repr(d)
	if err != nil {
		t.Fatalf("Repr: %v", err)
	}
	if want := "{'self': {...}}"; got != want {
		t.Errorf("Repr of a self-referencing dict = %s, want %s", got, want)
	}
}

func TestIntReprLimit(t *testing.T) {
	big1 := new(big.Int).Exp(big.NewInt(10), big.NewInt(4300), nil)
	if _, err := Repr(big1); err == nil {
		t.Error("an int of 4301 digits has no repr, want ValueError")
	} else if pe, ok := err.(*PyError); !ok || pe.Class != "ValueError" {
		t.Errorf("Repr of a 4301-digit int = %v, want a ValueError", err)
	}
}

func TestScanErrorCarriesMarks(t *testing.T) {
	_, err := Scan("a: b\n\tc: d\n")
	marked, ok := err.(*MarkedError)
	if !ok {
		t.Fatalf("Scan of a tab-indented mapping = %v, want a ScannerError", err)
	}
	if marked.Class != "ScannerError" || marked.ProblemMark == nil {
		t.Fatalf("error %#v, want a ScannerError with a problem mark", marked)
	}
	if marked.ProblemMark.Line != 1 || marked.ProblemMark.Column != 0 {
		t.Errorf("problem mark at %d:%d, want 1:0", marked.ProblemMark.Line, marked.ProblemMark.Column)
	}
	if snippet := marked.ProblemMark.Snippet(4, 75); snippet == "" {
		t.Error("the mark has no snippet")
	}
}

func TestLoadAllReturnsDocumentsBeforeAnError(t *testing.T) {
	docs, err := SafeLoadAll("a: 1\n---\nb: [\n")
	if len(docs) != 1 {
		t.Errorf("got %d documents before the error, want 1", len(docs))
	}
	if err == nil {
		t.Error("the second document is malformed, want an error")
	}
}

func TestDumpRoundTripsSharedValues(t *testing.T) {
	inner := NewDict()
	mustSet(inner, "x", big.NewInt(1))
	out, err := Dump(&List{Items: []any{inner, inner}}, NewDumpOptions())
	if err != nil {
		t.Fatalf("Dump: %v", err)
	}
	text, ok := out.(string)
	if !ok {
		t.Fatalf("Dump returned %T, want a string", out)
	}
	value, err := SafeLoad(text)
	if err != nil {
		t.Fatalf("SafeLoad of %q: %v", text, err)
	}
	list, ok := value.(*List)
	if !ok || len(list.Items) != 2 {
		t.Fatalf("loaded %#v, want a list of two items", value)
	}
	if list.Items[0] != list.Items[1] {
		t.Error("the alias did not load as one shared object")
	}
}
