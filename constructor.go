// Port of PyYAML's lib/yaml/constructor.py, up to FullConstructor (MIT License; Copyright (c) 2017-2021
// Ingy döt Net, Copyright (c) 2006-2016 Kirill Simonov; see LICENSE).

package yaml

import (
	"fmt"
	"math"
	"math/big"
	"regexp"
	"strconv"
	"strings"

	"github.com/yaml/pyyaml-go/internal/pyval"
)

const yamlTagPrefix = "tag:yaml.org,2002:"

// Loader selects a loader class. UnsafeLoader and Loader are not ported.
type Loader int

// The loaders.
const (
	BaseLoader Loader = iota
	SafeLoader
	FullLoader
)

// constructor builds Python values from nodes. The second half of a generator constructor of
// constructor.py is a func kept in stateGenerators until the document's end.
type constructor struct {
	*composer
	loader             Loader
	constructedObjects map[*Node]any
	recursiveObjects   map[*Node]bool
	stateGenerators    []func()
}

func newConstructor(stream any, loader Loader) *constructor {
	return &constructor{composer: newComposer(stream, resolver{implicit: loader != BaseLoader}), loader: loader,
		constructedObjects: map[*Node]any{}, recursiveObjects: map[*Node]bool{}}
}

func constructorError(context string, contextMark *Mark, problem string, problemMark *Mark) {
	var c *string
	if context != "" {
		c = &context
	}
	raise(markedError("ConstructorError", c, contextMark, problem, problemMark))
}

func (c *constructor) checkData() bool { return c.checkNode() }

func (c *constructor) getData() any {
	if c.checkNode() {
		return c.constructDocument(c.getNode())
	}
	return nil
}

func (c *constructor) getSingleData() any {
	node := c.getSingleNode()
	if node != nil {
		return c.constructDocument(node)
	}
	return nil
}

func (c *constructor) constructDocument(node *Node) any {
	data := c.constructObject(node)
	for len(c.stateGenerators) > 0 {
		generators := c.stateGenerators
		c.stateGenerators = nil
		for _, generator := range generators {
			generator()
		}
	}
	c.constructedObjects = map[*Node]any{}
	c.recursiveObjects = map[*Node]bool{}
	return data
}

func (c *constructor) constructObject(node *Node) any {
	if data, ok := c.constructedObjects[node]; ok {
		return data
	}
	if c.recursiveObjects[node] {
		constructorError("", nil, "found unconstructable recursive node", node.StartMark)
	}
	c.recursiveObjects[node] = true
	data, rest := c.dispatch(node)
	if rest != nil {
		c.stateGenerators = append(c.stateGenerators, rest)
	}
	c.constructedObjects[node] = data
	delete(c.recursiveObjects, node)
	return data
}

// dispatch picks the constructor of node's tag. It returns the value and, for a generator
// constructor, the part that runs after its yield.
func (c *constructor) dispatch(node *Node) (any, func()) {
	if c.loader == BaseLoader {
		switch node.Kind {
		case ScalarNode:
			return c.constructScalar(node), nil
		case SequenceNode:
			return &List{Items: c.constructSequence(node)}, nil
		}
		return c.baseConstructMapping(node), nil
	}
	tag := node.Tag
	if rest, ok := strings.CutPrefix(tag, yamlTagPrefix); ok {
		switch rest {
		case "null":
			return c.constructYAMLNull(node), nil
		case "bool":
			return c.constructYAMLBool(node), nil
		case "int":
			return c.constructYAMLInt(node), nil
		case "float":
			return c.constructYAMLFloat(node), nil
		case "binary":
			return c.constructYAMLBinary(node), nil
		case "timestamp":
			return c.constructYAMLTimestamp(node), nil
		case "omap":
			return c.constructYAMLPairs(node, "while constructing an ordered map")
		case "pairs":
			return c.constructYAMLPairs(node, "while constructing pairs")
		case "set":
			return c.constructYAMLSet(node)
		case "str":
			return c.constructScalar(node), nil
		case "seq":
			return c.constructYAMLSeq(node)
		case "map":
			return c.constructYAMLMap(node)
		}
		if c.loader == FullLoader {
			switch rest {
			case "python/none":
				return c.constructYAMLNull(node), nil
			case "python/bool":
				return c.constructYAMLBool(node), nil
			case "python/str", "python/unicode":
				return c.constructScalar(node), nil
			case "python/bytes":
				return c.constructYAMLBinary(node), nil
			case "python/int", "python/long":
				return c.constructYAMLInt(node), nil
			case "python/float":
				return c.constructYAMLFloat(node), nil
			case "python/complex":
				re, im, err := pyval.ParseComplex(c.constructScalar(node))
				if err != nil {
					raise(pyError(err))
				}
				return &Complex{Real: re, Imag: im}, nil
			case "python/list":
				return c.constructYAMLSeq(node)
			case "python/tuple":
				return &Tuple{Items: c.constructSequence(node)}, nil
			case "python/dict":
				return c.constructYAMLMap(node)
			}
		}
	}
	constructorError("", nil, "could not determine a constructor for the tag "+pyval.StrRepr(tag), node.StartMark)
	return nil, nil
}

