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

func validateRules(source string, rules ast.RulesSection, tokens ast.TokensSection, diagnostics []Diagnostic) []Diagnostic {
	if len(rules.Rules) > 1 {
		names := map[string]struct{}{}

		for idx := range rules.Rules {
			name := rules.Rules[idx].Name

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

	for idx := range rules.Rules {
		diagnostics = validateRuleReferences(source, rules.Rules[idx], tokens, diagnostics)
	}

	return diagnostics
}

func validateRuleReferences(source string, rule ast.Rule, tokens ast.TokensSection, diagnostics []Diagnostic) []Diagnostic {
	hasMatch := ruleHasMatch(rule)
	hasReport := ruleHasReport(rule)

	if !hasMatch {
		diagnostics = append(diagnostics, Diagnostic{
			Message: "Rule must contain a match statement.",
			Span:    rule.Match.Span,
		})
	}

	if !hasReport {
		diagnostics = append(diagnostics, Diagnostic{
			Message: "Rule must contain a report statement.",
			Span:    rule.Report.Span,
		})
	}

	if !hasMatch {
		return diagnostics
	}

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

	if hasReport && rule.Report.Target.Value(source) != "" && rule.Report.Target.Value(source) != matchedToken.Value(source) {
		diagnostics = append(diagnostics, Diagnostic{
			Message: `Report target must reference matched token "` + matchedToken.Value(source) + `".`,
			Span:    rule.Report.Target.Span,
		})
	}

	return diagnostics
}

func ruleHasMatch(rule ast.Rule) bool {
	return rule.Match.Token.Kind == token.TokenIdentifier
}

func ruleHasReport(rule ast.Rule) bool {
	return rule.Report.Span.Start != rule.Report.Severity.Span.Start
}

func tokenDeclared(source string, tokens ast.TokensSection, name string) bool {
	for idx := range tokens.Tokens {
		if tokens.Tokens[idx].Name.Value(source) == name {
			return true
		}
	}

	return false
}
