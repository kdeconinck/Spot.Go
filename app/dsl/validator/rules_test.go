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

func Test_Validate_Rules(t *testing.T) {
	t.Parallel()

	for tcName, tc := range map[string]struct {
		inSource        markedSource
		wantDiagnostics []expectedDiagnostic
	}{
		"When rule names are unique, no diagnostic is returned.": {
			inSource: markedMultilineLiteral(`
				scope {
					include "**/*.go"
				}
				tokens {
					Identifier = "id"
				}
				rules {
					rule PublicIdentifier {
						match Identifier
						report warn at Identifier "x"
					}
					rule LongIdentifier {
						match Identifier
						report warn at Identifier "x"
					}
				}
			`),
			wantDiagnostics: nil,
		},
		"When a rule name is declared twice, a diagnostic is returned.": {
			inSource: markedMultilineLiteral(`
				scope {
					include "**/*.go"
				}
				tokens {
					Identifier = "id"
				}
				rules {
					rule PublicIdentifier {
						match Identifier
						report warn at Identifier "x"
					}
					rule [[PublicIdentifier]] {
						match Identifier
						report warn at Identifier "x"
					}
				}
			`),
			wantDiagnostics: []expectedDiagnostic{
				expectDiagnostic(`Rule "PublicIdentifier" is already declared.`, 0),
			},
		},
		"When a rule name is declared three times, diagnostics are returned for the second and third declarations.": {
			inSource: markedMultilineLiteral(`
				scope {
					include "**/*.go"
				}
				tokens {
					Identifier = "id"
				}
				rules {
					rule PublicIdentifier {
						match Identifier
						report warn at Identifier "x"
					}
					rule [[PublicIdentifier]] {
						match Identifier
						report warn at Identifier "x"
					}
					rule [[PublicIdentifier]] {
						match Identifier
						report warn at Identifier "x"
					}
				}
			`),
			wantDiagnostics: []expectedDiagnostic{
				expectDiagnostic(`Rule "PublicIdentifier" is already declared.`, 0),
				expectDiagnostic(`Rule "PublicIdentifier" is already declared.`, 1),
			},
		},
		"When a rule matches an undeclared token, a diagnostic is returned.": {
			inSource: markedMultilineLiteral(`
				scope {
					include "**/*.go"
				}
				tokens {
					Identifier = "id"
				}
				rules {
					rule PublicIdentifier {
						match [[Missing]]
						report warn at Missing "x"
					}
				}
			`),
			wantDiagnostics: []expectedDiagnostic{
				expectDiagnostic(`Token "Missing" is not declared.`, 0),
			},
		},
		"When a where clause references a token other than the matched token, a diagnostic is returned.": {
			inSource: markedMultilineLiteral(`
				scope {
					include "**/*.go"
				}
				tokens {
					Identifier = "id"
					Keyword = "kw"
				}
				rules {
					rule PublicIdentifier {
						match Identifier
						where [[Keyword]].text == "public"
						report warn at Identifier "x"
					}
				}
			`),
			wantDiagnostics: []expectedDiagnostic{
				expectDiagnostic(`Where clause must reference matched token "Identifier".`, 0),
			},
		},
		"When a where clause references the text property, no diagnostic is returned.": {
			inSource: markedMultilineLiteral(`
				scope {
					include "**/*.go"
				}
				tokens {
					Identifier = "id"
				}
				rules {
					rule PublicIdentifier {
						match Identifier
						where Identifier.text == "public"
						report warn at Identifier "x"
					}
				}
			`),
			wantDiagnostics: nil,
		},
		"When a text property uses an inequality operator, no diagnostic is returned.": {
			inSource: markedMultilineLiteral(`
				scope {
					include "**/*.go"
				}
				tokens {
					Identifier = "id"
				}
				rules {
					rule PublicIdentifier {
						match Identifier
						where Identifier.text != "public"
						report warn at Identifier "x"
					}
				}
			`),
			wantDiagnostics: nil,
		},
		"When a text property uses an ordering operator, a diagnostic is returned.": {
			inSource: markedMultilineLiteral(`
				scope {
					include "**/*.go"
				}
				tokens {
					Identifier = "id"
				}
				rules {
					rule PublicIdentifier {
						match Identifier
						where Identifier.text [[>]] "public"
						report warn at Identifier "x"
					}
				}
			`),
			wantDiagnostics: []expectedDiagnostic{
				expectDiagnostic(`Token property "text" only supports equality operators.`, 0),
			},
		},
		"When a where clause references the length property, no diagnostic is returned.": {
			inSource: markedMultilineLiteral(`
				scope {
					include "**/*.go"
				}
				tokens {
					Identifier = "id"
				}
				rules {
					rule PublicIdentifier {
						match Identifier
						where Identifier.length > 1
						report warn at Identifier "x"
					}
				}
			`),
			wantDiagnostics: nil,
		},
		"When a where clause references an unknown property, a diagnostic is returned.": {
			inSource: markedMultilineLiteral(`
				scope {
					include "**/*.go"
				}
				tokens {
					Identifier = "id"
				}
				rules {
					rule PublicIdentifier {
						match Identifier
						where Identifier.[[unknown]] == "public"
						report warn at Identifier "x"
					}
				}
			`),
			wantDiagnostics: []expectedDiagnostic{
				expectDiagnostic(`Token property "unknown" is not declared.`, 0),
			},
		},
		"When a text property is compared with an integer literal, a diagnostic is returned.": {
			inSource: markedMultilineLiteral(`
				scope {
					include "**/*.go"
				}
				tokens {
					Identifier = "id"
				}
				rules {
					rule PublicIdentifier {
						match Identifier
						where Identifier.text == [[1]]
						report warn at Identifier "x"
					}
				}
			`),
			wantDiagnostics: []expectedDiagnostic{
				expectDiagnostic(`Token property "text" must be compared with a string literal.`, 0),
			},
		},
		"When a length property is compared with a string literal, a diagnostic is returned.": {
			inSource: markedMultilineLiteral(`
				scope {
					include "**/*.go"
				}
				tokens {
					Identifier = "id"
				}
				rules {
					rule PublicIdentifier {
						match Identifier
						where Identifier.length > [["public"]]
						report warn at Identifier "x"
					}
				}
			`),
			wantDiagnostics: []expectedDiagnostic{
				expectDiagnostic(`Token property "length" must be compared with an integer literal.`, 0),
			},
		},
		"When a report target references a token other than the matched token, a diagnostic is returned.": {
			inSource: markedMultilineLiteral(`
				scope {
					include "**/*.go"
				}
				tokens {
					Identifier = "id"
					Keyword = "kw"
				}
				rules {
					rule PublicIdentifier {
						match Identifier
						report warn at [[Keyword]] "x"
					}
				}
			`),
			wantDiagnostics: []expectedDiagnostic{
				expectDiagnostic(`Report target must reference matched token "Identifier".`, 0),
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

func Benchmark_Validate_Rules_0(b *testing.B)    { benchmark_Validate_Rules(b, 0) }
func Benchmark_Validate_Rules_1(b *testing.B)    { benchmark_Validate_Rules(b, 1) }
func Benchmark_Validate_Rules_10(b *testing.B)   { benchmark_Validate_Rules(b, 10) }
func Benchmark_Validate_Rules_100(b *testing.B)  { benchmark_Validate_Rules(b, 100) }
func Benchmark_Validate_Rules_1000(b *testing.B) { benchmark_Validate_Rules(b, 1000) }

func benchmark_Validate_Rules(b *testing.B, size int) {
	b.Helper()

	benchmark_Validate(b, rulesHappyPathDSL(size))
}
