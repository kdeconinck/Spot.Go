// =====================================================================================================================
// == LICENSE:                 Copyright (c) 2026 Kevin De Coninck.
// == SPDX-License-Identifier: LicenseRef-PolyForm-Noncommercial-1.0.0
// =====================================================================================================================

// Package parser parses Spot DSL source text into syntax data structures.
package parser

func (p *sizingParser) measureDocument() {
	p.measureOptionalScopeSection()
	p.measureOptionalDefinitionsSection()
	p.measureOptionalTokensSection()
	p.measureOptionalRulesSection()
}
