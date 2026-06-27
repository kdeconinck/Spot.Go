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
			wantDocument: ast.Document{
				Scope: ast.ScopeSection{
					Span: span(0, 8),
				},
				Tokens: ast.TokensSection{
					Span: span(9, 18),
				},
				Span: span(0, 18),
			},
		},
		"When parsing a definitions block followed by an empty tokens block, a document is returned.": {
			inSource: "scope {} definitions {} tokens {}",
			wantDocument: ast.Document{
				Scope: ast.ScopeSection{
					Span: span(0, 8),
				},
				Definitions: ast.DefinitionsSection{
					Span: span(9, 23),
				},
				Tokens: ast.TokensSection{
					Span: span(24, 33),
				},
				Span: span(0, 33),
			},
		},
		"When parsing a tokens block with a reference token, a document is returned.": {
			inSource: "scope {} tokens { Identifier = letter }",
			wantDocument: ast.Document{
				Scope: ast.ScopeSection{
					Span: span(0, 8),
				},
				Tokens: ast.TokensSection{
					Tokens: []ast.TokenDefinition{
						tokenDefinition("Identifier", 18, 28, referenceExpression("letter", 31, 37), 18, 37),
					},
					Span: span(9, 39),
				},
				Span: span(0, 39),
			},
		},
		"When parsing a tokens block with a string token, a document is returned.": {
			inSource: "scope {} tokens { KeywordPublic = \"public\" }",
			wantDocument: ast.Document{
				Scope: ast.ScopeSection{
					Span: span(0, 8),
				},
				Tokens: ast.TokensSection{
					Tokens: []ast.TokenDefinition{
						tokenDefinition("KeywordPublic", 18, 31, stringExpression("\"public\"", 34, 42), 18, 42),
					},
					Span: span(9, 44),
				},
				Span: span(0, 44),
			},
		},
		"When parsing a tokens block with a skipped token, a document is returned.": {
			inSource: "scope {} tokens { Whitespace = (' ' | '\\t')+ skip }",
			wantDocument: ast.Document{
				Scope: ast.ScopeSection{
					Span: span(0, 8),
				},
				Tokens: ast.TokensSection{
					Tokens: []ast.TokenDefinition{
						tokenDefinitionWithSkip("Whitespace", 18, 28, repetitionExpression(groupExpression(alternationExpression(characterExpression(token.TokenCharacter, "' '", 32, 35), characterExpression(token.TokenCharacter, "'\\t'", 38, 42)), 31, 43), token.TokenPlus, "+", 43, 44), 45, 49, 18, 49),
					},
					Span: span(9, 51),
				},
				Span: span(0, 51),
			},
		},
		"When the tokens opening brace is missing, a diagnostic is returned.": {
			inSource: "scope {} tokens }",
			wantDocument: ast.Document{
				Scope: ast.ScopeSection{
					Span: span(0, 8),
				},
				Tokens: ast.TokensSection{
					Span: span(9, 15),
				},
				Span: span(0, 15),
			},
			wantDiagnostics: []parser.Diagnostic{
				diagnostic("Expected '{', found '}'.", 16, 17),
			},
		},
		"When the tokens closing brace is missing, a diagnostic is returned.": {
			inSource: "scope {} tokens {",
			wantDocument: ast.Document{
				Scope: ast.ScopeSection{
					Span: span(0, 8),
				},
				Tokens: ast.TokensSection{
					Span: span(9, 17),
				},
				Span: span(0, 17),
			},
			wantDiagnostics: []parser.Diagnostic{
				diagnostic("Expected '}', found 'EOF'.", 17, 17),
			},
		},
		"When an unexpected token appears inside tokens, a diagnostic is returned.": {
			inSource: "scope {} tokens { 'a' }",
			wantDocument: ast.Document{
				Scope: ast.ScopeSection{
					Span: span(0, 8),
				},
				Tokens: ast.TokensSection{
					Span: span(9, 21),
				},
				Span: span(0, 21),
			},
			wantDiagnostics: []parser.Diagnostic{
				diagnostic("Expected 'identifier', found 'character'.", 18, 21),
			},
		},
		"When a token is missing an expression, a diagnostic is returned.": {
			inSource: "scope {} tokens { Identifier = }",
			wantDocument: ast.Document{
				Scope: ast.ScopeSection{
					Span: span(0, 8),
				},
				Tokens: ast.TokensSection{
					Tokens: []ast.TokenDefinition{
						tokenDefinition("Identifier", 18, 28, characterExpression(token.TokenRightBrace, "}", 31, 32), 18, 32),
					},
					Span: span(9, 32),
				},
				Span: span(0, 32),
			},
			wantDiagnostics: []parser.Diagnostic{
				diagnostic("Expected 'character', found '}'.", 31, 32),
			},
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
