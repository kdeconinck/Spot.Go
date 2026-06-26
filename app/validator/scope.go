// =====================================================================================================================
// == LICENSE:                 Copyright (c) 2026 Kevin De Coninck.
// == SPDX-License-Identifier: LicenseRef-PolyForm-Noncommercial-1.0.0
// =====================================================================================================================

// Package validator validates parsed Spot DSL syntax.
package validator

import "github.com/kdeconinck/spot/syntax"

func validateScope(scope syntax.ScopeSection, diagnostics []Diagnostic) []Diagnostic {
	hasInclude := false

	for idx := range scope.Entries {
		entry := scope.Entries[idx]

		if entry.Kind == syntax.ScopeEntryInclude {
			hasInclude = true
		}

		if entry.Pattern.Text == `""` {
			diagnostics = append(diagnostics, Diagnostic{
				Message: "Scope pattern must not be empty.",
				Span:    entry.Pattern.Span,
			})
		}
	}

	if !hasInclude {
		diagnostics = append(diagnostics, Diagnostic{
			Message: "Scope must contain at least one include.",
			Span:    scope.Span,
		})
	}

	return diagnostics
}
