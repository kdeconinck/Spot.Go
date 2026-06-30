// =====================================================================================================================
// == LICENSE:                 Copyright (c) 2026 Kevin De Coninck.
// == SPDX-License-Identifier: LicenseRef-PolyForm-Noncommercial-1.0.0
// =====================================================================================================================

// Package validator validates parsed Spot DSL syntax.
package validator

import (
	"github.com/kdeconinck/spot/dsl/ast"
	"github.com/kdeconinck/spot/dsl/resolver"
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

	case ast.SyntaxExpressionAlternation, ast.SyntaxExpressionConcatenation:
		for _, childID := range expressions.Children(expression) {
			diagnostics = validateSyntaxExpression(source, childID, resolution, diagnostics)
		}

	case ast.SyntaxExpressionGroup, ast.SyntaxExpressionRepetition:
		children := expressions.Children(expression)
		diagnostics = validateSyntaxExpression(source, children[0], resolution, diagnostics)
	}

	return diagnostics
}
