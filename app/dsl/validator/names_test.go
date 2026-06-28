// =====================================================================================================================
// == LICENSE:                 Copyright (c) 2026 Kevin De Coninck.
// == SPDX-License-Identifier: LicenseRef-PolyForm-Noncommercial-1.0.0
// =====================================================================================================================

// Verify the public API of the validator package.
//
// Tests in this package are written against the exported API only.
// This ensures that behavior is tested through the same surface that external consumers would use.
package validator_test

import (
	"testing"

	"github.com/kdeconinck/spot/dsl/parser"
	"github.com/kdeconinck/spot/dsl/validator"
	"github.com/kdeconinck/spot/qa/claim"
)

func Test_Validate_Names(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name            string
		inSource        string
		wantDiagnostics []validator.Diagnostic
	}{
		{
			name:     "When definition and token names are different, no diagnostic is returned.",
			inSource: `scope { include "**/*.go" } definitions { letter = 'a' } tokens { Identifier = letter }`,
		},
		{
			name:     "When a token name conflicts with a definition name, a diagnostic is returned.",
			inSource: `scope { include "**/*.go" } definitions { Identifier = 'a' } tokens { Identifier = "id" }`,
			wantDiagnostics: []validator.Diagnostic{
				diagnostic(`Token "Identifier" conflicts with a definition of the same name.`, 70, 80),
			},
		},
		{
			name:     "When multiple token names conflict with definition names, diagnostics are returned.",
			inSource: `scope { include "**/*.go" } definitions { Identifier = 'a' Keyword = 'k' } tokens { Identifier = "id" Keyword = "kw" }`,
			wantDiagnostics: []validator.Diagnostic{
				diagnostic(`Token "Identifier" conflicts with a definition of the same name.`, 84, 94),
				diagnostic(`Token "Keyword" conflicts with a definition of the same name.`, 102, 109),
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange.
			document, parseDiagnostics := parser.Parse(tc.inSource)

			// Act.
			gotDiagnostics := validator.Validate(tc.inSource, document)

			// Assert.
			claim.Equal(t, tc.name, 0, len(parseDiagnostics), "Parse Diagnostic Count")
			claim.DeepEqual(t, tc.name, tc.wantDiagnostics, gotDiagnostics, "Diagnostic")
		})
	}
}
