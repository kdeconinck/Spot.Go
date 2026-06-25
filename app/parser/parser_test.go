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

	"github.com/kdeconinck/spot/location"
	"github.com/kdeconinck/spot/parser"
	"github.com/kdeconinck/spot/qa/claim"
	"github.com/kdeconinck/spot/syntax"
)

func Test_Parse(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name            string
		inSource        string
		wantDocument    syntax.Document
		wantDiagnostics []parser.Diagnostic
	}{
		{
			name:     "When parsing an empty scope block, a document is returned.",
			inSource: "scope {}",
			wantDocument: syntax.Document{
				Scope: syntax.ScopeSection{
					Span: span(0, 8),
				},
				Span: span(0, 8),
			},
		},
		{
			name:     "When parsing a scope block with an include entry, a document is returned.",
			inSource: "scope { include \"**/*.go\" }",
			wantDocument: syntax.Document{
				Scope: syntax.ScopeSection{
					Entries: []syntax.ScopeEntry{
						scopeEntry(syntax.ScopeEntryInclude, syntax.TokenString, "\"**/*.go\"", 16, 25, 8, 25),
					},
					Span: span(0, 27),
				},
				Span: span(0, 27),
			},
		},
		{
			name:     "When parsing a scope block with an exclude entry, a document is returned.",
			inSource: "scope { exclude \"vendor/**\" }",
			wantDocument: syntax.Document{
				Scope: syntax.ScopeSection{
					Entries: []syntax.ScopeEntry{
						scopeEntry(syntax.ScopeEntryExclude, syntax.TokenString, "\"vendor/**\"", 16, 27, 8, 27),
					},
					Span: span(0, 29),
				},
				Span: span(0, 29),
			},
		},
		{
			name:     "When parsing a scope block with include and exclude entries, a document is returned.",
			inSource: "scope { include \"**/*.go\" exclude \"vendor/**\" }",
			wantDocument: syntax.Document{
				Scope: syntax.ScopeSection{
					Entries: []syntax.ScopeEntry{
						scopeEntry(syntax.ScopeEntryInclude, syntax.TokenString, "\"**/*.go\"", 16, 25, 8, 25),
						scopeEntry(syntax.ScopeEntryExclude, syntax.TokenString, "\"vendor/**\"", 34, 45, 26, 45),
					},
					Span: span(0, 47),
				},
				Span: span(0, 47),
			},
		},
		{
			name:     "When parsing an empty definitions block, a document is returned.",
			inSource: "scope {} definitions {}",
			wantDocument: syntax.Document{
				Scope: syntax.ScopeSection{
					Span: span(0, 8),
				},
				Definitions: syntax.DefinitionsSection{
					Span: span(9, 23),
				},
				Span: span(0, 23),
			},
		},
		{
			name:     "When parsing a definitions block with a character definition, a document is returned.",
			inSource: "scope {} definitions { letter = 'a' }",
			wantDocument: syntax.Document{
				Scope: syntax.ScopeSection{
					Span: span(0, 8),
				},
				Definitions: syntax.DefinitionsSection{
					Definitions: []syntax.Definition{
						characterDefinition("letter", 23, 29, syntax.TokenCharacter, "'a'", 32, 35, 23, 35),
					},
					Span: span(9, 37),
				},
				Span: span(0, 37),
			},
		},
		{
			name:     "When parsing a definitions block with a character range definition, a document is returned.",
			inSource: "scope {} definitions { letter = 'a'..'z' }",
			wantDocument: syntax.Document{
				Scope: syntax.ScopeSection{
					Span: span(0, 8),
				},
				Definitions: syntax.DefinitionsSection{
					Definitions: []syntax.Definition{
						rangeDefinition("letter", 23, 29, "'a'", 32, 35, "'z'", 37, 40, 23, 40),
					},
					Span: span(9, 42),
				},
				Span: span(0, 42),
			},
		},
		{
			name:     "When parsing a definitions block with a reference definition, a document is returned.",
			inSource: "scope {} definitions { identifierStart = letter }",
			wantDocument: syntax.Document{
				Scope: syntax.ScopeSection{
					Span: span(0, 8),
				},
				Definitions: syntax.DefinitionsSection{
					Definitions: []syntax.Definition{
						referenceDefinition("identifierStart", 23, 38, "letter", 41, 47, 23, 47),
					},
					Span: span(9, 49),
				},
				Span: span(0, 49),
			},
		},
		{
			name:     "When parsing a definitions block with character concatenation, a document is returned.",
			inSource: "scope {} definitions { value = 'a' 'b' }",
			wantDocument: syntax.Document{
				Scope: syntax.ScopeSection{
					Span: span(0, 8),
				},
				Definitions: syntax.DefinitionsSection{
					Definitions: []syntax.Definition{
						concatenationDefinition("value", 23, 28, 23, 38, characterExpression(syntax.TokenCharacter, "'a'", 31, 34), characterExpression(syntax.TokenCharacter, "'b'", 35, 38)),
					},
					Span: span(9, 40),
				},
				Span: span(0, 40),
			},
		},
		{
			name:     "When parsing a definitions block with repeated reference concatenation, a document is returned.",
			inSource: "scope {} definitions { value = letter digit* }",
			wantDocument: syntax.Document{
				Scope: syntax.ScopeSection{
					Span: span(0, 8),
				},
				Definitions: syntax.DefinitionsSection{
					Definitions: []syntax.Definition{
						concatenationDefinition("value", 23, 28, 23, 44, referenceExpression("letter", 31, 37), repetitionExpression(referenceExpression("digit", 38, 43), syntax.TokenStar, "*", 43, 44)),
					},
					Span: span(9, 46),
				},
				Span: span(0, 46),
			},
		},
		{
			name:     "When parsing a definitions block with grouped repetition concatenation, a document is returned.",
			inSource: "scope {} definitions { value = letter ('_' | digit)+ }",
			wantDocument: syntax.Document{
				Scope: syntax.ScopeSection{
					Span: span(0, 8),
				},
				Definitions: syntax.DefinitionsSection{
					Definitions: []syntax.Definition{
						concatenationDefinition("value", 23, 28, 23, 52, referenceExpression("letter", 31, 37), repetitionExpression(groupExpression(alternationExpression(characterExpression(syntax.TokenCharacter, "'_'", 39, 42), referenceExpression("digit", 45, 50)), 38, 51), syntax.TokenPlus, "+", 51, 52)),
					},
					Span: span(9, 54),
				},
				Span: span(0, 54),
			},
		},
		{
			name:     "When parsing multiple definitions after concatenation, a document is returned.",
			inSource: "scope {} definitions { letter = 'a' value = letter digit }",
			wantDocument: syntax.Document{
				Scope: syntax.ScopeSection{
					Span: span(0, 8),
				},
				Definitions: syntax.DefinitionsSection{
					Definitions: []syntax.Definition{
						characterDefinition("letter", 23, 29, syntax.TokenCharacter, "'a'", 32, 35, 23, 35),
						concatenationDefinition("value", 36, 41, 36, 56, referenceExpression("letter", 44, 50), referenceExpression("digit", 51, 56)),
					},
					Span: span(9, 58),
				},
				Span: span(0, 58),
			},
		},
		{
			name:     "When parsing a definitions block with character alternation, a document is returned.",
			inSource: "scope {} definitions { value = 'a' | 'b' }",
			wantDocument: syntax.Document{
				Scope: syntax.ScopeSection{
					Span: span(0, 8),
				},
				Definitions: syntax.DefinitionsSection{
					Definitions: []syntax.Definition{
						alternationDefinition("value", 23, 28, 23, 40, characterExpression(syntax.TokenCharacter, "'a'", 31, 34), characterExpression(syntax.TokenCharacter, "'b'", 37, 40)),
					},
					Span: span(9, 42),
				},
				Span: span(0, 42),
			},
		},
		{
			name:     "When parsing a definitions block with range alternation, a document is returned.",
			inSource: "scope {} definitions { letter = 'a'..'z' | 'A'..'Z' }",
			wantDocument: syntax.Document{
				Scope: syntax.ScopeSection{
					Span: span(0, 8),
				},
				Definitions: syntax.DefinitionsSection{
					Definitions: []syntax.Definition{
						alternationDefinition("letter", 23, 29, 23, 51, rangeExpression("'a'", 32, 35, syntax.TokenCharacter, "'z'", 37, 40), rangeExpression("'A'", 43, 46, syntax.TokenCharacter, "'Z'", 48, 51)),
					},
					Span: span(9, 53),
				},
				Span: span(0, 53),
			},
		},
		{
			name:     "When parsing a definitions block with reference alternation, a document is returned.",
			inSource: "scope {} definitions { identifierStart = letter | '_' }",
			wantDocument: syntax.Document{
				Scope: syntax.ScopeSection{
					Span: span(0, 8),
				},
				Definitions: syntax.DefinitionsSection{
					Definitions: []syntax.Definition{
						alternationDefinition("identifierStart", 23, 38, 23, 53, referenceExpression("letter", 41, 47), characterExpression(syntax.TokenCharacter, "'_'", 50, 53)),
					},
					Span: span(9, 55),
				},
				Span: span(0, 55),
			},
		},
		{
			name:     "When parsing a definitions block with concatenation before alternation, a document is returned.",
			inSource: "scope {} definitions { value = letter digit | '_' }",
			wantDocument: syntax.Document{
				Scope: syntax.ScopeSection{
					Span: span(0, 8),
				},
				Definitions: syntax.DefinitionsSection{
					Definitions: []syntax.Definition{
						alternationDefinition("value", 23, 28, 23, 49, concatenationExpression(referenceExpression("letter", 31, 37), referenceExpression("digit", 38, 43)), characterExpression(syntax.TokenCharacter, "'_'", 46, 49)),
					},
					Span: span(9, 51),
				},
				Span: span(0, 51),
			},
		},
		{
			name:     "When parsing a definitions block with a grouped expression, a document is returned.",
			inSource: "scope {} definitions { value = ('a' | 'b') }",
			wantDocument: syntax.Document{
				Scope: syntax.ScopeSection{
					Span: span(0, 8),
				},
				Definitions: syntax.DefinitionsSection{
					Definitions: []syntax.Definition{
						groupDefinition("value", 23, 28, groupExpression(alternationExpression(characterExpression(syntax.TokenCharacter, "'a'", 32, 35), characterExpression(syntax.TokenCharacter, "'b'", 38, 41)), 31, 42), 23, 42),
					},
					Span: span(9, 44),
				},
				Span: span(0, 44),
			},
		},
		{
			name:     "When parsing a definitions block with zero-or-one repetition, a document is returned.",
			inSource: "scope {} definitions { value = 'a'? }",
			wantDocument: syntax.Document{
				Scope: syntax.ScopeSection{
					Span: span(0, 8),
				},
				Definitions: syntax.DefinitionsSection{
					Definitions: []syntax.Definition{
						repetitionDefinition("value", 23, 28, repetitionExpression(characterExpression(syntax.TokenCharacter, "'a'", 31, 34), syntax.TokenQuestion, "?", 34, 35), 23, 35),
					},
					Span: span(9, 37),
				},
				Span: span(0, 37),
			},
		},
		{
			name:     "When parsing a definitions block with zero-or-more repetition, a document is returned.",
			inSource: "scope {} definitions { value = letter* }",
			wantDocument: syntax.Document{
				Scope: syntax.ScopeSection{
					Span: span(0, 8),
				},
				Definitions: syntax.DefinitionsSection{
					Definitions: []syntax.Definition{
						repetitionDefinition("value", 23, 28, repetitionExpression(referenceExpression("letter", 31, 37), syntax.TokenStar, "*", 37, 38), 23, 38),
					},
					Span: span(9, 40),
				},
				Span: span(0, 40),
			},
		},
		{
			name:     "When parsing a definitions block with one-or-more repetition, a document is returned.",
			inSource: "scope {} definitions { value = ('a' | 'b')+ }",
			wantDocument: syntax.Document{
				Scope: syntax.ScopeSection{
					Span: span(0, 8),
				},
				Definitions: syntax.DefinitionsSection{
					Definitions: []syntax.Definition{
						repetitionDefinition("value", 23, 28, repetitionExpression(groupExpression(alternationExpression(characterExpression(syntax.TokenCharacter, "'a'", 32, 35), characterExpression(syntax.TokenCharacter, "'b'", 38, 41)), 31, 42), syntax.TokenPlus, "+", 42, 43), 23, 43),
					},
					Span: span(9, 45),
				},
				Span: span(0, 45),
			},
		},
		{
			name:     "When the scope keyword is missing, a diagnostic is returned.",
			inSource: "x",
			wantDocument: syntax.Document{
				Scope: syntax.ScopeSection{
					Span: span(0, 1),
				},
				Span: span(0, 1),
			},
			wantDiagnostics: []parser.Diagnostic{
				diagnostic("Expected 'scope', found 'identifier'.", 0, 1),
			},
		},
		{
			name:     "When the opening brace is missing, a diagnostic is returned.",
			inSource: "scope }",
			wantDocument: syntax.Document{
				Scope: syntax.ScopeSection{
					Span: span(0, 5),
				},
				Span: span(0, 5),
			},
			wantDiagnostics: []parser.Diagnostic{
				diagnostic("Expected '{', found '}'.", 6, 7),
			},
		},
		{
			name:     "When the closing brace is missing, a diagnostic is returned.",
			inSource: "scope {",
			wantDocument: syntax.Document{
				Scope: syntax.ScopeSection{
					Span: span(0, 7),
				},
				Span: span(0, 7),
			},
			wantDiagnostics: []parser.Diagnostic{
				diagnostic("Expected '}', found 'EOF'.", 7, 7),
			},
		},
		{
			name:     "When an include entry has no string, a diagnostic is returned.",
			inSource: "scope { include }",
			wantDocument: syntax.Document{
				Scope: syntax.ScopeSection{
					Entries: []syntax.ScopeEntry{
						scopeEntry(syntax.ScopeEntryInclude, syntax.TokenRightBrace, "}", 16, 17, 8, 17),
					},
					Span: span(0, 17),
				},
				Span: span(0, 17),
			},
			wantDiagnostics: []parser.Diagnostic{
				diagnostic("Expected 'string', found '}'.", 16, 17),
			},
		},
		{
			name:     "When an unexpected token appears inside scope, a diagnostic is returned.",
			inSource: "scope { x }",
			wantDocument: syntax.Document{
				Scope: syntax.ScopeSection{
					Span: span(0, 9),
				},
				Span: span(0, 9),
			},
			wantDiagnostics: []parser.Diagnostic{
				diagnostic("Expected 'include', found 'identifier'.", 8, 9),
			},
		},
		{
			name:     "When a token appears after the scope block, a diagnostic is returned.",
			inSource: "scope {} x",
			wantDocument: syntax.Document{
				Scope: syntax.ScopeSection{
					Span: span(0, 8),
				},
				Span: span(0, 10),
			},
			wantDiagnostics: []parser.Diagnostic{
				diagnostic("Expected 'EOF', found 'identifier'.", 9, 10),
			},
		},
		{
			name:     "When the definitions opening brace is missing, a diagnostic is returned.",
			inSource: "scope {} definitions }",
			wantDocument: syntax.Document{
				Scope: syntax.ScopeSection{
					Span: span(0, 8),
				},
				Definitions: syntax.DefinitionsSection{
					Span: span(9, 20),
				},
				Span: span(0, 20),
			},
			wantDiagnostics: []parser.Diagnostic{
				diagnostic("Expected '{', found '}'.", 21, 22),
			},
		},
		{
			name:     "When the definitions closing brace is missing, a diagnostic is returned.",
			inSource: "scope {} definitions {",
			wantDocument: syntax.Document{
				Scope: syntax.ScopeSection{
					Span: span(0, 8),
				},
				Definitions: syntax.DefinitionsSection{
					Span: span(9, 22),
				},
				Span: span(0, 22),
			},
			wantDiagnostics: []parser.Diagnostic{
				diagnostic("Expected '}', found 'EOF'.", 22, 22),
			},
		},
		{
			name:     "When an unexpected token appears inside definitions, a diagnostic is returned.",
			inSource: "scope {} definitions { 'a' }",
			wantDocument: syntax.Document{
				Scope: syntax.ScopeSection{
					Span: span(0, 8),
				},
				Definitions: syntax.DefinitionsSection{
					Span: span(9, 26),
				},
				Span: span(0, 26),
			},
			wantDiagnostics: []parser.Diagnostic{
				diagnostic("Expected 'identifier', found 'character'.", 23, 26),
			},
		},
		{
			name:     "When a definition is missing an equal sign, a diagnostic is returned.",
			inSource: "scope {} definitions { letter 'a' }",
			wantDocument: syntax.Document{
				Scope: syntax.ScopeSection{
					Span: span(0, 8),
				},
				Definitions: syntax.DefinitionsSection{
					Definitions: []syntax.Definition{
						characterDefinition("letter", 23, 29, syntax.TokenCharacter, "'a'", 30, 33, 23, 33),
					},
					Span: span(9, 35),
				},
				Span: span(0, 35),
			},
			wantDiagnostics: []parser.Diagnostic{
				diagnostic("Expected '=', found 'character'.", 30, 33),
			},
		},
		{
			name:     "When a definition is missing an expression, a diagnostic is returned.",
			inSource: "scope {} definitions { letter = }",
			wantDocument: syntax.Document{
				Scope: syntax.ScopeSection{
					Span: span(0, 8),
				},
				Definitions: syntax.DefinitionsSection{
					Definitions: []syntax.Definition{
						characterDefinition("letter", 23, 29, syntax.TokenRightBrace, "}", 32, 33, 23, 33),
					},
					Span: span(9, 33),
				},
				Span: span(0, 33),
			},
			wantDiagnostics: []parser.Diagnostic{
				diagnostic("Expected 'character', found '}'.", 32, 33),
			},
		},
		{
			name:     "When a character range is missing an end character, a diagnostic is returned.",
			inSource: "scope {} definitions { letter = 'a'.. }",
			wantDocument: syntax.Document{
				Scope: syntax.ScopeSection{
					Span: span(0, 8),
				},
				Definitions: syntax.DefinitionsSection{
					Definitions: []syntax.Definition{
						rangeDefinitionWithEndKind("letter", 23, 29, "'a'", 32, 35, syntax.TokenRightBrace, "}", 38, 39, 23, 39),
					},
					Span: span(9, 39),
				},
				Span: span(0, 39),
			},
			wantDiagnostics: []parser.Diagnostic{
				diagnostic("Expected 'character', found '}'.", 38, 39),
			},
		},
		{
			name:     "When alternation is missing a right expression, a diagnostic is returned.",
			inSource: "scope {} definitions { value = 'a' | }",
			wantDocument: syntax.Document{
				Scope: syntax.ScopeSection{
					Span: span(0, 8),
				},
				Definitions: syntax.DefinitionsSection{
					Definitions: []syntax.Definition{
						alternationDefinition("value", 23, 28, 23, 38, characterExpression(syntax.TokenCharacter, "'a'", 31, 34), characterExpression(syntax.TokenRightBrace, "}", 37, 38)),
					},
					Span: span(9, 38),
				},
				Span: span(0, 38),
			},
			wantDiagnostics: []parser.Diagnostic{
				diagnostic("Expected 'character', found '}'.", 37, 38),
			},
		},
		{
			name:     "When a grouped expression is missing a closing parenthesis, a diagnostic is returned.",
			inSource: "scope {} definitions { value = ('a' | 'b' }",
			wantDocument: syntax.Document{
				Scope: syntax.ScopeSection{
					Span: span(0, 8),
				},
				Definitions: syntax.DefinitionsSection{
					Definitions: []syntax.Definition{
						groupDefinition("value", 23, 28, groupExpression(alternationExpression(characterExpression(syntax.TokenCharacter, "'a'", 32, 35), characterExpression(syntax.TokenCharacter, "'b'", 38, 41)), 31, 43), 23, 43),
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
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Act.
			gotDocument, gotDiagnostics := parser.Parse(tc.inSource)

			// Assert.
			claim.DeepEqual(t, tc.name, tc.wantDocument, gotDocument, "Document")
			claim.Equal(t, tc.name, len(tc.wantDiagnostics), len(gotDiagnostics), "Diagnostic Count")

			for idx := range tc.wantDiagnostics {
				claim.Equal(t, tc.name, tc.wantDiagnostics[idx], gotDiagnostics[idx], "Diagnostic")
			}
		})
	}
}

func benchmark_Parse(b *testing.B, source string) {
	b.Helper()

	for b.Loop() {
		_, _ = parser.Parse(source)
	}
}

func Benchmark_Parse_DSL_0(b *testing.B)    { benchmark_Parse_DSL(b, 0) }
func Benchmark_Parse_DSL_1(b *testing.B)    { benchmark_Parse_DSL(b, 1) }
func Benchmark_Parse_DSL_10(b *testing.B)   { benchmark_Parse_DSL(b, 10) }
func Benchmark_Parse_DSL_100(b *testing.B)  { benchmark_Parse_DSL(b, 100) }
func Benchmark_Parse_DSL_1000(b *testing.B) { benchmark_Parse_DSL(b, 1000) }

func benchmark_Parse_DSL(b *testing.B, size int) {
	b.Helper()

	benchmark_Parse(b, dsl(size))
}

func dsl(size int) string {
	return "scope {\n" +
		strings.Repeat("    include \"**/*.go\"\n    exclude \"vendor/**\"\n", size) +
		"}\n" +
		"definitions {\n" +
		strings.Repeat("    letter = 'a'..'z' | 'A'..'Z'\n    identifierStart = letter | '_'\n    value = letter ('a' | 'b')+\n", size) +
		"}"
}

func diagnostic(message string, start, end location.Position) parser.Diagnostic {
	return parser.Diagnostic{
		Message: message,
		Span:    span(start, end),
	}
}

func span(start, end location.Position) location.Span {
	return location.Span{
		Start: start,
		End:   end,
	}
}

func scopeEntry(kind syntax.ScopeEntryKind, patternKind syntax.TokenKind, pattern string, patternStart, patternEnd, entryStart, entryEnd location.Position) syntax.ScopeEntry {
	return syntax.ScopeEntry{
		Kind: kind,
		Pattern: syntax.Token{
			Kind: patternKind,
			Text: pattern,
			Span: span(patternStart, patternEnd),
		},
		Span: span(entryStart, entryEnd),
	}
}

func characterDefinition(name string, nameStart, nameEnd location.Position, expressionKind syntax.TokenKind, expression string, expressionStart, expressionEnd, definitionStart, definitionEnd location.Position) syntax.Definition {
	return syntax.Definition{
		Name: syntax.Token{
			Kind: syntax.TokenIdentifier,
			Text: name,
			Span: span(nameStart, nameEnd),
		},
		Expression: characterExpression(expressionKind, expression, expressionStart, expressionEnd),
		Span:       span(definitionStart, definitionEnd),
	}
}

func rangeDefinition(name string, nameStart, nameEnd location.Position, start string, startStart, startEnd location.Position, end string, endStart, endEnd, definitionStart, definitionEnd location.Position) syntax.Definition {
	return rangeDefinitionWithEndKind(name, nameStart, nameEnd, start, startStart, startEnd, syntax.TokenCharacter, end, endStart, endEnd, definitionStart, definitionEnd)
}

func rangeDefinitionWithEndKind(name string, nameStart, nameEnd location.Position, start string, startStart, startEnd location.Position, endKind syntax.TokenKind, end string, endStart, endEnd, definitionStart, definitionEnd location.Position) syntax.Definition {
	return syntax.Definition{
		Name: syntax.Token{
			Kind: syntax.TokenIdentifier,
			Text: name,
			Span: span(nameStart, nameEnd),
		},
		Expression: rangeExpression(start, startStart, startEnd, endKind, end, endStart, endEnd),
		Span:       span(definitionStart, definitionEnd),
	}
}

func referenceDefinition(name string, nameStart, nameEnd location.Position, reference string, referenceStart, referenceEnd, definitionStart, definitionEnd location.Position) syntax.Definition {
	return syntax.Definition{
		Name: syntax.Token{
			Kind: syntax.TokenIdentifier,
			Text: name,
			Span: span(nameStart, nameEnd),
		},
		Expression: referenceExpression(reference, referenceStart, referenceEnd),
		Span:       span(definitionStart, definitionEnd),
	}
}

func groupDefinition(name string, nameStart, nameEnd location.Position, expression syntax.DefinitionExpression, definitionStart, definitionEnd location.Position) syntax.Definition {
	return syntax.Definition{
		Name: syntax.Token{
			Kind: syntax.TokenIdentifier,
			Text: name,
			Span: span(nameStart, nameEnd),
		},
		Expression: expression,
		Span:       span(definitionStart, definitionEnd),
	}
}

func repetitionDefinition(name string, nameStart, nameEnd location.Position, expression syntax.DefinitionExpression, definitionStart, definitionEnd location.Position) syntax.Definition {
	return syntax.Definition{
		Name: syntax.Token{
			Kind: syntax.TokenIdentifier,
			Text: name,
			Span: span(nameStart, nameEnd),
		},
		Expression: expression,
		Span:       span(definitionStart, definitionEnd),
	}
}

func concatenationDefinition(name string, nameStart, nameEnd, definitionStart, definitionEnd location.Position, terms ...syntax.DefinitionExpression) syntax.Definition {
	return syntax.Definition{
		Name: syntax.Token{
			Kind: syntax.TokenIdentifier,
			Text: name,
			Span: span(nameStart, nameEnd),
		},
		Expression: concatenationExpression(terms...),
		Span:       span(definitionStart, definitionEnd),
	}
}

func alternationDefinition(name string, nameStart, nameEnd, definitionStart, definitionEnd location.Position, terms ...syntax.DefinitionExpression) syntax.Definition {
	return syntax.Definition{
		Name: syntax.Token{
			Kind: syntax.TokenIdentifier,
			Text: name,
			Span: span(nameStart, nameEnd),
		},
		Expression: alternationExpression(terms...),
		Span:       span(definitionStart, definitionEnd),
	}
}

func concatenationExpression(terms ...syntax.DefinitionExpression) syntax.DefinitionExpression {
	return syntax.DefinitionExpression{
		Kind:  syntax.DefinitionExpressionConcatenation,
		Terms: terms,
		Span:  span(terms[0].Span.Start, terms[len(terms)-1].Span.End),
	}
}

func alternationExpression(terms ...syntax.DefinitionExpression) syntax.DefinitionExpression {
	return syntax.DefinitionExpression{
		Kind:  syntax.DefinitionExpressionAlternation,
		Terms: terms,
		Span:  span(terms[0].Span.Start, terms[len(terms)-1].Span.End),
	}
}

func characterExpression(kind syntax.TokenKind, text string, start, end location.Position) syntax.DefinitionExpression {
	return syntax.DefinitionExpression{
		Kind: syntax.DefinitionExpressionCharacter,
		Start: syntax.Token{
			Kind: kind,
			Text: text,
			Span: span(start, end),
		},
		Span: span(start, end),
	}
}

func groupExpression(inner syntax.DefinitionExpression, start, end location.Position) syntax.DefinitionExpression {
	return syntax.DefinitionExpression{
		Kind:  syntax.DefinitionExpressionGroup,
		Inner: &inner,
		Span:  span(start, end),
	}
}

func repetitionExpression(inner syntax.DefinitionExpression, operatorKind syntax.TokenKind, operator string, operatorStart, operatorEnd location.Position) syntax.DefinitionExpression {
	return syntax.DefinitionExpression{
		Kind: syntax.DefinitionExpressionRepetition,
		Operator: syntax.Token{
			Kind: operatorKind,
			Text: operator,
			Span: span(operatorStart, operatorEnd),
		},
		Inner: &inner,
		Span:  span(inner.Span.Start, operatorEnd),
	}
}

func referenceExpression(text string, start, end location.Position) syntax.DefinitionExpression {
	return syntax.DefinitionExpression{
		Kind: syntax.DefinitionExpressionReference,
		Start: syntax.Token{
			Kind: syntax.TokenIdentifier,
			Text: text,
			Span: span(start, end),
		},
		Span: span(start, end),
	}
}

func rangeExpression(start string, startStart, startEnd location.Position, endKind syntax.TokenKind, end string, endStart, endEnd location.Position) syntax.DefinitionExpression {
	return syntax.DefinitionExpression{
		Kind: syntax.DefinitionExpressionRange,
		Start: syntax.Token{
			Kind: syntax.TokenCharacter,
			Text: start,
			Span: span(startStart, startEnd),
		},
		End: syntax.Token{
			Kind: endKind,
			Text: end,
			Span: span(endStart, endEnd),
		},
		Span: span(startStart, endEnd),
	}
}
