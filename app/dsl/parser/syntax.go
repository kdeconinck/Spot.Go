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

func (p *parser) parseOptionalSyntaxSection() (ast.SyntaxSection, error) {
	if !p.isAt(token.TokenSyntax) {
		return ast.SyntaxSection{}, nil
	}

	start, err := p.expect(token.TokenSyntax)

	if err != nil {
		return ast.SyntaxSection{}, err
	}

	if err := p.match(token.TokenLeftBrace); err != nil {
		return ast.SyntaxSection{}, err
	}

	firstSyntaxNode := uint32(len(p.document.SyntaxNodeList))

	for p.isAt(token.TokenNode) {
		syntaxNode, err := p.parseSyntaxNode()

		if err != nil {
			return ast.SyntaxSection{}, err
		}

		p.document.SyntaxNodeList = append(p.document.SyntaxNodeList, syntaxNode)
	}

	end, err := p.expectSectionEnd(token.TokenNode)

	if err != nil {
		return ast.SyntaxSection{}, err
	}

	return ast.SyntaxSection{
		FirstElementIdx:  firstSyntaxNode,
		AmountOfElements: uint32(len(p.document.SyntaxNodeList)) - firstSyntaxNode,
		Span:             span(start.Span.Start, end.Span.End),
	}, nil
}

func (p *parser) parseSyntaxNode() (ast.SyntaxNode, error) {
	start, err := p.expect(token.TokenNode)

	if err != nil {
		return ast.SyntaxNode{}, err
	}

	name, err := p.expect(token.TokenIdentifier)

	if err != nil {
		return ast.SyntaxNode{}, err
	}

	if _, err := p.expect(token.TokenEqual); err != nil {
		return ast.SyntaxNode{}, err
	}

	expressionID, err := p.parseSyntaxExpression()

	if err != nil {
		return ast.SyntaxNode{}, err
	}

	return ast.SyntaxNode{
		Name:       name,
		Expression: expressionID,
		Span:       span(start.Span.Start, p.syntaxExpressionNode(expressionID).Span.End),
	}, nil
}
