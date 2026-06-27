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

func (p *parser) parseExpression(allowString bool) ast.DefinitionExpression {
	first := p.parseConcatenation(allowString)

	if !p.at(token.TokenPipe) {
		return first
	}

	terms := []ast.DefinitionExpression{
		first,
	}

	for p.consume(token.TokenPipe) {
		terms = append(terms, p.parseConcatenation(allowString))
	}

	return ast.DefinitionExpression{
		Kind:  ast.DefinitionExpressionAlternation,
		Terms: terms,
		Span:  span(first.Span.Start, terms[len(terms)-1].Span.End),
	}
}

func (p *parser) parseConcatenation(allowString bool) ast.DefinitionExpression {
	first := p.parseRepetition(allowString)

	if !p.atExpressionContinuationStart(allowString) {
		return first
	}

	terms := []ast.DefinitionExpression{
		first,
	}

	for p.atExpressionContinuationStart(allowString) {
		terms = append(terms, p.parseRepetition(allowString))
	}

	return ast.DefinitionExpression{
		Kind:  ast.DefinitionExpressionConcatenation,
		Terms: terms,
		Span:  span(first.Span.Start, terms[len(terms)-1].Span.End),
	}
}

func (p *parser) parseRepetition(allowString bool) ast.DefinitionExpression {
	inner := p.parsePrimary(allowString)

	if !p.at(token.TokenQuestion) && !p.at(token.TokenStar) && !p.at(token.TokenPlus) {
		return inner
	}

	operator := p.current
	p.advance()

	return ast.DefinitionExpression{
		Kind:     ast.DefinitionExpressionRepetition,
		Operator: operator,
		Inner:    &inner,
		Span:     span(inner.Span.Start, operator.Span.End),
	}
}

func (p *parser) parsePrimary(allowString bool) ast.DefinitionExpression {
	if p.at(token.TokenLeftParen) {
		return p.parseGroupedExpression(allowString)
	}

	if allowString && p.at(token.TokenString) {
		start := p.expect(token.TokenString)

		return ast.DefinitionExpression{
			Kind:  ast.DefinitionExpressionString,
			Start: start,
			Span:  start.Span,
		}
	}

	if p.at(token.TokenIdentifier) {
		reference := p.expect(token.TokenIdentifier)

		return ast.DefinitionExpression{
			Kind:  ast.DefinitionExpressionReference,
			Start: reference,
			Span:  reference.Span,
		}
	}

	start := p.expect(token.TokenCharacter)

	if !p.consume(token.TokenDotDot) {
		return ast.DefinitionExpression{
			Kind:  ast.DefinitionExpressionCharacter,
			Start: start,
			Span:  start.Span,
		}
	}

	end := p.expect(token.TokenCharacter)

	return ast.DefinitionExpression{
		Kind:  ast.DefinitionExpressionRange,
		Start: start,
		End:   end,
		Span:  span(start.Span.Start, end.Span.End),
	}
}

func (p *parser) atExpressionContinuationStart(allowString bool) bool {
	if p.at(token.TokenIdentifier) && p.next.Kind == token.TokenEqual {
		return false
	}

	return p.at(token.TokenLeftParen) ||
		p.at(token.TokenIdentifier) ||
		p.at(token.TokenCharacter) ||
		allowString && p.at(token.TokenString)
}

func (p *parser) parseGroupedExpression(allowString bool) ast.DefinitionExpression {
	start := p.expect(token.TokenLeftParen)
	inner := p.parseExpression(allowString)
	end := p.expect(token.TokenRightParen)

	return ast.DefinitionExpression{
		Kind:  ast.DefinitionExpressionGroup,
		Inner: &inner,
		Span:  span(start.Span.Start, end.Span.End),
	}
}
