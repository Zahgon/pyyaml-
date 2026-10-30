// Port of PyYAML's lib/yaml/emitter.py (MIT License; Copyright (c) 2017-2021 Ingy döt Net, Copyright
// (c) 2006-2016 Kirill Simonov; see LICENSE).

package yaml

import (
	"fmt"
	"math/big"
	"sort"
	"strconv"
	"strings"

	"github.com/yaml/pyyaml-go/internal/pyval"
)

type scalarAnalysis struct {
	scalar            []rune
	empty             bool
	multiline         bool
	allowFlowPlain    bool
	allowBlockPlain   bool
	allowSingleQuoted bool
	allowDoubleQuoted bool
	allowBlock        bool
}

// noChar stands for Python's None where the writers look at the character after the text.
const noChar rune = -1

// emitter writes events as text, following emitter.py's state machine.
type emitter struct {
	out          []byte
	binary       bool
	encoding     string
	states       []func()
	state        func()
	events       []*Event
	event        *Event
	indents      []*int
	indent       *int
	flowLevel    int
	rootContext  bool
	seqContext   bool
	mapContext   bool
	simpleKeyCtx bool
	line         int
	column       int
	whitespace   bool
	indention    bool
	openEnded    bool
	canonical    bool
	allowUnicode bool
	bestIndent   int
	bestWidth    int
	bestBreak    string
	tagPrefixes  map[string]string
	prepAnchor   *string
	prepTag      *string
	analysis     *scalarAnalysis
	style        *string
}

func newEmitter(opts *DumpOptions, binary bool) *emitter {
	e := &emitter{binary: binary, whitespace: true, indention: true, canonical: opts.Canonical,
		allowUnicode: opts.AllowUnicode, bestIndent: 2, bestWidth: 80, bestBreak: "\n"}
	e.state = e.expectStreamStart
	if opts.Indent != 0 && opts.Indent > 1 && opts.Indent < 10 {
		e.bestIndent = opts.Indent
	}
	if opts.Width != 0 && opts.Width > e.bestIndent*2 {
		e.bestWidth = opts.Width
	}
	switch opts.LineBreak {
	case "\r", "\n", "\r\n":
		e.bestBreak = opts.LineBreak
	}
	return e
}

func emitterError(format string, args ...any) {
	raise(&YAMLError{Class: "EmitterError", Msg: fmt.Sprintf(format, args...)})
}

func eventRepr(ev *Event) string {
	s, _ := Repr(ev)
	return s
}

func (e *emitter) popState() {
	e.state = e.states[len(e.states)-1]
	e.states = e.states[:len(e.states)-1]
}

func (e *emitter) popIndent() {
	e.indent = e.indents[len(e.indents)-1]
	e.indents = e.indents[:len(e.indents)-1]
}

func (e *emitter) emit(event *Event) {
	e.events = append(e.events, event)
	for !e.needMoreEvents() {
		e.event = e.events[0]
		e.events = e.events[1:]
		e.state()
		e.event = nil
	}
}

func (e *emitter) needMoreEvents() bool {
	if len(e.events) == 0 {
		return true
	}
	switch e.events[0].Kind {
	case DocumentStartEvent:
		return e.needEvents(1)
	case SequenceStartEvent:
		return e.needEvents(2)
	case MappingStartEvent:
		return e.needEvents(3)
	}
	return false
}

func (e *emitter) needEvents(count int) bool {
	level := 0
	for _, event := range e.events[1:] {
		switch {
		case event.Kind == DocumentStartEvent || isCollectionStart(event):
			level++
		case event.Kind == DocumentEndEvent || isCollectionEnd(event):
			level--
		case event.Kind == StreamEndEvent:
			level = -1
		}
		if level < 0 {
			return false
		}
	}
	return len(e.events) < count+1
}

func (e *emitter) increaseIndent(flow, indentless bool) {
	e.indents = append(e.indents, e.indent)
	if e.indent == nil {
		v := 0
		if flow {
			v = e.bestIndent
		}
		e.indent = &v
	} else if !indentless {
		v := *e.indent + e.bestIndent
		e.indent = &v
	}
}

func (e *emitter) expectStreamStart() {
	if e.event.Kind != StreamStartEvent {
		emitterError("expected StreamStartEvent, but got %s", eventRepr(e.event))
	}
	if e.event.Encoding != nil && *e.event.Encoding != "" && e.binary {
		e.encoding = *e.event.Encoding
	}
	e.writeStreamStart()
	e.state = e.expectFirstDocumentStart
}

func (e *emitter) expectNothing() { emitterError("expected nothing, but got %s", eventRepr(e.event)) }

