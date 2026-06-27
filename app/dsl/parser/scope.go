// =====================================================================================================================
// == LICENSE:                 Copyright (c) 2026 Kevin De Coninck.
// == SPDX-License-Identifier: LicenseRef-PolyForm-Noncommercial-1.0.0
// =====================================================================================================================

// Package parser parses Spot DSL source text into syntax data structures.
package parser

import "github.com/kdeconinck/spot/dsl/token"

func (parser *parser) parseScopeSection() token.ScopeSection {
	if !parser.at(token.TokenScope) {
		parser.addDiagnostic(token.TokenScope)

		return token.ScopeSection{
			Span: parser.current.Span,
		}
	}

	start := parser.expect(token.TokenScope)

	if !parser.match(token.TokenLeftBrace) {
		return token.ScopeSection{
			Span: start.Span,
		}
	}

	var entries []token.ScopeEntry

	for parser.at(token.TokenInclude) || parser.at(token.TokenExclude) {
		entries = append(entries, parser.parseScopeEntry())
	}

	end := parser.expectSectionEnd(token.TokenInclude)

	return token.ScopeSection{
		Entries: entries,
		Span:    span(start.Span.Start, end.Span.End),
	}
}

func (parser *parser) parseScopeEntry() token.ScopeEntry {
	start := parser.current
	kind := token.ScopeEntryInclude

	if parser.at(token.TokenExclude) {
		kind = token.ScopeEntryExclude
	}

	parser.advance()
	pattern := parser.expect(token.TokenString)

	return token.ScopeEntry{
		Kind:    kind,
		Pattern: pattern,
		Span:    span(start.Span.Start, pattern.Span.End),
	}
}
