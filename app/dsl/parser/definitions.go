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

func (p *parser) parseOptionalDefinitionsSection() ast.DefinitionsSection {
	if !p.at(token.TokenDefinitions) {
		return ast.DefinitionsSection{}
	}

	return p.parseDefinitionsSection()
}

func (p *parser) parseDefinitionsSection() ast.DefinitionsSection {
	start := p.expect(token.TokenDefinitions)

	if !p.match(token.TokenLeftBrace) {
		return ast.DefinitionsSection{
			Span: start.Span,
		}
	}

	var definitions []ast.Definition

	for p.at(token.TokenIdentifier) {
		definitions = append(definitions, p.parseDefinition())
	}

	end := p.expectSectionEnd(token.TokenIdentifier)

	return ast.DefinitionsSection{
		Definitions: definitions,
		Span:        span(start.Span.Start, end.Span.End),
	}
}

func (p *parser) parseDefinition() ast.Definition {
	name := p.expect(token.TokenIdentifier)
	p.expect(token.TokenEqual)
	expression := p.parseExpression(false)

	return ast.Definition{
		Name:       name,
		Expression: expression,
		Span:       span(name.Span.Start, expression.Span.End),
	}
}
