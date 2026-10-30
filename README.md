PyYAML
======

A full-featured YAML processing framework for Go

This is the Go library that replaces the Python package this repository used to hold. It is a port of
PyYAML's pure Python implementation: the reader, scanner, parser, composer, constructor, resolver,
representer, serializer and emitter, with the Python 3.12 value semantics the library relies on.

## Installation

    go get github.com/yaml/pyyaml-go

The module has no dependencies outside the Go standard library.

## Usage

Load a document:

    value, err := yaml.SafeLoad("a: 1\nb: [x, y]\n")

Dump a value:

    text, err := yaml.Dump(value, yaml.NewDumpOptions())

`Load`, `LoadAll`, `Compose`, `ComposeAll` take a `Loader` (`BaseLoader`, `SafeLoader` or
`FullLoader`); `Dump`, `DumpAll`, `Serialize`, `SerializeAll` and `Emit` take `DumpOptions`, whose
fields are the keyword arguments of the Python functions and whose zero value, through
`NewDumpOptions`, is their defaults. If you do not trust the input, use `SafeLoad`.

`Scan` and `Parse` give the tokens and the events of a stream, and `Repr` renders any loaded value as
Python's `repr()` does.

## Values

A YAML document loads into the Go values that stand for Python's: `string` for `str`, `*big.Int` for
`int`, `float64` for `float`, `bool`, `nil` for `None`, `[]byte` for `bytes`, and `*List`, `*Tuple`,
`*Dict`, `*PySet`, `*Date`, `*DateTime` and `*Complex` for the containers and the other types. `*Dict`
keeps its insertion order and compares keys as Python does, so `1`, `1.0` and `true` are one key. Two
references to one of the pointer types are one object: an anchor and its aliases load as one value, and
a value shared by identity is dumped once with an anchor.

## Testing

    go test ./...

The tests in `recorded_test.go` and `internal/*/recorded_test.go` are recordings of the original Python
library and of CPython 3.12 themselves; they are generated, not written by hand.

## Further Information

* [PyYAML homepage](https://github.com/yaml/pyyaml).

* [PyYAML tutorial and reference](http://pyyaml.org/wiki/PyYAMLDocumentation).

* Submit bug reports and feature requests to the
  [PyYAML bug tracker](https://github.com/yaml/pyyaml/issues).
