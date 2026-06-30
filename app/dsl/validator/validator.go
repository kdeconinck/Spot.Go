// =====================================================================================================================
// == LICENSE:                 Copyright (c) 2026 Kevin De Coninck.
// == SPDX-License-Identifier: LicenseRef-PolyForm-Noncommercial-1.0.0
// =====================================================================================================================

// Package validator validates parsed Spot DSL syntax.
package validator

import "github.com/kdeconinck/spot/dsl/resolver"

// Validate validates resolved DSL syntax and returns semantic diagnostics.
func Validate(source string, resolution resolver.Resolution) []Diagnostic {
	var diagnostics []Diagnostic

	diagnostics = validateScope(source, resolution.Document.Scope, resolution.ScopeEntries, diagnostics)
	diagnostics = validateDefinitions(source, resolution, diagnostics)
	diagnostics = validateTokens(source, resolution, diagnostics)
	diagnostics = validateRules(source, resolution, diagnostics)

	return diagnostics
}
