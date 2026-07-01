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

		diagnostics = validateRulePropertyRules(where, rule, "Token", source, diagnostics)
		diagnostics = validateTokenConditionPaths(where, diagnostics)

	default:
		if _, ok := resolution.SyntaxNodeIndex(matchedTargetName); !ok {
			diagnostics = append(diagnostics, Diagnostic{
				Message: `Syntax node "` + matchedTargetName + `" is not declared.`,
				Span:    matchedTarget.Span,
			})
		}

		if rule.Match.RelationKind == ast.RuleMatchRelationAdjacentSibling {
			if _, ok := resolution.SyntaxNodeIndex(rule.Match.RelatedTarget.Value(source)); !ok {
				diagnostics = append(diagnostics, Diagnostic{
					Message: `Syntax node "` + rule.Match.RelatedTarget.Value(source) + `" is not declared.`,
					Span:    rule.Match.RelatedTarget.Span,
				})
			}
		}

		if where.Subject.Value(source) != "" && !validSyntaxConditionSubject(where.Subject.Value(source), matchedTargetName, rule.Match.RelationKind) {
			diagnostics = append(diagnostics, Diagnostic{
				Message: `Where clause references an invalid subject for syntax rule "` + matchedTargetName + `".`,
				Span:    where.Subject.Span,
			})
		}

		if where.OtherSubject.Value(source) != "" && !validSyntaxConditionSubject(where.OtherSubject.Value(source), matchedTargetName, rule.Match.RelationKind) {
			diagnostics = append(diagnostics, Diagnostic{
				Message: `Where clause references an invalid comparison subject for syntax rule "` + matchedTargetName + `".`,
				Span:    where.OtherSubject.Span,
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

		diagnostics = validateRulePropertyRules(where, rule, "Syntax node", source, diagnostics)
		diagnostics = validateSyntaxConditionPaths(source, where, rule, resolution, diagnostics)
	}

	if rule.Match.Kind == ast.RuleMatchToken && rule.Match.ScopeKind != ast.RuleMatchScopeNone {
		diagnostics = append(diagnostics, Diagnostic{
			Message: "Only syntax-node rules may use inside/outside constraints.",
			Span:    rule.Match.Span,
		})
	}

	return diagnostics
}

func validateRulePropertyRules(where ast.RuleCondition, rule ast.Rule, subjectLabel string, source string, diagnostics []Diagnostic) []Diagnostic {
	if where.Property.Value(source) == "" {
		return diagnostics
	}

	if !isSupportedProperty(where.Property.Value(source)) {
		diagnostics = append(diagnostics, Diagnostic{
			Message: subjectLabel + ` property "` + where.Property.Value(source) + `" is not declared.`,
			Span:    where.Property.Span,
		})
	}

	if where.OtherProperty.Value(source) != "" && !isSupportedProperty(where.OtherProperty.Value(source)) {
		diagnostics = append(diagnostics, Diagnostic{
			Message: subjectLabel + ` property "` + where.OtherProperty.Value(source) + `" is not declared.`,
			Span:    where.OtherProperty.Span,
		})
	}

	if where.Property.Value(source) == "blankLines" && where.Subject.Value(source) != "gap" {
		diagnostics = append(diagnostics, Diagnostic{
			Message: `Property "blankLines" is only valid on subject "gap".`,
			Span:    where.Property.Span,
		})
	}

	if where.OtherProperty.Value(source) == "blankLines" && where.OtherSubject.Value(source) != "gap" {
		diagnostics = append(diagnostics, Diagnostic{
			Message: `Property "blankLines" is only valid on subject "gap".`,
			Span:    where.OtherProperty.Span,
		})
	}

	if where.Value.Span != (token.Token{}).Span {
		if where.Property.Value(source) == "text" && where.Value.Kind != token.TokenString {
			diagnostics = append(diagnostics, Diagnostic{
				Message: subjectLabel + ` property "text" must be compared with a string literal.`,
				Span:    where.Value.Span,
			})
		}

		if (where.Property.Value(source) == "length" || where.Property.Value(source) == "blankLines") && where.Value.Kind != token.TokenInteger {
			diagnostics = append(diagnostics, Diagnostic{
				Message: subjectLabel + ` property "` + where.Property.Value(source) + `" must be compared with an integer literal.`,
				Span:    where.Value.Span,
			})
		}
	}

	if where.Operator.Kind == token.TokenStartsWith && where.Property.Value(source) != "text" {
		diagnostics = append(diagnostics, Diagnostic{
			Message: `Property "` + where.Property.Value(source) + `" only supports numeric or equality operators.`,
			Span:    where.Operator.Span,
		})
	}

	if where.OtherProperty.Value(source) != "" && propertyValueKind(where.Property.Value(source)) != propertyValueKind(where.OtherProperty.Value(source)) {
		diagnostics = append(diagnostics, Diagnostic{
			Message: "Compared properties must use compatible value types.",
			Span:    where.Span,
		})
	}

	if rule.Match.RelationKind == ast.RuleMatchRelationNone && (where.Subject.Value(source) == "left" || where.OtherSubject.Value(source) == "left" || where.Subject.Value(source) == "gap" || where.OtherSubject.Value(source) == "gap") {
		diagnostics = append(diagnostics, Diagnostic{
			Message: "Single-match rules cannot reference left or gap subjects.",
			Span:    where.Span,
		})
	}

	return diagnostics
}

func validateTokenConditionPaths(where ast.RuleCondition, diagnostics []Diagnostic) []Diagnostic {
	if len(where.Path) > 0 {
		diagnostics = append(diagnostics, Diagnostic{
			Message: "Token rules do not support named syntax-field paths.",
			Span:    where.Path[0].Span,
		})
	}

	if len(where.OtherPath) > 0 {
		diagnostics = append(diagnostics, Diagnostic{
			Message: "Token rules do not support named syntax-field paths.",
			Span:    where.OtherPath[0].Span,
		})
	}

	return diagnostics
}

func validateSyntaxConditionPaths(source string, where ast.RuleCondition, rule ast.Rule, resolution resolver.Resolution, diagnostics []Diagnostic) []Diagnostic {
	diagnostics = validateSyntaxConditionPath(source, where.Subject, where.Path, rule, resolution, diagnostics)
	diagnostics = validateSyntaxConditionPath(source, where.OtherSubject, where.OtherPath, rule, resolution, diagnostics)

	return diagnostics
}

func validateSyntaxConditionPath(source string, subject token.Token, path []token.Token, rule ast.Rule, resolution resolver.Resolution, diagnostics []Diagnostic) []Diagnostic {
	if len(path) == 0 {
		return diagnostics
	}

	if subject.Value(source) == "gap" {
		diagnostics = append(diagnostics, Diagnostic{
			Message: `Subject "gap" does not support named syntax-field paths.`,
			Span:    path[0].Span,
		})

		return diagnostics
	}

	possibleNodeNames := conditionSubjectSyntaxNodes(source, subject.Value(source), rule, resolution)

	if len(possibleNodeNames) == 0 {
		diagnostics = append(diagnostics, Diagnostic{
			Message: `Subject "` + subject.Value(source) + `" does not support named syntax-field paths.`,
			Span:    path[0].Span,
		})

		return diagnostics
	}

	for idx := range path {
		possibleNodeNames = capturedSyntaxTargets(source, resolution, possibleNodeNames, path[idx].Value(source))

		if len(possibleNodeNames) == 0 {
			diagnostics = append(diagnostics, Diagnostic{
				Message: `Named syntax field "` + path[idx].Value(source) + `" is not declared on the selected syntax path.`,
				Span:    path[idx].Span,
			})

			return diagnostics
		}
	}

	return diagnostics
}

func validSyntaxConditionSubject(subject, matchedTarget string, relationKind ast.RuleMatchRelationKind) bool {
	if subject == matchedTarget || subject == "right" {
		return true
	}

	if relationKind == ast.RuleMatchRelationAdjacentSibling && (subject == "left" || subject == "gap") {
		return true
	}

	return false
}

func isSupportedProperty(value string) bool {
	return value == "text" || value == "length" || value == "blankLines"
}

func propertyValueKind(value string) string {
	if value == "text" {
		return "string"
	}

	return "integer"
}

func conditionSubjectSyntaxNodes(source, subject string, rule ast.Rule, resolution resolver.Resolution) map[string]struct{} {
	if subject == "gap" {
		return nil
	}

	nodeNames := make(map[string]struct{}, 1)

	switch subject {
	case "left":
		if rule.Match.RelationKind == ast.RuleMatchRelationAdjacentSibling {
			nodeNames[rule.Match.RelatedTarget.Value(source)] = struct{}{}
		}

	case "right":
		if rule.Match.Kind == ast.RuleMatchNode {
			nodeNames[rule.Match.Target.Value(source)] = struct{}{}
		}

	default:
		if rule.Match.Kind == ast.RuleMatchNode {
			nodeNames[rule.Match.Target.Value(source)] = struct{}{}
		}
	}

	if len(nodeNames) == 0 {
		return nil
	}

	return nodeNames
}

func capturedSyntaxTargets(source string, resolution resolver.Resolution, nodeNames map[string]struct{}, fieldName string) map[string]struct{} {
	targets := make(map[string]struct{})

	for nodeName := range nodeNames {
		nodeIndex, ok := resolution.SyntaxNodeIndex(nodeName)

		if !ok {
			continue
		}

		collectFieldTargets(source, resolution, resolution.SyntaxNodes[nodeIndex].Expression, fieldName, targets)
	}

	if len(targets) == 0 {
		return nil
	}

	return targets
}

func collectFieldTargets(source string, resolution resolver.Resolution, expressionID ast.SyntaxExpressionID, fieldName string, targets map[string]struct{}) {
	expression := resolution.Document.SyntaxExpressions.Node(expressionID)

	if expression.Kind == ast.SyntaxExpressionCapture && expression.Field.Value(source) == fieldName {
		collectProducedSyntaxNodes(source, resolution, resolution.Document.SyntaxExpressions.Children(expression)[0], targets)

		return
	}

	for _, childID := range resolution.Document.SyntaxExpressions.Children(expression) {
		collectFieldTargets(source, resolution, childID, fieldName, targets)
	}
}

func collectProducedSyntaxNodes(source string, resolution resolver.Resolution, expressionID ast.SyntaxExpressionID, targets map[string]struct{}) {
	expression := resolution.Document.SyntaxExpressions.Node(expressionID)

	switch expression.Kind {
	case ast.SyntaxExpressionReference:
		name := expression.Reference.Value(source)

		if _, ok := resolution.SyntaxNodeIndex(name); ok {
			targets[name] = struct{}{}
		}

	case ast.SyntaxExpressionAny:
		return

	default:
		for _, childID := range resolution.Document.SyntaxExpressions.Children(expression) {
			collectProducedSyntaxNodes(source, resolution, childID, targets)
		}
	}
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

	case ast.SyntaxExpressionCapture, ast.SyntaxExpressionGroup, ast.SyntaxExpressionRepetition:
		markReferencedSyntaxNodes(source, resolution, expressions, expressions.Children(expression)[0], referenced)
	}
}
