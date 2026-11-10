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

/*
Grammar:

expression → equality ;
equality → comparison ( ( "!=" | "==" ) comparison )* ;
comparison → term ( ( ">" | ">=" | "<" | "<=" ) term )* ;
term → factor ( ( "-" | "+" ) factor )* ;
factor → unary ( ( "/" | "*" ) unary )* ;
unary → ( "!" | "-" ) unary | primary ;
primary → NUMBER | STRING | "true" | "false" | "nil" | "(" expression ")" ;
*/

func (p *Parser) Parse() (ast.Expr, error) {
	return p.expression()
}

func (p *Parser) expression() (ast.Expr, error) {
	return p.equality()
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

	return p.primary()
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

		_, err = p.consume(token.RightParen, "Expect ')' after expression")
		if err != nil {
			return nil, err
		}

		return ast.Grouping{Expr: expr}, nil
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
