// Port of PyYAML's lib/yaml/parser.py (MIT License; Copyright (c) 2017-2021 Ingy döt Net, Copyright
// (c) 2006-2016 Kirill Simonov; see LICENSE).

package yaml

import (
	"fmt"
	"math/big"

	"github.com/yaml/pyyaml-go/internal/pyval"
)

var defaultTags = map[string]string{"!": "!", "!!": "tag:yaml.org,2002:"}

// parser turns tokens into events, following parser.py's state machine.
type parser struct {
	*scanner
	currentEvent *Event
	yamlVersion  *Tuple
	tagHandles   map[string]string
	states       []func() *Event
	marks        []*Mark
	state        func() *Event
}

func newParser(s *scanner) *parser {
	p := &parser{scanner: s, tagHandles: map[string]string{}}
	p.state = p.parseStreamStart
	return p
}

func parserError(context string, contextMark *Mark, problem string, problemMark *Mark) {
	var c *string
	if context != "" {
		c = &context
	}
	raise(markedError("ParserError", c, contextMark, problem, problemMark))
}

func (p *parser) popState() func() *Event {
	st := p.states[len(p.states)-1]
	p.states = p.states[:len(p.states)-1]
	return st
}

func (p *parser) checkEvent(choices ...string) bool {
	if p.currentEvent == nil && p.state != nil {
		p.currentEvent = p.state()
	}
	if p.currentEvent != nil {
		if len(choices) == 0 {
			return true
		}
		for _, c := range choices {
			if p.currentEvent.Kind == c {
				return true
			}
		}
	}
	return false
}

func (p *parser) peekEvent() *Event {
	if p.currentEvent == nil && p.state != nil {
		p.currentEvent = p.state()
	}
	return p.currentEvent
}

func (p *parser) getEvent() *Event {
	if p.currentEvent == nil && p.state != nil {
		p.currentEvent = p.state()
	}
	value := p.currentEvent
	p.currentEvent = nil
	return value
}

func (p *parser) parseStreamStart() *Event {
	token := p.getToken()
	p.state = p.parseImplicitDocumentStart
	return &Event{Kind: StreamStartEvent, StartMark: token.StartMark, EndMark: token.EndMark, Encoding: token.Encoding}
}

func (p *parser) parseImplicitDocumentStart() *Event {
	if !p.checkToken(DirectiveToken, DocumentStartToken, StreamEndToken) {
		p.tagHandles = defaultTags
		token := p.peekToken()
		p.states = append(p.states, p.parseDocumentEnd)
		p.state = p.parseBlockNode
		return &Event{Kind: DocumentStartEvent, StartMark: token.StartMark, EndMark: token.StartMark}
	}
	return p.parseDocumentStart()
}

func (p *parser) parseDocumentStart() *Event {
	for p.checkToken(DocumentEndToken) {
		p.getToken()
	}
	if !p.checkToken(StreamEndToken) {
		token := p.peekToken()
		start := token.StartMark
		version, tags := p.processDirectives()
		if !p.checkToken(DocumentStartToken) {
			parserError("", nil, "expected '<document start>', but found "+pyval.StrRepr(p.peekToken().ID), p.peekToken().StartMark)
		}
		token = p.getToken()
		p.states = append(p.states, p.parseDocumentEnd)
		p.state = p.parseDocumentContent
		return &Event{Kind: DocumentStartEvent, StartMark: start, EndMark: token.EndMark, Explicit: true, Version: version, Tags: tags}
	}
	token := p.getToken()
	p.state = nil
	return &Event{Kind: StreamEndEvent, StartMark: token.StartMark, EndMark: token.EndMark}
}

func (p *parser) parseDocumentEnd() *Event {
	token := p.peekToken()
	start, end := token.StartMark, token.StartMark
	explicit := false
	if p.checkToken(DocumentEndToken) {
		token = p.getToken()
		end = token.EndMark
		explicit = true
	}
	p.state = p.parseDocumentStart
	return &Event{Kind: DocumentEndEvent, StartMark: start, EndMark: end, Explicit: explicit}
}

func (p *parser) parseDocumentContent() *Event {
	if p.checkToken(DirectiveToken, DocumentStartToken, DocumentEndToken, StreamEndToken) {
		event := p.processEmptyScalar(p.peekToken().StartMark)
		p.state = p.popState()
		return event
	}
	return p.parseBlockNode()
}

