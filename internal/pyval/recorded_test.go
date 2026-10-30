// Code generated from CPython 3.12's own answers. DO NOT EDIT.

package pyval

import "testing"

var floatReprRows = [][2]string{
	{"0.0", "0.0"},
	{"-0.0", "-0.0"},
	{"1.0", "1.0"},
	{"0.1", "0.1"},
	{"0.2", "0.2"},
	{"0.30000000000000004", "0.30000000000000004"},
	{"1e15", "1000000000000000.0"},
	{"1e16", "1e+16"},
	{"1e17", "1e+17"},
	{"9999999999999998.0", "9999999999999998.0"},
	{"1e22", "1e+22"},
	{"1e23", "1e+23"},
	{"1e-4", "0.0001"},
	{"1e-5", "1e-05"},
	{"0.0001", "0.0001"},
	{"0.00012", "0.00012"},
	{"5e-324", "5e-324"},
	{"1.7976931348623157e308", "1.7976931348623157e+308"},
	{"2.2250738585072014e-308", "2.2250738585072014e-308"},
	{"1.2345678901234567e30", "1.2345678901234567e+30"},
	{"123456789012345678.0", "1.2345678901234568e+17"},
	{"1234567890123456.0", "1234567890123456.0"},
	{"3.141592653589793", "3.141592653589793"},
	{"100.0", "100.0"},
	{"1e100", "1e+100"},
	{"inf", "inf"},
	{"-inf", "-inf"},
	{"nan", "nan"},
	{"2.675", "2.675"},
	{"4.35", "4.35"},
	{"1.1", "1.1"},
	{"8.41e21", "8.41e+21"},
	{"1e-320", "1e-320"},
	{"1e309", "inf"},
	{"-1e309", "-inf"},
}

var intParseRows = []struct {
	s    string
	base int
	out  string
	err  string
}{
	{"0", 10, "0", ""},
	{"-0", 10, "0", ""},
	{"10", 10, "10", ""},
	{"1_0", 10, "10", ""},
	{"0x1f", 16, "31", ""},
	{"ff", 16, "255", ""},
	{"0b101", 2, "5", ""},
	{"777", 8, "511", ""},
	{"  12  ", 10, "12", ""},
	{"+12", 10, "12", ""},
	{"1__0", 10, "", "ValueError: invalid literal for int() with base 10: '1__0'"},
	{"_1", 10, "", "ValueError: invalid literal for int() with base 10: '_1'"},
	{"1_", 10, "", "ValueError: invalid literal for int() with base 10: '1_'"},
	{"", 10, "", "ValueError: invalid literal for int() with base 10: ''"},
	{"x", 10, "", "ValueError: invalid literal for int() with base 10: 'x'"},
	{"\xd9\xa1\xd9\xa2", 10, "12", ""},
	{"\xdb\xb1", 10, "1", ""},
	{"1\xc2\xa02", 10, "", "ValueError: invalid literal for int() with base 10: '1\x5cxa02'"},
	{"1 2", 10, "", "ValueError: invalid literal for int() with base 10: '1 2'"},
	{"99999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999", 10, "", "ValueError: Exceeds the limit (4300 digits) for integer string conversion: value has 4301 digits; use sys.set_int_max_str_digits() to increase the limit"},
	{"1e3", 10, "", "ValueError: invalid literal for int() with base 10: '1e3'"},
	{"0x", 16, "", "ValueError: invalid literal for int() with base 16: '0x'"},
	{"0X1F", 16, "31", ""},
	{"-0b11", 2, "-3", ""},
	{"\xd9\xa3", 10, "3", ""},
}

