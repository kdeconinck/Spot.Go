// =====================================================================================================================
// == LICENSE:                 Copyright (c) 2026 Kevin De Coninck.
// == SPDX-License-Identifier: LicenseRef-PolyForm-Noncommercial-1.0.0
// =====================================================================================================================

// Package parser parses Spot DSL source text into syntax data structures.
package parser

import "github.com/kdeconinck/spot/dsl/token"

// Measures capacity for the optional syntax section and its node declarations.
func (p *sizingParser) measureOptionalSyntaxSection() {
	if !p.isAt(token.TokenSyntax) {
		return
	}

	p.expect(token.TokenSyntax)

	if !p.match(token.TokenLeftBrace) {
		return
	}

	for p.isAt(token.TokenNode) {
		p.measureSyntaxNode()
	}

	p.expectSectionEnd()
}

func (p *sizingParser) measureSyntaxNode() {
	p.capacity.amountOfSyntaxElements++
	p.expect(token.TokenNode)
	p.expect(token.TokenIdentifier)
	p.measureStructuredSyntaxNodeExpression()
}

func (p *sizingParser) measureStructuredSyntaxNodeExpression() {
	p.expect(token.TokenLeftBrace)
	p.measureStructuredSyntaxEntry()

	if p.isAt(token.TokenRightBrace) {
		p.expect(token.TokenRightBrace)

		return
	}

	childCount := 1

	for !p.isAt(token.TokenRightBrace) && !p.isAt(token.TokenEOF) {
		p.measureStructuredSyntaxEntry()
		childCount++
	}

	p.expect(token.TokenRightBrace)
	p.countMeasuredSyntaxExpressionNode(childCount)
}

func (p *sizingParser) measureStructuredSyntaxEntry() {
	if p.isAt(token.TokenOneOf) {
		p.measureStructuredSyntaxOneOf()

		return
	}

	if !p.isAt(token.TokenIdentifier) {
		p.measureStructuredSyntaxUnnamedEntry()

		return
	}

	p.expect(token.TokenIdentifier)
	hasOperator := p.atRepetitionOperator()

	if hasOperator {
		p.advance()
	}

	if p.consume(token.TokenColon) {
		p.measureStructuredSyntaxFieldTarget()

		if hasOperator {
			p.countMeasuredSyntaxExpressionNode(1)
		}

		p.countMeasuredSyntaxExpressionNode(1)

		return
	}

	p.countMeasuredSyntaxExpressionNode(0)

	if hasOperator {
		p.countMeasuredSyntaxExpressionNode(1)
	}
}

func (p *sizingParser) measureStructuredSyntaxUnnamedEntry() {
	if p.isAt(token.TokenOneOf) {
		p.measureStructuredSyntaxOneOf()

		return
	}

	if p.isAt(token.TokenLeftParen) {
		p.measureGroupedSyntaxExpressionCapacity()
	} else if p.isAt(token.TokenAny) {
		p.advance()
		p.countMeasuredSyntaxExpressionNode(0)
	} else {
		p.expect(token.TokenIdentifier)
		p.countMeasuredSyntaxExpressionNode(0)
	}

	if p.atRepetitionOperator() {
		p.advance()
		p.countMeasuredSyntaxExpressionNode(1)
	}
}

func (p *sizingParser) measureStructuredSyntaxFieldTarget() {
	if p.isAt(token.TokenOneOf) {
		p.measureStructuredSyntaxOneOf()

		return
	}

	p.measureStructuredSyntaxUnnamedEntry()
}

func (p *sizingParser) measureStructuredSyntaxOneOf() {
	p.expect(token.TokenOneOf)
	p.expect(token.TokenLeftBrace)
	p.measureStructuredSyntaxFieldTarget()

	if p.isAt(token.TokenRightBrace) {
		p.expect(token.TokenRightBrace)

		return
	}

	childCount := 1

	for !p.isAt(token.TokenRightBrace) && !p.isAt(token.TokenEOF) {
		p.measureStructuredSyntaxFieldTarget()
		childCount++
	}

	p.expect(token.TokenRightBrace)
	p.countMeasuredSyntaxExpressionNode(childCount)
}

func (p *sizingParser) measureSyntaxExpressionCapacity() {
	p.measureSyntaxConcatenationExpression()

	if !p.isAt(token.TokenPipe) {
		return
	}

	childCount := 1

	for p.consume(token.TokenPipe) {
		p.measureSyntaxConcatenationExpression()

		childCount++
	}

	p.countMeasuredSyntaxExpressionNode(childCount)
}

func (p *sizingParser) measureSyntaxConcatenationExpression() {
	childCount := 1

	p.measureSyntaxRepetitionExpression()

	if !p.startsSyntaxExpressionContinuation() {
		return
	}

	for p.startsSyntaxExpressionContinuation() {
		p.measureSyntaxRepetitionExpression()

		childCount++
	}

	p.countMeasuredSyntaxExpressionNode(childCount)
}

func (p *sizingParser) measureSyntaxRepetitionExpression() {
	p.measureSyntaxPrimaryExpression()

	if !p.atRepetitionOperator() {
		return
	}

	p.advance()
	p.countMeasuredSyntaxExpressionNode(1)
}

func (p *sizingParser) measureSyntaxPrimaryExpression() {
	if p.isAt(token.TokenIdentifier) && p.next.Kind == token.TokenColon {
		p.expect(token.TokenIdentifier)
		p.expect(token.TokenColon)
		p.measureSyntaxPrimaryExpression()
		p.countMeasuredSyntaxExpressionNode(1)

		return
	}

	if p.isAt(token.TokenLeftParen) {
		p.measureGroupedSyntaxExpressionCapacity()

		return
	}

	if p.isAt(token.TokenAny) {
		p.advance()
		p.countMeasuredSyntaxExpressionNode(0)

		return
	}

	p.expect(token.TokenIdentifier)
	p.countMeasuredSyntaxExpressionNode(0)
}

func (p *sizingParser) startsSyntaxExpressionContinuation() bool {
	if p.isAt(token.TokenIdentifier) && p.next.Kind == token.TokenEqual {
		return false
	}

	return p.isAt(token.TokenLeftParen) || p.isAt(token.TokenIdentifier) || p.isAt(token.TokenAny)
}

func (p *sizingParser) measureGroupedSyntaxExpressionCapacity() {
	p.expect(token.TokenLeftParen)
	p.measureSyntaxExpressionCapacity()
	p.expect(token.TokenRightParen)
	p.countMeasuredSyntaxExpressionNode(1)
}

func (p *sizingParser) countMeasuredSyntaxExpressionNode(childCount int) {
	p.capacity.amountOfSyntaxExpressionNodes++
	p.capacity.amountOfSyntaxExpressionChildren += childCount
}
