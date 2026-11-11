package interpreter

import "time"

type Callable interface {
	arity() int
	call(env Environment, args []any) any
}

type NativeFn struct {
	Arity int
	Fn    func(env Environment, args []any) any
}

func (f NativeFn) arity() int {
	return f.Arity
}

func (f NativeFn) call(env Environment, args []any) any {
	return f.Fn(env, args)
}

var clockFn = NativeFn{
	Arity: 0,
	Fn: func(env Environment, args []any) any {
		return float64(time.Now().UnixMilli()) / 1000.0
	},
}