var floatParseRows = []struct {
	s   string
	out string
	err string
}{
	{"1.5", "1.5", ""},
	{" 1.5 ", "1.5", ""},
	{"1_000.5", "1000.5", ""},
	{".5", "0.5", ""},
	{"5.", "5.0", ""},
	{"1e3", "1000.0", ""},
	{"1E3", "1000.0", ""},
	{"1e+3", "1000.0", ""},
	{"inf", "inf", ""},
	{"INF", "inf", ""},
	{"Infinity", "inf", ""},
	{"-infinity", "-inf", ""},
	{"nan", "nan", ""},
	{"NAN", "nan", ""},
	{"", "", "ValueError: could not convert string to float: ''"},
	{" ", "", "ValueError: could not convert string to float: ' '"},
	{"x", "", "ValueError: could not convert string to float: 'x'"},
	{"1.5x", "", "ValueError: could not convert string to float: '1.5x'"},
	{"1,5", "", "ValueError: could not convert string to float: '1,5'"},
	{"1e", "", "ValueError: could not convert string to float: '1e'"},
	{"1e+", "", "ValueError: could not convert string to float: '1e+'"},
	{"_1.5", "", "ValueError: could not convert string to float: '_1.5'"},
	{"1._5", "", "ValueError: could not convert string to float: '1._5'"},
	{"1.5_", "", "ValueError: could not convert string to float: '1.5_'"},
	{"\xd9\xa3.\xd9\xa5", "3.5", ""},
	{"1\xc2\xa05", "", "ValueError: could not convert string to float: '1\x5cxa05'"},
	{"1e400", "inf", ""},
	{"-1e400", "-inf", ""},
	{"1e-400", "0.0", ""},
	{"0x1p3", "", "ValueError: could not convert string to float: '0x1p3'"},
	{"1.5e3_0", "1.5e+30", ""},
}

var complexRows = []struct {
	s    string
	repr string
	err  string
}{
	{"0", "0j", ""},
	{"1", "(1+0j)", ""},
	{"1j", "1j", ""},
	{"-1j", "-1j", ""},
	{"1+2j", "(1+2j)", ""},
	{"1-2j", "(1-2j)", ""},
	{"-1-2j", "(-1-2j)", ""},
	{"1.5+2.5j", "(1.5+2.5j)", ""},
	{"(1+2j)", "(1+2j)", ""},
	{" (1+2j) ", "(1+2j)", ""},
	{"1e3+1e-3j", "(1000+0.001j)", ""},
	{"infj", "infj", ""},
	{"nanj", "nanj", ""},
	{"nan+nanj", "(nan+nanj)", ""},
	{"0-0j", "-0j", ""},
	{"-0-0j", "(-0-0j)", ""},
	{"+1j", "1j", ""},
	{"1J", "1j", ""},
	{"j", "1j", ""},
	{"+j", "1j", ""},
	{"-j", "-1j", ""},
	{"2j1", "", "ValueError: complex() arg is a malformed string"},
	{"1+", "", "ValueError: complex() arg is a malformed string"},
	{"", "", "ValueError: complex() arg is a malformed string"},
	{" ", "", "ValueError: complex() arg is a malformed string"},
	{"1 + 2j", "", "ValueError: complex() arg is a malformed string"},
	{"(1+2j", "", "ValueError: complex() arg is a malformed string"},
	{"1+2j)", "", "ValueError: complex() arg is a malformed string"},
	{"1_0+2_0j", "(10+20j)", ""},
	{"1e3j", "1000j", ""},
	{".5j", "0.5j", ""},
	{"5.j", "5j", ""},
	{"infinity", "(inf+0j)", ""},
	{"x", "", "ValueError: complex() arg is a malformed string"},
}

var strReprRows = [][2]string{
	{"", "''"},
	{"a", "'a'"},
	{"a'b", "\x22a'b\x22"},
	{"a\x22b", "'a\x22b'"},
	{"a'b\x22c", "'a\x5c'b\x22c'"},
	{"a\x5cb", "'a\x5c\x5cb'"},
	{"\x09\x0a\x0d", "'\x5ct\x5cn\x5cr'"},
	{"\x00\x01\x1f\x7f", "'\x5cx00\x5cx01\x5cx1f\x5cx7f'"},
	{"\xc2\x80\xc2\x9f\xc2\xa0\xc2\xad", "'\x5cx80\x5cx9f\x5cxa0\x5cxad'"},
	{"caf\xc3\xa9", "'caf\xc3\xa9'"},
	{"\xcd\xb8", "'\x5cu0378'"},
	{"\xe2\x80\x8b", "'\x5cu200b'"},
	{"\xe2\x80\xa8", "'\x5cu2028'"},
	{"\xe3\x80\x80", "'\x5cu3000'"},
	{"\xef\xbb\xbf", "'\x5cufeff'"},
	{"\xef\xbf\xbd", "'\xef\xbf\xbd'"},
	{"\xef\xbf\xbe", "'\x5cufffe'"},
	{"\xef\xbf\xbf", "'\x5cuffff'"},
	{"\xed\xa0\x80", "'\x5cud800'"},
	{"\xed\xbf\xbf", "'\x5cudfff'"},
	{"\xf0\x90\x80\x80", "'\xf0\x90\x80\x80'"},
	{"\xf0\x9f\x98\x80", "'\xf0\x9f\x98\x80'"},
	{"\xf3\xa0\x80\x81", "'\x5cU000e0001'"},
	{"\xf4\x8f\xbf\xbf", "'\x5cU0010ffff'"},
	{"\xc4\xb0", "'\xc4\xb0'"},
	{"\xce\xa3\xce\x91\xce\xa3", "'\xce\xa3\xce\x91\xce\xa3'"},
	{"\xc3\x9f", "'\xc3\x9f'"},
}

