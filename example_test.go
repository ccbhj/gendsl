package gendsl_test

import (
	"errors"
	"fmt"
	"os"
	"strings"

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

func ExampleExpr_Text() {
	printlnOp := func(ectx *gendsl.EvalCtx, args []gendsl.Expr, _ map[string]gendsl.Value) (gendsl.Value, error) {
		for _, arg := range args {
			v, err := arg.Eval()
			if err != nil {
				return nil, err
			}
			fmt.Println(v.Unwrap())
		}
		return gendsl.Nil{}, nil
	}

	letOp := func(ectx *gendsl.EvalCtx, args []gendsl.Expr, _ map[string]gendsl.Value) (gendsl.Value, error) {
		nameExpr := args[0]
		if nameExpr.Type() != gendsl.ExprTypeIdentifier {
			return nil, errors.New("expecting an identifier")
		}

		name := strings.TrimSpace(nameExpr.Text())

		val, err := args[1].Eval()
		if err != nil {
			return nil, err
		}

		return args[2].EvalWithEnv(ectx.Env().WithValue(name, val))
	}

	script := `
(LET foo "bar" 
	(PRINTLN foo))
	`

	_, err := gendsl.EvalExpr(script,
		gendsl.NewEnv().WithInt("ONE", 1).
			WithProcedure("LET", gendsl.Procedure{
				Eval: gendsl.CheckNArgs("3", letOp),
			}).
			WithProcedure("PRINTLN", gendsl.Procedure{
				Eval: gendsl.CheckNArgs("+", printlnOp),
			}),
	)

	if err != nil {
		panic(err)
	}
	// Output:
	// bar
}
