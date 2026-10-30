// Code generated from PyYAML at commit f5e4504a, running under CPython 3.12. DO NOT EDIT.

package yaml

import (
	"math/big"
	"strings"
	"testing"
)

// loaderByName maps the recorded loader name to its Loader.
func loaderByName(name string) Loader {
	switch name {
	case "base":
		return BaseLoader
	case "full":
		return FullLoader
	}
	return SafeLoader
}

type loadRow struct {
	input  string
	loader string
	repr   string
	err    string
}

type dumpRow struct {
	kind string
	lit  string
	opts string
	safe bool
	out  string
	err  string
}

type stageRow struct {
	input    string
	tokens   []string
	tokenErr string
	events   []string
	eventErr string
}

var loadRows = []loadRow{
	{"yes\x0a", "safe", "True", ""},
	{"Yes\x0a", "safe", "True", ""},
	{"yEs\x0a", "safe", "'yEs'", ""},
	{"off\x0a", "safe", "False", ""},
	{"~\x0a", "safe", "None", ""},
	{"null\x0a", "safe", "None", ""},
	{"''\x0a", "safe", "''", ""},
	{"\x0a", "safe", "None", ""},
	{"0\x0a", "safe", "0", ""},
	{"00\x0a", "safe", "0", ""},
	{"017\x0a", "safe", "15", ""},
	{"0b101\x0a", "safe", "5", ""},
	{"0x1f\x0a", "safe", "31", ""},
	{"1_000\x0a", "safe", "1000", ""},
	{"1:30\x0a", "safe", "90", ""},
	{"190:20:30\x0a", "safe", "685230", ""},
	{"1.5\x0a", "safe", "1.5", ""},
	{".5\x0a", "safe", "0.5", ""},
	{"1.\x0a", "safe", "1.0", ""},
	{"1e3\x0a", "safe", "'1e3'", ""},
	{"1.0e+3\x0a", "safe", "1000.0", ""},
	{".inf\x0a", "safe", "inf", ""},
	{"-.INF\x0a", "safe", "-inf", ""},
	{".nan\x0a", "safe", "nan", ""},
	{"1:30.5\x0a", "safe", "90.5", ""},
	{"1e16\x0a", "safe", "'1e16'", ""},
	{"!!float 1e16\x0a", "safe", "1e+16", ""},
	{"!!float 1e15\x0a", "safe", "1000000000000000.0", ""},
	{"!!float 5e-324\x0a", "safe", "5e-324", ""},
	{"!!float 0.1\x0a", "safe", "0.1", ""},
	{"!!int 0x10000000000000000\x0a", "safe", "18446744073709551616", ""},
	{"2001-12-14\x0a", "safe", "datetime.date(2001, 12, 14)", ""},
	{"2001-12-14 21:59:43.10 -5\x0a", "safe", "datetime.datetime(2001, 12, 14, 21, 59, 43, 100000, tzinfo=datetime.timezone(datetime.timedelta(days=-1, seconds=68400)))", ""},
	{"2001-12-14T21:59:43Z\x0a", "safe", "datetime.datetime(2001, 12, 14, 21, 59, 43, tzinfo=datetime.timezone.utc)", ""},
	{"2001-12-14 21:59:43+05:45\x0a", "safe", "datetime.datetime(2001, 12, 14, 21, 59, 43, tzinfo=datetime.timezone(datetime.timedelta(seconds=20700)))", ""},
	{"!!timestamp '2001-13-01'\x0a", "safe", "", "ValueError: month must be in 1..12"},
	{"!!timestamp 'x'\x0a", "safe", "", "AttributeError: 'NoneType' object has no attribute 'groupdict'"},
	{"!!binary aGVsbG8=\x0a", "safe", "b'hello'", ""},
	{"!!binary 'a'\x0a", "safe", "", "ConstructorError: failed to decode base64 data: Invalid base64-encoded string: number of data characters (1) cannot be 1 more than a multiple of 4\x0a  in \x22<unicode string>\x22, line 1, column 1:\x0a    !!binary 'a'\x0a    ^"},
	{"!!binary ''\x0a", "safe", "b''", ""},
	{"!!set {a, b, a}\x0a", "safe", "{'a', 'b'}", ""},
	{"!!omap [{a: 1}, {b: 2}]\x0a", "safe", "[('a', 1), ('b', 2)]", ""},
	{"!!pairs [{a: 1}, {a: 2}]\x0a", "safe", "[('a', 1), ('a', 2)]", ""},
	{"1: a\x0a1.0: b\x0atrue: c\x0a", "safe", "{1: 'c'}", ""},
	{"? [1, 2]\x0a: x\x0a", "safe", "", "ConstructorError: while constructing a mapping\x0a  in \x22<unicode string>\x22, line 1, column 1:\x0a    ? [1, 2]\x0a    ^\x0afound unhashable key\x0a  in \x22<unicode string>\x22, line 1, column 3:\x0a    ? [1, 2]\x0a      ^"},
	{"? {a: 1}\x0a: x\x0a", "safe", "", "ConstructorError: while constructing a mapping\x0a  in \x22<unicode string>\x22, line 1, column 1:\x0a    ? {a: 1}\x0a    ^\x0afound unhashable key\x0a  in \x22<unicode string>\x22, line 1, column 3:\x0a    ? {a: 1}\x0a      ^"},
	{"x: &a {p: 1}\x0ay:\x0a  <<: *a\x0a  p: 2\x0a", "safe", "{'x': {'p': 1}, 'y': {'p': 2}}", ""},
	{"x: &a {p: 1}\x0ay: &b {q: 2}\x0az:\x0a  <<: [*a, *b]\x0a", "safe", "{'x': {'p': 1}, 'y': {'q': 2}, 'z': {'p': 1, 'q': 2}}", ""},
	{"=: v\x0aa: 1\x0a", "safe", "{'=': 'v', 'a': 1}", ""},
	{"- &a [1]\x0a- *a\x0a", "safe", "[[1], [1]]", ""},
	{"&a [*a]\x0a", "safe", "[[...]]", ""},
	{"a: *b\x0a", "safe", "", "ComposerError: found undefined alias 'b'\x0a  in \x22<unicode string>\x22, line 1, column 4:\x0a    a: *b\x0a       ^"},
	{"- &a x\x0a- &a y\x0a", "safe", "", "ComposerError: found duplicate anchor 'a'; first occurrence\x0a  in \x22<unicode string>\x22, line 1, column 3:\x0a    - &a x\x0a      ^\x0asecond occurrence\x0a  in \x22<unicode string>\x22, line 2, column 3:\x0a    - &a y\x0a      ^"},
	{"\x22\x5cx41\x5cu0100\x5cU0001f600\x22\x0a", "safe", "'A\xc4\x80\xf0\x9f\x98\x80'", ""},
	{"\x22\x5c0\x5ca\x5cb\x5ct\x5cn\x5cv\x5cf\x5cr\x5ce\x22\x0a", "safe", "'\x5cx00\x5cx07\x5cx08\x5ct\x5cn\x5cx0b\x5cx0c\x5cr\x5cx1b'", ""},
	{"\x22\x5cN\x5c_\x5cL\x5cP\x22\x0a", "safe", "'\x5cx85\x5cxa0\x5cu2028\x5cu2029'", ""},
	{"\x22\x5cU00110000\x22\x0a", "safe", "", "ValueError: chr() arg not in range(0x110000)"},
	{"\x22\x5cq\x22\x0a", "safe", "", "ScannerError: while scanning a double-quoted scalar\x0a  in \x22<unicode string>\x22, line 1, column 1:\x0a    \x22\x5cq\x22\x0a    ^\x0afound unknown escape character 'q'\x0a  in \x22<unicode string>\x22, line 1, column 3:\x0a    \x22\x5cq\x22\x0a      ^"},
	{"'a''b'\x0a", "safe", "\x22a'b\x22", ""},
	{"|\x0a  a\x0a  b\x0a", "safe", "'a\x5cnb\x5cn'", ""},
	{"|-\x0a  a\x0a\x0a", "safe", "'a'", ""},
	{"|+\x0a  a\x0a\x0a", "safe", "'a\x5cn\x5cn'", ""},
	{">\x0a  a\x0a  b\x0a\x0a  c\x0a", "safe", "'a b\x5cnc\x5cn'", ""},
	{">2\x0a   a\x0a", "safe", "' a\x5cn'", ""},
	{"|0\x0a  a\x0a", "safe", "", "ScannerError: while scanning a block scalar\x0a  in \x22<unicode string>\x22, line 1, column 1:\x0a    |0\x0a    ^\x0aexpected indentation indicator in the range 1-9, but found 0\x0a  in \x22<unicode string>\x22, line 1, column 2:\x0a    |0\x0a     ^"},
	{"a: b\x0a\x09c: d\x0a", "safe", "", "ScannerError: while scanning for the next token\x0afound character '\x5ct' that cannot start any token\x0a  in \x22<unicode string>\x22, line 2, column 1:\x0a    \x09c: d\x0a    ^"},
	{"a: b: c\x0a", "safe", "", "ScannerError: mapping values are not allowed here\x0a  in \x22<unicode string>\x22, line 1, column 5:\x0a    a: b: c\x0a        ^"},
	{"[a b]\x0a", "safe", "['a b']", ""},
	{"{a: b\x0a", "safe", "", "ParserError: while parsing a flow mapping\x0a  in \x22<unicode string>\x22, line 1, column 1:\x0a    {a: b\x0a    ^\x0aexpected ',' or '}', but got '<stream end>'\x0a  in \x22<unicode string>\x22, line 2, column 1:\x0a    \x0a    ^"},
	{"- a\x0a- b\x0a", "base", "['a', 'b']", ""},
	{"yes\x0a", "base", "'yes'", ""},
	{"1\x0a", "base", "'1'", ""},
	{"!!int 1\x0a", "base", "'1'", ""},
	{"!!python/tuple [1, 2]\x0a", "full", "(1, 2)", ""},
	{"!!python/complex '1+2j'\x0a", "full", "(1+2j)", ""},
	{"!!python/complex 'infj'\x0a", "full", "infj", ""},
	{"!!python/bytes 'aGk='\x0a", "full", "b'hi'", ""},
	{"!!python/long '0x1f'\x0a", "full", "31", ""},
	{"!!python/tuple [1]\x0a", "safe", "", "ConstructorError: could not determine a constructor for the tag 'tag:yaml.org,2002:python/tuple'\x0a  in \x22<unicode string>\x22, line 1, column 1:\x0a    !!python/tuple [1]\x0a    ^"},
	{"!unknown x\x0a", "safe", "", "ConstructorError: could not determine a constructor for the tag '!unknown'\x0a  in \x22<unicode string>\x22, line 1, column 1:\x0a    !unknown x\x0a    ^"},
	{"!!str [1]\x0a", "safe", "", "ConstructorError: expected a scalar node, but found sequence\x0a  in \x22<unicode string>\x22, line 1, column 1:\x0a    !!str [1]\x0a    ^"},
	{"!!int ''\x0a", "safe", "", "IndexError: string index out of range"},
	{"!!bool x\x0a", "safe", "", "KeyError: 'x'"},
	{"!!int 'x'\x0a", "safe", "", "ValueError: invalid literal for int() with base 10: 'x'"},
	{"a: 1\x0aa: 2\x0a", "safe", "{'a': 2}", ""},
	{"? !!set {a}\x0a: x\x0a", "safe", "", "ConstructorError: while constructing a mapping\x0a  in \x22<unicode string>\x22, line 1, column 1:\x0a    ? !!set {a}\x0a    ^\x0afound unhashable key\x0a  in \x22<unicode string>\x22, line 1, column 3:\x0a    ? !!set {a}\x0a      ^"},
	{"- 2001-12-14 21:59:43.1234567\x0a", "safe", "[datetime.datetime(2001, 12, 14, 21, 59, 43, 123456)]", ""},
	{"- 9999-12-31 23:59:59.999999\x0a", "safe", "[datetime.datetime(9999, 12, 31, 23, 59, 59, 999999)]", ""},
	{"- 0001-01-01 00:00:00+01:00\x0a", "safe", "[datetime.datetime(1, 1, 1, 0, 0, tzinfo=datetime.timezone(datetime.timedelta(seconds=3600)))]", ""},
	{"- !!timestamp '2001-12-14 21:59:43+24:00'\x0a", "safe", "", "ValueError: offset must be a timedelta strictly between -timedelta(hours=24) and timedelta(hours=24), not datetime.timedelta(days=1)."},
	{"a: |\x0a  x\x0ab: >-\x0a  y\x0a", "safe", "{'a': 'x\x5cn', 'b': 'y'}", ""},
	{"---\x0aa\x0a---\x0ab\x0a", "safe", "", "ComposerError: expected a single document in the stream\x0a  in \x22<unicode string>\x22, line 2, column 1:\x0a    a\x0a    ^\x0abut found another document\x0a  in \x22<unicode string>\x22, line 3, column 1:\x0a    ---\x0a    ^"},
	{"a\x0a...\x0ab\x0a", "safe", "", "ParserError: expected '<document start>', but found '<scalar>'\x0a  in \x22<unicode string>\x22, line 3, column 1:\x0a    b\x0a    ^"},
}

