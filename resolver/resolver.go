package resolver

import (
	"fmt"

	"github.com/antoniskalou/glox/ast"
	"github.com/antoniskalou/glox/token"
)

type ResolverError struct {
	Token   token.Token
	Message string
}

func (e *ResolverError) Error() string {
	return fmt.Sprintf("[line %d] Resolver error: %s", e.Token.Line, e.Message)
}

type Resolver struct {
	scopes stack[map[string]bool]
}

func (r *Resolver) Resolve(statements []ast.Stmt) error {
	for _, stmt := range statements {
		if err := r.resolveStmt(stmt); err != nil {
			return err
		}
	}
	return nil
}

func (r *Resolver) resolveStmt(stmt ast.Stmt) error {
	switch stmt := stmt.(type) {
	case ast.Var:
		r.declare(stmt.Name)
		if stmt.Initializer != nil {
			r.resolveExpr(stmt.Initializer)
		}
		r.define(stmt.Name)
	case ast.Function:
		r.declare(stmt.Name)
		r.define(stmt.Name)
		r.resolveFunction(stmt.Params, stmt.Body)
	case ast.Expression:
		r.resolveExpr(stmt.Expr)
	case ast.If:
		r.resolveExpr(stmt.Condition)
		r.resolveStmt(stmt.ThenBranch)
		if stmt.ElseBranch != nil {
			r.resolveStmt(stmt.ElseBranch)
		}
	case ast.Print:
		r.resolveExpr(stmt.Expr)
	case ast.Return:
		if stmt.Value != nil {
			r.resolveExpr(stmt.Value)
		}
	case ast.While:
		r.resolveExpr(stmt.Condition)
		r.resolveStmt(stmt.Body)
	default:
		panic(fmt.Sprintf("logic error, unhandled statement type: %T", stmt))
	}

	return nil
}

func (r *Resolver) resolveExpr(expr ast.Expr) error {
	switch expr := expr.(type) {
	case ast.Variable:
		scope, ok := r.scopes.peek()
		// FIXME: will fail if expr doesn't exist
		if ok && scope[expr.Name.Lexeme] == false {
			return &ResolverError{
				Token:   expr.Name,
				Message: "can't read local variable in its own initalizer",
			}
		}

		r.resolveLocal(expr, expr.Name)
	case ast.Assign:
		r.resolveExpr(expr.Value)
		r.resolveLocal(expr, expr.Name)
	case ast.Binary:
		r.resolveExpr(expr.Left)
		r.resolveExpr(expr.Right)
	case ast.Logical:
		r.resolveExpr(expr.Left)
		r.resolveExpr(expr.Right)
	case ast.Unary:
		r.resolveExpr(expr.Right)
	case ast.Call:
		r.resolveExpr(expr.Callee)
		for _, arg := range expr.Arguments {
			r.resolveExpr(arg)
		}
	case ast.Grouping:
		r.resolveExpr(expr.Expr)
	case ast.Literal:
		// do nothing
	case ast.FunctionExpr:
		r.resolveFunction(expr.Params, expr.Body)
	default:
		panic(fmt.Sprintf("logic error, unhandled expression type: %T", expr))
	}

	return nil
}

func (r *Resolver) resolveFunction(params []token.Token, body ast.Block) {
	r.beginScope()
	defer r.endScope()
	for _, param := range params {
		r.declare(param)
		r.define(param)
	}
	r.Resolve(body.Statements)
}

func (r *Resolver) resolveLocal(expr ast.Expr, name token.Token) {
	for depth, scope := range r.scopes.iter() {
		if _, ok := scope[name.Lexeme]; ok {
			fmt.Println("expr:", expr, "depth:", depth)
			// TODO: interpreter resolve?
			return
		}
	}
}

func (r *Resolver) declare(name token.Token) {
	scope, ok := r.scopes.peek()
	if !ok {
		return
	}
	// false means "its not ready"
	scope[name.Lexeme] = false
}

func (r *Resolver) define(name token.Token) {
	scope, ok := r.scopes.peek()
	if ok {
		// true means "fully initialized and available for use"
		scope[name.Lexeme] = true
	}
}

func (r *Resolver) beginScope() {
	r.scopes.push(make(map[string]bool))
}

func (r *Resolver) endScope() {
	_, _ = r.scopes.pop() // discard
}