var bytesReprRows = []struct {
	data []byte
	repr string
}{
	{[]byte(""), "b''"},
	{[]byte("a"), "b'a'"},
	{[]byte("a'b"), "b\x22a'b\x22"},
	{[]byte("a\x22b"), "b'a\x22b'"},
	{[]byte("\x00\x01\x1f\x7f\x80\xff"), "b'\x5cx00\x5cx01\x5cx1f\x5cx7f\x5cx80\x5cxff'"},
	{[]byte("\x09\x0a\x0d\x5c"), "b'\x5ct\x5cn\x5cr\x5c\x5c'"},
	{[]byte("\x00\x01\x02\x03\x04\x05\x06\x07\x08\x09\x0a\x0b\x0c\x0d\x0e\x0f"), "b'\x5cx00\x5cx01\x5cx02\x5cx03\x5cx04\x5cx05\x5cx06\x5cx07\x5cx08\x5ct\x5cn\x5cx0b\x5cx0c\x5cr\x5cx0e\x5cx0f'"},
}

var lowerRows = [][2]string{
	{"", ""},
	{"a", "a"},
	{"a'b", "a'b"},
	{"a\x22b", "a\x22b"},
	{"a'b\x22c", "a'b\x22c"},
	{"a\x5cb", "a\x5cb"},
	{"\x09\x0a\x0d", "\x09\x0a\x0d"},
	{"\x00\x01\x1f\x7f", "\x00\x01\x1f\x7f"},
	{"\xc2\x80\xc2\x9f\xc2\xa0\xc2\xad", "\xc2\x80\xc2\x9f\xc2\xa0\xc2\xad"},
	{"caf\xc3\xa9", "caf\xc3\xa9"},
	{"\xcd\xb8", "\xcd\xb8"},
	{"\xe2\x80\x8b", "\xe2\x80\x8b"},
	{"\xe2\x80\xa8", "\xe2\x80\xa8"},
	{"\xe3\x80\x80", "\xe3\x80\x80"},
	{"\xef\xbb\xbf", "\xef\xbb\xbf"},
	{"\xef\xbf\xbd", "\xef\xbf\xbd"},
	{"\xef\xbf\xbe", "\xef\xbf\xbe"},
	{"\xef\xbf\xbf", "\xef\xbf\xbf"},
	{"\xed\xa0\x80", "\xed\xa0\x80"},
	{"\xed\xbf\xbf", "\xed\xbf\xbf"},
	{"\xf0\x90\x80\x80", "\xf0\x90\x80\x80"},
	{"\xf0\x9f\x98\x80", "\xf0\x9f\x98\x80"},
	{"\xf3\xa0\x80\x81", "\xf3\xa0\x80\x81"},
	{"\xf4\x8f\xbf\xbf", "\xf4\x8f\xbf\xbf"},
	{"\xc4\xb0", "i\xcc\x87"},
	{"\xce\xa3\xce\x91\xce\xa3", "\xcf\x83\xce\xb1\xcf\x82"},
	{"\xc3\x9f", "\xc3\x9f"},
}

