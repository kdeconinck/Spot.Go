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

func (p *parser) parseDocument() (ast.Document, error) {
	scope, err := p.parseScopeSection()

	if err != nil {
		return ast.Document{}, err
	}

	definitions, err := p.parseOptionalDefinitionsSection()

	if err != nil {
		return ast.Document{}, err
	}

	tokens, err := p.parseOptionalTokensSection()

	if err != nil {
		return ast.Document{}, err
	}

	rules, err := p.parseOptionalRulesSection()

	if err != nil {
		return ast.Document{}, err
	}

	end, err := p.expect(token.TokenEOF)

	if err != nil {
		return ast.Document{}, err
	}

	p.document.Scope = scope
	p.document.Definitions = definitions
	p.document.Tokens = tokens
	p.document.Rules = rules
	p.document.Span = span(scope.Span.Start, end.Span.End)

	return p.document, nil
}
