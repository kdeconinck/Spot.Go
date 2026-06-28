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
	"strings"
	"testing"

	"github.com/kdeconinck/spot/dsl/parser"
	"github.com/kdeconinck/spot/dsl/validator"
	"github.com/kdeconinck/spot/location"
	"github.com/kdeconinck/spot/qa/claim"
)

func Test_Validate_Scope(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name            string
		inSource        string
		wantDiagnostics []validator.Diagnostic
	}{
		{
			name:     "When scope contains an include pattern, no diagnostic is returned.",
			inSource: `scope { include "**/*.go" } tokens { Token = "x" }`,
		},
		{
			name:     "When scope contains no include pattern, a diagnostic is returned.",
			inSource: `scope { exclude "vendor/**" } tokens { Token = "x" }`,
			wantDiagnostics: []validator.Diagnostic{
				diagnostic("Scope must contain at least one include.", 0, 29),
			},
		},
		{
			name:     "When an include pattern is empty, a diagnostic is returned.",
			inSource: `scope { include "" } tokens { Token = "x" }`,
			wantDiagnostics: []validator.Diagnostic{
				diagnostic("Scope pattern must not be empty.", 16, 18),
			},
		},
		{
			name:     "When an exclude pattern is empty, a diagnostic is returned.",
			inSource: `scope { include "**/*.go" exclude "" } tokens { Token = "x" }`,
			wantDiagnostics: []validator.Diagnostic{
				diagnostic("Scope pattern must not be empty.", 34, 36),
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

func Benchmark_Validate_Scope_0(b *testing.B)    { benchmark_Validate_Scope(b, 0) }
func Benchmark_Validate_Scope_1(b *testing.B)    { benchmark_Validate_Scope(b, 1) }
func Benchmark_Validate_Scope_10(b *testing.B)   { benchmark_Validate_Scope(b, 10) }
func Benchmark_Validate_Scope_100(b *testing.B)  { benchmark_Validate_Scope(b, 100) }
func Benchmark_Validate_Scope_1000(b *testing.B) { benchmark_Validate_Scope(b, 1000) }

func benchmark_Validate_Scope(b *testing.B, size int) {
	b.Helper()

	source := scopeDSL(size)
	document, parseDiagnostics := parser.Parse(source)
	claim.Equal(b, "Scope benchmark.", 0, len(parseDiagnostics), "Parse Diagnostic Count")

	for b.Loop() {
		_ = validator.Validate(source, document)
	}
}

func scopeDSL(size int) string {
	return "scope {\n" +
		strings.Repeat("    include \"**/*.go\"\n    exclude \"vendor/**\"\n", size) +
		"}\n" +
		"tokens { Token = \"x\" }"
}

func diagnostic(message string, start, end location.Position) validator.Diagnostic {
	return validator.Diagnostic{
		Message: message,
		Span: location.Span{
			Start: start,
			End:   end,
		},
	}
}