var base64DecodeRows = []struct {
	s    string
	data []byte
	err  string
}{
	{"", []byte(""), ""},
	{"aGk=", []byte("hi"), ""},
	{"aGk", []byte(""), "Incorrect padding"},
	{"YQ==", []byte("a"), ""},
	{"YQ===", []byte("a"), ""},
	{"a", []byte(""), "Invalid base64-encoded string: number of data characters (1) cannot be 1 more than a multiple of 4"},
	{"ab", []byte(""), "Incorrect padding"},
	{"abc", []byte(""), "Incorrect padding"},
	{"abcd", []byte("i\xb7\x1d"), ""},
	{"=", []byte(""), ""},
	{"==", []byte(""), ""},
	{"Y Q==", []byte("a"), ""},
	{"YQ==!!", []byte("a"), ""},
	{"**YQ==", []byte("a"), ""},
	{"/w==", []byte("\xff"), ""},
	{"AP//", []byte("\x00\xff\xff"), ""},
	{"YQ\x0aYQ==", []byte("a\x06\x10"), ""},
	{"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA", []byte("\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00"), ""},
	{"Y===", []byte(""), "Invalid base64-encoded string: number of data characters (1) cannot be 1 more than a multiple of 4"},
	{"Q==", []byte(""), "Invalid base64-encoded string: number of data characters (1) cannot be 1 more than a multiple of 4"},
	{"!!!", []byte(""), ""},
}

var base64EncodeRows = []struct {
	data []byte
	text string
}{
	{[]byte(""), ""},
	{[]byte("a"), "YQ==\x0a"},
	{[]byte("ab"), "YWI=\x0a"},
	{[]byte("abc"), "YWJj\x0a"},
	{[]byte("\x00\x01\x02\x03\x04\x05\x06\x07\x08\x09\x0a\x0b\x0c\x0d\x0e\x0f\x10\x11\x12\x13\x14\x15\x16\x17\x18\x19\x1a\x1b\x1c\x1d\x1e\x1f !\x22#$%&'()*+,-./012345678"), "AAECAwQFBgcICQoLDA0ODxAREhMUFRYXGBkaGxwdHh8gISIjJCUmJygpKissLS4vMDEyMzQ1Njc4\x0a"},
	{[]byte("\x00\x01\x02\x03\x04\x05\x06\x07\x08\x09\x0a\x0b\x0c\x0d\x0e\x0f\x10\x11\x12\x13\x14\x15\x16\x17\x18\x19\x1a\x1b\x1c\x1d\x1e\x1f !\x22#$%&'()*+,-./0123456789"), "AAECAwQFBgcICQoLDA0ODxAREhMUFRYXGBkaGxwdHh8gISIjJCUmJygpKissLS4vMDEyMzQ1Njc4\x0aOQ==\x0a"},
	{[]byte("\x00\x01\x02\x03\x04\x05\x06\x07\x08\x09\x0a\x0b\x0c\x0d\x0e\x0f\x10\x11\x12\x13\x14\x15\x16\x17\x18\x19\x1a\x1b\x1c\x1d\x1e\x1f !\x22#$%&'()*+,-./0123456789:;<=>?@ABCDEFGHIJKLMNOPQRSTUVWXYZ[\x5c]^_`abcdefghijklmnopq"), "AAECAwQFBgcICQoLDA0ODxAREhMUFRYXGBkaGxwdHh8gISIjJCUmJygpKissLS4vMDEyMzQ1Njc4\x0aOTo7PD0+P0BBQkNERUZHSElKS0xNTk9QUVJTVFVWV1hZWltcXV5fYGFiY2RlZmdoaWprbG1ub3Bx\x0a"},
	{[]byte("\x00\x01\x02\x03\x04\x05\x06\x07\x08\x09\x0a\x0b\x0c\x0d\x0e\x0f\x10\x11\x12\x13\x14\x15\x16\x17\x18\x19\x1a\x1b\x1c\x1d\x1e\x1f !\x22#$%&'()*+,-./0123456789:;<=>?@ABCDEFGHIJKLMNOPQRSTUVWXYZ[\x5c]^_`abcdefghijklmnopqr"), "AAECAwQFBgcICQoLDA0ODxAREhMUFRYXGBkaGxwdHh8gISIjJCUmJygpKissLS4vMDEyMzQ1Njc4\x0aOTo7PD0+P0BBQkNERUZHSElKS0xNTk9QUVJTVFVWV1hZWltcXV5fYGFiY2RlZmdoaWprbG1ub3Bx\x0acg==\x0a"},
	{[]byte("\x00\x01\x02\x03\x04\x05\x06\x07\x08\x09\x0a\x0b\x0c\x0d\x0e\x0f\x10\x11\x12\x13\x14\x15\x16\x17\x18\x19\x1a\x1b\x1c\x1d\x1e\x1f !\x22#$%&'()*+,-./0123456789:;<=>?@ABCDEFGHIJKLMNOPQRSTUVWXYZ[\x5c]^_`abcdefghijklmnopqrstuvwxyz{|}~\x7f\x80\x81\x82\x83\x84\x85\x86\x87\x88\x89\x8a\x8b\x8c\x8d\x8e\x8f\x90\x91\x92\x93\x94\x95\x96\x97\x98\x99\x9a\x9b\x9c\x9d\x9e\x9f\xa0\xa1\xa2\xa3\xa4\xa5\xa6\xa7\xa8\xa9\xaa\xab\xac\xad\xae\xaf\xb0\xb1\xb2\xb3\xb4\xb5\xb6\xb7\xb8\xb9\xba\xbb\xbc\xbd\xbe\xbf\xc0\xc1\xc2\xc3\xc4\xc5\xc6\xc7"), "AAECAwQFBgcICQoLDA0ODxAREhMUFRYXGBkaGxwdHh8gISIjJCUmJygpKissLS4vMDEyMzQ1Njc4\x0aOTo7PD0+P0BBQkNERUZHSElKS0xNTk9QUVJTVFVWV1hZWltcXV5fYGFiY2RlZmdoaWprbG1ub3Bx\x0acnN0dXZ3eHl6e3x9fn+AgYKDhIWGh4iJiouMjY6PkJGSk5SVlpeYmZqbnJ2en6ChoqOkpaanqKmq\x0aq6ytrq+wsbKztLW2t7i5uru8vb6/wMHCw8TFxsc=\x0a"},
}

