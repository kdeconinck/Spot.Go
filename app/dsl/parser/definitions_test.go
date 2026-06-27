// =====================================================================================================================
// == LICENSE:                 Copyright (c) 2026 Kevin De Coninck.
// == SPDX-License-Identifier: LicenseRef-PolyForm-Noncommercial-1.0.0
// =====================================================================================================================

// Verify the public API of the parser package.
//
// Tests in this package are written against the exported API only.
// This ensures that behavior is tested through the same surface that external consumers would use.
package parser_test

import (
	"strings"
	"testing"

	"github.com/kdeconinck/spot/dsl/ast"
	"github.com/kdeconinck/spot/dsl/parser"
	"github.com/kdeconinck/spot/dsl/token"
	"github.com/kdeconinck/spot/qa/claim"
)

func Test_Parse_Definitions(t *testing.T) {
	t.Parallel()

	for tcName, tc := range map[string]struct {
		inSource        string
		wantDocument    ast.Document
		wantDiagnostics []parser.Diagnostic
	}{
		"When parsing an empty definitions block, a document is returned.": {
			inSource: "scope {} definitions {}",
			wantDocument: document(
				span(0, 23),
				scopeSection(span(0, 8)),
				definitionsSection(span(9, 23)),
				ast.TokensSection{},
				ast.RulesSection{},
			),
		},
		"When parsing a definitions block with a character definition, a document is returned.": {
			inSource: "scope {} definitions { letter = 'a' }",
			wantDocument: document(
				span(0, 37),
				scopeSection(span(0, 8)),
				definitionsSection(span(9, 37), defineCharacter("letter", span(23, 29), span(23, 35), charExpr("'a'", span(32, 35)))),
				ast.TokensSection{},
				ast.RulesSection{},
			),
		},
		"When parsing a definitions block with a character range definition, a document is returned.": {
			inSource: "scope {} definitions { letter = 'a'..'z' }",
			wantDocument: document(
				span(0, 42),
				scopeSection(span(0, 8)),
				definitionsSection(
					span(9, 42),
					rangeDefinition("letter", 23, 29, "'a'", 32, 35, "'z'", 37, 40, 23, 40),
				),
				ast.TokensSection{},
				ast.RulesSection{},
			),
		},
		"When parsing a definitions block with a reference definition, a document is returned.": {
			inSource: "scope {} definitions { identifierStart = letter }",
			wantDocument: document(
				span(0, 49),
				scopeSection(span(0, 8)),
				definitionsSection(
					span(9, 49),
					defineReference(
						"identifierStart",
						span(23, 38),
						span(23, 47),
						refExpr("letter", span(41, 47)),
					),
				),
				ast.TokensSection{},
				ast.RulesSection{},
			),
		},
		"When parsing a definitions block with character concatenation, a document is returned.": {
			inSource: "scope {} definitions { value = 'a' 'b' }",
			wantDocument: document(
				span(0, 40),
				scopeSection(span(0, 8)),
				definitionsSection(
					span(9, 40),
					defineConcatenation(
						"value",
						span(23, 28),
						span(23, 38),
						charExpr("'a'", span(31, 34)),
						charExpr("'b'", span(35, 38)),
					),
				),
				ast.TokensSection{},
				ast.RulesSection{},
			),
		},
		"When parsing a definitions block with repeated reference concatenation, a document is returned.": {
			inSource: "scope {} definitions { value = letter digit* }",
			wantDocument: document(
				span(0, 46),
				scopeSection(span(0, 8)),
				definitionsSection(
					span(9, 46),
					defineConcatenation(
						"value",
						span(23, 28),
						span(23, 44),
						refExpr("letter", span(31, 37)),
						zeroOrMore(refExpr("digit", span(38, 43)), span(43, 44)),
					),
				),
				ast.TokensSection{},
				ast.RulesSection{},
			),
		},
		"When parsing a definitions block with grouped repetition concatenation, a document is returned.": {
			inSource: "scope {} definitions { value = letter ('_' | digit)+ }",
			wantDocument: document(
				span(0, 54),
				scopeSection(span(0, 8)),
				definitionsSection(
					span(9, 54),
					defineConcatenation(
						"value",
						span(23, 28),
						span(23, 52),
						refExpr("letter", span(31, 37)),
						oneOrMore(groupExpr(alternationExpr(charExpr("'_'", span(39, 42)), refExpr("digit", span(45, 50))), span(38, 51)), span(51, 52)),
					),
				),
				ast.TokensSection{},
				ast.RulesSection{},
			),
		},
		"When parsing multiple definitions after concatenation, a document is returned.": {
			inSource: "scope {} definitions { letter = 'a' value = letter digit }",
			wantDocument: document(
				span(0, 58),
				scopeSection(span(0, 8)),
				definitionsSection(
					span(9, 58),
					defineCharacter("letter", span(23, 29), span(23, 35), charExpr("'a'", span(32, 35))),
					defineConcatenation("value", span(36, 41), span(36, 56), refExpr("letter", span(44, 50)), refExpr("digit", span(51, 56))),
				),
				ast.TokensSection{},
				ast.RulesSection{},
			),
		},
		"When parsing a definitions block with character alternation, a document is returned.": {
			inSource: "scope {} definitions { value = 'a' | 'b' }",
			wantDocument: document(
				span(0, 42),
				scopeSection(span(0, 8)),
				definitionsSection(
					span(9, 42),
					defineAlternation(
						"value",
						span(23, 28),
						span(23, 40),
						charExpr("'a'", span(31, 34)),
						charExpr("'b'", span(37, 40)),
					),
				),
				ast.TokensSection{},
				ast.RulesSection{},
			),
		},
		"When parsing a definitions block with range alternation, a document is returned.": {
			inSource: "scope {} definitions { letter = 'a'..'z' | 'A'..'Z' }",
			wantDocument: document(
				span(0, 53),
				scopeSection(span(0, 8)),
				definitionsSection(
					span(9, 53),
					defineAlternation(
						"letter",
						span(23, 29),
						span(23, 51),
						rangeExpr("'a'", span(32, 35), "'z'", span(37, 40)),
						rangeExpr("'A'", span(43, 46), "'Z'", span(48, 51)),
					),
				),
				ast.TokensSection{},
				ast.RulesSection{},
			),
		},
		"When parsing a definitions block with reference alternation, a document is returned.": {
			inSource: "scope {} definitions { identifierStart = letter | '_' }",
			wantDocument: document(
				span(0, 55),
				scopeSection(span(0, 8)),
				definitionsSection(
					span(9, 55),
					defineAlternation(
						"identifierStart",
						span(23, 38),
						span(23, 53),
						refExpr("letter", span(41, 47)),
						charExpr("'_'", span(50, 53)),
					),
				),
				ast.TokensSection{},
				ast.RulesSection{},
			),
		},
		"When parsing a definitions block with concatenation before alternation, a document is returned.": {
			inSource: "scope {} definitions { value = letter digit | '_' }",
			wantDocument: document(
				span(0, 51),
				scopeSection(span(0, 8)),
				definitionsSection(
					span(9, 51),
					defineAlternation(
						"value",
						span(23, 28),
						span(23, 49),
						concatenationExpr(
							refExpr("letter", span(31, 37)),
							refExpr("digit", span(38, 43)),
						),
						charExpr("'_'", span(46, 49)),
					),
				),
				ast.TokensSection{},
				ast.RulesSection{},
			),
		},
		"When parsing a definitions block with a grouped expression, a document is returned.": {
			inSource: "scope {} definitions { value = ('a' | 'b') }",
			wantDocument: document(
				span(0, 44),
				scopeSection(span(0, 8)),
				definitionsSection(
					span(9, 44),
					defineGroup(
						"value",
						span(23, 28),
						span(23, 42),
						groupExpr(
							alternationExpr(
								charExpr("'a'", span(32, 35)),
								charExpr("'b'", span(38, 41)),
							),
							span(31, 42),
						),
					),
				),
				ast.TokensSection{},
				ast.RulesSection{},
			),
		},
		"When parsing a definitions block with zero-or-one repetition, a document is returned.": {
			inSource: "scope {} definitions { value = 'a'? }",
			wantDocument: document(
				span(0, 37),
				scopeSection(span(0, 8)),
				definitionsSection(
					span(9, 37),
					defineRepetition(
						"value",
						span(23, 28),
						span(23, 35),
						zeroOrOne(charExpr("'a'", span(31, 34)), span(34, 35)),
					),
				),
				ast.TokensSection{},
				ast.RulesSection{},
			),
		},
		"When parsing a definitions block with zero-or-more repetition, a document is returned.": {
			inSource: "scope {} definitions { value = letter* }",
			wantDocument: document(
				span(0, 40),
				scopeSection(span(0, 8)),
				definitionsSection(
					span(9, 40),
					defineRepetition(
						"value",
						span(23, 28),
						span(23, 38),
						zeroOrMore(refExpr("letter", span(31, 37)), span(37, 38)),
					),
				),
				ast.TokensSection{},
				ast.RulesSection{},
			),
		},
		"When parsing a definitions block with one-or-more repetition, a document is returned.": {
			inSource: "scope {} definitions { value = ('a' | 'b')+ }",
			wantDocument: document(
				span(0, 45),
				scopeSection(span(0, 8)),
				definitionsSection(
					span(9, 45),
					defineRepetition(
						"value",
						span(23, 28),
						span(23, 43),
						oneOrMore(
							groupExpr(
								alternationExpr(
									charExpr("'a'", span(32, 35)),
									charExpr("'b'", span(38, 41)),
								),
								span(31, 42),
							),
							span(42, 43),
						),
					),
				),
				ast.TokensSection{},
				ast.RulesSection{},
			),
		},
		"When the definitions opening brace is missing, a diagnostic is returned.": {
			inSource:        "scope {} definitions }",
			wantDocument:    document(span(0, 20), scopeSection(span(0, 8)), definitionsSection(span(9, 20)), ast.TokensSection{}, ast.RulesSection{}),
			wantDiagnostics: []parser.Diagnostic{diagnostic("Expected '{', found '}'.", 21, 22)},
		},
		"When the definitions closing brace is missing, a diagnostic is returned.": {
			inSource:        "scope {} definitions {",
			wantDocument:    document(span(0, 22), scopeSection(span(0, 8)), definitionsSection(span(9, 22)), ast.TokensSection{}, ast.RulesSection{}),
			wantDiagnostics: []parser.Diagnostic{diagnostic("Expected '}', found 'EOF'.", 22, 22)},
		},
		"When an unexpected token appears inside definitions, a diagnostic is returned.": {
			inSource:        "scope {} definitions { 'a' }",
			wantDocument:    document(span(0, 26), scopeSection(span(0, 8)), definitionsSection(span(9, 26)), ast.TokensSection{}, ast.RulesSection{}),
			wantDiagnostics: []parser.Diagnostic{diagnostic("Expected 'identifier', found 'character'.", 23, 26)},
		},
		"When a definition is missing an equal sign, a diagnostic is returned.": {
			inSource: "scope {} definitions { letter 'a' }",
			wantDocument: document(
				span(0, 35),
				scopeSection(span(0, 8)),
				definitionsSection(
					span(9, 35),
					defineCharacter(
						"letter",
						span(23, 29),
						span(23, 33),
						charExpr("'a'", span(30, 33)),
					),
				),
				ast.TokensSection{},
				ast.RulesSection{},
			),
			wantDiagnostics: []parser.Diagnostic{diagnostic("Expected '=', found 'character'.", 30, 33)},
		},
		"When a definition is missing an expression, a diagnostic is returned.": {
			inSource: "scope {} definitions { letter = }",
			wantDocument: document(
				span(0, 33),
				scopeSection(span(0, 8)),
				definitionsSection(
					span(9, 33),
					defineCharacter(
						"letter",
						span(23, 29),
						span(23, 33),
						characterExpression(token.TokenRightBrace, "}", 32, 33),
					),
				),
				ast.TokensSection{},
				ast.RulesSection{},
			),
			wantDiagnostics: []parser.Diagnostic{diagnostic("Expected 'character', found '}'.", 32, 33)},
		},
		"When a character range is missing an end character, a diagnostic is returned.": {
			inSource: "scope {} definitions { letter = 'a'.. }",
			wantDocument: document(
				span(0, 39),
				scopeSection(span(0, 8)),
				definitionsSection(
					span(9, 39),
					rangeDefinitionWithEndKind("letter", 23, 29, "'a'", 32, 35, token.TokenRightBrace, "}", 38, 39, 23, 39),
				),
				ast.TokensSection{},
				ast.RulesSection{},
			),
			wantDiagnostics: []parser.Diagnostic{diagnostic("Expected 'character', found '}'.", 38, 39)},
		},
		"When alternation is missing a right expression, a diagnostic is returned.": {
			inSource: "scope {} definitions { value = 'a' | }",
			wantDocument: document(
				span(0, 38),
				scopeSection(span(0, 8)),
				definitionsSection(
					span(9, 38),
					defineAlternation(
						"value",
						span(23, 28),
						span(23, 38),
						charExpr("'a'", span(31, 34)),
						characterExpression(token.TokenRightBrace, "}", 37, 38),
					),
				),
				ast.TokensSection{},
				ast.RulesSection{},
			),
			wantDiagnostics: []parser.Diagnostic{diagnostic("Expected 'character', found '}'.", 37, 38)},
		},
		"When a grouped expression is missing a closing parenthesis, a diagnostic is returned.": {
			inSource: "scope {} definitions { value = ('a' | 'b' }",
			wantDocument: document(
				span(0, 43),
				scopeSection(span(0, 8)),
				definitionsSection(
					span(9, 43),
					defineGroup(
						"value",
						span(23, 28),
						span(23, 43),
						groupExpr(
							alternationExpr(
								charExpr("'a'", span(32, 35)),
								charExpr("'b'", span(38, 41)),
							),
							span(31, 43),
						),
					),
				),
				ast.TokensSection{},
				ast.RulesSection{},
			),
			wantDiagnostics: []parser.Diagnostic{diagnostic("Expected ')', found '}'.", 42, 43)},
		},
	} {
		t.Run(tcName, func(t *testing.T) {
			t.Parallel()

			// Act.
			gotDocument, gotDiagnostics := parser.Parse(tc.inSource)

			// Assert.
			claim.DeepEqual(t, tcName, tc.wantDocument, gotDocument, "Document")
			claim.Equal(t, tcName, len(tc.wantDiagnostics), len(gotDiagnostics), "Diagnostic Count")

			for idx := range tc.wantDiagnostics {
				claim.Equal(t, tcName, tc.wantDiagnostics[idx], gotDiagnostics[idx], "Diagnostic")
			}
		})
	}
}

func Benchmark_Parse_Definitions_0(b *testing.B)    { benchmark_Parse_Definitions(b, 0) }
func Benchmark_Parse_Definitions_1(b *testing.B)    { benchmark_Parse_Definitions(b, 1) }
func Benchmark_Parse_Definitions_10(b *testing.B)   { benchmark_Parse_Definitions(b, 10) }
func Benchmark_Parse_Definitions_100(b *testing.B)  { benchmark_Parse_Definitions(b, 100) }
func Benchmark_Parse_Definitions_1000(b *testing.B) { benchmark_Parse_Definitions(b, 1000) }

func benchmark_Parse_Definitions(b *testing.B, size int) {
	b.Helper()

	benchmark_Parse(b, definitionsDSL(size))
}

func definitionsDSL(size int) string {
	return "scope {}\n" +
		"definitions {\n" +
		strings.Repeat("    letter = 'a'..'z' | 'A'..'Z'\n    identifierStart = letter | '_'\n    value = letter ('a' | 'b')+\n", size) +
		"}"
}
