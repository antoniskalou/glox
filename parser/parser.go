package parser

import (
	"fmt"
	"slices"

	"github.com/antoniskalou/glox/ast"
	"github.com/antoniskalou/glox/token"
)

type ParseError struct {
	Token   token.Token
	Message string
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("[line %d] Parser error: %s", e.Token.Line, e.Message)
}

type Parser struct {
	tokens  []token.Token
	current int
}

func New(tokens []token.Token) *Parser {
	return &Parser{
		tokens:  tokens,
		current: 0,
	}
}

func (p *Parser) Parse() ([]ast.Stmt, error) {
	statements := make([]ast.Stmt, 0)

	for !p.isAtEnd() {
		stmt, err := p.declaration()
		if err != nil {
			p.synchronize()
			return nil, err
		}

		statements = append(statements, stmt)
	}

	return statements, nil
}

func (p *Parser) declaration() (ast.Stmt, error) {
	if p.match(token.Fun) {
		return p.function("function")
	}

	if p.match(token.Var) {
		return p.varDeclaration()
	}

	return p.statement()
}

func (p *Parser) function(kind string) (ast.Stmt, error) {
	name, err := p.consume(token.Identifier, "expect "+kind+" name")
	if err != nil {
		return nil, err
	}

	if _, err := p.consume(
		token.LeftParen, "expect '(' after "+kind+" name",
	); err != nil {
		return nil, err
	}

	params := make([]token.Token, 0)
	if !p.check(token.RightParen) {
		for {
			if len(params) > 255 {
				return nil, &ParseError{
					Token:   p.peek(),
					Message: "can't have more than 255 parameters",
				}
			}

			param, err := p.consume(token.Identifier, "expect parameter name")
			if err != nil {
				return nil, err
			}

			params = append(params, param)

			if !p.match(token.Comma) {
				break
			}
		}
	}

	if _, err := p.consume(
		token.RightParen, "expect ')' after parameters",
	); err != nil {
		return nil, err
	}

	if _, err := p.consume(
		token.LeftBrace, "expect '{' before "+kind+" body",
	); err != nil {
		return nil, err
	}

	body, err := p.block()
	if err != nil {
		return nil, err
	}

	return ast.Function{
		Name:   name,
		Params: params,
		Body:   body.(ast.Block), // shouldn't fail
	}, nil
}

func (p *Parser) varDeclaration() (ast.Stmt, error) {
	name, err := p.consume(token.Identifier, "expect variable name")
	if err != nil {
		return nil, err
	}

	// initializer is optional, comes with the territory of statement based
	// languages.
	var initializer ast.Expr
	if p.match(token.Equal) {
		initializer, err = p.expression()
		if err != nil {
			return nil, err
		}
	}

	p.consume(token.Semicolon, "expect ';' after variable declaration.")
	return ast.Var{Name: name, Initializer: initializer}, nil
}

func (p *Parser) statement() (ast.Stmt, error) {
	switch {
	case p.match(token.Print):
		return p.printStatement()
	case p.match(token.LeftBrace):
		return p.block()
	case p.match(token.If):
		return p.ifStatement()
	case p.match(token.While):
		return p.whileStatement()
	case p.match(token.For):
		return p.forStatement()
	case p.match(token.Return):
		return p.returnStatement()
	default:
		return p.expressionStatement()
	}
}

func (p *Parser) block() (ast.Stmt, error) {
	statements := make([]ast.Stmt, 0)

	for !p.check(token.RightBrace) && !p.isAtEnd() {
		stmt, err := p.declaration()
		if err != nil {
			return nil, err
		}
		statements = append(statements, stmt)
	}

	if _, err := p.consume(
		token.RightBrace, "expect '{' after block",
	); err != nil {
		return nil, err
	}

	return ast.Block{Statements: statements}, nil
}

func (p *Parser) ifStatement() (ast.Stmt, error) {
	if _, err := p.consume(token.LeftParen, "expect '(' after 'if'"); err != nil {
		return nil, err
	}

	condition, err := p.expression()
	if err != nil {
		return nil, err
	}

	if _, err := p.consume(
		token.RightParen, "expect ')' after if condition",
	); err != nil {
		return nil, err
	}

	thenBranch, err := p.statement()
	if err != nil {
		return nil, err
	}

	var elseBranch ast.Stmt
	if p.match(token.Else) {
		elseBranch, err = p.statement()
		if err != nil {
			return nil, err
		}
	}

	return ast.If{
		Condition:  condition,
		ThenBranch: thenBranch,
		ElseBranch: elseBranch,
	}, nil
}

func (p *Parser) whileStatement() (ast.Stmt, error) {
	if _, err := p.consume(
		token.LeftParen, "expect '(' after 'while'",
	); err != nil {
		return nil, err
	}

	condition, err := p.expression()
	if err != nil {
		return nil, err
	}

	if _, err := p.consume(
		token.RightParen, "expect ')' after condition",
	); err != nil {
		return nil, err
	}

	body, err := p.statement()
	if err != nil {
		return nil, err
	}

	return ast.While{
		Condition: condition,
		Body:      body,
	}, nil
}

