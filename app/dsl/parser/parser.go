// =====================================================================================================================
// == LICENSE:                 Copyright (c) 2026 Kevin De Coninck.
// == SPDX-License-Identifier: LicenseRef-PolyForm-Noncommercial-1.0.0
// =====================================================================================================================

// Package parser parses Spot DSL source text into syntax data structures.
package parser

import (
	"github.com/kdeconinck/spot/dsl/ast"
	"github.com/kdeconinck/spot/dsl/lexer"
	"github.com/kdeconinck/spot/dsl/token"
	"github.com/kdeconinck/spot/location"
)

// Parse parses DSL source text into a syntax document.
func Parse(src string) (ast.Document, []Diagnostic) {
	parser := parser{
		lexer: lexer.New(src),
	}

	parser.current = parser.lexer.Next()
	parser.next = parser.lexer.Next()
	document := parser.parseDocument()

	return document, parser.diagnostics
}

type parser struct {
	lexer       lexer.Lexer
	current     token.Token
	next        token.Token
	diagnostics []Diagnostic
}

func (p *parser) expectSectionEnd(entryKind token.TokenKind) token.Token {
	if p.at(token.TokenRightBrace) {
		return p.expect(token.TokenRightBrace)
	}

	if p.at(token.TokenEOF) {
		p.addDiagnostic(token.TokenRightBrace)
	} else {
		p.addDiagnostic(entryKind)
	}

	return p.current
}

func (p *parser) expect(kind token.TokenKind) token.Token {
	if p.at(kind) {
		token := p.current
		p.advance()

		return token
	}

	token := p.current
	p.addDiagnostic(kind)

	return token
}

func (p *parser) match(kind token.TokenKind) bool {
	if !p.at(kind) {
		p.addDiagnostic(kind)

		return false
	}

	p.advance()

	return true
}

func (p *parser) consume(kind token.TokenKind) bool {
	if !p.at(kind) {
		return false
	}

	p.advance()

	return true
}

func (p *parser) at(kind token.TokenKind) bool {
	return p.current.Kind == kind
}

func (p *parser) addDiagnostic(kind token.TokenKind) {
	p.diagnostics = append(p.diagnostics, Diagnostic{
		Message: "Expected '" + kind.String() + "', found '" + p.current.Kind.String() + "'.",
		Span:    p.current.Span,
	})
}

func (p *parser) advance() {
	p.current = p.next
	p.next = p.lexer.Next()
}

func span(start, end location.Position) location.Span {
	return location.Span{
		Start: start,
		End:   end,
	}
}
