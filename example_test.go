package gendsl_test

import (
	"fmt"
	"os"

	"github.com/ccbhj/gendsl"
)

func ExampleEvalExpr() {
	plusOp := func(_ *gendsl.EvalCtx, args []gendsl.Expr, _ map[string]gendsl.Value) (gendsl.Value, error) {
		var ret gendsl.Int
		for _, arg := range args {
			v, err := arg.Eval()
			if err != nil {
				return nil, err
			}
			if v.Type() == gendsl.ValueTypeInt {
				ret += v.(gendsl.Int)
			}
		}

		return ret, nil
	}

	result, err := gendsl.EvalExpr("(PLUS ONE $TWO @THREE 4)",
		gendsl.NewEnv().WithInt("ONE", 1).
			WithInt("$TWO", 2).
			WithInt("@THREE", 3).
			WithProcedure("PLUS", gendsl.Procedure{
				Eval: gendsl.CheckNArgs("+", plusOp),
			}),
	)

	if err != nil {
		panic(err)
	}
	fmt.Println(result.Unwrap())
	// Output: 10
}

func ExampleEvalExprWithData() {
	printlnOp := func(ectx *gendsl.EvalCtx, args []gendsl.Expr, _ map[string]gendsl.Value) (gendsl.Value, error) {
		output := ectx.UserData.(*os.File)
		for _, arg := range args {
			v, err := arg.Eval()
			if err != nil {
				return nil, err
			}
			fmt.Fprintln(output, v.Unwrap())
		}
		return gendsl.Nil{}, nil
	}

	out := os.Stdout
	_, err := gendsl.EvalExprWithData("(PRINTLN ONE $TWO @THREE 4)",
		gendsl.NewEnv().WithInt("ONE", 1).
			WithInt("$TWO", 2).
			WithInt("@THREE", 3).
			WithProcedure("PRINTLN", gendsl.Procedure{
				Eval: gendsl.CheckNArgs("+", printlnOp),
			}),
		out,
	)

	if err != nil {
		panic(err)
	}
	// Output:
	// 1
	// 2
	// 3
	// 4
}

func ExampleTest() {
	type Case struct {
		Cond gendsl.Expr
		Then gendsl.Expr
	}
	_case := func(ectx *gendsl.EvalCtx, args []gendsl.Expr, _ map[string]gendsl.Value) (gendsl.Value, error) {
		return &gendsl.UserData{V: Case{args[0], args[1]}}, nil
	}

	_switch := func(ectx *gendsl.EvalCtx, args []gendsl.Expr, _ map[string]gendsl.Value) (gendsl.Value, error) {
		var ret gendsl.Value
		env := ectx.Env().Clone().WithProcedure("CASE", gendsl.Procedure{
			Eval: gendsl.CheckNArgs("2", _case),
		})
		expect, err := args[0].Eval()
		if err != nil {
			return nil, err
		}
		for _, arg := range args[1:] {
			cv, err := arg.EvalWithEnv(env)
			if err != nil {
				return nil, err
			}
			c, ok := cv.Unwrap().(Case)
			if !ok {
				panic("expecting a cas ")
			}

			cond, err := c.Cond.Eval()
			if err != nil {
				return nil, err
			}
			if cond == expect {
				v, err := c.Then.Eval()
				if err != nil {
					return nil, err
				}
				ret = v
			}
		}

		return ret, nil
	}

	script := `
(SWITCH "FOO"
	(CASE "BAR" "no")
	(CASE "FOO" "yes")
)
	`

	val, err := gendsl.EvalExprWithData(script,
		gendsl.NewEnv().
			WithProcedure("SWITCH", gendsl.Procedure{
				Eval: gendsl.CheckNArgs("+", _switch),
			}),
		nil,
	)

	if err != nil {
		panic(err)
	}
	println(val.Unwrap().(string))
	// output: yes
}
