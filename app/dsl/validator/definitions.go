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

func validateDefinitions(source string, definitions ast.DefinitionsSection, definitionList []ast.Definition, expressions ast.DefinitionExpressionArena, diagnostics []Diagnostic) []Diagnostic {
	names := map[string]struct{}{}

	for idx := range definitionList {
		name := definitionList[idx].Name

		if _, ok := names[name.Value(source)]; ok {
			diagnostics = append(diagnostics, Diagnostic{
				Message: `Definition "` + name.Value(source) + `" is already declared.`,
				Span:    name.Span,
			})

			continue
		}

		names[name.Value(source)] = struct{}{}
	}

	for idx := range definitionList {
		diagnostics = validateDefinitionExpression(source, definitionList[idx].Expression, expressions, names, diagnostics)
	}

	diagnostics = validateDefinitionRecursion(source, definitionList, expressions, diagnostics)

	return diagnostics
}

func validateDefinitionExpression(source string, expressionID ast.DefinitionExpressionID, expressions ast.DefinitionExpressionArena, names map[string]struct{}, diagnostics []Diagnostic) []Diagnostic {
	expression := expressions.Node(expressionID)

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
		for _, childID := range expressions.Children(expression) {
			diagnostics = validateDefinitionExpression(source, childID, expressions, names, diagnostics)
		}

	case ast.DefinitionExpressionGroup, ast.DefinitionExpressionRepetition:
		children := expressions.Children(expression)
		diagnostics = validateDefinitionExpression(source, children[0], expressions, names, diagnostics)
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

func validateDefinitionRecursion(source string, definitions []ast.Definition, expressions ast.DefinitionExpressionArena, diagnostics []Diagnostic) []Diagnostic {
	if len(definitions) == 0 {
		return diagnostics
	}

	definitionByName := map[string]ast.Definition{}

	for idx := range definitions {
		definition := definitions[idx]

		if _, ok := definitionByName[definition.Name.Value(source)]; !ok {
			definitionByName[definition.Name.Value(source)] = definition
		}
	}

	states := map[string]uint8{}

	for idx := range definitions {
		diagnostics = validateDefinitionCycle(source, definitions[idx].Name.Value(source), definitionByName, expressions, states, diagnostics)
	}

	return diagnostics
}

func validateDefinitionCycle(source string, name string, definitions map[string]ast.Definition, expressions ast.DefinitionExpressionArena, states map[string]uint8, diagnostics []Diagnostic) []Diagnostic {
	switch states[name] {
	case 1, 2:
		return diagnostics
	}

	definition, ok := definitions[name]
	if !ok {
		return diagnostics
	}

	states[name] = 1
	diagnostics = validateDefinitionExpressionCycle(source, definition.Expression, expressions, definitions, states, diagnostics)
	states[name] = 2

	return diagnostics
}

func validateDefinitionExpressionCycle(source string, expressionID ast.DefinitionExpressionID, expressions ast.DefinitionExpressionArena, definitions map[string]ast.Definition, states map[string]uint8, diagnostics []Diagnostic) []Diagnostic {
	expression := expressions.Node(expressionID)

	switch expression.Kind {
	case ast.DefinitionExpressionReference:
		if states[expression.Start.Value(source)] == 1 {
			diagnostics = append(diagnostics, Diagnostic{
				Message: `Definition "` + expression.Start.Value(source) + `" is recursive.`,
				Span:    expression.Start.Span,
			})

			return diagnostics
		}

		diagnostics = validateDefinitionCycle(source, expression.Start.Value(source), definitions, expressions, states, diagnostics)

	case ast.DefinitionExpressionAlternation, ast.DefinitionExpressionConcatenation:
		for _, childID := range expressions.Children(expression) {
			diagnostics = validateDefinitionExpressionCycle(source, childID, expressions, definitions, states, diagnostics)
		}

	case ast.DefinitionExpressionGroup, ast.DefinitionExpressionRepetition:
		children := expressions.Children(expression)
		diagnostics = validateDefinitionExpressionCycle(source, children[0], expressions, definitions, states, diagnostics)
	}

	return diagnostics
}
