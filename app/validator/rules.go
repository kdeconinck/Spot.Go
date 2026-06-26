// =====================================================================================================================
// == LICENSE:                 Copyright (c) 2026 Kevin De Coninck.
// == SPDX-License-Identifier: LicenseRef-PolyForm-Noncommercial-1.0.0
// =====================================================================================================================

// Package validator validates parsed Spot DSL syntax.
package validator

import "github.com/kdeconinck/spot/syntax"

func validateRules(rules syntax.RulesSection, tokens syntax.TokensSection, diagnostics []Diagnostic) []Diagnostic {
	if len(rules.Rules) > 1 {
		names := map[string]struct{}{}

		for idx := range rules.Rules {
			name := rules.Rules[idx].Name

			if _, ok := names[name.Text]; ok {
				diagnostics = append(diagnostics, Diagnostic{
					Message: `Rule "` + name.Text + `" is already declared.`,
					Span:    name.Span,
				})

				continue
			}

			names[name.Text] = struct{}{}
		}
	}

	for idx := range rules.Rules {
		diagnostics = validateRuleReferences(rules.Rules[idx], tokens, diagnostics)
	}

	return diagnostics
}

func validateRuleReferences(rule syntax.Rule, tokens syntax.TokensSection, diagnostics []Diagnostic) []Diagnostic {
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

	if !tokenDeclared(tokens, matchedToken.Text) {
		diagnostics = append(diagnostics, Diagnostic{
			Message: `Token "` + matchedToken.Text + `" is not declared.`,
			Span:    matchedToken.Span,
		})
	}

	if rule.Where.Subject.Text != "" && rule.Where.Subject.Text != matchedToken.Text {
		diagnostics = append(diagnostics, Diagnostic{
			Message: `Where clause must reference matched token "` + matchedToken.Text + `".`,
			Span:    rule.Where.Subject.Span,
		})
	}

	if rule.Where.Property.Text != "" && rule.Where.Property.Text != "text" && rule.Where.Property.Text != "length" {
		diagnostics = append(diagnostics, Diagnostic{
			Message: `Token property "` + rule.Where.Property.Text + `" is not declared.`,
			Span:    rule.Where.Property.Span,
		})
	}

	if rule.Where.Property.Text == "text" &&
		rule.Where.Operator.Kind != syntax.TokenEqualEqual &&
		rule.Where.Operator.Kind != syntax.TokenBangEqual {
		diagnostics = append(diagnostics, Diagnostic{
			Message: `Token property "text" only supports equality operators.`,
			Span:    rule.Where.Operator.Span,
		})
	}

	if rule.Where.Property.Text == "text" && rule.Where.Value.Kind != syntax.TokenString {
		diagnostics = append(diagnostics, Diagnostic{
			Message: `Token property "text" must be compared with a string literal.`,
			Span:    rule.Where.Value.Span,
		})
	}

	if rule.Where.Property.Text == "length" && rule.Where.Value.Kind != syntax.TokenInteger {
		diagnostics = append(diagnostics, Diagnostic{
			Message: `Token property "length" must be compared with an integer literal.`,
			Span:    rule.Where.Value.Span,
		})
	}

	if hasReport && rule.Report.Target.Text != "" && rule.Report.Target.Text != matchedToken.Text {
		diagnostics = append(diagnostics, Diagnostic{
			Message: `Report target must reference matched token "` + matchedToken.Text + `".`,
			Span:    rule.Report.Target.Span,
		})
	}

	return diagnostics
}

func ruleHasMatch(rule syntax.Rule) bool {
	return rule.Match.Span.Start != rule.Match.Token.Span.Start
}

func ruleHasReport(rule syntax.Rule) bool {
	return rule.Report.Span.Start != rule.Report.Severity.Span.Start
}

func tokenDeclared(tokens syntax.TokensSection, name string) bool {
	for idx := range tokens.Tokens {
		if tokens.Tokens[idx].Name.Text == name {
			return true
		}
	}

	return false
}
