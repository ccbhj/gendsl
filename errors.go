package gendsl

import (
	"fmt"

	"github.com/pkg/errors"
)

// SyntaxError got thrown when a parsing error found.
type SyntaxError struct {
	pe *parseError
}

func (e *SyntaxError) Error() string {
	return e.pe.Error()
}

// EvaluateError got thrown during the evaluation.
type EvaluateError struct {
	// pos where the expression cannot be evaluated.
	BeginLine, EndLine int
	BeginSym, EndSym   int
	cause              error
}

func (s *EvaluateError) Error() string {
	return fmt.Sprintf("evaluate error (line %v symbol %v - line %v symbol %v): %s",
		s.BeginLine, s.BeginSym, s.EndLine, s.EndSym, s.cause)
}

func (s *EvaluateError) Unwrap() error {
	return s.cause
}

func (s *EvaluateError) Cause() error {
	return s.cause
}

func evalErrorf(c *ParseContext, node *node32, f string, args ...any) error {
	pos := translatePositions(c.p.buffer, []int{int(node.begin), int(node.end)})
	beg, end := pos[int(node.begin)], pos[int(node.end)]

	return &EvaluateError{
		BeginLine: beg.line,
		EndLine:   end.line,
		BeginSym:  beg.symbol,
		EndSym:    end.symbol,
		cause:     errors.Errorf(f, args...),
	}
}

// UnboundedIdentifierError got thrown when an identifier cannot be found in env.
type UnboundedIdentifierError struct {
	ID string // ID that cannot be found
	// position where the unbounded id found
	BeginLine, EndLine int
	BeginSym, EndSym   int
}

func (s *UnboundedIdentifierError) Error() string {
	return fmt.Sprintf("unbounded variable (line %v symbol %v - line %v symbol %v): %s",
		s.BeginLine, s.BeginSym, s.EndLine, s.EndSym, s.ID)
}

func newUnboundedIdentifierError(c *ParseContext, node *node32, id string) error {
	pos := translatePositions(c.p.buffer, []int{int(node.begin), int(node.end)})
	beg, end := pos[int(node.begin)], pos[int(node.end)]

	return &UnboundedIdentifierError{
		BeginLine: beg.line,
		EndLine:   end.line,
		BeginSym:  beg.symbol,
		EndSym:    end.symbol,
		ID:        id,
	}
}
