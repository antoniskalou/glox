package interpreter

import (
	"fmt"

	"github.com/antoniskalou/glox/ast"
	"github.com/antoniskalou/glox/resolver"
	"github.com/antoniskalou/glox/token"
)

type RuntimeErrorKind string

const (
	TypeError  RuntimeErrorKind = "Type error"
	ValueError RuntimeErrorKind = "Value error"
	NameError  RuntimeErrorKind = "Name error"
)

type RuntimeError struct {
	Token   token.Token
	Kind    RuntimeErrorKind
	Message string
}

func (e *RuntimeError) Error() string {
	return fmt.Sprintf("[line %d] %s: %s", e.Token.Line, e.Kind, e.Message)
}

type ReturnSignal struct {
	Value any
}

func (e *ReturnSignal) Error() string {
	return fmt.Sprintf("ReturnSignal: %v", e.Value)
}

type executor struct {
	scope   *Environment
	globals *Environment
	res     resolver.Resolution
}

func (e executor) withScope(scope Environment) executor {
	exec := e
	exec.scope = &scope
	return exec
}

type Interpreter struct {
	// global vars are treated differently and they have a "dynamic" scope, as in
	// globals defined later in the program are still accessible in places defined
	// before the global definition.
	//
	// Because of this, the store in a plain mutable map.
	globals Environment
}

func New() *Interpreter {
	return &Interpreter{
		globals: newEnv().define("clock", clockFn),
	}
}

func (i *Interpreter) Interpret(
	ast ast.AST,
	res resolver.Resolution,
) error {
	exec := executor{
		globals: &i.globals,
		scope:   &i.globals,
		res:     res,
	}

	for _, stmt := range ast {
		if err := execute(stmt, exec); err != nil {
			return err
		}
	}

	return nil
}

func execute(stmt ast.Stmt, exec executor) error {
	switch stmt := stmt.(type) {
	case ast.Expression:
		if _, err := evaluate(stmt.Expr, exec); err != nil {
			return err
		}
	case ast.Print:
		val, err := evaluate(stmt.Expr, exec)
		if err != nil {
			return err
		}
		fmt.Println(val)
	case ast.Var:
		var val any
		var err error
		// var is nil if not initalized
		if stmt.Initializer != nil {
			val, err = evaluate(stmt.Initializer, exec)
			if err != nil {
				return err
			}
		}

		*exec.scope = exec.scope.define(stmt.Name.Lexeme, val)
	case ast.Block:
		return executeBlock(stmt, exec)
	case ast.If:
		val, err := evaluate(stmt.Condition, exec)
		if err != nil {
			return err
		}

		if isTruthy(val) {
			return execute(stmt.ThenBranch, exec)
		} else if stmt.ElseBranch != nil {
			return execute(stmt.ElseBranch, exec)
		}
	case ast.While:
		for {
			val, err := evaluate(stmt.Condition, exec)
			if err != nil {
				return err
			}

			if !isTruthy(val) {
				break
			}

			if err := execute(stmt.Body, exec); err != nil {
				return err
			}
		}
	case ast.Function:
		// create the cell before capturing the environment so that the function
		// can reference itself recursively
		cell := &cell{}

		*exec.scope = exec.scope.defineCell(stmt.Name.Lexeme, cell)

		fn := LoxFn{
			Declaration: stmt,
			Closure:     *exec.scope,
			Globals:     exec.globals,
			Resolution:  exec.res,
		}

		cell.value = fn
	case ast.Return:
		var err error
		var val any

		if stmt.Value != nil {
			val, err = evaluate(stmt.Value, exec)
			if err != nil {
				return err
			}
		}

		return &ReturnSignal{Value: val}
	}

	return nil
}

func executeBlock(block ast.Block, exec executor) error {
	scoped := enclose(*exec.scope)
	return executeBlockIn(block, exec.withScope(scoped))
}

func executeBlockIn(block ast.Block, exec executor) error {
	for _, stmt := range block.Statements {
		if err := execute(stmt, exec); err != nil {
			return err
		}
	}
	return nil
}

func evaluate(expr ast.Expr, exec executor) (any, error) {
	switch expr := expr.(type) {
	case ast.Literal:
		return expr.Value, nil
	case ast.Grouping:
		return evaluate(expr.Expr, exec)
	case ast.Unary:
		return evalUnary(expr, exec)
	case ast.Binary:
		return evalBinary(expr, exec)
	case ast.Variable:
		val, ok := lookupVariable(exec, expr.Name)
		if !ok {
			return nil, &RuntimeError{
				Token:   expr.Name,
				Kind:    NameError, // is it really a "value" error?
				Message: fmt.Sprintf("variable %q is undefined", expr.Name.Lexeme),
			}
		}
		return val, nil
	case ast.Assign:
		val, err := evaluate(expr.Value, exec)
		if err != nil {
			return nil, err
		}

		if !assignVariable(exec, expr.Name, val) {
			return nil, &RuntimeError{
				Token:   expr.Name,
				Kind:    NameError,
				Message: fmt.Sprintf("undefined variable %q", expr.Name.Lexeme),
			}
		}

		return val, nil
	case ast.Logical:
		left, err := evaluate(expr.Left, exec)
		if err != nil {
			return nil, err
		}

		if expr.Operator.Type == token.Or {
			if isTruthy(left) {
				return left, nil
			}
		} else {
			if !isTruthy(left) {
				return left, nil
			}
		}

		return evaluate(expr.Right, exec)
	case ast.Call:
		callee, err := evaluate(expr.Callee, exec)
		if err != nil {
			return nil, err
		}

		args := make([]any, len(expr.Arguments))
		for i, arg := range expr.Arguments {
			val, err := evaluate(arg, exec)
			if err != nil {
				return nil, err
			}

			args[i] = val
		}

		fun, ok := callee.(Callable)
		if !ok {
			return nil, &RuntimeError{
				Token:   expr.Paren,
				Kind:    TypeError,
				Message: fmt.Sprintf("%T is not callable", callee),
			}
		}

		if len(args) != fun.arity() {
			return nil, &RuntimeError{
				Token: expr.Paren,
				Kind:  TypeError,
				Message: fmt.Sprintf("expected %d arguments, got %d",
					fun.arity(), len(args)),
			}
		}

		return fun.call(args)
	case ast.FunctionExpr:
		return LoxFn{
			Declaration: ast.Function{
				Params: expr.Params,
				Body:   expr.Body,
			},
			Closure:    *exec.scope,
			Globals:    exec.globals,
			Resolution: exec.res,
		}, nil
	default:
		panic(fmt.Errorf("logic error: no clause for expression %v", expr))
	}
}

func evalUnary(expr ast.Unary, exec executor) (any, error) {
	right, err := evaluate(expr.Right, exec)
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

func evalBinary(expr ast.Binary, exec executor) (any, error) {
	left, err := evaluate(expr.Left, exec)
	if err != nil {
		return nil, err
	}

	right, err := evaluate(expr.Right, exec)
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

func assignVariable(exec executor, name token.Token, val any) bool {
	distance, ok := exec.res.Distance(name)
	if ok {
		return exec.scope.assignAt(distance, name.Lexeme, val)
	}

	return exec.globals.assign(name, val)
}

func lookupVariable(exec executor, name token.Token) (any, bool) {
	distance, ok := exec.res.Distance(name)
	if ok {
		return exec.scope.getAt(distance, name.Lexeme)
	}

	return exec.globals.get(name)
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
