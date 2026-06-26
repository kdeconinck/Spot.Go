// =====================================================================================================================
// == LICENSE:                 Copyright (c) 2026 Kevin De Coninck.
// == SPDX-License-Identifier: LicenseRef-PolyForm-Noncommercial-1.0.0
// =====================================================================================================================

// Package validator validates parsed Spot DSL syntax.
package validator

import "github.com/kdeconinck/spot/syntax"

func validateTokens(tokens syntax.TokensSection, definitions syntax.DefinitionsSection, diagnostics []Diagnostic) []Diagnostic {
	names := map[string]struct{}{}

	for idx := range tokens.Tokens {
		name := tokens.Tokens[idx].Name

		if _, ok := names[name.Text]; ok {
			diagnostics = append(diagnostics, Diagnostic{
				Message: `Token "` + name.Text + `" is already declared.`,
				Span:    name.Span,
			})

			continue
		}

		names[name.Text] = struct{}{}
	}

	definitionNames := map[string]struct{}{}

	for idx := range definitions.Definitions {
		definitionNames[definitions.Definitions[idx].Name.Text] = struct{}{}
	}

	for idx := range tokens.Tokens {
		diagnostics = validateTokenExpression(tokens.Tokens[idx].Expression, definitionNames, diagnostics)
	}

	return diagnostics
}

func validateTokenExpression(expression syntax.DefinitionExpression, definitionNames map[string]struct{}, diagnostics []Diagnostic) []Diagnostic {
	switch expression.Kind {
	case syntax.DefinitionExpressionReference:
		if _, ok := definitionNames[expression.Start.Text]; !ok {
			diagnostics = append(diagnostics, Diagnostic{
				Message: `Definition "` + expression.Start.Text + `" is not declared.`,
				Span:    expression.Start.Span,
			})
		}

	case syntax.DefinitionExpressionAlternation, syntax.DefinitionExpressionConcatenation:
		for idx := range expression.Terms {
			diagnostics = validateTokenExpression(expression.Terms[idx], definitionNames, diagnostics)
		}

	case syntax.DefinitionExpressionGroup, syntax.DefinitionExpressionRepetition:
		diagnostics = validateTokenExpression(*expression.Inner, definitionNames, diagnostics)
	}

	return diagnostics
}
