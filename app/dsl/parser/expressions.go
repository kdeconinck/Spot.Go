// =====================================================================================================================
// == LICENSE:                 Copyright (c) 2026 Kevin De Coninck.
// == SPDX-License-Identifier: LicenseRef-PolyForm-Noncommercial-1.0.0
// =====================================================================================================================

// Package parser parses Spot DSL source text into syntax data structures.
package parser

import "github.com/kdeconinck/spot/dsl/token"

func (parser *parser) parseExpression(allowString bool) token.DefinitionExpression {
	first := parser.parseConcatenation(allowString)

	if !parser.at(token.TokenPipe) {
		return first
	}

	terms := []token.DefinitionExpression{
		first,
	}

	for parser.consume(token.TokenPipe) {
		terms = append(terms, parser.parseConcatenation(allowString))
	}

	return token.DefinitionExpression{
		Kind:  token.DefinitionExpressionAlternation,
		Terms: terms,
		Span:  span(first.Span.Start, terms[len(terms)-1].Span.End),
	}
}

func (parser *parser) parseConcatenation(allowString bool) token.DefinitionExpression {
	first := parser.parseRepetition(allowString)

	if !parser.atExpressionContinuationStart(allowString) {
		return first
	}

	terms := []token.DefinitionExpression{
		first,
	}

	for parser.atExpressionContinuationStart(allowString) {
		terms = append(terms, parser.parseRepetition(allowString))
	}

	return token.DefinitionExpression{
		Kind:  token.DefinitionExpressionConcatenation,
		Terms: terms,
		Span:  span(first.Span.Start, terms[len(terms)-1].Span.End),
	}
}

func (parser *parser) parseRepetition(allowString bool) token.DefinitionExpression {
	inner := parser.parsePrimary(allowString)

	if !parser.at(token.TokenQuestion) && !parser.at(token.TokenStar) && !parser.at(token.TokenPlus) {
		return inner
	}

	operator := parser.current
	parser.advance()

	return token.DefinitionExpression{
		Kind:     token.DefinitionExpressionRepetition,
		Operator: operator,
		Inner:    &inner,
		Span:     span(inner.Span.Start, operator.Span.End),
	}
}

func (parser *parser) parsePrimary(allowString bool) token.DefinitionExpression {
	if parser.at(token.TokenLeftParen) {
		return parser.parseGroupedExpression(allowString)
	}

	if allowString && parser.at(token.TokenString) {
		start := parser.expect(token.TokenString)

		return token.DefinitionExpression{
			Kind:  token.DefinitionExpressionString,
			Start: start,
			Span:  start.Span,
		}
	}

	if parser.at(token.TokenIdentifier) {
		reference := parser.expect(token.TokenIdentifier)

		return token.DefinitionExpression{
			Kind:  token.DefinitionExpressionReference,
			Start: reference,
			Span:  reference.Span,
		}
	}

	start := parser.expect(token.TokenCharacter)

	if !parser.consume(token.TokenDotDot) {
		return token.DefinitionExpression{
			Kind:  token.DefinitionExpressionCharacter,
			Start: start,
			Span:  start.Span,
		}
	}

	end := parser.expect(token.TokenCharacter)

	return token.DefinitionExpression{
		Kind:  token.DefinitionExpressionRange,
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

func (parser *parser) parseGroupedExpression(allowString bool) token.DefinitionExpression {
	start := parser.expect(token.TokenLeftParen)
	inner := parser.parseExpression(allowString)
	end := parser.expect(token.TokenRightParen)

	return token.DefinitionExpression{
		Kind:  token.DefinitionExpressionGroup,
		Inner: &inner,
		Span:  span(start.Span.Start, end.Span.End),
	}
}
