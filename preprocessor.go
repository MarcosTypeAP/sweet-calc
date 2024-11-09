package main

import (
	"fmt"
	"slices"
)

type preprocessor struct {
	inTokens  []lexerToken
	outTokens []lexerToken
	firstPos  int
	idx       int
}

func newPreprocessor(tokens []lexerToken) preprocessor {
	spaceCount := 0
	for _, token := range tokens {
		if token.kind == tokenKindSpace {
			spaceCount++
		}
	}
	firstPos := 0
	if len(tokens) > 0 {
		firstPos = tokens[0].pos
	}
	return preprocessor{
		inTokens:  tokens,
		outTokens: make([]lexerToken, 0, len(tokens)+spaceCount*2), // + brackets
		firstPos:  firstPos,
	}
}

func (p *preprocessor) hasNext() bool {
	return p.idx < len(p.inTokens)
}

func (p *preprocessor) prev() lexerToken {
	if p.idx == 0 {
		panic(fmt.Errorf("token %d: why do i call this first?", p.idx))
	}
	return p.inTokens[p.idx-1]
}

func (p *preprocessor) peek() lexerToken {
	if !p.hasNext() {
		panic(fmt.Errorf("token %d: tokens length not previously checked", p.idx))
	}
	return p.inTokens[p.idx]
}

func (p *preprocessor) consume() lexerToken {
	if !p.hasNext() {
		panic(fmt.Errorf("token %d: tokens length not previously checked", p.idx))
	}
	token := p.inTokens[p.idx]
	p.idx++
	return token
}

func (p *preprocessor) addToken(token lexerToken) {
	prevCap := cap(p.outTokens)
	p.outTokens = append(p.outTokens, token)
	if cap(p.outTokens) != prevCap {
		panic("new slice created")
	}
}

func (p *preprocessor) newError(msg string) parsingError {
	if !p.hasNext() {
		return newParsingError(fmt.Sprintf("preprocessor: token %d: %s", p.idx, msg), len(p.inTokens), 1)
	}
	token := p.peek()
	return newParsingError(fmt.Sprintf("preprocessor: token %d: %s", p.idx, msg), token.pos, token.size())
}

func (p *preprocessor) expandSpace() {
	prev := p.prev().kind
	p.consume()
	if prev == tokenKindBracketOpen {
		return
	}
	if prev == tokenKindOperator {
		p.addToken(lexerToken{kind: tokenKindBracketOpen, text: "("})
		openBrackets := 0
	Loop:
		for p.hasNext() {
			switch p.peek().kind {
			case tokenKindBracketOpen:
				openBrackets++
			case tokenKindBracketClose:
				if openBrackets == 0 {
					break Loop
				}
				openBrackets--
			case tokenKindSpace:
				if openBrackets == 0 && p.prev().kind != tokenKindOperator {
					break Loop
				}
				p.expandSpace()
				continue
			}
			p.addToken(p.consume())
		}
		p.addToken(lexerToken{kind: tokenKindBracketClose, text: ")"})
		return
	}

	if !p.hasNext() {
		return
	}

	next := p.peek().kind
	if next == tokenKindBracketClose {
		return
	}
	if next == tokenKindOperator {
		p.addToken(lexerToken{kind: tokenKindBracketClose, text: ")"})
		i := len(p.outTokens) - 1
		for ; i >= 0 && p.outTokens[i].kind != tokenKindBracketOpen; i-- {
		}
		if i < 0 {
			i = 0
		}

		prevCap := cap(p.outTokens)
		p.outTokens = slices.Insert(p.outTokens, i, lexerToken{kind: tokenKindBracketOpen, text: "("})
		if cap(p.outTokens) != prevCap {
			panic("new slice created")
		}
		return
	}
}

func (p *preprocessor) process() ([]lexerToken, error) {
	if !p.hasNext() {
		return nil, fmt.Errorf("preprocessor: empty input")
	}

	for i := 1; i < len(p.inTokens)-1; i++ {
		if p.inTokens[i].kind != tokenKindSpace {
			continue
		}
		prev := p.inTokens[i-1]
		next := p.inTokens[i+1]
		if prev.kind == tokenKindNumber && next.kind == tokenKindNumber {
			return nil, newParsingError(
				fmt.Sprintf("preprocessor: token: %d: two consecutive operands without operator", i),
				prev.pos,
				next.pos+next.size()-prev.pos,
			)
		}
	}

	if token := p.consume(); token.kind != tokenKindSpace {
		p.addToken(token)
	}
	for p.hasNext() {
		if p.peek().kind == tokenKindSpace {
			p.expandSpace()
			continue
		}
		p.addToken(p.consume())
	}

	recalcPositions(p.outTokens, p.firstPos)

	return p.outTokens, nil
}

func PreprocessTokens(tokens []lexerToken) ([]lexerToken, error) {
	preprocessor := newPreprocessor(tokens)
	newTokens, err := preprocessor.process()
	if err != nil {
		return nil, err
	}
	return newTokens, nil
}
