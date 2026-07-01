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

	expressionID, err := p.parseStructuredSyntaxNodeExpression()

	if err != nil {
		return ast.SyntaxNode{}, err
	}

	return ast.SyntaxNode{
		Name:       name,
		Expression: expressionID,
		Span:       span(start.Span.Start, p.syntaxExpressionNode(expressionID).Span.End),
	}, nil
}

func (p *parser) parseStructuredSyntaxNodeExpression() (ast.SyntaxExpressionID, error) {
	if _, err := p.expect(token.TokenLeftBrace); err != nil {
		return 0, err
	}

	first, err := p.parseStructuredSyntaxEntry()

	if err != nil {
		return 0, err
	}

	if p.isAt(token.TokenRightBrace) {
		if _, err := p.expect(token.TokenRightBrace); err != nil {
			return 0, err
		}

		return first, nil
	}

	var buffer [4]ast.SyntaxExpressionID
	entries := buffer[:0]
	entries = append(entries, first)

	for !p.isAt(token.TokenRightBrace) {
		entry, err := p.parseStructuredSyntaxEntry()

		if err != nil {
			return 0, err
		}

		entries = append(entries, entry)
	}

	end, err := p.expect(token.TokenRightBrace)

	if err != nil {
		return 0, err
	}

	return p.addSyntaxExpressionBranch(
		ast.SyntaxExpressionConcatenation,
		entries,
		span(p.syntaxExpressionNode(first).Span.Start, end.Span.End),
	), nil
}

func (p *parser) parseStructuredSyntaxEntry() (ast.SyntaxExpressionID, error) {
	if p.isAt(token.TokenOneOf) {
		return p.parseStructuredSyntaxOneOf()
	}

	if !p.isAt(token.TokenIdentifier) {
		return p.parseStructuredSyntaxUnnamedEntry()
	}

	first := p.current
	p.advance()

	operator := token.Token{}

	if p.isAt(token.TokenQuestion) || p.isAt(token.TokenStar) || p.isAt(token.TokenPlus) {
		operator = p.current
		p.advance()
	}

	if p.isAt(token.TokenColon) {
		p.advance()

		inner, err := p.parseStructuredSyntaxFieldTarget()

		if err != nil {
			return 0, err
		}

		if operator.Kind == token.TokenQuestion || operator.Kind == token.TokenStar || operator.Kind == token.TokenPlus {
			inner = p.wrapSyntaxExpressionWithLeadingRepetition(inner, operator)
		}

		return p.addSyntaxExpressionNode(ast.SyntaxExpressionNode{
			Kind:             ast.SyntaxExpressionCapture,
			Field:            first,
			FirstElementIdx:  p.appendSyntaxExpressionChildren(inner),
			AmountOfElements: 1,
			Span:             span(first.Span.Start, p.syntaxExpressionNode(inner).Span.End),
		}), nil
	}

	return p.finishStructuredSyntaxReference(first, operator), nil
}

func (p *parser) parseStructuredSyntaxUnnamedEntry() (ast.SyntaxExpressionID, error) {
	if p.isAt(token.TokenOneOf) {
		return p.parseStructuredSyntaxOneOf()
	}

	if p.isAt(token.TokenAny) {
		anyToken := p.current
		p.advance()

		expressionID := p.addSyntaxExpressionNode(ast.SyntaxExpressionNode{
			Kind: ast.SyntaxExpressionAny,
			Span: anyToken.Span,
		})

		if p.isAt(token.TokenQuestion) || p.isAt(token.TokenStar) || p.isAt(token.TokenPlus) {
			operator := p.current
			p.advance()

			return p.wrapSyntaxExpressionWithRepetition(expressionID, operator), nil
		}

		return expressionID, nil
	}

	if p.isAt(token.TokenLeftParen) {
		group, err := p.parseGroupedSyntaxExpression()

		if err != nil {
			return 0, err
		}

		if p.isAt(token.TokenQuestion) || p.isAt(token.TokenStar) || p.isAt(token.TokenPlus) {
			operator := p.current
			p.advance()

			return p.wrapSyntaxExpressionWithRepetition(group, operator), nil
		}

		return group, nil
	}

	reference, err := p.expect(token.TokenIdentifier)

	if err != nil {
		return 0, err
	}

	operator := token.Token{}

	if p.isAt(token.TokenQuestion) || p.isAt(token.TokenStar) || p.isAt(token.TokenPlus) {
		operator = p.current
		p.advance()
	}

	return p.finishStructuredSyntaxReference(reference, operator), nil
}

func (p *parser) parseStructuredSyntaxFieldTarget() (ast.SyntaxExpressionID, error) {
	if p.isAt(token.TokenOneOf) {
		return p.parseStructuredSyntaxOneOf()
	}

	return p.parseStructuredSyntaxUnnamedEntry()
}

func (p *parser) parseStructuredSyntaxOneOf() (ast.SyntaxExpressionID, error) {
	start, err := p.expect(token.TokenOneOf)

	if err != nil {
		return 0, err
	}

	if _, err := p.expect(token.TokenLeftBrace); err != nil {
		return 0, err
	}

	first, err := p.parseStructuredSyntaxFieldTarget()

	if err != nil {
		return 0, err
	}

	var buffer [4]ast.SyntaxExpressionID
	alternatives := buffer[:0]
	alternatives = append(alternatives, first)

	for !p.isAt(token.TokenRightBrace) {
		alternative, err := p.parseStructuredSyntaxFieldTarget()

		if err != nil {
			return 0, err
		}

		alternatives = append(alternatives, alternative)
	}

	end, err := p.expect(token.TokenRightBrace)

	if err != nil {
		return 0, err
	}

	if len(alternatives) == 1 {
		return alternatives[0], nil
	}

	return p.addSyntaxExpressionBranch(
		ast.SyntaxExpressionAlternation,
		alternatives,
		span(start.Span.Start, end.Span.End),
	), nil
}

func (p *parser) finishStructuredSyntaxReference(reference token.Token, operator token.Token) ast.SyntaxExpressionID {
	expressionID := p.addSyntaxExpressionNode(ast.SyntaxExpressionNode{
		Kind:      ast.SyntaxExpressionReference,
		Reference: reference,
		Span:      reference.Span,
	})

	if operator.Kind == token.TokenQuestion || operator.Kind == token.TokenStar || operator.Kind == token.TokenPlus {
		return p.wrapSyntaxExpressionWithRepetition(expressionID, operator)
	}

	return expressionID
}
