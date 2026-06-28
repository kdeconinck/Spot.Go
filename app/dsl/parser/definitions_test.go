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

func Test_Parse_Definitions(t *testing.T) {
	t.Parallel()

	for tcName, tc := range map[string]struct {
		inSource  string
		wantTree  string
		wantDiags string
	}{
		"When parsing an empty definitions block, a document is returned.": {
			inSource: "scope {} definitions {}",
			wantTree: snapshot(`
				Document
				  Scope
				  Definitions
			`),
		},
		"When parsing a definitions block with a character definition, a document is returned.": {
			inSource: "scope {} definitions { letter = 'a' }",
			wantTree: snapshot(`
				Document
				  Scope
				  Definitions
				    Definition letter
				      Character 'a'
			`),
		},
		"When parsing a definitions block with a character range definition, a document is returned.": {
			inSource: "scope {} definitions { letter = 'a'..'z' }",
			wantTree: snapshot(`
				Document
				  Scope
				  Definitions
				    Definition letter
				      Range 'a' 'z'
			`),
		},
		"When parsing a definitions block with a reference definition, a document is returned.": {
			inSource: "scope {} definitions { identifierStart = letter }",
			wantTree: snapshot(`
				Document
				  Scope
				  Definitions
				    Definition identifierStart
				      Reference letter
			`),
		},
		"When parsing a definitions block with character concatenation, a document is returned.": {
			inSource: "scope {} definitions { value = 'a' 'b' }",
			wantTree: snapshot(`
				Document
				  Scope
				  Definitions
				    Definition value
				      Concatenation
				        Character 'a'
				        Character 'b'
			`),
		},
		"When parsing a definitions block with repeated reference concatenation, a document is returned.": {
			inSource: "scope {} definitions { value = letter digit* }",
			wantTree: snapshot(`
				Document
				  Scope
				  Definitions
				    Definition value
				      Concatenation
				        Reference letter
				        Repetition *
				          Reference digit
			`),
		},
		"When parsing a definitions block with grouped repetition concatenation, a document is returned.": {
			inSource: "scope {} definitions { value = letter ('_' | digit)+ }",
			wantTree: snapshot(`
				Document
				  Scope
				  Definitions
				    Definition value
				      Concatenation
				        Reference letter
				        Repetition +
				          Group
				            Alternation
				              Character '_'
				              Reference digit
			`),
		},
		"When parsing multiple definitions after concatenation, a document is returned.": {
			inSource: "scope {} definitions { letter = 'a' value = letter digit }",
			wantTree: snapshot(`
				Document
				  Scope
				  Definitions
				    Definition letter
				      Character 'a'
				    Definition value
				      Concatenation
				        Reference letter
				        Reference digit
			`),
		},
		"When parsing a definitions block with character alternation, a document is returned.": {
			inSource: "scope {} definitions { value = 'a' | 'b' }",
			wantTree: snapshot(`
				Document
				  Scope
				  Definitions
				    Definition value
				      Alternation
				        Character 'a'
				        Character 'b'
			`),
		},
		"When parsing a definitions block with range alternation, a document is returned.": {
			inSource: "scope {} definitions { letter = 'a'..'z' | 'A'..'Z' }",
			wantTree: snapshot(`
				Document
				  Scope
				  Definitions
				    Definition letter
				      Alternation
				        Range 'a' 'z'
				        Range 'A' 'Z'
			`),
		},
		"When parsing a definitions block with reference alternation, a document is returned.": {
			inSource: "scope {} definitions { identifierStart = letter | '_' }",
			wantTree: snapshot(`
				Document
				  Scope
				  Definitions
				    Definition identifierStart
				      Alternation
				        Reference letter
				        Character '_'
			`),
		},
		"When parsing a definitions block with concatenation before alternation, a document is returned.": {
			inSource: "scope {} definitions { value = letter digit | '_' }",
			wantTree: snapshot(`
				Document
				  Scope
				  Definitions
				    Definition value
				      Alternation
				        Concatenation
				          Reference letter
				          Reference digit
				        Character '_'
			`),
		},
		"When parsing a definitions block with a grouped expression, a document is returned.": {
			inSource: "scope {} definitions { value = ('a' | 'b') }",
			wantTree: snapshot(`
				Document
				  Scope
				  Definitions
				    Definition value
				      Group
				        Alternation
				          Character 'a'
				          Character 'b'
			`),
		},
		"When parsing a definitions block with zero-or-one repetition, a document is returned.": {
			inSource: "scope {} definitions { value = 'a'? }",
			wantTree: snapshot(`
				Document
				  Scope
				  Definitions
				    Definition value
				      Repetition ?
				        Character 'a'
			`),
		},
		"When parsing a definitions block with zero-or-more repetition, a document is returned.": {
			inSource: "scope {} definitions { value = letter* }",
			wantTree: snapshot(`
				Document
				  Scope
				  Definitions
				    Definition value
				      Repetition *
				        Reference letter
			`),
		},
		"When parsing a definitions block with one-or-more repetition, a document is returned.": {
			inSource: "scope {} definitions { value = ('a' | 'b')+ }",
			wantTree: snapshot(`
				Document
				  Scope
				  Definitions
				    Definition value
				      Repetition +
				        Group
				          Alternation
				            Character 'a'
				            Character 'b'
			`),
		},
		"When the definitions opening brace is missing, a diagnostic is returned.": {
			inSource:  "scope {} definitions }",
			wantDiags: `Expected '{', found '}'. [21:22]`,
		},
		"When the definitions closing brace is missing, a diagnostic is returned.": {
			inSource:  "scope {} definitions {",
			wantDiags: `Expected '}', found 'EOF'. [22:22]`,
		},
		"When an unexpected token appears inside definitions, a diagnostic is returned.": {
			inSource:  `scope {} definitions { "x" }`,
			wantDiags: `Expected 'identifier', found 'string'. [23:26]`,
		},
		"When a definition is missing an equal sign, a diagnostic is returned.": {
			inSource:  "scope {} definitions { letter 'a' }",
			wantDiags: `Expected '=', found 'character'. [30:33]`,
		},
		"When a definition is missing an expression, a diagnostic is returned.": {
			inSource:  "scope {} definitions { letter = }",
			wantDiags: `Expected 'character', found '}'. [32:33]`,
		},
		"When a character range is missing an end character, a diagnostic is returned.": {
			inSource:  "scope {} definitions { letter = 'a'.. }",
			wantDiags: `Expected 'character', found '}'. [38:39]`,
		},
		"When alternation is missing a right expression, a diagnostic is returned.": {
			inSource:  "scope {} definitions { value = 'a' | }",
			wantDiags: `Expected 'character', found '}'. [37:38]`,
		},
		"When a grouped expression is missing a closing parenthesis, a diagnostic is returned.": {
			inSource:  "scope {} definitions { value = ('a' | 'b' }",
			wantDiags: `Expected ')', found '}'. [42:43]`,
		},
	} {
		t.Run(tcName, func(t *testing.T) {
			t.Parallel()

			// Act.
			gotDocument, gotDiagnostics := parser.Parse(tc.inSource)

			// Assert.
			claim.Equal(t, tcName, snapshot(tc.wantDiags), debugDiagnostics(gotDiagnostics), "Diagnostics")

			if tc.wantTree != "" {
				claim.Equal(t, tcName, tc.wantTree, debugDocument(gotDocument, false), "Document")
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
