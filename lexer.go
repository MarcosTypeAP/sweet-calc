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

type tokenKind byte

const (
	tokenKindInvalid tokenKind = iota
	tokenKindSpace
	tokenKindBracketOpen
	tokenKindBracketClose
	tokenKindNumber
	tokenKindOperator
	tokenKindOperation
)

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

func (l *lexer) consume() byte {
	if l.idx >= len(l.input) {
		panic(fmt.Errorf("input size not checked before consume: %d >= %d", l.idx, len(l.input)))
	}
	b := l.input[l.idx]
	l.idx++
	return b
}

func (l *lexer) lexNumber() {
	s := l.idx

	if l.peek() == '-' {
		l.consume()
	}
	if l.hasNext() && l.peek() == '.' {
		l.consume()
		for l.hasNext() && (l.peek() >= '0' && l.peek() <= '9') {
			l.consume()
		}
	} else {
		for l.hasNext() && (l.peek() >= '0' && l.peek() <= '9') {
			l.consume()
		}
		if l.hasNext() && l.peek() == '.' {
			l.consume()
			for l.hasNext() && (l.peek() >= '0' && l.peek() <= '9') {
				l.consume()
			}
		}
	}
	l.newToken(tokenKindNumber, s, string(l.input[s:l.idx]))
}

func (l *lexer) lexSpace() {
	s := l.idx
	for l.hasNext() && l.peek() == ' ' {
		l.consume()
	}
	l.newToken(tokenKindSpace, s, string(l.input[s:l.idx]))
}

func (l *lexer) newToken(kind tokenKind, pos int, text string) {
	l.tokens = append(l.tokens, newLexerToken(kind, pos, text))
}

func (l *lexer) newTokenConsume(kind tokenKind) {
	l.newToken(kind, l.idx, string(l.consume()))
}

func (l *lexer) newError(msg string) parsingError {
	return newParsingError(fmt.Sprintf("lexer: char %d: %s", l.idx, msg), l.idx, 1)
}

func (l *lexer) tokenize() ([]lexerToken, error) {
	for l.hasNext() {
		if l.peek() == '-' {
			lastTokenKind := tokenKindInvalid
			if len(l.tokens) > 0 {
				lastTokenKind = l.tokens[len(l.tokens)-1].kind
			}
			if lastTokenKind == tokenKindOperator || lastTokenKind == tokenKindBracketOpen || lastTokenKind == tokenKindInvalid {
				l.lexNumber()
				continue
			}
		}
		if (l.peek() >= '0' && l.peek() <= '9') || l.peek() == '.' {
			l.lexNumber()
			continue
		}

		switch l.peek() {
		case ' ':
			l.lexSpace()
		case '(':
			l.newTokenConsume(tokenKindBracketOpen)
		case ')':
			l.newTokenConsume(tokenKindBracketClose)
		case '+':
			l.newTokenConsume(tokenKindOperator)
		case '-':
			l.newTokenConsume(tokenKindOperator)
		case '*':
			l.consume()
			if l.hasNext() && l.peek() == '*' {
				l.consume()
				l.newToken(tokenKindOperator, l.idx-2, "**")
			} else {
				l.newToken(tokenKindOperator, l.idx-1, "*")
			}
		case '/':
			l.consume()
			if l.hasNext() && l.peek() == '/' {
				l.consume()
				l.newToken(tokenKindOperator, l.idx-2, "//")
			} else {
				l.newToken(tokenKindOperator, l.idx-1, "/")
			}
		case '%':
			l.newTokenConsume(tokenKindOperator)
		case 'v':
			l.newTokenConsume(tokenKindOperator)
		default:
			return nil, l.newError("unexpected character")
		}
	}

	return l.tokens, nil
}

func TokenizeInput(input []byte) ([]lexerToken, error) {
	lexer := newLexer(input)
	tokens, err := lexer.tokenize()
	if err != nil {
		return nil, err
	}
	return tokens, nil
}