func (e *emitter) expectFirstDocumentStart() { e.expectDocumentStartAt(true) }
func (e *emitter) expectDocumentStart()      { e.expectDocumentStartAt(false) }

func (e *emitter) expectDocumentStartAt(first bool) {
	switch e.event.Kind {
	case DocumentStartEvent:
		hasVersion := e.event.Version != nil && len(e.event.Version.Items) > 0
		if (hasVersion || len(e.event.Tags) > 0) && e.openEnded {
			e.writeIndicator("...", true, false, false)
			e.writeIndent()
		}
		if hasVersion {
			e.writeVersionDirective(e.prepareVersion(e.event.Version))
		}
		e.tagPrefixes = map[string]string{"!": "!", "tag:yaml.org,2002:": "!!"}
		if len(e.event.Tags) > 0 {
			handles := make([]string, 0, len(e.event.Tags))
			for h := range e.event.Tags {
				handles = append(handles, h)
			}
			sort.Strings(handles)
			for _, handle := range handles {
				prefix := e.event.Tags[handle]
				e.tagPrefixes[prefix] = handle
				handleText := e.prepareTagHandle(handle)
				prefixText := e.prepareTagPrefix(prefix)
				e.writeTagDirective(handleText, prefixText)
			}
		}
		implicit := first && !e.event.Explicit && !e.canonical && !hasVersion && len(e.event.Tags) == 0 && !e.checkEmptyDocument()
		if !implicit {
			e.writeIndent()
			e.writeIndicator("---", true, false, false)
			if e.canonical {
				e.writeIndent()
			}
		}
		e.state = e.expectDocumentRoot
	case StreamEndEvent:
		if e.openEnded {
			e.writeIndicator("...", true, false, false)
			e.writeIndent()
		}
		e.state = e.expectNothing
	default:
		emitterError("expected DocumentStartEvent, but got %s", eventRepr(e.event))
	}
}

func (e *emitter) expectDocumentEnd() {
	if e.event.Kind != DocumentEndEvent {
		emitterError("expected DocumentEndEvent, but got %s", eventRepr(e.event))
	}
	e.writeIndent()
	if e.event.Explicit {
		e.writeIndicator("...", true, false, false)
		e.writeIndent()
	}
	e.state = e.expectDocumentStart
}

func (e *emitter) expectDocumentRoot() {
	e.states = append(e.states, e.expectDocumentEnd)
	e.expectNode(true, false, false, false)
}

func (e *emitter) expectNode(root, sequence, mapping, simpleKey bool) {
	e.rootContext, e.seqContext, e.mapContext, e.simpleKeyCtx = root, sequence, mapping, simpleKey
	switch {
	case e.event.Kind == AliasEvent:
		e.expectAlias()
	case e.event.Kind == ScalarEvent || isCollectionStart(e.event):
		e.processAnchor("&")
		e.processTag()
		flow := e.event.FlowStyle != nil && *e.event.FlowStyle
		switch e.event.Kind {
		case ScalarEvent:
			e.expectScalar()
		case SequenceStartEvent:
			if e.flowLevel > 0 || e.canonical || flow || e.checkEmptySequence() {
				e.expectFlowSequence()
			} else {
				e.expectBlockSequence()
			}
		case MappingStartEvent:
			if e.flowLevel > 0 || e.canonical || flow || e.checkEmptyMapping() {
				e.expectFlowMapping()
			} else {
				e.expectBlockMapping()
			}
		}
	default:
		emitterError("expected NodeEvent, but got %s", eventRepr(e.event))
	}
}

func (e *emitter) expectAlias() {
	if e.event.Anchor == nil {
		emitterError("anchor is not specified for alias")
	}
	e.processAnchor("*")
	e.popState()
}

func (e *emitter) expectScalar() {
	e.increaseIndent(true, false)
	e.processScalar()
	e.popIndent()
	e.popState()
}

func (e *emitter) expectFlowSequence() {
	e.writeIndicator("[", true, true, false)
	e.flowLevel++
	e.increaseIndent(true, false)
	e.state = e.expectFirstFlowSequenceItem
}

func (e *emitter) expectFirstFlowSequenceItem() {
	if e.event.Kind == SequenceEndEvent {
		e.popIndent()
		e.flowLevel--
		e.writeIndicator("]", false, false, false)
		e.popState()
	} else {
		if e.canonical || e.column > e.bestWidth {
			e.writeIndent()
		}
		e.states = append(e.states, e.expectFlowSequenceItem)
		e.expectNode(false, true, false, false)
	}
}

