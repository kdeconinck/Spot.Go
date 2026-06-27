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

func (p *parser) parseOptionalTokensSection() ast.TokensSection {
	if !p.at(token.TokenTokens) {
		return ast.TokensSection{}
	}

	return p.parseTokensSection()
}

func (p *parser) parseTokensSection() ast.TokensSection {
	start := p.expect(token.TokenTokens)

	if !p.match(token.TokenLeftBrace) {
		return ast.TokensSection{
			Span: start.Span,
		}
	}

	var tokens []ast.TokenDefinition

	for p.at(token.TokenIdentifier) {
		tokens = append(tokens, p.parseTokenDefinition())
	}

	end := p.expectSectionEnd(token.TokenIdentifier)

	return ast.TokensSection{
		Tokens: tokens,
		Span:   span(start.Span.Start, end.Span.End),
	}
}

func (p *parser) parseTokenDefinition() ast.TokenDefinition {
	name := p.expect(token.TokenIdentifier)
	p.expect(token.TokenEqual)
	expression := p.parseExpression(true)
	end := expression.Span.End
	var skip token.Token

	if p.at(token.TokenSkip) {
		skip = p.expect(token.TokenSkip)
		end = skip.Span.End
	}

	return ast.TokenDefinition{
		Name:       name,
		Expression: expression,
		Skip:       skip,
		Span:       span(name.Span.Start, end),
	}
}
