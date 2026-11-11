package interpreter

import (
	"time"

	"github.com/antoniskalou/glox/ast"
)

type Callable interface {
	arity() int
	call(env Environment, args []any) (any, error)
}

type NativeFn struct {
	Arity int
	Fn    func(env Environment, args []any) (any, error)
}

func (f NativeFn) arity() int {
	return f.Arity
}

func (f NativeFn) call(env Environment, args []any) (any, error) {
	return f.Fn(env, args)
}

func (f NativeFn) String() string {
	return "<native fn>"
}

var clockFn = NativeFn{
	Arity: 0,
	Fn: func(env Environment, args []any) (any, error) {
		return float64(time.Now().UnixMilli()) / 1000.0, nil
	},
}

type LoxFn struct {
	Declaration ast.Function
}

func (f LoxFn) arity() int {
	return len(f.Declaration.Params)
}

func (f LoxFn) call(env Environment, args []any) (any, error) {
	scoped := newEnclosed(env)
	for i, param := range f.Declaration.Params {
		scoped.define(param.Lexeme, args[i])
	}

	err := executeBlock(f.Declaration.Body, scoped)
	return nil, err
}

func (f LoxFn) String() string {
	return "<fn " + f.Declaration.Name.Lexeme + ">"
}
