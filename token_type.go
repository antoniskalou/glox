package main

type TokenType string

const (
	// single char tokens
	LeftParen  TokenType = "L_PAREN"
	RightParen           = "R_PAREN"
	LeftBrace            = "L_BRACE"
	RightBrace           = "R_BRACE"
	Comma                = "COMMA"
	Dot                  = "DOT"
	Minus                = "MINUS"
	Plus                 = "PLUS"
	Semicolon            = "SEMICOLON"
	Slash                = "SLASH"
	Star                 = "STAR"

	// one or two character tokens
	Bang         = "BANG"
	BangEqual    = "BANG_EQUAL"
	Equal        = "EQUAL"
	EqualEqual   = "EQUAL_EQUAL"
	Greater      = "GREATER"
	GreaterEqual = "GREATER_EQUAL"
	Less         = "LESS"
	LessEqual    = "LESS_EQUAL"

	// literals
	Identifier = "IDENTIFIER"
	String     = "STRING"
	Number     = "NUMBER"

	// keywords
	And    = "AND"
	Class  = "CLASS"
	Else   = "ELSE"
	False  = "FALSE"
	Fun    = "FUN"
	For    = "FOR"
	If     = "IF"
	Nil    = "NIL"
	Or     = "OR"
	Print  = "PRINT"
	Return = "RETURN"
	Super  = "SUPER"
	This   = "THIS"
	True   = "TRUE"
	Var    = "VAR"
	While  = "WHILE"

	EOF
)
