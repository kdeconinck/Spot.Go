// =====================================================================================================================
// == LICENSE:                 Copyright (c) 2026 Kevin De Coninck.
// == SPDX-License-Identifier: LicenseRef-PolyForm-Noncommercial-1.0.0
// =====================================================================================================================

// Package parser parses Spot DSL source text into syntax data structures.
package parser

import "github.com/kdeconinck/spot/syntax"

func (parser *parser) parseOptionalDefinitionsSection() syntax.DefinitionsSection {
	if !parser.at(syntax.TokenDefinitions) {
		return syntax.DefinitionsSection{}
	}

	return parser.parseDefinitionsSection()
}

func (parser *parser) parseDefinitionsSection() syntax.DefinitionsSection {
	start := parser.expect(syntax.TokenDefinitions)

	if !parser.match(syntax.TokenLeftBrace) {
		return syntax.DefinitionsSection{
			Span: start.Span,
		}
	}

	var definitions []syntax.Definition

	for parser.at(syntax.TokenIdentifier) {
		definitions = append(definitions, parser.parseDefinition())
	}

	end := parser.expectSectionEnd(syntax.TokenIdentifier)

	return syntax.DefinitionsSection{
		Definitions: definitions,
		Span:        span(start.Span.Start, end.Span.End),
	}
}

func (parser *parser) parseDefinition() syntax.Definition {
	name := parser.expect(syntax.TokenIdentifier)
	parser.expect(syntax.TokenEqual)
	expression := parser.parseExpression(false)

	return syntax.Definition{
		Name:       name,
		Expression: expression,
		Span:       span(name.Span.Start, expression.Span.End),
	}
}
