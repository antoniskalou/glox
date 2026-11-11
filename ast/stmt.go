package ast

import "github.com/antoniskalou/glox/token"

type Stmt interface {
	isStmt()
}

type Expression struct {
	Expr Expr
}

func (Expression) isStmt() {}

type Print struct {
	Expr Expr
}

func (Print) isStmt() {}

type Var struct {
	Name        token.Token
	Initializer Expr
}

func (Var) isStmt() {}
