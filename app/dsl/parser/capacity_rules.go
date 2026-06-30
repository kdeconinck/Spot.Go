// =====================================================================================================================
// == LICENSE:                 Copyright (c) 2026 Kevin De Coninck.
// == SPDX-License-Identifier: LicenseRef-PolyForm-Noncommercial-1.0.0
// =====================================================================================================================

// Package parser parses Spot DSL source text into syntax data structures.
package parser

import "github.com/kdeconinck/spot/dsl/token"

// Measures capacity for the optional rules section and its clauses.
func (p *sizingParser) measureOptionalRulesSection() {
	if !p.isAt(token.TokenRules) {
		return
	}

	p.measureRulesSection()
}

func (p *sizingParser) measureRulesSection() {
	p.expect(token.TokenRules)

	if !p.match(token.TokenLeftBrace) {
		return
	}

	for p.startsRule() {
		p.measureRuleDeclaration()
	}

	p.expectSectionEnd()
}

func (p *sizingParser) startsRule() bool {
	return p.isAt(token.TokenRule) || p.atSeverity()
}

func (p *sizingParser) measureRuleDeclaration() {
	p.capacity.amountOfRuleElements++

	if p.atSeverity() {
		p.measureSelectorRule()

		return
	}

	p.expect(token.TokenRule)
	p.expect(token.TokenIdentifier)

	if !p.match(token.TokenLeftBrace) {
		return
	}

	p.measureMatchClause()

	if p.startsWhereClause() {
		p.measureWhereClause()
	}

	p.measureReportClause()
	p.expect(token.TokenRightBrace)
}

func (p *sizingParser) measureSelectorRule() {
	p.expectSeverityToken()
	p.expect(token.TokenString)
	p.expect(token.TokenColon)
	p.measureSelectorMatch()
}

func (p *sizingParser) measureSelectorMatch() {
	p.expect(token.TokenIdentifier)

	if p.isAt(token.TokenColon) {
		p.expect(token.TokenColon)
		p.expect(token.TokenNot)
		p.expect(token.TokenLeftParen)
		p.expect(token.TokenIdentifier)

		if p.isAt(token.TokenGreater) {
			p.advance()
		}

		p.expect(token.TokenStar)
		p.expect(token.TokenRightParen)

		return
	}

	if p.isAt(token.TokenGreater) {
		p.advance()
		p.expect(token.TokenIdentifier)

		return
	}

	if p.isAt(token.TokenIdentifier) {
		p.advance()
	}
}

func (p *sizingParser) measureMatchClause() {
	p.expect(token.TokenMatch)

	if p.isAt(token.TokenNode) {
		p.advance()
	}

	p.expect(token.TokenIdentifier)

	if p.isAt(token.TokenInside) || p.isAt(token.TokenOutside) {
		p.advance()
		p.expect(token.TokenIdentifier)
	}
}

func (p *sizingParser) startsWhereClause() bool {
	return p.isAt(token.TokenWhere)
}

func (p *sizingParser) measureWhereClause() {
	p.expect(token.TokenWhere)
	p.expect(token.TokenIdentifier)
	p.expect(token.TokenDot)
	p.expect(token.TokenIdentifier)
	p.expectComparisonOperatorToken()
	p.expectConditionLiteralToken()
}

func (p *sizingParser) expectComparisonOperatorToken() {
	if p.atComparisonOperator() {
		p.advance()

		return
	}

	p.advanceUnexpected()
}

func (p *sizingParser) atComparisonOperator() bool {
	return p.isAt(token.TokenEqualEqual) || p.isAt(token.TokenBangEqual) || p.isAt(token.TokenLess) || p.isAt(token.TokenLessEqual) || p.isAt(token.TokenGreater) || p.isAt(token.TokenGreaterEqual)
}

func (p *sizingParser) atConditionLiteral() bool {
	return p.isAt(token.TokenString) || p.isAt(token.TokenInteger)
}

func (p *sizingParser) expectConditionLiteralToken() {
	if p.atConditionLiteral() {
		p.advance()

		return
	}

	p.advanceUnexpected()
}

func (p *sizingParser) measureReportClause() {
	p.expect(token.TokenReport)
	p.expectSeverityToken()
	p.expect(token.TokenAt)
	p.expect(token.TokenIdentifier)
	p.expect(token.TokenString)
}

func (p *sizingParser) expectSeverityToken() {
	if p.atSeverity() {
		p.advance()

		return
	}

	p.advanceUnexpected()
}

func (p *sizingParser) atSeverity() bool {
	return p.isAt(token.TokenInfo) || p.isAt(token.TokenWarn) || p.isAt(token.TokenErr)
}
