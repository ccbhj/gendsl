package json

import (
	"encoding/json"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/ccbhj/gendsl"
)

func TestJson(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Json Suite")
}

var _ = Describe("Json", func() {
	var (
		env map[string]any
	)

	It("can return result", func() {
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

		var retPtr = new(map[string]any)
		jsonResult, err := gendsl.EvalExprWithInput(script, env, retPtr)
		Expect(err).Should(BeNil())
		if Expect(isTypeOf[[]byte](jsonResult)).To(BeTrue()) {
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