var utf8Rows = []struct {
	data   []byte
	out    string
	start  int
	reason string
}{
	{[]byte(""), "", -1, ""},
	{[]byte("a"), "a", -1, ""},
	{[]byte("\xc3\xa9"), "\xc3\xa9", -1, ""},
	{[]byte("\xe6\xbc\xa2"), "\xe6\xbc\xa2", -1, ""},
	{[]byte("\xf0\x9f\x98\x80"), "\xf0\x9f\x98\x80", -1, ""},
	{[]byte("\xff"), "", 0, "invalid start byte"},
	{[]byte("\xfe"), "", 0, "invalid start byte"},
	{[]byte("\x80"), "", 0, "invalid start byte"},
	{[]byte("\xbf"), "", 0, "invalid start byte"},
	{[]byte("\xc0\x80"), "", 0, "invalid start byte"},
	{[]byte("\xc1\xbf"), "", 0, "invalid start byte"},
	{[]byte("\xc2"), "", 0, "unexpected end of data"},
	{[]byte("\xe0\x80\x80"), "", 0, "invalid continuation byte"},
	{[]byte("\xe0\x9f\xbf"), "", 0, "invalid continuation byte"},
	{[]byte("\xed\xa0\x80"), "", 0, "invalid continuation byte"},
	{[]byte("\xed\xbf\xbf"), "", 0, "invalid continuation byte"},
	{[]byte("\xf0\x80\x80\x80"), "", 0, "invalid continuation byte"},
	{[]byte("\xf0\x8f\xbf\xbf"), "", 0, "invalid continuation byte"},
	{[]byte("\xf4\x90\x80\x80"), "", 0, "invalid continuation byte"},
	{[]byte("\xf5\x80\x80\x80"), "", 0, "invalid start byte"},
	{[]byte("\xe2\x82"), "", 0, "unexpected end of data"},
	{[]byte("\xf0\x9f\x98"), "", 0, "unexpected end of data"},
	{[]byte("\xe2(\xa1"), "", 0, "invalid continuation byte"},
	{[]byte("\xc2A"), "", 0, "invalid continuation byte"},
	{[]byte("a\xffb"), "", 1, "invalid start byte"},
}

