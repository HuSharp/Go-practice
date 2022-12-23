package parser

import (
	"src/lexer"
	"src/token"
)

type Parser struct {
	l *lexer.Lexer

	curToken  token.Token
	peekToken token.Token // support for more info
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{l: l}

	// make both token be set
	p.nextToken()
	p.nextToken()

	return p
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}
