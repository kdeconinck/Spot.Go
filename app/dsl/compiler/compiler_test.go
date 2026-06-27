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
	"strconv"
	"strings"
	"testing"

	"github.com/kdeconinck/spot/dsl/compiler"
	"github.com/kdeconinck/spot/dsl/parser"
	"github.com/kdeconinck/spot/dsl/validator"
	"github.com/kdeconinck/spot/qa/claim"
	"github.com/kdeconinck/spot/runtime/ir"
)

func Test_Compile_DSL(t *testing.T) {
	t.Parallel()

	// Arrange.
	source := dsl(0)
	document, parseDiagnostics := parser.Parse(source)
	validationDiagnostics := validator.Validate(document)
	wantProgram := ir.Program{
		Tokens: []ir.Token{
			{
				Name: "Identifier",
				Expression: ir.Expression{
					Kind: ir.ExpressionConcatenation,
					Terms: []ir.Expression{
						{
							Kind: ir.ExpressionAlternation,
							Terms: []ir.Expression{
								{Kind: ir.ExpressionRange, RangeStart: 'a', RangeEnd: 'z'},
								{Kind: ir.ExpressionRange, RangeStart: 'A', RangeEnd: 'Z'},
							},
						},
						{
							Kind: ir.ExpressionRepetition,
							Inner: pointer(ir.Expression{
								Kind: ir.ExpressionAlternation,
								Terms: []ir.Expression{
									{
										Kind: ir.ExpressionAlternation,
										Terms: []ir.Expression{
											{Kind: ir.ExpressionRange, RangeStart: 'a', RangeEnd: 'z'},
											{Kind: ir.ExpressionRange, RangeStart: 'A', RangeEnd: 'Z'},
										},
									},
									{Kind: ir.ExpressionRange, RangeStart: '0', RangeEnd: '9'},
									{Kind: ir.ExpressionCharacter, Character: '_'},
								},
							}),
							Repetition: ir.RepetitionZeroOrMore,
						},
					},
				},
			},
			{
				Name:       "KeywordPublic",
				Expression: ir.Expression{Kind: ir.ExpressionString, String: "public"},
			},
			{
				Name: "Whitespace",
				Expression: ir.Expression{
					Kind: ir.ExpressionRepetition,
					Inner: pointer(ir.Expression{
						Kind: ir.ExpressionAlternation,
						Terms: []ir.Expression{
							{Kind: ir.ExpressionCharacter, Character: ' '},
							{Kind: ir.ExpressionCharacter, Character: '\t'},
						},
					}),
					Repetition: ir.RepetitionOneOrMore,
				},
				Skip: true,
			},
		},
		Rules: []ir.Rule{
			{
				Name:       "PublicIdentifier",
				MatchToken: 0,
				Where: ir.Condition{
					Property: ir.ConditionPropertyText,
					Operator: ir.ConditionOperatorEqual,
					String:   "public",
				},
				Report: ir.Report{
					Severity:    ir.SeverityWarn,
					TargetToken: 0,
					Message:     "Public identifier found",
				},
			},
		},
	}

	// Act.
	gotProgram := compiler.Compile(document)

	// Assert.
	claim.Equal(t, "When compiling a full DSL file, parse diagnostics are not returned.", 0, len(parseDiagnostics), "Parse Diagnostic Count")
	claim.Equal(t, "When compiling a full DSL file, validation diagnostics are not returned.", 0, len(validationDiagnostics), "Validation Diagnostic Count")
	claim.DeepEqual(t, "When compiling a full DSL file, a program is returned.", wantProgram, gotProgram, "Program")
}

func Benchmark_Compile_DSL_0(b *testing.B)    { benchmark_Compile_DSL(b, 0) }
func Benchmark_Compile_DSL_1(b *testing.B)    { benchmark_Compile_DSL(b, 1) }
func Benchmark_Compile_DSL_10(b *testing.B)   { benchmark_Compile_DSL(b, 10) }
func Benchmark_Compile_DSL_100(b *testing.B)  { benchmark_Compile_DSL(b, 100) }
func Benchmark_Compile_DSL_1000(b *testing.B) { benchmark_Compile_DSL(b, 1000) }

