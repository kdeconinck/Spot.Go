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

func Test_Validate_Rules(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name                     string
		inSource                 string
		wantParseDiagnosticCount int
		wantDiagnostics          []validator.Diagnostic
	}{
		{
			name:     "When rule names are unique, no diagnostic is returned.",
			inSource: `scope { include "**/*.go" } tokens { Identifier = "id" } rules { rule PublicIdentifier { match Identifier report warn at Identifier "x" } rule LongIdentifier { match Identifier report warn at Identifier "x" } }`,
		},
		{
			name:     "When a rule name is declared twice, a diagnostic is returned.",
			inSource: `scope { include "**/*.go" } tokens { Identifier = "id" } rules { rule PublicIdentifier { match Identifier report warn at Identifier "x" } rule PublicIdentifier { match Identifier report warn at Identifier "x" } }`,
			wantDiagnostics: []validator.Diagnostic{
				diagnostic(`Rule "PublicIdentifier" is already declared.`, 143, 159),
			},
		},
		{
			name:     "When a rule name is declared three times, diagnostics are returned for the second and third declarations.",
			inSource: `scope { include "**/*.go" } tokens { Identifier = "id" } rules { rule PublicIdentifier { match Identifier report warn at Identifier "x" } rule PublicIdentifier { match Identifier report warn at Identifier "x" } rule PublicIdentifier { match Identifier report warn at Identifier "x" } }`,
			wantDiagnostics: []validator.Diagnostic{
				diagnostic(`Rule "PublicIdentifier" is already declared.`, 143, 159),
				diagnostic(`Rule "PublicIdentifier" is already declared.`, 216, 232),
			},
		},
		{
			name:                     "When a rule is missing a match statement, a diagnostic is returned.",
			inSource:                 `scope { include "**/*.go" } tokens { Identifier = "id" } rules { rule PublicIdentifier { report warn at Identifier "x" } }`,
			wantParseDiagnosticCount: 2,
			wantDiagnostics: []validator.Diagnostic{
				diagnostic("Rule must contain a match statement.", 89, 95),
			},
		},
		{
			name:                     "When a rule is missing a report statement, a diagnostic is returned.",
			inSource:                 `scope { include "**/*.go" } tokens { Identifier = "id" } rules { rule PublicIdentifier { match Identifier } }`,
			wantParseDiagnosticCount: 5,
			wantDiagnostics: []validator.Diagnostic{
				diagnostic("Rule must contain a report statement.", 106, 107),
			},
		},
		{
			name:     "When a rule matches an undeclared token, a diagnostic is returned.",
			inSource: `scope { include "**/*.go" } tokens { Identifier = "id" } rules { rule PublicIdentifier { match Missing report warn at Missing "x" } }`,
			wantDiagnostics: []validator.Diagnostic{
				diagnostic(`Token "Missing" is not declared.`, 95, 102),
			},
		},
		{
			name:     "When a where clause references a token other than the matched token, a diagnostic is returned.",
			inSource: `scope { include "**/*.go" } tokens { Identifier = "id" Keyword = "kw" } rules { rule PublicIdentifier { match Identifier where Keyword.text == "public" report warn at Identifier "x" } }`,
			wantDiagnostics: []validator.Diagnostic{
				diagnostic(`Where clause must reference matched token "Identifier".`, 127, 134),
			},
		},
		{
			name:     "When a where clause references the text property, no diagnostic is returned.",
			inSource: `scope { include "**/*.go" } tokens { Identifier = "id" } rules { rule PublicIdentifier { match Identifier where Identifier.text == "public" report warn at Identifier "x" } }`,
		},
		{
			name:     "When a text property uses an inequality operator, no diagnostic is returned.",
			inSource: `scope { include "**/*.go" } tokens { Identifier = "id" } rules { rule PublicIdentifier { match Identifier where Identifier.text != "public" report warn at Identifier "x" } }`,
		},
		{
			name:     "When a text property uses an ordering operator, a diagnostic is returned.",
			inSource: `scope { include "**/*.go" } tokens { Identifier = "id" } rules { rule PublicIdentifier { match Identifier where Identifier.text > "public" report warn at Identifier "x" } }`,
			wantDiagnostics: []validator.Diagnostic{
				diagnostic(`Token property "text" only supports equality operators.`, 128, 129),
			},
		},
		{
			name:     "When a where clause references the length property, no diagnostic is returned.",
			inSource: `scope { include "**/*.go" } tokens { Identifier = "id" } rules { rule PublicIdentifier { match Identifier where Identifier.length > 1 report warn at Identifier "x" } }`,
		},
		{
			name:     "When a where clause references an unknown property, a diagnostic is returned.",
			inSource: `scope { include "**/*.go" } tokens { Identifier = "id" } rules { rule PublicIdentifier { match Identifier where Identifier.unknown == "public" report warn at Identifier "x" } }`,
			wantDiagnostics: []validator.Diagnostic{
				diagnostic(`Token property "unknown" is not declared.`, 123, 130),
			},
		},
		{
			name:     "When a text property is compared with an integer literal, a diagnostic is returned.",
			inSource: `scope { include "**/*.go" } tokens { Identifier = "id" } rules { rule PublicIdentifier { match Identifier where Identifier.text == 1 report warn at Identifier "x" } }`,
			wantDiagnostics: []validator.Diagnostic{
				diagnostic(`Token property "text" must be compared with a string literal.`, 131, 132),
			},
		},
		{
			name:     "When a length property is compared with a string literal, a diagnostic is returned.",
			inSource: `scope { include "**/*.go" } tokens { Identifier = "id" } rules { rule PublicIdentifier { match Identifier where Identifier.length > "public" report warn at Identifier "x" } }`,
			wantDiagnostics: []validator.Diagnostic{
				diagnostic(`Token property "length" must be compared with an integer literal.`, 132, 140),
			},
		},
		{
			name:     "When a report target references a token other than the matched token, a diagnostic is returned.",
			inSource: `scope { include "**/*.go" } tokens { Identifier = "id" Keyword = "kw" } rules { rule PublicIdentifier { match Identifier report warn at Keyword "x" } }`,
			wantDiagnostics: []validator.Diagnostic{
				diagnostic(`Report target must reference matched token "Identifier".`, 136, 143),
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
			claim.Equal(t, tc.name, tc.wantParseDiagnosticCount, len(parseDiagnostics), "Parse Diagnostic Count")
			claim.DeepEqual(t, tc.name, tc.wantDiagnostics, gotDiagnostics, "Diagnostic")
		})
	}
}

