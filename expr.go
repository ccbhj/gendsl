package gendsl

type (
	EvalOpt struct {
		Env Env
	}
	Expr func(EvalOpt) (any, error)

	EvalFn func(evalCtx *EvalCtx, args []Expr) (any, error)

	Operator struct {
		NArg string
		Eval EvalFn
	}
)

func parseExpression(c *ParseContext, evalCtx *EvalCtx, node *node32) (any, error) {
	var (
		operator string
	)
	cur := node.up

	// ignore the LPAR
	cur = cur.next
	if cur.pegRule != ruleOperator {
		return nil, SyntaxErrorf(c, cur, "expecting an operator in expression, but got %s", cur.pegRule)
	}
	v, err := c.parseNode(cur, evalCtx)
	if err != nil {
		return nil, err
	}
	op, ok := v.(Operator)
	if !ok {
		return nil, SyntaxErrorf(c, node, "<%s> is not an operator", operator)
	}
	cur = cur.next

	operands := make([]Expr, 0)
	for ; cur != nil; cur = cur.next {
		switch cur.pegRule {
		case ruleLPAR, ruleRPAR:
			continue
		case ruleOperand:
			node := cur.up
			operands = append(operands, newExpr(c, evalCtx, node))
		default:
			// will NOT go here
			panic("invalid node in an expression")
		}
	}

	return op.Eval(evalCtx, operands)
}

func newExpr(c *ParseContext, evalCtx *EvalCtx, node *node32) Expr {
	return func(opt EvalOpt) (any, error) {
		env := opt.Env
		if env != nil {
			evalCtx = NewEvalCtx(evalCtx, evalCtx.UserData, env)
		}
		v, err := c.parseNode(node, evalCtx)
		if err != nil {
			return nil, err
		}
		return v, nil
	}
}

func (e Expr) Eval() (any, error) {
	return e(EvalOpt{})
}

func (e Expr) EvalWithEnv(env Env) (any, error) {
	return e(EvalOpt{Env: env})
}
