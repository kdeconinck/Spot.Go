// =====================================================================================================================
// == LICENSE:                 Copyright (c) 2026 Kevin De Coninck.
// == SPDX-License-Identifier: LicenseRef-PolyForm-Noncommercial-1.0.0
// =====================================================================================================================

// Package parser parses Spot DSL source text into syntax data structures.
package parser

import "github.com/kdeconinck/spot/syntax"

func (parser *parser) parseDocument() syntax.Document {
	diagnosticCount := len(parser.diagnostics)
	scope := parser.parseScopeSection()

	if len(parser.diagnostics) != diagnosticCount {
		return syntax.Document{
			Scope: scope,
			Span:  scope.Span,
		}
	}

	diagnosticCount = len(parser.diagnostics)
	definitions := parser.parseOptionalDefinitionsSection()

	if len(parser.diagnostics) != diagnosticCount {
		return syntax.Document{
			Scope:       scope,
			Definitions: definitions,
			Span:        span(scope.Span.Start, definitions.Span.End),
		}
	}

	diagnosticCount = len(parser.diagnostics)
	tokens := parser.parseOptionalTokensSection()

	if len(parser.diagnostics) != diagnosticCount {
		return syntax.Document{
			Scope:       scope,
			Definitions: definitions,
			Tokens:      tokens,
			Span:        span(scope.Span.Start, tokens.Span.End),
		}
	}

	diagnosticCount = len(parser.diagnostics)
	rules := parser.parseOptionalRulesSection()

	if len(parser.diagnostics) != diagnosticCount {
		return syntax.Document{
			Scope:       scope,
			Definitions: definitions,
			Tokens:      tokens,
			Rules:       rules,
			Span:        span(scope.Span.Start, rules.Span.End),
		}
	}

	end := parser.expect(syntax.TokenEOF)

	return syntax.Document{
		Scope:       scope,
		Definitions: definitions,
		Tokens:      tokens,
		Rules:       rules,
		Span:        span(scope.Span.Start, end.Span.End),
	}
}
