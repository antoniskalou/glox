package interpreter

import "github.com/antoniskalou/glox/token"

type Environment struct {
	values map[string]any
	parent *Environment
}

// creates a "global" environment without an enclosing scope
func newGlobal() Environment {
	return Environment{
		values: make(map[string]any),
	}
}

// create a new environment that is enclosed by another environment
func newEnclosed(env Environment) Environment {
	return Environment{
		values: make(map[string]any),
		parent: &env,
	}
}

func (e Environment) define(name string, val any) {
	e.values[name] = val
}

func (e Environment) assign(name token.Token, val any) bool {
	if _, ok := e.values[name.Lexeme]; ok {
		e.values[name.Lexeme] = val
		return true
	}

	// failed, try the parent
	if e.parent != nil {
		return e.parent.assign(name, val)
	}

	return false
}

func (e Environment) get(name token.Token) (any, bool) {
	val, ok := e.values[name.Lexeme]

	// try and fetch from enclosing env
	if !ok && e.parent != nil {
		return e.parent.get(name)
	}

	return val, ok
}