func (e *emitter) expectFlowSequenceItem() {
	if e.event.Kind == SequenceEndEvent {
		e.popIndent()
		e.flowLevel--
		if e.canonical {
			e.writeIndicator(",", false, false, false)
			e.writeIndent()
		}
		e.writeIndicator("]", false, false, false)
		e.popState()
	} else {
		e.writeIndicator(",", false, false, false)
		if e.canonical || e.column > e.bestWidth {
			e.writeIndent()
		}
		e.states = append(e.states, e.expectFlowSequenceItem)
		e.expectNode(false, true, false, false)
	}
}

func (e *emitter) expectFlowMapping() {
	e.writeIndicator("{", true, true, false)
	e.flowLevel++
	e.increaseIndent(true, false)
	e.state = e.expectFirstFlowMappingKey
}

func (e *emitter) expectFirstFlowMappingKey() {
	if e.event.Kind == MappingEndEvent {
		e.popIndent()
		e.flowLevel--
		e.writeIndicator("}", false, false, false)
		e.popState()
	} else {
		if e.canonical || e.column > e.bestWidth {
			e.writeIndent()
		}
		if !e.canonical && e.checkSimpleKey() {
			e.states = append(e.states, e.expectFlowMappingSimpleValue)
			e.expectNode(false, false, true, true)
		} else {
			e.writeIndicator("?", true, false, false)
			e.states = append(e.states, e.expectFlowMappingValue)
			e.expectNode(false, false, true, false)
		}
	}
}

func (e *emitter) expectFlowMappingKey() {
	if e.event.Kind == MappingEndEvent {
		e.popIndent()
		e.flowLevel--
		if e.canonical {
			e.writeIndicator(",", false, false, false)
			e.writeIndent()
		}
		e.writeIndicator("}", false, false, false)
		e.popState()
	} else {
		e.writeIndicator(",", false, false, false)
		if e.canonical || e.column > e.bestWidth {
			e.writeIndent()
		}
		if !e.canonical && e.checkSimpleKey() {
			e.states = append(e.states, e.expectFlowMappingSimpleValue)
			e.expectNode(false, false, true, true)
		} else {
			e.writeIndicator("?", true, false, false)
			e.states = append(e.states, e.expectFlowMappingValue)
			e.expectNode(false, false, true, false)
		}
	}
}

func (e *emitter) expectFlowMappingSimpleValue() {
	e.writeIndicator(":", false, false, false)
	e.states = append(e.states, e.expectFlowMappingKey)
	e.expectNode(false, false, true, false)
}

func (e *emitter) expectFlowMappingValue() {
	if e.canonical || e.column > e.bestWidth {
		e.writeIndent()
	}
	e.writeIndicator(":", true, false, false)
	e.states = append(e.states, e.expectFlowMappingKey)
	e.expectNode(false, false, true, false)
}

func (e *emitter) expectBlockSequence() {
	indentless := e.mapContext && !e.indention
	e.increaseIndent(false, indentless)
	e.state = e.expectFirstBlockSequenceItem
}

func (e *emitter) expectFirstBlockSequenceItem() { e.expectBlockSequenceItemAt(true) }
func (e *emitter) expectBlockSequenceItem()      { e.expectBlockSequenceItemAt(false) }

func (e *emitter) expectBlockSequenceItemAt(first bool) {
	if !first && e.event.Kind == SequenceEndEvent {
		e.popIndent()
		e.popState()
	} else {
		e.writeIndent()
		e.writeIndicator("-", true, false, true)
		e.states = append(e.states, e.expectBlockSequenceItem)
		e.expectNode(false, true, false, false)
	}
}

func (e *emitter) expectBlockMapping() {
	e.increaseIndent(false, false)
	e.state = e.expectFirstBlockMappingKey
}

func (e *emitter) expectFirstBlockMappingKey() { e.expectBlockMappingKeyAt(true) }
func (e *emitter) expectBlockMappingKey()      { e.expectBlockMappingKeyAt(false) }

func (e *emitter) expectBlockMappingKeyAt(first bool) {
	if !first && e.event.Kind == MappingEndEvent {
		e.popIndent()
		e.popState()
	} else {
		e.writeIndent()
		if e.checkSimpleKey() {
			e.states = append(e.states, e.expectBlockMappingSimpleValue)
			e.expectNode(false, false, true, true)
		} else {
			e.writeIndicator("?", true, false, true)
			e.states = append(e.states, e.expectBlockMappingValue)
			e.expectNode(false, false, true, false)
		}
	}
}

func (e *emitter) expectBlockMappingSimpleValue() {
	e.writeIndicator(":", false, false, false)
	e.states = append(e.states, e.expectBlockMappingKey)
	e.expectNode(false, false, true, false)
}

