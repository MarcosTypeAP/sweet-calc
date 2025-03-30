package main

import (
	"fmt"
)

type lexerToken struct {
	text string
	pos  int
	kind tokenKind
}

func newLexerToken(kind tokenKind, pos int, text string) lexerToken {
	return lexerToken{kind: kind, pos: pos, text: text}
}

func (t lexerToken) size() int {
	return len(t.text)
}

func (t lexerToken) String() string {
	return fmt.Sprintf("{%q, %d, %s}", t.text, t.pos, t.kind)
}

type tokenKind byte

const (
	tokenKindInvalid tokenKind = iota
	tokenKindSpace
	tokenKindSemicolon
	tokenKindBracketOpen
	tokenKindBracketClose
	tokenKindEqual
	tokenKindSymbol
	tokenKindNumber
	tokenKindOperator
	tokenKindFunction
)

func (t tokenKind) String() string {
	switch t {
	case tokenKindInvalid:
		return "kindInvalid"
	case tokenKindSpace:
		return "kindSpace"
	case tokenKindSemicolon:
		return "kindSemicolon"
	case tokenKindBracketOpen:
		return "kindBracketOpen"
	case tokenKindBracketClose:
		return "kindBracketClose"
	case tokenKindEqual:
		return "kindEqual"
	case tokenKindSymbol:
		return "kindSymbol"
	case tokenKindNumber:
		return "kindNumber"
	case tokenKindOperator:
		return "kindOperator"
	case tokenKindFunction:
		return "kindFunction"
	}
	panic("not implemented")
}

type lexer struct {
	tokens []lexerToken
	input  []byte
	idx    int
}

func newLexer(input []byte) lexer {
	return lexer{
		input:  input,
		tokens: make([]lexerToken, 0, len(input)),
	}
}

func (l *lexer) hasNext() bool {
	return l.idx < len(l.input)
}

func (l *lexer) peek() byte {
	if l.idx >= len(l.input) {
		panic(fmt.Errorf("input size not checked before peek: %d >= %d", l.idx, len(l.input)))
	}
	return l.input[l.idx]
}

func (l *lexer) backup() {
	if l.idx <= 0 {
		panic(fmt.Errorf("input size not checked before peek: %d <= %d", l.idx, len(l.input)))
	}
	l.idx--
}

func (l *lexer) consume() byte {
	if l.idx >= len(l.input) {
		panic(fmt.Errorf("input size not checked before consume: %d >= %d", l.idx, len(l.input)))
	}
	b := l.input[l.idx]
	l.idx++
	return b
}

func isNumber(char byte) bool {
	return '0' <= char && char <= '9'
}

func isAlpha(char byte) bool {
	return ('a' <= char && char <= 'z') || ('A' <= char && char <= 'Z')
}

func isAlphanumeric(char byte) bool {
	return ('a' <= char && char <= 'z') || ('A' <= char && char <= 'Z') || '0' <= char && char <= '9'
}

func (l *lexer) lexNumber() {
	s := l.idx

	if l.peek() == '.' {
		l.consume()
		for l.hasNext() && isNumber(l.peek()) {
			l.consume()
		}
	} else {
		for l.hasNext() && (isNumber(l.peek()) || l.peek() == '_') {
			l.consume()
		}
		if l.hasNext() && l.peek() == '.' {
			l.consume()
			for l.hasNext() && isNumber(l.peek()) {
				l.consume()
			}
		}
	}

	l.addToken(tokenKindNumber, s, string(l.input[s:l.idx]))
}

func (l *lexer) lexAlphanumeric() {
	s := l.idx

	for l.hasNext() {
		ch := l.peek()
		if isAlphanumeric(ch) || ch == '_' {
			l.consume()
			continue
		}
		break
	}

	l.addToken(tokenKindSymbol, s, string(l.input[s:l.idx]))
}

func (l *lexer) lexSpace() {
	s := l.idx
	for l.hasNext() && l.peek() == ' ' {
		l.consume()
	}
	l.addToken(tokenKindSpace, s, " ")
}

func (l *lexer) lexFunction(name string) (found bool) {
	consumed := 0
	for l.hasNext() && consumed < len(name) && l.peek() == name[consumed] {
		l.consume()
		consumed++
	}
	if consumed == len(name) && (!l.hasNext() || !isAlphanumeric(l.peek())) {
		l.addToken(tokenKindFunction, l.idx-len(name), name)
		return true
	}
	for range consumed {
		l.backup()
	}
	return false
}

func (l *lexer) addToken(kind tokenKind, pos int, text string) {
	l.tokens = append(l.tokens, newLexerToken(kind, pos, text))
}

func (l *lexer) addTokenConsume(kind tokenKind) {
	l.tokens = append(l.tokens, newLexerToken(kind, l.idx, string(l.consume())))
}

func (l *lexer) newError(msg string) parsingError {
	return newParsingError(fmt.Sprintf("lexer: char %d: %s", l.idx, msg), l.idx, 1)
}

func (l *lexer) tokenize() ([]lexerToken, error) {
	for l.hasNext() {
		ch := l.peek()

		if isNumber(ch) || ch == '.' {
			l.lexNumber()
			continue
		}

		switch ch {
		case 'v':
			l.consume()
			if !l.hasNext() || !isAlpha(l.peek()) {
				l.addToken(tokenKindOperator, l.idx-1, "v")
				continue
			}
			l.backup()
			l.lexAlphanumeric()
		case 's':
			if l.lexFunction("sin") {
				continue
			}
		case 'c':
			if l.lexFunction("cos") {
				continue
			}
		case 't':
			if l.lexFunction("tan") {
				continue
			}
		case 'a':
			if l.lexFunction("asin") {
				continue
			}
			if l.lexFunction("acos") {
				continue
			}
			if l.lexFunction("atan") {
				continue
			}
		}

		if ch != 'v' && isAlpha(ch) {
			l.lexAlphanumeric()
			continue
		}

		switch ch {
		case ' ':
			l.lexSpace()
		case ';':
			l.addTokenConsume(tokenKindSemicolon)
		case '=':
			l.addTokenConsume(tokenKindEqual)
		case '(':
			l.addTokenConsume(tokenKindBracketOpen)
		case ')':
			l.addTokenConsume(tokenKindBracketClose)
		case '+':
			l.addTokenConsume(tokenKindOperator)
		case '-':
			l.addTokenConsume(tokenKindOperator)
		case '*':
			l.consume()
			if l.hasNext() && l.peek() == '*' {
				l.consume()
				l.addToken(tokenKindOperator, l.idx-2, "**")
			} else {
				l.addToken(tokenKindOperator, l.idx-1, "*")
			}
		case '/':
			l.consume()
			if l.hasNext() && l.peek() == '/' {
				l.consume()
				l.addToken(tokenKindOperator, l.idx-2, "//")
			} else {
				l.addToken(tokenKindOperator, l.idx-1, "/")
			}
		case '%':
			l.addTokenConsume(tokenKindOperator)
		default:
			return nil, l.newError("unexpected character")
		}
	}

	return l.tokens, nil
}

func lexStatement(input []byte) ([]lexerToken, error) {
	lexer := newLexer(input)
	tokens, err := lexer.tokenize()
	if err != nil {
		return nil, err
	}
	return tokens, nil
}
