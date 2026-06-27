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
			wantDocument: ast.Document{
				Scope: ast.ScopeSection{
					Span: span(0, 8),
				},
				Definitions: ast.DefinitionsSection{
					Span: span(9, 23),
				},
				Span: span(0, 23),
			},
		},
		"When parsing a definitions block with a character definition, a document is returned.": {
			inSource: "scope {} definitions { letter = 'a' }",
			wantDocument: ast.Document{
				Scope: ast.ScopeSection{
					Span: span(0, 8),
				},
				Definitions: ast.DefinitionsSection{
					Definitions: []ast.Definition{
						characterDefinition("letter", 23, 29, token.TokenCharacter, "'a'", 32, 35, 23, 35),
					},
					Span: span(9, 37),
				},
				Span: span(0, 37),
			},
		},
		"When parsing a definitions block with a character range definition, a document is returned.": {
			inSource: "scope {} definitions { letter = 'a'..'z' }",
			wantDocument: ast.Document{
				Scope: ast.ScopeSection{
					Span: span(0, 8),
				},
				Definitions: ast.DefinitionsSection{
					Definitions: []ast.Definition{
						rangeDefinition("letter", 23, 29, "'a'", 32, 35, "'z'", 37, 40, 23, 40),
					},
					Span: span(9, 42),
				},
				Span: span(0, 42),
			},
		},
		"When parsing a definitions block with a reference definition, a document is returned.": {
			inSource: "scope {} definitions { identifierStart = letter }",
			wantDocument: ast.Document{
				Scope: ast.ScopeSection{
					Span: span(0, 8),
				},
				Definitions: ast.DefinitionsSection{
					Definitions: []ast.Definition{
						referenceDefinition("identifierStart", 23, 38, "letter", 41, 47, 23, 47),
					},
					Span: span(9, 49),
				},
				Span: span(0, 49),
			},
		},
		"When parsing a definitions block with character concatenation, a document is returned.": {
			inSource: "scope {} definitions { value = 'a' 'b' }",
			wantDocument: ast.Document{
				Scope: ast.ScopeSection{
					Span: span(0, 8),
				},
				Definitions: ast.DefinitionsSection{
					Definitions: []ast.Definition{
						concatenationDefinition("value", 23, 28, 23, 38, characterExpression(token.TokenCharacter, "'a'", 31, 34), characterExpression(token.TokenCharacter, "'b'", 35, 38)),
					},
					Span: span(9, 40),
				},
				Span: span(0, 40),
			},
		},
		"When parsing a definitions block with repeated reference concatenation, a document is returned.": {
			inSource: "scope {} definitions { value = letter digit* }",
			wantDocument: ast.Document{
				Scope: ast.ScopeSection{
					Span: span(0, 8),
				},
				Definitions: ast.DefinitionsSection{
					Definitions: []ast.Definition{
						concatenationDefinition("value", 23, 28, 23, 44, referenceExpression("letter", 31, 37), repetitionExpression(referenceExpression("digit", 38, 43), token.TokenStar, "*", 43, 44)),
					},
					Span: span(9, 46),
				},
				Span: span(0, 46),
			},
		},
		"When parsing a definitions block with grouped repetition concatenation, a document is returned.": {
			inSource: "scope {} definitions { value = letter ('_' | digit)+ }",
			wantDocument: ast.Document{
				Scope: ast.ScopeSection{
					Span: span(0, 8),
				},
				Definitions: ast.DefinitionsSection{
					Definitions: []ast.Definition{
						concatenationDefinition("value", 23, 28, 23, 52, referenceExpression("letter", 31, 37), repetitionExpression(groupExpression(alternationExpression(characterExpression(token.TokenCharacter, "'_'", 39, 42), referenceExpression("digit", 45, 50)), 38, 51), token.TokenPlus, "+", 51, 52)),
					},
					Span: span(9, 54),
				},
				Span: span(0, 54),
			},
		},
		"When parsing multiple definitions after concatenation, a document is returned.": {
			inSource: "scope {} definitions { letter = 'a' value = letter digit }",
			wantDocument: ast.Document{
				Scope: ast.ScopeSection{
					Span: span(0, 8),
				},
				Definitions: ast.DefinitionsSection{
					Definitions: []ast.Definition{
						characterDefinition("letter", 23, 29, token.TokenCharacter, "'a'", 32, 35, 23, 35),
						concatenationDefinition("value", 36, 41, 36, 56, referenceExpression("letter", 44, 50), referenceExpression("digit", 51, 56)),
					},
					Span: span(9, 58),
				},
				Span: span(0, 58),
			},
		},
		"When parsing a definitions block with character alternation, a document is returned.": {
			inSource: "scope {} definitions { value = 'a' | 'b' }",
			wantDocument: ast.Document{
				Scope: ast.ScopeSection{
					Span: span(0, 8),
				},
				Definitions: ast.DefinitionsSection{
					Definitions: []ast.Definition{
						alternationDefinition("value", 23, 28, 23, 40, characterExpression(token.TokenCharacter, "'a'", 31, 34), characterExpression(token.TokenCharacter, "'b'", 37, 40)),
					},
					Span: span(9, 42),
				},
				Span: span(0, 42),
			},
		},
		"When parsing a definitions block with range alternation, a document is returned.": {
			inSource: "scope {} definitions { letter = 'a'..'z' | 'A'..'Z' }",
			wantDocument: ast.Document{
				Scope: ast.ScopeSection{
					Span: span(0, 8),
				},
				Definitions: ast.DefinitionsSection{
					Definitions: []ast.Definition{
						alternationDefinition("letter", 23, 29, 23, 51, rangeExpression("'a'", 32, 35, token.TokenCharacter, "'z'", 37, 40), rangeExpression("'A'", 43, 46, token.TokenCharacter, "'Z'", 48, 51)),
					},
					Span: span(9, 53),
				},
				Span: span(0, 53),
			},
		},
		"When parsing a definitions block with reference alternation, a document is returned.": {
			inSource: "scope {} definitions { identifierStart = letter | '_' }",
			wantDocument: ast.Document{
				Scope: ast.ScopeSection{
					Span: span(0, 8),
				},
				Definitions: ast.DefinitionsSection{
					Definitions: []ast.Definition{
						alternationDefinition("identifierStart", 23, 38, 23, 53, referenceExpression("letter", 41, 47), characterExpression(token.TokenCharacter, "'_'", 50, 53)),
					},
					Span: span(9, 55),
				},
				Span: span(0, 55),
			},
		},
		"When parsing a definitions block with concatenation before alternation, a document is returned.": {
			inSource: "scope {} definitions { value = letter digit | '_' }",
			wantDocument: ast.Document{
				Scope: ast.ScopeSection{
					Span: span(0, 8),
				},
				Definitions: ast.DefinitionsSection{
					Definitions: []ast.Definition{
						alternationDefinition("value", 23, 28, 23, 49, concatenationExpression(referenceExpression("letter", 31, 37), referenceExpression("digit", 38, 43)), characterExpression(token.TokenCharacter, "'_'", 46, 49)),
					},
					Span: span(9, 51),
				},
				Span: span(0, 51),
			},
		},
		"When parsing a definitions block with a grouped expression, a document is returned.": {
			inSource: "scope {} definitions { value = ('a' | 'b') }",
			wantDocument: ast.Document{
				Scope: ast.ScopeSection{
					Span: span(0, 8),
				},
				Definitions: ast.DefinitionsSection{
					Definitions: []ast.Definition{
						groupDefinition("value", 23, 28, groupExpression(alternationExpression(characterExpression(token.TokenCharacter, "'a'", 32, 35), characterExpression(token.TokenCharacter, "'b'", 38, 41)), 31, 42), 23, 42),
					},
					Span: span(9, 44),
				},
				Span: span(0, 44),
			},
		},
		"When parsing a definitions block with zero-or-one repetition, a document is returned.": {
			inSource: "scope {} definitions { value = 'a'? }",
			wantDocument: ast.Document{
				Scope: ast.ScopeSection{
					Span: span(0, 8),
				},
				Definitions: ast.DefinitionsSection{
					Definitions: []ast.Definition{
						repetitionDefinition("value", 23, 28, repetitionExpression(characterExpression(token.TokenCharacter, "'a'", 31, 34), token.TokenQuestion, "?", 34, 35), 23, 35),
					},
					Span: span(9, 37),
				},
				Span: span(0, 37),
			},
		},
		"When parsing a definitions block with zero-or-more repetition, a document is returned.": {
			inSource: "scope {} definitions { value = letter* }",
			wantDocument: ast.Document{
				Scope: ast.ScopeSection{
					Span: span(0, 8),
				},
				Definitions: ast.DefinitionsSection{
					Definitions: []ast.Definition{
						repetitionDefinition("value", 23, 28, repetitionExpression(referenceExpression("letter", 31, 37), token.TokenStar, "*", 37, 38), 23, 38),
					},
					Span: span(9, 40),
				},
				Span: span(0, 40),
			},
		},
		"When parsing a definitions block with one-or-more repetition, a document is returned.": {
			inSource: "scope {} definitions { value = ('a' | 'b')+ }",
			wantDocument: ast.Document{
				Scope: ast.ScopeSection{
					Span: span(0, 8),
				},
				Definitions: ast.DefinitionsSection{
					Definitions: []ast.Definition{
						repetitionDefinition("value", 23, 28, repetitionExpression(groupExpression(alternationExpression(characterExpression(token.TokenCharacter, "'a'", 32, 35), characterExpression(token.TokenCharacter, "'b'", 38, 41)), 31, 42), token.TokenPlus, "+", 42, 43), 23, 43),
					},
					Span: span(9, 45),
				},
				Span: span(0, 45),
			},
		},
		"When the definitions opening brace is missing, a diagnostic is returned.": {
			inSource: "scope {} definitions }",
			wantDocument: ast.Document{
				Scope: ast.ScopeSection{
					Span: span(0, 8),
				},
				Definitions: ast.DefinitionsSection{
					Span: span(9, 20),
				},
				Span: span(0, 20),
			},
			wantDiagnostics: []parser.Diagnostic{
				diagnostic("Expected '{', found '}'.", 21, 22),
			},
		},
		"When the definitions closing brace is missing, a diagnostic is returned.": {
			inSource: "scope {} definitions {",
			wantDocument: ast.Document{
				Scope: ast.ScopeSection{
					Span: span(0, 8),
				},
				Definitions: ast.DefinitionsSection{
					Span: span(9, 22),
				},
				Span: span(0, 22),
			},
			wantDiagnostics: []parser.Diagnostic{
				diagnostic("Expected '}', found 'EOF'.", 22, 22),
			},
		},
		"When an unexpected token appears inside definitions, a diagnostic is returned.": {
			inSource: "scope {} definitions { 'a' }",
			wantDocument: ast.Document{
				Scope: ast.ScopeSection{
					Span: span(0, 8),
				},
				Definitions: ast.DefinitionsSection{
					Span: span(9, 26),
				},
				Span: span(0, 26),
			},
			wantDiagnostics: []parser.Diagnostic{
				diagnostic("Expected 'identifier', found 'character'.", 23, 26),
			},
		},
		"When a definition is missing an equal sign, a diagnostic is returned.": {
			inSource: "scope {} definitions { letter 'a' }",
			wantDocument: ast.Document{
				Scope: ast.ScopeSection{
					Span: span(0, 8),
				},
				Definitions: ast.DefinitionsSection{
					Definitions: []ast.Definition{
						characterDefinition("letter", 23, 29, token.TokenCharacter, "'a'", 30, 33, 23, 33),
					},
					Span: span(9, 35),
				},
				Span: span(0, 35),
			},
			wantDiagnostics: []parser.Diagnostic{
				diagnostic("Expected '=', found 'character'.", 30, 33),
			},
		},
		"When a definition is missing an expression, a diagnostic is returned.": {
			inSource: "scope {} definitions { letter = }",
			wantDocument: ast.Document{
				Scope: ast.ScopeSection{
					Span: span(0, 8),
				},
				Definitions: ast.DefinitionsSection{
					Definitions: []ast.Definition{
						characterDefinition("letter", 23, 29, token.TokenRightBrace, "}", 32, 33, 23, 33),
					},
					Span: span(9, 33),
				},
				Span: span(0, 33),
			},
			wantDiagnostics: []parser.Diagnostic{
				diagnostic("Expected 'character', found '}'.", 32, 33),
			},
		},
		"When a character range is missing an end character, a diagnostic is returned.": {
			inSource: "scope {} definitions { letter = 'a'.. }",
			wantDocument: ast.Document{
				Scope: ast.ScopeSection{
					Span: span(0, 8),
				},
				Definitions: ast.DefinitionsSection{
					Definitions: []ast.Definition{
						rangeDefinitionWithEndKind("letter", 23, 29, "'a'", 32, 35, token.TokenRightBrace, "}", 38, 39, 23, 39),
					},
					Span: span(9, 39),
				},
				Span: span(0, 39),
			},
			wantDiagnostics: []parser.Diagnostic{
				diagnostic("Expected 'character', found '}'.", 38, 39),
			},
		},
		"When alternation is missing a right expression, a diagnostic is returned.": {
			inSource: "scope {} definitions { value = 'a' | }",
			wantDocument: ast.Document{
				Scope: ast.ScopeSection{
					Span: span(0, 8),
				},
				Definitions: ast.DefinitionsSection{
					Definitions: []ast.Definition{
						alternationDefinition("value", 23, 28, 23, 38, characterExpression(token.TokenCharacter, "'a'", 31, 34), characterExpression(token.TokenRightBrace, "}", 37, 38)),
					},
					Span: span(9, 38),
				},
				Span: span(0, 38),
			},
			wantDiagnostics: []parser.Diagnostic{
				diagnostic("Expected 'character', found '}'.", 37, 38),
			},
		},
		"When a grouped expression is missing a closing parenthesis, a diagnostic is returned.": {
			inSource: "scope {} definitions { value = ('a' | 'b' }",
			wantDocument: ast.Document{
				Scope: ast.ScopeSection{
					Span: span(0, 8),
				},
				Definitions: ast.DefinitionsSection{
					Definitions: []ast.Definition{
						groupDefinition("value", 23, 28, groupExpression(alternationExpression(characterExpression(token.TokenCharacter, "'a'", 32, 35), characterExpression(token.TokenCharacter, "'b'", 38, 41)), 31, 43), 23, 43),
					},
					Span: span(9, 43),
				},
				Span: span(0, 43),
			},
			wantDiagnostics: []parser.Diagnostic{
				diagnostic("Expected ')', found '}'.", 42, 43),
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
