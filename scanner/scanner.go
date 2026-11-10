package scanner

import (
	"fmt"
	"strconv"

	"github.com/antoniskalou/glox/token"
)

type ScanError struct {
	Line    int
	Message string
}

func (e *ScanError) Error() string {
	return fmt.Sprintf("[line %d] Scan error: %s", e.Line, e.Message)
}

type Scanner struct {
	source string
	tokens []token.Token
	// points to the first character in the lexeme
	start int
	// points at the character currently being considered
	current int
	// track source line current is on
	line int
}

func New(source string) *Scanner {
	return &Scanner{
		source:  source,
		start:   0,
		current: 0,
		line:    1,
	}
}

func (s *Scanner) ScanTokens() ([]token.Token, error) {
	for !s.isAtEnd() {
		// we are the beginning of the next lexeme
		s.start = s.current

		if err := s.scanToken(); err != nil {
			// return what we managed to parse so far
			return s.tokens, err
		}
	}

	s.tokens = append(s.tokens, token.Token{
		Type:    token.EOF,
		Lexeme:  "",
		Literal: nil,
		Line:    s.line,
	})

	return s.tokens, nil
}

func (s *Scanner) scanToken() error {
	c := s.advance()
	switch c {
	case '(':
		s.addToken(token.LeftParen, nil)
	case ')':
		s.addToken(token.RightParen, nil)
	case '{':
		s.addToken(token.LeftBrace, nil)
	case '}':
		s.addToken(token.RightBrace, nil)
	case ',':
		s.addToken(token.Comma, nil)
	case '.':
		s.addToken(token.Dot, nil)
	case '-':
		s.addToken(token.Minus, nil)
	case '+':
		s.addToken(token.Plus, nil)
	case '*':
		s.addToken(token.Star, nil)
	case ';':
		s.addToken(token.Semicolon, nil)
	case '!':
		typ := token.Bang
		if s.match('=') {
			typ = token.BangEqual
		}
		s.addToken(typ, nil)
	case '=':
		typ := token.Equal
		if s.match('=') {
			typ = token.EqualEqual
		}
		s.addToken(typ, nil)
	case '<':
		typ := token.Less
		if s.match('=') {
			typ = token.LessEqual
		}
		s.addToken(typ, nil)
	case '>':
		typ := token.Greater
		if s.match('=') {
			typ = token.GreaterEqual
		}
		s.addToken(typ, nil)
	case '/':
		if s.match('/') {
			// comment goes until end of line
			for s.peek() != '\n' && !s.isAtEnd() {
				s.advance()
			}
		} else {
			s.addToken(token.Slash, nil)
		}
	case ' ', '\r', '\t':
		// ignore whitespace
	case '\n':
		s.line++
	case '"':
		return s.addString()
	default:
		if isDigit(c) {
			s.addNumber()
		} else if isAlpha(c) {
			s.addIdentifier()
		} else {
			return &ScanError{
				Line:    s.line,
				Message: fmt.Sprintf("unexpected character: %q", c),
			}
		}
	}

	return nil
}

var keywords = map[string]token.TokenType{
	"and":    token.And,
	"class":  token.Class,
	"else":   token.Else,
	"false":  token.False,
	"for":    token.For,
	"fun":    token.Fun,
	"if":     token.If,
	"nil":    token.Nil,
	"or":     token.Or,
	"print":  token.Print,
	"return": token.Return,
	"super":  token.Super,
	"this":   token.This,
	"true":   token.True,
	"var":    token.Var,
	"while":  token.While,
}

func (s *Scanner) addIdentifier() {
	for isAlphaNumeric(s.peek()) {
		s.advance()
	}

	text := s.source[s.start:s.current]
	typ, ok := keywords[text]
	if !ok {
		typ = token.Identifier
	}
	s.addToken(typ, nil)
}

func (s *Scanner) addNumber() {
	for isDigit(s.peek()) {
		s.advance()
	}

	// look for a fractional part
	if s.peek() == '.' && isDigit(s.peekNext()) {
		// consume the "."
		s.advance()

		for isDigit(s.peek()) {
			s.advance()
		}
	}

	f, err := strconv.ParseFloat(s.source[s.start:s.current], 64)
	if err != nil {
		// FIXME
		panic(err)
	}

	s.addToken(token.Number, f)
}

func (s *Scanner) addString() error {
	for s.peek() != '"' && !s.isAtEnd() {
		// handle multiline strings
		if s.peek() == '\n' {
			s.line++
		}
		s.advance()
	}

	if s.isAtEnd() {
		return &ScanError{
			Line: s.line,
			Message: fmt.Sprintf("Unterminated string: %s",
				s.source[s.start:s.current]),
		}
	}

	s.advance() // the closing "

	// trim surrounding quotes
	value := s.source[s.start+1 : s.current-1]
	s.addToken(token.String, value)

	return nil
}

// advances scanner, returns the next char
func (s *Scanner) advance() byte {
	s.current++
	return s.source[s.current-1]
}

// lookahead 2 bytes
func (s *Scanner) peekNext() byte {
	if s.current+1 >= len(s.source) {
		return '\x00'
	}

	return s.source[s.current+1]
}

// checks next byte without advancing scanner
func (s *Scanner) peek() byte {
	if s.isAtEnd() {
		return '\x00'
	}

	return s.source[s.current]
}

func (s *Scanner) addToken(typ token.TokenType, literal any) {
	text := s.source[s.start:s.current]
	s.tokens = append(s.tokens, token.Token{
		Type:    typ,
		Lexeme:  text,
		Literal: literal,
		Line:    s.line,
	})
}

func (s *Scanner) match(expected byte) bool {
	if s.isAtEnd() {
		return false
	}

	if s.source[s.current] != expected {
		return false
	}

	s.current++
	return true
}

// did we finish reading?
func (s *Scanner) isAtEnd() bool {
	return s.current >= len(s.source)
}

func isAlphaNumeric(c byte) bool {
	return isAlpha(c) || isDigit(c)
}

func isAlpha(c byte) bool {
	return (c >= 'a' && c <= 'z') ||
		(c >= 'A' && c <= 'Z') ||
		c == '_'
}

func isDigit(c byte) bool {
	return c >= '0' && c <= '9'
}
