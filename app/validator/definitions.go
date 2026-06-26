// =====================================================================================================================
// == LICENSE:                 Copyright (c) 2026 Kevin De Coninck.
// == SPDX-License-Identifier: LicenseRef-PolyForm-Noncommercial-1.0.0
// =====================================================================================================================

// Package validator validates parsed Spot DSL syntax.
package validator

import "github.com/kdeconinck/spot/syntax"

func validateDefinitions(definitions syntax.DefinitionsSection, diagnostics []Diagnostic) []Diagnostic {
	names := map[string]struct{}{}

	for idx := range definitions.Definitions {
		name := definitions.Definitions[idx].Name

		if _, ok := names[name.Text]; ok {
			diagnostics = append(diagnostics, Diagnostic{
				Message: `Definition "` + name.Text + `" is already declared.`,
				Span:    name.Span,
			})

			continue
		}

		names[name.Text] = struct{}{}
	}

	for idx := range definitions.Definitions {
		diagnostics = validateDefinitionExpression(definitions.Definitions[idx].Expression, names, diagnostics)
	}

	diagnostics = validateDefinitionRecursion(definitions, diagnostics)

	return diagnostics
}

func validateDefinitionExpression(expression syntax.DefinitionExpression, names map[string]struct{}, diagnostics []Diagnostic) []Diagnostic {
	switch expression.Kind {
	case syntax.DefinitionExpressionRange:
		if characterValue(expression.Start) > characterValue(expression.End) {
			diagnostics = append(diagnostics, Diagnostic{
				Message: "Character range start must be less than or equal to end.",
				Span:    expression.Span,
			})
		}

	case syntax.DefinitionExpressionReference:
		if _, ok := names[expression.Start.Text]; !ok {
			diagnostics = append(diagnostics, Diagnostic{
				Message: `Definition "` + expression.Start.Text + `" is not declared.`,
				Span:    expression.Start.Span,
			})
		}

	case syntax.DefinitionExpressionAlternation, syntax.DefinitionExpressionConcatenation:
		for idx := range expression.Terms {
			diagnostics = validateDefinitionExpression(expression.Terms[idx], names, diagnostics)
		}

	case syntax.DefinitionExpressionGroup, syntax.DefinitionExpressionRepetition:
		diagnostics = validateDefinitionExpression(*expression.Inner, names, diagnostics)
	}

	return diagnostics
}

func characterValue(token syntax.Token) byte {
	if token.Text[1] != '\\' {
		return token.Text[1]
	}

	switch token.Text[2] {
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

func validateDefinitionRecursion(definitions syntax.DefinitionsSection, diagnostics []Diagnostic) []Diagnostic {
	if len(definitions.Definitions) == 0 {
		return diagnostics
	}

	if len(definitions.Definitions) == 1 {
		definition := definitions.Definitions[0]

		return validateDefinitionSelfReference(definition.Name.Text, definition.Expression, diagnostics)
	}

	definitionByName := map[string]syntax.Definition{}

	for idx := range definitions.Definitions {
		definition := definitions.Definitions[idx]

		if _, ok := definitionByName[definition.Name.Text]; !ok {
			definitionByName[definition.Name.Text] = definition
		}
	}

	states := map[string]uint8{}

	for idx := range definitions.Definitions {
		diagnostics = validateDefinitionCycle(definitions.Definitions[idx].Name.Text, definitionByName, states, diagnostics)
	}

	return diagnostics
}

func validateDefinitionSelfReference(name string, expression syntax.DefinitionExpression, diagnostics []Diagnostic) []Diagnostic {
	switch expression.Kind {
	case syntax.DefinitionExpressionReference:
		if expression.Start.Text == name {
			diagnostics = append(diagnostics, Diagnostic{
				Message: `Definition "` + expression.Start.Text + `" is recursive.`,
				Span:    expression.Start.Span,
			})
		}

	case syntax.DefinitionExpressionAlternation, syntax.DefinitionExpressionConcatenation:
		for idx := range expression.Terms {
			diagnostics = validateDefinitionSelfReference(name, expression.Terms[idx], diagnostics)
		}

	case syntax.DefinitionExpressionGroup, syntax.DefinitionExpressionRepetition:
		diagnostics = validateDefinitionSelfReference(name, *expression.Inner, diagnostics)
	}

	return diagnostics
}

func validateDefinitionCycle(name string, definitions map[string]syntax.Definition, states map[string]uint8, diagnostics []Diagnostic) []Diagnostic {
	switch states[name] {
	case 1, 2:
		return diagnostics
	}

	definition, ok := definitions[name]
	if !ok {
		return diagnostics
	}

	states[name] = 1
	diagnostics = validateDefinitionExpressionCycle(definition.Expression, definitions, states, diagnostics)
	states[name] = 2

	return diagnostics
}

func validateDefinitionExpressionCycle(expression syntax.DefinitionExpression, definitions map[string]syntax.Definition, states map[string]uint8, diagnostics []Diagnostic) []Diagnostic {
	switch expression.Kind {
	case syntax.DefinitionExpressionReference:
		if states[expression.Start.Text] == 1 {
			diagnostics = append(diagnostics, Diagnostic{
				Message: `Definition "` + expression.Start.Text + `" is recursive.`,
				Span:    expression.Start.Span,
			})

			return diagnostics
		}

		diagnostics = validateDefinitionCycle(expression.Start.Text, definitions, states, diagnostics)

	case syntax.DefinitionExpressionAlternation, syntax.DefinitionExpressionConcatenation:
		for idx := range expression.Terms {
			diagnostics = validateDefinitionExpressionCycle(expression.Terms[idx], definitions, states, diagnostics)
		}

	case syntax.DefinitionExpressionGroup, syntax.DefinitionExpressionRepetition:
		diagnostics = validateDefinitionExpressionCycle(*expression.Inner, definitions, states, diagnostics)
	}

	return diagnostics
}
