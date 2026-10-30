// Port of PyYAML's lib/yaml/representer.py, without the Python-object representers (MIT License;
// Copyright (c) 2017-2021 Ingy döt Net, Copyright (c) 2006-2016 Kirill Simonov; see LICENSE).

package yaml

import (
	"math"
	"math/big"
	"sort"
	"strings"

	"github.com/yaml/pyyaml-go/internal/pyval"
)

// representer turns values into nodes. Aliases follow the identity of the pointer types.
type representer struct {
	safe               bool
	defaultStyle       string
	defaultFlowStyle   *bool
	sortKeys           bool
	representedObjects map[any]*Node
	aliasKey           any
}

func (r *representer) ignoreAliases(data any) bool {
	switch x := data.(type) {
	case nil, string, []byte, bool, *big.Int, float64:
		return true
	case *Tuple:
		return len(x.Items) == 0
	case *List, *Dict, *PySet, *Date, *DateTime, *Complex, *Timezone:
		return false
	}
	return true
}

func (r *representer) representData(data any) *Node {
	if r.ignoreAliases(data) {
		r.aliasKey = nil
	} else {
		r.aliasKey = data
		if node, ok := r.representedObjects[data]; ok {
			return node
		}
	}
	switch x := data.(type) {
	case nil:
		return r.representScalar(yamlTagPrefix+"null", "null", "")
	case string:
		return r.representScalar(yamlTagPrefix+"str", x, "")
	case []byte:
		return r.representScalar(yamlTagPrefix+"binary", pyval.EncodeBase64Lines(x), "|")
	case bool:
		if x {
			return r.representScalar(yamlTagPrefix+"bool", "true", "")
		}
		return r.representScalar(yamlTagPrefix+"bool", "false", "")
	case *big.Int:
		s, err := pyval.IntReprChecked(x)
		if err != nil {
			raise(pyError(err))
		}
		return r.representScalar(yamlTagPrefix+"int", s, "")
	case float64:
		return r.representScalar(yamlTagPrefix+"float", floatScalar(x), "")
	case *List:
		return r.representSequence(yamlTagPrefix+"seq", x.Items)
	case *Tuple:
		if r.safe {
			return r.representSequence(yamlTagPrefix+"seq", x.Items)
		}
		return r.representSequence(yamlTagPrefix+"python/tuple", x.Items)
	case *Dict:
		return r.representMapping(yamlTagPrefix+"map", x.Keys(), x.Values())
	case *PySet:
		items := x.Items()
		return r.representMapping(yamlTagPrefix+"set", items, make([]any, len(items)))
	case *Date:
		return r.representScalar(yamlTagPrefix+"timestamp", x.isoformat(), "")
	case *DateTime:
		return r.representScalar(yamlTagPrefix+"timestamp", x.isoformat(), "")
	case *Complex:
		if !r.safe {
			return r.representScalar(yamlTagPrefix+"python/complex", complexScalar(x), "")
		}
	}
	text, err := Repr(data)
	if err != nil {
		raise(err)
	}
	raise(&YAMLError{Class: "RepresenterError", Msg: "('cannot represent an object', " + text + ")"})
	return nil
}

func floatScalar(data float64) string {
	switch {
	case math.IsNaN(data):
		return ".nan"
	case math.IsInf(data, 1):
		return ".inf"
	case math.IsInf(data, -1):
		return "-.inf"
	}
	value := strings.ToLower(pyval.FloatRepr(data))
	if !strings.Contains(value, ".") && strings.Contains(value, "e") {
		value = strings.Replace(value, "e", ".0e", 1)
	}
	return value
}

func complexScalar(data *Complex) string {
	switch {
	case data.Imag == 0:
		return pyval.FloatRepr(data.Real)
	case data.Real == 0:
		return pyval.FloatRepr(data.Imag) + "j"
	case data.Imag > 0:
		return pyval.FloatRepr(data.Real) + "+" + pyval.FloatRepr(data.Imag) + "j"
	}
	return pyval.FloatRepr(data.Real) + pyval.FloatRepr(data.Imag) + "j"
}

func (r *representer) representScalar(tag, value, style string) *Node {
	if style == "" {
		style = r.defaultStyle
	}
	node := newScalarNode(tag, value, style, nil, nil)
	if r.aliasKey != nil {
		r.representedObjects[r.aliasKey] = node
	}
	return node
}

func (r *representer) representSequence(tag string, sequence []any) *Node {
	node := newCollectionNode(SequenceNode, tag, nil, nil)
	if r.aliasKey != nil {
		r.representedObjects[r.aliasKey] = node
	}
	best := true
	for _, item := range sequence {
		child := r.representData(item)
		if !(child.Kind == ScalarNode && child.Style == "") {
			best = false
		}
		node.Items = append(node.Items, child)
	}
	if r.defaultFlowStyle != nil {
		node.FlowStyle = r.defaultFlowStyle
	} else {
		node.FlowStyle = &best
	}
	return node
}

func (r *representer) representMapping(tag string, keys, values []any) *Node {
	node := newCollectionNode(MappingNode, tag, nil, nil)
	if r.aliasKey != nil {
		r.representedObjects[r.aliasKey] = node
	}
	best := true
	order := make([]int, len(keys))
	for i := range order {
		order[i] = i
	}
	if r.sortKeys {
		sorted := append([]int(nil), order...)
		failed := false
		sort.SliceStable(sorted, func(i, j int) bool {
			less, ok := pyLess(keys[sorted[i]], keys[sorted[j]])
			if !ok {
				failed = true
			}
			return less
		})
		if !failed {
			order = sorted
		}
	}
	for _, i := range order {
		nodeKey := r.representData(keys[i])
		nodeValue := r.representData(values[i])
		if !(nodeKey.Kind == ScalarNode && nodeKey.Style == "") {
			best = false
		}
		if !(nodeValue.Kind == ScalarNode && nodeValue.Style == "") {
			best = false
		}
		node.Pairs = append(node.Pairs, [2]*Node{nodeKey, nodeValue})
	}
	if r.defaultFlowStyle != nil {
		node.FlowStyle = r.defaultFlowStyle
	} else {
		node.FlowStyle = &best
	}
	return node
}
