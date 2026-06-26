// =====================================================================================================================
// == LICENSE:                 Copyright (c) 2026 Kevin De Coninck.
// == SPDX-License-Identifier: LicenseRef-PolyForm-Noncommercial-1.0.0
// =====================================================================================================================

// Package validator validates parsed Spot DSL syntax.
package validator

import "github.com/kdeconinck/spot/syntax"

func validateDefinitions(definitions syntax.DefinitionsSection, diagnostics []Diagnostic) []Diagnostic {
	if len(definitions.Definitions) < 2 {
		return diagnostics
	}

	names := map[string]struct{}{}

	for idx := range definitions.Definitions {
		name := definitions.Definitions[idx].Name

		if _, ok := names[name.Text]; ok {
			diagnostics = append(diagnostics, Diagnostic{
				Message: `Definition "` + name.Text + `" is already declared.`,
				Span:    name.Span,
			})

			continue
		}

		names[name.Text] = struct{}{}
	}

	return diagnostics
}
