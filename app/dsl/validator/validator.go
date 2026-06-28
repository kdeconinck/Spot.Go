// =====================================================================================================================
// == LICENSE:                 Copyright (c) 2026 Kevin De Coninck.
// == SPDX-License-Identifier: LicenseRef-PolyForm-Noncommercial-1.0.0
// =====================================================================================================================

// Package validator validates parsed Spot DSL syntax.
package validator

import "github.com/kdeconinck/spot/dsl/ast"

// Validate validates parsed DSL syntax and returns semantic diagnostics.
func Validate(source string, document ast.Document) []Diagnostic {
	var diagnostics []Diagnostic

	diagnostics = validateScope(source, document.Scope, diagnostics)
	diagnostics = validateDefinitions(source, document.Definitions, diagnostics)
	diagnostics = validateTokens(source, document.Tokens, document.Definitions, diagnostics)
	diagnostics = validateRules(source, document.Rules, document.Tokens, diagnostics)

	return diagnostics
}
