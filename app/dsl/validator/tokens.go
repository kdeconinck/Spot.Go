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

func validateTokens(source string, resolution resolver.Resolution, diagnostics []Diagnostic) []Diagnostic {
	tokens := resolution.Document.Tokens
	tokenList := resolution.Tokens

	if len(tokenList) == 0 {
		diagnostics = append(diagnostics, Diagnostic{
			Message: "Tokens must contain at least one token.",
			Span:    tokens.Span,
		})

		return diagnostics
	}

	for idx := range tokenList {
		name := tokenList[idx].Name
		nameValue := name.Value(source)

		if _, ok := resolution.DefinitionIndex(nameValue); ok {
			diagnostics = append(diagnostics, Diagnostic{
				Message: `Token "` + nameValue + `" conflicts with a definition of the same name.`,
				Span:    name.Span,
			})
		}

		if firstIndex, ok := resolution.TokenIndex(nameValue); ok && firstIndex != idx {
			diagnostics = append(diagnostics, Diagnostic{
				Message: `Token "` + nameValue + `" is already declared.`,
				Span:    name.Span,
			})

			continue
		}
	}

	for idx := range tokenList {
		expressionID := tokenList[idx].Expression
		expression := resolution.Document.Expressions.Node(expressionID)

		diagnostics = validateTokenExpression(source, expressionID, resolution, diagnostics)

		if tokenExpressionMatchesEmpty(source, expressionID, resolution, 0) {
			diagnostics = append(diagnostics, Diagnostic{
				Message: "Token expression must not match empty input.",
				Span:    expression.Span,
			})
		}
	}

	return diagnostics
}

func validateTokenExpression(source string, expressionID ast.DefinitionExpressionID, resolution resolver.Resolution, diagnostics []Diagnostic) []Diagnostic {
	expressions := resolution.Document.Expressions
	expression := expressions.Node(expressionID)

	switch expression.Kind {
	case ast.DefinitionExpressionReference:
		if _, ok := resolution.DefinitionIndex(expression.Start.Value(source)); !ok {
			diagnostics = append(diagnostics, Diagnostic{
				Message: `Definition "` + expression.Start.Value(source) + `" is not declared.`,
				Span:    expression.Start.Span,
			})
		}

	case ast.DefinitionExpressionAlternation, ast.DefinitionExpressionConcatenation:
		for _, childID := range expressions.Children(expression) {
			diagnostics = validateTokenExpression(source, childID, resolution, diagnostics)
		}

	case ast.DefinitionExpressionGroup, ast.DefinitionExpressionRepetition:
		children := expressions.Children(expression)
		diagnostics = validateTokenExpression(source, children[0], resolution, diagnostics)
	}

	return diagnostics
}

func tokenExpressionMatchesEmpty(source string, expressionID ast.DefinitionExpressionID, resolution resolver.Resolution, depth int) bool {
	matchesEmpty := false
	expressions := resolution.Document.Expressions
	expression := expressions.Node(expressionID)

	switch expression.Kind {
	case ast.DefinitionExpressionCharacter, ast.DefinitionExpressionRange:
		matchesEmpty = false

	case ast.DefinitionExpressionString:
		matchesEmpty = expression.Start.Value(source) == `""`

	case ast.DefinitionExpressionReference:
		if depth >= len(resolution.Definitions) {
			break
		}

		definitionIndex, ok := resolution.DefinitionIndex(expression.Start.Value(source))
		if !ok {
			break
		}

		matchesEmpty = tokenExpressionMatchesEmpty(source, resolution.Definitions[definitionIndex].Expression, resolution, depth+1)

	case ast.DefinitionExpressionConcatenation:
		matchesEmpty = true

		for _, childID := range expressions.Children(expression) {
			if !tokenExpressionMatchesEmpty(source, childID, resolution, depth) {
				matchesEmpty = false

				break
			}
		}

	case ast.DefinitionExpressionAlternation:
		for _, childID := range expressions.Children(expression) {
			if tokenExpressionMatchesEmpty(source, childID, resolution, depth) {
				matchesEmpty = true

				break
			}
		}

	case ast.DefinitionExpressionGroup:
		children := expressions.Children(expression)
		matchesEmpty = tokenExpressionMatchesEmpty(source, children[0], resolution, depth)

	case ast.DefinitionExpressionRepetition:
		if expression.Operator.Kind == token.TokenPlus {
			children := expressions.Children(expression)
			matchesEmpty = tokenExpressionMatchesEmpty(source, children[0], resolution, depth)

			break
		}

		matchesEmpty = true
	}

	return matchesEmpty
}
