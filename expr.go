package main

type Expr interface {
	isExpr()
}

type Binary struct {
	Left     Expr
	Right    Expr
	Operator Token
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
	Operator Token
	Right    Expr
}

func (Unary) isExpr() {}
