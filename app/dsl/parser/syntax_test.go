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

func Test_Parse_Syntax(t *testing.T) {
	t.Parallel()

	for tcName, tc := range map[string]struct {
		inSource  string
		wantTree  string
		wantDiags string
	}{
		"When parsing an empty syntax block, a document is returned.": {
			inSource: "scope {} syntax {}",
			wantTree: normalizeMultilineLiteral(`
				Document
				  Scope
				  Syntax
			`),
		},
		"When parsing a tokens block followed by an empty syntax block, a document is returned.": {
			inSource: "scope {} tokens {} syntax {}",
			wantTree: normalizeMultilineLiteral(`
				Document
				  Scope
				  Tokens
				  Syntax
			`),
		},
		"When parsing a syntax block with node declarations, a document is returned.": {
			inSource: "scope {} tokens { Identifier = \"id\" KeywordPublic = \"public\" KeywordInternal = \"internal\" } syntax { node Word = Identifier | KeywordPublic node WordPair = Word Word node OptionalWord = (Word | KeywordInternal)? }",
			wantTree: normalizeMultilineLiteral(`
				Document
				  Scope
				  Tokens
				    Token Identifier
				      String "id"
				    Token KeywordPublic
				      String "public"
				    Token KeywordInternal
				      String "internal"
				  Syntax
				    Node Word
				      Alternation
				        Reference Identifier
				        Reference KeywordPublic
				    Node WordPair
				      Concatenation
				        Reference Word
				        Reference Word
				    Node OptionalWord
				      Repetition ?
				        Group
				          Alternation
				            Reference Word
				            Reference KeywordInternal
			`),
		},
		"When the syntax opening brace is missing, a diagnostic is returned.": {
			inSource:  "scope {} syntax }",
			wantDiags: `Expected '{', found '}'. [16:17]`,
		},
		"When the syntax closing brace is missing, a diagnostic is returned.": {
			inSource:  "scope {} syntax {",
			wantDiags: `Expected '}', found 'EOF'. [17:17]`,
		},
		"When an unexpected token appears inside syntax, a diagnostic is returned.": {
			inSource:  "scope {} syntax { FileHeader = Missing }",
			wantDiags: `Expected 'node', found 'identifier'. [18:28]`,
		},
		"When a node declaration is missing its name, a diagnostic is returned.": {
			inSource:  "scope {} syntax { node = Missing }",
			wantDiags: `Expected 'identifier', found '='. [23:24]`,
		},
		"When a node declaration is missing its equal sign, a diagnostic is returned.": {
			inSource:  "scope {} syntax { node Word Missing }",
			wantDiags: `Expected '=', found 'identifier'. [28:35]`,
		},
		"When a node declaration is missing its expression, a diagnostic is returned.": {
			inSource:  "scope {} syntax { node Word = }",
			wantDiags: `Expected 'identifier', found '}'. [30:31]`,
		},
		"When a grouped syntax expression is missing its inner expression, a diagnostic is returned.": {
			inSource:  "scope {} syntax { node Word = ( ) }",
			wantDiags: `Expected 'identifier', found ')'. [32:33]`,
		},
		"When syntax alternation is missing a right expression, a diagnostic is returned.": {
			inSource:  "scope {} syntax { node Word = Missing | }",
			wantDiags: `Expected 'identifier', found '}'. [40:41]`,
		},
		"When a grouped syntax expression is missing a closing parenthesis, a diagnostic is returned.": {
			inSource:  "scope {} syntax { node Word = (Missing | Other? }",
			wantDiags: `Expected ')', found '}'. [48:49]`,
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

func Benchmark_Parse_Syntax_0(b *testing.B)    { benchmark_Parse_Syntax(b, 0) }
func Benchmark_Parse_Syntax_1(b *testing.B)    { benchmark_Parse_Syntax(b, 1) }
func Benchmark_Parse_Syntax_10(b *testing.B)   { benchmark_Parse_Syntax(b, 10) }
func Benchmark_Parse_Syntax_100(b *testing.B)  { benchmark_Parse_Syntax(b, 100) }
func Benchmark_Parse_Syntax_1000(b *testing.B) { benchmark_Parse_Syntax(b, 1000) }

func benchmark_Parse_Syntax(b *testing.B, size int) {
	b.Helper()

	benchmark_Parse(b, syntaxDSL(size))
}

func syntaxDSL(size int) string {
	return "scope {}\n" +
		"tokens {\n" +
		"    Identifier = \"id\"\n" +
		"    KeywordPublic = \"public\"\n" +
		"    KeywordInternal = \"internal\"\n" +
		"}\n" +
		"syntax {\n" +
		strings.Repeat(
			""+
				"    node Word = Identifier | KeywordPublic\n"+
				"    node WordPair = Word Word\n"+
				"    node OptionalWord = (Word | KeywordInternal)?\n"+
				"    node WordList = Word+\n",
			size,
		) +
		"}"
}