// desugars into a while loop
func (p *Parser) forStatement() (ast.Stmt, error) {
	if _, err := p.consume(
		token.LeftParen, "expect '(' after 'for'",
	); err != nil {
		return nil, err
	}

	var err error
	var init ast.Stmt
	switch {
	case p.match(token.Semicolon):
		// do nothing, init stays null
	case p.match(token.Var):
		init, err = p.varDeclaration()
	default:
		init, err = p.expressionStatement()
	}

	if err != nil {
		return nil, err
	}

	var cond ast.Expr
	if !p.check(token.Semicolon) {
		cond, err = p.expression()
		if err != nil {
			return nil, err
		}
	}

	if _, err := p.consume(
		token.Semicolon, "expect ';' after loop condition",
	); err != nil {
		return nil, err
	}

	var increment ast.Expr
	if !p.check(token.RightParen) {
		increment, err = p.expression()
		if err != nil {
			return nil, err
		}
	}

	if _, err := p.consume(
		token.RightParen, "expect ')' after loop condition",
	); err != nil {
		return nil, err
	}

	body, err := p.statement()
	if err != nil {
		return nil, err
	}

	if increment != nil {
		body = ast.Block{
			Statements: []ast.Stmt{
				body,
				ast.Expression{Expr: increment},
			},
		}
	}

	if cond == nil {
		cond = ast.Literal{Value: true}
	}

	body = ast.While{Condition: cond, Body: body}

	if init != nil {
		body = ast.Block{
			Statements: []ast.Stmt{init, body},
		}
	}

	return body, nil
}

func (p *Parser) returnStatement() (ast.Stmt, error) {
	keyword := p.previous()

	var err error
	var val ast.Expr
	if !p.check(token.Semicolon) {
		val, err = p.expression()
		if err != nil {
			return nil, err
		}
	}

	if _, err := p.consume(
		token.Semicolon, "expect ';' after return value",
	); err != nil {
		return nil, err
	}

	return ast.Return{
		Keyword: keyword,
		Value:   val,
	}, nil
}

func (p *Parser) printStatement() (ast.Stmt, error) {
	expr, err := p.expression()
	if err != nil {
		return nil, err
	}

	if _, err := p.consume(
		token.Semicolon, "expect ';' after expression",
	); err != nil {
		return nil, err
	}

	return ast.Print{Expr: expr}, nil
}

func (p *Parser) expressionStatement() (ast.Stmt, error) {
	expr, err := p.expression()
	if err != nil {
		return nil, err
	}

	if _, err := p.consume(
		token.Semicolon, "expect ';' after expression",
	); err != nil {
		return nil, err
	}

	return ast.Expression{Expr: expr}, nil
}

func (p *Parser) expression() (ast.Expr, error) {
	return p.assignment()
}

func (p *Parser) assignment() (ast.Expr, error) {
	expr, err := p.or()
	if err != nil {
		return nil, err
	}

	if !p.match(token.Equal) {
		return expr, nil
	}

	equals := p.previous()
	value, err := p.assignment()
	if err != nil {
		return nil, err
	}

	if expr, ok := expr.(ast.Variable); ok {
		return ast.Assign{
			Name:  expr.Name,
			Value: value,
		}, nil
	}

	return nil, &ParseError{
		Token:   equals,
		Message: "invalid assignment target.",
	}
}

func (p *Parser) or() (ast.Expr, error) {
	expr, err := p.and()
	if err != nil {
		return nil, err
	}

	for p.match(token.Or) {
		op := p.previous()
		right, err := p.and()
		if err != nil {
			return nil, err
		}

		expr = ast.Logical{
			Left:     expr,
			Operator: op,
			Right:    right,
		}
	}

	return expr, nil
}

func (p *Parser) and() (ast.Expr, error) {
	expr, err := p.equality()
	if err != nil {
		return nil, err
	}

	for p.match(token.And) {
		op := p.previous()
		right, err := p.equality()
		if err != nil {
			return nil, err
		}
		expr = ast.Logical{
			Left:     expr,
			Operator: op,
			Right:    right,
		}
	}

	return expr, nil
}

func (p *Parser) equality() (ast.Expr, error) {
	expr, err := p.comparison()
	if err != nil {
		return nil, err
	}

	for p.match(token.BangEqual, token.EqualEqual) {
		op := p.previous()

		right, err := p.comparison()
		if err != nil {
			return nil, err
		}

		expr = ast.Binary{
			Left:     expr,
			Operator: op,
			Right:    right,
		}
	}

	return expr, nil
}

func (p *Parser) comparison() (ast.Expr, error) {
	expr, err := p.term()
	if err != nil {
		return nil, err
	}

	for p.match(token.Greater, token.GreaterEqual, token.Less, token.LessEqual) {
		op := p.previous()

		right, err := p.term()
		if err != nil {
			return nil, err
		}

		expr = ast.Binary{
			Left:     expr,
			Operator: op,
			Right:    right,
		}
	}

	return expr, nil
}

