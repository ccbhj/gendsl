package gendsl

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

var ErrUnboundedIdentifier = errors.New("unbounded identifier")

type (
	ParseContext struct {
		p   *Parser
		env map[string]any // global env, will never modified
	}

	EvalCtx struct {
		parent   *EvalCtx
		env      Env // env for evaluation
		UserData interface{}
	}

	OptionList map[string]any
)

var parserTab map[pegRule]func(*ParseContext, *EvalCtx, *node32) (any, error)

func init() {
	parserTab = map[pegRule]func(*ParseContext, *EvalCtx, *node32) (any, error){
		ruleScript:         parseChild,
		ruleExpression:     parseExpression,
		ruleIdentifier:     parseIdentifier,
		ruleOperator:       parseOperator,
		ruleLiteral:        parseChild,
		ruleBoolLiteral:    parseBoolLiteral,
		ruleFloatLiteral:   parseFloatLiteral,
		ruleIntegerLiteral: parseIntegerLiteral,
		ruleStringLiteral:  parseStringLiteral,
	}
}

func EvalExpr(expr string, env Env) (any, error) {
	p, err := MakeParseContext(expr, env)
	if err != nil {
		return nil, err
	}
	return p.Run(NewEvalCtx(nil, nil, p.env))
}

func EvalExprWithInput(expr string, env Env, data any) (any, error) {
	p, err := MakeParseContext(expr, env)
	if err != nil {
		return nil, err
	}
	c := NewEvalCtx(nil, data, p.env)
	ret, err := p.Run(c)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func MakeParseContext(expr string, globalEnv Env) (*ParseContext, error) {
	parser := &Parser{
		Buffer: expr,
		Pretty: true,
	}
	if err := parser.Init(); err != nil {
		return nil, err
	}
	if err := parser.Parse(); err != nil {
		return nil, err
	}
	return &ParseContext{
		p:   parser,
		env: globalEnv,
	}, nil
}

func NewEvalCtx(p *EvalCtx, userData any, env map[string]any) *EvalCtx {
	return &EvalCtx{
		parent:   p,
		env:      env,
		UserData: userData,
	}
}

func (e *EvalCtx) lookup(id string) (any, bool) {
	v, ok := e.env[id]
	if ok {
		return v, ok
	}

	if e.parent == nil {
		return nil, false
	}

	return e.parent.lookup(id)
}

func (c *ParseContext) Run(evalCtx *EvalCtx) (any, error) {
	ast := c.p.AST()
	v, err := c.parseNode(ast, evalCtx)
	if err != nil {
		return nil, err
	}
	return v, nil
}

func (c *ParseContext) PrintAst() {
	c.p.PrintSyntaxTree()
}

func (c *ParseContext) nodeText(n *node32) string {
	return strings.TrimSpace(string(c.p.buffer[n.begin:n.end]))
}

func (c *ParseContext) readChars(node *node32) (string, error) {
	buf := &strings.Builder{}
	for node.next != nil {
		switch node.pegRule {
		case ruleLetterOrDigit, ruleLetter:
			buf.WriteString(c.nodeText(node))
		default:
			return "", SyntaxErrorf(c, node, "expecting a digit or char")
		}
		node = node.next
	}

	return buf.String(), nil
}

func (c *ParseContext) PrintTree() {
	c.p.PrintSyntaxTree()
}

type SyntaxError struct {
	beginLine, endLine int
	beginSym, endSym   int
	cause              error
}

func (s *SyntaxError) Error() string {
	return fmt.Sprintf("invalid syntax from line %d, symbol %d to line %d, symbol %d: %s",
		s.beginLine, s.beginSym, s.endLine, s.endSym, s.cause)
}

func (s *SyntaxError) Unwrap() error {
	return s.cause
}

func (s *SyntaxError) Cause() error {
	return s.cause
}

func SyntaxErrorf(c *ParseContext, node *node32, f string, args ...any) error {
	pos := translatePositions(c.p.buffer, []int{int(node.begin), int(node.end)})
	beg, end := pos[int(node.begin)], pos[int(node.end)]

	return &SyntaxError{
		beginLine: beg.line,
		endLine:   end.line,
		beginSym:  beg.symbol,
		endSym:    end.symbol,
		cause:     errors.Errorf(f, args...),
	}
}

// parsers
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

// parse whatever its child is
func parseChild(c *ParseContext, evalCtx *EvalCtx, node *node32) (any, error) {
	return c.parseNode(node.up, evalCtx)
}

// parseNodeText return the text of the token node
func parseNodeText(c *ParseContext, evalCtx *EvalCtx, node *node32) (any, error) {
	return c.nodeText(node), nil
}

func parseOptionList(c *ParseContext, evalCtx *EvalCtx, node *node32) (any, error) {
	opts := make(OptionList, 2)
	for node = node.up; node != nil; node = node.next {
		optID, optVal, err := parseOption(c, evalCtx, node)
		if err != nil {
			return nil, err
		}
		opts[optID] = optVal
	}
	return opts, nil
}

func parseOption(c *ParseContext, evalCtx *EvalCtx, node *node32) (string, any, error) {
	var (
		id    string
		idVal any
		val   any
		err   error
	)

	node = node.up
	idVal, err = c.parseNode(node, evalCtx)
	if err != nil {
		return "", nil, errors.Errorf("fail to parse option id(%q)", c.nodeText(node))
	}
	id = idVal.(string)

	val, err = c.parseNode(node.next, evalCtx)
	if err != nil {
		return "", nil, errors.Errorf("fail to parse option val(%q) for option #:%s",
			c.nodeText(node.next), idVal)
	}
	return id, val, nil
}

func parseIdentifier(c *ParseContext, evalCtx *EvalCtx, node *node32) (any, error) {
	id := strings.TrimSpace(c.nodeText(node))
	v, ok := evalCtx.lookup(id)
	if !ok {
		return nil, errors.WithMessagef(ErrUnboundedIdentifier, "%s", id)
	}
	return v, nil
}

func parseOperator(c *ParseContext, e *EvalCtx, node *node32) (any, error) {
	id, err := c.readChars(node.up.up)
	if err != nil {
		return nil, err
	}
	id = strings.TrimSpace(id)
	v, ok := e.lookup(id)
	if !ok {
		return nil, SyntaxErrorf(c, node, "unsupported operator <%s>", id)
	}
	op, ok := v.(Operator)
	if !ok {
		return nil, SyntaxErrorf(c, node, "<%s> not operator", id)
	}

	return op, nil
}

func parseStringLiteral(c *ParseContext, _ *EvalCtx, node *node32) (any, error) {
	unquoted, err := strconv.Unquote(c.nodeText(node))
	if err != nil {
		return nil, err
	}
	return unquoted, nil
}

func parseFloatLiteral(c *ParseContext, _ *EvalCtx, node *node32) (any, error) {
	text := c.nodeText(node)
	v, err := strconv.ParseFloat(text, 64)
	if err != nil {
		return nil, SyntaxErrorf(c, node, "invalid float literal %q: %s", text, err)
	}
	return v, nil
}

func parseBoolLiteral(c *ParseContext, _ *EvalCtx, node *node32) (any, error) {
	s := c.nodeText(node)
	if s == "#t" {
		return true, nil
	}
	return false, nil
}

func parseIntegerLiteral(c *ParseContext, _ *EvalCtx, node *node32) (any, error) {
	var (
		v   any
		err error
	)
	text := c.nodeText(node)
	suffix := text[len(text)-1]
	switch suffix {
	case 'f', 'F':
		v, err = strconv.ParseFloat(text[:len(text)-1], 64)
	case 'u', 'U':
		v, err = strconv.ParseUint(text[:len(text)-1], 10, 64)
	default:
		v, err = strconv.ParseInt(text, 10, 64)
	}

	if err != nil {
		return nil, SyntaxErrorf(c, node, "invalid literal %q: %s", text, err)
	}
	return v, nil
}
