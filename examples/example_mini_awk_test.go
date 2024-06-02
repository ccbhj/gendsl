package gendsl_test

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

// use optional pattern to set MiniAWK's attribute.
type MiniAWkOpt func(awk *MiniAWK)

// _begin for (BEGIN {expr}) set the begin part of the MiniAWK.
func _begin(_ *gendsl.EvalCtx, args []gendsl.Expr, _ map[string]gendsl.Value) (gendsl.Value, error) {
	return &gendsl.UserData{
		V: func(awk *MiniAWK) {
			awk.Begin = args[0]
		},
	}, nil
}

// _end for (END {expr}) set the end part of the MiniAWK.
func _end(_ *gendsl.EvalCtx, args []gendsl.Expr, _ map[string]gendsl.Value) (gendsl.Value, error) {
	return &gendsl.UserData{
		V: func(awk *MiniAWK) {
			awk.End = args[0]
		},
	}, nil
}

// _pattern for (PATTERN {match_expr} {action_expr}) to register a matcher and it action.
func _pattern(_ *gendsl.EvalCtx, args []gendsl.Expr, _ map[string]gendsl.Value) (gendsl.Value, error) {
	if len(args) < 2 {
		panic("need two or more arguments")
	}
	return &gendsl.UserData{
		V: func(awk *MiniAWK) {
			awk.Patterns = append(awk.Patterns, struct {
				Cond gendsl.Expr
				Then gendsl.Expr
			}{
				Cond: args[0],
				Then: args[1],
			})
		},
	}, nil
}

// _match for (match {pattern} {text}) to report if `text` match the regex expr {pattern}.
func _match(_ *gendsl.EvalCtx, args []gendsl.Expr, _ map[string]gendsl.Value) (gendsl.Value, error) {
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
	result, err := regexp.MatchString(p.Unwrap().(string), s.Unwrap().(string))
	if err != nil {
		return nil, err
	}

	return gendsl.Bool(result), nil
}

// _printf for (print {fmt} {arg...}) to behave like the fmt.Printf(fmt, arg...).
func _printf(_ *gendsl.EvalCtx, args []gendsl.Expr, _ map[string]gendsl.Value) (gendsl.Value, error) {
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

	fmt.Printf(s.Unwrap().(string), arr...)
	return gendsl.Nil{}, nil
}

// awkEnv used in the (awk ...) context.
var awkEnv = gendsl.NewEnv().
	WithProcedure("BEGIN", gendsl.Procedure{Eval: _begin}).
	WithProcedure("END", gendsl.Procedure{Eval: _end}).
	WithProcedure("PATTERN", gendsl.Procedure{Eval: _pattern}).
	WithProcedure("printf", gendsl.Procedure{Eval: _printf}).
	WithProcedure("match", gendsl.Procedure{Eval: _match})

// _awk accept an io.Reader before execution and some option from DSL then run the text parsing.
func _awk(evalCtx *gendsl.EvalCtx, args []gendsl.Expr, _ map[string]gendsl.Value) (gendsl.Value, error) {
	awk := &MiniAWK{}

	// read options
	for _, arg := range args {
		val, err := arg.EvalWithEnv(awkEnv)
		if err != nil {
			return nil, err
		}

		opt, ok := val.Unwrap().(func(*MiniAWK))
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

	// do the actual match-then work.
	scn := bufio.NewScanner(input)
	scn.Split(bufio.ScanLines)
	for scn.Scan() {
		line := scn.Text()
		vars := strings.Split(line, " ")

		// inject variable to local env for the sub-expression.
		localEnv := gendsl.NewEnv().WithString("$0", gendsl.String(line))
		for i, v := range vars {
			localEnv.WithString(fmt.Sprintf("$%d", i+1), gendsl.String(v))
		}

		for _, p := range awk.Patterns {
			ok, err := p.Cond.EvalWithEnv(localEnv)
			if err != nil {
				return nil, err
			}
			if ok.Unwrap().(bool) {
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

	return gendsl.Nil{}, nil
}

func EvalAWK(script string, input io.Reader) error {
	env := gendsl.NewEnv().WithProcedure("awk", gendsl.Procedure{
		Eval: _awk,
	})
	_, err := gendsl.EvalExprWithData(script, env, input)
	if err != nil {
		return err
	}

	return nil
}

// ExampleEvalExprWithData demonstrates a DSL that acts like the unix command 'awk'.
// Inside the (awk ...) block, you can define three sections:
//   - (BEGIN {action}) defines what to do before the text processing starts.
//   - (END {action}) defines what to do after the text processing done.
//   - (PATTERN {pattern} {action}) is the match-then part that tells the MiniAWK to execute {action} when a line of the input matches {pattern} (when {pattern} returns true).
//     More than one (PATTERN ...) expression is allowed and they will be executed one by one.
//     What's more, you can refer each column in a line by "$1" for first column, "$2" for second column, "$n" for n columns and $0 for the whole line.
//     In additional, a (match {regex_pattern} {column}) expression is provided for {pattern}, and (printf {format} {arg}...) for {action}.
func ExampleEvalExprWithData() {
	scripts := `
(awk 
	(BEGIN (printf "Language    FirstAppearAt\n" )) 
	; find out those language that matches ".*C.*", print their name the first appear time.
	(PATTERN (match ".*C.*" $1) (printf "%-8s    %s\n" $1 $3))  
)
	`

	input := `
C Static 1972
C++ Static 1985
Python Dynamic 1991
Ruby Dynamic 1995
	`

	err := EvalAWK(scripts, strings.NewReader(input))
	if err != nil {
		panic(err)
	}
	// Output:
	// Language    FirstAppearAt
	// C           1972
	// C++         1985
	//
}
