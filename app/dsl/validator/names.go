// =====================================================================================================================
// == LICENSE:                 Copyright (c) 2026 Kevin De Coninck.
// == SPDX-License-Identifier: LicenseRef-PolyForm-Noncommercial-1.0.0
// =====================================================================================================================

// Package validator validates parsed Spot DSL token.
package validator

import "github.com/kdeconinck/spot/dsl/token"

func validateNames(definitions token.DefinitionsSection, tokens token.TokensSection, diagnostics []Diagnostic) []Diagnostic {
	if len(definitions.Definitions) == 0 || len(tokens.Tokens) == 0 {
		return diagnostics
	}

	definitionNames := map[string]struct{}{}

	for idx := range definitions.Definitions {
		definitionNames[definitions.Definitions[idx].Name.Text] = struct{}{}
	}

	for idx := range tokens.Tokens {
		name := tokens.Tokens[idx].Name

		if _, ok := definitionNames[name.Text]; ok {
			diagnostics = append(diagnostics, Diagnostic{
				Message: `Token "` + name.Text + `" conflicts with a definition of the same name.`,
				Span:    name.Span,
			})
		}
	}

	return diagnostics
}
