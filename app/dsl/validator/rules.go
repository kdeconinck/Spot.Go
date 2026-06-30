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

	if len(rules) > 1 {
		for idx := range rules {
			name := rules[idx].Name
			nameValue := name.Value(source)

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
		diagnostics = validateRuleReferences(source, rules[idx], resolution, diagnostics)
	}

	return diagnostics
}

func validateRuleReferences(source string, rule ast.Rule, resolution resolver.Resolution, diagnostics []Diagnostic) []Diagnostic {
	matchedToken := rule.Match.Token
	matchedTokenName := matchedToken.Value(source)
	where := rule.Where
	report := rule.Report

	if _, ok := resolution.TokenIndex(matchedTokenName); !ok {
		diagnostics = append(diagnostics, Diagnostic{
			Message: `Token "` + matchedTokenName + `" is not declared.`,
			Span:    matchedToken.Span,
		})
	}

	if where.Subject.Value(source) != "" && where.Subject.Value(source) != matchedTokenName {
		diagnostics = append(diagnostics, Diagnostic{
			Message: `Where clause must reference matched token "` + matchedTokenName + `".`,
			Span:    where.Subject.Span,
		})
	}

	if where.Property.Value(source) != "" && where.Property.Value(source) != "text" && where.Property.Value(source) != "length" {
		diagnostics = append(diagnostics, Diagnostic{
			Message: `Token property "` + where.Property.Value(source) + `" is not declared.`,
			Span:    where.Property.Span,
		})
	}

	if where.Property.Value(source) == "text" && where.Operator.Kind != token.TokenEqualEqual && where.Operator.Kind != token.TokenBangEqual {
		diagnostics = append(diagnostics, Diagnostic{
			Message: `Token property "text" only supports equality operators.`,
			Span:    where.Operator.Span,
		})
	}

	if where.Property.Value(source) == "text" && where.Value.Kind != token.TokenString {
		diagnostics = append(diagnostics, Diagnostic{
			Message: `Token property "text" must be compared with a string literal.`,
			Span:    where.Value.Span,
		})
	}

	if where.Property.Value(source) == "length" && where.Value.Kind != token.TokenInteger {
		diagnostics = append(diagnostics, Diagnostic{
			Message: `Token property "length" must be compared with an integer literal.`,
			Span:    where.Value.Span,
		})
	}

	if report.Target.Value(source) != "" && report.Target.Value(source) != matchedTokenName {
		diagnostics = append(diagnostics, Diagnostic{
			Message: `Report target must reference matched token "` + matchedTokenName + `".`,
			Span:    report.Target.Span,
		})
	}

	return diagnostics
}
