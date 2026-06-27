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

func (parser *parser) parseOptionalDefinitionsSection() ast.DefinitionsSection {
	if !parser.at(token.TokenDefinitions) {
		return ast.DefinitionsSection{}
	}

	return parser.parseDefinitionsSection()
}

func (parser *parser) parseDefinitionsSection() ast.DefinitionsSection {
	start := parser.expect(token.TokenDefinitions)

	if !parser.match(token.TokenLeftBrace) {
		return ast.DefinitionsSection{
			Span: start.Span,
		}
	}

	var definitions []ast.Definition

	for parser.at(token.TokenIdentifier) {
		definitions = append(definitions, parser.parseDefinition())
	}

	end := parser.expectSectionEnd(token.TokenIdentifier)

	return ast.DefinitionsSection{
		Definitions: definitions,
		Span:        span(start.Span.Start, end.Span.End),
	}
}

func (parser *parser) parseDefinition() ast.Definition {
	name := parser.expect(token.TokenIdentifier)
	parser.expect(token.TokenEqual)
	expression := parser.parseExpression(false)

	return ast.Definition{
		Name:       name,
		Expression: expression,
		Span:       span(name.Span.Start, expression.Span.End),
	}
}
