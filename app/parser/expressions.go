// =====================================================================================================================
// == LICENSE:                 Copyright (c) 2026 Kevin De Coninck.
// == SPDX-License-Identifier: LicenseRef-PolyForm-Noncommercial-1.0.0
// =====================================================================================================================

// Package parser parses Spot DSL source text into syntax data structures.
package parser

import "github.com/kdeconinck/spot/syntax"

func (parser *parser) parseExpression(allowString bool) syntax.DefinitionExpression {
	first := parser.parseConcatenation(allowString)

	if !parser.at(syntax.TokenPipe) {
		return first
	}

	terms := []syntax.DefinitionExpression{
		first,
	}

	for parser.consume(syntax.TokenPipe) {
		terms = append(terms, parser.parseConcatenation(allowString))
	}

	return syntax.DefinitionExpression{
		Kind:  syntax.DefinitionExpressionAlternation,
		Terms: terms,
		Span:  span(first.Span.Start, terms[len(terms)-1].Span.End),
	}
}

func (parser *parser) parseConcatenation(allowString bool) syntax.DefinitionExpression {
	first := parser.parseRepetition(allowString)

	if !parser.atExpressionContinuationStart(allowString) {
		return first
	}

	terms := []syntax.DefinitionExpression{
		first,
	}

	for parser.atExpressionContinuationStart(allowString) {
		terms = append(terms, parser.parseRepetition(allowString))
	}

	return syntax.DefinitionExpression{
		Kind:  syntax.DefinitionExpressionConcatenation,
		Terms: terms,
		Span:  span(first.Span.Start, terms[len(terms)-1].Span.End),
	}
}

func (parser *parser) parseRepetition(allowString bool) syntax.DefinitionExpression {
	inner := parser.parsePrimary(allowString)

	if !parser.at(syntax.TokenQuestion) && !parser.at(syntax.TokenStar) && !parser.at(syntax.TokenPlus) {
		return inner
	}

	operator := parser.current
	parser.advance()

	return syntax.DefinitionExpression{
		Kind:     syntax.DefinitionExpressionRepetition,
		Operator: operator,
		Inner:    &inner,
		Span:     span(inner.Span.Start, operator.Span.End),
	}
}

func (parser *parser) parsePrimary(allowString bool) syntax.DefinitionExpression {
	if parser.at(syntax.TokenLeftParen) {
		return parser.parseGroupedExpression(allowString)
	}

	if allowString && parser.at(syntax.TokenString) {
		start := parser.expect(syntax.TokenString)

		return syntax.DefinitionExpression{
			Kind:  syntax.DefinitionExpressionString,
			Start: start,
			Span:  start.Span,
		}
	}

	if parser.at(syntax.TokenIdentifier) {
		reference := parser.expect(syntax.TokenIdentifier)

		return syntax.DefinitionExpression{
			Kind:  syntax.DefinitionExpressionReference,
			Start: reference,
			Span:  reference.Span,
		}
	}

	start := parser.expect(syntax.TokenCharacter)

	if !parser.consume(syntax.TokenDotDot) {
		return syntax.DefinitionExpression{
			Kind:  syntax.DefinitionExpressionCharacter,
			Start: start,
			Span:  start.Span,
		}
	}

	end := parser.expect(syntax.TokenCharacter)

	return syntax.DefinitionExpression{
		Kind:  syntax.DefinitionExpressionRange,
		Start: start,
		End:   end,
		Span:  span(start.Span.Start, end.Span.End),
	}
}

func (parser *parser) atExpressionContinuationStart(allowString bool) bool {
	if parser.at(syntax.TokenIdentifier) && parser.next.Kind == syntax.TokenEqual {
		return false
	}

	return parser.at(syntax.TokenLeftParen) ||
		parser.at(syntax.TokenIdentifier) ||
		parser.at(syntax.TokenCharacter) ||
		allowString && parser.at(syntax.TokenString)
}

func (parser *parser) parseGroupedExpression(allowString bool) syntax.DefinitionExpression {
	start := parser.expect(syntax.TokenLeftParen)
	inner := parser.parseExpression(allowString)
	end := parser.expect(syntax.TokenRightParen)

	return syntax.DefinitionExpression{
		Kind:  syntax.DefinitionExpressionGroup,
		Inner: &inner,
		Span:  span(start.Span.Start, end.Span.End),
	}
}
