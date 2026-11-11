package interpreter

import (
	"fmt"

	"github.com/antoniskalou/glox/ast"
	"github.com/antoniskalou/glox/token"
)

type RuntimeErrorKind string

const (
	TypeError  RuntimeErrorKind = "Type error"
	ValueError RuntimeErrorKind = "Value error"
)

type RuntimeError struct {
	Token   token.Token
	Kind    RuntimeErrorKind
	Message string
}

func (e *RuntimeError) Error() string {
	return fmt.Sprintf("[line %d] %s: %s", e.Token.Line, e.Kind, e.Message)
}

type Interpreter struct {
	env Environment
}

func New() *Interpreter {
	return &Interpreter{
		env: newEnvironment(),
	}
}

func (i *Interpreter) Interpret(statements []ast.Stmt) error {
	for _, stmt := range statements {
		if err := execute(stmt, i.env); err != nil {
			return err
		}
	}

	return nil
}

func execute(stmt ast.Stmt, env Environment) error {
	switch stmt := stmt.(type) {
	case ast.Expression:
		if _, err := evaluate(stmt.Expr, env); err != nil {
			return err
		}
	case ast.Print:
		val, err := evaluate(stmt.Expr, env)
		if err != nil {
			return err
		}
		fmt.Println(val)
	case ast.Var:
		var val any
		var err error
		// var is nil if not initalized
		if stmt.Initializer != nil {
			val, err = evaluate(stmt.Initializer, env)
			if err != nil {
				return err
			}
		}
		env.define(stmt.Name.Lexeme, val)
	}

	return nil
}

func evaluate(expr ast.Expr, env Environment) (any, error) {
	switch expr := expr.(type) {
	case ast.Literal:
		return expr.Value, nil
	case ast.Grouping:
		return evaluate(expr.Expr, env)
	case ast.Unary:
		return evalUnary(expr, env)
	case ast.Binary:
		return evalBinary(expr, env)
	case ast.Variable:
		val, ok := env.get(expr.Name)
		if !ok {
			return nil, &RuntimeError{
				Token:   expr.Name,
				Kind:    ValueError, // is it really a "value" error?
				Message: fmt.Sprintf("variable %q is undefined", expr.Name.Lexeme),
			}
		}
		return val, nil
	default:
		panic(fmt.Errorf("logic error: no clause for expression %v", expr))
	}
}

func evalUnary(expr ast.Unary, env Environment) (any, error) {
	right, err := evaluate(expr.Right, env)
	if err != nil {
		return nil, err
	}

	switch expr.Operator.Type {
	case token.Minus:
		return -right.(float64), nil
	case token.Bang:
		return !isTruthy(right), nil
	}

	panic("unreachable")
}

func evalBinary(expr ast.Binary, env Environment) (any, error) {
	left, err := evaluate(expr.Left, env)
	if err != nil {
		return nil, err
	}

	right, err := evaluate(expr.Right, env)
	if err != nil {
		return nil, err
	}

	typeError := func() error {
		return &RuntimeError{
			Token:   expr.Operator,
			Kind:    TypeError,
			Message: fmt.Sprintf("%T %s %T", left, expr.Operator.Lexeme, right),
		}
	}

	switch expr.Operator.Type {
	case token.Plus:
		if v, ok := addFloats(left, right); ok {
			return v, nil
		}

		if v, ok := addStrings(left, right); ok {
			return v, nil
		}

		return nil, typeError()
	case token.Minus:
		if l, r, ok := comparable(left, right); ok {
			return l - r, nil
		}

		return nil, typeError()
	case token.Slash:
		l, r, ok := comparable(left, right)
		if !ok {
			return nil, typeError()
		}

		if r == 0 {
			return nil, &RuntimeError{
				Token:   expr.Operator,
				Kind:    ValueError,
				Message: "division by zero",
			}
		}

		return l / r, nil
	case token.Star:
		if l, r, ok := comparable(left, right); ok {
			return l * r, nil
		}

		return nil, typeError()
	case token.Greater:
		if l, r, ok := comparable(left, right); ok {
			return l > r, nil
		}

		return nil, typeError()
	case token.GreaterEqual:
		if l, r, ok := comparable(left, right); ok {
			return l >= r, nil
		}

		return nil, typeError()
	case token.Less:
		if l, r, ok := comparable(left, right); ok {
			return l < r, nil
		}

		return nil, typeError()
	case token.LessEqual:
		if l, r, ok := comparable(left, right); ok {
			return l <= r, nil
		}

		return nil, typeError()
	case token.BangEqual:
		return left != right, nil
	case token.EqualEqual:
		return left == right, nil
	}

	panic("unreachable")
}

func comparable(l any, r any) (float64, float64, bool) {
	vl, okl := l.(float64)
	vr, okr := r.(float64)

	return vl, vr, okl && okr
}

func isTruthy(val any) bool {
	if val == nil {
		return false
	}

	if val, ok := val.(bool); ok {
		return val
	}

	return true
}

func addFloats(l any, r any) (float64, bool) {
	vl, okl := l.(float64)
	vr, okr := r.(float64)

	if okl && okr {
		return vl + vr, true
	}

	return 0, false
}

func addStrings(l any, r any) (string, bool) {
	vl, okl := l.(string)
	vr, okr := r.(string)

	if okl && okr {
		return vl + vr, true
	}

	return "", false
}
