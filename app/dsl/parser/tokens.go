// =====================================================================================================================
// == LICENSE:                 Copyright (c) 2026 Kevin De Coninck.
// == SPDX-License-Identifier: LicenseRef-PolyForm-Noncommercial-1.0.0
// =====================================================================================================================

// Package parser parses Spot DSL source text into syntax data structures.
package parser

import "github.com/kdeconinck/spot/dsl/token"

func (parser *parser) parseOptionalTokensSection() token.TokensSection {
	if !parser.at(token.TokenTokens) {
		return token.TokensSection{}
	}

	return parser.parseTokensSection()
}

func (parser *parser) parseTokensSection() token.TokensSection {
	start := parser.expect(token.TokenTokens)

	if !parser.match(token.TokenLeftBrace) {
		return token.TokensSection{
			Span: start.Span,
		}
	}

	var tokens []token.TokenDefinition

	for parser.at(token.TokenIdentifier) {
		tokens = append(tokens, parser.parseTokenDefinition())
	}

	end := parser.expectSectionEnd(token.TokenIdentifier)

	return token.TokensSection{
		Tokens: tokens,
		Span:   span(start.Span.Start, end.Span.End),
	}
}

func (parser *parser) parseTokenDefinition() token.TokenDefinition {
	name := parser.expect(token.TokenIdentifier)
	parser.expect(token.TokenEqual)
	expression := parser.parseExpression(true)
	end := expression.Span.End
	var skip token.Token

	if parser.at(token.TokenSkip) {
		skip = parser.expect(token.TokenSkip)
		end = skip.Span.End
	}

	return token.TokenDefinition{
		Name:       name,
		Expression: expression,
		Skip:       skip,
		Span:       span(name.Span.Start, end),
	}
}
