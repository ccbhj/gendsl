package gendsl

import (
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

type (
	// EvalOpt for some options to control the evaluate behavior
	EvalOpt struct {
		Env *Env
	}
	// Expr wraps the evalution of an ast node, or in another word, an expression,
	// It allows you control when the evalution can be done, or the env for evalution, so
	// that you can program your procedure to act like a macro.
	Expr func(EvalOpt) (Value, error)

	// ProcedureFn specify the behavior of an [gendsl.Procedure].
	// `evalCtx` carry some information that might be used during the evaluation, see [gendsl.EvalCtx]
	ProcedureFn func(evalCtx *EvalCtx, args []Expr) (Value, error)
)

func parseExpression(c *ParseContext, evalCtx *EvalCtx, node *node32) (any, error) {
	cur := node.up
	cur = cur.next                   // ignore the LPAR
	if cur.pegRule != ruleOperator { // won't go into here, the parser will make sure of it
		return nil, evalErrorf(c, cur, "expecting an operator in expression, but got %s", cur.pegRule)
	}
	v, err := c.parseNode(cur, evalCtx)
	if err != nil {
		return nil, err
	}
	op, ok := v.(Procedure)
	if !ok {
		return nil, evalErrorf(c, node, "<%s> is not an procedure", v)
	}
	if op.Eval == nil {
		return nil, evalErrorf(c, node, "procedure <%s> not provide an evaluate function", v)
	}

	operands := make([]Expr, 0)
	for cur = cur.next; cur != nil; cur = cur.next {
		switch cur.pegRule {
		case ruleLPAR, ruleRPAR:
			continue
		case ruleValue:
			node := cur.up
			operands = append(operands, newOperand(c, evalCtx, node))
		default:
			// will NOT go here, parser will make sure of it
			panic("invalid node in an expression")
		}
	}

	return op.Eval(evalCtx, operands)
}

func newOperand(c *ParseContext, evalCtx *EvalCtx, node *node32) Expr {
	return func(opt EvalOpt) (Value, error) {
		env := opt.Env
		if env != nil {
			evalCtx = NewEvalCtx(evalCtx, evalCtx.UserData, env)
		}
		v, err := c.parseNode(node, evalCtx)
		if err != nil {
			return nil, err
		}
		tv, ok := v.(Value)
		if !ok {
			return nil, evalErrorf(c, node, "expression should return a Value, but got %v", v)
		}
		return tv, nil
	}
}

// Eval evaluate an [gendsl.Expr], return the result of this expression.
//
// These errors might be returned:
//   - [gendsl.SyntaxError] - when a syntax error is found in this expression
//   - [gendsl.UnboundedIdentifierError] - when an undefined id is used in this expression.
func (e Expr) Eval() (Value, error) {
	return e(EvalOpt{})
}

// EvalWithEnv evaluate an [gendsl.Expr] with a new env, return the result of this expression.
// We will lookup an identifier in `env` first, and we will look it up again in the parent env when its value is not found in `env`.
//
// These errors might be returned:
//   - [gendsl.SyntaxError] - when a syntax error is found in this expression
//   - [gendsl.UnboundedIdentifierError] - when an undefined id is used in this expression.
func (e Expr) EvalWithEnv(env *Env) (Value, error) {
	return e(EvalOpt{Env: env})
}

// CheckNArgs check the amount of the operands for a procedure by wrapping an EvalFn.
//
// `nargs` specify the pattern of the amount of operands:
//   - "+" to accept one or more than one operands
//   - "*" or "" to accept any amount of operands
//   - "?" to accept no or one operand
//   - "n" for whatever number the strconv.Atoi can accept, to check the exact amount of the operands
func CheckNArgs(nargs string, evalFn ProcedureFn) ProcedureFn {
	switch strings.TrimSpace(nargs) {
	case "+": // one or more args
		return func(evalCtx *EvalCtx, args []Expr) (Value, error) {
			if len(args) < 1 {
				return nil, errors.Errorf("expecting one or more argument, but got %d", len(args))
			}
			return evalFn(evalCtx, args)
		}
	case "?": // one or no arg
		return func(evalCtx *EvalCtx, args []Expr) (Value, error) {
			if len(args) > 1 {
				return nil, errors.Errorf("expecting one or no argument, but got %d", len(args))
			}
			return evalFn(evalCtx, args)
		}
	case "*", "": // one or no arg
		return evalFn
	}

	n, err := strconv.Atoi(nargs)
	if err != nil {
		panic(errors.Errorf("invalid nargs(%q) passed to CheckNArgs", nargs))
	}
	return func(evalCtx *EvalCtx, args []Expr) (Value, error) {
		if len(args) != n {
			return nil, errors.Errorf("expecting %d argument(s), but got %d", n, len(args))
		}
		return evalFn(evalCtx, args)
	}
}

// EvalCtx holds some information used for evaluation.
type EvalCtx struct {
	parent   *EvalCtx // EvalCtx from the outter scope, nil for top level scope
	env      *Env     // env for current scope
	UserData any      // UserData that is used across the entire script evaluation
}

// NewEvalCtx creates a new EvalCtx with `p` as the output scope EvalCtx(nil is allowed),
// `userData` is argument used across the whole evaluation,
// `env` as the env for current scope evaluation, nil is allowed here and an empty env will be created for it.
func NewEvalCtx(p *EvalCtx, userData any, env *Env) *EvalCtx {
	if env == nil {
		env = NewEnv()
	}
	return &EvalCtx{
		parent:   p,
		env:      env,
		UserData: userData,
	}
}

func (e *EvalCtx) Derive(newEnv *Env) *EvalCtx {
	return NewEvalCtx(e, e.UserData, newEnv)
}

func (e *EvalCtx) Env() *Env {
	return e.env.Clone()
}

// Lookup looks up an identifier in the current env, and try to look it up in its outter scope recurssively if not found.
func (e *EvalCtx) Lookup(id string) (any, bool) {
	if e == nil || e.env == nil {
		return nil, false
	}
	v, ok := e.env.Lookup(id)
	if ok {
		return v, ok
	}

	if e.parent == nil {
		return nil, false
	}

	return e.parent.Lookup(id)
}

// OutScopeEvalCtx returns [gendsl.EvalCtx] from the outter scope.
func (e *EvalCtx) OutScopeEvalCtx() *EvalCtx {
	return e.parent
}
