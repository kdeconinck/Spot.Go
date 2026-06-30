// =====================================================================================================================
// == LICENSE:                 Copyright (c) 2026 Kevin De Coninck.
// == SPDX-License-Identifier: LicenseRef-PolyForm-Noncommercial-1.0.0
// =====================================================================================================================

// Verify the public API of the parser package.
//
// Tests in this package are written against the exported API only.
// This ensures that behavior is tested through the same surface that external consumers would use.
package parser_test

import (
	"testing"

	"github.com/kdeconinck/spot/dsl/parser"
	"github.com/kdeconinck/spot/qa/claim"
)

func Test_Parse_Document(t *testing.T) {
	t.Parallel()

	for tcName, tc := range map[string]struct {
		inSource string
		wantErr  string
	}{
		"When parsing a document without a scope section, a parse error is returned.": {
			inSource: "tokens {}",
			wantErr:  "Expected 'scope', found 'tokens'. [0:6]",
		},
		"When parsing a document with trailing tokens, a parse error is returned.": {
			inSource: "scope {} trailing",
			wantErr:  "Expected 'EOF', found 'identifier'. [9:17]",
		},
	} {
		t.Run(tcName, func(t *testing.T) {
			t.Parallel()

			// Act.
			_, gotErr := parser.Parse(tc.inSource)

			// Assert.
			claim.Equal(t, tcName, tc.wantErr, formatParseError(gotErr), "Parse Error")
		})
	}
}
