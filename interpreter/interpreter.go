package interpreter

import (
	"fmt"

	"github.com/antoniskalou/glox/ast"
	"github.com/antoniskalou/glox/token"
)

type RuntimeError struct {
	Token   token.Token
	Message string
}

func (e *RuntimeError) Error() string {
	return fmt.Sprintf("[line %d] Runtime error: %s", e.Token.Line, e.Message)
}

type Interpreter struct{}

func New() *Interpreter {
	return &Interpreter{}
}

func evaluate(expr ast.Expr) (any, error) {
	switch expr := expr.(type) {
	case ast.Literal:
		return expr.Value, nil
	case ast.Grouping:
		return evaluate(expr.Expr)
	case ast.Unary:
		return evalUnary(expr)
	case ast.Binary:
		return evalBinary(expr)
	default:
		panic(fmt.Errorf("logic error: no clause for expression %v", expr))
	}
}

func evalUnary(expr ast.Unary) (any, error) {
	right, err := evaluate(expr.Right)
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

func evalBinary(expr ast.Binary) (any, error) {
	left, err := evaluate(expr.Left)
	if err != nil {
		return nil, err
	}

	right, err := evaluate(expr.Right)
	if err != nil {
		return nil, err
	}

	switch expr.Operator.Type {
	case token.Minus:
		return left.(float64) - right.(float64), nil
	case token.Slash:
		return left.(float64) / right.(float64), nil
	case token.Star:
		return left.(float64) * right.(float64), nil
	case token.Plus:
		if v, ok := addFloats(left, right); ok {
			return v, nil
		}

		if v, ok := addStrings(left, right); ok {
			return v, nil
		}

		return nil, &RuntimeError{
			Token:   expr.Operator,
			Message: fmt.Sprintf("type error: %T + %T", left, right),
		}
	case token.Greater:
		return left.(float64) > right.(float64), nil
	case token.GreaterEqual:
		return left.(float64) >= right.(float64), nil
	case token.Less:
		return left.(float64) < right.(float64), nil
	case token.LessEqual:
		return left.(float64) <= right.(float64), nil
	case token.BangEqual:
		return left != right, nil
	case token.EqualEqual:
		return left == right, nil
	}

	panic("unreachable")
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
