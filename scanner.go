// Port of PyYAML's lib/yaml/scanner.py (MIT License; Copyright (c) 2017-2021 Ingy döt Net, Copyright
// (c) 2006-2016 Kirill Simonov; see LICENSE).

package yaml

import (
	"fmt"
	"math/big"

	"github.com/yaml/pyyaml-go/internal/pyval"
)

type simpleKey struct {
	level       int
	tokenNumber int
	required    bool
	index       int
	line        int
	column      int
	mark        *Mark
}

// scanner turns the characters of a reader into tokens, following scanner.py method by method.
type scanner struct {
	*reader
	done            bool
	flowLevel       int
	tokens          []*Token
	tokensTaken     int
	indent          int
	indents         []int
	allowSimpleKey  bool
	possibleSimples []*simpleKey
}

func newScanner(r *reader) *scanner {
	s := &scanner{reader: r, indent: -1, allowSimpleKey: true}
	s.fetchStreamStart()
	return s
}

func scannerError(context string, contextMark *Mark, problem string, problemMark *Mark) {
	var c *string
	if context != "" {
		c = &context
	}
	raise(markedError("ScannerError", c, contextMark, problem, problemMark))
}

func rch(r rune) string { return pyval.StrRepr(pyval.FromRunes([]rune{r})) }

func isIn(r rune, set string) bool {
	for _, c := range set {
		if c == r {
			return true
		}
	}
	return false
}

// Character classes of scanner.py; a NUL ends the stream.
func isZBLT(r rune) bool { // '\0 \t\r\n\x85\u2028\u2029'
	switch r {
	case 0, ' ', '\t', '\r', '\n', 0x85, 0x2028, 0x2029:
		return true
	}
	return false
}

func isZBL(r rune) bool { // '\0 \r\n\x85\u2028\u2029'
	switch r {
	case 0, ' ', '\r', '\n', 0x85, 0x2028, 0x2029:
		return true
	}
	return false
}

func isZL(r rune) bool { // '\0\r\n\x85\u2028\u2029'
	switch r {
	case 0, '\r', '\n', 0x85, 0x2028, 0x2029:
		return true
	}
	return false
}

func isL(r rune) bool { // '\r\n\x85\u2028\u2029'
	switch r {
	case '\r', '\n', 0x85, 0x2028, 0x2029:
		return true
	}
	return false
}

func isHex(r rune) bool {
	return (r >= '0' && r <= '9') || (r >= 'A' && r <= 'F') || (r >= 'a' && r <= 'f')
}