func (e *emitter) expectBlockMappingValue() {
	e.writeIndent()
	e.writeIndicator(":", true, false, true)
	e.states = append(e.states, e.expectBlockMappingKey)
	e.expectNode(false, false, true, false)
}

func (e *emitter) checkEmptySequence() bool {
	return e.event.Kind == SequenceStartEvent && len(e.events) > 0 && e.events[0].Kind == SequenceEndEvent
}

func (e *emitter) checkEmptyMapping() bool {
	return e.event.Kind == MappingStartEvent && len(e.events) > 0 && e.events[0].Kind == MappingEndEvent
}

func (e *emitter) checkEmptyDocument() bool {
	if e.event.Kind != DocumentStartEvent || len(e.events) == 0 {
		return false
	}
	event := e.events[0]
	return event.Kind == ScalarEvent && event.Anchor == nil && event.Tag == nil && event.Value == ""
}

func (e *emitter) checkSimpleKey() bool {
	length := 0
	if isNodeEvent(e.event) && e.event.Anchor != nil {
		if e.prepAnchor == nil {
			v := e.prepareAnchor(*e.event.Anchor)
			e.prepAnchor = &v
		}
		length += pyval.Len(*e.prepAnchor)
	}
	if (e.event.Kind == ScalarEvent || isCollectionStart(e.event)) && e.event.Tag != nil {
		if e.prepTag == nil {
			v := e.prepareTag(*e.event.Tag)
			e.prepTag = &v
		}
		length += pyval.Len(*e.prepTag)
	}
	if e.event.Kind == ScalarEvent {
		if e.analysis == nil {
			e.analysis = e.analyzeScalar(pyval.Runes(e.event.Value))
		}
		length += len(e.analysis.scalar)
	}
	return length < 128 && (e.event.Kind == AliasEvent ||
		(e.event.Kind == ScalarEvent && !e.analysis.empty && !e.analysis.multiline) ||
		e.checkEmptySequence() || e.checkEmptyMapping())
}

func (e *emitter) processAnchor(indicator string) {
	if e.event.Anchor == nil {
		e.prepAnchor = nil
		return
	}
	if e.prepAnchor == nil {
		v := e.prepareAnchor(*e.event.Anchor)
		e.prepAnchor = &v
	}
	if *e.prepAnchor != "" {
		e.writeIndicator(indicator+*e.prepAnchor, true, false, false)
	}
	e.prepAnchor = nil
}

func (e *emitter) processTag() {
	tag := e.event.Tag
	if e.event.Kind == ScalarEvent {
		if e.style == nil {
			v := e.chooseScalarStyle()
			e.style = &v
		}
		if (!e.canonical || tag == nil) &&
			((*e.style == "" && e.event.Implicit[0]) || (*e.style != "" && e.event.Implicit[1])) {
			e.prepTag = nil
			return
		}
		if e.event.Implicit[0] && tag == nil {
			tag = ptr("!")
			e.prepTag = nil
		}
	} else if (!e.canonical || tag == nil) && e.event.Implicit[0] {
		e.prepTag = nil
		return
	}
	if tag == nil {
		emitterError("tag is not specified")
	}
	if e.prepTag == nil {
		v := e.prepareTag(*tag)
		e.prepTag = &v
	}
	if *e.prepTag != "" {
		e.writeIndicator(*e.prepTag, true, false, false)
	}
	e.prepTag = nil
}

func (e *emitter) chooseScalarStyle() string {
	if e.analysis == nil {
		e.analysis = e.analyzeScalar(pyval.Runes(e.event.Value))
	}
	style := e.event.Style
	if style == "\"" || e.canonical {
		return "\""
	}
	if style == "" && e.event.Implicit[0] {
		if !(e.simpleKeyCtx && (e.analysis.empty || e.analysis.multiline)) &&
			((e.flowLevel > 0 && e.analysis.allowFlowPlain) || (e.flowLevel == 0 && e.analysis.allowBlockPlain)) {
			return ""
		}
	}
	if style != "" && strings.Contains("|>", style) {
		if e.flowLevel == 0 && !e.simpleKeyCtx && e.analysis.allowBlock {
			return style
		}
	}
	if style == "" || style == "'" {
		if e.analysis.allowSingleQuoted && !(e.simpleKeyCtx && e.analysis.multiline) {
			return "'"
		}
	}
	return "\""
}

