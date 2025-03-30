package main

import (
	"fmt"
	"slices"
)

type preprocessor struct {
	inTokens  []lexerToken
	outTokens []lexerToken
	idx       int
}

func newPreprocessor(tokens []lexerToken) preprocessor {
	spaceCount := 0
	for _, token := range tokens {
		if token.kind == tokenKindSpace {
			spaceCount++
		}
	}
	return preprocessor{
		inTokens:  tokens,
		outTokens: make([]lexerToken, 0, len(tokens)+spaceCount*2), // + brackets
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
	if prev == tokenKindOperator || prev == tokenKindFunction {
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
				if openBrackets == 0 && p.prev().kind != tokenKindOperator && p.prev().kind != tokenKindFunction {
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

	// merge "-" with the number next to it if it isn't a subtraction
	for i := 0; i < len(p.inTokens)-1; i++ {
		prevIdx := i - 1
		prev := lexerToken{kind: tokenKindInvalid}
		if prevIdx >= 0 {
			if p.inTokens[prevIdx].kind == tokenKindSpace {
				if prevIdx-1 >= 0 {
					prevIdx--
					prev = p.inTokens[prevIdx]
				}
			} else {
				prev = p.inTokens[prevIdx]
			}
		}
		curr := p.inTokens[i]
		nextIdx := i + 1
		next := p.inTokens[nextIdx]
		if curr.text != "-" {
			continue
		}
		if next.kind == tokenKindSpace {
			if nextIdx+1 >= len(p.inTokens) {
				recalcPositions(p.inTokens, -1)
				return p.inTokens, newParsingError(
					fmt.Sprintf("preprocessor: token: %d: missing operand at the end", i),
					curr.pos,
					curr.size(),
				)
			}
			nextIdx++
			next = p.inTokens[nextIdx]
		}
		if next.kind != tokenKindNumber {
			continue
		}
		if prev.kind == tokenKindNumber {
			continue
		}
		p.inTokens[nextIdx].text = "-" + next.text
		p.inTokens = slices.Delete(p.inTokens, i, i+1)
	}

	// check two numbers have operator in between
	for i := 1; i < len(p.inTokens)-1; i++ {
		if p.inTokens[i].kind != tokenKindSpace {
			continue
		}
		prev := p.inTokens[i-1]
		next := p.inTokens[i+1]
		if prev.kind == tokenKindNumber && next.kind == tokenKindNumber {
			recalcPositions(p.inTokens, -1)
			return p.inTokens, newParsingError(
				fmt.Sprintf("preprocessor: token: %d: two consecutive operands without operator", i),
				prev.pos,
				next.pos+next.size()-prev.pos,
			)
		}
	}

	// check there isn't two operators in a row
	for i := 0; i < len(p.inTokens)-1; i++ {
		curr := p.inTokens[i]
		if curr.kind != tokenKindOperator {
			continue
		}
		nextIdx := i + 1
		next := p.inTokens[nextIdx]
		if next.kind == tokenKindSpace {
			if i+2 >= len(p.inTokens) {
				break
			}
			nextIdx = i + 2
			next = p.inTokens[nextIdx]
		}
		if next.kind == tokenKindOperator {
			recalcPositions(p.inTokens, -1)
			return p.inTokens, newParsingError(
				fmt.Sprintf("preprocessor: token: %d: two consecutive operators", i),
				curr.pos,
				next.pos+next.size()-curr.pos,
			)
		}
	}

	// remove spaces in a row (due to previous steps)
	for i := 0; i < len(p.inTokens)-2; i++ {
		curr := p.inTokens[i]
		next := p.inTokens[i+1]
		if curr.kind == tokenKindSpace && next.kind == tokenKindSpace {
			p.inTokens = slices.Delete(p.inTokens, i, (i+1)+1)
		}
	}

	if p.hasNext() {
		if token := p.consume(); token.kind != tokenKindSpace {
			p.addToken(token)
		}
	}
	for p.hasNext() {
		if p.peek().kind == tokenKindSpace {
			p.expandSpace()
			continue
		}
		p.addToken(p.consume())
	}

	recalcPositions(p.outTokens, -1)

	return p.outTokens, nil
}

func PreprocessTokens(tokens []lexerToken) ([]lexerToken, error) {
	preprocessor := newPreprocessor(tokens)
	newTokens, err := preprocessor.process()
	if err != nil {
		return newTokens, err
	}
	return newTokens, nil
}
