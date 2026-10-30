// Port of the module-level functions of PyYAML's lib/yaml/__init__.py and of the dumper classes of
// lib/yaml/dumper.py (MIT License; Copyright (c) 2017-2021 Ingy döt Net, Copyright (c) 2006-2016 Kirill
// Simonov; see LICENSE).

// Package yaml is a Go port of PyYAML's pure Python library: the reader, scanner, parser, composer,
// constructor, resolver, representer, serializer and emitter. It reproduces what the Python library
// returns and raises under CPython 3.12, including the behaviour it gets from Python itself. See the
// repository's README for the Go types that stand for Python values.
package yaml

// Scan is yaml.scan(stream): the tokens produced, and the exception that stopped the scan. stream is a
// string (str) or a []byte (bytes).
func Scan(stream any) (tokens []*Token, err error) {
	defer catch(&err)
	s := newScanner(newReader(stream))
	for s.checkToken() {
		tokens = append(tokens, s.getToken())
	}
	return tokens, nil
}

// Parse is yaml.parse(stream): the events produced, and the exception that stopped the parse.
func Parse(stream any) (events []*Event, err error) {
	defer catch(&err)
	p := newParser(newScanner(newReader(stream)))
	for p.checkEvent() {
		events = append(events, p.getEvent())
	}
	return events, nil
}

// Compose is yaml.compose(stream, loader): the root node of the only document, nil for an empty stream.
func Compose(stream any, loader Loader) (node *Node, err error) {
	defer catch(&err)
	return newComposer(stream, resolver{implicit: loader != BaseLoader}).getSingleNode(), nil
}

// ComposeAll is yaml.compose_all(stream, loader): the root nodes produced, and the exception that
// stopped the composition.
func ComposeAll(stream any, loader Loader) (nodes []*Node, err error) {
	defer catch(&err)
	c := newComposer(stream, resolver{implicit: loader != BaseLoader})
	for c.checkNode() {
		nodes = append(nodes, c.getNode())
	}
	return nodes, nil
}

// Load is yaml.load(stream, loader).
func Load(stream any, loader Loader) (data any, err error) {
	defer catch(&err)
	return newConstructor(stream, loader).getSingleData(), nil
}

// LoadAll is yaml.load_all(stream, loader): the documents produced, and the exception that stopped the
// load.
func LoadAll(stream any, loader Loader) (docs []any, err error) {
	defer catch(&err)
	c := newConstructor(stream, loader)
	for c.checkData() {
		docs = append(docs, c.getData())
	}
	return docs, nil
}

// SafeLoad is yaml.safe_load(stream).
func SafeLoad(stream any) (any, error) { return Load(stream, SafeLoader) }

// SafeLoadAll is yaml.safe_load_all(stream).
func SafeLoadAll(stream any) ([]any, error) { return LoadAll(stream, SafeLoader) }

// FullLoad is yaml.full_load(stream).
func FullLoad(stream any) (any, error) { return Load(stream, FullLoader) }

// FullLoadAll is yaml.full_load_all(stream).
func FullLoadAll(stream any) ([]any, error) { return LoadAll(stream, FullLoader) }

// Dumper selects a dumper class. BaseDumper is not ported.
type Dumper int

// The dumpers: DefaultDumper is yaml.Dumper.
const (
	DefaultDumper Dumper = iota
	SafeDumper
)

// DumpOptions holds the keywords of yaml.dump_all, each field standing for the keyword of the same name.
// Start from NewDumpOptions, which gives Python's defaults. Canonical, AllowUnicode, ExplicitStart and
// ExplicitEnd false, Indent and Width 0, and LineBreak, DefaultStyle and Encoding "" are None; Version nil
// and Tags nil are None; DefaultFlowStyle nil is None.
type DumpOptions struct {
	Dumper           Dumper
	DefaultStyle     string
	DefaultFlowStyle *bool
	Canonical        bool
	Indent           int
	Width            int
	AllowUnicode     bool
	LineBreak        string
	Encoding         string
	ExplicitStart    bool
	ExplicitEnd      bool
	Version          *Tuple
	Tags             map[string]string
	SortKeys         bool
}

// NewDumpOptions returns Python's defaults: Dumper, default_flow_style=False, sort_keys=True, all else None.
func NewDumpOptions() DumpOptions {
	block := false
	return DumpOptions{DefaultFlowStyle: &block, SortKeys: true}
}

