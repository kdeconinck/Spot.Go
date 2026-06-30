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

func Test_Parse_Rules(t *testing.T) {
	t.Parallel()

	for tcName, tc := range map[string]struct {
		inSource  string
		wantTree  string
		wantDiags string
	}{
		"When parsing an empty rules block, a document is returned.": {
			inSource: "scope {} rules {}",
			wantTree: normalizeMultilineLiteral(`
				Document
				  Scope
				  Rules
			`),
		},
		"When parsing a rules block with a rule, a document is returned.": {
			inSource: `scope {} rules { rule PublicIdentifier { match Identifier report warn at Identifier "Public identifier found" } }`,
			wantTree: normalizeMultilineLiteral(`
				Document
				  Scope
				  Rules
				    Rule PublicIdentifier
				      Match Identifier
				      Report
				        Severity warn
				        At Identifier
				        Message "Public identifier found"
			`),
		},
		"When parsing a rules block with a string condition, a document is returned.": {
			inSource: `scope {} rules { rule PublicIdentifier { match Identifier where Identifier.text == "public" report warn at Identifier "Public identifier found" } }`,
			wantTree: normalizeMultilineLiteral(`
				Document
				  Scope
				  Rules
				    Rule PublicIdentifier
				      Match Identifier
				      Where
				        Subject Identifier
				        Property text
				        Operator ==
				        Value "public"
				      Report
				        Severity warn
				        At Identifier
				        Message "Public identifier found"
			`),
		},
		"When parsing a rules block with an integer condition, a document is returned.": {
			inSource: `scope {} rules { rule LongIdentifier { match Identifier where Identifier.length > 3 report warn at Identifier "Long identifier found" } }`,
			wantTree: normalizeMultilineLiteral(`
				Document
				  Scope
				  Rules
				    Rule LongIdentifier
				      Match Identifier
				      Where
				        Subject Identifier
				        Property length
				        Operator >
				        Value 3
				      Report
				        Severity warn
				        At Identifier
				        Message "Long identifier found"
			`),
		},
		"When parsing a rules block with a syntax-node match, a document is returned.": {
			inSource: `scope {} syntax { node File = Identifier } rules { rule FileRule { match node File where File.length > 0 report warn at File "File found" } }`,
			wantTree: normalizeMultilineLiteral(`
				Document
				  Scope
				  Syntax
				    Node File
				      Reference Identifier
				  Rules
				    Rule FileRule
				      Match node File
				      Where
				        Subject File
				        Property length
				        Operator >
				        Value 0
				      Report
				        Severity warn
				        At File
				        Message "File found"
			`),
		},
		"When the rules opening brace is missing, a diagnostic is returned.": {
			inSource:  "scope {} rules }",
			wantDiags: `Expected '{', found '}'. [15:16]`,
		},
		"When the rules closing brace is missing, a diagnostic is returned.": {
			inSource:  "scope {} rules {",
			wantDiags: `Expected '}', found 'EOF'. [16:16]`,
		},
		"When an unexpected token appears inside rules, a diagnostic is returned.": {
			inSource:  "scope {} rules { x }",
			wantDiags: `Expected 'rule', found 'identifier'. [17:18]`,
		},
		"When a rule is missing its name, a diagnostic is returned.": {
			inSource:  `scope {} rules { rule { match Identifier report warn at Identifier "x" } }`,
			wantDiags: `Expected 'identifier', found '{'. [22:23]`,
		},
		"When a rule opening brace is missing, a diagnostic is returned.": {
			inSource:  "scope {} rules { rule PublicIdentifier }",
			wantDiags: `Expected '{', found '}'. [39:40]`,
		},
		"When a rule is missing a match statement, diagnostics are returned.": {
			inSource:  "scope {} rules { rule PublicIdentifier { report warn at Identifier \"x\" } }",
			wantDiags: `Expected 'match', found 'report'. [41:47]`,
		},
		"When a rule match statement is missing its token name, a diagnostic is returned.": {
			inSource:  `scope {} rules { rule PublicIdentifier { match report warn at Identifier "x" } }`,
			wantDiags: `Expected 'identifier', found 'report'. [47:53]`,
		},
		"When a where clause is missing its subject, a diagnostic is returned.": {
			inSource:  `scope {} rules { rule PublicIdentifier { match Identifier where .text == "public" report warn at Identifier "x" } }`,
			wantDiags: `Expected 'identifier', found '.'. [64:65]`,
		},
		"When a where clause is missing its dot, a diagnostic is returned.": {
			inSource:  `scope {} rules { rule PublicIdentifier { match Identifier where Identifier text == "public" report warn at Identifier "x" } }`,
			wantDiags: `Expected '.', found 'identifier'. [75:79]`,
		},
		"When a where property is missing, a diagnostic is returned.": {
			inSource:  `scope {} rules { rule PublicIdentifier { match Identifier where Identifier. == "public" report warn at Identifier "x" } }`,
			wantDiags: `Expected 'identifier', found '=='. [76:78]`,
		},
		"When a where operator is missing, a diagnostic is returned.": {
			inSource:  `scope {} rules { rule PublicIdentifier { match Identifier where Identifier.text "public" report warn at Identifier "x" } }`,
			wantDiags: `Expected '==', found 'string'. [80:88]`,
		},
		"When a where literal is missing, a diagnostic is returned.": {
			inSource:  `scope {} rules { rule PublicIdentifier { match Identifier where Identifier.text == report warn at Identifier "x" } }`,
			wantDiags: `Expected 'string', found 'report'. [83:89]`,
		},
		"When a rule is missing its report statement, a diagnostic is returned.": {
			inSource:  `scope {} rules { rule PublicIdentifier { match Identifier } }`,
			wantDiags: `Expected 'report', found '}'. [58:59]`,
		},
		"When a report severity is missing, a diagnostic is returned.": {
			inSource:  `scope {} rules { rule PublicIdentifier { match Identifier report at Identifier "x" } }`,
			wantDiags: `Expected 'warn', found 'at'. [65:67]`,
		},
		"When a report statement is missing its at keyword, a diagnostic is returned.": {
			inSource:  `scope {} rules { rule PublicIdentifier { match Identifier report warn Identifier "x" } }`,
			wantDiags: `Expected 'at', found 'identifier'. [70:80]`,
		},
		"When a report statement is missing its target, a diagnostic is returned.": {
			inSource:  `scope {} rules { rule PublicIdentifier { match Identifier report warn at "x" } }`,
			wantDiags: `Expected 'identifier', found 'string'. [73:76]`,
		},
		"When a report statement is missing its message, a diagnostic is returned.": {
			inSource:  `scope {} rules { rule PublicIdentifier { match Identifier report warn at Identifier } }`,
			wantDiags: `Expected 'string', found '}'. [84:85]`,
		},
		"When a rule is missing its closing brace, a diagnostic is returned.": {
			inSource:  `scope {} rules { rule PublicIdentifier { match Identifier report warn at Identifier "x"`,
			wantDiags: `Expected '}', found 'EOF'. [87:87]`,
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

func Benchmark_Parse_Rules_0(b *testing.B)    { benchmark_Parse_Rules(b, 0) }
func Benchmark_Parse_Rules_1(b *testing.B)    { benchmark_Parse_Rules(b, 1) }
func Benchmark_Parse_Rules_10(b *testing.B)   { benchmark_Parse_Rules(b, 10) }
func Benchmark_Parse_Rules_100(b *testing.B)  { benchmark_Parse_Rules(b, 100) }
func Benchmark_Parse_Rules_1000(b *testing.B) { benchmark_Parse_Rules(b, 1000) }

func benchmark_Parse_Rules(b *testing.B, size int) {
	b.Helper()

	benchmark_Parse(b, rulesDSL(size))
}

func rulesDSL(size int) string {
	return "scope {}\n" +
		"tokens {\n" +
		"    Identifier = 'a'..'z'+\n" +
		"}\n" +
		"syntax {\n" +
		"    node Word = Identifier\n" +
		"    node Root = Word+\n" +
		"}\n" +
		"rules {\n" +
		strings.Repeat(
			""+
				"    rule PublicIdentifier {\n"+
				"        match Identifier\n"+
				"        where Identifier.text == \"public\"\n"+
				"        report warn at Identifier \"Public identifier found\"\n"+
				"    }\n"+
				"    rule InternalIdentifier {\n"+
				"        match Identifier\n"+
				"        where Identifier.text != \"internal\"\n"+
				"        report info at Identifier \"Internal identifier found\"\n"+
				"    }\n"+
				"    rule ShortIdentifier {\n"+
				"        match Identifier\n"+
				"        where Identifier.length < 3\n"+
				"        report info at Identifier \"Short identifier found\"\n"+
				"    }\n"+
				"    rule MediumIdentifier {\n"+
				"        match Identifier\n"+
				"        where Identifier.length <= 4\n"+
				"        report warn at Identifier \"Medium identifier found\"\n"+
				"    }\n"+
				"    rule LongIdentifier {\n"+
				"        match Identifier\n"+
				"        where Identifier.length > 5\n"+
				"        report err at Identifier \"Long identifier found\"\n"+
				"    }\n"+
				"    rule VeryLongIdentifier {\n"+
				"        match Identifier\n"+
				"        where Identifier.length >= 6\n"+
				"        report err at Identifier \"Very long identifier found\"\n"+
				"    }\n"+
				"    rule AnyIdentifier {\n"+
				"        match Identifier\n"+
				"        report info at Identifier \"Identifier found\"\n"+
				"    }\n"+
				"    rule RootNode {\n"+
				"        match node Root\n"+
				"        where Root.length > 0\n"+
				"        report warn at Root \"Root found\"\n"+
				"    }\n",
			size,
		) +
		"}"
}