func (e *emitter) processScalar() {
	if e.analysis == nil {
		e.analysis = e.analyzeScalar(pyval.Runes(e.event.Value))
	}
	if e.style == nil {
		v := e.chooseScalarStyle()
		e.style = &v
	}
	split := !e.simpleKeyCtx
	switch *e.style {
	case "\"":
		e.writeDoubleQuoted(e.analysis.scalar, split)
	case "'":
		e.writeSingleQuoted(e.analysis.scalar, split)
	case ">":
		e.writeFolded(e.analysis.scalar)
	case "|":
		e.writeLiteral(e.analysis.scalar)
	default:
		e.writePlain(e.analysis.scalar, split)
	}
	e.analysis = nil
	e.style = nil
}

func (e *emitter) prepareVersion(version *Tuple) string {
	major, minor := version.Items[0].(*big.Int), version.Items[1].(*big.Int)
	if major.Cmp(big.NewInt(1)) != 0 {
		emitterError("unsupported YAML version: %s.%s", major, minor)
	}
	return major.String() + "." + minor.String()
}

func isAlnumWord(ch rune) bool {
	return (ch >= '0' && ch <= '9') || (ch >= 'A' && ch <= 'Z') || (ch >= 'a' && ch <= 'z')
}

func (e *emitter) prepareTagHandle(handle string) string {
	if handle == "" {
		emitterError("tag handle must not be empty")
	}
	rs := pyval.Runes(handle)
	if rs[0] != '!' || rs[len(rs)-1] != '!' {
		emitterError("tag handle must start and end with '!': %s", pyval.StrRepr(handle))
	}
	if len(rs) > 1 {
		for _, ch := range rs[1 : len(rs)-1] {
			if !isAlnumWord(ch) && ch != '-' && ch != '_' {
				emitterError("invalid character %s in the tag handle: %s", rch(ch), pyval.StrRepr(handle))
			}
		}
	}
	return handle
}

func (e *emitter) prepareTagPrefix(prefix string) string {
	if prefix == "" {
		emitterError("tag prefix must not be empty")
	}
	rs := pyval.Runes(prefix)
	var chunks strings.Builder
	start, end := 0, 0
	if rs[0] == '!' {
		end = 1
	}
	for end < len(rs) {
		ch := rs[end]
		if isAlnumWord(ch) || isIn(ch, "-;/?!:@&=+$,_.~*'()[]") {
			end++
		} else {
			if start < end {
				chunks.WriteString(pyval.FromRunes(rs[start:end]))
			}
			start = end + 1
			end = start
			if _, encErr := pyval.Encode(pyval.FromRunes([]rune{ch}), "utf-8"); encErr != nil {
				raise(&PyError{Class: "UnicodeEncodeError", Msg: encErr.Msg})
			}
			raise(&PyError{Class: "TypeError", Msg: "ord() expected string of length 1, but int found"})
		}
	}
	if start < end {
		chunks.WriteString(pyval.FromRunes(rs[start:end]))
	}
	return chunks.String()
}

func (e *emitter) prepareTag(tag string) string {
	if tag == "" {
		emitterError("tag must not be empty")
	}
	if tag == "!" {
		return tag
	}
	handle := ""
	hasHandle := false
	suffix := tag
	prefixes := make([]string, 0, len(e.tagPrefixes))
	for p := range e.tagPrefixes {
		prefixes = append(prefixes, p)
	}
	sort.Strings(prefixes)
	for _, prefix := range prefixes {
		if strings.HasPrefix(tag, prefix) && (prefix == "!" || len(pyval.Runes(prefix)) < len(pyval.Runes(tag))) {
			handle = e.tagPrefixes[prefix]
			hasHandle = true
			suffix = tag[len(prefix):]
		}
	}
	rs := pyval.Runes(suffix)
	var chunks strings.Builder
	start, end := 0, 0
	for end < len(rs) {
		ch := rs[end]
		if isAlnumWord(ch) || isIn(ch, "-;/?:@&=+$,_.~*'()[]") || (ch == '!' && handle != "!") {
			end++
		} else {
			if start < end {
				chunks.WriteString(pyval.FromRunes(rs[start:end]))
			}
			start = end + 1
			end = start
			data, encErr := pyval.Encode(pyval.FromRunes([]rune{ch}), "utf-8")
			if encErr != nil {
				raise(&PyError{Class: "UnicodeEncodeError", Msg: encErr.Msg})
			}
			for _, c := range data {
				fmt.Fprintf(&chunks, "%%%02X", c)
			}
		}
	}
	if start < end {
		chunks.WriteString(pyval.FromRunes(rs[start:end]))
	}
	if hasHandle && handle != "" {
		return handle + chunks.String()
	}
	return "!<" + chunks.String() + ">"
}

func (e *emitter) prepareAnchor(anchor string) string {
	if anchor == "" {
		emitterError("anchor must not be empty")
	}
	for _, ch := range pyval.Runes(anchor) {
		if !isAlnumWord(ch) && ch != '-' && ch != '_' {
			emitterError("invalid character %s in the anchor: %s", rch(ch), pyval.StrRepr(anchor))
		}
	}
	return anchor
}

