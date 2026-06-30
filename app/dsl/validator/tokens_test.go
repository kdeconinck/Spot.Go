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

	"github.com/kdeconinck/spot/dsl/parser"
	"github.com/kdeconinck/spot/dsl/validator"
	"github.com/kdeconinck/spot/qa/claim"
)

func Test_Validate_Tokens(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name            string
		inSource        string
		wantDiagnostics []validator.Diagnostic
	}{
		{
			name:     "When the tokens section is missing, a diagnostic is returned.",
			inSource: `scope { include "**/*.go" }`,
			wantDiagnostics: []validator.Diagnostic{
				diagnostic("Tokens must contain at least one token.", 0, 0),
			},
		},
		{
			name:     "When the tokens section is empty, a diagnostic is returned.",
			inSource: `scope { include "**/*.go" } tokens {}`,
			wantDiagnostics: []validator.Diagnostic{
				diagnostic("Tokens must contain at least one token.", 28, 37),
			},
		},
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
		{
			name:     "When a token references a declared definition, no diagnostic is returned.",
			inSource: `scope { include "**/*.go" } definitions { letter = 'a' } tokens { Identifier = letter }`,
		},
		{
			name:     "When a token references an undeclared definition, a diagnostic is returned.",
			inSource: `scope { include "**/*.go" } tokens { Identifier = missing }`,
			wantDiagnostics: []validator.Diagnostic{
				diagnostic(`Definition "missing" is not declared.`, 50, 57),
			},
		},
		{
			name:     "When a token references an undeclared definition and other definitions exist, a diagnostic is returned.",
			inSource: `scope { include "**/*.go" } definitions { digit = '0' } tokens { Identifier = missing }`,
			wantDiagnostics: []validator.Diagnostic{
				diagnostic(`Definition "missing" is not declared.`, 78, 85),
			},
		},
		{
			name:     "When a token grouped repetition references undeclared definitions, diagnostics are returned.",
			inSource: `scope { include "**/*.go" } tokens { Identifier = (letter | digit)+ }`,
			wantDiagnostics: []validator.Diagnostic{
				diagnostic(`Definition "letter" is not declared.`, 51, 57),
				diagnostic(`Definition "digit" is not declared.`, 60, 65),
			},
		},
		{
			name:     "When a token concatenation references undeclared definitions, diagnostics are returned.",
			inSource: `scope { include "**/*.go" } tokens { Identifier = letter digit missing }`,
			wantDiagnostics: []validator.Diagnostic{
				diagnostic(`Definition "letter" is not declared.`, 50, 56),
				diagnostic(`Definition "digit" is not declared.`, 57, 62),
				diagnostic(`Definition "missing" is not declared.`, 63, 70),
			},
		},
		{
			name:     "When a token expression is a string literal, no definition lookup diagnostic is returned.",
			inSource: `scope { include "**/*.go" } tokens { Keyword = "public" }`,
		},
		{
			name:     "When a token expression uses zero-or-more repetition, a diagnostic is returned.",
			inSource: `scope { include "**/*.go" } definitions { digit = '0'..'9' } tokens { Empty = digit* }`,
			wantDiagnostics: []validator.Diagnostic{
				diagnostic("Token expression must not match empty input.", 78, 84),
			},
		},
		{
			name:     "When a token expression uses zero-or-one repetition, a diagnostic is returned.",
			inSource: `scope { include "**/*.go" } definitions { digit = '0'..'9' } tokens { Optional = digit? }`,
			wantDiagnostics: []validator.Diagnostic{
				diagnostic("Token expression must not match empty input.", 81, 87),
			},
		},
		{
			name:     "When a grouped token expression uses zero-or-one repetition, a diagnostic is returned.",
			inSource: `scope { include "**/*.go" } definitions { digit = '0'..'9' } tokens { Maybe = (digit | '_')? }`,
			wantDiagnostics: []validator.Diagnostic{
				diagnostic("Token expression must not match empty input.", 78, 92),
			},
		},
		{
			name:     "When a token expression concatenates only optional terms, a diagnostic is returned.",
			inSource: `scope { include "**/*.go" } definitions { digit = '0' } tokens { Maybe = digit? digit? }`,
			wantDiagnostics: []validator.Diagnostic{
				diagnostic("Token expression must not match empty input.", 73, 86),
			},
		},
		{
			name:     "When a token expression has an alternative that matches empty input, a diagnostic is returned.",
			inSource: `scope { include "**/*.go" } definitions { digit = '0' letter = 'a' } tokens { Maybe = digit* | letter }`,
			wantDiagnostics: []validator.Diagnostic{
				diagnostic("Token expression must not match empty input.", 86, 101),
			},
		},
		{
			name:     "When a token expression is an empty string literal, a diagnostic is returned.",
			inSource: `scope { include "**/*.go" } tokens { Empty = "" }`,
			wantDiagnostics: []validator.Diagnostic{
				diagnostic("Token expression must not match empty input.", 45, 47),
			},
		},
		{
			name:     "When a token expression uses one-or-more repetition, no empty input diagnostic is returned.",
			inSource: `scope { include "**/*.go" } definitions { digit = '0'..'9' } tokens { Number = digit+ }`,
		},
		{
			name:     "When a token expression concatenates required and optional terms, no empty input diagnostic is returned.",
			inSource: `scope { include "**/*.go" } definitions { letter = 'a' digit = '0' } tokens { Identifier = letter digit* }`,
		},
		{
			name:     "When a token expression references a recursive definition, no empty input diagnostic is returned.",
			inSource: `scope { include "**/*.go" } definitions { a = b b = a } tokens { Value = a }`,
			wantDiagnostics: []validator.Diagnostic{
				diagnostic(`Definition "a" is recursive.`, 52, 53),
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange.
			document, parseErr := parser.Parse(tc.inSource)

			// Act.
			gotDiagnostics := validator.Validate(tc.inSource, document)

			// Assert.
			claim.Equal(t, tc.name, error(nil), parseErr, "Parse Error")
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

	source := tokensDSL(size)
	document, parseErr := parser.Parse(source)
	claim.Equal(b, "Tokens benchmark.", error(nil), parseErr, "Parse Error")

	for b.Loop() {
		_ = validator.Validate(source, document)
	}
}

func tokensDSL(size int) string {
	var sb strings.Builder

	sb.WriteString("scope { include \"**/*.go\" }\n")
	sb.WriteString("definitions { base = 'a' }\n")
	sb.WriteString("tokens {\n")

	for idx := range size {
		sb.WriteString("    Token")
		sb.WriteString(strconv.Itoa(idx))
		sb.WriteString(" = base\n")
	}

	sb.WriteString("}")

	return sb.String()
}
