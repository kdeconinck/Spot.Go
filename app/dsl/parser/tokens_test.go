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

func Test_Parse_Tokens(t *testing.T) {
	t.Parallel()

	for tcName, tc := range map[string]struct {
		inSource        string
		wantDocument    ast.Document
		wantDiagnostics []parser.Diagnostic
	}{
		"When parsing an empty tokens block, a document is returned.": {
			inSource: "scope {} tokens {}",
			wantDocument: document(
				span(0, 18),
				scopeSection(span(0, 8)),
				ast.DefinitionsSection{},
				tokensSection(span(9, 18)),
				ast.RulesSection{},
			),
		},
		"When parsing a definitions block followed by an empty tokens block, a document is returned.": {
			inSource: "scope {} definitions {} tokens {}",
			wantDocument: document(
				span(0, 33),
				scopeSection(span(0, 8)),
				definitionsSection(span(9, 23)),
				tokensSection(span(24, 33)),
				ast.RulesSection{},
			),
		},
		"When parsing a tokens block with a reference token, a document is returned.": {
			inSource: "scope {} tokens { Identifier = letter }",
			wantDocument: document(
				span(0, 39),
				scopeSection(span(0, 8)),
				ast.DefinitionsSection{},
				tokensSection(
					span(9, 39),
					defineToken(
						"Identifier",
						span(18, 28),
						span(18, 37),
						refExpr("letter", span(31, 37)),
					),
				),
				ast.RulesSection{},
			),
		},
		"When parsing a tokens block with a string token, a document is returned.": {
			inSource: "scope {} tokens { KeywordPublic = \"public\" }",
			wantDocument: document(
				span(0, 44),
				scopeSection(span(0, 8)),
				ast.DefinitionsSection{},
				tokensSection(
					span(9, 44),
					defineToken(
						"KeywordPublic",
						span(18, 31),
						span(18, 42),
						stringExpr(`"public"`, span(34, 42)),
					),
				),
				ast.RulesSection{},
			),
		},
		"When parsing a tokens block with a skipped token, a document is returned.": {
			inSource: "scope {} tokens { Whitespace = (' ' | '\\t')+ skip }",
			wantDocument: document(
				span(0, 51),
				scopeSection(span(0, 8)),
				ast.DefinitionsSection{},
				tokensSection(
					span(9, 51),
					defineSkippedToken(
						"Whitespace",
						span(18, 28),
						span(45, 49),
						span(18, 49),
						oneOrMore(
							groupExpr(
								alternationExpr(
									charExpr("' '", span(32, 35)),
									charExpr(`'\t'`, span(38, 42)),
								),
								span(31, 43),
							),
							span(43, 44),
						),
					),
				),
				ast.RulesSection{},
			),
		},
		"When the tokens opening brace is missing, a diagnostic is returned.": {
			inSource:        "scope {} tokens }",
			wantDocument:    document(span(0, 15), scopeSection(span(0, 8)), ast.DefinitionsSection{}, tokensSection(span(9, 15)), ast.RulesSection{}),
			wantDiagnostics: []parser.Diagnostic{diagnostic("Expected '{', found '}'.", 16, 17)},
		},
		"When the tokens closing brace is missing, a diagnostic is returned.": {
			inSource:        "scope {} tokens {",
			wantDocument:    document(span(0, 17), scopeSection(span(0, 8)), ast.DefinitionsSection{}, tokensSection(span(9, 17)), ast.RulesSection{}),
			wantDiagnostics: []parser.Diagnostic{diagnostic("Expected '}', found 'EOF'.", 17, 17)},
		},
		"When an unexpected token appears inside tokens, a diagnostic is returned.": {
			inSource:        "scope {} tokens { 'a' }",
			wantDocument:    document(span(0, 21), scopeSection(span(0, 8)), ast.DefinitionsSection{}, tokensSection(span(9, 21)), ast.RulesSection{}),
			wantDiagnostics: []parser.Diagnostic{diagnostic("Expected 'identifier', found 'character'.", 18, 21)},
		},
		"When a token is missing an expression, a diagnostic is returned.": {
			inSource: "scope {} tokens { Identifier = }",
			wantDocument: document(
				span(0, 32),
				scopeSection(span(0, 8)),
				ast.DefinitionsSection{},
				tokensSection(
					span(9, 32),
					defineToken(
						"Identifier",
						span(18, 28),
						span(18, 32),
						characterExpression(token.TokenRightBrace, "}", 31, 32),
					),
				),
				ast.RulesSection{},
			),
			wantDiagnostics: []parser.Diagnostic{diagnostic("Expected 'character', found '}'.", 31, 32)},
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

func Benchmark_Parse_Tokens_0(b *testing.B)    { benchmark_Parse_Tokens(b, 0) }
func Benchmark_Parse_Tokens_1(b *testing.B)    { benchmark_Parse_Tokens(b, 1) }
func Benchmark_Parse_Tokens_10(b *testing.B)   { benchmark_Parse_Tokens(b, 10) }
func Benchmark_Parse_Tokens_100(b *testing.B)  { benchmark_Parse_Tokens(b, 100) }
func Benchmark_Parse_Tokens_1000(b *testing.B) { benchmark_Parse_Tokens(b, 1000) }

func benchmark_Parse_Tokens(b *testing.B, size int) {
	b.Helper()

	benchmark_Parse(b, tokensDSL(size))
}

func tokensDSL(size int) string {
	return "scope {}\n" +
		"definitions {\n" +
		"    identifierStart = 'a'..'z' | '_'\n" +
		"    value = identifierStart*\n" +
		"}\n" +
		"tokens {\n" +
		strings.Repeat("    Identifier = identifierStart value*\n    KeywordPublic = \"public\"\n    Whitespace = (' ' | '\\t')+ skip\n", size) +
		"}"
}
