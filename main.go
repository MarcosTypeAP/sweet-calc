package main

import (
	"bytes"
	"fmt"
	"os"
	"strings"
)

const (
	ansiFgRed   = "\033[31m"
	ansiFgGreen = "\033[32m"
	ansiFgBlue  = "\033[34m"
	ansiFgReset = "\033[39m"
)

type parsingError struct {
	msg  string
	pos  int
	size int
}

func newParsingError(msg string, pos int, size int) parsingError {
	return parsingError{msg: msg, pos: pos, size: size}
}

func (p parsingError) Error() string {
	return p.msg
}

func printError(input string, err error, exit bool) {
	os.Stdout = os.Stderr

	fmt.Println()

	if _, ok := err.(parsingError); !ok {
		fmt.Printf(ansiFgRed+"error"+ansiFgReset+": %s\n", err)
		fmt.Println()
		if exit {
			os.Exit(1)
		}
	}

	perr := err.(parsingError)

	if perr.pos+perr.size > len(input) {
		input += "_"
	}

	inputLeft := input[:perr.pos]
	inputMid := input[perr.pos : perr.pos+perr.size]
	inputRight := input[perr.pos+perr.size:]

	fmt.Println("    " + inputLeft + ansiFgRed + inputMid + ansiFgReset + inputRight)
	fmt.Println("    " + ansiFgRed + strings.Repeat(" ", perr.pos) + strings.Repeat("^", perr.size) + ansiFgReset)

	fmt.Printf(ansiFgRed+"error"+ansiFgReset+" at position %d:\n", perr.pos)
	fmt.Println("    " + perr.msg)
	fmt.Println()
	if exit {
		os.Exit(1)
	}
}

func tokensToString(tokens []lexerToken) string {
	builder := strings.Builder{}
	for _, token := range tokens {
		builder.WriteString(token.text)
	}
	return builder.String()
}

func Calc(input []byte) (res float64, processed string, err error) {
	tokens, err := TokenizeInput(input)
	if err != nil {
		return 0, string(input), err
	}

	tokens, err = PreprocessTokens(tokens)
	if err != nil {
		return 0, string(input), err
	}

	tree, err := ParseTokens(tokens)
	if err != nil {
		return 0, tokensToString(tokens), err
	}

	result, err := EvalTree(tree)
	if err != nil {
		return 0, tokensToString(tokens), err
	}

	return result, tokensToString(tokens), nil
}

func EvalTree(node *parserNode) (float64, error) {
	if node.kind == tokenKindNumber {
		return node.number, nil
	}
	if node.lhs == nil || node.rhs == nil {
		panic("operator with nil lhs and rhs")
	}

	lhs, err := EvalTree(node.lhs)
	if err != nil {
		return 0, err
	}
	rhs, err := EvalTree(node.rhs)
	if err != nil {
		return 0, err
	}
	res, err := node.op.operation(lhs, rhs)
	if err != nil {
		return 0, newParsingError(
			fmt.Sprintf("eval tree: %s", err),
			node.token.pos,
			node.token.size(),
		)
	}
	return res, nil
}

func main() {
	if len(os.Args) > 1 {
		input := []byte(os.Args[1])
		processInput(input, true)
		return
	}

	buf := bytes.Buffer{}
	for {
		byteBuf := []byte{0}
		for byteBuf[0] != '\n' {
			_, err := os.Stdin.Read(byteBuf)
			if err != nil {
				panic(err)
			}
			buf.WriteByte(byteBuf[0])
		}
		input := bytes.Trim(buf.Bytes(), " \t\r\n")
		if len(input) > 0 {
			processInput(input, false)
		}
		buf.Reset()
	}
}

func processInput(input []byte, exitOnError bool) {
	res, processed, err := Calc(input)
	if string(input) != processed {
		fmt.Println("=", processed)
	}
	if err != nil {
		printError(processed, err, exitOnError)
	} else {
		fmt.Println("=", res)
	}
}
