// =====================================================================================================================
// == LICENSE:                 Copyright (c) 2026 Kevin De Coninck.
// == SPDX-License-Identifier: LicenseRef-PolyForm-Noncommercial-1.0.0
// =====================================================================================================================

// Package parser parses Spot DSL source text into syntax data structures.
package parser

import (
	"github.com/kdeconinck/spot/dsl/ast"
	"github.com/kdeconinck/spot/dsl/token"
)

func (parser *parser) parseOptionalTokensSection() ast.TokensSection {
	if !parser.at(token.TokenTokens) {
		return ast.TokensSection{}
	}

	return parser.parseTokensSection()
}

func (parser *parser) parseTokensSection() ast.TokensSection {
	start := parser.expect(token.TokenTokens)

	if !parser.match(token.TokenLeftBrace) {
		return ast.TokensSection{
			Span: start.Span,
		}
	}

	var tokens []ast.TokenDefinition

	for parser.at(token.TokenIdentifier) {
		tokens = append(tokens, parser.parseTokenDefinition())
	}

	end := parser.expectSectionEnd(token.TokenIdentifier)

	return ast.TokensSection{
		Tokens: tokens,
		Span:   span(start.Span.Start, end.Span.End),
	}
}

func (parser *parser) parseTokenDefinition() ast.TokenDefinition {
	name := parser.expect(token.TokenIdentifier)
	parser.expect(token.TokenEqual)
	expression := parser.parseExpression(true)
	end := expression.Span.End
	var skip token.Token

	if parser.at(token.TokenSkip) {
		skip = parser.expect(token.TokenSkip)
		end = skip.Span.End
	}

	return ast.TokenDefinition{
		Name:       name,
		Expression: expression,
		Skip:       skip,
		Span:       span(name.Span.Start, end),
	}
}
