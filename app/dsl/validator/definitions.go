// =====================================================================================================================
// == LICENSE:                 Copyright (c) 2026 Kevin De Coninck.
// == SPDX-License-Identifier: LicenseRef-PolyForm-Noncommercial-1.0.0
// =====================================================================================================================

// Package validator validates parsed Spot DSL syntax.
package validator

import (
	"github.com/kdeconinck/spot/dsl/ast"
	"github.com/kdeconinck/spot/dsl/token"
)

func validateDefinitions(source string, definitions ast.DefinitionsSection, diagnostics []Diagnostic) []Diagnostic {
	names := map[string]struct{}{}

	for idx := range definitions.Definitions {
		name := definitions.Definitions[idx].Name

		if _, ok := names[name.Value(source)]; ok {
			diagnostics = append(diagnostics, Diagnostic{
				Message: `Definition "` + name.Value(source) + `" is already declared.`,
				Span:    name.Span,
			})

			continue
		}

		names[name.Value(source)] = struct{}{}
	}

	for idx := range definitions.Definitions {
		diagnostics = validateDefinitionExpression(source, definitions.Definitions[idx].Expression, names, diagnostics)
	}

	diagnostics = validateDefinitionRecursion(source, definitions, diagnostics)

	return diagnostics
}

func validateDefinitionExpression(source string, expression ast.DefinitionExpression, names map[string]struct{}, diagnostics []Diagnostic) []Diagnostic {
	switch expression.Kind {
	case ast.DefinitionExpressionRange:
		if characterValue(source, expression.Start) > characterValue(source, expression.End) {
			diagnostics = append(diagnostics, Diagnostic{
				Message: "Character range start must be less than or equal to end.",
				Span:    expression.Span,
			})
		}

	case ast.DefinitionExpressionReference:
		if _, ok := names[expression.Start.Value(source)]; !ok {
			diagnostics = append(diagnostics, Diagnostic{
				Message: `Definition "` + expression.Start.Value(source) + `" is not declared.`,
				Span:    expression.Start.Span,
			})
		}

	case ast.DefinitionExpressionAlternation, ast.DefinitionExpressionConcatenation:
		for idx := range expression.Terms {
			diagnostics = validateDefinitionExpression(source, expression.Terms[idx], names, diagnostics)
		}

	case ast.DefinitionExpressionGroup, ast.DefinitionExpressionRepetition:
		diagnostics = validateDefinitionExpression(source, *expression.Inner, names, diagnostics)
	}

	return diagnostics
}

func characterValue(source string, token token.Token) byte {
	if token.Value(source)[1] != '\\' {
		return token.Value(source)[1]
	}

	switch token.Value(source)[2] {
	case '\\':
		return '\\'

	case '\'':
		return '\''

	case 'n':
		return '\n'

	case 'r':
		return '\r'
	}

	return '\t'
}

func validateDefinitionRecursion(source string, definitions ast.DefinitionsSection, diagnostics []Diagnostic) []Diagnostic {
	if len(definitions.Definitions) == 0 {
		return diagnostics
	}

	definitionByName := map[string]ast.Definition{}

	for idx := range definitions.Definitions {
		definition := definitions.Definitions[idx]

		if _, ok := definitionByName[definition.Name.Value(source)]; !ok {
			definitionByName[definition.Name.Value(source)] = definition
		}
	}

	states := map[string]uint8{}

	for idx := range definitions.Definitions {
		diagnostics = validateDefinitionCycle(source, definitions.Definitions[idx].Name.Value(source), definitionByName, states, diagnostics)
	}

	return diagnostics
}

func validateDefinitionCycle(source string, name string, definitions map[string]ast.Definition, states map[string]uint8, diagnostics []Diagnostic) []Diagnostic {
	switch states[name] {
	case 1, 2:
		return diagnostics
	}

	definition, ok := definitions[name]
	if !ok {
		return diagnostics
	}

	states[name] = 1
	diagnostics = validateDefinitionExpressionCycle(source, definition.Expression, definitions, states, diagnostics)
	states[name] = 2

	return diagnostics
}

func validateDefinitionExpressionCycle(source string, expression ast.DefinitionExpression, definitions map[string]ast.Definition, states map[string]uint8, diagnostics []Diagnostic) []Diagnostic {
	switch expression.Kind {
	case ast.DefinitionExpressionReference:
		if states[expression.Start.Value(source)] == 1 {
			diagnostics = append(diagnostics, Diagnostic{
				Message: `Definition "` + expression.Start.Value(source) + `" is recursive.`,
				Span:    expression.Start.Span,
			})

			return diagnostics
		}

		diagnostics = validateDefinitionCycle(source, expression.Start.Value(source), definitions, states, diagnostics)

	case ast.DefinitionExpressionAlternation, ast.DefinitionExpressionConcatenation:
		for idx := range expression.Terms {
			diagnostics = validateDefinitionExpressionCycle(source, expression.Terms[idx], definitions, states, diagnostics)
		}

	case ast.DefinitionExpressionGroup, ast.DefinitionExpressionRepetition:
		diagnostics = validateDefinitionExpressionCycle(source, *expression.Inner, definitions, states, diagnostics)
	}

	return diagnostics
}