func isWordChar(r rune) bool {
	return (r >= '0' && r <= '9') || (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') || r == '-' || r == '_'
}

func (s *scanner) checkToken(choices ...string) bool {
	for s.needMoreTokens() {
		s.fetchMoreTokens()
	}
	if len(s.tokens) > 0 {
		if len(choices) == 0 {
			return true
		}
		for _, c := range choices {
			if s.tokens[0].Kind == c {
				return true
			}
		}
	}
	return false
}

func (s *scanner) peekToken() *Token {
	for s.needMoreTokens() {
		s.fetchMoreTokens()
	}
	if len(s.tokens) > 0 {
		return s.tokens[0]
	}
	return nil
}

func (s *scanner) getToken() *Token {
	for s.needMoreTokens() {
		s.fetchMoreTokens()
	}
	if len(s.tokens) > 0 {
		s.tokensTaken++
		t := s.tokens[0]
		s.tokens = s.tokens[1:]
		return t
	}
	return nil
}

func (s *scanner) needMoreTokens() bool {
	if s.done {
		return false
	}
	if len(s.tokens) == 0 {
		return true
	}
	s.stalePossibleSimpleKeys()
	n, ok := s.nextPossibleSimpleKey()
	return ok && n == s.tokensTaken
}

func (s *scanner) fetchMoreTokens() {
	s.scanToNextToken()
	s.stalePossibleSimpleKeys()
	s.unwindIndent(s.column)
	ch := s.peek(0)
	switch {
	case ch == 0:
		s.fetchStreamEnd()
	case ch == '%' && s.column == 0:
		s.fetchDirective()
	case ch == '-' && s.checkDocumentIndicator("---"):
		s.fetchDocumentIndicator(DocumentStartToken)
	case ch == '.' && s.checkDocumentIndicator("..."):
		s.fetchDocumentIndicator(DocumentEndToken)
	case ch == '[':
		s.fetchFlowCollectionStart(FlowSequenceStartToken)
	case ch == '{':
		s.fetchFlowCollectionStart(FlowMappingStartToken)
	case ch == ']':
		s.fetchFlowCollectionEnd(FlowSequenceEndToken)
	case ch == '}':
		s.fetchFlowCollectionEnd(FlowMappingEndToken)
	case ch == ',':
		s.fetchFlowEntry()
	case ch == '-' && isZBLT(s.peek(1)):
		s.fetchBlockEntry()
	case ch == '?' && (s.flowLevel != 0 || isZBLT(s.peek(1))):
		s.fetchKey()
	case ch == ':' && (s.flowLevel != 0 || isZBLT(s.peek(1))):
		s.fetchValue()
	case ch == '*':
		s.savePossibleSimpleKey()
		s.allowSimpleKey = false
		s.tokens = append(s.tokens, s.scanAnchor(AliasToken))
	case ch == '&':
		s.savePossibleSimpleKey()
		s.allowSimpleKey = false
		s.tokens = append(s.tokens, s.scanAnchor(AnchorToken))
	case ch == '!':
		s.savePossibleSimpleKey()
		s.allowSimpleKey = false
		s.tokens = append(s.tokens, s.scanTag())
	case ch == '|' && s.flowLevel == 0:
		s.fetchBlockScalar('|')
	case ch == '>' && s.flowLevel == 0:
		s.fetchBlockScalar('>')
	case ch == '\'':
		s.fetchFlowScalar('\'')
	case ch == '"':
		s.fetchFlowScalar('"')
	case s.checkPlain():
		s.fetchPlain()
	default:
		scannerError("while scanning for the next token", nil,
			fmt.Sprintf("found character %s that cannot start any token", rch(ch)), s.mark())
	}
}

func (s *scanner) nextPossibleSimpleKey() (int, bool) {
	found := false
	minNumber := 0
	for _, key := range s.possibleSimples {
		if !found || key.tokenNumber < minNumber {
			found = true
			minNumber = key.tokenNumber
		}
	}
	return minNumber, found
}

func (s *scanner) stalePossibleSimpleKeys() {
	kept := s.possibleSimples[:0:0]
	for i, key := range s.possibleSimples {
		if key.line != s.line || s.index-key.index > 1024 {
			if key.required {
				s.possibleSimples = append(kept, s.possibleSimples[i:]...)
				scannerError("while scanning a simple key", key.mark, "could not find expected ':'", s.mark())
			}
			continue
		}
		kept = append(kept, key)
	}
	s.possibleSimples = kept
}

func (s *scanner) savePossibleSimpleKey() {
	required := s.flowLevel == 0 && s.indent == s.column
	if s.allowSimpleKey {
		s.removePossibleSimpleKey()
		s.possibleSimples = append(s.possibleSimples, &simpleKey{
			level: s.flowLevel, tokenNumber: s.tokensTaken + len(s.tokens), required: required,
			index: s.index, line: s.line, column: s.column, mark: s.mark(),
		})
	}
}

func (s *scanner) simpleKeyAt(level int) int {
	for i, key := range s.possibleSimples {
		if key.level == level {
			return i
		}
	}
	return -1
}

func (s *scanner) dropSimpleKey(i int) {
	s.possibleSimples = append(s.possibleSimples[:i:i], s.possibleSimples[i+1:]...)
}

func (s *scanner) removePossibleSimpleKey() {
	if i := s.simpleKeyAt(s.flowLevel); i >= 0 {
		key := s.possibleSimples[i]
		if key.required {
			scannerError("while scanning a simple key", key.mark, "could not find expected ':'", s.mark())
		}
		s.dropSimpleKey(i)
	}
}

func (s *scanner) unwindIndent(column int) {
	// In the flow context indentation is ignored. A stray ']' or '}' in the block context makes the
	// level negative, which Python also treats as "in flow".
	if s.flowLevel != 0 {
		return
	}
	for s.indent > column {
		mark := s.mark()
		s.indent = s.indents[len(s.indents)-1]
		s.indents = s.indents[:len(s.indents)-1]
		s.tokens = append(s.tokens, newToken(BlockEndToken, mark, mark))
	}
}

func (s *scanner) addIndent(column int) bool {
	if s.indent < column {
		s.indents = append(s.indents, s.indent)
		s.indent = column
		return true
	}
	return false
}

func (s *scanner) fetchStreamStart() {
	mark := s.mark()
	t := newToken(StreamStartToken, mark, mark)
	t.Encoding = s.encoding
	s.tokens = append(s.tokens, t)
}

func (s *scanner) fetchStreamEnd() {
	s.unwindIndent(-1)
	s.removePossibleSimpleKey()
	s.allowSimpleKey = false
	s.possibleSimples = nil
	mark := s.mark()
	s.tokens = append(s.tokens, newToken(StreamEndToken, mark, mark))
	s.done = true
}

func (s *scanner) fetchDirective() {
	s.unwindIndent(-1)
	s.removePossibleSimpleKey()
	s.allowSimpleKey = false
	s.tokens = append(s.tokens, s.scanDirective())
}

func (s *scanner) fetchDocumentIndicator(kind string) {
	s.unwindIndent(-1)
	s.removePossibleSimpleKey()
	s.allowSimpleKey = false
	start := s.mark()
	s.forward(3)
	s.tokens = append(s.tokens, newToken(kind, start, s.mark()))
}

func (s *scanner) fetchFlowCollectionStart(kind string) {
	s.savePossibleSimpleKey()
	s.flowLevel++
	s.allowSimpleKey = true
	start := s.mark()
	s.forward(1)
	s.tokens = append(s.tokens, newToken(kind, start, s.mark()))
}

func (s *scanner) fetchFlowCollectionEnd(kind string) {
	s.removePossibleSimpleKey()
	s.flowLevel--
	s.allowSimpleKey = false
	start := s.mark()
	s.forward(1)
	s.tokens = append(s.tokens, newToken(kind, start, s.mark()))
}

func (s *scanner) fetchFlowEntry() {
	s.allowSimpleKey = true
	s.removePossibleSimpleKey()
	start := s.mark()
	s.forward(1)
	s.tokens = append(s.tokens, newToken(FlowEntryToken, start, s.mark()))
}

func (s *scanner) fetchBlockEntry() {
	if s.flowLevel == 0 {
		if !s.allowSimpleKey {
			scannerError("", nil, "sequence entries are not allowed here", s.mark())
		}
		if s.addIndent(s.column) {
			mark := s.mark()
			s.tokens = append(s.tokens, newToken(BlockSequenceStartToken, mark, mark))
		}
	}
	s.allowSimpleKey = true
	s.removePossibleSimpleKey()
	start := s.mark()
	s.forward(1)
	s.tokens = append(s.tokens, newToken(BlockEntryToken, start, s.mark()))
}

func (s *scanner) fetchKey() {
	if s.flowLevel == 0 {
		if !s.allowSimpleKey {
			scannerError("", nil, "mapping keys are not allowed here", s.mark())
		}
		if s.addIndent(s.column) {
			mark := s.mark()
			s.tokens = append(s.tokens, newToken(BlockMappingStartToken, mark, mark))
		}
	}
	s.allowSimpleKey = s.flowLevel == 0
	s.removePossibleSimpleKey()
	start := s.mark()
	s.forward(1)
	s.tokens = append(s.tokens, newToken(KeyToken, start, s.mark()))
}

func (s *scanner) insertToken(at int, t *Token) {
	s.tokens = append(s.tokens, nil)
	copy(s.tokens[at+1:], s.tokens[at:])
	s.tokens[at] = t
}

func (s *scanner) fetchValue() {
	if i := s.simpleKeyAt(s.flowLevel); i >= 0 {
		key := s.possibleSimples[i]
		s.dropSimpleKey(i)
		s.insertToken(key.tokenNumber-s.tokensTaken, newToken(KeyToken, key.mark, key.mark))
		if s.flowLevel == 0 {
			if s.addIndent(key.column) {
				s.insertToken(key.tokenNumber-s.tokensTaken, newToken(BlockMappingStartToken, key.mark, key.mark))
			}
		}
		s.allowSimpleKey = false
	} else {
		if s.flowLevel == 0 {
			if !s.allowSimpleKey {
				scannerError("", nil, "mapping values are not allowed here", s.mark())
			}
		}
		if s.flowLevel == 0 {
			if s.addIndent(s.column) {
				mark := s.mark()
				s.tokens = append(s.tokens, newToken(BlockMappingStartToken, mark, mark))
			}
		}
		s.allowSimpleKey = s.flowLevel == 0
		s.removePossibleSimpleKey()
	}
	start := s.mark()
	s.forward(1)
	s.tokens = append(s.tokens, newToken(ValueToken, start, s.mark()))
}

func (s *scanner) fetchBlockScalar(style rune) {
	s.allowSimpleKey = true
	s.removePossibleSimpleKey()
	s.tokens = append(s.tokens, s.scanBlockScalar(style))
}

func (s *scanner) fetchFlowScalar(style rune) {
	s.savePossibleSimpleKey()
	s.allowSimpleKey = false
	s.tokens = append(s.tokens, s.scanFlowScalar(style))
}

func (s *scanner) fetchPlain() {
	s.savePossibleSimpleKey()
	s.allowSimpleKey = false
	s.tokens = append(s.tokens, s.scanPlain())
}

func (s *scanner) checkDocumentIndicator(text string) bool {
	return s.column == 0 && s.prefix(3) == text && isZBLT(s.peek(3))
}

func (s *scanner) checkPlain() bool {
	ch := s.peek(0)
	if !isZBLT(ch) && !isIn(ch, "-?:,[]{}#&*!|>'\"%@`") {
		return true
	}
	return !isZBLT(s.peek(1)) && (ch == '-' || (s.flowLevel == 0 && (ch == '?' || ch == ':')))
}

func (s *scanner) scanToNextToken() {
	if s.index == 0 && s.peek(0) == 0xFEFF {
		s.forward(1)
	}
	found := false
	for !found {
		for s.peek(0) == ' ' {
			s.forward(1)
		}
		if s.peek(0) == '#' {
			for !isZL(s.peek(0)) {
				s.forward(1)
			}
		}
		if s.scanLineBreak() != "" {
			if s.flowLevel == 0 {
				s.allowSimpleKey = true
			}
		} else {
			found = true
		}
	}
}

func (s *scanner) scanDirective() *Token {
	start := s.mark()
	s.forward(1)
	name := s.scanDirectiveName(start)
	var value any
	var end *Mark
	switch name {
	case "YAML":
		value = s.scanYAMLDirectiveValue(start)
		end = s.mark()
	case "TAG":
		value = s.scanTagDirectiveValue(start)
		end = s.mark()
	default:
		end = s.mark()
		for !isZL(s.peek(0)) {
			s.forward(1)
		}
	}
	s.scanDirectiveIgnoredLine(start)
	t := newToken(DirectiveToken, start, end)
	t.Name = name
	t.Value = value
	return t
}

func (s *scanner) scanDirectiveName(start *Mark) string {
	length := 0
	ch := s.peek(length)
	for isWordChar(ch) {
		length++
		ch = s.peek(length)
	}
	if length == 0 {
		scannerError("while scanning a directive", start,
			fmt.Sprintf("expected alphabetic or numeric character, but found %s", rch(ch)), s.mark())
	}
	value := s.prefix(length)
	s.forward(length)
	ch = s.peek(0)
	if !isZBL(ch) {
		scannerError("while scanning a directive", start,
			fmt.Sprintf("expected alphabetic or numeric character, but found %s", rch(ch)), s.mark())
	}
	return value
}

func (s *scanner) scanYAMLDirectiveValue(start *Mark) *Tuple {
	for s.peek(0) == ' ' {
		s.forward(1)
	}
	major := s.scanYAMLDirectiveNumber(start)
	if s.peek(0) != '.' {
		scannerError("while scanning a directive", start,
			fmt.Sprintf("expected a digit or '.', but found %s", rch(s.peek(0))), s.mark())
	}
	s.forward(1)
	minor := s.scanYAMLDirectiveNumber(start)
	if !isZBL(s.peek(0)) {
		scannerError("while scanning a directive", start,
			fmt.Sprintf("expected a digit or ' ', but found %s", rch(s.peek(0))), s.mark())
	}
	return &Tuple{Items: []any{major, minor}}
}

func (s *scanner) scanYAMLDirectiveNumber(start *Mark) *big.Int {
	ch := s.peek(0)
	if ch < '0' || ch > '9' {
		scannerError("while scanning a directive", start,
			fmt.Sprintf("expected a digit, but found %s", rch(ch)), s.mark())
	}
	length := 0
	for s.peek(length) >= '0' && s.peek(length) <= '9' {
		length++
	}
	value, err := pyval.ParseInt(s.prefix(length), 10)
	if err != nil {
		raise(pyError(err))
	}
	s.forward(length)
	return value
}

func (s *scanner) scanTagDirectiveValue(start *Mark) *Tuple {
	for s.peek(0) == ' ' {
		s.forward(1)
	}
	handle := s.scanTagDirectiveHandle(start)
	for s.peek(0) == ' ' {
		s.forward(1)
	}
	prefix := s.scanTagDirectivePrefix(start)
	return &Tuple{Items: []any{handle, prefix}}
}

func (s *scanner) scanTagDirectiveHandle(start *Mark) string {
	value := s.scanTagHandle("directive", start)
	ch := s.peek(0)
	if ch != ' ' {
		scannerError("while scanning a directive", start, fmt.Sprintf("expected ' ', but found %s", rch(ch)), s.mark())
	}
	return value
}

func (s *scanner) scanTagDirectivePrefix(start *Mark) string {
	value := s.scanTagURI("directive", start)
	ch := s.peek(0)
	if !isZBL(ch) {
		scannerError("while scanning a directive", start, fmt.Sprintf("expected ' ', but found %s", rch(ch)), s.mark())
	}
	return value
}

func (s *scanner) scanDirectiveIgnoredLine(start *Mark) {
	for s.peek(0) == ' ' {
		s.forward(1)
	}
	if s.peek(0) == '#' {
		for !isZL(s.peek(0)) {
			s.forward(1)
		}
	}
	ch := s.peek(0)
	if !isZL(ch) {
		scannerError("while scanning a directive", start,
			fmt.Sprintf("expected a comment or a line break, but found %s", rch(ch)), s.mark())
	}
	s.scanLineBreak()
}

func (s *scanner) scanAnchor(kind string) *Token {
	start := s.mark()
	name := "anchor"
	if s.peek(0) == '*' {
		name = "alias"
	}
	s.forward(1)
	length := 0
	ch := s.peek(length)
	for isWordChar(ch) {
		length++
		ch = s.peek(length)
	}
	if length == 0 {
		scannerError("while scanning an "+name, start,
			fmt.Sprintf("expected alphabetic or numeric character, but found %s", rch(ch)), s.mark())
	}
	value := s.prefix(length)
	s.forward(length)
	ch = s.peek(0)
	if !isZBLT(ch) && !isIn(ch, "?:,]}%@`") {
		scannerError("while scanning an "+name, start,
			fmt.Sprintf("expected alphabetic or numeric character, but found %s", rch(ch)), s.mark())
	}
	t := newToken(kind, start, s.mark())
	t.Value = value
	return t
}

func (s *scanner) scanTag() *Token {
	start := s.mark()
	ch := s.peek(1)
	var handle any
	var suffix string
	switch {
	case ch == '<':
		s.forward(2)
		suffix = s.scanTagURI("tag", start)
		if s.peek(0) != '>' {
			scannerError("while parsing a tag", start, fmt.Sprintf("expected '>', but found %s", rch(s.peek(0))), s.mark())
		}
		s.forward(1)
	case isZBLT(ch):
		suffix = "!"
		s.forward(1)
	default:
		length := 1
		useHandle := false
		for !isZBL(ch) {
			if ch == '!' {
				useHandle = true
				break
			}
			length++
			ch = s.peek(length)
		}
		if useHandle {
			handle = s.scanTagHandle("tag", start)
		} else {
			handle = "!"
			s.forward(1)
		}
		suffix = s.scanTagURI("tag", start)
	}
	ch = s.peek(0)
	if !isZBL(ch) {
		scannerError("while scanning a tag", start, fmt.Sprintf("expected ' ', but found %s", rch(ch)), s.mark())
	}
	t := newToken(TagToken, start, s.mark())
	t.Value = &Tuple{Items: []any{handle, suffix}}
	return t
}

func (s *scanner) scanBlockScalar(style rune) *Token {
	folded := style == '>'
	var chunks []byte
	start := s.mark()
	s.forward(1)
	chomping, increment := s.scanBlockScalarIndicators(start)
	s.scanBlockScalarIgnoredLine(start)
	minIndent := s.indent + 1
	if minIndent < 1 {
		minIndent = 1
	}
	var breaks []string
	var end *Mark
	var indent int
	if increment == 0 {
		var maxIndent int
		breaks, maxIndent, end = s.scanBlockScalarIndentation()
		indent = max(minIndent, maxIndent)
	} else {
		indent = minIndent + increment - 1
		breaks, end = s.scanBlockScalarBreaks(indent)
	}
	lineBreak := ""
	for s.column == indent && s.peek(0) != 0 {
		for _, br := range breaks {
			chunks = append(chunks, br...)
		}
		leadingNonSpace := s.peek(0) != ' ' && s.peek(0) != '\t'
		length := 0
		for !isZL(s.peek(length)) {
			length++
		}
		chunks = append(chunks, s.prefix(length)...)
		s.forward(length)
		lineBreak = s.scanLineBreak()
		breaks, end = s.scanBlockScalarBreaks(indent)
		if s.column == indent && s.peek(0) != 0 {
			if folded && lineBreak == "\n" && leadingNonSpace && s.peek(0) != ' ' && s.peek(0) != '\t' {
				if len(breaks) == 0 {
					chunks = append(chunks, ' ')
				}
			} else {
				chunks = append(chunks, lineBreak...)
			}
		} else {
			break
		}
	}
	if chomping >= 0 {
		chunks = append(chunks, lineBreak...)
	}
	if chomping > 0 {
		for _, br := range breaks {
			chunks = append(chunks, br...)
		}
	}
	t := newToken(ScalarToken, start, end)
	t.Value = string(chunks)
	t.Style = string(style)
	return t
}

// scanBlockScalarIndicators returns the chomping (+1 True, -1 False, 0 None) and the increment
// (0 for None).
func (s *scanner) scanBlockScalarIndicators(start *Mark) (int, int) {
	chomping, increment := 0, 0
	badZero := func() {
		scannerError("while scanning a block scalar", start,
			"expected indentation indicator in the range 1-9, but found 0", s.mark())
	}
	ch := s.peek(0)
	if ch == '+' || ch == '-' {
		chomping = -1
		if ch == '+' {
			chomping = 1
		}
		s.forward(1)
		ch = s.peek(0)
		if ch >= '0' && ch <= '9' {
			increment = int(ch - '0')
			if increment == 0 {
				badZero()
			}
			s.forward(1)
		}
	} else if ch >= '0' && ch <= '9' {
		increment = int(ch - '0')
		if increment == 0 {
			badZero()
		}
		s.forward(1)
		ch = s.peek(0)
		if ch == '+' || ch == '-' {
			chomping = -1
			if ch == '+' {
				chomping = 1
			}
			s.forward(1)
		}
	}
	ch = s.peek(0)
	if !isZBL(ch) {
		scannerError("while scanning a block scalar", start,
			fmt.Sprintf("expected chomping or indentation indicators, but found %s", rch(ch)), s.mark())
	}
	return chomping, increment
}

func (s *scanner) scanBlockScalarIgnoredLine(start *Mark) {
	for s.peek(0) == ' ' {
		s.forward(1)
	}
	if s.peek(0) == '#' {
		for !isZL(s.peek(0)) {
			s.forward(1)
		}
	}
	ch := s.peek(0)
	if !isZL(ch) {
		scannerError("while scanning a block scalar", start,
			fmt.Sprintf("expected a comment or a line break, but found %s", rch(ch)), s.mark())
	}
	s.scanLineBreak()
}

func (s *scanner) scanBlockScalarIndentation() ([]string, int, *Mark) {
	var chunks []string
	maxIndent := 0
	end := s.mark()
	for s.peek(0) == ' ' || isL(s.peek(0)) {
		if s.peek(0) != ' ' {
			chunks = append(chunks, s.scanLineBreak())
			end = s.mark()
		} else {
			s.forward(1)
			if s.column > maxIndent {
				maxIndent = s.column
			}
		}
	}
	return chunks, maxIndent, end
}

func (s *scanner) scanBlockScalarBreaks(indent int) ([]string, *Mark) {
	var chunks []string
	end := s.mark()
	for s.column < indent && s.peek(0) == ' ' {
		s.forward(1)
	}
	for isL(s.peek(0)) {
		chunks = append(chunks, s.scanLineBreak())
		end = s.mark()
		for s.column < indent && s.peek(0) == ' ' {
			s.forward(1)
		}
	}
	return chunks, end
}

func (s *scanner) scanFlowScalar(style rune) *Token {
	double := style == '"'
	var chunks []byte
	start := s.mark()
	quote := s.peek(0)
	s.forward(1)
	chunks = s.scanFlowScalarNonSpaces(chunks, double, start)
	for s.peek(0) != quote {
		chunks = s.scanFlowScalarSpaces(chunks, double, start)
		chunks = s.scanFlowScalarNonSpaces(chunks, double, start)
	}
	s.forward(1)
	t := newToken(ScalarToken, start, s.mark())
	t.Value = string(chunks)
	t.Style = string(style)
	return t
}

var escapeReplacements = map[rune]rune{
	'0': 0, 'a': 0x07, 'b': 0x08, 't': 0x09, '\t': 0x09, 'n': 0x0A, 'v': 0x0B, 'f': 0x0C, 'r': 0x0D,
	'e': 0x1B, ' ': 0x20, '"': '"', '\\': '\\', '/': '/', 'N': 0x85, '_': 0xA0, 'L': 0x2028, 'P': 0x2029,
}

var escapeCodes = map[rune]int{'x': 2, 'u': 4, 'U': 8}

func (s *scanner) scanFlowScalarNonSpaces(chunks []byte, double bool, start *Mark) []byte {
	for {
		length := 0
		for {
			c := s.peek(length)
			if isZBLT(c) || c == '\'' || c == '"' || c == '\\' {
				break
			}
			length++
		}
		if length > 0 {
			chunks = append(chunks, s.prefix(length)...)
			s.forward(length)
		}
		ch := s.peek(0)
		switch {
		case !double && ch == '\'' && s.peek(1) == '\'':
			chunks = append(chunks, '\'')
			s.forward(2)
		case (double && ch == '\'') || (!double && (ch == '"' || ch == '\\')):
			chunks = pyval.AppendRune(chunks, ch)
			s.forward(1)
		case double && ch == '\\':
			s.forward(1)
			ch = s.peek(0)
			if r, ok := escapeReplacements[ch]; ok {
				chunks = pyval.AppendRune(chunks, r)
				s.forward(1)
			} else if n, ok := escapeCodes[ch]; ok {
				s.forward(1)
				for k := 0; k < n; k++ {
					if !isHex(s.peek(k)) {
						scannerError("while scanning a double-quoted scalar", start,
							fmt.Sprintf("expected escape sequence of %d hexadecimal numbers, but found %s", n, rch(s.peek(k))), s.mark())
					}
				}
				code, _ := new(big.Int).SetString(s.prefix(n), 16)
				switch {
				case code.Cmp(big.NewInt(0x7FFFFFFF)) > 0:
					raise(&PyError{Class: "OverflowError", Msg: "Python int too large to convert to C int"})
				case code.Cmp(big.NewInt(0x10FFFF)) > 0:
					raise(&PyError{Class: "ValueError", Msg: "chr() arg not in range(0x110000)"})
				}
				chunks = pyval.AppendRune(chunks, rune(code.Int64()))
				s.forward(n)
			} else if isL(ch) {
				s.scanLineBreak()
				chunks = s.scanFlowScalarBreaks(chunks, start)
			} else {
				scannerError("while scanning a double-quoted scalar", start,
					fmt.Sprintf("found unknown escape character %s", rch(ch)), s.mark())
			}
		default:
			return chunks
		}
	}
}

func (s *scanner) scanFlowScalarSpaces(chunks []byte, double bool, start *Mark) []byte {
	length := 0
	for s.peek(length) == ' ' || s.peek(length) == '\t' {
		length++
	}
	whitespaces := s.prefix(length)
	s.forward(length)
	ch := s.peek(0)
	if ch == 0 {
		scannerError("while scanning a quoted scalar", start, "found unexpected end of stream", s.mark())
	} else if isL(ch) {
		lineBreak := s.scanLineBreak()
		breaks := s.scanFlowScalarBreaks(nil, start)
		if lineBreak != "\n" {
			chunks = append(chunks, lineBreak...)
		} else if len(breaks) == 0 {
			chunks = append(chunks, ' ')
		}
		chunks = append(chunks, breaks...)
	} else {
		chunks = append(chunks, whitespaces...)
	}
	return chunks
}

func (s *scanner) scanFlowScalarBreaks(chunks []byte, start *Mark) []byte {
	for {
		prefix := s.prefix(3)
		if (prefix == "---" || prefix == "...") && isZBLT(s.peek(3)) {
			scannerError("while scanning a quoted scalar", start, "found unexpected document separator", s.mark())
		}
		for s.peek(0) == ' ' || s.peek(0) == '\t' {
			s.forward(1)
		}
		if isL(s.peek(0)) {
			chunks = append(chunks, s.scanLineBreak()...)
		} else {
			return chunks
		}
	}
}

func (s *scanner) scanPlain() *Token {
	var chunks []byte
	start := s.mark()
	end := start
	indent := s.indent + 1
	var spacesChunks []string
	for {
		length := 0
		if s.peek(0) == '#' {
			break
		}
		for {
			ch := s.peek(length)
			if isZBLT(ch) {
				break
			}
			if ch == ':' {
				next := s.peek(length + 1)
				if isZBLT(next) || (s.flowLevel != 0 && isIn(next, ",[]{}")) {
					break
				}
			}
			if s.flowLevel != 0 && isIn(ch, ",?[]{}") {
				break
			}
			length++
		}
		if length == 0 {
			break
		}
		s.allowSimpleKey = false
		for _, c := range spacesChunks {
			chunks = append(chunks, c...)
		}
		chunks = append(chunks, s.prefix(length)...)
		s.forward(length)
		end = s.mark()
		var ok bool
		spacesChunks, ok = s.scanPlainSpaces()
		if !ok || len(spacesChunks) == 0 || s.peek(0) == '#' || (s.flowLevel == 0 && s.column < indent) {
			break
		}
	}
	t := newToken(ScalarToken, start, end)
	t.Value = string(chunks)
	t.Plain = true
	return t
}

// scanPlainSpaces returns ok false where Python returns None.
func (s *scanner) scanPlainSpaces() ([]string, bool) {
	var chunks []string
	length := 0
	for s.peek(length) == ' ' {
		length++
	}
	whitespaces := s.prefix(length)
	s.forward(length)
	ch := s.peek(0)
	if isL(ch) {
		lineBreak := s.scanLineBreak()
		s.allowSimpleKey = true
		prefix := s.prefix(3)
		if (prefix == "---" || prefix == "...") && isZBLT(s.peek(3)) {
			return nil, false
		}
		var breaks []string
		for s.peek(0) == ' ' || isL(s.peek(0)) {
			if s.peek(0) == ' ' {
				s.forward(1)
			} else {
				breaks = append(breaks, s.scanLineBreak())
				prefix = s.prefix(3)
				if (prefix == "---" || prefix == "...") && isZBLT(s.peek(3)) {
					return nil, false
				}
			}
		}
		if lineBreak != "\n" {
			chunks = append(chunks, lineBreak)
		} else if len(breaks) == 0 {
			chunks = append(chunks, " ")
		}
		chunks = append(chunks, breaks...)
	} else if whitespaces != "" {
		chunks = append(chunks, whitespaces)
	}
	return chunks, true
}

func (s *scanner) scanTagHandle(name string, start *Mark) string {
	ch := s.peek(0)
	if ch != '!' {
		scannerError("while scanning a "+name, start, fmt.Sprintf("expected '!', but found %s", rch(ch)), s.mark())
	}
	length := 1
	ch = s.peek(length)
	if ch != ' ' {
		for isWordChar(ch) {
			length++
			ch = s.peek(length)
		}
		if ch != '!' {
			s.forward(length)
			scannerError("while scanning a "+name, start, fmt.Sprintf("expected '!', but found %s", rch(ch)), s.mark())
		}
		length++
	}
	value := s.prefix(length)
	s.forward(length)
	return value
}

func (s *scanner) scanTagURI(name string, start *Mark) string {
	var chunks []byte
	any := false
	length := 0
	ch := s.peek(length)
	for (ch >= '0' && ch <= '9') || (ch >= 'A' && ch <= 'Z') || (ch >= 'a' && ch <= 'z') ||
		isIn(ch, "-;/?:@&=+$,_.!~*'()[]%") {
		if ch == '%' {
			chunks = append(chunks, s.prefix(length)...)
			any = true
			s.forward(length)
			length = 0
			chunks = append(chunks, s.scanURIEscapes(name, start)...)
		} else {
			length++
		}
		ch = s.peek(length)
	}
	if length > 0 {
		chunks = append(chunks, s.prefix(length)...)
		any = true
		s.forward(length)
	}
	if !any {
		scannerError("while parsing a "+name, start, fmt.Sprintf("expected URI, but found %s", rch(ch)), s.mark())
	}
	return string(chunks)
}

func (s *scanner) scanURIEscapes(name string, start *Mark) string {
	var codes []byte
	mark := s.mark()
	for s.peek(0) == '%' {
		s.forward(1)
		for k := 0; k < 2; k++ {
			if !isHex(s.peek(k)) {
				scannerError("while scanning a "+name, start,
					fmt.Sprintf("expected URI escape sequence of 2 hexadecimal numbers, but found %s", rch(s.peek(k))), s.mark())
			}
		}
		code, _ := new(big.Int).SetString(s.prefix(2), 16)
		codes = append(codes, byte(code.Int64()))
		s.forward(2)
	}
	rs, derr := pyval.DecodeUTF8(codes)
	if derr != nil {
		scannerError("while scanning a "+name, start, derr.Error(), mark)
	}
	return pyval.FromRunes(rs)
}

func (s *scanner) scanLineBreak() string {
	ch := s.peek(0)
	switch ch {
	case '\r', '\n', 0x85:
		if s.prefix(2) == "\r\n" {
			s.forward(2)
		} else {
			s.forward(1)
		}
		return "\n"
	case 0x2028, 0x2029:
		s.forward(1)
		return string(ch)
	}
	return ""
}
