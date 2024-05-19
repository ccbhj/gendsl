package gendsl

import (
	"encoding/json"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type KV struct {
	K string
	V any
}

func _json(evalCtx *EvalCtx, args []Expr) (any, error) {
	ret := make(map[string]any)
	for _, arg := range args {
		v, err := arg.Eval()
		if err != nil {
			return nil, err
		}
		kv, ok := v.(KV)
		if !ok {
			panic("expecting a kv")
		}
		ret[kv.K] = kv.V
	}
	if typeIs[*map[string]any](evalCtx.UserData) {
		*(evalCtx.UserData.(*map[string]any)) = ret
	}

	return json.Marshal(ret)
}

func _kv(evalCtx *EvalCtx, args []Expr) (any, error) {
	key, err := args[0].Eval()
	if err != nil {
		return nil, err
	}

	val, err := args[1].Eval()
	if err != nil {
		return nil, err
	}
	if !typeIs[string](key) {
		panic(fmt.Sprintf("key is not string, but %+v", key))
	}
	return KV{K: key.(string), V: val}, nil
}

func _array(evalCtx *EvalCtx, args []Expr) (any, error) {
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

func _dict(evalCtx *EvalCtx, args []Expr) (any, error) {
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

var _ = Describe("Expr", func() {
	Describe("json", func() {
		var (
			env map[string]any
		)
		BeforeEach(func() {
			env = map[string]any{
				"json": Operator{
					Eval: _json,
				},
				"array": Operator{
					Eval: _array,
				},
				"kv": Operator{
					Eval: _kv,
				},
				"dict": Operator{
					Eval: _dict,
				},
			}
		})

		It("can return result", func() {
			script := `
(json
 ; :use_number #t
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

			var retPtr = new(map[string]any)
			jsonResult, err := EvalExprWithInput(script, env, retPtr)
			Expect(err).Should(BeNil())
			if Expect(typeIs[[]byte](jsonResult)).To(BeTrue()) {
				GinkgoLogr.Info("json result", "result", string(jsonResult.([]byte)))
				jsonMap := make(map[string]any)
				Expect(json.Unmarshal(jsonResult.([]byte), &jsonMap)).Should(BeNil())

				Expect(*retPtr).To(HaveKeyWithValue("foo", "bar"))
				Expect(*retPtr).To(HaveKeyWithValue("language", []any{"c", "c++", "javascript"}))
				Expect(*retPtr).To(HaveKeyWithValue("typing", map[string]any{
					"c":          "static",
					"c++":        "static",
					"javascript": "dynamic",
				}))
				Expect(*retPtr).To(HaveKeyWithValue("first_appear", map[string]any{
					"c":          int64(1972),
					"c++":        int64(1985),
					"javascript": int64(1995),
				}))
			}
		})
	})
})

func typeIs[T any](v any) bool {
	_, ok := v.(T)
	return ok
}
