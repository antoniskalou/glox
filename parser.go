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
package main

import "errors"

var ErrParse = errors.New("parsing error")

type Parser struct {
	tokens  []Token
	current int
}

func NewParser(tokens []Token) *Parser {
	return &Parser{
		tokens:  tokens,
		current: 0,
	}
}

func (p *Parser) Parse() Expr {
	return p.expression()
}

func (p *Parser) expression() Expr {
	return p.equality()
}

func (p *Parser) equality() Expr {
	expr := p.comparison()

	for p.match(BangEqual, EqualEqual) {
		op := p.previous()
		right := p.comparison()
		expr = Binary{
			Left:     expr,
			Operator: op,
			Right:    right,
		}
	}

	return expr
}

func (p *Parser) comparison() Expr {
	expr := p.term()

	for p.match(Greater, GreaterEqual, Less, LessEqual) {
		op := p.previous()
		right := p.term()
		expr = Binary{
			Left:     expr,
			Operator: op,
			Right:    right,
		}
	}

	return expr
}

func (p *Parser) term() Expr {
	expr := p.factor()

	for p.match(Minus, Plus) {
		op := p.previous()
		right := p.factor()
		expr = Binary{
			Left:     expr,
			Operator: op,
			Right:    right,
		}
	}

	return expr
}

func (p *Parser) factor() Expr {
	expr := p.unary()

	for p.match(Slash, Star) {
		op := p.previous()
		right := p.unary()
		expr = Binary{
			Left:     expr,
			Operator: op,
			Right:    right,
		}
	}

	return expr
}

func (p *Parser) unary() Expr {
	if p.match(Bang, Minus) {
		op := p.previous()
		// match advances remember, so this won't cause an infinite loop
		right := p.unary()
		return Unary{
			Operator: op,
			Right:    right,
		}
	}

	return p.primary()
}

func (p *Parser) primary() Expr {
	switch {
	case p.match(False):
		return Literal{false}
	case p.match(True):
		return Literal{true}
	case p.match(Nil):
		return Literal{nil}
	case p.match(Number, String):
		return Literal{p.previous().Literal}
	case p.match(LeftParen):
		expr := p.expression()
		p.consume(RightParen, "Expect ')' after expression")
		return Grouping{expr}
	}

	panic(parseError(p.peek(), "Expect expression."))
}

func (p *Parser) synchronize() {
	p.advance()

	for !p.isAtEnd() {
		if p.previous().Type == Semicolon {
			return
		}

		switch p.peek().Type {
		case Class, Fun, Var, For, If, While, Print, Return:
			return
		}

		p.advance()
	}
}

func (p *Parser) consume(typ TokenType, message string) Token {
	if p.check(typ) {
		return p.advance()
	}

	panic(parseError(p.peek(), message))
}

// checks if current token has any of the given types, if so, it
// consumes the token and returns true.
func (p *Parser) match(types ...TokenType) bool {
	for _, typ := range types {
		if p.check(typ) {
			p.advance()
			return true
		}
	}

	return false
}

// returns true if the current token is of the given type
func (p *Parser) check(typ TokenType) bool {
	if p.isAtEnd() {
		return false
	}

	return p.peek().Type == typ
}

// consumes the current token and returns it
func (p *Parser) advance() Token {
	if !p.isAtEnd() {
		p.current++
	}

	return p.previous()
}

// returns the current token we have yet to consume
func (p *Parser) peek() Token {
	// FIXME: what if it doesn't exist
	return p.tokens[p.current]
}

// returns the most recently consumed token
func (p *Parser) previous() Token {
	// FIXME: what if there's no previous
	return p.tokens[p.current-1]
}

// have we run out of tokens to parse?
func (p *Parser) isAtEnd() bool {
	return p.peek().Type == EOF
}

func parseError(token Token, message string) error {
	printError(token, message)
	return ErrParse
}
