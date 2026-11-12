package interpreter

import (
	"github.com/antoniskalou/glox/token"
	"github.com/benbjohnson/immutable"
)

// Mutable variable storage
type cell struct {
	value any
}

type Environment struct {
	values *immutable.Map[string, *cell]
	parent *Environment
}

// creates a "global" environment without an enclosing scope
func newEnv() Environment {
	return Environment{
		values: immutable.NewMap[string, *cell](nil),
	}
}

// create a new environment that is enclosed by another environment
func enclose(parent Environment) Environment {
	return Environment{
		values: immutable.NewMap[string, *cell](nil),
		parent: &parent,
	}
}

// creates a new environment version containing the bdingin.
func (e Environment) define(name string, val any) Environment {
	return e.defineCell(name, &cell{value: val})
}

func (e Environment) defineCell(name string, cell *cell) Environment {
	return Environment{
		values: e.values.Set(name, cell),
		parent: e.parent,
	}
}

// assign updates an existing binding through mutation
func (e Environment) assign(name token.Token, val any) bool {
	if cell, ok := e.values.Get(name.Lexeme); ok {
		cell.value = val
		return true
	}

	if e.parent != nil {
		return e.parent.assign(name, val)
	}

	return false
}

func (e Environment) get(name token.Token) (any, bool) {
	if cell, ok := e.values.Get(name.Lexeme); ok {
		return cell.value, true
	}

	if e.parent != nil {
		return e.parent.get(name)
	}

	return nil, false
}