func benchmark_Compile_DSL(b *testing.B, size int) {
	b.Helper()

	document, parseDiagnostics := parser.Parse(dsl(size))
	validationDiagnostics := validator.Validate(document)
	claim.Equal(b, "Compile DSL benchmark parse diagnostics.", 0, len(parseDiagnostics), "Parse Diagnostic Count")
	claim.Equal(b, "Compile DSL benchmark validation diagnostics.", 0, len(validationDiagnostics), "Validation Diagnostic Count")

	for b.Loop() {
		_ = compiler.Compile(document)
	}
}

func dsl(size int) string {
	var sb strings.Builder

	sb.WriteString("scope {\n")
	sb.WriteString("    include \"**/*.go\"\n")
	sb.WriteString("    exclude \"vendor/**\"\n")

	for range size {
		sb.WriteString("    include \"**/*.go\"\n")
		sb.WriteString("    exclude \"vendor/**\"\n")
	}

	sb.WriteString("}\n")
	sb.WriteString("definitions {\n")
	sb.WriteString("    letter = 'a'..'z' | 'A'..'Z'\n")
	sb.WriteString("    digit = '0'..'9'\n")
	sb.WriteString("    identifierStart = letter\n")
	sb.WriteString("    identifier = identifierStart (identifierStart | digit | '_')*\n")

	for idx := 1; idx <= size; idx++ {
		sb.WriteString("    letter")
		sb.WriteString(strconv.Itoa(idx))
		sb.WriteString(" = 'a'..'z' | 'A'..'Z'\n")
		sb.WriteString("    digit")
		sb.WriteString(strconv.Itoa(idx))
		sb.WriteString(" = '0'..'9'\n")
		sb.WriteString("    identifierStart")
		sb.WriteString(strconv.Itoa(idx))
		sb.WriteString(" = letter")
		sb.WriteString(strconv.Itoa(idx))
		sb.WriteString("\n")
		sb.WriteString("    identifier")
		sb.WriteString(strconv.Itoa(idx))
		sb.WriteString(" = identifierStart")
		sb.WriteString(strconv.Itoa(idx))
		sb.WriteString(" (identifierStart")
		sb.WriteString(strconv.Itoa(idx))
		sb.WriteString(" | digit")
		sb.WriteString(strconv.Itoa(idx))
		sb.WriteString(" | '_')*\n")
	}

	sb.WriteString("}\n")
	sb.WriteString("tokens {\n")
	sb.WriteString("    Identifier = identifier\n")
	sb.WriteString("    KeywordPublic = \"public\"\n")
	sb.WriteString("    Whitespace = (' ' | '\\t')+ skip\n")

	for idx := 1; idx <= size; idx++ {
		sb.WriteString("    Identifier")
		sb.WriteString(strconv.Itoa(idx))
		sb.WriteString(" = identifier")
		sb.WriteString(strconv.Itoa(idx))
		sb.WriteString("\n")
		sb.WriteString("    KeywordPublic")
		sb.WriteString(strconv.Itoa(idx))
		sb.WriteString(" = \"public\"\n")
		sb.WriteString("    Whitespace")
		sb.WriteString(strconv.Itoa(idx))
		sb.WriteString(" = (' ' | '\\t')+ skip\n")
	}

	sb.WriteString("}\n")
	sb.WriteString("rules {\n")
	sb.WriteString("    rule PublicIdentifier {\n")
	sb.WriteString("        match Identifier\n")
	sb.WriteString("        where Identifier.text == \"public\"\n")
	sb.WriteString("        report warn at Identifier \"Public identifier found\"\n")
	sb.WriteString("    }\n")

	for idx := 1; idx <= size; idx++ {
		sb.WriteString("    rule PublicIdentifier")
		sb.WriteString(strconv.Itoa(idx))
		sb.WriteString(" {\n")
		sb.WriteString("        match Identifier")
		sb.WriteString(strconv.Itoa(idx))
		sb.WriteString("\n")
		sb.WriteString("        where Identifier")
		sb.WriteString(strconv.Itoa(idx))
		sb.WriteString(".text == \"public\"\n")
		sb.WriteString("        report warn at Identifier")
		sb.WriteString(strconv.Itoa(idx))
		sb.WriteString(" \"Public identifier found\"\n")
		sb.WriteString("    }\n")
	}

	sb.WriteString("}")

	return sb.String()
}