// constructScalar is BaseConstructor.construct_scalar, with SafeConstructor's reading of a mapping's
// value key on top for the loaders that have it.
func (c *constructor) constructScalar(node *Node) string {
	if c.loader != BaseLoader && node.Kind == MappingNode {
		for _, pair := range node.Pairs {
			if pair[0].Tag == "tag:yaml.org,2002:value" {
				return c.constructScalar(pair[1])
			}
		}
	}
	if node.Kind != ScalarNode {
		constructorError("", nil, "expected a scalar node, but found "+node.ID, node.StartMark)
	}
	return node.Value
}

func (c *constructor) constructSequence(node *Node) []any {
	if node.Kind != SequenceNode {
		constructorError("", nil, "expected a sequence node, but found "+node.ID, node.StartMark)
	}
	items := make([]any, 0, len(node.Items))
	for _, child := range node.Items {
		items = append(items, c.constructObject(child))
	}
	return items
}

func isHashableType(v any) bool {
	switch v.(type) {
	case *List, *Dict, *PySet:
		return false
	}
	return true
}

func (c *constructor) baseConstructMapping(node *Node) *Dict {
	if node.Kind != MappingNode {
		constructorError("", nil, "expected a mapping node, but found "+node.ID, node.StartMark)
	}
	mapping := NewDict()
	for _, pair := range node.Pairs {
		key := c.constructObject(pair[0])
		if !isHashableType(key) {
			constructorError("while constructing a mapping", node.StartMark, "found unhashable key", pair[0].StartMark)
		}
		value := c.constructObject(pair[1])
		if err := mapping.Set(key, value); err != nil {
			raise(err)
		}
	}
	return mapping
}

func (c *constructor) constructMapping(node *Node) *Dict {
	if node.Kind == MappingNode {
		c.flattenMapping(node)
	}
	return c.baseConstructMapping(node)
}

// keyIdentity stands for id(key_node.value): a collection's value list belongs to its node; a scalar's
// str is an object of its own, except that CPython shares the empty string and each one-character
// string below U+0100.
func keyIdentity(node *Node) any {
	if node.Kind != ScalarNode {
		return node
	}
	if node.Value == "" {
		return "shared"
	}
	rs := pyval.Runes(node.Value)
	if len(rs) == 1 && rs[0] < 0x100 {
		return rs[0]
	}
	return node
}

func (c *constructor) flattenMapping(node *Node) {
	var merge [][2]*Node
	seen := map[any]bool{}
	addMerged := func(pairs [][2]*Node) {
		for _, pair := range pairs {
			id := keyIdentity(pair[0])
			if !seen[id] {
				seen[id] = true
				merge = append(merge, pair)
			}
		}
	}
	index := 0
	for index < len(node.Pairs) {
		key, value := node.Pairs[index][0], node.Pairs[index][1]
		switch key.Tag {
		case "tag:yaml.org,2002:merge":
			node.Pairs = append(node.Pairs[:index:index], node.Pairs[index+1:]...)
			switch value.Kind {
			case MappingNode:
				c.flattenMapping(value)
				addMerged(value.Pairs)
			case SequenceNode:
				for _, sub := range value.Items {
					if sub.Kind != MappingNode {
						constructorError("while constructing a mapping", node.StartMark,
							"expected a mapping for merging, but found "+sub.ID, sub.StartMark)
					}
					c.flattenMapping(sub)
					addMerged(sub.Pairs)
				}
			default:
				constructorError("while constructing a mapping", node.StartMark,
					"expected a mapping or list of mappings for merging, but found "+value.ID, value.StartMark)
			}
		case "tag:yaml.org,2002:value":
			key.Tag = "tag:yaml.org,2002:str"
			index++
		default:
			index++
		}
	}
	if len(merge) > 0 {
		node.Pairs = append(merge, node.Pairs...)
	}
}

func (c *constructor) constructYAMLNull(node *Node) any {
	c.constructScalar(node)
	return nil
}

func (c *constructor) constructYAMLBool(node *Node) any {
	value := pyval.Lower(c.constructScalar(node))
	switch value {
	case "yes", "true", "on":
		return true
	case "no", "false", "off":
		return false
	}
	raise(&PyError{Class: "KeyError", Msg: pyval.StrRepr(value)})
	return nil
}

