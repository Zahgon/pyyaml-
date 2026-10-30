// Port of PyYAML's lib/yaml/composer.py (MIT License; Copyright (c) 2017-2021 Ingy döt Net, Copyright
// (c) 2006-2016 Kirill Simonov; see LICENSE).

package yaml

import "github.com/yaml/pyyaml-go/internal/pyval"

// composer turns events into representation trees.
type composer struct {
	*parser
	resolver
	anchors map[string]*Node
}

func newComposer(stream any, r resolver) *composer {
	return &composer{parser: newParser(newScanner(newReader(stream))), resolver: r, anchors: map[string]*Node{}}
}

func composerError(context string, contextMark *Mark, problem string, problemMark *Mark) {
	var c *string
	if context != "" {
		c = &context
	}
	raise(markedError("ComposerError", c, contextMark, problem, problemMark))
}

func (c *composer) checkNode() bool {
	if c.checkEvent(StreamStartEvent) {
		c.getEvent()
	}
	return !c.checkEvent(StreamEndEvent)
}

func (c *composer) getNode() *Node {
	if !c.checkEvent(StreamEndEvent) {
		return c.composeDocument()
	}
	return nil
}

func (c *composer) getSingleNode() *Node {
	c.getEvent()
	var document *Node
	if !c.checkEvent(StreamEndEvent) {
		document = c.composeDocument()
	}
	if !c.checkEvent(StreamEndEvent) {
		event := c.getEvent()
		composerError("expected a single document in the stream", document.StartMark,
			"but found another document", event.StartMark)
	}
	c.getEvent()
	return document
}

func (c *composer) composeDocument() *Node {
	c.getEvent()
	node := c.composeNode()
	c.getEvent()
	c.anchors = map[string]*Node{}
	return node
}

func (c *composer) composeNode() *Node {
	if c.checkEvent(AliasEvent) {
		event := c.getEvent()
		anchor := *event.Anchor
		node, ok := c.anchors[anchor]
		if !ok {
			composerError("", nil, "found undefined alias "+pyval.StrRepr(anchor), event.StartMark)
		}
		return node
	}
	event := c.peekEvent()
	if event.Anchor != nil {
		if first, ok := c.anchors[*event.Anchor]; ok {
			composerError("found duplicate anchor "+pyval.StrRepr(*event.Anchor)+"; first occurrence", first.StartMark,
				"second occurrence", event.StartMark)
		}
	}
	switch {
	case c.checkEvent(ScalarEvent):
		return c.composeScalarNode(event.Anchor)
	case c.checkEvent(SequenceStartEvent):
		return c.composeSequenceNode(event.Anchor)
	}
	return c.composeMappingNode(event.Anchor)
}

func (c *composer) composeScalarNode(anchor *string) *Node {
	event := c.getEvent()
	var tag string
	if event.Tag == nil || *event.Tag == "!" {
		tag = c.resolve(ScalarNode, event.Value, event.Implicit[0])
	} else {
		tag = *event.Tag
	}
	node := newScalarNode(tag, event.Value, event.Style, event.StartMark, event.EndMark)
	if anchor != nil {
		c.anchors[*anchor] = node
	}
	return node
}

func (c *composer) composeSequenceNode(anchor *string) *Node {
	start := c.getEvent()
	var tag string
	if start.Tag == nil || *start.Tag == "!" {
		tag = c.resolve(SequenceNode, "", false)
	} else {
		tag = *start.Tag
	}
	node := newCollectionNode(SequenceNode, tag, start.FlowStyle, start.StartMark)
	if anchor != nil {
		c.anchors[*anchor] = node
	}
	for !c.checkEvent(SequenceEndEvent) {
		node.Items = append(node.Items, c.composeNode())
	}
	node.EndMark = c.getEvent().EndMark
	return node
}

func (c *composer) composeMappingNode(anchor *string) *Node {
	start := c.getEvent()
	var tag string
	if start.Tag == nil || *start.Tag == "!" {
		tag = c.resolve(MappingNode, "", false)
	} else {
		tag = *start.Tag
	}
	node := newCollectionNode(MappingNode, tag, start.FlowStyle, start.StartMark)
	if anchor != nil {
		c.anchors[*anchor] = node
	}
	for !c.checkEvent(MappingEndEvent) {
		key := c.composeNode()
		value := c.composeNode()
		node.Pairs = append(node.Pairs, [2]*Node{key, value})
	}
	node.EndMark = c.getEvent().EndMark
	return node
}
