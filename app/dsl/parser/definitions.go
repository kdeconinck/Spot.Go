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

func (p *parser) parseOptionalDefinitionsSection() (ast.DefinitionsSection, error) {
	if !p.isAt(token.TokenDefinitions) {
		return ast.DefinitionsSection{}, nil
	}

	return p.parseDefinitionsSection()
}

func (p *parser) parseDefinitionsSection() (ast.DefinitionsSection, error) {
	start := p.current

	p.advance()

	if err := p.match(token.TokenLeftBrace); err != nil {
		return ast.DefinitionsSection{}, err
	}

	firstDefinition := uint32(len(p.document.DefinitionList))

	for p.isAt(token.TokenIdentifier) {
		definition, err := p.parseDefinition()

		if err != nil {
			return ast.DefinitionsSection{}, err
		}

		p.document.DefinitionList = append(p.document.DefinitionList, definition)
	}

	end, err := p.expectSectionEnd(token.TokenIdentifier)

	if err != nil {
		return ast.DefinitionsSection{}, err
	}

	return ast.DefinitionsSection{
		FirstElementsIdx: firstDefinition,
		AmountOfElements: uint32(len(p.document.DefinitionList)) - firstDefinition,
		Span:             span(start.Span.Start, end.Span.End),
	}, nil
}

func (p *parser) parseDefinition() (ast.Definition, error) {
	name := p.current

	p.advance()

	if _, err := p.expect(token.TokenEqual); err != nil {
		return ast.Definition{}, err
	}

	expressionID, err := p.parseExpression(false)

	if err != nil {
		return ast.Definition{}, err
	}

	return ast.Definition{
		Name:       name,
		Expression: expressionID,
		Span:       span(name.Span.Start, p.expressionNode(expressionID).Span.End),
	}, nil
}
