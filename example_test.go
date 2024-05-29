package gendsl_test

import (
	"fmt"
	"os"

	"github.com/ccbhj/gendsl"
)

func ExampleEvalExpr() {
	plusOp := func(_ *gendsl.EvalCtx, args []gendsl.Expr) (gendsl.Value, error) {
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
	printlnOp := func(ectx *gendsl.EvalCtx, args []gendsl.Expr) (gendsl.Value, error) {
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
