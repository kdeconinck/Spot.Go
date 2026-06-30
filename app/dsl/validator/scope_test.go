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

func Test_Validate_Scope(t *testing.T) {
	t.Parallel()

	for tcName, tc := range map[string]struct {
		inSource        markedSource
		wantDiagnostics []expectedDiagnostic
	}{
		"When scope contains an include pattern, no diagnostic is returned.": {
			inSource: markedMultilineLiteral(`
				scope {
					include "**/*.go"
				}
				tokens {
					Token = "x"
				}
			`),
			wantDiagnostics: nil,
		},
		"When an include pattern is empty, a diagnostic is returned.": {
			inSource: markedMultilineLiteral(`
				scope {
					include [[""]]
				}
				tokens {
					Token = "x"
				}
			`),
			wantDiagnostics: []expectedDiagnostic{
				expectDiagnostic("Scope pattern must not be empty.", 0),
			},
		},
		"When an exclude pattern is empty, a diagnostic is returned.": {
			inSource: markedMultilineLiteral(`
				scope {
					include "**/*.go"
					exclude [[""]]
				}
				tokens {
					Token = "x"
				}
			`),
			wantDiagnostics: []expectedDiagnostic{
				expectDiagnostic("Scope pattern must not be empty.", 0),
			},
		},
		"When scope contains no include pattern, a diagnostic is returned.": {
			inSource: markedMultilineLiteral(`
				[[scope {
					exclude "vendor/**"
				}]]
				tokens {
					Token = "x"
				}
			`),
			wantDiagnostics: []expectedDiagnostic{
				expectDiagnostic("Scope must contain at least one include.", 0),
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

func Benchmark_Validate_Scope_0(b *testing.B)    { benchmark_Validate_Scope(b, 0) }
func Benchmark_Validate_Scope_1(b *testing.B)    { benchmark_Validate_Scope(b, 1) }
func Benchmark_Validate_Scope_10(b *testing.B)   { benchmark_Validate_Scope(b, 10) }
func Benchmark_Validate_Scope_100(b *testing.B)  { benchmark_Validate_Scope(b, 100) }
func Benchmark_Validate_Scope_1000(b *testing.B) { benchmark_Validate_Scope(b, 1000) }

func benchmark_Validate_Scope(b *testing.B, size int) {
	b.Helper()

	benchmark_Validate(b, scopeHappyPathDSL(size))
}