func strIndexError() { raise(&PyError{Class: "IndexError", Msg: "string index out of range"}) }

func parseIntOrRaise(s string, base int) *big.Int {
	z, err := pyval.ParseInt(s, base)
	if err != nil {
		raise(pyError(err))
	}
	return z
}

func (c *constructor) constructYAMLInt(node *Node) any {
	value := strings.ReplaceAll(c.constructScalar(node), "_", "")
	if value == "" {
		strIndexError()
	}
	sign := int64(1)
	if value[0] == '-' {
		sign = -1
	}
	if value[0] == '+' || value[0] == '-' {
		value = value[1:]
	}
	if value == "0" {
		return new(big.Int)
	}
	var out *big.Int
	switch {
	case strings.HasPrefix(value, "0b"):
		out = parseIntOrRaise(value[2:], 2)
	case strings.HasPrefix(value, "0x"):
		out = parseIntOrRaise(value[2:], 16)
	case value == "":
		strIndexError()
	case value[0] == '0':
		out = parseIntOrRaise(value, 8)
	case strings.Contains(value, ":"):
		parts := strings.Split(value, ":")
		digits := make([]*big.Int, len(parts))
		for i, part := range parts {
			digits[i] = parseIntOrRaise(part, 10)
		}
		base := big.NewInt(1)
		out = new(big.Int)
		for i := len(digits) - 1; i >= 0; i-- {
			out.Add(out, new(big.Int).Mul(digits[i], base))
			base.Mul(base, big.NewInt(60))
		}
	default:
		out = parseIntOrRaise(value, 10)
	}
	return out.Mul(out, big.NewInt(sign))
}

// nanValue is the one float object SafeConstructor.nan_value; newNaN gives every other NaN result bits
// of its own, since a dict compares floats by identity before equality.
var (
	nanValue  = math.Float64frombits(0x7FF8000000000000)
	nanSerial uint64
)

func newNaN() float64 {
	nanSerial++
	return math.Float64frombits(0x7FF8000000000000 | nanSerial)
}

func parseFloatOrRaise(s string) float64 {
	f, err := pyval.ParseFloat(s)
	if err != nil {
		raise(pyError(err))
	}
	if math.IsNaN(f) {
		return newNaN()
	}
	return f
}

func (c *constructor) constructYAMLFloat(node *Node) any {
	value := pyval.Lower(strings.ReplaceAll(c.constructScalar(node), "_", ""))
	if value == "" {
		strIndexError()
	}
	sign := 1.0
	if value[0] == '-' {
		sign = -1.0
	}
	if value[0] == '+' || value[0] == '-' {
		value = value[1:]
	}
	switch {
	case value == ".inf":
		return sign * math.Inf(1)
	case value == ".nan":
		return nanValue
	case strings.Contains(value, ":"):
		parts := strings.Split(value, ":")
		digits := make([]float64, len(parts))
		for i, part := range parts {
			digits[i] = parseFloatOrRaise(part)
		}
		base := big.NewInt(1)
		total := 0.0
		for i := len(digits) - 1; i >= 0; i-- {
			factor, _ := new(big.Float).SetInt(base).Float64()
			if math.IsInf(factor, 0) {
				raise(&PyError{Class: "OverflowError", Msg: "int too large to convert to float"})
			}
			total += digits[i] * factor
			base.Mul(base, big.NewInt(60))
		}
		if math.IsNaN(total) {
			return newNaN()
		}
		return sign * total
	}
	f := parseFloatOrRaise(value)
	if math.IsNaN(f) {
		return newNaN()
	}
	return sign * f
}

func (c *constructor) constructYAMLBinary(node *Node) any {
	scalar := c.constructScalar(node)
	value, encErr := pyval.Encode(scalar, "ascii")
	if encErr != nil {
		constructorError("", nil, "failed to convert base64 data into ascii: "+encErr.Msg, node.StartMark)
	}
	data, msg := pyval.DecodeBase64(value)
	if msg != "" {
		constructorError("", nil, "failed to decode base64 data: "+msg, node.StartMark)
	}
	if data == nil {
		data = []byte{}
	}
	return data
}

var timestampRegexp = regexp.MustCompile(`^([0-9][0-9][0-9][0-9])-([0-9][0-9]?)-([0-9][0-9]?)(?:(?:[Tt]|[ \t]+)([0-9][0-9]?):([0-9][0-9]):([0-9][0-9])(?:\.([0-9]*))?(?:[ \t]*(Z|([-+])([0-9][0-9]?)(?::([0-9][0-9]))?))?)?\n?\z`)

