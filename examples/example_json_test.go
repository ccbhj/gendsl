package gendsl_test

import (
	"encoding/json"
	"fmt"
	"time"

	. "github.com/onsi/gomega"
	"github.com/pkg/errors"

	"github.com/ccbhj/gendsl"
)

type KV struct {
	K string
	V any
}

// only in json env would these operators can be accessed
var jsonOps = gendsl.NewEnv().
	WithProcedure("if", gendsl.Procedure{Eval: _if}).
	WithProcedure("array", gendsl.Procedure{Eval: _array}).
	WithProcedure("kv", gendsl.Procedure{Eval: _kv}).
	WithProcedure("dict", gendsl.Procedure{Eval: _dict})

func _json(evalCtx *gendsl.EvalCtx, args []gendsl.Expr, _ map[string]gendsl.Value) (gendsl.Value, error) {
	ret := make(map[string]any)
	for _, arg := range args {
		v, err := arg.EvalWithEnv(jsonOps) // inject some operators in this context
		if err != nil {
			return nil, err
		}
		kv, ok := v.Unwrap().(KV)
		if !ok {
			panic("expecting a kv")
		}
		ret[kv.K] = kv.V
	}

	b, err := json.Marshal(ret)
	if err != nil {
		return nil, err
	}

	return &gendsl.UserData{V: b}, nil
}

// _kv evaluate a key and its value then assemble them to a KV struct
func _kv(_ *gendsl.EvalCtx, args []gendsl.Expr, _ map[string]gendsl.Value) (gendsl.Value, error) {
	key, err := args[0].Eval()
	if err != nil {
		return nil, err
	}
	if key.Type() != gendsl.ValueTypeString {
		return nil, errors.Errorf("key is not string, but %+v", key)
	}

	val, err := args[1].Eval()
	if err != nil {
		return nil, err
	}
	// return &gendsl.UserData{} to wrap any data into gendsl.Value
	return &gendsl.UserData{V: KV{K: string(key.(gendsl.String)), V: val.Unwrap()}}, nil
}

// _array evaluate any data and assemble them to a array
func _array(_ *gendsl.EvalCtx, args []gendsl.Expr, _ map[string]gendsl.Value) (gendsl.Value, error) {
	val := make([]any, 0, 1)
	for _, arg := range args {
		v, err := arg.Eval()
		if err != nil {
			return nil, err
		}
		val = append(val, v.Unwrap())
	}

	// return &gendsl.UserData{} to wrap any data into gendsl.Value
	return &gendsl.UserData{V: val}, nil
}

// _dict evaluate any KV pair and assemble them to a dict(map[string]any)
func _dict(_ *gendsl.EvalCtx, args []gendsl.Expr, _ map[string]gendsl.Value) (gendsl.Value, error) {
	val := make(map[string]any)
	for _, arg := range args {
		v, err := arg.Eval()
		if err != nil {
			return nil, err
		}
		kv, ok := v.Unwrap().(KV)
		if !ok {
			return nil, errors.New("expecting a kv")
		}
		val[kv.K] = kv.V
	}

	return &gendsl.UserData{V: val}, nil
}

func _laterThan(_ *gendsl.EvalCtx, args []gendsl.Expr, _ map[string]gendsl.Value) (gendsl.Value, error) {
	left, err := args[0].Eval()
	if err != nil {
		return nil, err
	}
	right, err := args[1].Eval()
	if err != nil {
		return nil, err
	}

	return gendsl.Bool(left.Unwrap().(int64) > right.Unwrap().(int64)), nil
}

func _year() gendsl.Int {
	return gendsl.Int(time.Now().Year())
}

// _if check the condition of the first argument,
// if condition returns not nil, evaluate the second argument, otherwise the third argument
func _if(_ *gendsl.EvalCtx, args []gendsl.Expr, _ map[string]gendsl.Value) (gendsl.Value, error) {
	condEnv := gendsl.NewEnv().
		WithProcedure("later-than", gendsl.Procedure{Eval: gendsl.CheckNArgs("2", _laterThan)}).
		WithInt("$NOW", _year())

	cond, err := args[0].EvalWithEnv(condEnv)
	if err != nil {
		return nil, err
	}
	if cond.Type() == gendsl.ValueTypeBool && cond.Unwrap().(bool) == true {
		return args[1].Eval()
	}

	return args[2].Eval()
}

func isTypeOf[T any](v any) bool {
	_, ok := v.(T)
	return ok
}

var env = gendsl.NewEnv().WithProcedure("json", gendsl.Procedure{
	Eval: _json,
})

func EvalJSON(script string) ([]byte, error) {
	jsonResult, err := gendsl.EvalExpr(script, env)
	if err != nil {
		return nil, err
	}

	return jsonResult.(*gendsl.UserData).V.([]byte), nil
}

// ExampleEvalExpr demonstrates a DSL that defines and output a JSON.
// Inside the (json ...) block you can use (kv {key} {value}) to add an key-value pair in this JSON.
// The value could be any string, int, float or array by using (array {value}...) or dict by using (dict (kv {key} {value})...)
func ExampleEvalExpr() {
	RegisterFailHandler(func(message string, callerSkip ...int) {
		panic(message)
	})
	script := `
(json
 (if (later-than $NOW 2012)                                  ; condition
     (kv "language" (array "c" "c++" "javascript" "elixir")) ; then
     (kv "language" (array "c" "c++" "javascript")))         ; else
 (kv "typing"
     (dict
      (kv "c" "static")
      (kv "c++" "static")
      (kv "javascript" "dynamic")))
 )
`

	jsonResult, err := EvalJSON(script)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(jsonResult))
	// Output: {"language":["c","c++","javascript","elixir"],"typing":{"c":"static","c++":"static","javascript":"dynamic"}}
}