var utf16Rows = []struct {
	data   []byte
	big    bool
	out    string
	start  int
	reason string
}{
	{[]byte(""), false, "", -1, ""},
	{[]byte("a\x00"), false, "a", -1, ""},
	{[]byte("\x00a"), true, "a", -1, ""},
	{[]byte("a"), false, "", 0, "truncated data"},
	{[]byte("\x00\xd8\x00\xdc"), false, "\xf0\x90\x80\x80", -1, ""},
	{[]byte("\x00\xdc\x00\xd8"), false, "", 0, "illegal encoding"},
	{[]byte("\x00\xd8"), false, "", 0, "unexpected end of data"},
	{[]byte("\x00\xd8a\x00"), false, "", 0, "illegal UTF-16 surrogate"},
	{[]byte("\xd8\x00\xdc\x00"), true, "\xf0\x90\x80\x80", -1, ""},
	{[]byte("\xdc\x00"), true, "", 0, "illegal encoding"},
	{[]byte("\xd8\x00"), true, "", 0, "unexpected end of data"},
	{[]byte("\xff\xfea\x00"), false, "\xef\xbb\xbfa", -1, ""},
	{[]byte("a\x00\x00"), false, "", 2, "truncated data"},
}

var encodeRows = []struct {
	s        string
	encoding string
	data     []byte
	err      string
}{
	{"a", "utf-8", []byte("a"), ""},
	{"caf\xc3\xa9", "utf-8", []byte("caf\xc3\xa9"), ""},
	{"\xed\xa0\x80", "utf-8", []byte(""), "'utf-8' codec can't encode character '\x5cud800' in position 0: surrogates not allowed"},
	{"a\xed\xbf\xbfb", "utf-8", []byte(""), "'utf-8' codec can't encode character '\x5cudfff' in position 1: surrogates not allowed"},
	{"a", "ascii", []byte("a"), ""},
	{"\xc3\xa9", "ascii", []byte(""), "'ascii' codec can't encode character '\x5cxe9' in position 0: ordinal not in range(128)"},
	{"\xc3\xa9", "latin-1", []byte("\xe9"), ""},
	{"\xc4\x80", "latin-1", []byte(""), "'latin-1' codec can't encode character '\x5cu0100' in position 0: ordinal not in range(256)"},
	{"\xf0\x9f\x98\x80", "utf-16-le", []byte("=\xd8\x00\xde"), ""},
	{"\xed\xa0\x80", "utf-16-le", []byte(""), "'utf-16-le' codec can't encode character '\x5cud800' in position 0: surrogates not allowed"},
	{"a\xc3\xa9", "utf-16-be", []byte("\x00a\x00\xe9"), ""},
	{"ab\xc4\x80\xc4\x81", "ascii", []byte(""), "'ascii' codec can't encode characters in position 2-3: ordinal not in range(128)"},
}

func TestRecordedFloatRepr(t *testing.T) {
	for _, row := range floatReprRows {
		f, err := ParseFloat(row[0])
		if err != nil {
			t.Fatalf("ParseFloat(%q): %v", row[0], err)
		}
		if got := FloatRepr(f); got != row[1] {
			t.Errorf("FloatRepr(%s) = %s, want %s", row[0], got, row[1])
		}
	}
}

func TestRecordedIntParse(t *testing.T) {
	for _, row := range intParseRows {
		z, err := ParseInt(row.s, row.base)
		if err != nil {
			if got := err.(*Error).Class + ": " + err.(*Error).Msg; got != row.err {
				t.Errorf("ParseInt(%q, %d) raised %q, want %q", row.s, row.base, got, row.err)
			}
			continue
		}
		if row.err != "" {
			t.Errorf("ParseInt(%q, %d) = %s, want the exception %q", row.s, row.base, z, row.err)
			continue
		}
		if got := IntRepr(z); got != row.out {
			t.Errorf("ParseInt(%q, %d) = %s, want %s", row.s, row.base, got, row.out)
		}
	}
}

func TestRecordedFloatParse(t *testing.T) {
	for _, row := range floatParseRows {
		f, err := ParseFloat(row.s)
		if err != nil {
			if got := err.(*Error).Class + ": " + err.(*Error).Msg; got != row.err {
				t.Errorf("ParseFloat(%q) raised %q, want %q", row.s, got, row.err)
			}
			continue
		}
		if row.err != "" {
			t.Errorf("ParseFloat(%q) = %v, want the exception %q", row.s, f, row.err)
			continue
		}
		if got := FloatRepr(f); got != row.out {
			t.Errorf("ParseFloat(%q) = %s, want %s", row.s, got, row.out)
		}
	}
}

