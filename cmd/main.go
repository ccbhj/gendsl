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

	input := os.Args[1]
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
	pctx, err := gendsl.MakeParseContext(input, nil)
	if err != nil {
		panic(err)
	}
	if printTree != nil && *printTree {
		pctx.PrintTree()
	}
	result, err := pctx.Run(nil)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%+v\n", result)
}
