// Port of PyYAML's lib/yaml/nodes.py (MIT License; Copyright (c) 2017-2021 Ingy döt Net, Copyright
// (c) 2006-2016 Kirill Simonov; see LICENSE).

package yaml

import (
	"strings"

	"github.com/yaml/pyyaml-go/internal/pyval"
)

// Node kinds: the class names of nodes.py.
const (
	ScalarNode   = "ScalarNode"
	SequenceNode = "SequenceNode"
	MappingNode  = "MappingNode"
)

// Node is one node of a representation tree. Kind is its class name and ID its id attribute ("scalar",
// "sequence" or "mapping"). A scalar holds Value and Style ("" is None), a sequence Items and a mapping
// Pairs (key, value); collections have FlowStyle (nil is None). A *Node has identity: an alias in the
// document is a second reference to the one node, so the graph can be cyclic.
type Node struct {
	Kind      string
	ID        string
	Tag       string
	Value     string
	Items     []*Node
	Pairs     [][2]*Node
	Style     string
	FlowStyle *bool
	StartMark *Mark
	EndMark   *Mark
}

func newScalarNode(tag, value, style string, start, end *Mark) *Node {
	return &Node{Kind: ScalarNode, ID: "scalar", Tag: tag, Value: value, Style: style, StartMark: start, EndMark: end}
}

func newCollectionNode(kind, tag string, flow *bool, start *Mark) *Node {
	id := "sequence"
	if kind == MappingNode {
		id = "mapping"
	}
	return &Node{Kind: kind, ID: id, Tag: tag, FlowStyle: flow, StartMark: start}
}

func (st *reprState) writeNode(b *strings.Builder, n *Node) error {
	b.WriteString(n.Kind)
	b.WriteString("(tag=")
	b.WriteString(pyval.StrRepr(n.Tag))
	b.WriteString(", value=")
	switch n.Kind {
	case ScalarNode:
		b.WriteString(pyval.StrRepr(n.Value))
	case SequenceNode:
		if len(n.Items) == 0 {
			b.WriteString("[]")
			break
		}
		if st.active[n] {
			b.WriteString("[...]")
			break
		}
		st.active[n] = true
		defer delete(st.active, n)
		b.WriteByte('[')
		for i, item := range n.Items {
			if i > 0 {
				b.WriteString(", ")
			}
			if err := st.writeNode(b, item); err != nil {
				return err
			}
		}
		b.WriteByte(']')
	case MappingNode:
		if len(n.Pairs) == 0 {
			b.WriteString("[]")
			break
		}
		if st.active[n] {
			b.WriteString("[...]")
			break
		}
		st.active[n] = true
		defer delete(st.active, n)
		b.WriteByte('[')
		for i, pair := range n.Pairs {
			if i > 0 {
				b.WriteString(", ")
			}
			b.WriteByte('(')
			if err := st.writeNode(b, pair[0]); err != nil {
				return err
			}
			b.WriteString(", ")
			if err := st.writeNode(b, pair[1]); err != nil {
				return err
			}
			b.WriteByte(')')
		}
		b.WriteByte(']')
	}
	b.WriteByte(')')
	return nil
}
