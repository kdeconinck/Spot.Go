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

	case 't':
		return '\t'

	default:
		return token.Text[2]
	}
}