func isBreakChar(ch rune) bool { return ch == '\n' || ch == 0x85 || ch == 0x2028 || ch == 0x2029 }

func (e *emitter) analyzeScalar(scalar []rune) *scalarAnalysis {
	if len(scalar) == 0 {
		return &scalarAnalysis{scalar: scalar, empty: true, allowBlockPlain: true, allowSingleQuoted: true, allowDoubleQuoted: true}
	}
	var blockIndicators, flowIndicators, lineBreaks, specialCharacters bool
	var leadingSpace, leadingBreak, trailingSpace, trailingBreak, breakSpace, spaceBreak bool
	if len(scalar) >= 3 && ((scalar[0] == '-' && scalar[1] == '-' && scalar[2] == '-') ||
		(scalar[0] == '.' && scalar[1] == '.' && scalar[2] == '.')) {
		blockIndicators = true
		flowIndicators = true
	}
	precededByWhitespace := true
	followedByWhitespace := len(scalar) == 1 || isZBLT(scalar[1])
	previousSpace, previousBreak := false, false
	for index := 0; index < len(scalar); {
		ch := scalar[index]
		if index == 0 {
			if isIn(ch, "#,[]{}&*!|>'\"%@`") {
				flowIndicators = true
				blockIndicators = true
			}
			if ch == '?' || ch == ':' {
				flowIndicators = true
				if followedByWhitespace {
					blockIndicators = true
				}
			}
			if ch == '-' && followedByWhitespace {
				flowIndicators = true
				blockIndicators = true
			}
		} else {
			if isIn(ch, ",?[]{}") {
				flowIndicators = true
			}
			if ch == ':' {
				flowIndicators = true
				if followedByWhitespace {
					blockIndicators = true
				}
			}
			if ch == '#' && precededByWhitespace {
				flowIndicators = true
				blockIndicators = true
			}
		}
		if isBreakChar(ch) {
			lineBreaks = true
		}
		if !(ch == '\n' || (ch >= 0x20 && ch <= 0x7E)) {
			if (ch == 0x85 || (ch >= 0xA0 && ch <= 0xD7FF) || (ch >= 0xE000 && ch <= 0xFFFD) ||
				(ch >= 0x10000 && ch < 0x10FFFF)) && ch != 0xFEFF {
				if !e.allowUnicode {
					specialCharacters = true
				}
			} else {
				specialCharacters = true
			}
		}
		if ch == ' ' {
			if index == 0 {
				leadingSpace = true
			}
			if index == len(scalar)-1 {
				trailingSpace = true
			}
			if previousBreak {
				breakSpace = true
			}
			previousSpace = true
			previousBreak = false
		} else if isBreakChar(ch) {
			if index == 0 {
				leadingBreak = true
			}
			if index == len(scalar)-1 {
				trailingBreak = true
			}
			if previousSpace {
				spaceBreak = true
			}
			previousSpace = false
			previousBreak = true
		} else {
			previousSpace = false
			previousBreak = false
		}
		index++
		precededByWhitespace = isZBLT(ch)
		followedByWhitespace = index+1 >= len(scalar) || isZBLT(scalar[index+1])
	}
	a := &scalarAnalysis{scalar: scalar, allowFlowPlain: true, allowBlockPlain: true, allowSingleQuoted: true,
		allowDoubleQuoted: true, allowBlock: true, multiline: lineBreaks}
	if leadingSpace || leadingBreak || trailingSpace || trailingBreak {
		a.allowFlowPlain = false
		a.allowBlockPlain = false
	}
	if trailingSpace {
		a.allowBlock = false
	}
	if breakSpace {
		a.allowFlowPlain = false
		a.allowBlockPlain = false
		a.allowSingleQuoted = false
	}
	if spaceBreak || specialCharacters {
		a.allowFlowPlain = false
		a.allowBlockPlain = false
		a.allowSingleQuoted = false
		a.allowBlock = false
	}
	if lineBreaks {
		a.allowFlowPlain = false
		a.allowBlockPlain = false
	}
	if flowIndicators {
		a.allowFlowPlain = false
	}
	if blockIndicators {
		a.allowBlockPlain = false
	}
	return a
}

func (e *emitter) write(data string) {
	if e.encoding != "" {
		encoded, encErr := pyval.Encode(data, e.encoding)
		if encErr != nil {
			raise(&PyError{Class: "UnicodeEncodeError", Msg: encErr.Msg})
		}
		e.out = append(e.out, encoded...)
		return
	}
	e.out = append(e.out, data...)
}

