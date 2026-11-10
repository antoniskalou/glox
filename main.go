package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/antoniskalou/glox/ast"
	"github.com/antoniskalou/glox/parser"
	"github.com/antoniskalou/glox/scanner"
)

// used for skipping code execution that has a known error
//
// TODO: don't use a global var
var hadError = false

func main() {
	switch len(os.Args) {
	case 1:
		runPrompt()
	case 2:
		runFile(os.Args[1])
	default:
		fmt.Println("Usage: glox [script]")
		os.Exit(64)
	}
}

func runFile(path string) {
	bytes, err := os.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}
	run(string(bytes))

	if hadError {
		os.Exit(65)
	}
}

func runPrompt() {
	stdin := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("> ")

		line, _, err := stdin.ReadLine()
		if err == io.EOF {
			os.Exit(0)
		} else if err != nil {
			log.Fatal(err)
		}

		if len(line) == 0 {
			break
		}

		run(string(line))
		hadError = false
	}
}

func run(source string) {
	scanner := scanner.New(source)
	tokens, err := scanner.ScanTokens()
	if err != nil {
		log.Fatal(err)
	}

	parser := parser.New(tokens)
	expr, err := parser.Parse()
	if err != nil {
		log.Fatal(err)
	}

	if hadError {
		return
	}

	ast, err := ast.Format(expr)
	if err != nil {
		panic(err)
	}
	fmt.Println(ast)
}
