package main

import "fmt"

type Token struct {
	Type    TokenType
	Lexeme  string
	Literal any // fixme: don't use any
	Line    int
}

func (t Token) String() string {
	return fmt.Sprintf("%s %s %v", t.Type, t.Lexeme, t.Literal)
}
