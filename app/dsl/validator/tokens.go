// =====================================================================================================================
// == LICENSE:                 Copyright (c) 2026 Kevin De Coninck.
// == SPDX-License-Identifier: LicenseRef-PolyForm-Noncommercial-1.0.0
// =====================================================================================================================

// Package validator validates parsed Spot DSL token.
package validator

import "github.com/kdeconinck/spot/dsl/token"

func validateTokens(tokens token.TokensSection, definitions token.DefinitionsSection, diagnostics []Diagnostic) []Diagnostic {
	if len(tokens.Tokens) == 0 {
		diagnostics = append(diagnostics, Diagnostic{
			Message: "Tokens must contain at least one token.",
			Span:    tokens.Span,
		})

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

	definitionsByName := map[string]token.Definition{}

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

func validateTokenExpression(expression token.DefinitionExpression, definitions map[string]token.Definition, diagnostics []Diagnostic) []Diagnostic {
	switch expression.Kind {
	case token.DefinitionExpressionReference:
		if _, ok := definitions[expression.Start.Text]; !ok {
			diagnostics = append(diagnostics, Diagnostic{
				Message: `Definition "` + expression.Start.Text + `" is not declared.`,
				Span:    expression.Start.Span,
			})
		}

	case token.DefinitionExpressionAlternation, token.DefinitionExpressionConcatenation:
		for idx := range expression.Terms {
			diagnostics = validateTokenExpression(expression.Terms[idx], definitions, diagnostics)
		}

	case token.DefinitionExpressionGroup, token.DefinitionExpressionRepetition:
		diagnostics = validateTokenExpression(*expression.Inner, definitions, diagnostics)
	}

	return diagnostics
}

func tokenExpressionMatchesEmpty(expression token.DefinitionExpression, definitions map[string]token.Definition, depth int) bool {
	matchesEmpty := false

	switch expression.Kind {
	case token.DefinitionExpressionCharacter, token.DefinitionExpressionRange:
		matchesEmpty = false

	case token.DefinitionExpressionString:
		matchesEmpty = expression.Start.Text == `""`

	case token.DefinitionExpressionReference:
		if depth >= len(definitions) {
			break
		}

		definition, ok := definitions[expression.Start.Text]
		if !ok {
			break
		}

		matchesEmpty = tokenExpressionMatchesEmpty(definition.Expression, definitions, depth+1)

	case token.DefinitionExpressionConcatenation:
		matchesEmpty = true

		for idx := range expression.Terms {
			if !tokenExpressionMatchesEmpty(expression.Terms[idx], definitions, depth) {
				matchesEmpty = false

				break
			}
		}

	case token.DefinitionExpressionAlternation:
		for idx := range expression.Terms {
			if tokenExpressionMatchesEmpty(expression.Terms[idx], definitions, depth) {
				matchesEmpty = true

				break
			}
		}

	case token.DefinitionExpressionGroup:
		matchesEmpty = tokenExpressionMatchesEmpty(*expression.Inner, definitions, depth)

	case token.DefinitionExpressionRepetition:
		if expression.Operator.Kind == token.TokenPlus {
			matchesEmpty = tokenExpressionMatchesEmpty(*expression.Inner, definitions, depth)

			break
		}

		matchesEmpty = true
	}

	return matchesEmpty
}
