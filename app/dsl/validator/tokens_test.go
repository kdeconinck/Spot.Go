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

func Test_Validate_Tokens(t *testing.T) {
	t.Parallel()

	for tcName, tc := range map[string]struct {
		inSource        markedSource
		wantDiagnostics []expectedDiagnostic
	}{
		"When the tokens section is missing, a diagnostic is returned.": {
			inSource: markedMultilineLiteral(`
				scope {
					include "**/*.go"
				}
			`),
			wantDiagnostics: []expectedDiagnostic{
				expectDiagnostic("Tokens must contain at least one token.", -1),
			},
		},
		"When the tokens section is empty, a diagnostic is returned.": {
			inSource: markedMultilineLiteral(`
				scope {
					include "**/*.go"
				}
				[[tokens {
				}]]
			`),
			wantDiagnostics: []expectedDiagnostic{
				expectDiagnostic("Tokens must contain at least one token.", 0),
			},
		},
		"When token names are unique, no diagnostic is returned.": {
			inSource: markedMultilineLiteral(`
				scope {
					include "**/*.go"
				}
				tokens {
					Identifier = "id"
					Keyword = "kw"
				}
			`),
			wantDiagnostics: nil,
		},
		"When a fallback token is declared, no diagnostic is returned.": {
			inSource: markedMultilineLiteral(`
				scope {
					include "**/*.go"
				}
				tokens {
					Unknown = fallback
				}
			`),
			wantDiagnostics: nil,
		},
		"When multiple fallback tokens are declared, a diagnostic is returned.": {
			inSource: markedMultilineLiteral(`
				scope {
					include "**/*.go"
				}
				tokens {
					Unknown = fallback
					Other = [[fallback]]
				}
			`),
			wantDiagnostics: []expectedDiagnostic{
				expectDiagnostic("Tokens may contain at most one fallback token.", 0),
			},
		},
		"When a token name conflicts with a definition name, a diagnostic is returned.": {
			inSource: markedMultilineLiteral(`
				scope {
					include "**/*.go"
				}
				definitions {
					Identifier = 'a'
				}
				tokens {
					[[Identifier]] = "id"
				}
			`),
			wantDiagnostics: []expectedDiagnostic{
				expectDiagnostic(`Token "Identifier" conflicts with a definition of the same name.`, 0),
			},
		},
		"When multiple token names conflict with definition names, diagnostics are returned.": {
			inSource: markedMultilineLiteral(`
				scope {
					include "**/*.go"
				}
				definitions {
					Identifier = 'a'
					Keyword = 'k'
				}
				tokens {
					[[Identifier]] = "id"
					[[Keyword]] = "kw"
				}
			`),
			wantDiagnostics: []expectedDiagnostic{
				expectDiagnostic(`Token "Identifier" conflicts with a definition of the same name.`, 0),
				expectDiagnostic(`Token "Keyword" conflicts with a definition of the same name.`, 1),
			},
		},
		"When a token name is declared twice, a diagnostic is returned.": {
			inSource: markedMultilineLiteral(`
				scope {
					include "**/*.go"
				}
				tokens {
					Identifier = "id"
					[[Identifier]] = "other"
				}
			`),
			wantDiagnostics: []expectedDiagnostic{
				expectDiagnostic(`Token "Identifier" is already declared.`, 0),
			},
		},
		"When a token name is declared three times, diagnostics are returned for the second and third declarations.": {
			inSource: markedMultilineLiteral(`
				scope {
					include "**/*.go"
				}
				tokens {
					Identifier = "id"
					[[Identifier]] = "other"
					[[Identifier]] = "again"
				}
			`),
			wantDiagnostics: []expectedDiagnostic{
				expectDiagnostic(`Token "Identifier" is already declared.`, 0),
				expectDiagnostic(`Token "Identifier" is already declared.`, 1),
			},
		},
		"When definition and token names are different, no diagnostic is returned.": {
			inSource: markedMultilineLiteral(`
				scope {
					include "**/*.go"
				}
				definitions {
					letter = 'a'
				}
				tokens {
					Identifier = letter
				}
			`),
			wantDiagnostics: nil,
		},
		"When a token references a declared definition, no diagnostic is returned.": {
			inSource: markedMultilineLiteral(`
				scope {
					include "**/*.go"
				}
				definitions {
					letter = 'a'
				}
				tokens {
					Identifier = letter
				}
			`),
			wantDiagnostics: nil,
		},
		"When a token references an undeclared definition, a diagnostic is returned.": {
			inSource: markedMultilineLiteral(`
				scope {
					include "**/*.go"
				}
				tokens {
					Identifier = [[missing]]
				}
			`),
			wantDiagnostics: []expectedDiagnostic{
				expectDiagnostic(`Definition "missing" is not declared.`, 0),
			},
		},
		"When a token references an undeclared definition and other definitions exist, a diagnostic is returned.": {
			inSource: markedMultilineLiteral(`
				scope {
					include "**/*.go"
				}
				definitions {
					digit = '0'
				}
				tokens {
					Identifier = [[missing]]
				}
			`),
			wantDiagnostics: []expectedDiagnostic{
				expectDiagnostic(`Definition "missing" is not declared.`, 0),
			},
		},
		"When a token grouped repetition references undeclared definitions, diagnostics are returned.": {
			inSource: markedMultilineLiteral(`
				scope {
					include "**/*.go"
				}
				tokens {
					Identifier = ([[letter]] | [[digit]])+
				}
			`),
			wantDiagnostics: []expectedDiagnostic{
				expectDiagnostic(`Definition "letter" is not declared.`, 0),
				expectDiagnostic(`Definition "digit" is not declared.`, 1),
			},
		},
		"When a token concatenation references undeclared definitions, diagnostics are returned.": {
			inSource: markedMultilineLiteral(`
				scope {
					include "**/*.go"
				}
				tokens {
					Identifier = [[letter]] [[digit]] [[missing]]
				}
			`),
			wantDiagnostics: []expectedDiagnostic{
				expectDiagnostic(`Definition "letter" is not declared.`, 0),
				expectDiagnostic(`Definition "digit" is not declared.`, 1),
				expectDiagnostic(`Definition "missing" is not declared.`, 2),
			},
		},
		"When a token expression is a string literal, no definition lookup diagnostic is returned.": {
			inSource: markedMultilineLiteral(`
				scope {
					include "**/*.go"
				}
				tokens {
					Keyword = "public"
				}
			`),
			wantDiagnostics: nil,
		},
		"When a token expression uses zero-or-more repetition, a diagnostic is returned.": {
			inSource: markedMultilineLiteral(`
				scope {
					include "**/*.go"
				}
				definitions {
					digit = '0'..'9'
				}
				tokens {
					Empty = [[digit*]]
				}
			`),
			wantDiagnostics: []expectedDiagnostic{
				expectDiagnostic("Token expression must not match empty input.", 0),
			},
		},
		"When a token expression uses zero-or-one repetition, a diagnostic is returned.": {
			inSource: markedMultilineLiteral(`
				scope {
					include "**/*.go"
				}
				definitions {
					digit = '0'..'9'
				}
				tokens {
					Optional = [[digit?]]
				}
			`),
			wantDiagnostics: []expectedDiagnostic{
				expectDiagnostic("Token expression must not match empty input.", 0),
			},
		},
		"When a grouped token expression uses zero-or-one repetition, a diagnostic is returned.": {
			inSource: markedMultilineLiteral(`
				scope {
					include "**/*.go"
				}
				definitions {
					digit = '0'..'9'
				}
				tokens {
					Maybe = [[(digit | '_')?]]
				}
			`),
			wantDiagnostics: []expectedDiagnostic{
				expectDiagnostic("Token expression must not match empty input.", 0),
			},
		},
		"When a token expression concatenates only optional terms, a diagnostic is returned.": {
			inSource: markedMultilineLiteral(`
				scope {
					include "**/*.go"
				}
				definitions {
					digit = '0'
				}
				tokens {
					Maybe = [[digit? digit?]]
				}
			`),
			wantDiagnostics: []expectedDiagnostic{
				expectDiagnostic("Token expression must not match empty input.", 0),
			},
		},
		"When a token expression has an alternative that matches empty input, a diagnostic is returned.": {
			inSource: markedMultilineLiteral(`
				scope {
					include "**/*.go"
				}
				definitions {
					digit = '0'
					letter = 'a'
				}
				tokens {
					Maybe = [[digit* | letter]]
				}
			`),
			wantDiagnostics: []expectedDiagnostic{
				expectDiagnostic("Token expression must not match empty input.", 0),
			},
		},
		"When a token expression is an empty string literal, a diagnostic is returned.": {
			inSource: markedMultilineLiteral(`
				scope {
					include "**/*.go"
				}
				tokens {
					Empty = [[""]]
				}
			`),
			wantDiagnostics: []expectedDiagnostic{
				expectDiagnostic("Token expression must not match empty input.", 0),
			},
		},
		"When a token expression uses one-or-more repetition, no empty input diagnostic is returned.": {
			inSource: markedMultilineLiteral(`
				scope {
					include "**/*.go"
				}
				definitions {
					digit = '0'..'9'
				}
				tokens {
					Number = digit+
				}
			`),
			wantDiagnostics: nil,
		},
		"When a token expression concatenates required and optional terms, no empty input diagnostic is returned.": {
			inSource: markedMultilineLiteral(`
				scope {
					include "**/*.go"
				}
				definitions {
					letter = 'a'
					digit = '0'
				}
				tokens {
					Identifier = letter digit*
				}
			`),
			wantDiagnostics: nil,
		},
		"When a token expression references a recursive definition, no empty input diagnostic is returned.": {
			inSource: markedMultilineLiteral(`
				scope {
					include "**/*.go"
				}
				definitions {
					a = b
					b = [[a]]
				}
				tokens {
					Value = a
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

func Benchmark_Validate_Tokens_0(b *testing.B)    { benchmark_Validate_Tokens(b, 0) }
func Benchmark_Validate_Tokens_1(b *testing.B)    { benchmark_Validate_Tokens(b, 1) }
func Benchmark_Validate_Tokens_10(b *testing.B)   { benchmark_Validate_Tokens(b, 10) }
func Benchmark_Validate_Tokens_100(b *testing.B)  { benchmark_Validate_Tokens(b, 100) }
func Benchmark_Validate_Tokens_1000(b *testing.B) { benchmark_Validate_Tokens(b, 1000) }

func benchmark_Validate_Tokens(b *testing.B, size int) {
	b.Helper()

	benchmark_Validate(b, tokensHappyPathDSL(size))
}
