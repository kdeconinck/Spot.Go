// =====================================================================================================================
// == LICENSE:                 Copyright (c) 2026 Kevin De Coninck.
// == SPDX-License-Identifier: LicenseRef-PolyForm-Noncommercial-1.0.0
// =====================================================================================================================

// Package parser parses Spot DSL source text into syntax data structures.
package parser

import (
	"github.com/kdeconinck/spot/dsl/ast"
	"github.com/kdeconinck/spot/dsl/token"
)

func (p *parser) parseScopeSection() (ast.ScopeSection, error) {
	if !p.isAt(token.TokenScope) {
		return ast.ScopeSection{}, p.unexpected(token.TokenScope)
	}

	start := p.current

	p.advance()

	if err := p.match(token.TokenLeftBrace); err != nil {
		return ast.ScopeSection{}, err
	}

	firstElementIdx := uint32(len(p.document.ScopeEntries))

	for p.isAt(token.TokenInclude) || p.isAt(token.TokenExclude) {
		entry, err := p.parseScopeEntry()

		if err != nil {
			return ast.ScopeSection{}, err
		}

		p.document.ScopeEntries = append(p.document.ScopeEntries, entry)
	}

	end, err := p.expectSectionEnd(token.TokenInclude)

	if err != nil {
		return ast.ScopeSection{}, err
	}

	return ast.ScopeSection{
		FirstElementIdx:  firstElementIdx,
		AmountOfElements: uint32(len(p.document.ScopeEntries)) - firstElementIdx,
		Span:             span(start.Span.Start, end.Span.End),
	}, nil
}

func (p *parser) parseScopeEntry() (ast.ScopeEntry, error) {
	start := p.current
	kind := ast.ScopeEntryInclude

	if p.isAt(token.TokenExclude) {
		kind = ast.ScopeEntryExclude
	}

	p.advance()

	pattern, err := p.expect(token.TokenString)

	if err != nil {
		return ast.ScopeEntry{}, err
	}

	return ast.ScopeEntry{
		Kind:    kind,
		Pattern: pattern,
		Span:    span(start.Span.Start, pattern.Span.End),
	}, nil
}
