package interpreter

import "github.com/antoniskalou/glox/token"

type Environment map[string]any

func newEnvironment() Environment {
	return Environment{}
}

func (e Environment) define(name string, val any) {
	e[name] = val
}

func (e Environment) get(name token.Token) (any, bool) {
	val, ok := e[name.Lexeme]
	return val, ok
}
