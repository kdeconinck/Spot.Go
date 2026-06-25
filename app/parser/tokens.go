// =====================================================================================================================
// == LICENSE:                 Copyright (c) 2026 Kevin De Coninck.
// == SPDX-License-Identifier: LicenseRef-PolyForm-Noncommercial-1.0.0
// =====================================================================================================================

// Package parser parses Spot DSL source text into syntax data structures.
package parser

import "github.com/kdeconinck/spot/syntax"

func (parser *parser) parseOptionalTokensSection() syntax.TokensSection {
	if !parser.at(syntax.TokenTokens) {
		return syntax.TokensSection{}
	}

	return parser.parseTokensSection()
}

func (parser *parser) parseTokensSection() syntax.TokensSection {
	start := parser.expect(syntax.TokenTokens)

	if !parser.match(syntax.TokenLeftBrace) {
		return syntax.TokensSection{
			Span: start.Span,
		}
	}

	var tokens []syntax.TokenDefinition

	for parser.at(syntax.TokenIdentifier) {
		tokens = append(tokens, parser.parseTokenDefinition())
	}

	end := parser.expectSectionEnd(syntax.TokenIdentifier)

	return syntax.TokensSection{
		Tokens: tokens,
		Span:   span(start.Span.Start, end.Span.End),
	}
}

func (parser *parser) parseTokenDefinition() syntax.TokenDefinition {
	name := parser.expect(syntax.TokenIdentifier)
	parser.expect(syntax.TokenEqual)
	expression := parser.parseExpression(true)
	end := expression.Span.End
	var skip syntax.Token

	if parser.at(syntax.TokenSkip) {
		skip = parser.expect(syntax.TokenSkip)
		end = skip.Span.End
	}

	return syntax.TokenDefinition{
		Name:       name,
		Expression: expression,
		Skip:       skip,
		Span:       span(name.Span.Start, end),
	}
}