func checkDate(year, month, day int) {
	if year < 1 || year > 9999 {
		raise(&PyError{Class: "ValueError", Msg: fmt.Sprintf("year %d is out of range", year)})
	}
	if month < 1 || month > 12 {
		raise(&PyError{Class: "ValueError", Msg: "month must be in 1..12"})
	}
	days := [...]int{31, 28, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31}[month-1]
	if month == 2 && isLeap(year) {
		days = 29
	}
	if day < 1 || day > days {
		raise(&PyError{Class: "ValueError", Msg: "day is out of range for month"})
	}
}

func (c *constructor) constructYAMLTimestamp(node *Node) any {
	c.constructScalar(node)
	if node.Kind != ScalarNode {
		raise(&PyError{Class: "TypeError", Msg: "expected string or bytes-like object, got 'list'"})
	}
	m := timestampRegexp.FindStringSubmatch(node.Value)
	if m == nil {
		raise(&PyError{Class: "AttributeError", Msg: "'NoneType' object has no attribute 'groupdict'"})
	}
	atoi := func(s string) int {
		n, _ := strconv.Atoi(s)
		return n
	}
	year, month, day := atoi(m[1]), atoi(m[2]), atoi(m[3])
	if m[4] == "" {
		checkDate(year, month, day)
		return &Date{Year: year, Month: month, Day: day}
	}
	hour, minute, second := atoi(m[4]), atoi(m[5]), atoi(m[6])
	fraction := 0
	if m[7] != "" {
		digits := m[7]
		if len(digits) > 6 {
			digits = digits[:6]
		}
		for len(digits) < 6 {
			digits += "0"
		}
		fraction = atoi(digits)
	}
	var tz *Timezone
	if m[9] != "" {
		delta := int64(atoi(m[10]))*3600 + int64(atoi(m[11]))*60
		if m[9] == "-" {
			delta = -delta
		}
		if delta >= 86400 || delta <= -86400 {
			raise(&PyError{Class: "ValueError", Msg: "offset must be a timedelta strictly between -timedelta(hours=24) and timedelta(hours=24), not " +
				timedeltaRepr(delta*1000000) + "."})
		}
		tz = &Timezone{OffsetMicroseconds: delta * 1000000}
	} else if m[8] != "" {
		tz = &Timezone{}
	}
	checkDate(year, month, day)
	if hour > 23 {
		raise(&PyError{Class: "ValueError", Msg: "hour must be in 0..23"})
	}
	if minute > 59 {
		raise(&PyError{Class: "ValueError", Msg: "minute must be in 0..59"})
	}
	if second > 59 {
		raise(&PyError{Class: "ValueError", Msg: "second must be in 0..59"})
	}
	return &DateTime{Year: year, Month: month, Day: day, Hour: hour, Minute: minute, Second: second, Microsecond: fraction, TZ: tz}
}

// constructYAMLPairs is construct_yaml_omap and construct_yaml_pairs, the same code with their own
// context.
func (c *constructor) constructYAMLPairs(node *Node, context string) (any, func()) {
	data := &List{}
	return data, func() {
		if node.Kind != SequenceNode {
			constructorError(context, node.StartMark, "expected a sequence, but found "+node.ID, node.StartMark)
		}
		for _, sub := range node.Items {
			if sub.Kind != MappingNode {
				constructorError(context, node.StartMark, "expected a mapping of length 1, but found "+sub.ID, sub.StartMark)
			}
			if len(sub.Pairs) != 1 {
				constructorError(context, node.StartMark,
					fmt.Sprintf("expected a single mapping item, but found %d items", len(sub.Pairs)), sub.StartMark)
			}
			key := c.constructObject(sub.Pairs[0][0])
			value := c.constructObject(sub.Pairs[0][1])
			data.Items = append(data.Items, &Tuple{Items: []any{key, value}})
		}
	}
}

func (c *constructor) constructYAMLSet(node *Node) (any, func()) {
	data := NewSet()
	return data, func() {
		value := c.constructMapping(node)
		for _, key := range value.Keys() {
			if err := data.Add(key); err != nil {
				raise(err)
			}
		}
	}
}

func (c *constructor) constructYAMLSeq(node *Node) (any, func()) {
	data := &List{}
	return data, func() { data.Items = append(data.Items, c.constructSequence(node)...) }
}

func (c *constructor) constructYAMLMap(node *Node) (any, func()) {
	data := NewDict()
	return data, func() {
		value := c.constructMapping(node)
		keys, values := value.Keys(), value.Values()
		for i, key := range keys {
			if err := data.Set(key, values[i]); err != nil {
				raise(err)
			}
		}
	}
}
