// =====================================================================================================================
// == LICENSE:                 Copyright (c) 2026 Kevin De Coninck.
// == SPDX-License-Identifier: LicenseRef-PolyForm-Noncommercial-1.0.0
// =====================================================================================================================

// Package parser parses Spot DSL source text into syntax data structures.
package parser

import (
	"github.com/kdeconinck/spot/lexer"
	"github.com/kdeconinck/spot/location"
	"github.com/kdeconinck/spot/syntax"
)

// Parse parses DSL source text into a syntax document.
func Parse(src string) (syntax.Document, []Diagnostic) {
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
	current     syntax.Token
	next        syntax.Token
	diagnostics []Diagnostic
}

func (parser *parser) expectSectionEnd(entryKind syntax.TokenKind) syntax.Token {
	if parser.at(syntax.TokenRightBrace) {
		return parser.expect(syntax.TokenRightBrace)
	}

	if parser.at(syntax.TokenEOF) {
		parser.addDiagnostic(syntax.TokenRightBrace)
	} else {
		parser.addDiagnostic(entryKind)
	}

	return parser.current
}

func (parser *parser) expect(kind syntax.TokenKind) syntax.Token {
	if parser.at(kind) {
		token := parser.current
		parser.advance()

		return token
	}

	token := parser.current
	parser.addDiagnostic(kind)

	return token
}

func (parser *parser) match(kind syntax.TokenKind) bool {
	if !parser.at(kind) {
		parser.addDiagnostic(kind)

		return false
	}

	parser.advance()

	return true
}

func (parser *parser) consume(kind syntax.TokenKind) bool {
	if !parser.at(kind) {
		return false
	}

	parser.advance()

	return true
}

func (parser *parser) at(kind syntax.TokenKind) bool {
	return parser.current.Kind == kind
}

func (parser *parser) addDiagnostic(kind syntax.TokenKind) {
	parser.diagnostics = append(parser.diagnostics, Diagnostic{
		Message: "Expected '" + kind.String() + "', found '" + parser.current.Kind.String() + "'.",
		Span:    parser.current.Span,
	})
}

func (parser *parser) advance() {
	parser.current = parser.next
	parser.next = parser.lexer.Next()
}

func span(start, end location.Position) location.Span {
	return location.Span{
		Start: start,
		End:   end,
	}
}
