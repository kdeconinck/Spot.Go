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

func (p *parser) parseScopeSection() ast.ScopeSection {
	if !p.at(token.TokenScope) {
		p.addDiagnostic(token.TokenScope)

		return ast.ScopeSection{
			Span: p.current.Span,
		}
	}

	start := p.expect(token.TokenScope)

	if !p.match(token.TokenLeftBrace) {
		return ast.ScopeSection{
			Span: start.Span,
		}
	}

	var entries []ast.ScopeEntry

	for p.at(token.TokenInclude) || p.at(token.TokenExclude) {
		entries = append(entries, p.parseScopeEntry())
	}

	end := p.expectSectionEnd(token.TokenInclude)

	return ast.ScopeSection{
		Entries: entries,
		Span:    span(start.Span.Start, end.Span.End),
	}
}

func (p *parser) parseScopeEntry() ast.ScopeEntry {
	start := p.current
	kind := ast.ScopeEntryInclude

	if p.at(token.TokenExclude) {
		kind = ast.ScopeEntryExclude
	}

	p.advance()
	pattern := p.expect(token.TokenString)

	return ast.ScopeEntry{
		Kind:    kind,
		Pattern: pattern,
		Span:    span(start.Span.Start, pattern.Span.End),
	}
}
