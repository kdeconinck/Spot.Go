// =====================================================================================================================
// == LICENSE:                 Copyright (c) 2026 Kevin De Coninck.
// == SPDX-License-Identifier: LicenseRef-PolyForm-Noncommercial-1.0.0
// =====================================================================================================================

// Package validator validates parsed Spot DSL syntax.
package validator

import "github.com/kdeconinck/spot/dsl/ast"

func validateScope(source string, scope ast.ScopeSection, entries []ast.ScopeEntry, diagnostics []Diagnostic) []Diagnostic {
	hasInclude := false

	for idx := range entries {
		entry := entries[idx]

		if entry.Kind == ast.ScopeEntryInclude {
			hasInclude = true
		}

		if entry.Pattern.Value(source) == `""` {
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
