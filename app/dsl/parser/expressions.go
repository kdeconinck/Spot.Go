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

func (parser *parser) parseExpression(allowString bool) ast.DefinitionExpression {
	first := parser.parseConcatenation(allowString)

	if !parser.at(token.TokenPipe) {
		return first
	}

	terms := []ast.DefinitionExpression{
		first,
	}

	for parser.consume(token.TokenPipe) {
		terms = append(terms, parser.parseConcatenation(allowString))
	}

	return ast.DefinitionExpression{
		Kind:  ast.DefinitionExpressionAlternation,
		Terms: terms,
		Span:  span(first.Span.Start, terms[len(terms)-1].Span.End),
	}
}

func (parser *parser) parseConcatenation(allowString bool) ast.DefinitionExpression {
	first := parser.parseRepetition(allowString)

	if !parser.atExpressionContinuationStart(allowString) {
		return first
	}

	terms := []ast.DefinitionExpression{
		first,
	}

	for parser.atExpressionContinuationStart(allowString) {
		terms = append(terms, parser.parseRepetition(allowString))
	}

	return ast.DefinitionExpression{
		Kind:  ast.DefinitionExpressionConcatenation,
		Terms: terms,
		Span:  span(first.Span.Start, terms[len(terms)-1].Span.End),
	}
}

func (parser *parser) parseRepetition(allowString bool) ast.DefinitionExpression {
	inner := parser.parsePrimary(allowString)

	if !parser.at(token.TokenQuestion) && !parser.at(token.TokenStar) && !parser.at(token.TokenPlus) {
		return inner
	}

	operator := parser.current
	parser.advance()

	return ast.DefinitionExpression{
		Kind:     ast.DefinitionExpressionRepetition,
		Operator: operator,
		Inner:    &inner,
		Span:     span(inner.Span.Start, operator.Span.End),
	}
}

func (parser *parser) parsePrimary(allowString bool) ast.DefinitionExpression {
	if parser.at(token.TokenLeftParen) {
		return parser.parseGroupedExpression(allowString)
	}

	if allowString && parser.at(token.TokenString) {
		start := parser.expect(token.TokenString)

		return ast.DefinitionExpression{
			Kind:  ast.DefinitionExpressionString,
			Start: start,
			Span:  start.Span,
		}
	}

	if parser.at(token.TokenIdentifier) {
		reference := parser.expect(token.TokenIdentifier)

		return ast.DefinitionExpression{
			Kind:  ast.DefinitionExpressionReference,
			Start: reference,
			Span:  reference.Span,
		}
	}

	start := parser.expect(token.TokenCharacter)

	if !parser.consume(token.TokenDotDot) {
		return ast.DefinitionExpression{
			Kind:  ast.DefinitionExpressionCharacter,
			Start: start,
			Span:  start.Span,
		}
	}

	end := parser.expect(token.TokenCharacter)

	return ast.DefinitionExpression{
		Kind:  ast.DefinitionExpressionRange,
		Start: start,
		End:   end,
		Span:  span(start.Span.Start, end.Span.End),
	}
}

func (parser *parser) atExpressionContinuationStart(allowString bool) bool {
	if parser.at(token.TokenIdentifier) && parser.next.Kind == token.TokenEqual {
		return false
	}

	return parser.at(token.TokenLeftParen) ||
		parser.at(token.TokenIdentifier) ||
		parser.at(token.TokenCharacter) ||
		allowString && parser.at(token.TokenString)
}

func (parser *parser) parseGroupedExpression(allowString bool) ast.DefinitionExpression {
	start := parser.expect(token.TokenLeftParen)
	inner := parser.parseExpression(allowString)
	end := parser.expect(token.TokenRightParen)

	return ast.DefinitionExpression{
		Kind:  ast.DefinitionExpressionGroup,
		Inner: &inner,
		Span:  span(start.Span.Start, end.Span.End),
	}
}
