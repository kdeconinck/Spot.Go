// =====================================================================================================================
// == LICENSE:                 Copyright (c) 2026 Kevin De Coninck.
// == SPDX-License-Identifier: LicenseRef-PolyForm-Noncommercial-1.0.0
// =====================================================================================================================

package validator_test

import (
	"strconv"
	"strings"
	"testing"

	"github.com/kdeconinck/spot/parser"
	"github.com/kdeconinck/spot/qa/claim"
	"github.com/kdeconinck/spot/validator"
)

func Test_Validate_Tokens(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name            string
		inSource        string
		wantDiagnostics []validator.Diagnostic
	}{
		{
			name:     "When token names are unique, no diagnostic is returned.",
			inSource: `scope { include "**/*.go" } tokens { Identifier = "id" Keyword = "kw" }`,
		},
		{
			name:     "When a token name is declared twice, a diagnostic is returned.",
			inSource: `scope { include "**/*.go" } tokens { Identifier = "id" Identifier = "other" }`,
			wantDiagnostics: []validator.Diagnostic{
				diagnostic(`Token "Identifier" is already declared.`, 55, 65),
			},
		},
		{
			name:     "When a token name is declared three times, diagnostics are returned for the second and third declarations.",
			inSource: `scope { include "**/*.go" } tokens { Identifier = "id" Identifier = "other" Identifier = "again" }`,
			wantDiagnostics: []validator.Diagnostic{
				diagnostic(`Token "Identifier" is already declared.`, 55, 65),
				diagnostic(`Token "Identifier" is already declared.`, 76, 86),
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange.
			document, parseDiagnostics := parser.Parse(tc.inSource)

			// Act.
			gotDiagnostics := validator.Validate(document)

			// Assert.
			claim.Equal(t, tc.name, 0, len(parseDiagnostics), "Parse Diagnostic Count")
			claim.DeepEqual(t, tc.name, tc.wantDiagnostics, gotDiagnostics, "Diagnostic")
		})
	}
}

func Benchmark_Validate_Tokens_0(b *testing.B)    { benchmark_Validate_Tokens(b, 0) }
func Benchmark_Validate_Tokens_1(b *testing.B)    { benchmark_Validate_Tokens(b, 1) }
func Benchmark_Validate_Tokens_10(b *testing.B)   { benchmark_Validate_Tokens(b, 10) }
func Benchmark_Validate_Tokens_100(b *testing.B)  { benchmark_Validate_Tokens(b, 100) }
func Benchmark_Validate_Tokens_1000(b *testing.B) { benchmark_Validate_Tokens(b, 1000) }

func benchmark_Validate_Tokens(b *testing.B, size int) {
	b.Helper()

	document, parseDiagnostics := parser.Parse(tokensDSL(size))
	claim.Equal(b, "Tokens benchmark.", 0, len(parseDiagnostics), "Parse Diagnostic Count")

	for b.Loop() {
		_ = validator.Validate(document)
	}
}

func tokensDSL(size int) string {
	var sb strings.Builder

	sb.WriteString("scope { include \"**/*.go\" }\n")
	sb.WriteString("tokens {\n")

	for idx := range size {
		sb.WriteString("    Token")
		sb.WriteString(strconv.Itoa(idx))
		sb.WriteString(" = \"token")
		sb.WriteString(strconv.Itoa(idx))
		sb.WriteString("\"\n")
	}

	sb.WriteString("}")

	return sb.String()
}
