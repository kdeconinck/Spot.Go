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

func Test_Validate_Syntax(t *testing.T) {
	t.Parallel()

	for tcName, tc := range map[string]struct {
		inSource        markedSource
		wantDiagnostics []expectedDiagnostic
	}{
		"When syntax node names are unique, no diagnostic is returned.": {
			inSource: markedMultilineLiteral(`
				scope {
					include "**/*.go"
				}
				tokens {
					Identifier = "id"
					KeywordPublic = "public"
				}
				syntax {
					node Word = Identifier | KeywordPublic
					node WordPair = Word Word
				}
			`),
			wantDiagnostics: nil,
		},
		"When a syntax node name is declared twice, a diagnostic is returned.": {
			inSource: markedMultilineLiteral(`
				scope {
					include "**/*.go"
				}
				tokens {
					Identifier = "id"
				}
				syntax {
					node Word = Identifier
					node [[Word]] = Identifier
				}
			`),
			wantDiagnostics: []expectedDiagnostic{
				expectDiagnostic(`Syntax node "Word" is already declared.`, 0),
			},
		},
		"When a syntax node name conflicts with a token name, a diagnostic is returned.": {
			inSource: markedMultilineLiteral(`
				scope {
					include "**/*.go"
				}
				tokens {
					Identifier = "id"
					Name = "name"
				}
				syntax {
					node [[Identifier]] = Name
				}
			`),
			wantDiagnostics: []expectedDiagnostic{
				expectDiagnostic(`Syntax node "Identifier" conflicts with token "Identifier".`, 0),
			},
		},
		"When a syntax expression references a declared token, no diagnostic is returned.": {
			inSource: markedMultilineLiteral(`
				scope {
					include "**/*.go"
				}
				tokens {
					Identifier = "id"
				}
				syntax {
					node Word = Identifier
				}
			`),
			wantDiagnostics: nil,
		},
		"When a syntax expression references a declared syntax node, no diagnostic is returned.": {
			inSource: markedMultilineLiteral(`
				scope {
					include "**/*.go"
				}
				tokens {
					Identifier = "id"
				}
				syntax {
					node Word = Identifier
					node WordPair = Word Word
				}
			`),
			wantDiagnostics: nil,
		},
		"When a syntax expression references an undeclared name, a diagnostic is returned.": {
			inSource: markedMultilineLiteral(`
				scope {
					include "**/*.go"
				}
				tokens {
					Identifier = "id"
				}
				syntax {
					node Word = [[Missing]]
				}
			`),
			wantDiagnostics: []expectedDiagnostic{
				expectDiagnostic(`Syntax reference "Missing" is not declared as a token or syntax node.`, 0),
			},
		},
		"When syntax alternation references undeclared names, diagnostics are returned.": {
			inSource: markedMultilineLiteral(`
				scope {
					include "**/*.go"
				}
				tokens {
					Identifier = "id"
				}
				syntax {
					node Word = Identifier | [[Missing]]
					node Pair = [[Other]] Word
				}
			`),
			wantDiagnostics: []expectedDiagnostic{
				expectDiagnostic(`Syntax reference "Missing" is not declared as a token or syntax node.`, 0),
				expectDiagnostic(`Syntax reference "Other" is not declared as a token or syntax node.`, 1),
			},
		},
		"When a syntax repetition uses one-or-more over non-empty input, no diagnostic is returned.": {
			inSource: markedMultilineLiteral(`
				scope {
					include "**/*.go"
				}
				tokens {
					Unknown = fallback
				}
				syntax {
					node UnknownStatement = Unknown+
					node Root = UnknownStatement*
				}
			`),
			wantDiagnostics: nil,
		},
		"When a syntax any expression is declared, no diagnostic is returned.": {
			inSource: markedMultilineLiteral(`
				scope {
					include "**/*.go"
				}
				tokens {
					Identifier = "id"
				}
				syntax {
					node UnknownStatement = any+
				}
			`),
			wantDiagnostics: nil,
		},
		"When a syntax node is recursive, a diagnostic is returned.": {
			inSource: markedMultilineLiteral(`
				scope {
					include "**/*.go"
				}
				tokens {
					Identifier = "id"
				}
				syntax {
					node Word = [[Word]]
				}
			`),
			wantDiagnostics: []expectedDiagnostic{
				expectDiagnostic(`Syntax node "Word" is recursive.`, 0),
			},
		},
		"When syntax recursion is indirect, a diagnostic is returned.": {
			inSource: markedMultilineLiteral(`
				scope {
					include "**/*.go"
				}
				tokens {
					Identifier = "id"
				}
				syntax {
					node Word = Pair
					node Pair = [[Word]]
				}
			`),
			wantDiagnostics: []expectedDiagnostic{
				expectDiagnostic(`Syntax node "Word" is recursive.`, 0),
			},
		},
		"When a syntax repetition can match empty input, a diagnostic is returned.": {
			inSource: markedMultilineLiteral(`
				scope {
					include "**/*.go"
				}
				tokens {
					Identifier = "id"
				}
				syntax {
					node OptionalWord = Identifier?
					node WordList = [[OptionalWord*]]
				}
			`),
			wantDiagnostics: []expectedDiagnostic{
				expectDiagnostic("Syntax repetition expression must not match empty input.", 0),
			},
		},
	} {
		t.Run(tcName, func(t *testing.T) {
			t.Parallel()

			// Arrange.
			resolution := parseResolution(t, tcName, tc.inSource.text)

			// Act.
			got := validator.Validate(tc.inSource.text, resolution)

			// Assert.
			claim.DeepEqual(t, tcName, realizeDiagnostics(tc.inSource, tc.wantDiagnostics), got, "Diagnostics")
		})
	}
}

func Benchmark_Validate_Syntax_0(b *testing.B)    { benchmark_Validate_Syntax(b, 0) }
func Benchmark_Validate_Syntax_1(b *testing.B)    { benchmark_Validate_Syntax(b, 1) }
func Benchmark_Validate_Syntax_10(b *testing.B)   { benchmark_Validate_Syntax(b, 10) }
func Benchmark_Validate_Syntax_100(b *testing.B)  { benchmark_Validate_Syntax(b, 100) }
func Benchmark_Validate_Syntax_1000(b *testing.B) { benchmark_Validate_Syntax(b, 1000) }

func benchmark_Validate_Syntax(b *testing.B, size int) {
	b.Helper()

	benchmark_Validate(b, syntaxHappyPathDSL(size))
}
