package gendsl

type Indexable interface {
	Index(idx string) (Value, bool)
}

// ValueType specify the type of an [Value]
type ValueType uint64

const (
	ValueTypeInt       = 1 << iota // Int
	ValueTypeUInt                  // Uint
	ValueTypeString                // String
	ValueTypeBool                  // Bool
	ValueTypeFloat                 // Float
	ValueTypeProcedure             // Procedure
	ValueTypeUserData              // UserData
	ValueTypeNil                   // Nil
)

func (v ValueType) String() string {
	switch v {
	case ValueTypeInt:
		return "int"
	case ValueTypeUInt:
		return "uint"
	case ValueTypeString:
		return "string"
	case ValueTypeBool:
		return "bool"
	case ValueTypeFloat:
		return "float"
	case ValueTypeProcedure:
		return "procedure"
	case ValueTypeUserData:
		return "userdata"
	case ValueTypeNil:
		return "nil"
	}
	return "unknown"
}

// Value represent all values that are used in the DSL[gendsl.ValueType]
//
// A Value can be converted to a go value by calling Unwrap(), the type mapping is as following:
//   - Int		    -> int64
//   - Uint       -> uint64
//   - String     -> string
//   - Float      -> float64
//   - Bool       -> bool
//   - Nil        -> nil
//   - Procedure  -> EvalFn
//   - UserData   -> any
type Value interface {
	// Type return the ValueType of a Value.
	Type() ValueType
	// Unwrap convert Value to go value, so that it can be used with type conversion syntax.
	Unwrap() any
	_value()
}

type Int int64

var _ Value = Int(0)

func (Int) _value()         {}
func (Int) Type() ValueType { return ValueTypeInt }
func (i Int) Unwrap() any   { return int64(i) }

type Uint uint64

var _ Value = Uint(0)

func (Uint) _value()         {}
func (Uint) Type() ValueType { return ValueTypeUInt }
func (u Uint) Unwrap() any   { return uint64(u) }

type Bool bool

var _ Value = Bool(false)

func (Bool) _value()         {}
func (Bool) Type() ValueType { return ValueTypeBool }
func (b Bool) Unwrap() any   { return bool(b) }

type String string

var _ Value = String("")

func (String) _value()         {}
func (String) Type() ValueType { return ValueTypeString }
func (s String) Unwrap() any   { return string(s) }

type Float float64

func (Float) _value()         {}
func (Float) Type() ValueType { return ValueTypeFloat }
func (f Float) Unwrap() any   { return float64(f) }

// Use Nil instead of nil literal to represent nil
type Nil struct{}

var _ Value = Nil{}

func (Nil) _value()         {}
func (Nil) Type() ValueType { return ValueTypeNil }
func (Nil) Unwrap() any     { return nil }

func (Nil) String() string { return "nil" }

// UserData wraps any value.
// You can use it when no type of Value can be used.
type UserData struct {
	V any
}

var _ Value = &UserData{}

func (*UserData) _value()         {}
func (*UserData) Type() ValueType { return ValueTypeUserData }
func (u *UserData) Unwrap() any   { return u.V }

// Procedure define how an expression in the format of (X Y Z...) got evaluated.
type Procedure struct {
	Eval ProcedureFn
}

var _ Value = Procedure{}

func (Procedure) _value()           {}
func (o Procedure) Type() ValueType { return ValueTypeProcedure }
func (o Procedure) Unwrap() any     { return o.Eval }
