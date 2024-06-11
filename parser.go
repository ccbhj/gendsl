// Package gendsl provides framework a DSL in [Lisp](https://en.wikipedia.org/wiki/Lisp_(programming_language)) style  and allows you to customize your own expressions so that you can integrate it into your own golang application without accessing any lexer or parser.
package gendsl

import (
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

type (
	// ParseContext holds the stateless parser context for a compiled script.
	// It can be reused and re-evaluated with different [gendsl.EvalCtx].
	ParseContext struct {
		p *parser
	}

	// 	OptionList map[string]any
)

var parserTab map[pegRule]func(*ParseContext, *EvalCtx, *node32) (any, error)

func init() {
	parserTab = map[pegRule]func(*ParseContext, *EvalCtx, *node32) (any, error){
		ruleScript:            parseFirstNonSpaceChild,
		ruleValue:             parseFirstNonSpaceChild,
		ruleExpression:        parseExpression,
		ruleIdentifier:        parseIdentifier,
		ruleIdentifierAttr:    parseIdentifierAttr,
		ruleOperator:          parseOperator,
		ruleLiteral:           parseChild,
		ruleBoolLiteral:       parseBoolLiteral,
		ruleFloatLiteral:      parseFloatLiteral,
		ruleIntegerLiteral:    parseIntegerLiteral,
		ruleStringLiteral:     parseStringLiteral,
		ruleLongStringLiteral: parseLongStringLiteral,
		ruleNilLiteral:        parseNilLiteral,
	}
}

// EvalExpr evaluates an `expr` with `env` and return a Value result which could be nil as the result.
func EvalExpr(expr string, env *Env) (Value, error) {
	p, err := MakeParseContext(expr)
	if err != nil {
		return nil, err
	}
	return p.Eval(NewEvalCtx(nil, nil, env))
}

// EvalExprWithData evaluates an `expr` with `env`,
// and allow you to pass a data to use across the entrie evaluation.
func EvalExprWithData(expr string, env *Env, data any) (Value, error) {
	p, err := MakeParseContext(expr)
	if err != nil {
		return nil, err
	}
	c := NewEvalCtx(nil, data, env)
	ret, err := p.Eval(c)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

// MakeParseContext parses the expr and compiles it into an ast.
// You can save the *ParseContext for later usage, since there is no side-affect during evaluation.
func MakeParseContext(expr string) (*ParseContext, error) {
	parser := &parser{
		Buffer: expr,
		Pretty: true,
	}
	if err := parser.Init(); err != nil {
		return nil, err
	}
	if err := parser.Parse(); err != nil {
		pe := new(parseError)
		if errors.Is(err, pe) {
			return nil, &SyntaxError{pe: pe}
		}
		return nil, err
	}
	return &ParseContext{
		p: parser,
	}, nil
}

// Eval evaluates the compiled script with an evalCtx.
// It will panic if evalCtx is nil.
func (c *ParseContext) Eval(evalCtx *EvalCtx) (Value, error) {
	if evalCtx == nil {
		panic("evalCtx cannot be nil")
	}
	ast := c.p.AST()
	v, err := c.parseNode(ast, evalCtx)
	if err != nil {
		return nil, err
	}
	if v == nil {
		return nil, nil
	}
	ret, ok := v.(Value)
	if !ok {
		return nil, errors.Errorf("invalid Value got returned, expecting Value but got %v", v)
	}
	return ret, nil
}

func (c *ParseContext) nodeText(n *node32) string {
	return strings.TrimSpace(string(c.p.buffer[n.begin:n.end]))
}

// PrintTree output the syntax tree to the stdio
func (c *ParseContext) PrintTree() {
	c.p.PrintSyntaxTree()
}

// parseNode parses an ast node.
// make sure that you have registered a parser func in parserTab for the node's rule.
// NOTE that you should use it whenevr you are not sure of the node's type or how to parse it.
func (c *ParseContext) parseNode(node *node32, evalCtx *EvalCtx) (any, error) {
	parser := parserTab[node.pegRule]
	if parser == nil {
		panic("parser for rule " + node.pegRule.String() + " not found")
	}
	return parser(c, evalCtx, node)
}

func (r pegRule) String() string {
	return rul3s[r]
}

func parseFirstNonSpaceChild(c *ParseContext, evalCtx *EvalCtx, node *node32) (any, error) {
	cur := node.up
	// skip the space
	for ; cur != nil && cur.pegRule == ruleSpacing; cur = cur.next {
	}
	return c.parseNode(cur, evalCtx)
}

// parse whatever its child is.
func parseChild(c *ParseContext, evalCtx *EvalCtx, node *node32) (any, error) {
	return c.parseNode(node.up, evalCtx)
}

// parseNodeText return the text of the token node.
func parseNodeText(c *ParseContext, _ *EvalCtx, node *node32) (any, error) {
	return c.nodeText(node), nil
}

// func parseOptionList(c *ParseContext, evalCtx *EvalCtx, node *node32) (any, error) {
// 	opts := make(OptionList, 2)
// 	for node = node.up; node != nil; node = node.next {
// 		optID, optVal, err := parseOption(c, evalCtx, node)
// 		if err != nil {
// 			return nil, err
// 		}
// 		opts[optID] = optVal
// 	}
// 	return opts, nil
// }

// func parseOption(c *ParseContext, evalCtx *EvalCtx, node *node32) (string, any, error) {
// 	var (
// 		id    string
// 		idVal any
// 		val   any
// 		err   error
// 	)
//
// 	node = node.up
// 	idVal, err = c.parseNode(node, evalCtx)
// 	if err != nil {
// 		return "", nil, errors.Errorf("fail to parse option id(%q)", c.nodeText(node))
// 	}
// 	id = idVal.(string)
//
// 	val, err = c.parseNode(node.next, evalCtx)
// 	if err != nil {
// 		return "", nil, errors.Errorf("fail to parse option val(%q) for option #:%s",
// 			c.nodeText(node.next), idVal)
// 	}
// 	return id, val, nil
// }

func readIdentifierText(c *ParseContext, node *node32) string {
	id := strings.TrimSpace(c.nodeText(node))
	return id
}

func parseIdentifier(c *ParseContext, evalCtx *EvalCtx, node *node32) (any, error) {
	return _parseIdentifier(c, evalCtx, node)
}

func _parseIdentifier(c *ParseContext, evalCtx *EvalCtx, node *node32) (Value, error) {
	id := readIdentifierText(c, node)
	v, ok := evalCtx.Lookup(id)
	if !ok {
		return nil, newUnboundedIdentifierError(c, node, id)
	}
	return v, nil
}

func parseIdentifierAttr(c *ParseContext, evalCtx *EvalCtx, node *node32) (any, error) {
	cur := node.up
	val, err := _parseIdentifier(c, evalCtx, cur)
	if err != nil {
		return nil, err
	}

	cur = cur.next
	for ; cur != nil; cur = cur.next {
		idxer, ok := val.Unwrap().(Selector)
		if !ok {
			return nil, evalErrorf(c, cur, "value is not indexable")
		}

		path := readIdentifierText(c, cur.up) // skip the '.'
		v, in := idxer.Select(path)
		if !in {
			return nil, evalErrorf(c, cur, "index(%s) not found for value(type=%v)", path, v)
		}
		val = v
	}
	return val, nil
}

func parseOperator(c *ParseContext, e *EvalCtx, node *node32) (any, error) {
	id := strings.TrimSpace(c.nodeText(node.up))
	v, ok := e.Lookup(id)
	if !ok {
		return nil, evalErrorf(c, node, "unsupported operator %s", id)
	}
	op, ok := v.(Procedure)
	if !ok {
		return nil, evalErrorf(c, node, "<%s> not operator", id)
	}

	return op, nil
}

func parseLongStringLiteral(c *ParseContext, _ *EvalCtx, node *node32) (any, error) {
	text := c.nodeText(node)
	return String(text[3 : len(text)-3]), nil
}

func parseStringLiteral(c *ParseContext, _ *EvalCtx, node *node32) (any, error) {
	text := c.nodeText(node)
	unquoted, err := strconv.Unquote(text)
	if err != nil {
		return nil, evalErrorf(c, node, "invalid string literal")
	}
	return String(unquoted), nil
}

func parseFloatLiteral(c *ParseContext, _ *EvalCtx, node *node32) (any, error) {
	text := c.nodeText(node)
	v, err := strconv.ParseFloat(text, 64)
	if err != nil {
		return nil, evalErrorf(c, node, "invalid float literal %q: %s", text, err)
	}
	return Float(v), nil
}

func parseBoolLiteral(c *ParseContext, _ *EvalCtx, node *node32) (any, error) {
	s := c.nodeText(node)
	if s == "#t" {
		return Bool(true), nil
	}
	return Bool(false), nil
}

func parseNilLiteral(c *ParseContext, _ *EvalCtx, node *node32) (any, error) {
	return Nil{}, nil
}

func parseIntegerLiteral(c *ParseContext, _ *EvalCtx, node *node32) (any, error) {
	var (
		v   any
		err error
	)
	text := c.nodeText(node)
	suffix := text[len(text)-1]
	switch suffix {
	case 'u', 'U':
		v, err = strconv.ParseUint(text[:len(text)-1], 0, 64)
		if err == nil {
			v = Uint(v.(uint64))
		}
	default:
		v, err = strconv.ParseInt(text, 0, 64)
		if err == nil {
			v = Int(v.(int64))
		}
	}

	if err != nil {
		return nil, evalErrorf(c, node, "invalid literal %q: %s", text, err)
	}
	return v, nil
}
