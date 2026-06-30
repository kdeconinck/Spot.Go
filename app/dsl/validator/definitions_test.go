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

	"github.com/kdeconinck/spot/dsl/validator"
	"github.com/kdeconinck/spot/qa/claim"
)

func Test_Validate_Definitions(t *testing.T) {
	t.Parallel()

	for tcName, tc := range map[string]struct {
		inSource        markedSource
		wantDiagnostics []expectedDiagnostic
	}{
		"When definition names are unique, no diagnostic is returned.": {
			inSource: markedMultilineLiteral(`
				scope {
					include "**/*.go"
				}
				definitions {
					letter = 'a'
					digit = '0'
				}
				tokens {
					Token = "x"
				}
			`),
			wantDiagnostics: nil,
		},
		"When a definition name is declared twice, a diagnostic is returned.": {
			inSource: markedMultilineLiteral(`
				scope {
					include "**/*.go"
				}
				definitions {
					letter = 'a'
					[[letter]] = 'b'
				}
				tokens {
					Token = "x"
				}
			`),
			wantDiagnostics: []expectedDiagnostic{
				expectDiagnostic(`Definition "letter" is already declared.`, 0),
			},
		},
		"When a definition name is declared three times, diagnostics are returned for the second and third declarations.": {
			inSource: markedMultilineLiteral(`
				scope {
					include "**/*.go"
				}
				definitions {
					letter = 'a'
					[[letter]] = 'b'
					[[letter]] = 'c'
				}
				tokens {
					Token = "x"
				}
			`),
			wantDiagnostics: []expectedDiagnostic{
				expectDiagnostic(`Definition "letter" is already declared.`, 0),
				expectDiagnostic(`Definition "letter" is already declared.`, 1),
			},
		},
		"When a character range start is less than the end, no diagnostic is returned.": {
			inSource: markedMultilineLiteral(`
				scope {
					include "**/*.go"
				}
				definitions {
					letter = 'a'..'z'
				}
				tokens {
					Token = "x"
				}
			`),
			wantDiagnostics: nil,
		},
		"When a character range start is greater than the end, a diagnostic is returned.": {
			inSource: markedMultilineLiteral(`
				scope {
					include "**/*.go"
				}
				definitions {
					letter = [['z'..'a']]
				}
				tokens {
					Token = "x"
				}
			`),
			wantDiagnostics: []expectedDiagnostic{
				expectDiagnostic("Character range start must be less than or equal to end.", 0),
			},
		},
		"When an escaped character range start is greater than the end, a diagnostic is returned.": {
			inSource: markedMultilineLiteral(`
				scope {
					include "**/*.go"
				}
				definitions {
					whitespace = [['\r'..'\n']]
				}
				tokens {
					Token = "x"
				}
			`),
			wantDiagnostics: []expectedDiagnostic{
				expectDiagnostic("Character range start must be less than or equal to end.", 0),
			},
		},
		"When an escaped tab character range start is less than the end, no diagnostic is returned.": {
			inSource: markedMultilineLiteral(`
				scope {
					include "**/*.go"
				}
				definitions {
					whitespace = '\t'..'\n'
				}
				tokens {
					Token = "x"
				}
			`),
			wantDiagnostics: nil,
		},
		"When a character range uses escaped quote and backslash, no diagnostic is returned.": {
			inSource: markedMultilineLiteral(`
				scope {
					include "**/*.go"
				}
				definitions {
					quote = '\''..'\\'
				}
				tokens {
					Token = "x"
				}
			`),
			wantDiagnostics: nil,
		},
		"When a definition references a declared definition, no diagnostic is returned.": {
			inSource: markedMultilineLiteral(`
				scope {
					include "**/*.go"
				}
				definitions {
					letter = 'a'
					identifier = letter
				}
				tokens {
					Token = "x"
				}
			`),
			wantDiagnostics: nil,
		},
		"When a definition references an undeclared definition, a diagnostic is returned.": {
			inSource: markedMultilineLiteral(`
				scope {
					include "**/*.go"
				}
				definitions {
					identifier = [[missing]]
				}
				tokens {
					Token = "x"
				}
			`),
			wantDiagnostics: []expectedDiagnostic{
				expectDiagnostic(`Definition "missing" is not declared.`, 0),
			},
		},
		"When one of multiple definitions references an undeclared definition, a diagnostic is returned.": {
			inSource: markedMultilineLiteral(`
				scope {
					include "**/*.go"
				}
				definitions {
					a = [[missing]]
					b = 'b'
				}
				tokens {
					Token = "x"
				}
			`),
			wantDiagnostics: []expectedDiagnostic{
				expectDiagnostic(`Definition "missing" is not declared.`, 0),
			},
		},
		"When an alternation references undeclared definitions, diagnostics are returned.": {
			inSource: markedMultilineLiteral(`
				scope {
					include "**/*.go"
				}
				definitions {
					value = [[letter]] | [[digit]]
				}
				tokens {
					Token = "x"
				}
			`),
			wantDiagnostics: []expectedDiagnostic{
				expectDiagnostic(`Definition "letter" is not declared.`, 0),
				expectDiagnostic(`Definition "digit" is not declared.`, 1),
			},
		},
		"When grouped repetition references undeclared definitions, diagnostics are returned.": {
			inSource: markedMultilineLiteral(`
				scope {
					include "**/*.go"
				}
				definitions {
					value = [[letter]] ([[digit]] | '_')+
				}
				tokens {
					Token = "x"
				}
			`),
			wantDiagnostics: []expectedDiagnostic{
				expectDiagnostic(`Definition "letter" is not declared.`, 0),
				expectDiagnostic(`Definition "digit" is not declared.`, 1),
			},
		},
		"When concatenation references undeclared definitions, diagnostics are returned.": {
			inSource: markedMultilineLiteral(`
				scope {
					include "**/*.go"
				}
				definitions {
					value = [[letter]] [[digit]] [[missing]]
				}
				tokens {
					Token = "x"
				}
			`),
			wantDiagnostics: []expectedDiagnostic{
				expectDiagnostic(`Definition "letter" is not declared.`, 0),
				expectDiagnostic(`Definition "digit" is not declared.`, 1),
				expectDiagnostic(`Definition "missing" is not declared.`, 2),
			},
		},
		"When a definition directly references itself, a diagnostic is returned.": {
			inSource: markedMultilineLiteral(`
				scope {
					include "**/*.go"
				}
				definitions {
					a = [[a]]
				}
				tokens {
					Token = "x"
				}
			`),
			wantDiagnostics: []expectedDiagnostic{
				expectDiagnostic(`Definition "a" is recursive.`, 0),
			},
		},
		"When definitions reference each other, a diagnostic is returned.": {
			inSource: markedMultilineLiteral(`
				scope {
					include "**/*.go"
				}
				definitions {
					a = b
					b = [[a]]
				}
				tokens {
					Token = "x"
				}
			`),
			wantDiagnostics: []expectedDiagnostic{
				expectDiagnostic(`Definition "a" is recursive.`, 0),
			},
		},
		"When grouped repetition references a recursive definition, a diagnostic is returned.": {
			inSource: markedMultilineLiteral(`
				scope {
					include "**/*.go"
				}
				definitions {
					a = (b | 'x')+
					b = [[a]]
				}
				tokens {
					Token = "x"
				}
			`),
			wantDiagnostics: []expectedDiagnostic{
				expectDiagnostic(`Definition "a" is recursive.`, 0),
			},
		},
	} {
		t.Run(tcName, func(t *testing.T) {
			t.Parallel()

			// Arrange.
			resolution := parseResolution(t, tcName, tc.inSource.text)

			// Act.
			gotDiagnostics := validator.Validate(tc.inSource.text, resolution)

			// Assert.
			claim.DeepEqual(t, tcName, realizeDiagnostics(tc.inSource, tc.wantDiagnostics), gotDiagnostics, "Diagnostic")
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

	benchmark_Validate(b, definitionsHappyPathDSL(size))
}
