package interpreter

import (
	"errors"
	"time"

	"github.com/antoniskalou/glox/ast"
	"github.com/antoniskalou/glox/resolver"
)

type Callable interface {
	arity() int
	call(args []any) (any, error)
}

type NativeFn struct {
	Arity int
	Fn    func(args []any) (any, error)
}

func (f NativeFn) arity() int {
	return f.Arity
}

func (f NativeFn) call(args []any) (any, error) {
	return f.Fn(args)
}

func (f NativeFn) String() string {
	return "<native fn>"
}

var clockFn = NativeFn{
	Arity: 0,
	Fn: func(args []any) (any, error) {
		return float64(time.Now().UnixMilli()) / 1000.0, nil
	},
}

type LoxFn struct {
	Declaration ast.Function
	Closure     Environment
	Globals     *Environment
	Resolution  resolver.Resolution
}

func (f LoxFn) arity() int {
	return len(f.Declaration.Params)
}

func (f LoxFn) call(args []any) (any, error) {
	local := enclose(f.Closure)
	for i, param := range f.Declaration.Params {
		local = local.define(param.Lexeme, args[i])
	}

	err := executeBlockIn(f.Declaration.Body, executor{
		scope:   &local,
		globals: f.Globals,
		res:     f.Resolution,
	})
	// explicit return, give the value back
	if signal, ok := errors.AsType[*ReturnSignal](err); ok {
		return signal.Value, nil
	}
	// implicit return
	return nil, err
}

func (f LoxFn) String() string {
	return "<fn " + f.Declaration.Name.Lexeme + ">"
}