func (e *emitter) writeRunes(text []rune, count bool) {
	data := pyval.FromRunes(text)
	if count {
		e.column += len(text)
	}
	e.write(data)
}

func (e *emitter) writeStreamStart() {
	if strings.HasPrefix(e.encoding, "utf-16") {
		e.write(string(rune(0xFEFF)))
	}
}

func (e *emitter) writeIndicator(indicator string, needWhitespace, whitespace, indention bool) {
	data := indicator
	if !e.whitespace && needWhitespace {
		data = " " + indicator
	}
	e.whitespace = whitespace
	e.indention = e.indention && indention
	e.column += pyval.Len(data)
	e.openEnded = false
	e.write(data)
}

func (e *emitter) writeIndent() {
	indent := 0
	if e.indent != nil {
		indent = *e.indent
	}
	if !e.indention || e.column > indent || (e.column == indent && !e.whitespace) {
		e.writeLineBreak("")
	}
	if e.column < indent {
		e.whitespace = true
		data := strings.Repeat(" ", indent-e.column)
		e.column = indent
		e.write(data)
	}
}

func (e *emitter) writeLineBreak(data string) {
	if data == "" {
		data = e.bestBreak
	}
	e.whitespace = true
	e.indention = true
	e.line++
	e.column = 0
	e.write(data)
}

func (e *emitter) writeVersionDirective(text string) {
	e.write("%YAML " + text)
	e.writeLineBreak("")
}

func (e *emitter) writeTagDirective(handle, prefix string) {
	e.write("%TAG " + handle + " " + prefix)
	e.writeLineBreak("")
}

func (e *emitter) writeBreaks(text []rune) {
	for _, br := range text {
		if br == '\n' {
			e.writeLineBreak("")
		} else {
			e.writeLineBreak(string(br))
		}
	}
}

func (e *emitter) writeSingleQuoted(text []rune, split bool) {
	e.writeIndicator("'", true, false, false)
	spacesRun, breaks := false, false
	start, end := 0, 0
	for end <= len(text) {
		ch := noChar
		if end < len(text) {
			ch = text[end]
		}
		if spacesRun {
			if ch == noChar || ch != ' ' {
				if start+1 == end && e.column > e.bestWidth && split && start != 0 && end != len(text) {
					e.writeIndent()
				} else {
					e.writeRunes(text[start:end], true)
				}
				start = end
			}
		} else if breaks {
			if ch == noChar || !isBreakChar(ch) {
				if text[start] == '\n' {
					e.writeLineBreak("")
				}
				e.writeBreaks(text[start:end])
				e.writeIndent()
				start = end
			}
		} else {
			if ch == noChar || ch == ' ' || isBreakChar(ch) || ch == '\'' {
				if start < end {
					e.writeRunes(text[start:end], true)
					start = end
				}
			}
		}
		if ch == '\'' {
			e.column += 2
			e.write("''")
			start = end + 1
		}
		if ch != noChar {
			spacesRun = ch == ' '
			breaks = isBreakChar(ch)
		}
		end++
	}
	e.writeIndicator("'", false, false, false)
}

var emitterEscapes = map[rune]string{
	0: "0", 0x07: "a", 0x08: "b", 0x09: "t", 0x0A: "n", 0x0B: "v", 0x0C: "f", 0x0D: "r", 0x1B: "e",
	'"': "\"", '\\': "\\", 0x85: "N", 0xA0: "_", 0x2028: "L", 0x2029: "P",
}

func (e *emitter) writeDoubleQuoted(text []rune, split bool) {
	e.writeIndicator("\"", true, false, false)
	start, end := 0, 0
	for end <= len(text) {
		ch := noChar
		if end < len(text) {
			ch = text[end]
		}
		if ch == noChar || ch == '"' || ch == '\\' || ch == 0x85 || ch == 0x2028 || ch == 0x2029 || ch == 0xFEFF ||
			!((ch >= 0x20 && ch <= 0x7E) || (e.allowUnicode && ((ch >= 0xA0 && ch <= 0xD7FF) || (ch >= 0xE000 && ch <= 0xFFFD)))) {
			if start < end {
				e.writeRunes(text[start:end], true)
				start = end
			}
			if ch != noChar {
				var data string
				if r, ok := emitterEscapes[ch]; ok {
					data = "\\" + r
				} else if ch <= 0xFF {
					data = fmt.Sprintf("\\x%02X", ch)
				} else if ch <= 0xFFFF {
					data = fmt.Sprintf("\\u%04X", ch)
				} else {
					data = fmt.Sprintf("\\U%08X", ch)
				}
				e.column += len(data)
				e.write(data)
				start = end + 1
			}
		}
		if end > 0 && end < len(text)-1 && (ch == ' ' || start >= end) && e.column+(end-start) > e.bestWidth && split {
			var slice []rune
			if start < end {
				slice = text[start:end]
			}
			data := pyval.FromRunes(slice) + "\\"
			if start < end {
				start = end
			}
			e.column += len(slice) + 1
			e.write(data)
			e.writeIndent()
			e.whitespace = false
			e.indention = false
			if text[start] == ' ' {
				e.column++
				e.write("\\")
			}
		}
		end++
	}
	e.writeIndicator("\"", false, false, false)
}

