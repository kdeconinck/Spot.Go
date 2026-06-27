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

func (p *parser) parseDocument() ast.Document {
	diagnosticCount := len(p.diagnostics)
	scope := p.parseScopeSection()

	if len(p.diagnostics) != diagnosticCount {
		return ast.Document{
			Scope: scope,
			Span:  scope.Span,
		}
	}

	diagnosticCount = len(p.diagnostics)
	definitions := p.parseOptionalDefinitionsSection()

	if len(p.diagnostics) != diagnosticCount {
		return ast.Document{
			Scope:       scope,
			Definitions: definitions,
			Span:        span(scope.Span.Start, definitions.Span.End),
		}
	}

	diagnosticCount = len(p.diagnostics)
	tokens := p.parseOptionalTokensSection()

	if len(p.diagnostics) != diagnosticCount {
		return ast.Document{
			Scope:       scope,
			Definitions: definitions,
			Tokens:      tokens,
			Span:        span(scope.Span.Start, tokens.Span.End),
		}
	}

	diagnosticCount = len(p.diagnostics)
	rules := p.parseOptionalRulesSection()

	if len(p.diagnostics) != diagnosticCount {
		return ast.Document{
			Scope:       scope,
			Definitions: definitions,
			Tokens:      tokens,
			Rules:       rules,
			Span:        span(scope.Span.Start, rules.Span.End),
		}
	}

	end := p.expect(token.TokenEOF)

	return ast.Document{
		Scope:       scope,
		Definitions: definitions,
		Tokens:      tokens,
		Rules:       rules,
		Span:        span(scope.Span.Start, end.Span.End),
	}
}
