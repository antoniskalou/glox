package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/antoniskalou/glox/ast"
	"github.com/antoniskalou/glox/interpreter"
	"github.com/antoniskalou/glox/parser"
	"github.com/antoniskalou/glox/scanner"
)

const debug = false

func main() {
	inter := interpreter.New()

	switch len(os.Args) {
	case 1:
		runPrompt(inter)
	case 2:
		runFile(inter, os.Args[1])
	default:
		fmt.Println("Usage: glox [script]")
		os.Exit(64)
	}
}

func runFile(inter *interpreter.Interpreter, path string) {
	bytes, err := os.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}

	if err := run(inter, string(bytes)); err != nil {
		fmt.Fprintln(os.Stderr, err)

		switch err.(type) {
		case *interpreter.RuntimeError:
			os.Exit(70)
		default:
			os.Exit(65)
		}
	}
}

func runPrompt(inter *interpreter.Interpreter) {
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

		if err := run(inter, string(line)); err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
	}
}

func run(inter *interpreter.Interpreter, source string) error {
	scanner := scanner.New(source)
	tokens, err := scanner.ScanTokens()
	if err != nil {
		return err
	}

	parser := parser.New(tokens)
	statements, err := parser.Parse()
	if err != nil {
		return err
	}

	if debug {
		ast, err := ast.FormatMany(statements)
		if err != nil {
			return err
		}
		fmt.Println("[DEBUG] AST:", ast)
	}

	return inter.Interpret(statements)
}
