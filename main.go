package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
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

func printError(input string, err error, repl bool) {
	os.Stdout = os.Stderr

	fmt.Println()

	var perr parsingError
	if ok := errors.As(err, &perr); !ok {
		fmt.Printf(ansiFgRed+"error"+ansiFgReset+": %s\n", err)
		fmt.Println()
		if !repl {
			os.Exit(1)
		}
		return
	}

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
	if !repl {
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

func EvalStatement(statement []byte, vars map[string]float64) (res float64, isAssignment bool, processed string, err error) {
	tokens, err := lexStatement(statement)
	if err != nil {
		return 0, false, string(statement), err
	}

	if len(tokens) == 0 {
		return 0, false, tokensToString(tokens), errors.New("statement eval: empty statement")
	}

	if tokens[0].kind == tokenKindSpace || tokens[len(tokens)-1].kind == tokenKindSpace {
		panic("statement not trimmed properlly")
	}

	recalcPositions(tokens, 0)

	isAssignment = false
	for _, token := range tokens {
		if token.kind == tokenKindEqual {
			isAssignment = true
			break
		}
	}

	if !isAssignment {
		res, processed, err = EvalExpression(tokens, vars)
		if err != nil {
			return 0, false, processed, err
		}
		return res, false, processed, nil
	}

	symbol := tokens[0]
	idx := 1

	if idx < len(tokens) && tokens[idx].kind == tokenKindSpace {
		idx++
	}

	idx++ // the "="

	if idx < len(tokens) && tokens[idx].kind == tokenKindSpace {
		idx++
	}

	if idx >= len(tokens) {
		return 0, true, tokensToString(tokens), newParsingError("statement eval: expression expected", idx, 1)
	}

	res, processed, err = EvalExpression(tokens[idx:], vars)
	processed = symbol.text + " = " + processed
	if err != nil {
		if perr, ok := err.(parsingError); ok {
			perr.pos += 2
			return 0, true, processed, perr
		}
		return 0, true, processed, err
	}
	vars[symbol.text] = res
	return res, true, processed, nil
}

func recalcPositions(tokens []lexerToken, start int) {
	nextPos := start
	for i := range tokens {
		tokens[i].pos = nextPos
		nextPos += tokens[i].size()
	}
}

func EvalExpression(tokens []lexerToken, vars map[string]float64) (res float64, processed string, err error) {
	tokens, err = PreprocessTokens(tokens)
	if err != nil {
		return 0, tokensToString(tokens), err
	}

	tree, err := ParseTokens(tokens)
	if err != nil {
		return 0, tokensToString(tokens), err
	}

	result, err := EvalTree(tree, vars)
	if err != nil {
		return 0, tokensToString(tokens), err
	}

	return result, tokensToString(tokens), nil
}

func EvalTree(node *parserNode, vars map[string]float64) (float64, error) {
	if node.kind == tokenKindNumber {
		return node.number, nil
	}
	if node.kind == tokenKindSymbol {
		if value, exists := vars[node.token.text]; exists {
			return value, nil
		}
		return 0, newParsingError(
			fmt.Sprintf("eval tree: undefined variable: %q", node.token.text),
			node.token.pos,
			node.token.size(),
		)
	}
	if node.lhs == nil || node.rhs == nil {
		panic("operator with nil lhs and rhs")
	}

	lhs, err := EvalTree(node.lhs, vars)
	if err != nil {
		return 0, err
	}
	rhs, err := EvalTree(node.rhs, vars)
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

func processInput(input []byte, vars map[string]float64, repl bool) {
	statements := bytes.Split(input, []byte{';'})
	for i, stmt := range statements {
		stmt = bytes.Trim(stmt, " \t\r\n")
		res, isAssignment, processed, err := EvalStatement(stmt, vars)
		if isAssignment {
			fmt.Println("=", processed)
		}
		if err != nil {
			printError(processed, fmt.Errorf("statement %d: %w", i, err), repl)
		} else {
			fmt.Println("=", res)
		}
		if repl || len(statements) > 0 && i < len(statements)-1 {
			fmt.Println()
		}
	}
}

func main() {
	vars := make(map[string]float64)

	if len(os.Args) > 1 {
		input := []byte(os.Args[1])
		processInput(input, vars, false)
		return
	}

	fmt.Println("Sweet Calculator REPL")

	buf := bytes.Buffer{}
	for {
		byteBuf := []byte{0}
		for byteBuf[0] != '\n' {
			_, err := os.Stdin.Read(byteBuf)
			if err != nil {
				if err == io.EOF {
					return
				}
				panic(err)
			}
			buf.WriteByte(byteBuf[0])
		}
		if buf.Len() > 0 {
			processInput(buf.Bytes(), vars, true)
		}
		buf.Reset()
	}
}