func (p *parser) processDirectives() (*Tuple, map[string]string) {
	p.yamlVersion = nil
	p.tagHandles = map[string]string{}
	for p.checkToken(DirectiveToken) {
		token := p.getToken()
		switch token.Name {
		case "YAML":
			if p.yamlVersion != nil {
				parserError("", nil, "found duplicate YAML directive", token.StartMark)
			}
			value := token.Value.(*Tuple)
			if value.Items[0].(*big.Int).Cmp(big.NewInt(1)) != 0 {
				parserError("", nil, "found incompatible YAML document (version 1.* is required)", token.StartMark)
			}
			p.yamlVersion = value
		case "TAG":
			value := token.Value.(*Tuple)
			handle, prefix := value.Items[0].(string), value.Items[1].(string)
			if _, ok := p.tagHandles[handle]; ok {
				parserError("", nil, "duplicate tag handle "+pyval.StrRepr(handle), token.StartMark)
			}
			p.tagHandles[handle] = prefix
		}
	}
	var tags map[string]string
	if len(p.tagHandles) > 0 {
		tags = make(map[string]string, len(p.tagHandles))
		for k, v := range p.tagHandles {
			tags[k] = v
		}
	}
	for k, v := range defaultTags {
		if _, ok := p.tagHandles[k]; !ok {
			p.tagHandles[k] = v
		}
	}
	return p.yamlVersion, tags
}

func (p *parser) parseBlockNode() *Event                     { return p.parseNode(true, false) }
func (p *parser) parseFlowNode() *Event                      { return p.parseNode(false, false) }
func (p *parser) parseBlockNodeOrIndentlessSequence() *Event { return p.parseNode(true, true) }

func flowStyle(v bool) *bool { return &v }

func (p *parser) parseNode(block, indentlessSequence bool) *Event {
	if p.checkToken(AliasToken) {
		token := p.getToken()
		p.state = p.popState()
		anchor := token.Value.(string)
		return &Event{Kind: AliasEvent, Anchor: &anchor, StartMark: token.StartMark, EndMark: token.EndMark}
	}
	var anchor, tag *string
	var start, end, tagMark *Mark
	var tagValue *Tuple
	if p.checkToken(AnchorToken) {
		token := p.getToken()
		start, end = token.StartMark, token.EndMark
		value := token.Value.(string)
		anchor = &value
		if p.checkToken(TagToken) {
			token = p.getToken()
			tagMark, end = token.StartMark, token.EndMark
			tagValue = token.Value.(*Tuple)
		}
	} else if p.checkToken(TagToken) {
		token := p.getToken()
		start, tagMark, end = token.StartMark, token.StartMark, token.EndMark
		tagValue = token.Value.(*Tuple)
		if p.checkToken(AnchorToken) {
			token = p.getToken()
			end = token.EndMark
			value := token.Value.(string)
			anchor = &value
		}
	}
	if tagValue != nil {
		suffix := tagValue.Items[1].(string)
		if handle, ok := tagValue.Items[0].(string); ok {
			prefix, defined := p.tagHandles[handle]
			if !defined {
				parserError("while parsing a node", start, "found undefined tag handle "+pyval.StrRepr(handle), tagMark)
			}
			value := prefix + suffix
			tag = &value
		} else {
			tag = &suffix
		}
	}
	if start == nil {
		start = p.peekToken().StartMark
		end = start
	}
	implicit := tag == nil || *tag == "!"
	collection := func(kind string, end *Mark, flow *bool, next func() *Event) *Event {
		p.state = next
		return &Event{Kind: kind, Anchor: anchor, Tag: tag, Implicit: [2]bool{implicit, false}, FlowStyle: flow, StartMark: start, EndMark: end}
	}
	if indentlessSequence && p.checkToken(BlockEntryToken) {
		return collection(SequenceStartEvent, p.peekToken().EndMark, nil, p.parseIndentlessSequenceEntry)
	}
	switch {
	case p.checkToken(ScalarToken):
		token := p.getToken()
		var pair [2]bool
		switch {
		case (token.Plain && tag == nil) || (tag != nil && *tag == "!"):
			pair = [2]bool{true, false}
		case tag == nil:
			pair = [2]bool{false, true}
		}
		p.state = p.popState()
		return &Event{Kind: ScalarEvent, Anchor: anchor, Tag: tag, Implicit: pair, Value: token.Value.(string),
			Style: token.Style, StartMark: start, EndMark: token.EndMark}
	case p.checkToken(FlowSequenceStartToken):
		return collection(SequenceStartEvent, p.peekToken().EndMark, flowStyle(true), p.parseFlowSequenceFirstEntry)
	case p.checkToken(FlowMappingStartToken):
		return collection(MappingStartEvent, p.peekToken().EndMark, flowStyle(true), p.parseFlowMappingFirstKey)
	case block && p.checkToken(BlockSequenceStartToken):
		return collection(SequenceStartEvent, p.peekToken().StartMark, flowStyle(false), p.parseBlockSequenceFirstEntry)
	case block && p.checkToken(BlockMappingStartToken):
		return collection(MappingStartEvent, p.peekToken().StartMark, flowStyle(false), p.parseBlockMappingFirstKey)
	case anchor != nil || tag != nil:
		p.state = p.popState()
		return &Event{Kind: ScalarEvent, Anchor: anchor, Tag: tag, Implicit: [2]bool{implicit, false}, StartMark: start, EndMark: end}
	}
	node := "flow"
	if block {
		node = "block"
	}
	token := p.peekToken()
	parserError(fmt.Sprintf("while parsing a %s node", node), start,
		"expected the node content, but found "+pyval.StrRepr(token.ID), token.StartMark)
	return nil
}

