// Port of PyYAML's lib/yaml/tokens.py (MIT License; Copyright (c) 2017-2021 Ingy döt Net, Copyright
// (c) 2006-2016 Kirill Simonov; see LICENSE).

package yaml

import (
	"strings"

	"github.com/yaml/pyyaml-go/internal/pyval"
)

// Token kinds: the class names of tokens.py.
const (
	DirectiveToken          = "DirectiveToken"
	DocumentStartToken      = "DocumentStartToken"
	DocumentEndToken        = "DocumentEndToken"
	StreamStartToken        = "StreamStartToken"
	StreamEndToken          = "StreamEndToken"
	BlockSequenceStartToken = "BlockSequenceStartToken"
	BlockMappingStartToken  = "BlockMappingStartToken"
	BlockEndToken           = "BlockEndToken"
	FlowSequenceStartToken  = "FlowSequenceStartToken"
	FlowMappingStartToken   = "FlowMappingStartToken"
	FlowSequenceEndToken    = "FlowSequenceEndToken"
	FlowMappingEndToken     = "FlowMappingEndToken"
	KeyToken                = "KeyToken"
	ValueToken              = "ValueToken"
	BlockEntryToken         = "BlockEntryToken"
	FlowEntryToken          = "FlowEntryToken"
	AliasToken              = "AliasToken"
	AnchorToken             = "AnchorToken"
	TagToken                = "TagToken"
	ScalarToken             = "ScalarToken"
)

var tokenIDs = map[string]string{
	DirectiveToken:          "<directive>",
	DocumentStartToken:      "<document start>",
	DocumentEndToken:        "<document end>",
	StreamStartToken:        "<stream start>",
	StreamEndToken:          "<stream end>",
	BlockSequenceStartToken: "<block sequence start>",
	BlockMappingStartToken:  "<block mapping start>",
	BlockEndToken:           "<block end>",
	FlowSequenceStartToken:  "[",
	FlowMappingStartToken:   "{",
	FlowSequenceEndToken:    "]",
	FlowMappingEndToken:     "}",
	KeyToken:                "?",
	ValueToken:              ":",
	BlockEntryToken:         "-",
	FlowEntryToken:          ",",
	AliasToken:              "<alias>",
	AnchorToken:             "<anchor>",
	TagToken:                "<tag>",
	ScalarToken:             "<scalar>",
}

// Token is one token of Scan. Kind is its class name and ID its id attribute. Name is the directive's
// name. Value is the str of an alias, anchor or scalar; for a tag a *Tuple (handle or nil, suffix);
// for a directive a *Tuple (major, minor as *big.Int) or (handle, prefix), or nil. Plain and Style
// belong to a scalar (Style "" is None). Encoding belongs to the stream start (nil is None).
type Token struct {
	Kind      string
	ID        string
	Name      string
	Value     any
	Plain     bool
	Style     string
	Encoding  *string
	StartMark *Mark
	EndMark   *Mark
}

func newToken(kind string, start, end *Mark) *Token {
	return &Token{Kind: kind, ID: tokenIDs[kind], StartMark: start, EndMark: end}
}

func reprOptStr(s *string) string {
	if s == nil {
		return "None"
	}
	return pyval.StrRepr(*s)
}

func (st *reprState) writeToken(b *strings.Builder, t *Token) error {
	b.WriteString(t.Kind)
	b.WriteByte('(')
	switch t.Kind {
	case DirectiveToken:
		b.WriteString("name=")
		b.WriteString(pyval.StrRepr(t.Name))
		b.WriteString(", value=")
		if err := st.write(b, t.Value); err != nil {
			return err
		}
	case StreamStartToken:
		b.WriteString("encoding=")
		b.WriteString(reprOptStr(t.Encoding))
	case AliasToken, AnchorToken, TagToken:
		b.WriteString("value=")
		if err := st.write(b, t.Value); err != nil {
			return err
		}
	case ScalarToken:
		if t.Plain {
			b.WriteString("plain=True")
		} else {
			b.WriteString("plain=False")
		}
		b.WriteString(", style=")
		if t.Style == "" {
			b.WriteString("None")
		} else {
			b.WriteString(pyval.StrRepr(t.Style))
		}
		b.WriteString(", value=")
		if err := st.write(b, t.Value); err != nil {
			return err
		}
	}
	b.WriteByte(')')
	return nil
}