func (p *Parser) term() (ast.Expr, error) {
	expr, err := p.factor()
	if err != nil {
		return nil, err
	}

	for p.match(token.Minus, token.Plus) {
		op := p.previous()

		right, err := p.factor()
		if err != nil {
			return nil, err
		}

		expr = ast.Binary{
			Left:     expr,
			Operator: op,
			Right:    right,
		}
	}

	return expr, nil
}

func (p *Parser) factor() (ast.Expr, error) {
	expr, err := p.unary()
	if err != nil {
		return nil, err
	}

	for p.match(token.Slash, token.Star) {
		op := p.previous()

		right, err := p.unary()
		if err != nil {
			return nil, err
		}

		expr = ast.Binary{
			Left:     expr,
			Operator: op,
			Right:    right,
		}
	}

	return expr, nil
}

func (p *Parser) unary() (ast.Expr, error) {
	if p.match(token.Bang, token.Minus) {
		op := p.previous()

		// match advances remember, so this won't cause an infinite loop
		right, err := p.unary()
		if err != nil {
			return nil, err
		}

		return ast.Unary{
			Operator: op,
			Right:    right,
		}, nil
	}

	return p.call()
}

func (p *Parser) call() (ast.Expr, error) {
	expr, err := p.primary()
	if err != nil {
		return nil, err
	}

	for {
		if p.match(token.LeftParen) {
			expr, err = p.finishCall(expr)
			if err != nil {
				return nil, err
			}
		} else {
			break
		}
	}

	return expr, nil
}

func (p *Parser) finishCall(callee ast.Expr) (ast.Expr, error) {
	args := make([]ast.Expr, 0)
	if !p.check(token.RightParen) {
		for {
			expr, err := p.expression()
			if err != nil {
				return nil, err
			}

			if len(args) >= 255 {
				return nil, &ParseError{
					Token:   p.peek(),
					Message: "can't have more than 255 arguments",
				}
			}

			args = append(args, expr)
			if !p.match(token.Comma) {
				break
			}
		}
	}

	paren, err := p.consume(token.RightParen, "expect ')' after arguments")
	if err != nil {
		return nil, err
	}

	return ast.Call{
		Callee:    callee,
		Paren:     paren,
		Arguments: args,
	}, nil
}

func (p *Parser) primary() (ast.Expr, error) {
	switch {
	case p.match(token.False):
		return ast.Literal{Value: false}, nil
	case p.match(token.True):
		return ast.Literal{Value: true}, nil
	case p.match(token.Nil):
		return ast.Literal{Value: nil}, nil
	case p.match(token.Number, token.String):
		return ast.Literal{Value: p.previous().Literal}, nil
	case p.match(token.LeftParen):
		expr, err := p.expression()
		if err != nil {
			return nil, err
		}

		_, err = p.consume(token.RightParen, "expect ')' after expression")
		if err != nil {
			return nil, err
		}

		return ast.Grouping{Expr: expr}, nil
	case p.match(token.Identifier):
		return ast.Variable{
			Name: p.previous(),
		}, nil
	default:
		token := p.peek()
		return nil, &ParseError{
			Token:   token,
			Message: fmt.Sprintf("%q is not a valid expression", token.Lexeme),
		}
	}
}

func (p *Parser) synchronize() {
	p.advance()

	for !p.isAtEnd() {
		if p.previous().Type == token.Semicolon {
			return
		}

		switch p.peek().Type {
		case token.Class, token.Fun, token.Var,
			token.For, token.If, token.While,
			token.Print, token.Return:
			return
		}

		p.advance()
	}
}

func (p *Parser) consume(typ token.TokenType, message string) (token.Token, error) {
	if p.check(typ) {
		return p.advance(), nil
	}

	return token.Token{}, &ParseError{
		Token:   p.peek(),
		Message: message,
	}
}

// checks if current token has any of the given types, if so, it
// consumes the token and returns true.
func (p *Parser) match(types ...token.TokenType) bool {
	if slices.ContainsFunc(types, p.check) {
		p.advance()
		return true
	}

	return false
}

// returns true if the current token is of the given type
func (p *Parser) check(typ token.TokenType) bool {
	if p.isAtEnd() {
		return false
	}

	return p.peek().Type == typ
}

// consumes the current token and returns it
func (p *Parser) advance() token.Token {
	if !p.isAtEnd() {
		p.current++
	}

	return p.previous()
}

// returns the current token we have yet to consume
func (p *Parser) peek() token.Token {
	// FIXME: what if it doesn't exist
	return p.tokens[p.current]
}

// returns the most recently consumed token
func (p *Parser) previous() token.Token {
	// FIXME: what if there's no previous
	return p.tokens[p.current-1]
}

// have we run out of tokens to parse?
func (p *Parser) isAtEnd() bool {
	return p.peek().Type == token.EOF
}
