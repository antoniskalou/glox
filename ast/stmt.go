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

type Block struct {
	Statements []Stmt
}

func (Block) isStmt() {}

type If struct {
	Condition  Expr
	ThenBranch Stmt
	ElseBranch Stmt
}

func (If) isStmt() {}

type While struct {
	Condition Expr
	Body      Stmt
}

func (While) isStmt() {}

type Function struct {
	Name   token.Token
	Params []token.Token
	Body   Block
}

func (Function) isStmt() {}

type Return struct {
	Keyword token.Token
	Value   Expr
}

func (Return) isStmt() {}
