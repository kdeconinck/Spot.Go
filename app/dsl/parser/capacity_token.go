// =====================================================================================================================
// == LICENSE:                 Copyright (c) 2026 Kevin De Coninck.
// == SPDX-License-Identifier: LicenseRef-PolyForm-Noncommercial-1.0.0
// =====================================================================================================================

// Package parser parses Spot DSL source text into syntax data structures.
package parser

import "github.com/kdeconinck/spot/dsl/token"

func (p *sizingParser) measureOptionalTokensSection() {
	if !p.isAt(token.TokenTokens) {
		return
	}

	p.measureTokensSection()
}

func (p *sizingParser) measureTokensSection() {
	p.expect(token.TokenTokens)

	if !p.match(token.TokenLeftBrace) {
		return
	}

	for p.isAt(token.TokenIdentifier) {
		p.measureTokenDeclaration()
	}

	p.expectSectionEnd()
}

func (p *sizingParser) measureTokenDeclaration() {
	p.capacity.amountOfTokenElements++

	p.expect(token.TokenIdentifier)
	p.expect(token.TokenEqual)
	p.measureExpressionCapacity(true)

	if p.isAt(token.TokenSkip) {
		p.advance()
	}
}
