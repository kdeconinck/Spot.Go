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

func validateTokens(source string, tokens ast.TokensSection, definitions ast.DefinitionsSection, diagnostics []Diagnostic) []Diagnostic {
	if len(tokens.Tokens) == 0 {
		diagnostics = append(diagnostics, Diagnostic{
			Message: "Tokens must contain at least one token.",
			Span:    tokens.Span,
		})

		return diagnostics
	}

	definitionsByName := map[string]ast.Definition{}

	for idx := range definitions.Definitions {
		definition := definitions.Definitions[idx]

		if _, ok := definitionsByName[definition.Name.Value(source)]; !ok {
			definitionsByName[definition.Name.Value(source)] = definition
		}
	}

	names := map[string]struct{}{}

	for idx := range tokens.Tokens {
		name := tokens.Tokens[idx].Name

		if _, ok := definitionsByName[name.Value(source)]; ok {
			diagnostics = append(diagnostics, Diagnostic{
				Message: `Token "` + name.Value(source) + `" conflicts with a definition of the same name.`,
				Span:    name.Span,
			})
		}

		if _, ok := names[name.Value(source)]; ok {
			diagnostics = append(diagnostics, Diagnostic{
				Message: `Token "` + name.Value(source) + `" is already declared.`,
				Span:    name.Span,
			})

			continue
		}

		names[name.Value(source)] = struct{}{}
	}

	for idx := range tokens.Tokens {
		expression := tokens.Tokens[idx].Expression

		diagnostics = validateTokenExpression(source, expression, definitionsByName, diagnostics)

		if tokenExpressionMatchesEmpty(source, expression, definitionsByName, 0) {
			diagnostics = append(diagnostics, Diagnostic{
				Message: "Token expression must not match empty input.",
				Span:    expression.Span,
			})
		}
	}

	return diagnostics
}

func validateTokenExpression(source string, expression ast.DefinitionExpression, definitions map[string]ast.Definition, diagnostics []Diagnostic) []Diagnostic {
	switch expression.Kind {
	case ast.DefinitionExpressionReference:
		if _, ok := definitions[expression.Start.Value(source)]; !ok {
			diagnostics = append(diagnostics, Diagnostic{
				Message: `Definition "` + expression.Start.Value(source) + `" is not declared.`,
				Span:    expression.Start.Span,
			})
		}

	case ast.DefinitionExpressionAlternation, ast.DefinitionExpressionConcatenation:
		for idx := range expression.Terms {
			diagnostics = validateTokenExpression(source, expression.Terms[idx], definitions, diagnostics)
		}

	case ast.DefinitionExpressionGroup, ast.DefinitionExpressionRepetition:
		diagnostics = validateTokenExpression(source, *expression.Inner, definitions, diagnostics)
	}

	return diagnostics
}

func tokenExpressionMatchesEmpty(source string, expression ast.DefinitionExpression, definitions map[string]ast.Definition, depth int) bool {
	matchesEmpty := false

	switch expression.Kind {
	case ast.DefinitionExpressionCharacter, ast.DefinitionExpressionRange:
		matchesEmpty = false

	case ast.DefinitionExpressionString:
		matchesEmpty = expression.Start.Value(source) == `""`

	case ast.DefinitionExpressionReference:
		if depth >= len(definitions) {
			break
		}

		definition, ok := definitions[expression.Start.Value(source)]
		if !ok {
			break
		}

		matchesEmpty = tokenExpressionMatchesEmpty(source, definition.Expression, definitions, depth+1)

	case ast.DefinitionExpressionConcatenation:
		matchesEmpty = true

		for idx := range expression.Terms {
			if !tokenExpressionMatchesEmpty(source, expression.Terms[idx], definitions, depth) {
				matchesEmpty = false

				break
			}
		}

	case ast.DefinitionExpressionAlternation:
		for idx := range expression.Terms {
			if tokenExpressionMatchesEmpty(source, expression.Terms[idx], definitions, depth) {
				matchesEmpty = true

				break
			}
		}

	case ast.DefinitionExpressionGroup:
		matchesEmpty = tokenExpressionMatchesEmpty(source, *expression.Inner, definitions, depth)

	case ast.DefinitionExpressionRepetition:
		if expression.Operator.Kind == token.TokenPlus {
			matchesEmpty = tokenExpressionMatchesEmpty(source, *expression.Inner, definitions, depth)

			break
		}

		matchesEmpty = true
	}

	return matchesEmpty
}
