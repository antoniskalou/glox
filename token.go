package main

import "fmt"

type Token struct {
	Type    TokenType
	Lexeme  string
	Literal any // fixme: don't use any
	Line    int
}

func (t Token) String() string {
	return fmt.Sprintf("%v\t%v\t%v", t.Type, t.Lexeme, t.Literal)
}