func Benchmark_Validate_Rules_0(b *testing.B)    { benchmark_Validate_Rules(b, 0) }
func Benchmark_Validate_Rules_1(b *testing.B)    { benchmark_Validate_Rules(b, 1) }
func Benchmark_Validate_Rules_10(b *testing.B)   { benchmark_Validate_Rules(b, 10) }
func Benchmark_Validate_Rules_100(b *testing.B)  { benchmark_Validate_Rules(b, 100) }
func Benchmark_Validate_Rules_1000(b *testing.B) { benchmark_Validate_Rules(b, 1000) }

func benchmark_Validate_Rules(b *testing.B, size int) {
	b.Helper()

	source := rulesDSL(size)
	document, parseDiagnostics := parser.Parse(source)
	claim.Equal(b, "Rules benchmark.", 0, len(parseDiagnostics), "Parse Diagnostic Count")

	for b.Loop() {
		_ = validator.Validate(source, document)
	}
}

func rulesDSL(size int) string {
	var sb strings.Builder

	sb.WriteString("scope { include \"**/*.go\" }\n")
	sb.WriteString("tokens { Identifier = \"id\" }\n")
	sb.WriteString("rules {\n")

	for idx := range size {
		sb.WriteString("    rule Rule")
		sb.WriteString(strconv.Itoa(idx))
		sb.WriteString(" { match Identifier report warn at Identifier \"x\" }\n")
	}

	sb.WriteString("}")

	return sb.String()
}
