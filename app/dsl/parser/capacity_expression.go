// =====================================================================================================================
// == LICENSE:                 Copyright (c) 2026 Kevin De Coninck.
// == SPDX-License-Identifier: LicenseRef-PolyForm-Noncommercial-1.0.0
// =====================================================================================================================

// Package parser parses Spot DSL source text into syntax data structures.
package parser

import "github.com/kdeconinck/spot/dsl/token"

func (p *sizingParser) measureExpressionCapacity(allowString bool) {
	p.measureConcatenationExpression(allowString)

	if !p.isAt(token.TokenPipe) {
		return
	}

	childCount := 1

	for p.consume(token.TokenPipe) {
		p.measureConcatenationExpression(allowString)

		childCount++
	}

	p.countMeasuredExpressionNode(childCount)
}

func (p *sizingParser) measureConcatenationExpression(allowString bool) {
	childCount := 1

	p.measureRepetitionExpression(allowString)

	if !p.startsExpressionContinuation(allowString) {
		return
	}

	for p.startsExpressionContinuation(allowString) {
		p.measureRepetitionExpression(allowString)

		childCount++
	}

	p.countMeasuredExpressionNode(childCount)
}

func (p *sizingParser) measureRepetitionExpression(allowString bool) {
	p.measurePrimaryExpression(allowString)

	if !p.atRepetitionOperator() {
		return
	}

	p.advance()
	p.countMeasuredExpressionNode(1)
}

func (p *sizingParser) measurePrimaryExpression(allowString bool) {
	if p.isAt(token.TokenLeftParen) {
		p.measureGroupedExpressionCapacity(allowString)

		return
	}

	if allowString && p.isAt(token.TokenString) {
		p.advance()
		p.countMeasuredExpressionNode(0)

		return
	}

	if p.isAt(token.TokenIdentifier) {
		p.advance()
		p.countMeasuredExpressionNode(0)

		return
	}

	p.expect(token.TokenCharacter)

	if !p.consume(token.TokenDotDot) {
		p.countMeasuredExpressionNode(0)

		return
	}

	p.expect(token.TokenCharacter)
	p.countMeasuredExpressionNode(0)
}

func (p *sizingParser) startsExpressionContinuation(allowString bool) bool {
	if p.isAt(token.TokenIdentifier) && p.next.Kind == token.TokenEqual {
		return false
	}

	return p.isAt(token.TokenLeftParen) || p.isAt(token.TokenIdentifier) || p.isAt(token.TokenCharacter) || allowString && p.isAt(token.TokenString)
}

func (p *sizingParser) measureGroupedExpressionCapacity(allowString bool) {
	p.expect(token.TokenLeftParen)
	p.measureExpressionCapacity(allowString)
	p.expect(token.TokenRightParen)
	p.countMeasuredExpressionNode(1)
}

func (p *sizingParser) atRepetitionOperator() bool {
	return p.isAt(token.TokenQuestion) || p.isAt(token.TokenStar) || p.isAt(token.TokenPlus)
}

func (p *sizingParser) countMeasuredExpressionNode(childCount int) {
	p.capacity.amountOfExpressionNodes++
	p.capacity.amountOfExpressionChildReferences += childCount
}
