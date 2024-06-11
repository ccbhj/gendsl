package gendsl

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
)

type Map map[string]Value

func (m Map) Index(path string) (Value, bool) {
	v, in := m[path]
	return v, in
}

func _array(_ *EvalCtx, args []Expr, _ map[string]Value) (Value, error) {
	ret := make([]Value, 0, len(args))
	for _, arg := range args {
		v, err := arg.Eval()
		if err != nil {
			return nil, err
		}

		ret = append(ret, v)
	}

	return &UserData{V: ret}, nil
}

func _return(evalCtx *EvalCtx, args []Expr, _ map[string]Value) (Value, error) {
	return args[0].Eval()
}

func _plus(evalCtx *EvalCtx, args []Expr, _ map[string]Value) (Value, error) {
	var ret Int
	for i, v := range args {
		x, err := v.Eval()
		if err != nil {
			return nil, err
		}
		if x.Type() != ValueTypeInt {
			return nil, errors.Errorf("invalid type for #%d arg: %s", i, x.Type())
		}

		ret += x.(Int)
	}

	return ret, nil
}

func _define(evalCtx *EvalCtx, args []Expr, _ map[string]Value) (Value, error) {
	id, err := args[0].Eval()
	if err != nil {
		return nil, err
	}
	if id.Type() != ValueTypeString {
		return nil, errors.New("expecting an string as id")
	}
	val, err := args[1].Eval()
	if err != nil {
		return nil, err
	}
	return args[2].EvalWithEnv(NewEnv().WithValue(string(id.(String)), val))
}

func _block(evalCtx *EvalCtx, args []Expr, _ map[string]Value) (Value, error) {
	var ret Value
	for _, arg := range args {
		v, err := arg.Eval()
		if err != nil {
			return nil, err
		}
		ret = v
	}

	if ret == nil {
		return Nil{}, nil
	}
	return ret, nil
}

