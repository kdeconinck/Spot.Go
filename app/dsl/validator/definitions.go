// =====================================================================================================================
// == LICENSE:                 Copyright (c) 2026 Kevin De Coninck.
// == SPDX-License-Identifier: LicenseRef-PolyForm-Noncommercial-1.0.0
// =====================================================================================================================

// Package validator validates parsed Spot DSL syntax.
package validator

import (
	"github.com/kdeconinck/spot/dsl/ast"
	"github.com/kdeconinck/spot/dsl/resolver"
	"github.com/kdeconinck/spot/dsl/token"
)

func validateDefinitions(source string, resolution resolver.Resolution, diagnostics []Diagnostic) []Diagnostic {
	definitions := resolution.Definitions

	for idx := range definitions {
		name := definitions[idx].Name
		nameValue := name.Value(source)

		if firstIndex, ok := resolution.DefinitionIndex(nameValue); ok && firstIndex != idx {
			diagnostics = append(diagnostics, Diagnostic{
				Message: `Definition "` + nameValue + `" is already declared.`,
				Span:    name.Span,
			})

			continue
		}
	}

	for idx := range definitions {
		diagnostics = validateDefinitionExpression(source, definitions[idx].Expression, resolution, diagnostics)
	}

	diagnostics = validateDefinitionRecursion(source, resolution, diagnostics)

	return diagnostics
}

func validateDefinitionExpression(source string, expressionID ast.DefinitionExpressionID, resolution resolver.Resolution, diagnostics []Diagnostic) []Diagnostic {
	expressions := resolution.Document.Expressions
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
		if _, ok := resolution.DefinitionIndex(expression.Start.Value(source)); !ok {
			diagnostics = append(diagnostics, Diagnostic{
				Message: `Definition "` + expression.Start.Value(source) + `" is not declared.`,
				Span:    expression.Start.Span,
			})
		}

	case ast.DefinitionExpressionAlternation, ast.DefinitionExpressionConcatenation:
		for _, childID := range expressions.Children(expression) {
			diagnostics = validateDefinitionExpression(source, childID, resolution, diagnostics)
		}

	case ast.DefinitionExpressionGroup, ast.DefinitionExpressionRepetition:
		children := expressions.Children(expression)
		diagnostics = validateDefinitionExpression(source, children[0], resolution, diagnostics)
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

func validateDefinitionRecursion(source string, resolution resolver.Resolution, diagnostics []Diagnostic) []Diagnostic {
	definitions := resolution.Definitions

	if len(definitions) == 0 {
		return diagnostics
	}

	states := make([]uint8, len(definitions))

	for idx := range definitions {
		diagnostics = validateDefinitionCycle(source, idx, resolution, states, diagnostics)
	}

	return diagnostics
}

func validateDefinitionCycle(source string, definitionIndex int, resolution resolver.Resolution, states []uint8, diagnostics []Diagnostic) []Diagnostic {
	switch states[definitionIndex] {
	case 1, 2:
		return diagnostics
	}

	definition := resolution.Definitions[definitionIndex]

	states[definitionIndex] = 1
	diagnostics = validateDefinitionExpressionCycle(source, definition.Expression, resolution, states, diagnostics)
	states[definitionIndex] = 2

	return diagnostics
}

func validateDefinitionExpressionCycle(source string, expressionID ast.DefinitionExpressionID, resolution resolver.Resolution, states []uint8, diagnostics []Diagnostic) []Diagnostic {
	expressions := resolution.Document.Expressions
	expression := expressions.Node(expressionID)

	switch expression.Kind {
	case ast.DefinitionExpressionReference:
		referencedIndex, ok := resolution.DefinitionIndex(expression.Start.Value(source))

		if !ok {
			return diagnostics
		}

		if states[referencedIndex] == 1 {
			diagnostics = append(diagnostics, Diagnostic{
				Message: `Definition "` + expression.Start.Value(source) + `" is recursive.`,
				Span:    expression.Start.Span,
			})

			return diagnostics
		}

		diagnostics = validateDefinitionCycle(source, referencedIndex, resolution, states, diagnostics)

	case ast.DefinitionExpressionAlternation, ast.DefinitionExpressionConcatenation:
		for _, childID := range expressions.Children(expression) {
			diagnostics = validateDefinitionExpressionCycle(source, childID, resolution, states, diagnostics)
		}

	case ast.DefinitionExpressionGroup, ast.DefinitionExpressionRepetition:
		children := expressions.Children(expression)
		diagnostics = validateDefinitionExpressionCycle(source, children[0], resolution, states, diagnostics)
	}

	return diagnostics
}
