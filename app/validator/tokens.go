// =====================================================================================================================
// == LICENSE:                 Copyright (c) 2026 Kevin De Coninck.
// == SPDX-License-Identifier: LicenseRef-PolyForm-Noncommercial-1.0.0
// =====================================================================================================================

// Package validator validates parsed Spot DSL syntax.
package validator

import "github.com/kdeconinck/spot/syntax"

func validateTokens(tokens syntax.TokensSection, definitions syntax.DefinitionsSection, diagnostics []Diagnostic) []Diagnostic {
	if len(tokens.Tokens) == 0 {
		return diagnostics
	}

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

	definitionsByName := map[string]syntax.Definition{}

	for idx := range definitions.Definitions {
		definition := definitions.Definitions[idx]

		if _, ok := definitionsByName[definition.Name.Text]; !ok {
			definitionsByName[definition.Name.Text] = definition
		}
	}

	for idx := range tokens.Tokens {
		expression := tokens.Tokens[idx].Expression

		diagnostics = validateTokenExpression(expression, definitionsByName, diagnostics)

		if tokenExpressionMatchesEmpty(expression, definitionsByName, 0) {
			diagnostics = append(diagnostics, Diagnostic{
				Message: "Token expression must not match empty input.",
				Span:    expression.Span,
			})
		}
	}

	return diagnostics
}

func validateTokenExpression(expression syntax.DefinitionExpression, definitions map[string]syntax.Definition, diagnostics []Diagnostic) []Diagnostic {
	switch expression.Kind {
	case syntax.DefinitionExpressionReference:
		if _, ok := definitions[expression.Start.Text]; !ok {
			diagnostics = append(diagnostics, Diagnostic{
				Message: `Definition "` + expression.Start.Text + `" is not declared.`,
				Span:    expression.Start.Span,
			})
		}

	case syntax.DefinitionExpressionAlternation, syntax.DefinitionExpressionConcatenation:
		for idx := range expression.Terms {
			diagnostics = validateTokenExpression(expression.Terms[idx], definitions, diagnostics)
		}

	case syntax.DefinitionExpressionGroup, syntax.DefinitionExpressionRepetition:
		diagnostics = validateTokenExpression(*expression.Inner, definitions, diagnostics)
	}

	return diagnostics
}

func tokenExpressionMatchesEmpty(expression syntax.DefinitionExpression, definitions map[string]syntax.Definition, depth int) bool {
	matchesEmpty := false

	switch expression.Kind {
	case syntax.DefinitionExpressionCharacter, syntax.DefinitionExpressionRange:
		matchesEmpty = false

	case syntax.DefinitionExpressionString:
		matchesEmpty = expression.Start.Text == `""`

	case syntax.DefinitionExpressionReference:
		if depth >= len(definitions) {
			break
		}

		definition, ok := definitions[expression.Start.Text]
		if !ok {
			break
		}

		matchesEmpty = tokenExpressionMatchesEmpty(definition.Expression, definitions, depth+1)

	case syntax.DefinitionExpressionConcatenation:
		matchesEmpty = true

		for idx := range expression.Terms {
			if !tokenExpressionMatchesEmpty(expression.Terms[idx], definitions, depth) {
				matchesEmpty = false

				break
			}
		}

	case syntax.DefinitionExpressionAlternation:
		for idx := range expression.Terms {
			if tokenExpressionMatchesEmpty(expression.Terms[idx], definitions, depth) {
				matchesEmpty = true

				break
			}
		}

	case syntax.DefinitionExpressionGroup:
		matchesEmpty = tokenExpressionMatchesEmpty(*expression.Inner, definitions, depth)

	case syntax.DefinitionExpressionRepetition:
		if expression.Operator.Kind == syntax.TokenPlus {
			matchesEmpty = tokenExpressionMatchesEmpty(*expression.Inner, definitions, depth)

			break
		}

		matchesEmpty = true
	}

	return matchesEmpty
}