func TestRecordedComplex(t *testing.T) {
	for _, row := range complexRows {
		re, im, err := ParseComplex(row.s)
		if err != nil {
			if got := err.(*Error).Class + ": " + err.(*Error).Msg; got != row.err {
				t.Errorf("ParseComplex(%q) raised %q, want %q", row.s, got, row.err)
			}
			continue
		}
		if row.err != "" {
			t.Errorf("ParseComplex(%q) = (%v, %v), want the exception %q", row.s, re, im, row.err)
			continue
		}
		if got := ComplexRepr(re, im); got != row.repr {
			t.Errorf("ParseComplex(%q) = %s, want %s", row.s, got, row.repr)
		}
	}
}

func TestRecordedStrRepr(t *testing.T) {
	for _, row := range strReprRows {
		if got := StrRepr(row[0]); got != row[1] {
			t.Errorf("StrRepr(%q) = %s, want %s", row[0], got, row[1])
		}
	}
}

func TestRecordedBytesRepr(t *testing.T) {
	for _, row := range bytesReprRows {
		if got := BytesRepr(row.data); got != row.repr {
			t.Errorf("BytesRepr(%q) = %s, want %s", row.data, got, row.repr)
		}
	}
}

func TestRecordedLower(t *testing.T) {
	for _, row := range lowerRows {
		if got := Lower(row[0]); got != row[1] {
			t.Errorf("Lower(%q) = %q, want %q", row[0], got, row[1])
		}
	}
}

func TestRecordedBase64(t *testing.T) {
	for _, row := range base64DecodeRows {
		data, msg := DecodeBase64([]byte(row.s))
		if msg != row.err {
			t.Errorf("DecodeBase64(%q) raised %q, want %q", row.s, msg, row.err)
			continue
		}
		if msg == "" && string(data) != string(row.data) {
			t.Errorf("DecodeBase64(%q) = %q, want %q", row.s, data, row.data)
		}
	}
	for _, row := range base64EncodeRows {
		if got := EncodeBase64Lines(row.data); got != row.text {
			t.Errorf("EncodeBase64Lines(%q) = %q, want %q", row.data, got, row.text)
		}
	}
}

func TestRecordedCodecs(t *testing.T) {
	for _, row := range utf8Rows {
		rs, err := DecodeUTF8(row.data)
		if err != nil {
			if err.Start != row.start || err.Reason != row.reason {
				t.Errorf("DecodeUTF8(%q) failed at %d (%s), want %d (%s)", row.data, err.Start, err.Reason, row.start, row.reason)
			}
			continue
		}
		if row.start >= 0 {
			t.Errorf("DecodeUTF8(%q) succeeded, want a failure at %d", row.data, row.start)
			continue
		}
		if got := FromRunes(rs); got != row.out {
			t.Errorf("DecodeUTF8(%q) = %q, want %q", row.data, got, row.out)
		}
	}
	for _, row := range utf16Rows {
		rs, err := DecodeUTF16(row.data, row.big)
		if err != nil {
			if err.Start != row.start || err.Reason != row.reason {
				t.Errorf("DecodeUTF16(%q, %v) failed at %d (%s), want %d (%s)", row.data, row.big, err.Start, err.Reason,
					row.start, row.reason)
			}
			continue
		}
		if row.start >= 0 {
			t.Errorf("DecodeUTF16(%q, %v) succeeded, want a failure at %d", row.data, row.big, row.start)
			continue
		}
		if got := FromRunes(rs); got != row.out {
			t.Errorf("DecodeUTF16(%q, %v) = %q, want %q", row.data, row.big, got, row.out)
		}
	}
	for _, row := range encodeRows {
		data, err := Encode(row.s, row.encoding)
		if err != nil {
			if err.Msg != row.err {
				t.Errorf("Encode(%q, %s) raised %q, want %q", row.s, row.encoding, err.Msg, row.err)
			}
			continue
		}
		if row.err != "" {
			t.Errorf("Encode(%q, %s) succeeded, want the exception %q", row.s, row.encoding, row.err)
			continue
		}
		if string(data) != string(row.data) {
			t.Errorf("Encode(%q, %s) = %q, want %q", row.s, row.encoding, data, row.data)
		}
	}
}
