package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"slices"
	"strings"
	"syscall"
	"unsafe"
)

const (
	ansiFgRed     = "\033[31m"
	ansiFgGreen   = "\033[32m"
	ansiFgYellow  = "\033[33m"
	ansiFgBlue    = "\033[34m"
	ansiFgMagenta = "\033[35m"

	ansiReset     = "\033[0m"
	ansiBold      = "\033[1m"
	ansiUnderline = "\033[4m"
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
	if !repl {
		os.Stdout = os.Stderr
	}

	fmt.Println()

	var perr parsingError
	if ok := errors.As(err, &perr); !ok {
		fmt.Printf(ansiFgRed+"error"+ansiReset+": %s\n", err)
		fmt.Println()
		if !repl {
			os.Exit(1)
		}
		return
	}

	if perr.pos >= len(input) {
		if perr.pos != len(input) {
			panic("miscalculated token position")
		}
		input += " "
		perr.size = 1
	}

	inputLeft := input[:perr.pos]
	inputMid := input[perr.pos : perr.pos+perr.size]
	inputRight := input[perr.pos+perr.size:]

	fmt.Println("    " + inputLeft + ansiFgRed + inputMid + ansiReset + inputRight)
	fmt.Println("    " + ansiFgRed + strings.Repeat(" ", perr.pos) + strings.Repeat("^", perr.size) + ansiReset)

	fmt.Printf(ansiFgRed+"error"+ansiReset+" at position %d:\n", perr.pos)
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

func EvalStatement(statement []byte, vars map[string]float64) (res float64, assignedSymbol string, processed string, err error) {
	tokens, err := lexStatement(statement)
	if err != nil {
		return 0, "", string(statement), err
	}

	if len(tokens) == 0 {
		return 0, "", tokensToString(tokens), errors.New("statement eval: empty statement")
	}

	if tokens[0].kind == tokenKindSpace || tokens[len(tokens)-1].kind == tokenKindSpace {
		panic("statement not trimmed properlly")
	}

	recalcPositions(tokens, 0)

	isAssignment := false
	for _, token := range tokens {
		if token.kind == tokenKindEqual {
			isAssignment = true
			break
		}
	}

	if !isAssignment {
		res, processed, err = EvalExpression(tokens, vars)
		if err != nil {
			return 0, "", processed, err
		}
		return res, "", processed, nil
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
		return 0, symbol.text, tokensToString(tokens), newParsingError("statement eval: expression expected", idx, 1)
	}

	res, processed, err = EvalExpression(tokens[idx:], vars)
	processed = symbol.text + " = " + processed
	if err != nil {
		if perr, ok := err.(parsingError); ok {
			perr.pos += 2 // spaces around the equal
			return 0, symbol.text, processed, perr
		}
		return 0, symbol.text, processed, err
	}
	vars[symbol.text] = res

	return res, symbol.text, processed, nil
}

func recalcPositions(tokens []lexerToken, start int) {
	nextPos := start
	if start == -1 {
		nextPos = tokens[0].pos
	}
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
	switch n := node.data.(type) {
	case nodeKindNumber:
		return n.number, nil

	case nodeKindSymbol:
		if value, exists := vars[node.token.text]; exists {
			return value, nil
		}
		return 0, newParsingError(
			fmt.Sprintf("eval tree: undefined variable: %q", node.token.text),
			node.token.pos,
			node.token.size(),
		)

	case nodeKindFunction:
		if n.arg == nil {
			panic("function with nil arg")
		}
		arg, err := EvalTree(n.arg, vars)
		if err != nil {
			return 0, err
		}
		res, err := n.fn.fn(arg)
		if err != nil {
			return 0, newParsingError(
				fmt.Sprintf("eval tree: %s", err),
				node.token.pos,
				node.token.size(),
			)
		}
		return res, nil

	case nodeKindOperation:
		if n.lhs == nil || n.rhs == nil {
			panic("operator with nil lhs and rhs")
		}

		lhs, err := EvalTree(n.lhs, vars)
		if err != nil {
			return 0, err
		}
		rhs, err := EvalTree(n.rhs, vars)
		if err != nil {
			return 0, err
		}
		res, err := n.op.operation(lhs, rhs)
		if err != nil {
			return 0, newParsingError(
				fmt.Sprintf("eval tree: %s", err),
				node.token.pos,
				node.token.size(),
			)
		}
		return res, nil

	default:
		panic("unexpected parser node kind")
	}
}

func processInput(input []byte, vars map[string]float64, repl bool) {
	statements := bytes.Split(input, []byte{';'})

	for i, stmt := range statements {
		stmt = bytes.Trim(stmt, " \t\r\n")
		if len(stmt) == 0 {
			continue
		}
		res, assignedSymbol, processed, err := EvalStatement(stmt, vars)
		if err != nil {
			printError(processed, fmt.Errorf("statement %d: %w", i, err), repl)
		} else {
			integerPart := int64(math.Abs(res))
			decimalPart := math.Abs(res - float64(int64(res)))

			const maxNormalIntegerLength = 5

			str := fmt.Sprint(integerPart)
			if len(str) > maxNormalIntegerLength {
				s := make([]byte, 0, len(str)+len(str)/maxNormalIntegerLength+1)

				for i := range str {
					if i%3 == 0 && i > 0 {
						s = append(s, '_')
					}
					s = append(s, byte(str[len(str)-1-i]))
				}
				if res < 0 {
					s = append(s, '-')
				}

				slices.Reverse(s)
				str = string(s)
			} else if res < 0 {
				str = "-" + str
			}
			if decimalPart > 0 {
				str += strings.TrimRight(strings.TrimLeft(fmt.Sprintf("%.6f", decimalPart), "0"), "0")
				if str[len(str)-1] == '.' {
					str += "0"
				}
			}

			if len(assignedSymbol) > 0 {
				fmt.Println("=", assignedSymbol, "=", ansiFgYellow+str+ansiReset)
			} else {
				fmt.Println("=", ansiFgYellow+str+ansiReset)
			}
		}
	}
	if repl {
		fmt.Println()
	}
}

func getTermios() syscall.Termios {
	var term syscall.Termios

	_, _, errno := syscall.Syscall(
		syscall.SYS_IOCTL,
		os.Stdin.Fd(),
		syscall.TCGETS,
		uintptr(unsafe.Pointer(&term)),
	)
	if errno != 0 {
		panic(fmt.Errorf("error getting term attributes: errno=%d", errno))
	}

	return term
}

func setTermios(term syscall.Termios) {
	_, _, errno := syscall.Syscall(
		syscall.SYS_IOCTL,
		os.Stdin.Fd(),
		syscall.TCSETS,
		uintptr(unsafe.Pointer(&term)),
	)
	if errno != 0 {
		panic(fmt.Errorf("error setting term attributes: errno=%d", errno))
	}
}

type TerminalInput struct {
	cursorIdx int
	line      []byte
}

func (t *TerminalInput) Reset() {
	t.cursorIdx = 0
	t.line = t.line[:0]
}

func (t *TerminalInput) CursorPosition() int {
	return t.cursorIdx
}

func (t *TerminalInput) Line() string {
	return string(t.line)
}

func (t *TerminalInput) MoveCursorLeft() {
	if t.cursorIdx == 0 {
		return
	}
	t.cursorIdx--
}

func (t *TerminalInput) MoveCursorRight() {
	if t.cursorIdx == len(t.line) {
		return
	}
	t.cursorIdx++
}

func (t *TerminalInput) WriteChar(char byte) {
	t.line = slices.Insert(t.line, t.cursorIdx, char)
	t.MoveCursorRight()
}

func (t *TerminalInput) DeleteLeft() {
	if t.cursorIdx == 0 {
		return
	}
	t.line = slices.Delete(t.line, t.cursorIdx-1, t.cursorIdx)
	t.MoveCursorLeft()
}

func (t *TerminalInput) DeleteRight() {
	if t.cursorIdx == len(t.line) {
		return
	}
	t.line = slices.Delete(t.line, t.cursorIdx, t.cursorIdx+1)
}

func main() {
	vars := make(map[string]float64)
	vars["PI"] = math.Pi

	if len(os.Args) > 1 {
		input := []byte(os.Args[1])
		processInput(input, vars, false)
		return
	}

	stat, err := os.Stdin.Stat()
	if err != nil {
		panic(fmt.Errorf("error getting stdin stat: %w", err))
	}
	if stat.Mode()&os.ModeCharDevice == 0 {
		input, err := io.ReadAll(os.Stdin)
		if err == nil {
			processInput(input, vars, false)
			return
		}
	}

	termOriginal := getTermios()
	defer setTermios(termOriginal)

	termNew := termOriginal
	// See termios(3)
	termNew.Lflag &^= syscall.ICANON | syscall.ISIG | syscall.ECHO
	termNew.Cc[syscall.VMIN] = 1
	termNew.Cc[syscall.VTIME] = 0
	setTermios(termNew)

	// See https://en.wikipedia.org/wiki/Control_character
	const (
		EndOfText         = 0x03 // ^C
		EndofTransmission = 0x04 // ^D
		Escape            = 0x1b // ^[
		LineFeed          = 0x0a // \n
	)

	const (
		EraseLine  = "\033[0K\033[1K" // ESC[2K don't work properlly
		MoveCursor = "\033[%dG"       // 1-indexed
	)

	const allowedChars = ";%/()=*+-._ "

	stdinFd := int(os.Stdin.Fd())
	readChar := func() byte {
		charBuf := []byte{0}
		_, err := syscall.Read(stdinFd, charBuf)
		if err != nil {
			panic(err)
		}
		return charBuf[0]
	}

	printPrompt := func(input *TerminalInput) {
		fmt.Print(EraseLine)
		fmt.Printf(MoveCursor, 1)
		fmt.Print(ansiFgBlue + ansiBold + "> " + ansiReset + input.Line())
		fmt.Printf(MoveCursor, 3+input.CursorPosition())
	}

	history := []TerminalInput{{}} // the last one is always the new input
	historyIdx := 0

	for {
		input := &history[historyIdx]

		printPrompt(input)

	LineLoop:
		for {
			ch := readChar()

			switch {
			case ch == EndOfText:
				fmt.Println()
				return

			case ch == EndofTransmission:
				historyIdx = len(history) - 1
				history[len(history)-1].Reset()
				break LineLoop

			case ch == Escape:
				if readChar() != '[' {
					break
				}
				switch readChar() {
				case 'A': // Arrow UP
					historyIdx = max(0, historyIdx-1)
					break LineLoop

				case 'B': // Arrow Down
					historyIdx = min(len(history)-1, historyIdx+1)
					break LineLoop

				case 'C': // Arrow Right
					input.MoveCursorRight()

				case 'D': // Arrow Left
					input.MoveCursorLeft()

				case 0x33: // Delete
					if readChar() != 0x7e {
						break
					}
					input.DeleteRight()
				}

			case ch == 0x7f: // Backspace
				input.DeleteLeft()

			case ch == LineFeed:
				if input.Line() == "" {
					continue
				}

				fmt.Println()
				processInput([]byte(input.Line()), vars, true)

				switch {
				case len(history) == 1:
					history = append(history, TerminalInput{})

				case len(history) >= 2 && input.Line() != history[len(history)-2].Line():
					if historyIdx != len(history)-1 {
						history[len(history)-1] = *input
					}
					history = append(history, TerminalInput{})
				}

				history[len(history)-1].Reset()
				historyIdx = len(history) - 1

				break LineLoop

			case isAlphanumeric(ch) || strings.Contains(allowedChars, string(ch)):
				input.WriteChar(ch)
			}

			printPrompt(input)
		}
	}
}
