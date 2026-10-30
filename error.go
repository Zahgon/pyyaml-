// Port of PyYAML's lib/yaml/error.py and of the exception classes of its other modules
// (MIT License; Copyright (c) 2017-2021 Ingy döt Net, Copyright (c) 2006-2016 Kirill Simonov; see LICENSE).

package yaml

import (
	"errors"
	"fmt"
	"strings"

	"github.com/yaml/pyyaml-go/internal/pyval"
)

// Mark is a position in a stream. Name is the stream's name; Index, Line and Column count code points
// from 0; Buffer is the decoded stream with its trailing NUL and Pointer the position in it.
type Mark struct {
	Name    string
	Index   int
	Line    int
	Column  int
	Buffer  []rune
	Pointer int
}

func isSnippetStop(r rune) bool {
	switch r {
	case 0, '\r', '\n', 0x85, 0x2028, 0x2029:
		return true
	}
	return false
}

// pySlice is buffer[start:end] with Python's indexing: a negative index counts from the end, and an
// empty range gives an empty slice. get_snippet reaches a negative end for a small max_length.
func pySlice(buffer []rune, start, end int) []rune {
	clamp := func(i int) int {
		if i < 0 {
			i += len(buffer)
			if i < 0 {
				return 0
			}
		}
		if i > len(buffer) {
			return len(buffer)
		}
		return i
	}
	start, end = clamp(start), clamp(end)
	if start >= end {
		return nil
	}
	return buffer[start:end]
}

func spaces(n int) string {
	if n <= 0 {
		return ""
	}
	return strings.Repeat(" ", n)
}

// Snippet is get_snippet(indent, max_length).
func (m *Mark) Snippet(indent, maxLength int) string {
	half := float64(maxLength)/2 - 1
	head := ""
	start := m.Pointer
	for start > 0 && !isSnippetStop(m.Buffer[start-1]) {
		start--
		if float64(m.Pointer-start) > half {
			head = " ... "
			start += 5
			break
		}
	}
	tail := ""
	end := m.Pointer
	for end < len(m.Buffer) && !isSnippetStop(m.Buffer[end]) {
		end++
		if float64(end-m.Pointer) > half {
			tail = " ... "
			end -= 5
			break
		}
	}
	return spaces(indent) + head + pyval.FromRunes(pySlice(m.Buffer, start, end)) + tail + "\n" +
		spaces(indent+m.Pointer-start+len(head)) + "^"
}

// String is str(mark).
func (m *Mark) String() string {
	return fmt.Sprintf("  in \"%s\", line %d, column %d:\n%s", m.Name, m.Line+1, m.Column+1, m.Snippet(4, 75))
}

// MarkedError is a MarkedYAMLError. Class is ScannerError, ParserError, ComposerError or
// ConstructorError; the other fields are its attributes, nil where Python has None.
type MarkedError struct {
	Class       string
	Context     *string
	ContextMark *Mark
	Problem     *string
	ProblemMark *Mark
	Note        *string
}

// Error is str(exc).
func (e *MarkedError) Error() string {
	var lines []string
	if e.Context != nil {
		lines = append(lines, *e.Context)
	}
	if e.ContextMark != nil && (e.Problem == nil || e.ProblemMark == nil ||
		e.ContextMark.Name != e.ProblemMark.Name || e.ContextMark.Line != e.ProblemMark.Line ||
		e.ContextMark.Column != e.ProblemMark.Column) {
		lines = append(lines, e.ContextMark.String())
	}
	if e.Problem != nil {
		lines = append(lines, *e.Problem)
	}
	if e.ProblemMark != nil {
		lines = append(lines, e.ProblemMark.String())
	}
	if e.Note != nil {
		lines = append(lines, *e.Note)
	}
	return strings.Join(lines, "\n")
}

// ReaderError is PyYAML's ReaderError. Character is the code point or the undecodable byte.
type ReaderError struct {
	Name      string
	Position  int
	Character int
	Encoding  string
	Reason    string
}

// Error is str(exc). Python indexes a bytes object for an undecodable byte, which gives an int, so
// both kinds of error print the same way.
func (e *ReaderError) Error() string {
	return fmt.Sprintf("unacceptable character #x%04x: %s\n  in \"%s\", position %d", e.Character, e.Reason, e.Name, e.Position)
}

// YAMLError is a YAMLError without marks. Class is EmitterError, SerializerError or RepresenterError
// and Msg is str(exc).
type YAMLError struct {
	Class string
	Msg   string
}

// Error is str(exc).
func (e *YAMLError) Error() string { return e.Msg }

// PyError is an exception Python itself raises while PyYAML runs. Class is the name of its class and
// Msg is str(exc).
type PyError struct {
	Class string
	Msg   string
}

// Error is str(exc).
func (e *PyError) Error() string { return e.Msg }

// pyError converts an error of internal/pyval.
func pyError(err error) error {
	var pe *pyval.Error
	if errors.As(err, &pe) {
		return &PyError{Class: pe.Class, Msg: pe.Msg}
	}
	return err
}

func ptr(s string) *string { return &s }

func markedError(class string, context *string, contextMark *Mark, problem string, problemMark *Mark) *MarkedError {
	return &MarkedError{Class: class, Context: context, ContextMark: contextMark, Problem: &problem, ProblemMark: problemMark}
}

// raised carries an exception up through the state machines to the entry points in api.go.
type raised struct{ err error }

func raise(err error) { panic(raised{err}) }

// catch turns a raised exception into the error of an entry point.
func catch(err *error) {
	if r := recover(); r != nil {
		rr, ok := r.(raised)
		if !ok {
			panic(r)
		}
		*err = rr.err
	}
}
