package ast

import "github.com/antoniskalou/glox/token"

type Expr interface {
	isExpr()
}

type Binary struct {
	Left     Expr
	Right    Expr
	Operator token.Token
}

func (Binary) isExpr() {}

type Grouping struct {
	Expr Expr
}

func (Grouping) isExpr() {}

type Literal struct {
	Value any
}

func (Literal) isExpr() {}

type Unary struct {
	Operator token.Token
	Right    Expr
}

func (Unary) isExpr() {}

type Variable struct {
	Name token.Token
}

func (Variable) isExpr() {}

type Assign struct {
	Name  token.Token
	Value Expr
}

func (Assign) isExpr() {}
