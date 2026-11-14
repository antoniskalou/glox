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

type Resolution struct {
	// FIXME: Nystrom stores these as Expr, what do?
	locals map[token.Token]int
}

func (r Resolution) Distance(name token.Token) (int, bool) {
	depth, ok := r.locals[name]
	return depth, ok
}

type functionType uint8

const (
	functionTypeNone functionType = iota
	functionTypeFunction
)

type Resolver struct {
	scopes          *stack[map[string]bool]
	currentFunction functionType
}

func New() *Resolver {
	return &Resolver{
		scopes: newStack[map[string]bool](0),
	}
}

func (r *Resolver) Resolve(ast ast.AST) (Resolution, error) {
	result := Resolution{
		locals: make(map[token.Token]int),
	}

	if err := r.resolve(ast, &result); err != nil {
		return Resolution{}, err
	}

	return result, nil
}

func (r *Resolver) resolve(statements []ast.Stmt, result *Resolution) error {
	for _, stmt := range statements {
		if err := r.resolveStmt(stmt, result); err != nil {
			return err
		}
	}

	return nil
}

func (r *Resolver) resolveStmt(stmt ast.Stmt, result *Resolution) error {
	switch stmt := stmt.(type) {
	case ast.Var:
		if err := r.declare(stmt.Name); err != nil {
			return err
		}

		if stmt.Initializer != nil {
			if err := r.resolveExpr(stmt.Initializer, result); err != nil {
				return err
			}
		}
		r.define(stmt.Name)
	case ast.Function:
		if err := r.declare(stmt.Name); err != nil {
			return err
		}
		r.define(stmt.Name)
		return r.resolveFunction(
			stmt.Params, stmt.Body, functionTypeFunction, result,
		)
	case ast.Expression:
		return r.resolveExpr(stmt.Expr, result)
	case ast.If:
		if err := r.resolveExpr(stmt.Condition, result); err != nil {
			return err
		}

		if err := r.resolveStmt(stmt.ThenBranch, result); err != nil {
			return err
		}

		if stmt.ElseBranch != nil {
			return r.resolveStmt(stmt.ElseBranch, result)
		}
	case ast.Print:
		return r.resolveExpr(stmt.Expr, result)
	case ast.Return:
		if r.currentFunction == functionTypeNone {
			return &ResolverError{
				Token:   stmt.Keyword,
				Message: "can't return from top-level code",
			}
		}

		if stmt.Value != nil {
			return r.resolveExpr(stmt.Value, result)
		}
	case ast.While:
		if err := r.resolveExpr(stmt.Condition, result); err != nil {
			return err
		}
		return r.resolveStmt(stmt.Body, result)
	case ast.Block:
		r.beginScope()
		defer r.endScope()
		return r.resolve(stmt.Statements, result)
	default:
		panic(fmt.Sprintf("logic error, unhandled statement type: %T", stmt))
	}

	return nil
}

func (r *Resolver) resolveExpr(expr ast.Expr, result *Resolution) error {
	switch expr := expr.(type) {
	case ast.Variable:
		scope, ok := r.scopes.peek()
		if ok {
			if defined, exists := scope[expr.Name.Lexeme]; exists && !defined {
				return &ResolverError{
					Token:   expr.Name,
					Message: "can't read local variable in its own initalizer",
				}
			}
		}

		r.resolveLocal(expr.Name, result)
	case ast.Assign:
		if err := r.resolveExpr(expr.Value, result); err != nil {
			return err
		}
		r.resolveLocal(expr.Name, result)
	case ast.Binary:
		if err := r.resolveExpr(expr.Left, result); err != nil {
			return err
		}
		return r.resolveExpr(expr.Right, result)
	case ast.Logical:
		if err := r.resolveExpr(expr.Left, result); err != nil {
			return err
		}
		return r.resolveExpr(expr.Right, result)
	case ast.Unary:
		return r.resolveExpr(expr.Right, result)
	case ast.Call:
		if err := r.resolveExpr(expr.Callee, result); err != nil {
			return err
		}

		for _, arg := range expr.Arguments {
			if err := r.resolveExpr(arg, result); err != nil {
				return err
			}
		}
	case ast.Grouping:
		return r.resolveExpr(expr.Expr, result)
	case ast.Literal:
		// do nothing
	case ast.FunctionExpr:
		return r.resolveFunction(
			expr.Params, expr.Body, functionTypeFunction, result,
		)
	default:
		panic(fmt.Sprintf("logic error, unhandled expression type: %T", expr))
	}

	return nil
}

func (r *Resolver) resolveFunction(
	params []token.Token,
	body ast.Block,
	ftype functionType,
	result *Resolution,
) error {
	encosingFun := r.currentFunction
	defer func() {
		r.currentFunction = encosingFun
	}()
	r.currentFunction = ftype

	r.beginScope()
	defer r.endScope()
	for _, param := range params {
		if err := r.declare(param); err != nil {
			return err
		}
		r.define(param)
	}

	return r.resolve(body.Statements, result)
}

func (r *Resolver) resolveLocal(name token.Token, result *Resolution) {
	for i, scope := range r.scopes.iter() {
		if _, ok := scope[name.Lexeme]; ok {
			depth := r.scopes.size() - 1 - i
			result.locals[name] = depth
			return
		}
	}
}

func (r *Resolver) declare(name token.Token) error {
	scope, ok := r.scopes.peek()
	if !ok {
		return nil
	}

	if _, ok := scope[name.Lexeme]; ok {
		return &ResolverError{
			Token:   name,
			Message: "already a variable with this name in this scope",
		}
	}

	// false means "its not ready"
	scope[name.Lexeme] = false
	return nil
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
