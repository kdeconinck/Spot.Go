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

func validateRules(source string, resolution resolver.Resolution, diagnostics []Diagnostic) []Diagnostic {
	rules := resolution.Rules
	hasSyntaxRules := false

	if len(rules) > 1 {
		for idx := range rules {
			name := rules[idx].Name
			nameValue := name.Value(source)

			if nameValue == "" {
				continue
			}

			if firstIndex, ok := resolution.RuleIndex(nameValue); ok && firstIndex != idx {
				diagnostics = append(diagnostics, Diagnostic{
					Message: `Rule "` + nameValue + `" is already declared.`,
					Span:    name.Span,
				})

				continue
			}
		}
	}

	for idx := range rules {
		if rules[idx].Match.Kind == ast.RuleMatchNode {
			hasSyntaxRules = true
		}

		diagnostics = validateRuleReferences(source, rules[idx], resolution, diagnostics)
	}

	if hasSyntaxRules {
		rootCount := amountOfSyntaxRoots(source, resolution)

		if rootCount != 1 {
			diagnostics = append(diagnostics, Diagnostic{
				Message: "Syntax rules require exactly one root syntax node.",
				Span:    resolution.Document.Rules.Span,
			})
		}
	}

	return diagnostics
}

func validateRuleReferences(source string, rule ast.Rule, resolution resolver.Resolution, diagnostics []Diagnostic) []Diagnostic {
	matchedTarget := rule.Match.Target
	matchedTargetName := matchedTarget.Value(source)
	where := rule.Where
	report := rule.Report

	switch rule.Match.Kind {
	case ast.RuleMatchToken:
		if _, ok := resolution.TokenIndex(matchedTargetName); !ok {
			diagnostics = append(diagnostics, Diagnostic{
				Message: `Token "` + matchedTargetName + `" is not declared.`,
				Span:    matchedTarget.Span,
			})
		}

		if where.Subject.Value(source) != "" && where.Subject.Value(source) != matchedTargetName {
			diagnostics = append(diagnostics, Diagnostic{
				Message: `Where clause must reference matched token "` + matchedTargetName + `".`,
				Span:    where.Subject.Span,
			})
		}

		if report.Target.Value(source) != "" && report.Target.Value(source) != matchedTargetName {
			diagnostics = append(diagnostics, Diagnostic{
				Message: `Report target must reference matched token "` + matchedTargetName + `".`,
				Span:    report.Target.Span,
			})
		}

		diagnostics = validateRulePropertyRules(where, "Token", source, diagnostics)

	default:
		if _, ok := resolution.SyntaxNodeIndex(matchedTargetName); !ok {
			diagnostics = append(diagnostics, Diagnostic{
				Message: `Syntax node "` + matchedTargetName + `" is not declared.`,
				Span:    matchedTarget.Span,
			})
		}

		if where.Subject.Value(source) != "" && where.Subject.Value(source) != matchedTargetName {
			diagnostics = append(diagnostics, Diagnostic{
				Message: `Where clause must reference matched syntax node "` + matchedTargetName + `".`,
				Span:    where.Subject.Span,
			})
		}

		if report.Target.Value(source) != "" && report.Target.Value(source) != matchedTargetName {
			diagnostics = append(diagnostics, Diagnostic{
				Message: `Report target must reference matched syntax node "` + matchedTargetName + `".`,
				Span:    report.Target.Span,
			})
		}

		if rule.Match.ScopeKind != ast.RuleMatchScopeNone {
			if _, ok := resolution.SyntaxNodeIndex(rule.Match.ScopeTarget.Value(source)); !ok {
				diagnostics = append(diagnostics, Diagnostic{
					Message: `Syntax node "` + rule.Match.ScopeTarget.Value(source) + `" is not declared.`,
					Span:    rule.Match.ScopeTarget.Span,
				})
			}
		}

		diagnostics = validateRulePropertyRules(where, "Syntax node", source, diagnostics)
	}

	if rule.Match.Kind == ast.RuleMatchToken && rule.Match.ScopeKind != ast.RuleMatchScopeNone {
		diagnostics = append(diagnostics, Diagnostic{
			Message: "Only syntax-node rules may use inside/outside constraints.",
			Span:    rule.Match.Span,
		})
	}

	return diagnostics
}

func validateRulePropertyRules(where ast.RuleCondition, subjectLabel string, source string, diagnostics []Diagnostic) []Diagnostic {
	if where.Property.Value(source) != "" && where.Property.Value(source) != "text" && where.Property.Value(source) != "length" {
		diagnostics = append(diagnostics, Diagnostic{
			Message: subjectLabel + ` property "` + where.Property.Value(source) + `" is not declared.`,
			Span:    where.Property.Span,
		})
	}

	if where.Property.Value(source) == "text" && where.Operator.Kind != token.TokenEqualEqual && where.Operator.Kind != token.TokenBangEqual {
		diagnostics = append(diagnostics, Diagnostic{
			Message: subjectLabel + ` property "text" only supports equality operators.`,
			Span:    where.Operator.Span,
		})
	}

	if where.Property.Value(source) == "text" && where.Value.Kind != token.TokenString {
		diagnostics = append(diagnostics, Diagnostic{
			Message: subjectLabel + ` property "text" must be compared with a string literal.`,
			Span:    where.Value.Span,
		})
	}

	if where.Property.Value(source) == "length" && where.Value.Kind != token.TokenInteger {
		diagnostics = append(diagnostics, Diagnostic{
			Message: subjectLabel + ` property "length" must be compared with an integer literal.`,
			Span:    where.Value.Span,
		})
	}

	return diagnostics
}

func amountOfSyntaxRoots(source string, resolution resolver.Resolution) int {
	referenced := make([]bool, len(resolution.SyntaxNodes))
	expressions := resolution.Document.SyntaxExpressions

	for idx := range resolution.SyntaxNodes {
		markReferencedSyntaxNodes(source, resolution, expressions, resolution.SyntaxNodes[idx].Expression, referenced)
	}

	amountOfRoots := 0

	for idx := range referenced {
		if !referenced[idx] {
			amountOfRoots++
		}
	}

	return amountOfRoots
}

func markReferencedSyntaxNodes(source string, resolution resolver.Resolution, expressions ast.SyntaxExpressionArena, expressionID ast.SyntaxExpressionID, referenced []bool) {
	expression := expressions.Node(expressionID)

	switch expression.Kind {
	case ast.SyntaxExpressionReference:
		referencedIndex, ok := resolution.SyntaxNodeIndex(expression.Reference.Value(source))

		if ok {
			referenced[referencedIndex] = true
		}

	case ast.SyntaxExpressionAny:
	case ast.SyntaxExpressionAlternation, ast.SyntaxExpressionConcatenation:
		for _, childID := range expressions.Children(expression) {
			markReferencedSyntaxNodes(source, resolution, expressions, childID, referenced)
		}

	case ast.SyntaxExpressionGroup, ast.SyntaxExpressionRepetition:
		markReferencedSyntaxNodes(source, resolution, expressions, expressions.Children(expression)[0], referenced)
	}
}
