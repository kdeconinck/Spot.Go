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
			wantTree: snapshot(`
				Document
				  Scope
				  Rules
			`),
		},
		"When parsing a rules block with a rule, a document is returned.": {
			inSource: `scope {} rules { rule PublicIdentifier { match Identifier report warn at Identifier "Public identifier found" } }`,
			wantTree: snapshot(`
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
			wantTree: snapshot(`
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
			wantTree: snapshot(`
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
		"When a rule opening brace is missing, a diagnostic is returned.": {
			inSource:  "scope {} rules { rule PublicIdentifier }",
			wantDiags: `Expected '{', found '}'. [39:40]`,
		},
		"When a rule is missing a match statement, diagnostics are returned.": {
			inSource: "scope {} rules { rule PublicIdentifier { report warn at Identifier \"x\" } }",
			wantDiags: snapshot(`
				Expected 'match', found 'report'. [41:47]
				Expected 'identifier', found 'report'. [41:47]
			`),
		},
		"When a report severity is missing, a diagnostic is returned.": {
			inSource:  `scope {} rules { rule PublicIdentifier { match Identifier report at Identifier "x" } }`,
			wantDiags: `Expected 'warn', found 'at'. [65:67]`,
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
		"rules {\n" +
		strings.Repeat("    rule PublicIdentifier {\n        match Identifier\n        where Identifier.text == \"public\"\n        report warn at Identifier \"Public identifier found\"\n    }\n", size) +
		"}"
}
