package main

import "fmt"

type ParseError struct {
	Token   Token
	Message string
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("[line %d] Parser error: %s", e.Token.Line, e.Message)
}

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

func (p *Parser) Parse() (Expr, error) {
	return p.expression()
}

func (p *Parser) expression() (Expr, error) {
	return p.equality()
}

func (p *Parser) equality() (Expr, error) {
	expr, err := p.comparison()
	if err != nil {
		return nil, err
	}

	for p.match(BangEqual, EqualEqual) {
		op := p.previous()

		right, err := p.comparison()
		if err != nil {
			return nil, err
		}

		expr = Binary{
			Left:     expr,
			Operator: op,
			Right:    right,
		}
	}

	return expr, nil
}

func (p *Parser) comparison() (Expr, error) {
	expr, err := p.term()
	if err != nil {
		return nil, err
	}

	for p.match(Greater, GreaterEqual, Less, LessEqual) {
		op := p.previous()

		right, err := p.term()
		if err != nil {
			return nil, err
		}

		expr = Binary{
			Left:     expr,
			Operator: op,
			Right:    right,
		}
	}

	return expr, nil
}

func (p *Parser) term() (Expr, error) {
	expr, err := p.factor()
	if err != nil {
		return nil, err
	}

	for p.match(Minus, Plus) {
		op := p.previous()

		right, err := p.factor()
		if err != nil {
			return nil, err
		}

		expr = Binary{
			Left:     expr,
			Operator: op,
			Right:    right,
		}
	}

	return expr, nil
}

func (p *Parser) factor() (Expr, error) {
	expr, err := p.unary()
	if err != nil {
		return nil, err
	}

	for p.match(Slash, Star) {
		op := p.previous()

		right, err := p.unary()
		if err != nil {
			return nil, err
		}

		expr = Binary{
			Left:     expr,
			Operator: op,
			Right:    right,
		}
	}

	return expr, nil
}

func (p *Parser) unary() (Expr, error) {
	if p.match(Bang, Minus) {
		op := p.previous()

		// match advances remember, so this won't cause an infinite loop
		right, err := p.unary()
		if err != nil {
			return nil, err
		}

		return Unary{
			Operator: op,
			Right:    right,
		}, nil
	}

	return p.primary()
}

func (p *Parser) primary() (Expr, error) {
	switch {
	case p.match(False):
		return Literal{false}, nil
	case p.match(True):
		return Literal{true}, nil
	case p.match(Nil):
		return Literal{nil}, nil
	case p.match(Number, String):
		return Literal{p.previous().Literal}, nil
	case p.match(LeftParen):
		expr, err := p.expression()
		if err != nil {
			return nil, err
		}

		_, err = p.consume(RightParen, "Expect ')' after expression")
		if err != nil {
			return nil, err
		}

		return Grouping{expr}, nil
	default:
		return nil, &ParseError{
			Token:   p.peek(),
			Message: "Expect expression.",
		}
	}
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

func (p *Parser) consume(typ TokenType, message string) (Token, error) {
	if p.check(typ) {
		return p.advance(), nil
	}

	return Token{}, &ParseError{
		Token:   p.peek(),
		Message: message,
	}
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