// DumpStream is an open Dumper object writing to a buffer of its own: the object behind dump_all and
// serialize_all, for callers that drive it step by step.
type DumpStream struct {
	ser *serializer
	rep *representer
}

// NewDumpStream is Dumper(stream, **options) for a new in-memory stream, bytes when Encoding is set.
func NewDumpStream(opts DumpOptions) *DumpStream {
	binary := opts.Encoding != ""
	safe := opts.Dumper == SafeDumper
	ser := &serializer{emitter: newEmitter(&opts, binary), resolver: resolver{implicit: true},
		useExplicitStart: opts.ExplicitStart, useExplicitEnd: opts.ExplicitEnd, useVersion: opts.Version, useTags: opts.Tags,
		serializedNodes: map[*Node]bool{}, anchors: map[*Node]*string{}}
	if binary {
		ser.useEncoding = &opts.Encoding
	}
	return &DumpStream{ser: ser, rep: &representer{safe: safe, defaultStyle: opts.DefaultStyle,
		defaultFlowStyle: opts.DefaultFlowStyle, sortKeys: opts.SortKeys, representedObjects: map[any]*Node{}}}
}

// Open is dumper.open().
func (d *DumpStream) Open() (err error) {
	defer catch(&err)
	d.ser.open()
	return nil
}

// Close is dumper.close().
func (d *DumpStream) Close() (err error) {
	defer catch(&err)
	d.ser.close()
	return nil
}

// Serialize is dumper.serialize(node).
func (d *DumpStream) Serialize(node *Node) (err error) {
	defer catch(&err)
	d.ser.serialize(node)
	return nil
}

// Represent is dumper.represent(data).
func (d *DumpStream) Represent(data any) (err error) {
	defer catch(&err)
	node := d.rep.representData(data)
	d.ser.serialize(node)
	d.rep.representedObjects = map[any]*Node{}
	d.rep.aliasKey = nil
	return nil
}

// EmitEvent is dumper.emit(event).
func (d *DumpStream) EmitEvent(event *Event) (err error) {
	defer catch(&err)
	d.ser.emit(event)
	return nil
}

// Output is stream.getvalue(): a string, or a []byte when Encoding is set.
func (d *DumpStream) Output() any {
	if d.ser.binary {
		return append([]byte{}, d.ser.out...)
	}
	return string(d.ser.out)
}

// Emit is yaml.emit(events, **options) with no stream: the text produced.
func Emit(events []*Event, opts DumpOptions) (text string, err error) {
	defer catch(&err)
	e := newEmitter(&opts, false)
	for _, event := range events {
		e.emit(event)
	}
	return string(e.out), nil
}

// SerializeAll is yaml.serialize_all(nodes, **options): a string, or a []byte when Encoding is set.
func SerializeAll(nodes []*Node, opts DumpOptions) (any, error) {
	d := NewDumpStream(opts)
	if err := d.Open(); err != nil {
		return nil, err
	}
	for _, node := range nodes {
		if err := d.Serialize(node); err != nil {
			return nil, err
		}
	}
	if err := d.Close(); err != nil {
		return nil, err
	}
	return d.Output(), nil
}

// Serialize is yaml.serialize(node, **options).
func Serialize(node *Node, opts DumpOptions) (any, error) { return SerializeAll([]*Node{node}, opts) }

// DumpAll is yaml.dump_all(documents, **options): a string, or a []byte when Encoding is set.
func DumpAll(documents []any, opts DumpOptions) (any, error) {
	d := NewDumpStream(opts)
	if err := d.Open(); err != nil {
		return nil, err
	}
	for _, data := range documents {
		if err := d.Represent(data); err != nil {
			return nil, err
		}
	}
	if err := d.Close(); err != nil {
		return nil, err
	}
	return d.Output(), nil
}

// Dump is yaml.dump(data, **options).
func Dump(data any, opts DumpOptions) (any, error) { return DumpAll([]any{data}, opts) }

// SafeDumpAll is yaml.safe_dump_all(documents, **options).
func SafeDumpAll(documents []any, opts DumpOptions) (any, error) {
	opts.Dumper = SafeDumper
	return DumpAll(documents, opts)
}

// SafeDump is yaml.safe_dump(data, **options).
func SafeDump(data any, opts DumpOptions) (any, error) { return SafeDumpAll([]any{data}, opts) }
