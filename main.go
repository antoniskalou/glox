package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
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
	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Println("> ")

		scanned := scanner.Scan()
		if !scanned {
			return
		}

		line := scanner.Text()
		run(line)
		hadError = false
	}
}

func run(source string) {
	scanner := scanner.New(source)
	tokens := scanner.ScanTokens()

	for _, token := range tokens {
		fmt.Println(token)
	}
}

func printError(line int, message string) {
	printReport(line, "", message)
}

func printReport(line int, where, message string) {
	fmt.Fprintf(os.Stderr, "[line %d] Error %s: %s", where, message)
	hadError = true
}