var _ = Describe("Expr", func() {
	var testEnv *Env
	BeforeEach(func() {
		testEnv = NewEnv().
			WithProcedure("RETURN", Procedure{Eval: CheckNArgs("1", _return)}).
			WithProcedure("DEFINE", Procedure{Eval: CheckNArgs("3", _define)}).
			WithProcedure("BLOCK", Procedure{Eval: CheckNArgs("*", _block)}).
			WithProcedure("PLUS", Procedure{Eval: CheckNArgs("*", _plus)}).
			WithProcedure("ARRAY", Procedure{Eval: CheckNArgs("*", _array)})
	})
	Describe("Literal", func() {
		var (
			evalFn = func(s string) (Value, error) {
				return EvalExpr(s, testEnv)
			}
		)

		It("can eval a simple integer", func() {
			Expect(evalFn("10")).Should(BeIdenticalTo(Int(10)))
			Expect(evalFn("-10")).Should(BeIdenticalTo(Int(-10)))

			Expect(evalFn("0x10")).Should(BeIdenticalTo(Int(16)))
			Expect(evalFn("-0x10")).Should(BeIdenticalTo(Int(-16)))
		})

		It("can eval a simple integer with 'u' suffix as uint", func() {
			Expect(evalFn("10u")).Should(BeIdenticalTo(Uint(10)))
			Expect(evalFn("10U")).Should(BeIdenticalTo(Uint(10)))

			Expect(evalFn("0x10U")).Should(BeIdenticalTo(Uint(16)))
			Expect(evalFn("0x10u")).Should(BeIdenticalTo(Uint(16)))
		})

		It("can eval a simple integer with '.0' suffix as float", func() {
			Expect(evalFn("10.0")).Should(BeIdenticalTo(Float(10)))
			Expect(evalFn("10.0")).Should(BeIdenticalTo(Float(10)))
		})

		It("can eval a simple float", func() {
			Expect(evalFn("0.")).Should(BeIdenticalTo(Float(0)))
			Expect(evalFn("74.40")).Should(BeIdenticalTo(Float(74.40)))

			Expect(evalFn("1.e+0")).Should(BeIdenticalTo(Float(1)))
			Expect(evalFn("6.67428e-2")).Should(BeIdenticalTo(Float(0.0667428)))
			Expect(evalFn("1E6")).Should(BeIdenticalTo(Float(1000000)))
			Expect(evalFn(".25")).Should(BeIdenticalTo(Float(0.25)))
			Expect(evalFn(".12345E+5")).Should(BeIdenticalTo(Float(12345)))
			Expect(evalFn("0.15e+0_2")).Should(BeIdenticalTo(Float(15.0)))
		})

		It("can eval a nil literal", func() {
			Expect(evalFn("nil")).Should(BeIdenticalTo(Nil{}))
			Expect(EvalExpr("nil", testEnv.Clone().WithInt("nil", 10))).Should(BeIdenticalTo(Nil{}))
		})

		It("can eval a simple string", func() {
			Expect(evalFn(`"hello"`)).Should(BeIdenticalTo(String("hello")))
			Expect(evalFn(`"你好"`)).Should(BeIdenticalTo(String("你好")))

			Expect(evalFn(`"'"`)).Should(BeIdenticalTo(String(`'`)))
			Expect(evalFn(`"\""`)).Should(BeIdenticalTo(String(`"`)))
			Expect(evalFn(`"\n"`)).Should(BeIdenticalTo(String("\n")))
			Expect(evalFn(`"\r"`)).Should(BeIdenticalTo(String("\r")))

			Expect(evalFn(`"\xc3"`)).Should(BeIdenticalTo(String([]byte{0xc3})))
			Expect(evalFn(`"\x61"`)).Should(BeIdenticalTo(String("a")))
			Expect(evalFn(`"\u65e5本\U00008a9e"`)).Should(BeIdenticalTo(String("日本語")))
			Expect(evalFn(`"\\"`)).Should(BeIdenticalTo(String(`\`)))
			Expect(evalFn(`"\\\\"`)).Should(BeIdenticalTo(String(`\\`)))
			Expect(evalFn(`"\\\\\n"`)).Should(BeIdenticalTo(String(`\\` + "\n")))
			Expect(evalFn(`"\\\n\\\n"`)).Should(BeIdenticalTo(String(`\` + "\n" + `\` + "\n")))
		})

		It("can eval a long string", func() {
			Expect(evalFn(`"""hello"""`)).Should(BeIdenticalTo(String("hello")))
			Expect(evalFn(`"""你好"""`)).Should(BeIdenticalTo(String("你好")))
			Expect(evalFn(`"""line
 break"""`)).Should(BeIdenticalTo(String("line\n break")))
			Expect(evalFn(`"""\n"""`)).Should(BeIdenticalTo(String(`\n`)))
			Expect(evalFn(`"""\xc3"""`)).Should(BeIdenticalTo(String(`\xc3`)))
			Expect(evalFn(`"""\\\"""`)).Should(BeIdenticalTo(String(`\\\`)))
			Expect(evalFn(`"""''"""`)).Should(BeIdenticalTo(String(`''`)))
			Expect(evalFn(`(ARRAY """""" "")`)).
				Should(BeEquivalentTo(&UserData{[]Value{String(""), String("")}}))
			Expect(evalFn(`(ARRAY "" """""" "")`)).
				Should(BeEquivalentTo(&UserData{[]Value{String(""), String(""), String("")}}))
		})
	})

	Describe("Expr", func() {
		It("can accept no argument", func() {
			Expect(EvalExpr(`(PLUS)`, testEnv)).Should(BeEquivalentTo(0))
		})

		It("can eval an simple expr with literals as argument", func() {
			Expect(EvalExpr(`(RETURN 10)`, testEnv)).Should(BeEquivalentTo(10))
			Expect(EvalExpr(`(PLUS 1 2 3)`, testEnv)).Should(BeEquivalentTo(6))
		})

		It("can eval an simple expr with expr as argument", func() {
			Expect(EvalExpr(`(RETURN (PLUS 1 2 3))`, testEnv)).Should(BeEquivalentTo(6))
			Expect(EvalExpr(`(RETURN (PLUS 1 2 (PLUS 1 2)))`, testEnv)).Should(BeEquivalentTo(6))
		})

		It("can check the expr type for arguments", func() {
			check := func(evalCtx *EvalCtx,
				args []Expr,
				options map[string]Value) (Value, error) {
				Expect(args[0].Type()).Should(BeEquivalentTo(ExprTypeLiteral))
				Expect(args[1].Type()).Should(BeEquivalentTo(ExprTypeExpr))
				Expect(args[2].Type()).Should(BeEquivalentTo(ExprTypeIdentifier))
				return Nil{}, nil
			}
			env := testEnv.Clone().WithProcedure("CHECK", Procedure{Eval: check}).
				WithInt("bar", 10)
			Expect(extractErr2(EvalExpr, `(CHECK "foo" (RETURN 1) bar)`, env)).
				Should(BeNil())
		})

		Describe("argument check", func() {
			It("can check extract count of argument", func() {
				var env = NewEnv().
					WithProcedure("PLUS", Procedure{Eval: CheckNArgs("2", _plus)})
				Expect(extractErr2(EvalExpr, `(PLUS 1 2 3)`, env)).Should(MatchError("expecting 2 argument(s), but got 3"))
				Expect(EvalExpr(`(PLUS 1 2)`, env)).Should(BeEquivalentTo(3))

			})

			It("can check one or more argument", func() {
				env := NewEnv().
					WithProcedure("PLUS", Procedure{Eval: CheckNArgs("+", _plus)})
				Expect(extractErr2(EvalExpr, `(PLUS )`, env)).Should(MatchError("expecting one or more argument, but got 0"))
				Expect(EvalExpr(`(PLUS 1 2)`, env)).Should(BeEquivalentTo(3))
				Expect(EvalExpr(`(PLUS 1)`, env)).Should(BeEquivalentTo(1))
			})

			It("can check no or more argument", func() {
				env := NewEnv().
					WithProcedure("PLUS", Procedure{Eval: CheckNArgs("*", _plus)})
				Expect(EvalExpr(`(PLUS )`, env)).Should(BeEquivalentTo(0))
				Expect(EvalExpr(`(PLUS 1)`, env)).Should(BeEquivalentTo(1))
				Expect(EvalExpr(`(PLUS 1 2)`, env)).Should(BeEquivalentTo(3))
			})

			It("can check one or no argument", func() {
				env := NewEnv().
					WithProcedure("PLUS", Procedure{Eval: CheckNArgs("?", _plus)})
				Expect(extractErr2(EvalExpr, `(PLUS 1 2 3)`, env)).Should(MatchError("expecting one or no argument, but got 3"))
				Expect(EvalExpr(`(PLUS 1)`, env)).Should(BeEquivalentTo(1))
				Expect(EvalExpr(`(PLUS)`, env)).Should(BeEquivalentTo(0))
			})

		})
	})

	Describe("option value", func() {
		var env *Env
		BeforeEach(func() {
			p := Procedure{
				Eval: func(evalCtx *EvalCtx, args []Expr, opts map[string]Value) (Value, error) {
					return &UserData{V: opts}, nil
				},
			}
			env = testEnv.Clone().WithProcedure("OPTIONS", p)
		})

		It("can declare literal option in a procedure", func() {
			Expect(EvalExpr(`(OPTIONS #:foo "bar" #:one 1)`, env)).Should(BeEquivalentTo(&UserData{
				V: map[string]Value{
					"foo": String("bar"),
					"one": Int(1),
				},
			}))
		})

		It("can declare value env in the option in a procedure", func() {
			e := env.Clone().WithInt("ONE", 1)
			Expect(EvalExpr(`(OPTIONS #:foo "bar" #:one ONE)`, e)).Should(BeEquivalentTo(&UserData{
				V: map[string]Value{
					"foo": String("bar"),
					"one": Int(1),
				},
			}))
		})

		It("can use option along with some value", func() {
			p := Procedure{
				Eval: func(evalCtx *EvalCtx, args []Expr, opts map[string]Value) (Value, error) {
					var ret Int
					limit := opts["N"].(Int)
					for i := 0; i < int(limit) && i < len(args); i++ {
						v, err := args[i].Eval()
						if err != nil {
							return nil, err
						}
						ret += v.(Int)
					}
					return ret, nil
				},
			}
			// only allow two values
			env = testEnv.Clone().WithProcedure("PLUS_N", p)
			Expect(EvalExpr(`(PLUS_N #:N 2 10 20)`, env)).Should(BeIdenticalTo(Int(30)))
			Expect(EvalExpr(`(PLUS_N #:N 2 (RETURN 10) 20)`, env)).Should(BeIdenticalTo(Int(30)))
			Expect(EvalExpr(`(PLUS_N #:N 2 10 20 30 40 50)`, env)).Should(BeIdenticalTo(Int(30)))
		})

		It("cannot use an procedure expression in option", func() {
			Expect(extractErr2(EvalExpr, `(OPTIONS #:foo "bar" #:one (RETURN 1))`, env)).
				Should(MatchError(ContainSubstring("parse error")))
		})

	})

	Describe("env", func() {
		var (
			evalFn = func(s string, id string, val Value) (Value, error) {
				env := NewEnv().WithValue(id, val).WithProcedure("RETURN", Procedure{
					Eval: _return,
				})
				return EvalExpr(s, env)
			}
		)

		It("can evaluate an identifier", func() {
			Expect(evalFn("foo", "foo", Int(10))).To(BeIdenticalTo(Int(10)))
			Expect(evalFn("FOO", "FOO", Int(10))).To(BeIdenticalTo(Int(10)))
			Expect(evalFn("@foo", "@foo", Int(10))).To(BeIdenticalTo(Int(10)))
			Expect(evalFn("$foo", "$foo", Int(10))).To(BeIdenticalTo(Int(10)))
			Expect(evalFn("(RETURN $foo)", "$foo", Int(10))).To(BeIdenticalTo(Int(10)))
			Expect(evalFn("(RETURN _foo)", "_foo", Int(10))).To(BeIdenticalTo(Int(10)))
			Expect(evalFn("(RETURN foo_)", "foo_", Int(10))).To(BeIdenticalTo(Int(10)))
			Expect(evalFn("(RETURN foo-bar)", "foo-bar", Int(10))).To(BeIdenticalTo(Int(10)))
		})

		It("can get attributes of an identifier", func() {
			Expect(evalFn("foo.bar", "foo", &UserData{
				V: Map{"bar": Int(10)},
			})).To(BeIdenticalTo(Int(10)))

			Expect(evalFn("foo.bar.$eww", "foo", &UserData{
				V: Map{"bar": &UserData{V: Map{"$eww": Int(10)}}},
			})).To(BeIdenticalTo(Int(10)))

			_, err := evalFn("foo.bar.$error", "foo", &UserData{
				V: Map{"bar": &UserData{V: Map{"$eww": Int(10)}}},
			})
			Expect(err).To(MatchError(ContainSubstring("index($error) not found")))
		})

		It("can refer to a value with its id", func() {
			env := NewEnv().WithInt("FOO", Int(10)).
				WithProcedure("PLUS", Procedure{Eval: CheckNArgs("*", _plus)})
			Expect(EvalExpr("(PLUS FOO 10)", env)).Should(BeEquivalentTo(20))
		})

		It("can refer to a value with its id in its parent's env", func() {
			env := NewEnv().WithInt("FOO", Int(10)).
				WithProcedure("DEFINE", Procedure{Eval: CheckNArgs("3", _define)}).
				WithProcedure("PLUS", Procedure{Eval: CheckNArgs("*", _plus)})
			script := `
			(DEFINE "BAR" 10
				(PLUS FOO BAR)
			)
			`
			Expect(EvalExpr(script, env)).Should(BeEquivalentTo(20))

			// look up from grandparent's env
			script = `
			(DEFINE "BAR" 10
				(DEFINE "ONE" 1 
			    (PLUS FOO BAR ONE))
			)
			`
			Expect(EvalExpr(script, env)).Should(BeEquivalentTo(21))
		})

		It("can set an value to an env in a procedure", func() {
			expr := `
			(DEFINE "foo" 20
				(RETURN foo))
			`
			Expect(EvalExpr(expr, testEnv)).Should(BeEquivalentTo(20))
		})

		It("can throw unbounded variable error outside its scope", func() {
			expr := `
			(BLOCK 
				(DEFINE "foo" 20
					(RETURN foo))
				(RETURN foo)
			)
			`
			err := extractErr2(EvalExpr, expr, testEnv)
			if Expect(err).ShouldNot(BeNil()) {
				Expect(err).Should(BeAssignableToTypeOf(&UnboundedIdentifierError{}))
				Expect(err).Should(MatchError(ContainSubstring("unbounded")))
			}
		})

		It("can override value in parent enviroment", func() {
			expr := `
			(DEFINE "foo" 10
				(DEFINE "foo" "bar" (RETURN foo))
			)
			`
			Expect(EvalExpr(expr, testEnv)).Should(BeEquivalentTo(String("bar")))
		})
	})
})

func extractErr2[X, Y, T any](fn func(X, Y) (T, error), x X, y Y) error {
	_, err := fn(x, y)
	return err
}

var testEnv = NewEnv().WithProcedure("RETURN", Procedure{
	Eval: CheckNArgs("1", func(_ *EvalCtx, args []Expr, _ map[string]Value) (Value, error) {
		return args[0].Eval()
	}),
},
)

func FuzzExprLiteralInt(f *testing.F) {
	// Int
	f.Add(int64(1))
	f.Fuzz(func(t *testing.T, i int64) {
		var f string
		switch i & 0x1 {
		case 0:
			f = fmt.Sprintf("%d", i)
		case 1:
			if i >= 0 {
				f = fmt.Sprintf("0x%x", i)
			} else {
				f = fmt.Sprintf("-0x%x", -i)
			}
		}
		result, err := EvalExpr(fmt.Sprintf("(RETURN %s)", f), testEnv)
		if err != nil {
			t.Errorf("eval error: %s, input=%q", err, f)
			t.FailNow()
		}
		if result.Type() != ValueTypeInt {
			t.Errorf("wrong type: %s, input=%q", result.Type(), f)
			t.FailNow()
		}

		if result.Unwrap().(int64) != i {
			t.Errorf("not equal: %s, input=%q", err, f)
			t.FailNow()
		}
	})
}

func FuzzExprLiteralUint(f *testing.F) {
	// Uint
	f.Add(uint64(1))
	f.Fuzz(func(t *testing.T, i uint64) {
		suffix := "u"
		if i%2 == 0 {
			suffix = "U"
		}
		result, err := EvalExpr(fmt.Sprintf("(RETURN %d%s)", i, suffix), testEnv)
		if err != nil {
			t.Errorf("eval error: %s, input=%v", err, i)
			t.FailNow()
		}
		if result.Type() != ValueTypeUInt {
			t.FailNow()
		}

		if result.Unwrap().(uint64) != i {
			t.FailNow()
		}
	})
}

func FuzzExprLiteralString(f *testing.F) {
	// String
	f.Add("\xd2")
	f.Fuzz(func(t *testing.T, s string) {
		script := fmt.Sprintf(`(RETURN %s)`, strconv.Quote(s))
		result, err := EvalExpr(script, testEnv)
		if err != nil {
			_, e := strconv.Unquote(fmt.Sprintf(`"%s"`, s))
			if e != nil {
				// expecting an err not to be nil
				return
			}
			t.Errorf("eval error: %s, input=%v", err, s)
			t.FailNow()
		}
		if result.Type() != ValueTypeString {
			t.Fatalf("result type is not string, result=%v", result)
		}
		if string(result.(String)) != s {
			t.Fatalf("result is not the same, result=%s, s=%s", hex.EncodeToString([]byte(result.(String))), hex.EncodeToString([]byte(s)))
		}
	})

}

func FuzzExprLiteralFloat(f *testing.F) {
	// Float
	f.Add(2.0)
	f.Fuzz(func(t *testing.T, s float64) {
		var fn func(float64) string
		switch int(s*1000) % 5 {
		default:
			fn = func(f float64) string {
				return fmt.Sprintf("%f", f)
			}
		case 1:
			fn = func(f float64) string {
				return fmt.Sprintf("%e", f)
			}
		case 2:
			fn = func(f float64) string {
				return fmt.Sprintf("%E", f)
			}
		case 3:
			fn = func(f float64) string {
				return fmt.Sprintf("%g", f)
			}
		case 4:
			fn = func(f float64) string {
				return fmt.Sprintf("%G", f)
			}
		}
		script := fmt.Sprintf(`(RETURN %s)`, fn(s))
		result, err := EvalExpr(script, testEnv)
		if err != nil {
			t.Errorf("eval error: %s, input=%v", err, s)
			t.FailNow()
		}
		if result.Type() != ValueTypeFloat {
			t.Fatalf("result type is not float, input=%q, result=%v", fn(s), result)
		}
		if fn(s) != fn(result.Unwrap().(float64)) {
			t.Fatalf("result not the same, input=%q, result=%v", fn(s), fn(result.Unwrap().(float64)))
		}
	})

}
