package gendsl

// Env stores the mapping of identifiers and values.
// Note that an Env is not concurrently safe.
type Env struct {
	m map[string]Value
}

// NewEnv creates a new Env.
func NewEnv() *Env {
	return &Env{
		m: make(map[string]Value),
	}
}

// Clone deep copy a new env
func (e *Env) Clone() *Env {
	newE := NewEnv()
	for k, v := range e.m {
		newE = newE.WithValue(k, v)
	}

	return newE
}

// Lookup looks up the value of `id`.
// `found` report whether an value could be found in the env.
func (e *Env) Lookup(id string) (v Value, found bool) {
	v, ok := e.m[id]
	return v, ok
}

// WithProcedure registers a [gendsl.Procedure] into the env.
func (e *Env) WithProcedure(id string, p Procedure) *Env {
	e.m[id] = p
	return e
}

// WithValue registers any [gendsl.Value] into the env,
// it will panic if val == nil.
func (e *Env) WithValue(id string, val Value) *Env {
	if val == nil {
		panic("Value cannot be nil, use Nil instead")
	}
	e.m[id] = val
	return e
}

// WithInt registers a [gendsl.Int] into the env.
func (e *Env) WithInt(id string, i Int) *Env {
	e.m[id] = i
	return e
}

// WithUint registers a [gendsl.Uint] into the env.
func (e *Env) WithUint(id string, u Uint) *Env {
	e.m[id] = u
	return e
}

// WithBool registers a [gendsl.Bool] into the env.
func (e *Env) WithBool(id string, b Bool) *Env {
	e.m[id] = b
	return e
}

// WithFloat registers a [gendsl.Float] into the env.
func (e *Env) WithFloat(id string, f Float) *Env {
	e.m[id] = f
	return e
}

// WithString registers a [gendsl.String] into the env.
func (e *Env) WithString(id string, s String) *Env {
	e.m[id] = s
	return e
}

// WithUserData registers a [gendsl.UserData] into the env.
func (e *Env) WithUserData(id string, ud *UserData) *Env {
	e.m[id] = ud
	return e
}

// WithNil registers a [gendsl.Nil] into the env.
func (e *Env) WithNil(id string, n Nil) *Env {
	e.m[id] = n
	return e
}
