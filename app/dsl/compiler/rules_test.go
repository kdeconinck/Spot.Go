// =====================================================================================================================
// == LICENSE:                 Copyright (c) 2026 Kevin De Coninck.
// == SPDX-License-Identifier: LicenseRef-PolyForm-Noncommercial-1.0.0
// =====================================================================================================================

// Verify the public API of the compiler package.
//
// Tests in this package are written against the exported API only.
// This ensures that behavior is tested through the same surface that external consumers would use.
package compiler_test

import (
	"testing"

	"github.com/kdeconinck/spot/dsl/compiler"
	"github.com/kdeconinck/spot/dsl/parser"
	"github.com/kdeconinck/spot/dsl/validator"
	"github.com/kdeconinck/spot/qa/claim"
	"github.com/kdeconinck/spot/runtime/ir"
)

func Test_Compile_Rules_PreservesSourceOrder(t *testing.T) {
	t.Parallel()

	// Arrange.
	source := `scope { include "**/*.go" } tokens { Identifier = "id" Number = ('0'..'9')+ } rules {
		rule First { match Identifier report info at Identifier "first" }
		rule Second { match Number where Number.length >= 2 report err at Number "second" }
	}`
	document, parseDiagnostics := parser.Parse(source)
	validationDiagnostics := validator.Validate(source, document)
	wantProgram := ir.Program{
		Tokens: []ir.Token{
			{Name: "Identifier", Expression: ir.Expression{Kind: ir.ExpressionString, String: "id"}},
			{
				Name: "Number",
				Expression: ir.Expression{
					Kind:       ir.ExpressionRepetition,
					Inner:      pointer(ir.Expression{Kind: ir.ExpressionRange, RangeStart: '0', RangeEnd: '9'}),
					Repetition: ir.RepetitionOneOrMore,
				},
			},
		},
		Rules: []ir.Rule{
			{
				Name:       "First",
				MatchToken: 0,
				Where: ir.Condition{
					Property: ir.ConditionPropertyNone,
				},
				Report: ir.Report{
					Severity:    ir.SeverityInfo,
					TargetToken: 0,
					Message:     "first",
				},
			},
			{
				Name:       "Second",
				MatchToken: 1,
				Where: ir.Condition{
					Property: ir.ConditionPropertyLength,
					Operator: ir.ConditionOperatorGreaterEqual,
					Integer:  2,
				},
				Report: ir.Report{
					Severity:    ir.SeverityErr,
					TargetToken: 1,
					Message:     "second",
				},
			},
		},
	}

	// Act.
	gotProgram := compiler.Compile(source, document)

	// Assert.
	claim.Equal(t, "When compiling rules, parse diagnostics are not returned.", 0, len(parseDiagnostics), "Parse Diagnostic Count")
	claim.Equal(t, "When compiling rules, validation diagnostics are not returned.", 0, len(validationDiagnostics), "Validation Diagnostic Count")
	claim.DeepEqual(t, "When compiling rules, source order is preserved.", wantProgram, gotProgram, "Program")
}

func Test_Compile_Rules_CompilesConditionOperators(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name        string
		inSource    string
		wantProgram ir.Program
	}{
		{
			name:     "When compiling a text inequality rule, the text condition is compiled.",
			inSource: `scope { include "**/*.go" } tokens { Identifier = "id" } rules { rule NotPublic { match Identifier where Identifier.text != "public" report warn at Identifier "message" } }`,
			wantProgram: ir.Program{
				Tokens: []ir.Token{
					{Name: "Identifier", Expression: ir.Expression{Kind: ir.ExpressionString, String: "id"}},
				},
				Rules: []ir.Rule{
					{
						Name:       "NotPublic",
						MatchToken: 0,
						Where: ir.Condition{
							Property: ir.ConditionPropertyText,
							Operator: ir.ConditionOperatorNotEqual,
							String:   "public",
						},
						Report: ir.Report{
							Severity:    ir.SeverityWarn,
							TargetToken: 0,
							Message:     "message",
						},
					},
				},
			},
		},
		{
			name:     "When compiling a length less-than rule, the numeric condition is compiled.",
			inSource: `scope { include "**/*.go" } tokens { Identifier = "id" } rules { rule Short { match Identifier where Identifier.length < 10 report warn at Identifier "message" } }`,
			wantProgram: ir.Program{
				Tokens: []ir.Token{
					{Name: "Identifier", Expression: ir.Expression{Kind: ir.ExpressionString, String: "id"}},
				},
				Rules: []ir.Rule{
					{
						Name:       "Short",
						MatchToken: 0,
						Where: ir.Condition{
							Property: ir.ConditionPropertyLength,
							Operator: ir.ConditionOperatorLess,
							Integer:  10,
						},
						Report: ir.Report{
							Severity:    ir.SeverityWarn,
							TargetToken: 0,
							Message:     "message",
						},
					},
				},
			},
		},
		{
			name:     "When compiling a length less-than-or-equal rule, the numeric condition is compiled.",
			inSource: `scope { include "**/*.go" } tokens { Identifier = "id" } rules { rule ShortOrEqual { match Identifier where Identifier.length <= 10 report warn at Identifier "message" } }`,
			wantProgram: ir.Program{
				Tokens: []ir.Token{
					{Name: "Identifier", Expression: ir.Expression{Kind: ir.ExpressionString, String: "id"}},
				},
				Rules: []ir.Rule{
					{
						Name:       "ShortOrEqual",
						MatchToken: 0,
						Where: ir.Condition{
							Property: ir.ConditionPropertyLength,
							Operator: ir.ConditionOperatorLessEqual,
							Integer:  10,
						},
						Report: ir.Report{
							Severity:    ir.SeverityWarn,
							TargetToken: 0,
							Message:     "message",
						},
					},
				},
			},
		},
		{
			name:     "When compiling a length greater-than rule, the numeric condition is compiled.",
			inSource: `scope { include "**/*.go" } tokens { Identifier = "id" } rules { rule Long { match Identifier where Identifier.length > 10 report warn at Identifier "message" } }`,
			wantProgram: ir.Program{
				Tokens: []ir.Token{
					{Name: "Identifier", Expression: ir.Expression{Kind: ir.ExpressionString, String: "id"}},
				},
				Rules: []ir.Rule{
					{
						Name:       "Long",
						MatchToken: 0,
						Where: ir.Condition{
							Property: ir.ConditionPropertyLength,
							Operator: ir.ConditionOperatorGreater,
							Integer:  10,
						},
						Report: ir.Report{
							Severity:    ir.SeverityWarn,
							TargetToken: 0,
							Message:     "message",
						},
					},
				},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange.
			document, parseDiagnostics := parser.Parse(tc.inSource)
			validationDiagnostics := validator.Validate(tc.inSource, document)

			// Act.
			gotProgram := compiler.Compile(tc.inSource, document)

			// Assert.
			claim.Equal(t, tc.name, 0, len(parseDiagnostics), "Parse Diagnostic Count")
			claim.Equal(t, tc.name, 0, len(validationDiagnostics), "Validation Diagnostic Count")
			claim.DeepEqual(t, tc.name, tc.wantProgram, gotProgram, "Program")
		})
	}
}
