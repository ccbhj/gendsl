# gendsl
`gendsl` provides a DSL in [Lisp](https://en.wikipedia.org/wiki/Lisp_(programming_language)) style  and allows you to customize your own expressions so that you can integrate it into your own golang application without accessing any lexer or parser.

## ✨ Features
- ⌨️ **Highly customizable**: Invoke your own golang functions in the DSL and actually define your own expressions like 'if-then-else', 'switch-case'.
- 🔌 **Easy and lightweight**: Syntax with simplicity and explicity, easy to learn.
- 🎯 **Value injection**: Access and inject any pre-defined variables/functions in the DSL.
- 🔎 **Lexical scope**: Variable name reference are lexical-scoped.

## 📦 Installation
`go get github.com/ccbhj/gendsl`

## 📋 Usage

### Setup environment 
An environment is basically a table to lookup identifiers. By injecting values into a environment and pass the environment when evaluating expression, you can access values by its name in your expressions. Here are how you declare and inject value into an environment:
```golang
env := gendsl.NewEnv().
         WithInt("ONE", 1).                           // inject an integer 1 named ONE
         WithString("@TWO", "2").                     // inject an string 2 named @TWO
         WithProcedure("PRINTLN", gendsl.Procedure{   // inject an procedure named PRINTLN with one or more arguments
             Eval: gendsl.CheckNArgs("+", printlnOp),
         }).
         WithProcedure("PLUS", gendsl.Procedure{   
             Eval: gendsl.CheckNArgs("+", plusOp),
         })
```
That's it, now you have an environment for expression evaluation. Note that values used in our expression are typed. Currently we support *Int/Uint/Bool/String/Float/UserData/Nil/Procedure*. If you cannot find any type that can satisfy your need, use UserData, and use Nil instead of nil literal as posible as you can.

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
(PRINTLN (MINUS 1 2)) ; => invalid, since 'MINUS' is not defined in env.
```

### Define procedures
#### Basic
A procedure is just a simple function that accept a bunch of expressions and return a value.
```golang 
_plus := func(ectx *gendsl.EvalCtx, args []gendsl.Expr) (gendsl.Value, error) {
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
The `EvalCtx` provides some information for evaluation including the env of the out scope, and `args` are some expressions as arguments. Inject your function wrapped with `gendsl.Procedure` into an env then you are good to go use it in your expressions. You may want to use `CheckNArgs()` to save you from the trouble of checking the amount of arguments everywhere.

#### Control the evaluation of an expression
By calling `arg.Eval()`, we can evaluate the sub-expression for this procedure. This means that **the sub-expression(or the sub-ast) is not evaluated until we call the `Eval()` method**. With this ability, you can define your own 'if-else-then' like this:
```golang
// _if takes first argument C as condition,
// returns the second expression if C is not nil, and returns execute the third expression otherwise
// example: 
//   (IF (EQUAL 1 1) "foo" "bar") ; => "foo"
// Note that expression "bar" will not be evaluated.
_if := func(_ *gendsl.EvalCtx, args []gendsl.Expr) (gendsl.Value, error) {
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

Also, **you can inject values during an procedure's evaluation by `arg.EvalWithEnv(e)`**:
```golang
// _block lets the procedure PLUS can only visible inside expressions of the procedure BLOCK.
// example:
//   (BLOCK 1 (PLUS 2 3)) ; => Int(5)
//   (PLUS 2 3)           ; => error! PLUS is not defined outside BLOCK
_block := func(_ *gendsl.EvalCtx, args []gendsl.Expr) (gendsl.Value, error) {
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

// print "helloworld" to stderr
_, err := gendsl.EvalExprWithData(`(PRINTLN "helloworld")`,
    gendsl.NewEnv().WithProcedure("PRINTLN", gendsl.Procedure{
        Eval: gendsl.CheckNArgs("+", printlnOp),
    }),
    os.Stderr,
)
```
See [ExampleEvalExprWithData](https://github.com/ccbhj/gendsl/blob/main/examples/example_mini_awk_test.go) for a more detailed example of a mini awk.

## Syntax
The syntax is pretty simple since **everything is just nothing more that an expression which produces a value**.<br>

### Expression
Our DSL can only be an single expression: <br>
```
DSL        = Expression
Expression = Int 
           | Uint 
           | Float 
           | Boolean
           | String
           | Nil
           | Identifier
           | '(' Identifier Expression... ')'
```
This means you can pass these expression to `EvalExpr`:
```
> 1                ; => integer 1
> $1               ; => any thing named "$1"
> (PLUS 1 1)       ; => Invoke procedure named PLUS
```

### Comment
Just like Common-Lisp, Scheme and Clojure, anything following ';' are treated as comments.

### Data Types
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
| procedure | ValueTypeProcedure | func(*EvalCtx,[]Expr) (Value, error) |
#### numbers(int/uint/float)
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
#### string
```
String  = '"' Char+ '"'
Char    = '\u' HexDigit HexDigit HexDigit HexDigit
        | '\U' HexDigit HexDigit HexDigit HexDigit HexDigit HexDigit HexDigit HexDigit
        | '\x' HexDigit HexDigit
        | '\' [abfnrtv\"']
        | .*
```
These expressions are parsed as a string:
```
> "'"                    ; String(`'`)
> "\\"                   ; String(`\`)
> "\u65e5"               ; String("日")
> "\U00008a9e"           ; String("语")
> "\u65e5本\U00008a9e"   ; String("日本语")
> "\x61"                 ; String("a")
```

#### bool
```
BoolLiteral = '#' [t/f]
```
These expressions are parsed as a bool:
```
> #t      ; true
> #f      ; false
```

### Nil
```
NilLiteral = 'nil'
```
`nil` represents `nil` in Go:
```
> nil
```
