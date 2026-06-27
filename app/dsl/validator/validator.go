// =====================================================================================================================
// == LICENSE:                 Copyright (c) 2026 Kevin De Coninck.
// == SPDX-License-Identifier: LicenseRef-PolyForm-Noncommercial-1.0.0
// =====================================================================================================================

// Package validator validates parsed Spot DSL syntax.
package validator

import "github.com/kdeconinck/spot/dsl/token"

// Validate validates parsed DSL syntax and returns semantic diagnostics.
func Validate(document token.Document) []Diagnostic {
	var diagnostics []Diagnostic

	diagnostics = validateScope(document.Scope, diagnostics)
	diagnostics = validateDefinitions(document.Definitions, diagnostics)
	diagnostics = validateTokens(document.Tokens, document.Definitions, diagnostics)
	diagnostics = validateRules(document.Rules, document.Tokens, diagnostics)
	diagnostics = validateNames(document.Definitions, document.Tokens, diagnostics)

	return diagnostics
}
