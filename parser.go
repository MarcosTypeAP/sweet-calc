package main

import (
	"errors"
	"fmt"
	"math"
	"strconv"
)

type operator struct {
	operation  func(float64, float64) (float64, error)
	precedence int
	symbol     string
}

func (o operator) String() string {
	if len(o.symbol) == 0 {
		panic(fmt.Errorf("symbol not assigned to operator: %#v", o))
	}
	return o.symbol
}

var (
	opAddition = operator{
		operation:  func(lhs float64, rhs float64) (float64, error) { return lhs + rhs, nil },
		precedence: 1,
		symbol:     "+",
	}
	opSubtraction = operator{
		operation:  func(lhs float64, rhs float64) (float64, error) { return lhs - rhs, nil },
		precedence: 1,
		symbol:     "-",
	}
	opMultiplication = operator{
		operation:  func(lhs float64, rhs float64) (float64, error) { return lhs * rhs, nil },
		precedence: 2,
		symbol:     "*",
	}
	opModulo = operator{
		operation: func(lhs float64, rhs float64) (float64, error) {
			if rhs == 0 {
				return 0, errors.New("division by 0")
			}
			return math.Mod(lhs, rhs), nil
		},
		precedence: 2,
		symbol:     "%",
	}
	opDivision = operator{
		operation: func(lhs float64, rhs float64) (float64, error) {
			if rhs == 0 {
				return 0, errors.New("division by 0")
			}
			return lhs / rhs, nil
		},
		precedence: 2,
		symbol:     "/",
	}
	opFloorDivision = operator{
		operation: func(lhs float64, rhs float64) (float64, error) {
			if rhs == 0 {
				return 0, errors.New("division by 0")
			}
			return math.Floor(lhs / rhs), nil
		},
		precedence: 2,
		symbol:     "//",
	}
	opRoot = operator{
		operation:  func(lhs float64, rhs float64) (float64, error) { return math.Pow(rhs, 1/lhs), nil },
		precedence: 3,
		symbol:     "v",
	}
	opPower = operator{
		operation:  func(lhs float64, rhs float64) (float64, error) { return math.Pow(lhs, rhs), nil },
		precedence: 3,
		symbol:     "**",
	}
)

type parserNode struct {
	number float64
	op     *operator
	lhs    *parserNode
	rhs    *parserNode
	token  lexerToken
	kind   tokenKind
}

func newParserNodeOperation(op *operator, lhs, rhs *parserNode, token lexerToken) *parserNode {
	return &parserNode{
		op:    op,
		lhs:   lhs,
		rhs:   rhs,
		kind:  tokenKindOperation,
		token: token,
	}
}

func newParserNodeNumber(number float64, token lexerToken) *parserNode {
	return &parserNode{
		number: number,
		kind:   tokenKindNumber,
		token:  token,
	}
}

type parser struct {
	tokens []lexerToken
	idx    int
}

func newParser(tokens []lexerToken) parser {
	return parser{
		tokens: tokens,
	}
}

func (p *parser) hasNext() bool {
	return p.idx < len(p.tokens)
}

func (p *parser) lastToken() lexerToken {
	if p.idx == 0 {
		panic(fmt.Errorf("token %d: why do i call this first?", p.idx))
	}
	return p.tokens[p.idx-1]
}

func (p *parser) peek() lexerToken {
	if !p.hasNext() {
		panic(fmt.Errorf("token %d: tokens length not previously checked", p.idx))
	}
	return p.tokens[p.idx]
}

func (p *parser) consume() lexerToken {
	if !p.hasNext() {
		panic(fmt.Errorf("token %d: tokens length not previously checked", p.idx))
	}
	token := p.tokens[p.idx]
	p.idx++
	return token
}

func (p *parser) newError(msg string) parsingError {
	if !p.hasNext() {
		return newParsingError(fmt.Sprintf("parser: token %d: %s", p.idx, msg), len(p.tokens), 1)
	}
	token := p.peek()
	return newParsingError(fmt.Sprintf("parser: token %d: %s", p.idx, msg), token.pos, token.size())
}

func (p *parser) parse(inBrackets bool, minPrecedence int) (*parserNode, error) {
	lhs, err := p.parsePrimary()
	if err != nil {
		return nil, err
	}

	for p.hasNext() {
		if p.peek().kind == tokenKindBracketClose {
			if !inBrackets {
				return nil, p.newError("closing bracket never opened")
			}
			p.consume()
			return lhs, nil
		}

		var op *operator
		opToken := p.peek()

		switch opToken.kind {
		case tokenKindOperator:
			op = parseOperator(p.peek().text)
		case tokenKindBracketOpen:
			op = &opMultiplication
		case tokenKindNumber:
			if p.lastToken().kind != tokenKindBracketClose {
				return nil, p.newError("operator expected")
			}
			op = &opMultiplication
		default:
			return nil, p.newError("operator expected")
		}

		if op.precedence < minPrecedence {
			return lhs, nil
		}

		if p.peek().kind == tokenKindOperator {
			p.consume()
		}

		rhs, err := p.parse(inBrackets, op.precedence+1)
		if err != nil {
			return nil, err
		}

		lhs = newParserNodeOperation(op, lhs, rhs, opToken)
	}

	return lhs, nil
}

func (p *parser) parsePrimary() (*parserNode, error) {
	if !p.hasNext() {
		return nil, p.newError("expression expected")
	}

	if p.peek().kind == tokenKindBracketOpen {
		p.consume()
		return p.parse(true, 0)
	}
	if p.peek().kind != tokenKindNumber {
		return nil, p.newError("number expected")
	}
	token := p.consume()
	number, err := strconv.ParseFloat(token.text, 64)
	if err != nil {
		return nil, p.newError(fmt.Sprintf("parsing number: %s", err))
	}
	node := newParserNodeNumber(number, token)
	return node, nil
}

func ParseTokens(tokens []lexerToken) (*parserNode, error) {
	parser := newParser(tokens)
	tree, err := parser.parse(false, 0)
	if err != nil {
		return nil, err
	}
	return tree, nil
}

func parseOperator(text string) *operator {
	switch text {
	case opAddition.symbol:
		return &opAddition
	case opSubtraction.symbol:
		return &opSubtraction
	case opMultiplication.symbol:
		return &opMultiplication
	case opDivision.symbol:
		return &opDivision
	case opFloorDivision.symbol:
		return &opFloorDivision
	case opModulo.symbol:
		return &opModulo
	case opRoot.symbol:
		return &opRoot
	case opPower.symbol:
		return &opPower
	}
	panic("unexpected operator")
}
