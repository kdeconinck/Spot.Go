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

func validateSyntax(source string, resolution resolver.Resolution, diagnostics []Diagnostic) []Diagnostic {
	syntaxNodes := resolution.SyntaxNodes

	for idx := range syntaxNodes {
		name := syntaxNodes[idx].Name
		nameValue := name.Value(source)

		if firstIndex, ok := resolution.SyntaxNodeIndex(nameValue); ok && firstIndex != idx {
			diagnostics = append(diagnostics, Diagnostic{
				Message: `Syntax node "` + nameValue + `" is already declared.`,
				Span:    name.Span,
			})

			continue
		}

		if _, ok := resolution.TokenIndex(nameValue); ok {
			diagnostics = append(diagnostics, Diagnostic{
				Message: `Syntax node "` + nameValue + `" conflicts with token "` + nameValue + `".`,
				Span:    name.Span,
			})
		}
	}

	for idx := range syntaxNodes {
		diagnostics = validateSyntaxExpression(source, syntaxNodes[idx].Expression, resolution, diagnostics)
	}

	diagnostics = validateSyntaxRecursion(source, resolution, diagnostics)

	return diagnostics
}

func validateSyntaxExpression(source string, expressionID ast.SyntaxExpressionID, resolution resolver.Resolution, diagnostics []Diagnostic) []Diagnostic {
	expressions := resolution.Document.SyntaxExpressions
	expression := expressions.Node(expressionID)

	switch expression.Kind {
	case ast.SyntaxExpressionReference:
		referenceName := expression.Reference.Value(source)
		_, tokenDeclared := resolution.TokenIndex(referenceName)
		_, nodeDeclared := resolution.SyntaxNodeIndex(referenceName)

		if !tokenDeclared && !nodeDeclared {
			diagnostics = append(diagnostics, Diagnostic{
				Message: `Syntax reference "` + referenceName + `" is not declared as a token or syntax node.`,
				Span:    expression.Reference.Span,
			})
		}

	case ast.SyntaxExpressionAny:
	case ast.SyntaxExpressionAlternation, ast.SyntaxExpressionConcatenation:
		for _, childID := range expressions.Children(expression) {
			diagnostics = validateSyntaxExpression(source, childID, resolution, diagnostics)
		}

	case ast.SyntaxExpressionGroup, ast.SyntaxExpressionRepetition:
		children := expressions.Children(expression)
		diagnostics = validateSyntaxExpression(source, children[0], resolution, diagnostics)

		if expression.Kind == ast.SyntaxExpressionRepetition && syntaxExpressionMatchesEmpty(source, children[0], resolution, 0) {
			diagnostics = append(diagnostics, Diagnostic{
				Message: "Syntax repetition expression must not match empty input.",
				Span:    expression.Span,
			})
		}
	}

	return diagnostics
}

func validateSyntaxRecursion(source string, resolution resolver.Resolution, diagnostics []Diagnostic) []Diagnostic {
	syntaxNodes := resolution.SyntaxNodes

	if len(syntaxNodes) == 0 {
		return diagnostics
	}

	states := make([]uint8, len(syntaxNodes))

	for idx := range syntaxNodes {
		diagnostics = validateSyntaxCycle(source, idx, resolution, states, diagnostics)
	}

	return diagnostics
}

func validateSyntaxCycle(source string, syntaxNodeIndex int, resolution resolver.Resolution, states []uint8, diagnostics []Diagnostic) []Diagnostic {
	switch states[syntaxNodeIndex] {
	case 1, 2:
		return diagnostics
	}

	syntaxNode := resolution.SyntaxNodes[syntaxNodeIndex]

	states[syntaxNodeIndex] = 1
	diagnostics = validateSyntaxExpressionCycle(source, syntaxNode.Expression, resolution, states, diagnostics)
	states[syntaxNodeIndex] = 2

	return diagnostics
}

func validateSyntaxExpressionCycle(source string, expressionID ast.SyntaxExpressionID, resolution resolver.Resolution, states []uint8, diagnostics []Diagnostic) []Diagnostic {
	expressions := resolution.Document.SyntaxExpressions
	expression := expressions.Node(expressionID)

	switch expression.Kind {
	case ast.SyntaxExpressionReference:
		referencedIndex, ok := resolution.SyntaxNodeIndex(expression.Reference.Value(source))

		if !ok {
			return diagnostics
		}

		if states[referencedIndex] == 1 {
			diagnostics = append(diagnostics, Diagnostic{
				Message: `Syntax node "` + expression.Reference.Value(source) + `" is recursive.`,
				Span:    expression.Reference.Span,
			})

			return diagnostics
		}

		diagnostics = validateSyntaxCycle(source, referencedIndex, resolution, states, diagnostics)

	case ast.SyntaxExpressionAny:
	case ast.SyntaxExpressionAlternation, ast.SyntaxExpressionConcatenation:
		for _, childID := range expressions.Children(expression) {
			diagnostics = validateSyntaxExpressionCycle(source, childID, resolution, states, diagnostics)
		}

	case ast.SyntaxExpressionGroup, ast.SyntaxExpressionRepetition:
		children := expressions.Children(expression)
		diagnostics = validateSyntaxExpressionCycle(source, children[0], resolution, states, diagnostics)
	}

	return diagnostics
}

func syntaxExpressionMatchesEmpty(source string, expressionID ast.SyntaxExpressionID, resolution resolver.Resolution, depth int) bool {
	if depth > len(resolution.SyntaxNodes) {
		return false
	}

	expressions := resolution.Document.SyntaxExpressions
	expression := expressions.Node(expressionID)
	children := expressions.Children(expression)

	switch expression.Kind {
	case ast.SyntaxExpressionReference:
		if _, ok := resolution.TokenIndex(expression.Reference.Value(source)); ok {
			return false
		}

		syntaxNodeIndex, ok := resolution.SyntaxNodeIndex(expression.Reference.Value(source))

		if !ok {
			return false
		}

		return syntaxExpressionMatchesEmpty(source, resolution.SyntaxNodes[syntaxNodeIndex].Expression, resolution, depth+1)

	case ast.SyntaxExpressionAny:
		return false

	case ast.SyntaxExpressionConcatenation:
		for _, childID := range children {
			if !syntaxExpressionMatchesEmpty(source, childID, resolution, depth) {
				return false
			}
		}

		return true

	case ast.SyntaxExpressionAlternation:
		for _, childID := range children {
			if syntaxExpressionMatchesEmpty(source, childID, resolution, depth) {
				return true
			}
		}

		return false

	case ast.SyntaxExpressionGroup:
		return syntaxExpressionMatchesEmpty(source, children[0], resolution, depth)

	case ast.SyntaxExpressionRepetition:
		if expression.Operator.Kind == token.TokenPlus {
			return syntaxExpressionMatchesEmpty(source, children[0], resolution, depth)
		}

		return true

	default:
		return false
	}
}
