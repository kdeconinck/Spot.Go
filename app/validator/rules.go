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

	if rule.Report.Target.Text != "" && rule.Report.Target.Text != matchedToken.Text {
		diagnostics = append(diagnostics, Diagnostic{
			Message: `Report target must reference matched token "` + matchedToken.Text + `".`,
			Span:    rule.Report.Target.Span,
		})
	}

	return diagnostics
}

func tokenDeclared(tokens syntax.TokensSection, name string) bool {
	for idx := range tokens.Tokens {
		if tokens.Tokens[idx].Name.Text == name {
			return true
		}
	}

	return false
}
