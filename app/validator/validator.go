// =====================================================================================================================
// == LICENSE:                 Copyright (c) 2026 Kevin De Coninck.
// == SPDX-License-Identifier: LicenseRef-PolyForm-Noncommercial-1.0.0
// =====================================================================================================================

// Package validator validates parsed Spot DSL syntax.
package validator

import "github.com/kdeconinck/spot/syntax"

// Validate validates parsed DSL syntax and returns semantic diagnostics.
func Validate(document syntax.Document) []Diagnostic {
	var diagnostics []Diagnostic

	validateScope(document.Scope, &diagnostics)

	return diagnostics
}
