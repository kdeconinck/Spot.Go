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

func (parser *parser) parseScopeSection() ast.ScopeSection {
	if !parser.at(token.TokenScope) {
		parser.addDiagnostic(token.TokenScope)

		return ast.ScopeSection{
			Span: parser.current.Span,
		}
	}

	start := parser.expect(token.TokenScope)

	if !parser.match(token.TokenLeftBrace) {
		return ast.ScopeSection{
			Span: start.Span,
		}
	}

	var entries []ast.ScopeEntry

	for parser.at(token.TokenInclude) || parser.at(token.TokenExclude) {
		entries = append(entries, parser.parseScopeEntry())
	}

	end := parser.expectSectionEnd(token.TokenInclude)

	return ast.ScopeSection{
		Entries: entries,
		Span:    span(start.Span.Start, end.Span.End),
	}
}

func (parser *parser) parseScopeEntry() ast.ScopeEntry {
	start := parser.current
	kind := ast.ScopeEntryInclude

	if parser.at(token.TokenExclude) {
		kind = ast.ScopeEntryExclude
	}

	parser.advance()
	pattern := parser.expect(token.TokenString)

	return ast.ScopeEntry{
		Kind:    kind,
		Pattern: pattern,
		Span:    span(start.Span.Start, pattern.Span.End),
	}
}
