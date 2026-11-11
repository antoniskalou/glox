package interpreter

import "github.com/antoniskalou/glox/token"

type Environment map[string]any

// creates a "global" environment without an enclosing scope
func newGlobal() Environment {
	return Environment{}
}

func (e Environment) define(name string, val any) {
	e[name] = val
}

func (e Environment) assign(name token.Token, val any) bool {
	if _, ok := e[name.Lexeme]; ok {
		e[name.Lexeme] = val
		return true
	}

	return false
}

func (e Environment) get(name token.Token) (any, bool) {
	val, ok := e[name.Lexeme]
	return val, ok
}