var dumpRows = []dumpRow{
	{"str", "yes", "{}", false, "\x22'yes'\x5cn\x22", ""},
	{"str", "", "{}", false, "\x22''\x5cn\x22", ""},
	{"str", " a", "{}", false, "\x22' a'\x5cn\x22", ""},
	{"str", "a\x0ab", "{}", false, "\x22'a\x5cn\x5cn  b'\x5cn\x22", ""},
	{"str", "a: b", "{}", false, "\x22'a: b'\x5cn\x22", ""},
	{"str", "- a", "{}", false, "\x22'- a'\x5cn\x22", ""},
	{"str", "#a", "{}", false, "\x22'#a'\x5cn\x22", ""},
	{"str", "a\x00b", "{}", false, "'\x22a\x5c\x5c0b\x22\x5cn'", ""},
	{"str", "caf\xc3\xa9", "{}", false, "'\x22caf\x5c\x5cxE9\x22\x5cn'", ""},
	{"str", "caf\xc3\xa9", "{\x22allow_unicode\x22: true}", false, "'caf\xc3\xa9\x5cn...\x5cn'", ""},
	{"str", "\xf0\x9f\x98\x80", "{}", false, "'\x22\x5c\x5cU0001F600\x22\x5cn'", ""},
	{"str", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", "{}", false, "'aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa\x5cn...\x5cn'", ""},
	{"str", "word word word word word word word word word word word word word word word word word word word word word word word word word word word word word word", "{}", false, "'word word word word word word word word word word word word word word word word word\x5cn  word word word word word word word word word word word word word\x5cn...\x5cn'", ""},
	{"str", "word word word word word word word word word word word word word word word word word word word word word word word word word word word word word word", "{\x22width\x22: 20}", false, "'word word word word word\x5cn  word word word word\x5cn  word word word word\x5cn  word word word word\x5cn  word word word word\x5cn  word word word word\x5cn  word word word word\x5cn  word\x5cn...\x5cn'", ""},
	{"str", "word word word word word word word word word word word word word word word word word word word word word word word word word word word word word word", "{\x22width\x22: 1}", false, "'word word word word word word word word word word word word word word word word word\x5cn  word word word word word word word word word word word word word\x5cn...\x5cn'", ""},
	{"str", "a\x0ab\x0a", "{\x22default_style\x22: \x22|\x22}", false, "'|\x5cn  a\x5cn  b\x5cn'", ""},
	{"str", "a\x0ab\x0a", "{\x22default_style\x22: \x22>\x22}", false, "'>\x5cn  a\x5cn\x5cn  b\x5cn'", ""},
	{"str", "it's", "{\x22default_style\x22: \x22'\x22}", false, "\x22'it''s'\x5cn\x22", ""},
	{"str", "x", "{\x22default_style\x22: \x22\x5c\x22\x22}", false, "'\x22x\x22\x5cn'", ""},
	{"float", "0.1", "{}", false, "'0.1\x5cn...\x5cn'", ""},
	{"float", "1e+16", "{}", false, "'1.0e+16\x5cn...\x5cn'", ""},
	{"float", "1000000000000000.0", "{}", false, "'1000000000000000.0\x5cn...\x5cn'", ""},
	{"float", "inf", "{}", false, "'.inf\x5cn...\x5cn'", ""},
	{"float", "-inf", "{}", false, "'-.inf\x5cn...\x5cn'", ""},
	{"float", "nan", "{}", false, "'.nan\x5cn...\x5cn'", ""},
	{"float", "-0.0", "{}", false, "'-0.0\x5cn...\x5cn'", ""},
	{"float", "5e-324", "{}", false, "'5.0e-324\x5cn...\x5cn'", ""},
	{"float", "1e+100", "{}", false, "'1.0e+100\x5cn...\x5cn'", ""},
	{"int", "1180591620717411303424", "{}", false, "'1180591620717411303424\x5cn...\x5cn'", ""},
	{"int", "-1180591620717411303424", "{}", false, "'-1180591620717411303424\x5cn...\x5cn'", ""},
	{"bytes", "aGVsbG8=", "{}", false, "'!!binary |\x5cn  aGVsbG8=\x5cn'", ""},
	{"bytes", "AAECAwQFBgcICQoLDA0ODxAREhMUFRYXGBkaGxwdHh8gISIjJCUmJygpKissLS4vMDEyMzQ1Njc4OTo7", "{}", false, "'!!binary |\x5cn  AAECAwQFBgcICQoLDA0ODxAREhMUFRYXGBkaGxwdHh8gISIjJCUmJygpKissLS4vMDEyMzQ1Njc4\x5cn  OTo7\x5cn'", ""},
	{"bytes", "", "{}", false, "'!!binary \x22\x22\x5cn'", ""},
	{"date", "", "{}", false, "'2001-12-14\x5cn...\x5cn'", ""},
	{"datetime", "", "{}", false, "'2001-12-14 21:59:43.100000\x5cn...\x5cn'", ""},
	{"datetime-tz", "", "{}", false, "'2001-12-14 21:59:43+05:45\x5cn...\x5cn'", ""},
	{"map", "", "{}", true, "'a:\x5cn- 1\x5cn- true\x5cn- null\x5cnb: 2\x5cnc: x\x5cn'", ""},
	{"map", "", "{\x22default_flow_style\x22: true}", false, "'{a: [1, true, null], b: 2, c: x}\x5cn'", ""},
	{"map", "", "{\x22default_flow_style\x22: null}", false, "'a: [1, true, null]\x5cnb: 2\x5cnc: x\x5cn'", ""},
	{"map", "", "{\x22indent\x22: 9}", false, "'a:\x5cn- 1\x5cn- true\x5cn- null\x5cnb: 2\x5cnc: x\x5cn'", ""},
	{"map", "", "{\x22canonical\x22: true}", false, "'---\x5cn!!map {\x5cn  ? !!str \x22a\x22\x5cn  : !!seq [\x5cn    !!int \x221\x22,\x5cn    !!bool \x22true\x22,\x5cn    !!null \x22null\x22,\x5cn  ],\x5cn  ? !!str \x22b\x22\x5cn  : !!int \x222\x22,\x5cn  ? !!str \x22c\x22\x5cn  : !!str \x22x\x22,\x5cn}\x5cn'", ""},
	{"map", "", "{\x22sort_keys\x22: false}", false, "'b: 2\x5cna:\x5cn- 1\x5cn- true\x5cn- null\x5cnc: x\x5cn'", ""},
	{"nested", "", "{}", false, "'a:\x5cn  b:\x5cn  - c:\x5cn    - 1\x5cn    - 2\x5cn  - d: {}\x5cne: []\x5cn'", ""},
	{"nested", "", "{\x22default_flow_style\x22: true}", false, "'{a: {b: [{c: [1, 2]}, {d: {}}]}, e: []}\x5cn'", ""},
	{"nested", "", "{\x22indent\x22: 5, \x22width\x22: 30}", false, "'a:\x5cn     b:\x5cn     -    c:\x5cn          - 1\x5cn          - 2\x5cn     -    d: {}\x5cne: []\x5cn'", ""},
	{"shared", "", "{}", false, "'- &id001\x5cn  x: 1\x5cn- *id001\x5cn- y: *id001\x5cn'", ""},
	{"recursive", "", "{}", false, "'&id001\x5cn- 1\x5cn- *id001\x5cn'", ""},
	{"tuple", "", "{}", true, "'- 1\x5cn- - 2\x5cn  - a\x5cn'", ""},
	{"set", "", "{}", true, "'!!set\x5cn1: null\x5cn2: null\x5cn3: null\x5cn'", ""},
	{"complex", "", "{}", true, "", "RepresenterError: ('cannot represent an object', (1+2j))"},
	{"mixedkeys", "", "{}", false, "'1: a\x5cnb: 2\x5cn1.5: c\x5cnnull: d\x5cn'", ""},
	{"map", "", "{\x22explicit_start\x22: true, \x22explicit_end\x22: true}", false, "'---\x5cna:\x5cn- 1\x5cn- true\x5cn- null\x5cnb: 2\x5cnc: x\x5cn...\x5cn'", ""},
	{"map", "", "{\x22version\x22: [1, 1]}", false, "'%YAML 1.1\x5cn---\x5cna:\x5cn- 1\x5cn- true\x5cn- null\x5cnb: 2\x5cnc: x\x5cn'", ""},
	{"map", "", "{\x22tags\x22: {\x22!e!\x22: \x22tag:example.com,2000:\x22}}", false, "'%TAG !e! tag:example.com,2000:\x5cn---\x5cna:\x5cn- 1\x5cn- true\x5cn- null\x5cnb: 2\x5cnc: x\x5cn'", ""},
	{"str", "caf\xc3\xa9", "{\x22encoding\x22: \x22utf-8\x22}", false, "b'\x22caf\x5c\x5cxE9\x22\x5cn'", ""},
	{"str", "a", "{\x22encoding\x22: \x22utf-16\x22}", false, "b'\x5cxff\x5cxfe\x5cxff\x5cxfe\x5cxff\x5cxfea\x5cx00\x5cxff\x5cxfe\x5cn\x5cx00\x5cxff\x5cxfe.\x5cx00.\x5cx00.\x5cx00\x5cxff\x5cxfe\x5cn\x5cx00'", ""},
	{"str", "a\x0ab", "{\x22line_break\x22: \x22\x5cr\x5cn\x22}", false, "\x22'a\x5cr\x5cn\x5cr\x5cn  b'\x5cr\x5cn\x22", ""},
	{"tuple", "", "{}", true, "'- 1\x5cn- - 2\x5cn  - a\x5cn'", ""},
	{"complex", "", "{}", true, "", "RepresenterError: ('cannot represent an object', (1+2j))"},
	{"set", "", "{}", true, "'!!set\x5cn1: null\x5cn2: null\x5cn3: null\x5cn'", ""},
	{"map", "", "{}", true, "'a:\x5cn- 1\x5cn- true\x5cn- null\x5cnb: 2\x5cnc: x\x5cn'", ""},
}

var stageRows = []stageRow{
	{"a: 1\x0a", []string{"<stream start> StreamStartToken(encoding=None) 0:0", "<block mapping start> BlockMappingStartToken() 0:0", "? KeyToken() 0:0", "<scalar> ScalarToken(plain=True, style=None, value='a') 0:0", ": ValueToken() 0:1", "<scalar> ScalarToken(plain=True, style=None, value='1') 0:3", "<block end> BlockEndToken() 1:0", "<stream end> StreamEndToken() 1:0"}, "", []string{"StreamStartEvent()", "DocumentStartEvent()", "MappingStartEvent(anchor=None, tag=None, implicit=True)", "ScalarEvent(anchor=None, tag=None, implicit=(True, False), value='a')", "ScalarEvent(anchor=None, tag=None, implicit=(True, False), value='1')", "MappingEndEvent()", "DocumentEndEvent()", "StreamEndEvent()"}, ""},
	{"- &a x\x0a- *a\x0a", []string{"<stream start> StreamStartToken(encoding=None) 0:0", "<block sequence start> BlockSequenceStartToken() 0:0", "- BlockEntryToken() 0:0", "<anchor> AnchorToken(value='a') 0:2", "<scalar> ScalarToken(plain=True, style=None, value='x') 0:5", "- BlockEntryToken() 1:0", "<alias> AliasToken(value='a') 1:2", "<block end> BlockEndToken() 2:0", "<stream end> StreamEndToken() 2:0"}, "", []string{"StreamStartEvent()", "DocumentStartEvent()", "SequenceStartEvent(anchor=None, tag=None, implicit=True)", "ScalarEvent(anchor='a', tag=None, implicit=(True, False), value='x')", "AliasEvent(anchor='a')", "SequenceEndEvent()", "DocumentEndEvent()", "StreamEndEvent()"}, ""},
	{"%YAML 1.1\x0a--- !!str x\x0a", []string{"<stream start> StreamStartToken(encoding=None) 0:0", "<directive> DirectiveToken(name='YAML', value=(1, 1)) 0:0", "<document start> DocumentStartToken() 1:0", "<tag> TagToken(value=('!!', 'str')) 1:4", "<scalar> ScalarToken(plain=True, style=None, value='x') 1:10", "<stream end> StreamEndToken() 2:0"}, "", []string{"StreamStartEvent()", "DocumentStartEvent()", "ScalarEvent(anchor=None, tag='tag:yaml.org,2002:str', implicit=(False, False), value='x')", "DocumentEndEvent()", "StreamEndEvent()"}, ""},
	{"{a: [1, 2]}\x0a", []string{"<stream start> StreamStartToken(encoding=None) 0:0", "{ FlowMappingStartToken() 0:0", "? KeyToken() 0:1", "<scalar> ScalarToken(plain=True, style=None, value='a') 0:1", ": ValueToken() 0:2", "[ FlowSequenceStartToken() 0:4", "<scalar> ScalarToken(plain=True, style=None, value='1') 0:5", ", FlowEntryToken() 0:6", "<scalar> ScalarToken(plain=True, style=None, value='2') 0:8", "] FlowSequenceEndToken() 0:9", "} FlowMappingEndToken() 0:10", "<stream end> StreamEndToken() 1:0"}, "", []string{"StreamStartEvent()", "DocumentStartEvent()", "MappingStartEvent(anchor=None, tag=None, implicit=True)", "ScalarEvent(anchor=None, tag=None, implicit=(True, False), value='a')", "SequenceStartEvent(anchor=None, tag=None, implicit=True)", "ScalarEvent(anchor=None, tag=None, implicit=(True, False), value='1')", "ScalarEvent(anchor=None, tag=None, implicit=(True, False), value='2')", "SequenceEndEvent()", "MappingEndEvent()", "DocumentEndEvent()", "StreamEndEvent()"}, ""},
	{"a: b\x0a\x09c\x0a", []string{"<stream start> StreamStartToken(encoding=None) 0:0", "<block mapping start> BlockMappingStartToken() 0:0", "? KeyToken() 0:0", "<scalar> ScalarToken(plain=True, style=None, value='a') 0:0", ": ValueToken() 0:1", "<scalar> ScalarToken(plain=True, style=None, value='b') 0:3"}, "ScannerError: while scanning for the next token\x0afound character '\x5ct' that cannot start any token\x0a  in \x22<unicode string>\x22, line 2, column 1:\x0a    \x09c\x0a    ^", []string{"StreamStartEvent()", "DocumentStartEvent()", "MappingStartEvent(anchor=None, tag=None, implicit=True)", "ScalarEvent(anchor=None, tag=None, implicit=(True, False), value='a')", "ScalarEvent(anchor=None, tag=None, implicit=(True, False), value='b')"}, "ScannerError: while scanning for the next token\x0afound character '\x5ct' that cannot start any token\x0a  in \x22<unicode string>\x22, line 2, column 1:\x0a    \x09c\x0a    ^"},
	{"? a\x0a: b\x0a", []string{"<stream start> StreamStartToken(encoding=None) 0:0", "<block mapping start> BlockMappingStartToken() 0:0", "? KeyToken() 0:0", "<scalar> ScalarToken(plain=True, style=None, value='a') 0:2", ": ValueToken() 1:0", "<scalar> ScalarToken(plain=True, style=None, value='b') 1:2", "<block end> BlockEndToken() 2:0", "<stream end> StreamEndToken() 2:0"}, "", []string{"StreamStartEvent()", "DocumentStartEvent()", "MappingStartEvent(anchor=None, tag=None, implicit=True)", "ScalarEvent(anchor=None, tag=None, implicit=(True, False), value='a')", "ScalarEvent(anchor=None, tag=None, implicit=(True, False), value='b')", "MappingEndEvent()", "DocumentEndEvent()", "StreamEndEvent()"}, ""},
}

// errText is the class and message of an exception, as mktests.py recorded it.
func errText(err error) string {
	switch e := err.(type) {
	case *MarkedError:
		return e.Class + ": " + e.Error()
	case *ReaderError:
		return "ReaderError: " + e.Error()
	case *YAMLError:
		return e.Class + ": " + e.Error()
	case *PyError:
		return e.Class + ": " + e.Error()
	}
	return "unknown: " + err.Error()
}

func TestRecordedLoad(t *testing.T) {
	for _, row := range loadRows {
		v, err := Load(row.input, loaderByName(row.loader))
		if err != nil {
			if got := errText(err); got != row.err {
				t.Errorf("Load(%q, %s) raised %s, want %s", row.input, row.loader, got, row.err)
			}
			continue
		}
		if row.err != "" {
			t.Errorf("Load(%q, %s) returned a value, want the exception %s", row.input, row.loader, row.err)
			continue
		}
		got, err := Repr(v)
		if err != nil {
			t.Errorf("Repr after Load(%q, %s): %v", row.input, row.loader, err)
			continue
		}
		if got != row.repr {
			t.Errorf("Load(%q, %s) = %s, want %s", row.input, row.loader, got, row.repr)
		}
	}
}

func TestRecordedDump(t *testing.T) {
	for _, row := range dumpRows {
		opts, err := dumpOptionsFromJSON(row.opts)
		if err != nil {
			t.Fatalf("options %s: %v", row.opts, err)
		}
		v := testValue(row.kind, row.lit)
		var out any
		if row.safe {
			out, err = SafeDump(v, opts)
		} else {
			out, err = Dump(v, opts)
		}
		if err != nil {
			if got := errText(err); got != row.err {
				t.Errorf("Dump(%s, %s) raised %s, want %s", row.kind, row.opts, got, row.err)
			}
			continue
		}
		if row.err != "" {
			t.Errorf("Dump(%s, %s) returned output, want the exception %s", row.kind, row.opts, row.err)
			continue
		}
		var got string
		switch text := out.(type) {
		case string:
			got, err = Repr(text)
		case []byte:
			got, err = Repr(text)
		}
		if err != nil {
			t.Errorf("Repr of the dump of %s: %v", row.kind, err)
			continue
		}
		if got != row.out {
			t.Errorf("Dump(%s, %s) = %s, want %s", row.kind, row.opts, got, row.out)
		}
	}
}

func TestRecordedStages(t *testing.T) {
	for _, row := range stageRows {
		tokens, err := Scan(row.input)
		var lines []string
		for _, tok := range tokens {
			text, rerr := Repr(tok)
			if rerr != nil {
				t.Fatalf("Repr of a token of %q: %v", row.input, rerr)
			}
			lines = append(lines, tok.ID+" "+text+" "+itoa(tok.StartMark.Line)+":"+itoa(tok.StartMark.Column))
		}
		if strings.Join(lines, "|") != strings.Join(row.tokens, "|") {
			t.Errorf("Scan(%q) = %v, want %v", row.input, lines, row.tokens)
		}
		if got := errOrEmpty(err); got != row.tokenErr {
			t.Errorf("Scan(%q) raised %q, want %q", row.input, got, row.tokenErr)
		}
		events, err := Parse(row.input)
		lines = nil
		for _, ev := range events {
			text, rerr := Repr(ev)
			if rerr != nil {
				t.Fatalf("Repr of an event of %q: %v", row.input, rerr)
			}
			lines = append(lines, text)
		}
		if strings.Join(lines, "|") != strings.Join(row.events, "|") {
			t.Errorf("Parse(%q) = %v, want %v", row.input, lines, row.events)
		}
		if got := errOrEmpty(err); got != row.eventErr {
			t.Errorf("Parse(%q) raised %q, want %q", row.input, got, row.eventErr)
		}
	}
}

// testValue builds the value mktests.py dumped under this name; lit carries the literal of a scalar.
func testValue(kind, lit string) any {
	switch kind {
	case "str":
		return lit
	case "float":
		return mustFloat(lit)
	case "int":
		return mustInt(lit)
	case "bytes":
		return mustBytes(lit)
	case "map":
		d := NewDict()
		mustSet(d, "b", big.NewInt(2))
		mustSet(d, "a", &List{Items: []any{big.NewInt(1), true, nil}})
		mustSet(d, "c", "x")
		return d
	case "nested":
		inner := NewDict()
		mustSet(inner, "c", &List{Items: []any{big.NewInt(1), big.NewInt(2)}})
		empty := NewDict()
		second := NewDict()
		mustSet(second, "d", empty)
		b := NewDict()
		mustSet(b, "b", &List{Items: []any{inner, second}})
		out := NewDict()
		mustSet(out, "a", b)
		mustSet(out, "e", &List{})
		return out
	case "shared":
		inner := NewDict()
		mustSet(inner, "x", big.NewInt(1))
		holder := NewDict()
		mustSet(holder, "y", inner)
		return &List{Items: []any{inner, inner, holder}}
	case "recursive":
		out := &List{Items: []any{big.NewInt(1)}}
		out.Items = append(out.Items, out)
		return out
	case "tuple":
		return &Tuple{Items: []any{big.NewInt(1), &Tuple{Items: []any{big.NewInt(2), "a"}}}}
	case "set":
		s := NewSet()
		for _, n := range []int64{1, 2, 3} {
			if err := s.Add(big.NewInt(n)); err != nil {
				panic(err)
			}
		}
		return s
	case "complex":
		return &Complex{Real: 1, Imag: 2}
	case "mixedkeys":
		d := NewDict()
		mustSetKey(d, big.NewInt(1), "a")
		mustSetKey(d, "b", big.NewInt(2))
		mustSetKey(d, 1.5, "c")
		mustSetKey(d, nil, "d")
		return d
	case "date":
		return &Date{Year: 2001, Month: 12, Day: 14}
	case "datetime":
		return &DateTime{Year: 2001, Month: 12, Day: 14, Hour: 21, Minute: 59, Second: 43, Microsecond: 100000}
	case "datetime-tz":
		return &DateTime{Year: 2001, Month: 12, Day: 14, Hour: 21, Minute: 59, Second: 43,
			TZ: &Timezone{OffsetMicroseconds: (5*3600 + 45*60) * 1000000}}
	}
	panic("unknown test value " + kind)
}
