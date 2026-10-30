// Port of PyYAML's lib/yaml/serializer.py (MIT License; Copyright (c) 2017-2021 Ingy döt Net, Copyright
// (c) 2006-2016 Kirill Simonov; see LICENSE).

package yaml

import "fmt"

// serializer turns nodes into events for an emitter.
type serializer struct {
	*emitter
	resolver
	useEncoding      *string
	useExplicitStart bool
	useExplicitEnd   bool
	useVersion       *Tuple
	useTags          map[string]string
	serializedNodes  map[*Node]bool
	anchors          map[*Node]*string
	lastAnchorID     int
	opened           bool
	closed           bool
}

func serializerError(msg string) { raise(&YAMLError{Class: "SerializerError", Msg: msg}) }

func (s *serializer) open() {
	switch {
	case !s.opened:
		s.emit(&Event{Kind: StreamStartEvent, Encoding: s.useEncoding})
		s.opened = true
	case s.closed:
		serializerError("serializer is closed")
	default:
		serializerError("serializer is already opened")
	}
}

func (s *serializer) close() {
	if !s.opened {
		serializerError("serializer is not opened")
	} else if !s.closed {
		s.emit(&Event{Kind: StreamEndEvent})
		s.closed = true
	}
}

func (s *serializer) serialize(node *Node) {
	if !s.opened {
		serializerError("serializer is not opened")
	} else if s.closed {
		serializerError("serializer is closed")
	}
	s.emit(&Event{Kind: DocumentStartEvent, Explicit: s.useExplicitStart, Version: s.useVersion, Tags: s.useTags})
	s.anchorNode(node)
	s.serializeNode(node)
	s.emit(&Event{Kind: DocumentEndEvent, Explicit: s.useExplicitEnd})
	s.serializedNodes = map[*Node]bool{}
	s.anchors = map[*Node]*string{}
	s.lastAnchorID = 0
}

func (s *serializer) anchorNode(node *Node) {
	if anchor, ok := s.anchors[node]; ok {
		if anchor == nil {
			s.lastAnchorID++
			name := fmt.Sprintf("id%03d", s.lastAnchorID)
			s.anchors[node] = &name
		}
		return
	}
	s.anchors[node] = nil
	switch node.Kind {
	case SequenceNode:
		for _, item := range node.Items {
			s.anchorNode(item)
		}
	case MappingNode:
		for _, pair := range node.Pairs {
			s.anchorNode(pair[0])
			s.anchorNode(pair[1])
		}
	}
}

func (s *serializer) serializeNode(node *Node) {
	alias := s.anchors[node]
	if s.serializedNodes[node] {
		s.emit(&Event{Kind: AliasEvent, Anchor: alias})
		return
	}
	s.serializedNodes[node] = true
	tag := node.Tag
	switch node.Kind {
	case ScalarNode:
		detected := s.resolve(ScalarNode, node.Value, true)
		def := s.resolve(ScalarNode, node.Value, false)
		s.emit(&Event{Kind: ScalarEvent, Anchor: alias, Tag: &tag, Implicit: [2]bool{tag == detected, tag == def},
			Value: node.Value, Style: node.Style})
	case SequenceNode:
		s.emit(&Event{Kind: SequenceStartEvent, Anchor: alias, Tag: &tag,
			Implicit: [2]bool{tag == s.resolve(SequenceNode, "", true), false}, FlowStyle: node.FlowStyle})
		for _, item := range node.Items {
			s.serializeNode(item)
		}
		s.emit(&Event{Kind: SequenceEndEvent})
	case MappingNode:
		s.emit(&Event{Kind: MappingStartEvent, Anchor: alias, Tag: &tag,
			Implicit: [2]bool{tag == s.resolve(MappingNode, "", true), false}, FlowStyle: node.FlowStyle})
		for _, pair := range node.Pairs {
			s.serializeNode(pair[0])
			s.serializeNode(pair[1])
		}
		s.emit(&Event{Kind: MappingEndEvent})
	}
}
