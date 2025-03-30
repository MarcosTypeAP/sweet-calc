package main

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
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
		operation: func(lhs float64, rhs float64) (float64, error) {
			res := math.Pow(rhs, 1/lhs)
			if math.IsNaN(res) {
				return 0, fmt.Errorf("%vv%v = NaN", lhs, rhs)
			}
			return res, nil
		},
		precedence: 3,
		symbol:     "v",
	}
	opPower = operator{
		operation: func(lhs float64, rhs float64) (float64, error) {
			res := math.Pow(lhs, rhs)
			if math.IsNaN(res) {
				return 0, fmt.Errorf("%v**%v = NaN", lhs, rhs)
			}
			return res, nil
		},
		precedence: 3,
		symbol:     "**",
	}
)

const FunctionPrecedence = 100

type function struct {
	fn     func(arg float64) (float64, error)
	symbol string
}

var (
	fnSin function = function{
		fn: func(x float64) (float64, error) {
			res := math.Sin(x)
			if math.IsNaN(res) {
				return 0, fmt.Errorf("sin(%v) = NaN", x)
			}
			return res, nil
		},
		symbol: "sin",
	}
	fnCos function = function{
		fn: func(x float64) (float64, error) {
			res := math.Cos(x)
			if math.IsNaN(res) {
				return 0, fmt.Errorf("cos(%v) = NaN", x)
			}
			return res, nil
		},
		symbol: "cos",
	}
	fnTan function = function{
		fn: func(x float64) (float64, error) {
			res := math.Tan(x)
			if math.IsNaN(res) {
				return 0, fmt.Errorf("tan(%v) = NaN", x)
			}
			return res, nil
		},
		symbol: "tan",
	}
	fnAsin function = function{
		fn: func(x float64) (float64, error) {
			res := math.Asin(x)
			if math.IsNaN(res) {
				return 0, fmt.Errorf("asin(%v) = NaN", x)
			}
			return res, nil
		},
		symbol: "asin",
	}
	fnAcos function = function{
		fn: func(x float64) (float64, error) {
			res := math.Acos(x)
			if math.IsNaN(res) {
				return 0, fmt.Errorf("acos(%v) = NaN", x)
			}
			return res, nil
		},
		symbol: "acos",
	}
	fnAtan function = function{
		fn: func(x float64) (float64, error) {
			return math.Atan(x), nil
		},
		symbol: "atan",
	}
)

type parserNode struct {
	data  any
	token lexerToken
}

type nodeKindOperation struct {
	op  operator
	lhs *parserNode
	rhs *parserNode
}

func newParserNodeOperation(token lexerToken, op operator, lhs, rhs *parserNode) *parserNode {
	return &parserNode{
		data: nodeKindOperation{
			op:  op,
			lhs: lhs,
			rhs: rhs,
		},
		token: token,
	}
}

type nodeKindFunction struct {
	fn  function
	arg *parserNode
}

func newParserNodeFunction(token lexerToken, fn function, arg *parserNode) *parserNode {
	return &parserNode{
		data: nodeKindFunction{
			fn:  fn,
			arg: arg,
		},
		token: token,
	}
}

type nodeKindNumber struct {
	number float64
}

func newParserNodeNumber(token lexerToken, number float64) *parserNode {
	return &parserNode{
		data: nodeKindNumber{
			number: number,
		},
		token: token,
	}
}

type nodeKindSymbol struct {
}

func newParserNodeSymbol(token lexerToken) *parserNode {
	return &parserNode{
		data:  nodeKindSymbol{},
		token: token,
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
		lastToken := p.lastToken()
		return newParsingError(fmt.Sprintf("parser: token %d: %s", p.idx, msg), lastToken.pos+lastToken.size(), 1)
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

		var op operator
		opToken := p.peek()

		switch opToken.kind {
		case tokenKindOperator:
			op = parseOperator(p.peek().text)
		case tokenKindBracketOpen:
			op = opMultiplication
		case tokenKindNumber:
			if p.lastToken().kind != tokenKindBracketClose {
				return nil, p.newError("operator expected")
			}
			op = opMultiplication
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

		lhs = newParserNodeOperation(opToken, op, lhs, rhs)
	}

	return lhs, nil
}

func parseNumber(text string) (float64, error) {
	text = strings.ReplaceAll(text, "_", "")
	return strconv.ParseFloat(text, 64)
}

func (p *parser) parsePrimary() (*parserNode, error) {
	if !p.hasNext() {
		return nil, p.newError("expression expected")
	}

	switch p.peek().kind {
	case tokenKindBracketOpen:
		p.consume()
		return p.parse(true, 0)

	case tokenKindNumber:
		token := p.consume()
		number, err := parseNumber(token.text)
		if err != nil {
			return nil, p.newError(fmt.Sprintf("parsing number: %v", err))
		}
		node := newParserNodeNumber(token, number)
		return node, nil

	case tokenKindSymbol:
		node := newParserNodeSymbol(p.consume())
		return node, nil

	case tokenKindFunction:
		token := p.consume()
		arg, err := p.parsePrimary()
		if err != nil {
			return nil, err
		}
		node := newParserNodeFunction(token, parseFunction(token.text), arg)
		return node, nil
	}
	return nil, p.newError("expression expected")
}

func ParseTokens(tokens []lexerToken) (*parserNode, error) {
	parser := newParser(tokens)
	tree, err := parser.parse(false, 0)
	if err != nil {
		return nil, err
	}
	return tree, nil
}

func parseOperator(text string) operator {
	switch text {
	case opAddition.symbol:
		return opAddition
	case opSubtraction.symbol:
		return opSubtraction
	case opMultiplication.symbol:
		return opMultiplication
	case opDivision.symbol:
		return opDivision
	case opFloorDivision.symbol:
		return opFloorDivision
	case opModulo.symbol:
		return opModulo
	case opRoot.symbol:
		return opRoot
	case opPower.symbol:
		return opPower
	}
	panic("unexpected operator")
}

func parseFunction(text string) function {
	switch text {
	case fnSin.symbol:
		return fnSin
	case fnCos.symbol:
		return fnCos
	case fnTan.symbol:
		return fnTan
	case fnAsin.symbol:
		return fnAsin
	case fnAcos.symbol:
		return fnAcos
	case fnAtan.symbol:
		return fnAtan
	}
	panic("unexpected function")
}
