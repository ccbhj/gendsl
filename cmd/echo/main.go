// package main proveides a cmd that reads an expression from input or file and print the result
// or pretty print the ast for debugging. An procedure called "ECHO" that accepts any amount of arguments, print them and return the amount of arguments printed is already provided in the cmd.
package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/ccbhj/gendsl"
)

func main() {
	var (
		printTree = flag.Bool("pt", false, "print tree")
		fromFile  = flag.String("file", "", "read input from file")
	)
	flag.Parse()

	input := flag.Args()[0]
	if fromFile != nil && *fromFile != "" {
		file, err := os.OpenFile(*fromFile, os.O_RDONLY, os.ModePerm)
		if err != nil {
			panic(err)
		}
		defer file.Close()

		bs, err := io.ReadAll(file)
		if err != nil {
			panic(err)
		}
		input = string(bs)
	}

	env := gendsl.NewEnv().WithProcedure("ECHO", gendsl.Procedure{
		Eval: func(_ *gendsl.EvalCtx, args []gendsl.Expr) (gendsl.Value, error) {
			for _, arg := range args {
				v, err := arg.Eval()
				if err != nil {
					return nil, err
				}
				fmt.Println(v.Unwrap())
			}
			return gendsl.Int(len(args)), nil
		},
	})

	pctx, err := gendsl.MakeParseContext(input)
	if err != nil {
		panic(err)
	}
	if printTree != nil && *printTree {
		pctx.PrintTree()
	}

	ret, err := pctx.Eval(gendsl.NewEvalCtx(nil, nil, env))
	if err != nil {
		panic(err)
	}
	fmt.Println(ret.Unwrap())
}
