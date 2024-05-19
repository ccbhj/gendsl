package awk

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/ccbhj/gendsl"
)

type MiniAWK struct {
	Begin    gendsl.Expr
	End      gendsl.Expr
	Patterns []struct {
		Cond gendsl.Expr
		Then gendsl.Expr
	}
}

// use optional pattern to set MiniAWk's attribute

type MiniAWkOpt func(awk *MiniAWK)

// _begin for (BEGIN {expr}) set the begin part of the MiniAWk
func _begin(_ *gendsl.EvalCtx, args []gendsl.Expr) (any, error) {
	return func(awk *MiniAWK) {
		awk.Begin = args[0]
	}, nil
}

// _end for (END {expr}) set the end part of the MiniAWk
func _end(_ *gendsl.EvalCtx, args []gendsl.Expr) (any, error) {
	return func(awk *MiniAWK) {
		awk.End = args[0]
	}, nil
}

// _pattern for (PATTERN {match_expr} {action_expr}) to register a matcher and it action
func _pattern(_ *gendsl.EvalCtx, args []gendsl.Expr) (any, error) {
	if len(args) < 2 {
		panic("need two or more arguments")
	}
	return func(awk *MiniAWK) {
		awk.Patterns = append(awk.Patterns, struct {
			Cond gendsl.Expr
			Then gendsl.Expr
		}{
			Cond: args[0],
			Then: args[1],
		})
	}, nil
}

// _match for (match {pattern} {text}) to return true if `text` match the regex expr {pattern} and false otherwise
func _match(_ *gendsl.EvalCtx, args []gendsl.Expr) (any, error) {
	if len(args) < 1 {
		panic("need one or more arguments")
	}
	p, err := args[0].Eval()
	if err != nil {
		panic(err)
	}
	s, err := args[1].Eval()
	if err != nil {
		panic(err)
	}
	return regexp.MatchString(p.(string), s.(string))
}

// _printf for (print {fmt} {arg...}) to behave like the fmt.Printf(fmt, arg...)
func _printf(_ *gendsl.EvalCtx, args []gendsl.Expr) (any, error) {
	if len(args) < 1 {
		panic("need one or more arguments")
	}
	s, err := args[0].Eval()
	if err != nil {
		panic(err)
	}

	arr := make([]any, 0, 2)
	for _, arg := range args[1:] {
		v, err := arg.Eval()
		if err != nil {
			return nil, err
		}
		arr = append(arr, v)
	}

	fmt.Printf(s.(string), arr...)
	return nil, nil
}

// awkEnv used in the (awk ...) context
var awkEnv = gendsl.NewEnv().
	WithOperator("BEGIN", gendsl.Operator{Eval: _begin}).
	WithOperator("END", gendsl.Operator{Eval: _end}).
	WithOperator("PATTERN", gendsl.Operator{Eval: _pattern}).
	WithOperator("printf", gendsl.Operator{Eval: _printf}).
	WithOperator("match", gendsl.Operator{Eval: _match})

// _awk accept an io.Reader before execution and some option from DSL then run the text parsing
func _awk(evalCtx *gendsl.EvalCtx, args []gendsl.Expr) (any, error) {
	awk := &MiniAWK{}

	// read options
	for _, arg := range args {
		val, err := arg.EvalWithEnv(awkEnv)
		if err != nil {
			return nil, err
		}

		opt, ok := val.(func(*MiniAWK))
		if !ok {
			panic("expecting an option func")
		}
		opt(awk)
	}

	// io.Reader should be provided before execution
	input, ok := evalCtx.UserData.(io.Reader)
	if !ok {
		panic("expecting an io.Reader")
	}

	// perform the BEGIN part if any
	if awk.Begin != nil {
		_, err := awk.Begin.Eval()
		if err != nil {
			panic(err)
		}
	}

	// do the actual match-then work
	scn := bufio.NewScanner(input)
	scn.Split(bufio.ScanLines)
	for scn.Scan() {
		line := scn.Text()
		vars := strings.Split(line, " ")
		localEnv := gendsl.NewEnv().WithString("$0", line)
		for i, v := range vars {
			localEnv.WithString(fmt.Sprintf("$%d", i+1), v)
		}

		for _, p := range awk.Patterns {
			ok, err := p.Cond.EvalWithEnv(localEnv)
			if err != nil {
				return nil, err
			}
			if ok.(bool) {
				_, err := p.Then.EvalWithEnv(localEnv)
				if err != nil {
					return nil, err
				}
			}
		}
	}

	// perform the BEGIN part if any
	if awk.End != nil {
		_, err := awk.End.Eval()
		if err != nil {
			return nil, err
		}
	}

	return awk, nil
}

func EvalAWK(script string, input io.Reader) error {
	env := map[string]any{
		"awk": gendsl.Operator{
			Eval: _awk,
		},
	}

	_, err := gendsl.EvalExprWithInput(script, env, input)
	if err != nil {
		return err
	}

	return nil
}
