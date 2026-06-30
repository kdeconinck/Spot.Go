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

func validateRules(source string, rules ast.RulesSection, ruleList []ast.Rule, tokens []ast.TokenDefinition, diagnostics []Diagnostic) []Diagnostic {
	if len(ruleList) > 1 {
		names := map[string]struct{}{}

		for idx := range ruleList {
			name := ruleList[idx].Name

			if _, ok := names[name.Value(source)]; ok {
				diagnostics = append(diagnostics, Diagnostic{
					Message: `Rule "` + name.Value(source) + `" is already declared.`,
					Span:    name.Span,
				})

				continue
			}

			names[name.Value(source)] = struct{}{}
		}
	}

	for idx := range ruleList {
		diagnostics = validateRuleReferences(source, ruleList[idx], tokens, diagnostics)
	}

	return diagnostics
}

func validateRuleReferences(source string, rule ast.Rule, tokens []ast.TokenDefinition, diagnostics []Diagnostic) []Diagnostic {
	matchedToken := rule.Match.Token

	if !tokenDeclared(source, tokens, matchedToken.Value(source)) {
		diagnostics = append(diagnostics, Diagnostic{
			Message: `Token "` + matchedToken.Value(source) + `" is not declared.`,
			Span:    matchedToken.Span,
		})
	}

	if rule.Where.Subject.Value(source) != "" && rule.Where.Subject.Value(source) != matchedToken.Value(source) {
		diagnostics = append(diagnostics, Diagnostic{
			Message: `Where clause must reference matched token "` + matchedToken.Value(source) + `".`,
			Span:    rule.Where.Subject.Span,
		})
	}

	if rule.Where.Property.Value(source) != "" && rule.Where.Property.Value(source) != "text" && rule.Where.Property.Value(source) != "length" {
		diagnostics = append(diagnostics, Diagnostic{
			Message: `Token property "` + rule.Where.Property.Value(source) + `" is not declared.`,
			Span:    rule.Where.Property.Span,
		})
	}

	if rule.Where.Property.Value(source) == "text" &&
		rule.Where.Operator.Kind != token.TokenEqualEqual &&
		rule.Where.Operator.Kind != token.TokenBangEqual {
		diagnostics = append(diagnostics, Diagnostic{
			Message: `Token property "text" only supports equality operators.`,
			Span:    rule.Where.Operator.Span,
		})
	}

	if rule.Where.Property.Value(source) == "text" && rule.Where.Value.Kind != token.TokenString {
		diagnostics = append(diagnostics, Diagnostic{
			Message: `Token property "text" must be compared with a string literal.`,
			Span:    rule.Where.Value.Span,
		})
	}

	if rule.Where.Property.Value(source) == "length" && rule.Where.Value.Kind != token.TokenInteger {
		diagnostics = append(diagnostics, Diagnostic{
			Message: `Token property "length" must be compared with an integer literal.`,
			Span:    rule.Where.Value.Span,
		})
	}

	if rule.Report.Target.Value(source) != "" && rule.Report.Target.Value(source) != matchedToken.Value(source) {
		diagnostics = append(diagnostics, Diagnostic{
			Message: `Report target must reference matched token "` + matchedToken.Value(source) + `".`,
			Span:    rule.Report.Target.Span,
		})
	}

	return diagnostics
}

func tokenDeclared(source string, tokens []ast.TokenDefinition, name string) bool {
	for idx := range tokens {
		if tokens[idx].Name.Value(source) == name {
			return true
		}
	}

	return false
}
