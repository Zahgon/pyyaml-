// Port of PyYAML's lib/yaml/resolver.py (MIT License; Copyright (c) 2017-2021 Ingy döt Net, Copyright
// (c) 2006-2016 Kirill Simonov; see LICENSE).

package yaml

import (
	"regexp"

	"github.com/yaml/pyyaml-go/internal/pyval"
)

const (
	defaultScalarTag   = "tag:yaml.org,2002:str"
	defaultSequenceTag = "tag:yaml.org,2002:seq"
	defaultMappingTag  = "tag:yaml.org,2002:map"
)

type implicitResolver struct {
	tag    string
	regexp *regexp.Regexp
	first  string
	empty  bool
}

// Python's `$` also matches before a final newline, hence `\n?\z`.
func compileResolver(body string) *regexp.Regexp { return regexp.MustCompile(`^(?:` + body + `)\n?\z`) }

// implicitResolvers are in add_implicit_resolver order: a first character's list keeps that order.
var implicitResolvers = []implicitResolver{
	{tag: "tag:yaml.org,2002:bool",
		regexp: compileResolver(`yes|Yes|YES|no|No|NO|true|True|TRUE|false|False|FALSE|on|On|ON|off|Off|OFF`),
		first:  "yYnNtTfFoO"},
	{tag: "tag:yaml.org,2002:float",
		regexp: compileResolver(`[-+]?(?:[0-9][0-9_]*)\.[0-9_]*(?:[eE][-+][0-9]+)?|\.[0-9][0-9_]*(?:[eE][-+][0-9]+)?|[-+]?[0-9][0-9_]*(?::[0-5]?[0-9])+\.[0-9_]*|[-+]?\.(?:inf|Inf|INF)|\.(?:nan|NaN|NAN)`),
		first:  "-+0123456789."},
	{tag: "tag:yaml.org,2002:int",
		regexp: compileResolver(`[-+]?0b[0-1_]+|[-+]?0[0-7_]+|[-+]?(?:0|[1-9][0-9_]*)|[-+]?0x[0-9a-fA-F_]+|[-+]?[1-9][0-9_]*(?::[0-5]?[0-9])+`),
		first:  "-+0123456789"},
	{tag: "tag:yaml.org,2002:merge", regexp: compileResolver(`<<`), first: "<"},
	{tag: "tag:yaml.org,2002:null", regexp: compileResolver(`~|null|Null|NULL|`), first: "~nN", empty: true},
	{tag: "tag:yaml.org,2002:timestamp",
		regexp: compileResolver(`[0-9][0-9][0-9][0-9]-[0-9][0-9]-[0-9][0-9]|[0-9][0-9][0-9][0-9]-[0-9][0-9]?-[0-9][0-9]?(?:[Tt]|[ \t]+)[0-9][0-9]?:[0-9][0-9]:[0-9][0-9](?:\.[0-9]*)?(?:[ \t]*(?:Z|[-+][0-9][0-9]?(?::[0-9][0-9])?))?`),
		first:  "0123456789"},
	{tag: "tag:yaml.org,2002:value", regexp: compileResolver(`=`), first: "="},
	{tag: "tag:yaml.org,2002:yaml", regexp: compileResolver(`!|&|\*`), first: "!&*"},
}

// resolver is BaseResolver (implicit false) or Resolver (implicit true); path resolvers are not
// ported.
type resolver struct {
	implicit bool
}

func (r resolver) resolve(kind, value string, implicit bool) string {
	if kind == ScalarNode && implicit && r.implicit {
		var first rune
		if value != "" {
			first, _ = pyval.DecodeRune(value)
		}
		for i := range implicitResolvers {
			ir := &implicitResolvers[i]
			var candidate bool
			if value == "" {
				candidate = ir.empty
			} else {
				candidate = isIn(first, ir.first)
			}
			if candidate && ir.regexp.MatchString(value) {
				return ir.tag
			}
		}
	}
	switch kind {
	case SequenceNode:
		return defaultSequenceTag
	case MappingNode:
		return defaultMappingTag
	}
	return defaultScalarTag
}
