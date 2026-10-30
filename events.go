// Port of PyYAML's lib/yaml/events.py (MIT License; Copyright (c) 2017-2021 Ingy döt Net, Copyright
// (c) 2006-2016 Kirill Simonov; see LICENSE).

package yaml

import (
	"strings"

	"github.com/yaml/pyyaml-go/internal/pyval"
)

// Event kinds: the class names of events.py.
const (
	StreamStartEvent   = "StreamStartEvent"
	StreamEndEvent     = "StreamEndEvent"
	DocumentStartEvent = "DocumentStartEvent"
	DocumentEndEvent   = "DocumentEndEvent"
	AliasEvent         = "AliasEvent"
	ScalarEvent        = "ScalarEvent"
	SequenceStartEvent = "SequenceStartEvent"
	SequenceEndEvent   = "SequenceEndEvent"
	MappingStartEvent  = "MappingStartEvent"
	MappingEndEvent    = "MappingEndEvent"
)

// Event is one event of Parse or Emit. Kind is its class name. Anchor and Tag are nil for None.
// Implicit is the (plain, quoted) pair of a scalar; a collection start uses Implicit[0] alone. Style
// is "" for None. FlowStyle is nil for None. Version is nil for None, else a *Tuple of two *big.Int;
// Tags is nil for None; Encoding is nil for None. Fields a Python event class lacks stay at their zero
// value. Marks are nil for events not produced by Parse.
type Event struct {
	Kind      string
	Anchor    *string
	Tag       *string
	Implicit  [2]bool
	Value     string
	Style     string
	FlowStyle *bool
	Explicit  bool
	Version   *Tuple
	Tags      map[string]string
	Encoding  *string
	StartMark *Mark
	EndMark   *Mark
}

func boolRepr(v bool) string {
	if v {
		return "True"
	}
	return "False"
}

func (st *reprState) writeEvent(b *strings.Builder, e *Event) error {
	b.WriteString(e.Kind)
	b.WriteByte('(')
	switch e.Kind {
	case AliasEvent:
		b.WriteString("anchor=")
		b.WriteString(reprOptStr(e.Anchor))
	case ScalarEvent:
		b.WriteString("anchor=")
		b.WriteString(reprOptStr(e.Anchor))
		b.WriteString(", tag=")
		b.WriteString(reprOptStr(e.Tag))
		b.WriteString(", implicit=(")
		b.WriteString(boolRepr(e.Implicit[0]))
		b.WriteString(", ")
		b.WriteString(boolRepr(e.Implicit[1]))
		b.WriteString("), value=")
		b.WriteString(pyval.StrRepr(e.Value))
	case SequenceStartEvent, MappingStartEvent:
		b.WriteString("anchor=")
		b.WriteString(reprOptStr(e.Anchor))
		b.WriteString(", tag=")
		b.WriteString(reprOptStr(e.Tag))
		b.WriteString(", implicit=")
		b.WriteString(boolRepr(e.Implicit[0]))
	}
	b.WriteByte(')')
	return nil
}

func isNodeEvent(e *Event) bool {
	switch e.Kind {
	case AliasEvent, ScalarEvent, SequenceStartEvent, MappingStartEvent:
		return true
	}
	return false
}

func isCollectionStart(e *Event) bool {
	return e.Kind == SequenceStartEvent || e.Kind == MappingStartEvent
}

func isCollectionEnd(e *Event) bool {
	return e.Kind == SequenceEndEvent || e.Kind == MappingEndEvent
}
