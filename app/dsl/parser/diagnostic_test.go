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
	"github.com/kdeconinck/spot/location"
	"github.com/kdeconinck/spot/qa/claim"
)

func Test_Diagnostic_Error(t *testing.T) {
	t.Parallel()

	// Arrange.
	diagnostic := parser.Diagnostic{
		Message: "Expected 'scope', found 'tokens'",
		Span: location.Span{
			Start: 0,
			End:   6,
		},
	}

	// Act.
	got := diagnostic.Error()

	// Assert.
	claim.Equal(t, "When formatting a parser diagnostic as an error, the message includes the source span.", "Expected 'scope', found 'tokens'. [0:6]", got, "Error")
}
