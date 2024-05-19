package gendsl

type Env map[string]any

func NewEnv() Env {
	return make(map[string]any)
}

func (e Env) WithOperator(name string, op Operator) Env {
	e[name] = op
	return e
}

func (e Env) WithInt(name string, i int64) Env {
	e[name] = i
	return e
}

func (e Env) WithUint(name string, i uint64) Env {
	e[name] = i
	return e
}

func (e Env) WithBool(name string, b bool) Env {
	e[name] = b
	return e
}

func (e Env) WithFloat(name string, f float64) Env {
	e[name] = f
	return e
}

func (e Env) WithString(name string, s string) Env {
	e[name] = s
	return e
}
