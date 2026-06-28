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
			wantTree: snapshot(`
				Document
				  Scope
				  Tokens
			`),
		},
		"When parsing a definitions block followed by an empty tokens block, a document is returned.": {
			inSource: "scope {} definitions {} tokens {}",
			wantTree: snapshot(`
				Document
				  Scope
				  Definitions
				  Tokens
			`),
		},
		"When parsing a tokens block with a reference token, a document is returned.": {
			inSource: "scope {} tokens { Identifier = letter }",
			wantTree: snapshot(`
				Document
				  Scope
				  Tokens
				    Token Identifier
				      Reference letter
			`),
		},
		"When parsing a tokens block with a string token, a document is returned.": {
			inSource: `scope {} tokens { KeywordPublic = "public" }`,
			wantTree: snapshot(`
				Document
				  Scope
				  Tokens
				    Token KeywordPublic
				      String "public"
			`),
		},
		"When parsing a tokens block with a skipped token, a document is returned.": {
			inSource: `scope {} tokens { Whitespace = (' ' | '\t')+ skip }`,
			wantTree: snapshot(`
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
		"When a token is missing an expression, a diagnostic is returned.": {
			inSource:  "scope {} tokens { Identifier = }",
			wantDiags: `Expected 'character', found '}'. [31:32]`,
		},
	} {
		t.Run(tcName, func(t *testing.T) {
			t.Parallel()

			// Act.
			gotDocument, gotDiagnostics := parser.Parse(tc.inSource)

			// Assert.
			claim.Equal(t, tcName, snapshot(tc.wantDiags), debugDiagnostics(gotDiagnostics), "Diagnostics")

			if tc.wantTree != "" {
				claim.Equal(t, tcName, tc.wantTree, debugDocument(tc.inSource, gotDocument, false), "Document")
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
