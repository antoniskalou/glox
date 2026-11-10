package main

type Scanner struct {
	source string
	tokens []Token
	// points to the first character in the lexeme
	start int
	// points at the character currently being considered
	current int
	// track source line current is on
	line int
}

func NewScanner(source string) *Scanner {
	return &Scanner{
		source:  source,
		start:   0,
		current: 0,
		line:    1,
	}
}

func (s *Scanner) ScanTokens() []Token {
	for !s.isAtEnd() {
		// we are the beginning of the next lexeme
		s.start = s.current
		s.scanToken()
	}

	s.tokens = append(s.tokens, Token{
		Type:    EOF,
		Lexeme:  "",
		Literal: nil,
		Line:    s.line,
	})

	return s.tokens
}

func (s *Scanner) scanToken() {
	c := s.advance()
	switch c {
	case '(':
		s.addToken(LeftParen, nil)
	case ')':
		s.addToken(RightParen, nil)
	case '{':
		s.addToken(LeftBrace, nil)
	case '}':
		s.addToken(RightBrace, nil)
	case ',':
		s.addToken(Comma, nil)
	case '.':
		s.addToken(Dot, nil)
	case '-':
		s.addToken(Minus, nil)
	case '+':
		s.addToken(Plus, nil)
	case '*':
		s.addToken(Star, nil)
	case ';':
		s.addToken(Semicolon, nil)
	default:
		// hate it, but it is what it is
		printError(s.line, "unexpected character")
	}
}

// advances scanner, returns the next char
func (s *Scanner) advance() byte {
	s.current++
	return s.source[s.current-1]
}

func (s *Scanner) addToken(typ TokenType, literal any) {
	text := s.source[s.start:s.current]
	s.tokens = append(s.tokens, Token{
		Type:    typ,
		Lexeme:  text,
		Literal: literal,
		Line:    s.line,
	})
}

// did we finish reading?
func (s *Scanner) isAtEnd() bool {
	return s.current >= len(s.source)
}
