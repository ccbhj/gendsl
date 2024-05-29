package gendsl_test

import (
	"encoding/json"
	"fmt"

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
	WithProcedure("array", gendsl.Procedure{Eval: _array}).
	WithProcedure("kv", gendsl.Procedure{Eval: _kv}).
	WithProcedure("dict", gendsl.Procedure{Eval: _dict})

func _json(evalCtx *gendsl.EvalCtx, args []gendsl.Expr) (gendsl.Value, error) {
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
func _kv(_ *gendsl.EvalCtx, args []gendsl.Expr) (gendsl.Value, error) {
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
func _array(_ *gendsl.EvalCtx, args []gendsl.Expr) (gendsl.Value, error) {
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

// _array evaluate any KV pair and assemble them to a dict(map[string]any)
func _dict(_ *gendsl.EvalCtx, args []gendsl.Expr) (gendsl.Value, error) {
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

// ExampleEvalExpr demonstrates a DSL that defines and output a JSON []byte.
// Inside the (json ...) block you can use (kv {key} {value}) to add an key-value pair in this JSON.
// The value could be any string, int, float or array by using (array {value}...) or dict by using (dict (kv {key} {value})...)
func ExampleEvalExpr() {
	RegisterFailHandler(func(message string, callerSkip ...int) {
		panic(message)
	})

	script := `
(json
 (kv "foo" "bar")
 (kv "language" (array "c" "c++" "javascript"))
 (kv "typing" 
    (dict
      (kv "c" "static")
      (kv "c++" "static")
      (kv "javascript" "dynamic")
    )
  )
 (kv "first_appear" 
    (dict
      (kv "c" 1972)
      (kv "c++" 1985)
      (kv "javascript" 1995)
    )
  )
)`

	jsonResult, err := EvalJSON(script)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(jsonResult))
	// Output: {"first_appear":{"c":1972,"c++":1985,"javascript":1995},"foo":"bar","language":["c","c++","javascript"],"typing":{"c":"static","c++":"static","javascript":"dynamic"}}
}
