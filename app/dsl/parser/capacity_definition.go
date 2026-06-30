// =====================================================================================================================
// == LICENSE:                 Copyright (c) 2026 Kevin De Coninck.
// == SPDX-License-Identifier: LicenseRef-PolyForm-Noncommercial-1.0.0
// =====================================================================================================================

// Package parser parses Spot DSL source text into syntax data structures.
package parser

import "github.com/kdeconinck/spot/dsl/token"

func (p *sizingParser) measureOptionalDefinitionsSection() {
	if !p.isAt(token.TokenDefinitions) {
		return
	}

	p.measureDefinitionsSection()
}

func (p *sizingParser) measureDefinitionsSection() {
	p.expect(token.TokenDefinitions)

	if !p.match(token.TokenLeftBrace) {
		return
	}

	for p.isAt(token.TokenIdentifier) {
		p.measureDefinitionDeclaration()
	}

	p.expectSectionEnd()
}

func (p *sizingParser) measureDefinitionDeclaration() {
	p.capacity.amountOfDefinitionElements++

	p.expect(token.TokenIdentifier)
	p.expect(token.TokenEqual)
	p.measureExpressionCapacity(false)
}
