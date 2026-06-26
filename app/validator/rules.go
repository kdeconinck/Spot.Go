// =====================================================================================================================
// == LICENSE:                 Copyright (c) 2026 Kevin De Coninck.
// == SPDX-License-Identifier: LicenseRef-PolyForm-Noncommercial-1.0.0
// =====================================================================================================================

// Package validator validates parsed Spot DSL syntax.
package validator

import "github.com/kdeconinck/spot/syntax"

func validateRules(rules syntax.RulesSection, diagnostics []Diagnostic) []Diagnostic {
	if len(rules.Rules) < 2 {
		return diagnostics
	}

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

	return diagnostics
}
