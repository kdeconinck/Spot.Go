// =====================================================================================================================
// == LICENSE:                 Copyright (c) 2026 Kevin De Coninck.
// == SPDX-License-Identifier: LicenseRef-PolyForm-Noncommercial-1.0.0
// =====================================================================================================================

// Package parser parses Spot DSL source text into syntax data structures.
package parser

import "github.com/kdeconinck/spot/syntax"

func (parser *parser) parseScopeSection() syntax.ScopeSection {
	if !parser.at(syntax.TokenScope) {
		parser.addDiagnostic(syntax.TokenScope)

		return syntax.ScopeSection{
			Span: parser.current.Span,
		}
	}

	start := parser.expect(syntax.TokenScope)

	if !parser.match(syntax.TokenLeftBrace) {
		return syntax.ScopeSection{
			Span: start.Span,
		}
	}

	var entries []syntax.ScopeEntry

	for parser.at(syntax.TokenInclude) || parser.at(syntax.TokenExclude) {
		entries = append(entries, parser.parseScopeEntry())
	}

	end := parser.expectSectionEnd(syntax.TokenInclude)

	return syntax.ScopeSection{
		Entries: entries,
		Span:    span(start.Span.Start, end.Span.End),
	}
}

func (parser *parser) parseScopeEntry() syntax.ScopeEntry {
	start := parser.current
	kind := syntax.ScopeEntryInclude

	if parser.at(syntax.TokenExclude) {
		kind = syntax.ScopeEntryExclude
	}

	parser.advance()
	pattern := parser.expect(syntax.TokenString)

	return syntax.ScopeEntry{
		Kind:    kind,
		Pattern: pattern,
		Span:    span(start.Span.Start, pattern.Span.End),
	}
}