func (p *parser) parseBlockSequenceFirstEntry() *Event {
	token := p.getToken()
	p.marks = append(p.marks, token.StartMark)
	return p.parseBlockSequenceEntry()
}

func (p *parser) lastMark() *Mark { return p.marks[len(p.marks)-1] }

func (p *parser) parseBlockSequenceEntry() *Event {
	if p.checkToken(BlockEntryToken) {
		token := p.getToken()
		if !p.checkToken(BlockEntryToken, BlockEndToken) {
			p.states = append(p.states, p.parseBlockSequenceEntry)
			return p.parseBlockNode()
		}
		p.state = p.parseBlockSequenceEntry
		return p.processEmptyScalar(token.EndMark)
	}
	if !p.checkToken(BlockEndToken) {
		token := p.peekToken()
		parserError("while parsing a block collection", p.lastMark(),
			"expected <block end>, but found "+pyval.StrRepr(token.ID), token.StartMark)
	}
	token := p.getToken()
	p.state = p.popState()
	p.marks = p.marks[:len(p.marks)-1]
	return &Event{Kind: SequenceEndEvent, StartMark: token.StartMark, EndMark: token.EndMark}
}

func (p *parser) parseIndentlessSequenceEntry() *Event {
	if p.checkToken(BlockEntryToken) {
		token := p.getToken()
		if !p.checkToken(BlockEntryToken, KeyToken, ValueToken, BlockEndToken) {
			p.states = append(p.states, p.parseIndentlessSequenceEntry)
			return p.parseBlockNode()
		}
		p.state = p.parseIndentlessSequenceEntry
		return p.processEmptyScalar(token.EndMark)
	}
	token := p.peekToken()
	p.state = p.popState()
	return &Event{Kind: SequenceEndEvent, StartMark: token.StartMark, EndMark: token.StartMark}
}

func (p *parser) parseBlockMappingFirstKey() *Event {
	token := p.getToken()
	p.marks = append(p.marks, token.StartMark)
	return p.parseBlockMappingKey()
}

func (p *parser) parseBlockMappingKey() *Event {
	if p.checkToken(KeyToken) {
		token := p.getToken()
		if !p.checkToken(KeyToken, ValueToken, BlockEndToken) {
			p.states = append(p.states, p.parseBlockMappingValue)
			return p.parseBlockNodeOrIndentlessSequence()
		}
		p.state = p.parseBlockMappingValue
		return p.processEmptyScalar(token.EndMark)
	}
	if !p.checkToken(BlockEndToken) {
		token := p.peekToken()
		parserError("while parsing a block mapping", p.lastMark(),
			"expected <block end>, but found "+pyval.StrRepr(token.ID), token.StartMark)
	}
	token := p.getToken()
	p.state = p.popState()
	p.marks = p.marks[:len(p.marks)-1]
	return &Event{Kind: MappingEndEvent, StartMark: token.StartMark, EndMark: token.EndMark}
}

func (p *parser) parseBlockMappingValue() *Event {
	if p.checkToken(ValueToken) {
		token := p.getToken()
		if !p.checkToken(KeyToken, ValueToken, BlockEndToken) {
			p.states = append(p.states, p.parseBlockMappingKey)
			return p.parseBlockNodeOrIndentlessSequence()
		}
		p.state = p.parseBlockMappingKey
		return p.processEmptyScalar(token.EndMark)
	}
	p.state = p.parseBlockMappingKey
	return p.processEmptyScalar(p.peekToken().StartMark)
}

func (p *parser) parseFlowSequenceFirstEntry() *Event {
	token := p.getToken()
	p.marks = append(p.marks, token.StartMark)
	return p.parseFlowSequenceEntryAt(true)
}

func (p *parser) parseFlowSequenceEntry() *Event { return p.parseFlowSequenceEntryAt(false) }

