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
	"strconv"
	"strings"
	"testing"

	"github.com/kdeconinck/spot/parser"
	"github.com/kdeconinck/spot/qa/claim"
	"github.com/kdeconinck/spot/validator"
)

func Test_Validate_Definitions(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name            string
		inSource        string
		wantDiagnostics []validator.Diagnostic
	}{
		{
			name:     "When definition names are unique, no diagnostic is returned.",
			inSource: "scope { include \"**/*.go\" } definitions { letter = 'a' digit = '0' }",
		},
		{
			name:     "When a definition name is declared twice, a diagnostic is returned.",
			inSource: "scope { include \"**/*.go\" } definitions { letter = 'a' letter = 'b' }",
			wantDiagnostics: []validator.Diagnostic{
				diagnostic(`Definition "letter" is already declared.`, 55, 61),
			},
		},
		{
			name:     "When a definition name is declared three times, diagnostics are returned for the second and third declarations.",
			inSource: "scope { include \"**/*.go\" } definitions { letter = 'a' letter = 'b' letter = 'c' }",
			wantDiagnostics: []validator.Diagnostic{
				diagnostic(`Definition "letter" is already declared.`, 55, 61),
				diagnostic(`Definition "letter" is already declared.`, 68, 74),
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange.
			document, parseDiagnostics := parser.Parse(tc.inSource)
			claim.Equal(t, tc.name, 0, len(parseDiagnostics), "Parse Diagnostic Count")

			// Act.
			gotDiagnostics := validator.Validate(document)

			// Assert.
			claim.DeepEqual(t, tc.name, tc.wantDiagnostics, gotDiagnostics, "Diagnostic")
		})
	}
}

func Benchmark_Validate_Definitions_0(b *testing.B)    { benchmark_Validate_Definitions(b, 0) }
func Benchmark_Validate_Definitions_1(b *testing.B)    { benchmark_Validate_Definitions(b, 1) }
func Benchmark_Validate_Definitions_10(b *testing.B)   { benchmark_Validate_Definitions(b, 10) }
func Benchmark_Validate_Definitions_100(b *testing.B)  { benchmark_Validate_Definitions(b, 100) }
func Benchmark_Validate_Definitions_1000(b *testing.B) { benchmark_Validate_Definitions(b, 1000) }

func benchmark_Validate_Definitions(b *testing.B, size int) {
	b.Helper()

	document, parseDiagnostics := parser.Parse(definitionsDSL(size))
	claim.Equal(b, "Definitions benchmark.", 0, len(parseDiagnostics), "Parse Diagnostic Count")

	for b.Loop() {
		_ = validator.Validate(document)
	}
}

func definitionsDSL(size int) string {
	var sb strings.Builder

	sb.WriteString("scope { include \"**/*.go\" }\n")
	sb.WriteString("definitions {\n")

	for idx := 0; idx < size; idx++ {
		sb.WriteString("    definition")
		sb.WriteString(strconv.Itoa(idx))
		sb.WriteString(" = 'a'\n")
	}

	sb.WriteString("}")

	return sb.String()
}
