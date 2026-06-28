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

func Test_Parse_Scope(t *testing.T) {
	t.Parallel()

	for tcName, tc := range map[string]struct {
		inSource  string
		wantTree  string
		wantDiags string
	}{
		"When parsing an empty scope block, a document is returned.": {
			inSource: "scope {}",
			wantTree: snapshot(`
				Document
				  Scope
			`),
		},
		"When parsing a scope block with an include entry, a document is returned.": {
			inSource: `scope { include "**/*.go" }`,
			wantTree: snapshot(`
				Document
				  Scope
				    Include "**/*.go"
			`),
		},
		"When parsing a scope block with an exclude entry, a document is returned.": {
			inSource: `scope { exclude "vendor/**" }`,
			wantTree: snapshot(`
				Document
				  Scope
				    Exclude "vendor/**"
			`),
		},
		"When parsing a scope block with include and exclude entries, a document is returned.": {
			inSource: `scope { include "**/*.go" exclude "vendor/**" }`,
			wantTree: snapshot(`
				Document
				  Scope
				    Include "**/*.go"
				    Exclude "vendor/**"
			`),
		},
		"When the scope keyword is missing, a diagnostic is returned.": {
			inSource:  "x",
			wantDiags: `Expected 'scope', found 'identifier'. [0:1]`,
		},
		"When the opening brace is missing, a diagnostic is returned.": {
			inSource:  "scope }",
			wantDiags: `Expected '{', found '}'. [6:7]`,
		},
		"When the closing brace is missing, a diagnostic is returned.": {
			inSource:  "scope {",
			wantDiags: `Expected '}', found 'EOF'. [7:7]`,
		},
		"When an include entry has no string, a diagnostic is returned.": {
			inSource:  "scope { include }",
			wantDiags: `Expected 'string', found '}'. [16:17]`,
		},
		"When an unexpected token appears inside scope, a diagnostic is returned.": {
			inSource:  "scope { x }",
			wantDiags: `Expected 'include', found 'identifier'. [8:9]`,
		},
		"When a token appears after the scope block, a diagnostic is returned.": {
			inSource:  "scope {} x",
			wantDiags: `Expected 'EOF', found 'identifier'. [9:10]`,
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

func Benchmark_Parse_Scope_0(b *testing.B)    { benchmark_Parse_Scope(b, 0) }
func Benchmark_Parse_Scope_1(b *testing.B)    { benchmark_Parse_Scope(b, 1) }
func Benchmark_Parse_Scope_10(b *testing.B)   { benchmark_Parse_Scope(b, 10) }
func Benchmark_Parse_Scope_100(b *testing.B)  { benchmark_Parse_Scope(b, 100) }
func Benchmark_Parse_Scope_1000(b *testing.B) { benchmark_Parse_Scope(b, 1000) }

func benchmark_Parse_Scope(b *testing.B, size int) {
	b.Helper()

	benchmark_Parse(b, scopeDSL(size))
}

func scopeDSL(size int) string {
	return "scope {\n" +
		strings.Repeat("    include \"**/*.go\"\n    exclude \"vendor/**\"\n", size) +
		"}"
}
