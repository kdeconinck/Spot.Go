// =====================================================================================================================
// == LICENSE:                 Copyright (c) 2026 Kevin De Coninck.
// == SPDX-License-Identifier: LicenseRef-PolyForm-Noncommercial-1.0.0
// =====================================================================================================================

// Package validator validates parsed Spot DSL syntax.
package validator

import "github.com/kdeconinck/spot/syntax"

func validateTokens(tokens syntax.TokensSection, diagnostics []Diagnostic) []Diagnostic {
	if len(tokens.Tokens) < 2 {
		return diagnostics
	}

	names := map[string]struct{}{}

	for idx := range tokens.Tokens {
		name := tokens.Tokens[idx].Name

		if _, ok := names[name.Text]; ok {
			diagnostics = append(diagnostics, Diagnostic{
				Message: `Token "` + name.Text + `" is already declared.`,
				Span:    name.Span,
			})

			continue
		}

		names[name.Text] = struct{}{}
	}

	return diagnostics
}
