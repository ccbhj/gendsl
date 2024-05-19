package json

import (
	"encoding/json"
	"fmt"

	"github.com/ccbhj/gendsl"
)

type KV struct {
	K string
	V any
}

// only in json env would these operators can be accessed
var jsonOps = gendsl.NewEnv().
	WithOperator("array", gendsl.Operator{Eval: _array}).
	WithOperator("kv", gendsl.Operator{Eval: _kv}).
	WithOperator("dict", gendsl.Operator{Eval: _dict})

func _json(evalCtx *gendsl.EvalCtx, args []gendsl.Expr) (any, error) {
	ret := make(map[string]any)
	for _, arg := range args {
		v, err := arg.EvalWithEnv(jsonOps) // inject some operators in this context
		if err != nil {
			return nil, err
		}
		kv, ok := v.(KV)
		if !ok {
			panic("expecting a kv")
		}
		ret[kv.K] = kv.V
	}
	if isTypeOf[*map[string]any](evalCtx.UserData) {
		*(evalCtx.UserData.(*map[string]any)) = ret
	}

	return json.Marshal(ret)
}

func _kv(_ *gendsl.EvalCtx, args []gendsl.Expr) (any, error) {
	key, err := args[0].Eval()
	if err != nil {
		return nil, err
	}

	val, err := args[1].Eval()
	if err != nil {
		return nil, err
	}
	if !isTypeOf[string](key) {
		panic(fmt.Sprintf("key is not string, but %+v", key))
	}
	return KV{K: key.(string), V: val}, nil
}

func _array(_ *gendsl.EvalCtx, args []gendsl.Expr) (any, error) {
	val := make([]any, 0, 1)
	for _, arg := range args {
		v, err := arg.Eval()
		if err != nil {
			return nil, err
		}
		val = append(val, v)
	}

	return val, nil
}

func _dict(_ *gendsl.EvalCtx, args []gendsl.Expr) (any, error) {
	val := make(map[string]any)
	for _, arg := range args {
		v, err := arg.Eval()
		if err != nil {
			return nil, err
		}
		kv, ok := v.(KV)
		if !ok {
			panic("expecting a kv")
		}
		val[kv.K] = kv.V
	}

	return val, nil
}

func isTypeOf[T any](v any) bool {
	_, ok := v.(T)
	return ok
}

var env = gendsl.NewEnv().WithOperator("json", gendsl.Operator{
	Eval: _json,
})

func EvalJSON(script string) (map[string]any, []byte, error) {
	var retPtr = new(map[string]any)

	env = map[string]any{
		"json": gendsl.Operator{
			Eval: _json,
		},
	}
	jsonResult, err := gendsl.EvalExprWithInput(script, env, retPtr)
	if err != nil {
		return nil, nil, err
	}

	return *retPtr, jsonResult.([]byte), nil
}
