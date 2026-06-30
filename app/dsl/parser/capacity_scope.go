// =====================================================================================================================
// == LICENSE:                 Copyright (c) 2026 Kevin De Coninck.
// == SPDX-License-Identifier: LicenseRef-PolyForm-Noncommercial-1.0.0
// =====================================================================================================================

// Package parser parses Spot DSL source text into syntax data structures.
package parser

import "github.com/kdeconinck/spot/dsl/token"

func (p *sizingParser) measureOptionalScopeSection() {
	if !p.isAt(token.TokenScope) {
		return
	}

	p.measureScopeSection()
}

func (p *sizingParser) measureScopeSection() {
	p.expect(token.TokenScope)

	if !p.match(token.TokenLeftBrace) {
		return
	}

	for p.startsScopeEntry() {
		p.measureScopeEntryDeclaration()
	}

	p.expectSectionEnd()
}

func (p *sizingParser) measureScopeEntryDeclaration() {
	p.capacity.amountOfScopeElements++

	p.advance()
	p.expect(token.TokenString)
}

func (p *sizingParser) startsScopeEntry() bool {
	return p.isAt(token.TokenInclude) || p.isAt(token.TokenExclude)
}
