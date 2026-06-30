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

func Test_Validate_Definitions(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name            string
		inSource        string
		wantDiagnostics []validator.Diagnostic
	}{
		{
			name:     "When definition names are unique, no diagnostic is returned.",
			inSource: "scope { include \"**/*.go\" } definitions { letter = 'a' digit = '0' } tokens { Token = \"x\" }",
		},
		{
			name:     "When a definition name is declared twice, a diagnostic is returned.",
			inSource: "scope { include \"**/*.go\" } definitions { letter = 'a' letter = 'b' } tokens { Token = \"x\" }",
			wantDiagnostics: []validator.Diagnostic{
				diagnostic(`Definition "letter" is already declared.`, 55, 61),
			},
		},
		{
			name:     "When a definition name is declared three times, diagnostics are returned for the second and third declarations.",
			inSource: "scope { include \"**/*.go\" } definitions { letter = 'a' letter = 'b' letter = 'c' } tokens { Token = \"x\" }",
			wantDiagnostics: []validator.Diagnostic{
				diagnostic(`Definition "letter" is already declared.`, 55, 61),
				diagnostic(`Definition "letter" is already declared.`, 68, 74),
			},
		},
		{
			name:     "When a definition references a declared definition, no diagnostic is returned.",
			inSource: "scope { include \"**/*.go\" } definitions { letter = 'a' identifier = letter } tokens { Token = \"x\" }",
		},
		{
			name:     "When a definition references an undeclared definition, a diagnostic is returned.",
			inSource: "scope { include \"**/*.go\" } definitions { identifier = missing } tokens { Token = \"x\" }",
			wantDiagnostics: []validator.Diagnostic{
				diagnostic(`Definition "missing" is not declared.`, 55, 62),
			},
		},
		{
			name:     "When one of multiple definitions references an undeclared definition, a diagnostic is returned.",
			inSource: "scope { include \"**/*.go\" } definitions { a = missing b = 'b' } tokens { Token = \"x\" }",
			wantDiagnostics: []validator.Diagnostic{
				diagnostic(`Definition "missing" is not declared.`, 46, 53),
			},
		},
		{
			name:     "When an alternation references undeclared definitions, diagnostics are returned.",
			inSource: "scope { include \"**/*.go\" } definitions { value = letter | digit } tokens { Token = \"x\" }",
			wantDiagnostics: []validator.Diagnostic{
				diagnostic(`Definition "letter" is not declared.`, 50, 56),
				diagnostic(`Definition "digit" is not declared.`, 59, 64),
			},
		},
		{
			name:     "When grouped repetition references undeclared definitions, diagnostics are returned.",
			inSource: "scope { include \"**/*.go\" } definitions { value = letter (digit | '_')+ } tokens { Token = \"x\" }",
			wantDiagnostics: []validator.Diagnostic{
				diagnostic(`Definition "letter" is not declared.`, 50, 56),
				diagnostic(`Definition "digit" is not declared.`, 58, 63),
			},
		},
		{
			name:     "When concatenation references undeclared definitions, diagnostics are returned.",
			inSource: "scope { include \"**/*.go\" } definitions { value = letter digit missing } tokens { Token = \"x\" }",
			wantDiagnostics: []validator.Diagnostic{
				diagnostic(`Definition "letter" is not declared.`, 50, 56),
				diagnostic(`Definition "digit" is not declared.`, 57, 62),
				diagnostic(`Definition "missing" is not declared.`, 63, 70),
			},
		},
		{
			name:     "When a definition directly references itself, a diagnostic is returned.",
			inSource: "scope { include \"**/*.go\" } definitions { a = a } tokens { Token = \"x\" }",
			wantDiagnostics: []validator.Diagnostic{
				diagnostic(`Definition "a" is recursive.`, 46, 47),
			},
		},
		{
			name:     "When definitions reference each other, a diagnostic is returned.",
			inSource: "scope { include \"**/*.go\" } definitions { a = b b = a } tokens { Token = \"x\" }",
			wantDiagnostics: []validator.Diagnostic{
				diagnostic(`Definition "a" is recursive.`, 52, 53),
			},
		},
		{
			name:     "When grouped repetition references a recursive definition, a diagnostic is returned.",
			inSource: "scope { include \"**/*.go\" } definitions { a = (b | 'x')+ b = a } tokens { Token = \"x\" }",
			wantDiagnostics: []validator.Diagnostic{
				diagnostic(`Definition "a" is recursive.`, 61, 62),
			},
		},
		{
			name:     "When a character range start is less than the end, no diagnostic is returned.",
			inSource: "scope { include \"**/*.go\" } definitions { letter = 'a'..'z' } tokens { Token = \"x\" }",
		},
		{
			name:     "When a character range start is greater than the end, a diagnostic is returned.",
			inSource: "scope { include \"**/*.go\" } definitions { letter = 'z'..'a' } tokens { Token = \"x\" }",
			wantDiagnostics: []validator.Diagnostic{
				diagnostic("Character range start must be less than or equal to end.", 51, 59),
			},
		},
		{
			name:     "When an escaped character range start is greater than the end, a diagnostic is returned.",
			inSource: "scope { include \"**/*.go\" } definitions { whitespace = '\\r'..'\\n' } tokens { Token = \"x\" }",
			wantDiagnostics: []validator.Diagnostic{
				diagnostic("Character range start must be less than or equal to end.", 55, 65),
			},
		},
		{
			name:     "When an escaped tab character range start is less than the end, no diagnostic is returned.",
			inSource: "scope { include \"**/*.go\" } definitions { whitespace = '\\t'..'\\n' } tokens { Token = \"x\" }",
		},
		{
			name:     "When a character range uses escaped quote and backslash, no diagnostic is returned.",
			inSource: "scope { include \"**/*.go\" } definitions { quote = '\\''..'\\\\' } tokens { Token = \"x\" }",
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

func Benchmark_Validate_Definitions_0(b *testing.B)    { benchmark_Validate_Definitions(b, 0) }
func Benchmark_Validate_Definitions_1(b *testing.B)    { benchmark_Validate_Definitions(b, 1) }
func Benchmark_Validate_Definitions_10(b *testing.B)   { benchmark_Validate_Definitions(b, 10) }
func Benchmark_Validate_Definitions_100(b *testing.B)  { benchmark_Validate_Definitions(b, 100) }
func Benchmark_Validate_Definitions_1000(b *testing.B) { benchmark_Validate_Definitions(b, 1000) }

func benchmark_Validate_Definitions(b *testing.B, size int) {
	b.Helper()

	source := definitionsDSL(size)
	document, parseErr := parser.Parse(source)
	claim.Equal(b, "Definitions benchmark.", error(nil), parseErr, "Parse Error")

	for b.Loop() {
		_ = validator.Validate(source, document)
	}
}

func definitionsDSL(size int) string {
	var sb strings.Builder

	sb.WriteString("scope { include \"**/*.go\" }\n")
	sb.WriteString("definitions {\n")
	sb.WriteString("    base = 'a'..'z'\n")

	for idx := range size {
		sb.WriteString("    definition")
		sb.WriteString(strconv.Itoa(idx))
		sb.WriteString(" = base\n")
	}

	sb.WriteString("}")
	sb.WriteString("\n")
	sb.WriteString("tokens { Token = \"x\" }")

	return sb.String()
}
