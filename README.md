# gendsl
`gendsl` provides a framework to create a DSL in [Lisp](https://en.wikipedia.org/wiki/Lisp_(programming_language)) style and allows you to customize your own expressions so that you can integrate it into your own application without accessing any lexer or parser.

## ✨ Features
- ⌨️ **Highly customizable**: Inject your own golang functions in the DSL and actually define your own expressions like 'if-then-else', 'switch-case'.
- 🔌 **Easy and lightweight**: Syntax with simplicity and explicity, easy to learn.
- 🌐 **Option value**: Option value syntax like [Racket](https://docs.racket-lang.org/rebellion/Option_Values.html) for data description.
- 🎯 **Value injection**: Access and inject any variables/functions in your DSL.
- 🔎 **Lexical scope**: Variable name reference are lexical-scoped.

## 📦 Installation
`go get github.com/ccbhj/gendsl`

## 📋 Usage

### Setup environment 
An environment is basically a table to lookup identifiers. By injecting values into an environment and pass the environment when evaluating expression, you can access values by its name in your expressions. Here are how you declare and inject value into an environment:
```golang
env := gendsl.NewEnv().
         WithInt("ONE", 1).                           // inject an integer 1 named ONE
         WithString("@TWO", "2").                     // inject an string 2 named @TWO
         WithProcedure("PRINTLN", gendsl.Procedure{   // inject an procedure named PRINTLN with one or more arguments
             Eval: gendsl.CheckNArgs("+", printlnOp),
         }).
         WithProcedure("PLUS", gendsl.Procedure{   
             Eval: gendsl.CheckNArgs("2", plusOp),
         })
```
That's it, now you have an environment for expression evaluation. Note that values used in our expression are typed. Currently we support *[Int](https://pkg.go.dev/github.com/ccbhj/gendsl#Int)/[Uint](https://pkg.go.dev/github.com/ccbhj/gendsl#Uint)/[Bool](https://pkg.go.dev/github.com/ccbhj/gendsl#Bool)/[String](https://pkg.go.dev/github.com/ccbhj/gendsl#String)/[Float](https://pkg.go.dev/github.com/ccbhj/gendsl#Float)/[UserData](https://pkg.go.dev/github.com/ccbhj/gendsl#UserData)/[Nil](https://pkg.go.dev/github.com/ccbhj/gendsl#Nil)/[Procedure](https://pkg.go.dev/github.com/ccbhj/gendsl#Procedure)*. If you cannot find any type that can satisfy your need, use [UserData](https://pkg.go.dev/github.com/ccbhj/gendsl#UserData), and use [Nil](https://pkg.go.dev/github.com/ccbhj/gendsl#Nil) instead of nil literal as possible as you can.

### Evaluate expressions
With `EvalExpr(expr string, env *Env) (Value, error)` you can evaluate an expression into a value. The expression syntax is the same as the parenthesized syntax of Lisp which means that an expression is either an value `X` or parenthesized list `(X Y Z ...)` where `X` is considered as a procedure and `Y`, `Z` ... are its arguments. With the env we defined [before](#setup-environment), we can write expressions like:
```
"HELLO WORLD"         ; => String("HELLO WORLD")
100                   ; => Int(100)
100.0                 ; => Float(100)
100u                  ; => Uint(100)
#t                    ; => Bool(true)
nil                   ; => Nil
(PRINTLN 10)          ; => 10
(PRINTLN (PLUS 1 2))  ; => 3
(PRINTLN (MINUS 1 2)) ; => error!!! since 'MINUS' is not defined in env.
(PRINTLN :out "stderr" (PLUS 1 2)) ; => 3, output to the stderr
```

### Define procedures
#### Basic
A procedure is just a simple function that accept a bunch of expressions and some options then return a value.
```golang 
_plus := func(ectx *gendsl.EvalCtx, args []gendsl.Expr, options map[string]gendsl.Value) (gendsl.Value, error) {
    var ret gendsl.Int
    for _, arg := range args {
        v, err := arg.Eval()            // evaluate arguments.
        if err != nil {
            return nil, err
        }
        if v.Type() == gendsl.ValueTypeInt {
            ret += v.(gendsl.Int)
        }
    }

    return ret, nil
}

env := gendsl.NewEnv().WithProcedure("PLUS", gendsl.Procedure{   
    Eval: gendsl.CheckNArgs("+", _plus),
})

```
The `EvalCtx` provides some information for evaluation including the env of the outer scope, and `args` are some expressions as arguments. Inject your function wrapped by `gendsl.Procedure` into an env then you are good to go use it in your expressions. You may want to use `CheckNArgs()` to save you from checking the amount of arguments everywhere.

#### Control the evaluation of an expression
By calling `arg.Eval()`, we can evaluate the sub-expression for this procedure. This means that **the sub-expression(or the sub-ast) is not evaluated until we call the `Eval()` method**. With this ability, you can define your own 'if-else-then' like this:
```golang
// _if takes first argument C as condition,
// returns the second expression if C is not nil, and returns execute the third expression otherwise
// example: 
//   (IF (EQUAL 1 1) "foo" "bar") ; => "foo"
// Note that expression "bar" will not be evaluated.
_if := func(_ *gendsl.EvalCtx, args []gendsl.Expr, options map[string]gendsl.Value) (gendsl.Value, error) {
    cond, err := args[0].Eval()
    if err != nil {
        return nil, err
    }
    if cond.Type() != gendsl.ValueTypeNil { // not-nil is considered as true
        return args[1].Eval()
    } else {
        return args[2].Eval()
    }
}
```

Also, **you can inject values during a procedure's evaluation by `arg.EvalWithEnv(e)`** so that you can have something like 'package' scope:
```golang
// _block lets the procedure PLUS can only be visible inside expressions of the procedure BLOCK.
// example:
//   (BLOCK 1 (PLUS 2 3)) ; => Int(5)
//   (PLUS 2 3)           ; => error! PLUS is not defined outside BLOCK
_block := func(_ *gendsl.EvalCtx, args []gendsl.Expr, options map[string]gendsl.Value) (gendsl.Value, error) {
    localEnv := gendsl.NewEnv().
        WithProcedure("PLUS", gendsl.Procedure{Eval: plusOp})

    var ret gendsl.Value
    for i, arg := range args {
        v, err := arg.EvalWithEnv(localEnv)
        if err != nil {
            return nil, err
        }
        ret = v
    }

    return v, nil
}
```
See the [ExampleEvalExpr](https://github.com/ccbhj/gendsl/blob/main/examples/example_json_test.go) for a more detailed example that defines and output a JSON.

#### Access input data in procedures
You might want to access some data that used as the input of your expression across all the procedures. Use `EvalExprWithData(expr string, env *Env, data any) (Value, error)` pass the data before evaluation and then access it by reading the `ectx.UserData`.
```golang 
printlnOp := func(ectx *gendsl.EvalCtx, args []gendsl.Expr, options map[string]gendsl.Value) (gendsl.Value, error) {
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

// print "helloworld" to stderr
_, err := gendsl.EvalExprWithData(`(PRINTLN "helloworld")`,
    gendsl.NewEnv().WithProcedure("PRINTLN", gendsl.Procedure{
        Eval: gendsl.CheckNArgs("+", printlnOp),
    }),
    os.Stderr,
)
```
See [ExampleEvalExprWithData](https://github.com/ccbhj/gendsl/blob/main/examples/example_mini_awk_test.go) for a more detailed example of a mini awk.

#### Use option value for data declaration
We also support `:#option {value}` for simple data declaration where {value} can only be a simple litreal. You can also use it to control the behavior of a procedure.
```golang 
printlnOp := func(ectx *gendsl.EvalCtx, args []gendsl.Expr, options map[string]gendsl.Value) (gendsl.Value, error) {
    output := os.Stdout
    outputOpt := options["out"]
    switch outputOpt.Unwrap().(string) {
    case "stdout":
        output = os.Stdout
    case "stderr":
        output = os.Stderr
    }
    for _, arg := range args {
        v, err := arg.Eval()
        if err != nil {
            return nil, err
        }
        fmt.Fprintln(output, v.Unwrap())
    }
    return gendsl.Nil{}, nil
}

// print "helloworld" to stderr
_, err := gendsl.EvalExprWithData(`(PRINTLN #:out "stderr"  "helloworld")`,
    gendsl.NewEnv().WithProcedure("PRINTLN", gendsl.Procedure{
        Eval: gendsl.CheckNArgs("+", printlnOp),
    }),
    nil,
)
```

## 🛠️ Syntax
The syntax is pretty simple since **everything is just nothing more that an expression which produces a value**.<br>

### Expression
Our DSL can only be an single expression(for now): <br>
```
DSL        = Expression
Expression = Int 
           | Uint 
           | Float 
           | Boolean
           | String
           | Nil
           | Identifier
           | '(' Identifier Options? Expression... ')'
```

Here are some examples:
```
> 1                ; => integer 1
> $1               ; => any thing named "$1"
> (PLUS 1 1)       ; => Invoke procedure named PLUS
> (PLUS #:N 2 1 1)  ; => Invoke procedure named PLUS with some options
```
Noted that options can only be declared before any argument.

### Comment
Just like Common-Lisp, Scheme and Clojure, anything following ';' are treated as comments.
```
> 1                ; This is a comment
```
### Literal Data Types
We support these types of data and they can be `Unwrap()` into Go value.

|           | Type               | Go Type                              |
|-----------|--------------------|--------------------------------------|
| int       | ValueTypeInt       | int64                                |
| uint      | ValueTypeUint      | uint64                               |
| string    | ValueTypeString    | string                               |
| float     | ValueTypeFloat     | float64                              |
| bool      | ValueTypeBool      | bool                                 |
| nil       | ValueType          | nil                                  |
| any       | ValueTypeUserData  | any                                  |
| procedure | ValueTypeProcedure | ProcedureFn                          |
#### numbers(Int/Uint/Float)
```
Int                    = [+-]? IntegerLiteral
IntegerLiteral         = '0x' HexDigit+
                       | '0X' HexDigit+
                       | DecimalDigit+

UnsignedIntegerLiteral = IntegerLiteral 'u'

FloatLiteral           = [+-]? DecimalDigit+ '.' DecimalDigit+? Exponent?
                       | [+-]? DecimalDigit+ Exponent
                       | [+-]? '.' DecimalDigit+ Exponent?
Exponent               =  [eE]? [+-]? DecimalDigit+
```
These expressions are parsed as numbers:
```
> 1              ; Int(1)
> -1             ; Int(-1)
> 0x10           ; Int(16)
> 0x10u          ; Uint(16)
> 0X10u          ; Uint(16)
> 0.             ; Float(0)
> .25            ; Float(0.25)
> 1.1            ; Float(1.1)
> 1.e+0          ; Float(1.0)
> 1.1e+0         ; Float(1.1)
> 1.1e-1         ; Float(0.11)
> 1E6            ; Float(1000000)
> 0.15e2         ; Float(15)
```
#### String
```
LongString = '"""' (![\"] .)* '"""'
String     = '"' Char+ '"'
Char       = '\u' HexDigit HexDigit HexDigit HexDigit
           | '\U' HexDigit HexDigit HexDigit HexDigit HexDigit HexDigit HexDigit HexDigit
           | '\x' HexDigit HexDigit
           | '\' [abfnrtv\"']
           | .*
```
These expressions are parsed as a string, long string is supported as well:
```
> "'"                    ; String(`'`)
> "\\"                   ; String(`\`)
> "\u65e5"               ; String("日")
> "\U00008a9e"           ; String("语")
> "\u65e5本\U00008a9e"   ; String("日本语")
> "\x61"                 ; String("a")
> """x61"""              ; String("x61")
> """\\"""               ; String(`\\`)
> """\u65e5"""           ; String(`\u65e5`)
> """line                ; String("line\nbreak")
break"""                 
```

#### Bool
```
BoolLiteral = '#' [t/f]
```
These expressions are parsed as a bool:
```
> #t      ; true
> #f      ; false
```

#### Nil
```
NilLiteral = 'nil'
```
`nil` represents `nil` in Go:
```
> nil
```
Noted that injecting a variable called 'nil' makes no sense, and you will get a 'nil' value instead of an identifier.

### Identifiers
```
Identifier = [a-zA-Z~!@$%^&*_?|<>] (LetterOrDigit / [~!@$%^&*_?|<>] / '-')*
```
These expressions are parsed as identifiers:
```
> foo
> @bar
> ?hello
> _world
> foo-bar
> a0
> a_0
> is-string?
```

Attribute select is also supported. If an UserData with an object that implements [gendsl.Selector]: 
```golang
// Selector can be used for the '.' syntax to get fields b
type Selector interface {
	// Select queries attributes by `idx`,
    // reports whether a field can be found and its value if any
	Select(idx string) (Value, bool)
}
```
Then you can refer an object's fields by '.': 
```
> foo.fields   ; returns foo.Unwrap().Select("fields")
```
An error will be thrown if Select() reports false.

### Options 
Option is a key-value pair inside an expression.
```
Option = "#:" Identifier Value
```
Option can be placed before, after or between produce arguments:
```
> (PRINTLN #:out "stderr" "foobar")
> (PLUS "hello" "world" #:type "string")
```

## 💡 Examples
<details><summary>Swith case expression</summary>

``` golang
type Case struct {
    Cond gendsl.Expr
    Then gendsl.Expr
}
_case := func(ectx *gendsl.EvalCtx, args []gendsl.Expr, _ map[string]gendsl.Value) (gendsl.Value, error) {
    return &gendsl.UserData{V: Case{args[0], args[1]}}, nil
}

_switch := func(ectx *gendsl.EvalCtx, args []gendsl.Expr, _ map[string]gendsl.Value) (gendsl.Value, error) {
    var ret gendsl.Value
    env := ectx.Env().Clone().WithProcedure("CASE", gendsl.Procedure{
        Eval: gendsl.CheckNArgs("2", _case),
    })
    expect, err := args[0].Eval()
    if err != nil {
        return nil, err
    }
    for _, arg := range args[1:] {
        cv, err := arg.EvalWithEnv(env)
        if err != nil {
            return nil, err
        }
        c, ok := cv.Unwrap().(Case)
        if !ok {
            panic("expecting a cas ")
        }

        cond, err := c.Cond.Eval()
        if err != nil {
            return nil, err
        }
        if cond == expect {
            v, err := c.Then.Eval()
            if err != nil {
                return nil, err
            }
            ret = v
        }
    }

    return ret, nil
}

script := `
(SWITCH "FOO"
(CASE "BAR" "no")
(CASE "FOO" "yes")
)
`

val, err := gendsl.EvalExprWithData(script,
    gendsl.NewEnv().
        WithProcedure("SWITCH", gendsl.Procedure{
            Eval: gendsl.CheckNArgs("+", _switch),
        }),
    nil,
)

if err != nil {
    panic(err)
}
println(val.Unwrap().(string))
// output: yes

```

</details>

<details><summary>Local variable injection</summary>

``` golang
func _let(_ *gendsl.EvalCtx, args []gendsl.Expr, options map[string]gendsl.Value) (gendsl.Value, error) {
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
(LET foo 10
    (PRINTLN foo)
)
`
env := gendsl.NewEnv().
        WithProcedure("let", gendsl.Procedure{Eval: gendsl.CheckNArgs("3", _let)}))
gendsl.EvalExpr(script, env) 
// output: 10
```

</details>
