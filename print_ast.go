package main

import (
	"fmt"
	"strings"
)

func formatAST(expr Expr) (string, error) {
	switch expr := expr.(type) {
	case Binary:
		return parenthesize(expr.Operator.Lexeme, expr.Left, expr.Right)
	case Grouping:
		return parenthesize("group", expr.Expr)
	case Literal:
		if expr.Value == nil {
			return "nil", nil
		}
		return formatValue(expr.Value), nil
	case Unary:
		return parenthesize(expr.Operator.Lexeme, expr.Right)
	default:
		return "", fmt.Errorf("invalid AST expression: %+v", expr)
	}
}

func formatValue(v any) string {
	return fmt.Sprint(v)
}

func parenthesize(name string, exprs ...Expr) (string, error) {
	var b strings.Builder

	b.WriteByte('(')
	b.WriteString(name)
	for _, expr := range exprs {
		b.WriteByte(' ')

		s, err := formatAST(expr)
		if err != nil {
			return "", err
		}
		b.WriteString(s)
	}
	b.WriteByte(')')

	return b.String(), nil
}