func (e *emitter) determineBlockHints(text []rune) string {
	hints := ""
	if len(text) > 0 {
		if text[0] == ' ' || isBreakChar(text[0]) {
			hints += strconv.Itoa(e.bestIndent)
		}
		if !isBreakChar(text[len(text)-1]) {
			hints += "-"
		} else if len(text) == 1 || isBreakChar(text[len(text)-2]) {
			hints += "+"
		}
	}
	return hints
}

func (e *emitter) writeFolded(text []rune) {
	hints := e.determineBlockHints(text)
	e.writeIndicator(">"+hints, true, false, false)
	if strings.HasSuffix(hints, "+") {
		e.openEnded = true
	}
	e.writeLineBreak("")
	leadingSpace := true
	spacesRun := false
	breaks := true
	start, end := 0, 0
	for end <= len(text) {
		ch := noChar
		if end < len(text) {
			ch = text[end]
		}
		if breaks {
			if ch == noChar || !isBreakChar(ch) {
				if !leadingSpace && ch != noChar && ch != ' ' && text[start] == '\n' {
					e.writeLineBreak("")
				}
				leadingSpace = ch == ' '
				e.writeBreaks(text[start:end])
				if ch != noChar {
					e.writeIndent()
				}
				start = end
			}
		} else if spacesRun {
			if ch != ' ' {
				if start+1 == end && e.column > e.bestWidth {
					e.writeIndent()
				} else {
					e.writeRunes(text[start:end], true)
				}
				start = end
			}
		} else {
			if ch == noChar || ch == ' ' || isBreakChar(ch) {
				e.writeRunes(text[start:end], true)
				if ch == noChar {
					e.writeLineBreak("")
				}
				start = end
			}
		}
		if ch != noChar {
			breaks = isBreakChar(ch)
			spacesRun = ch == ' '
		}
		end++
	}
}

func (e *emitter) writeLiteral(text []rune) {
	hints := e.determineBlockHints(text)
	e.writeIndicator("|"+hints, true, false, false)
	if strings.HasSuffix(hints, "+") {
		e.openEnded = true
	}
	e.writeLineBreak("")
	breaks := true
	start, end := 0, 0
	for end <= len(text) {
		ch := noChar
		if end < len(text) {
			ch = text[end]
		}
		if breaks {
			if ch == noChar || !isBreakChar(ch) {
				e.writeBreaks(text[start:end])
				if ch != noChar {
					e.writeIndent()
				}
				start = end
			}
		} else {
			if ch == noChar || isBreakChar(ch) {
				e.writeRunes(text[start:end], false)
				if ch == noChar {
					e.writeLineBreak("")
				}
				start = end
			}
		}
		if ch != noChar {
			breaks = isBreakChar(ch)
		}
		end++
	}
}

func (e *emitter) writePlain(text []rune, split bool) {
	if e.rootContext {
		e.openEnded = true
	}
	if len(text) == 0 {
		return
	}
	if !e.whitespace {
		e.column++
		e.write(" ")
	}
	e.whitespace = false
	e.indention = false
	spacesRun, breaks := false, false
	start, end := 0, 0
	for end <= len(text) {
		ch := noChar
		if end < len(text) {
			ch = text[end]
		}
		if spacesRun {
			if ch != ' ' {
				if start+1 == end && e.column > e.bestWidth && split {
					e.writeIndent()
					e.whitespace = false
					e.indention = false
				} else {
					e.writeRunes(text[start:end], true)
				}
				start = end
			}
		} else if breaks {
			if !isBreakChar(ch) {
				if text[start] == '\n' {
					e.writeLineBreak("")
				}
				e.writeBreaks(text[start:end])
				e.writeIndent()
				e.whitespace = false
				e.indention = false
				start = end
			}
		} else {
			if ch == noChar || ch == ' ' || isBreakChar(ch) {
				e.writeRunes(text[start:end], true)
				start = end
			}
		}
		if ch != noChar {
			spacesRun = ch == ' '
			breaks = isBreakChar(ch)
		}
		end++
	}
}
