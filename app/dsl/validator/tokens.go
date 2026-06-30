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

func validateTokens(source string, tokens ast.TokensSection, tokenList []ast.TokenDefinition, definitions []ast.Definition, expressions ast.DefinitionExpressionArena, diagnostics []Diagnostic) []Diagnostic {
	if len(tokenList) == 0 {
		diagnostics = append(diagnostics, Diagnostic{
			Message: "Tokens must contain at least one token.",
			Span:    tokens.Span,
		})

		return diagnostics
	}

	definitionsByName := map[string]ast.Definition{}

	for idx := range definitions {
		definition := definitions[idx]

		if _, ok := definitionsByName[definition.Name.Value(source)]; !ok {
			definitionsByName[definition.Name.Value(source)] = definition
		}
	}

	names := map[string]struct{}{}

	for idx := range tokenList {
		name := tokenList[idx].Name

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

	for idx := range tokenList {
		expressionID := tokenList[idx].Expression
		expression := expressions.Node(expressionID)

		diagnostics = validateTokenExpression(source, expressionID, expressions, definitionsByName, diagnostics)

		if tokenExpressionMatchesEmpty(source, expressionID, expressions, definitionsByName, 0) {
			diagnostics = append(diagnostics, Diagnostic{
				Message: "Token expression must not match empty input.",
				Span:    expression.Span,
			})
		}
	}

	return diagnostics
}

func validateTokenExpression(source string, expressionID ast.DefinitionExpressionID, expressions ast.DefinitionExpressionArena, definitions map[string]ast.Definition, diagnostics []Diagnostic) []Diagnostic {
	expression := expressions.Node(expressionID)

	switch expression.Kind {
	case ast.DefinitionExpressionReference:
		if _, ok := definitions[expression.Start.Value(source)]; !ok {
			diagnostics = append(diagnostics, Diagnostic{
				Message: `Definition "` + expression.Start.Value(source) + `" is not declared.`,
				Span:    expression.Start.Span,
			})
		}

	case ast.DefinitionExpressionAlternation, ast.DefinitionExpressionConcatenation:
		for _, childID := range expressions.Children(expression) {
			diagnostics = validateTokenExpression(source, childID, expressions, definitions, diagnostics)
		}

	case ast.DefinitionExpressionGroup, ast.DefinitionExpressionRepetition:
		children := expressions.Children(expression)
		diagnostics = validateTokenExpression(source, children[0], expressions, definitions, diagnostics)
	}

	return diagnostics
}

func tokenExpressionMatchesEmpty(source string, expressionID ast.DefinitionExpressionID, expressions ast.DefinitionExpressionArena, definitions map[string]ast.Definition, depth int) bool {
	matchesEmpty := false
	expression := expressions.Node(expressionID)

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

		matchesEmpty = tokenExpressionMatchesEmpty(source, definition.Expression, expressions, definitions, depth+1)

	case ast.DefinitionExpressionConcatenation:
		matchesEmpty = true

		for _, childID := range expressions.Children(expression) {
			if !tokenExpressionMatchesEmpty(source, childID, expressions, definitions, depth) {
				matchesEmpty = false

				break
			}
		}

	case ast.DefinitionExpressionAlternation:
		for _, childID := range expressions.Children(expression) {
			if tokenExpressionMatchesEmpty(source, childID, expressions, definitions, depth) {
				matchesEmpty = true

				break
			}
		}

	case ast.DefinitionExpressionGroup:
		children := expressions.Children(expression)
		matchesEmpty = tokenExpressionMatchesEmpty(source, children[0], expressions, definitions, depth)

	case ast.DefinitionExpressionRepetition:
		if expression.Operator.Kind == token.TokenPlus {
			children := expressions.Children(expression)
			matchesEmpty = tokenExpressionMatchesEmpty(source, children[0], expressions, definitions, depth)

			break
		}

		matchesEmpty = true
	}

	return matchesEmpty
}
