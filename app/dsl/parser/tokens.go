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

func (p *parser) parseOptionalTokensSection() (ast.TokensSection, error) {
	if !p.isAt(token.TokenTokens) {
		return ast.TokensSection{}, nil
	}

	return p.parseTokensSection()
}

func (p *parser) parseTokensSection() (ast.TokensSection, error) {
	start := p.current

	p.advance()

	if err := p.match(token.TokenLeftBrace); err != nil {
		return ast.TokensSection{}, err
	}

	firstToken := uint32(len(p.document.TokenList))

	for p.isAt(token.TokenIdentifier) {
		tokenDefinition, err := p.parseTokenDefinition()

		if err != nil {
			return ast.TokensSection{}, err
		}

		p.document.TokenList = append(p.document.TokenList, tokenDefinition)
	}

	end, err := p.expectSectionEnd(token.TokenIdentifier)

	if err != nil {
		return ast.TokensSection{}, err
	}

	return ast.TokensSection{
		FirstElementIdx:  firstToken,
		AmountOfElements: uint32(len(p.document.TokenList)) - firstToken,
		Span:             span(start.Span.Start, end.Span.End),
	}, nil
}

func (p *parser) parseTokenDefinition() (ast.TokenDefinition, error) {
	name := p.current

	p.advance()

	if _, err := p.expect(token.TokenEqual); err != nil {
		return ast.TokenDefinition{}, err
	}

	expressionID, err := p.parseExpression(true)

	if err != nil {
		return ast.TokenDefinition{}, err
	}

	end := p.expressionNode(expressionID).Span.End

	var skipToken token.Token

	if p.isAt(token.TokenSkip) {
		skipToken = p.current

		p.advance()

		end = skipToken.Span.End
	}

	return ast.TokenDefinition{
		Name:       name,
		Expression: expressionID,
		Skip:       skipToken,
		Span:       span(name.Span.Start, end),
	}, nil
}
