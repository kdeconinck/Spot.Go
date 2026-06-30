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

	"github.com/kdeconinck/spot/dsl/parser"
	"github.com/kdeconinck/spot/qa/claim"
)

func Test_Parse_Tokens(t *testing.T) {
	t.Parallel()

	for tcName, tc := range map[string]struct {
		inSource  string
		wantTree  string
		wantDiags string
	}{
		"When parsing an empty tokens block, a document is returned.": {
			inSource: "scope {} tokens {}",
			wantTree: normalizeMultilineLiteral(`
				Document
				  Scope
				  Tokens
			`),
		},
		"When parsing a definitions block followed by an empty tokens block, a document is returned.": {
			inSource: "scope {} definitions {} tokens {}",
			wantTree: normalizeMultilineLiteral(`
				Document
				  Scope
				  Definitions
				  Tokens
			`),
		},
		"When parsing a tokens block with a string token, a document is returned.": {
			inSource: `scope {} tokens { KeywordPublic = "public" }`,
			wantTree: normalizeMultilineLiteral(`
				Document
				  Scope
				  Tokens
				    Token KeywordPublic
				      String "public"
			`),
		},
		"When parsing a tokens block with a reference token, a document is returned.": {
			inSource: "scope {} tokens { Identifier = letter }",
			wantTree: normalizeMultilineLiteral(`
				Document
				  Scope
				  Tokens
				    Token Identifier
				      Reference letter
			`),
		},
		"When parsing a tokens block with a character token, a document is returned.": {
			inSource: "scope {} tokens { Plus = '+' }",
			wantTree: normalizeMultilineLiteral(`
				Document
				  Scope
				  Tokens
				    Token Plus
				      Character '+'
			`),
		},
		"When parsing a tokens block with a range token, a document is returned.": {
			inSource: "scope {} tokens { Lower = 'a'..'z' }",
			wantTree: normalizeMultilineLiteral(`
				Document
				  Scope
				  Tokens
				    Token Lower
				      Range 'a' 'z'
			`),
		},
		"When parsing a tokens block with string concatenation, a document is returned.": {
			inSource: `scope {} tokens { Keyword = "public" "static" }`,
			wantTree: normalizeMultilineLiteral(`
				Document
				  Scope
				  Tokens
				    Token Keyword
				      Concatenation
				        String "public"
				        String "static"
			`),
		},
		"When parsing a tokens block with alternation, a document is returned.": {
			inSource: `scope {} tokens { Sign = "+" | "-" }`,
			wantTree: normalizeMultilineLiteral(`
				Document
				  Scope
				  Tokens
				    Token Sign
				      Alternation
				        String "+"
				        String "-"
			`),
		},
		"When parsing a tokens block with zero-or-one repetition, a document is returned.": {
			inSource: `scope {} tokens { OptionalSign = ("+" | "-")? }`,
			wantTree: normalizeMultilineLiteral(`
				Document
				  Scope
				  Tokens
				    Token OptionalSign
				      Repetition ?
				        Group
				          Alternation
				            String "+"
				            String "-"
			`),
		},
		"When parsing a tokens block with a skipped token, a document is returned.": {
			inSource: `scope {} tokens { Whitespace = (' ' | '\t')+ skip }`,
			wantTree: normalizeMultilineLiteral(`
				Document
				  Scope
				  Tokens
				    Token Whitespace
				      Repetition +
				        Group
				          Alternation
				            Character ' '
				            Character '\t'
				      Skip
			`),
		},
		"When the tokens opening brace is missing, a diagnostic is returned.": {
			inSource:  "scope {} tokens }",
			wantDiags: `Expected '{', found '}'. [16:17]`,
		},
		"When the tokens closing brace is missing, a diagnostic is returned.": {
			inSource:  "scope {} tokens {",
			wantDiags: `Expected '}', found 'EOF'. [17:17]`,
		},
		"When an unexpected token appears inside tokens, a diagnostic is returned.": {
			inSource:  "scope {} tokens { 'a' }",
			wantDiags: `Expected 'identifier', found 'character'. [18:21]`,
		},
		"When a token is missing an equal sign, a diagnostic is returned.": {
			inSource:  `scope {} tokens { Identifier "public" }`,
			wantDiags: `Expected '=', found 'string'. [29:37]`,
		},
		"When a token is missing an expression, a diagnostic is returned.": {
			inSource:  "scope {} tokens { Identifier = }",
			wantDiags: `Expected 'character', found '}'. [31:32]`,
		},
		"When a token grouped expression is missing its inner expression, a diagnostic is returned.": {
			inSource:  `scope {} tokens { OptionalSign = ( ) }`,
			wantDiags: `Expected 'character', found ')'. [35:36]`,
		},
		"When a token range is missing an end character, a diagnostic is returned.": {
			inSource:  "scope {} tokens { Lower = 'a'.. }",
			wantDiags: `Expected 'character', found '}'. [32:33]`,
		},
		"When a token concatenation is missing a valid right expression, a diagnostic is returned.": {
			inSource:  `scope {} tokens { Pair = "a" ( }`,
			wantDiags: `Expected 'character', found '}'. [31:32]`,
		},
		"When token alternation is missing a right expression, a diagnostic is returned.": {
			inSource:  `scope {} tokens { Sign = "+" | }`,
			wantDiags: `Expected 'character', found '}'. [31:32]`,
		},
		"When a token grouped expression is missing a closing parenthesis, a diagnostic is returned.": {
			inSource:  `scope {} tokens { OptionalSign = ("+" | "-"? }`,
			wantDiags: `Expected ')', found '}'. [45:46]`,
		},
	} {
		t.Run(tcName, func(t *testing.T) {
			t.Parallel()

			// Act.
			gotDocument, gotErr := parser.Parse(tc.inSource)

			// Assert.
			claim.Equal(t, tcName, normalizeMultilineLiteral(tc.wantDiags), formatParseError(gotErr), "Parse Error")

			if tc.wantTree != "" {
				claim.Equal(t, tcName, tc.wantTree, renderDocument(tc.inSource, gotDocument, false), "Document")
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
		"    lower = 'a'..'z'\n" +
		"    upper = 'A'..'Z'\n" +
		"    digit = '0'..'9'\n" +
		"    underscore = '_'\n" +
		"    letter = lower | upper\n" +
		"    identifierStart = letter | underscore\n" +
		"    identifierPart = letter | digit | underscore\n" +
		"    optionalSign = ('+' | '-')?\n" +
		"}\n" +
		"tokens {\n" +
		strings.Repeat(
			""+
				"    Plus = '+'\n"+
				"    Lower = 'a'..'z'\n"+
				"    Identifier = identifierStart identifierPart*\n"+
				"    KeywordPublic = \"public\"\n"+
				"    Sign = \"+\" | \"-\"\n"+
				"    OptionalSign = (\"+\" | \"-\")?\n"+
				"    SignedInteger = optionalSign digit+\n"+
				"    Whitespace = (' ' | '\\t' | '\\n' | '\\r')+ skip\n",
			size,
		) +
		"}"
}
