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
	p.expect(token.TokenEqual)
	p.measureSyntaxExpressionCapacity()
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