func (p *parser) parseFlowSequenceEntryAt(first bool) *Event {
	if !p.checkToken(FlowSequenceEndToken) {
		if !first {
			if p.checkToken(FlowEntryToken) {
				p.getToken()
			} else {
				token := p.peekToken()
				parserError("while parsing a flow sequence", p.lastMark(),
					"expected ',' or ']', but got "+pyval.StrRepr(token.ID), token.StartMark)
			}
		}
		if p.checkToken(KeyToken) {
			token := p.peekToken()
			p.state = p.parseFlowSequenceEntryMappingKey
			return &Event{Kind: MappingStartEvent, Implicit: [2]bool{true, false}, FlowStyle: flowStyle(true),
				StartMark: token.StartMark, EndMark: token.EndMark}
		} else if !p.checkToken(FlowSequenceEndToken) {
			p.states = append(p.states, p.parseFlowSequenceEntry)
			return p.parseFlowNode()
		}
	}
	token := p.getToken()
	p.state = p.popState()
	p.marks = p.marks[:len(p.marks)-1]
	return &Event{Kind: SequenceEndEvent, StartMark: token.StartMark, EndMark: token.EndMark}
}

func (p *parser) parseFlowSequenceEntryMappingKey() *Event {
	token := p.getToken()
	if !p.checkToken(ValueToken, FlowEntryToken, FlowSequenceEndToken) {
		p.states = append(p.states, p.parseFlowSequenceEntryMappingValue)
		return p.parseFlowNode()
	}
	p.state = p.parseFlowSequenceEntryMappingValue
	return p.processEmptyScalar(token.EndMark)
}

func (p *parser) parseFlowSequenceEntryMappingValue() *Event {
	if p.checkToken(ValueToken) {
		token := p.getToken()
		if !p.checkToken(FlowEntryToken, FlowSequenceEndToken) {
			p.states = append(p.states, p.parseFlowSequenceEntryMappingEnd)
			return p.parseFlowNode()
		}
		p.state = p.parseFlowSequenceEntryMappingEnd
		return p.processEmptyScalar(token.EndMark)
	}
	p.state = p.parseFlowSequenceEntryMappingEnd
	return p.processEmptyScalar(p.peekToken().StartMark)
}

func (p *parser) parseFlowSequenceEntryMappingEnd() *Event {
	p.state = p.parseFlowSequenceEntry
	token := p.peekToken()
	return &Event{Kind: MappingEndEvent, StartMark: token.StartMark, EndMark: token.StartMark}
}

func (p *parser) parseFlowMappingFirstKey() *Event {
	token := p.getToken()
	p.marks = append(p.marks, token.StartMark)
	return p.parseFlowMappingKeyAt(true)
}

func (p *parser) parseFlowMappingKey() *Event { return p.parseFlowMappingKeyAt(false) }

func (p *parser) parseFlowMappingKeyAt(first bool) *Event {
	if !p.checkToken(FlowMappingEndToken) {
		if !first {
			if p.checkToken(FlowEntryToken) {
				p.getToken()
			} else {
				token := p.peekToken()
				parserError("while parsing a flow mapping", p.lastMark(),
					"expected ',' or '}', but got "+pyval.StrRepr(token.ID), token.StartMark)
			}
		}
		if p.checkToken(KeyToken) {
			token := p.getToken()
			if !p.checkToken(ValueToken, FlowEntryToken, FlowMappingEndToken) {
				p.states = append(p.states, p.parseFlowMappingValue)
				return p.parseFlowNode()
			}
			p.state = p.parseFlowMappingValue
			return p.processEmptyScalar(token.EndMark)
		} else if !p.checkToken(FlowMappingEndToken) {
			p.states = append(p.states, p.parseFlowMappingEmptyValue)
			return p.parseFlowNode()
		}
	}
	token := p.getToken()
	p.state = p.popState()
	p.marks = p.marks[:len(p.marks)-1]
	return &Event{Kind: MappingEndEvent, StartMark: token.StartMark, EndMark: token.EndMark}
}

func (p *parser) parseFlowMappingValue() *Event {
	if p.checkToken(ValueToken) {
		token := p.getToken()
		if !p.checkToken(FlowEntryToken, FlowMappingEndToken) {
			p.states = append(p.states, p.parseFlowMappingKey)
			return p.parseFlowNode()
		}
		p.state = p.parseFlowMappingKey
		return p.processEmptyScalar(token.EndMark)
	}
	p.state = p.parseFlowMappingKey
	return p.processEmptyScalar(p.peekToken().StartMark)
}

func (p *parser) parseFlowMappingEmptyValue() *Event {
	p.state = p.parseFlowMappingKey
	return p.processEmptyScalar(p.peekToken().StartMark)
}

func (p *parser) processEmptyScalar(mark *Mark) *Event {
	return &Event{Kind: ScalarEvent, Implicit: [2]bool{true, false}, StartMark: mark, EndMark: mark}
}
