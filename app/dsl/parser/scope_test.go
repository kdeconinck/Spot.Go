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

func Test_Parse_Scope(t *testing.T) {
	t.Parallel()

	for tcName, tc := range map[string]struct {
		inSource        string
		wantDocument    ast.Document
		wantDiagnostics []parser.Diagnostic
	}{
		"When parsing an empty scope block, a document is returned.": {
			inSource:     "scope {}",
			wantDocument: document(span(0, 8), scopeSection(span(0, 8)), ast.DefinitionsSection{}, ast.TokensSection{}, ast.RulesSection{}),
		},
		"When parsing a scope block with an include entry, a document is returned.": {
			inSource: "scope { include \"**/*.go\" }",
			wantDocument: document(
				span(0, 27),
				scopeSection(span(0, 27), includeScopeEntry(`"**/*.go"`, span(16, 25), span(8, 25))),
				ast.DefinitionsSection{},
				ast.TokensSection{},
				ast.RulesSection{},
			),
		},
		"When parsing a scope block with an exclude entry, a document is returned.": {
			inSource: "scope { exclude \"vendor/**\" }",
			wantDocument: document(
				span(0, 29),
				scopeSection(span(0, 29), excludeScopeEntry(`"vendor/**"`, span(16, 27), span(8, 27))),
				ast.DefinitionsSection{},
				ast.TokensSection{},
				ast.RulesSection{},
			),
		},
		"When parsing a scope block with include and exclude entries, a document is returned.": {
			inSource: "scope { include \"**/*.go\" exclude \"vendor/**\" }",
			wantDocument: document(
				span(0, 47),
				scopeSection(
					span(0, 47),
					includeScopeEntry(`"**/*.go"`, span(16, 25), span(8, 25)),
					excludeScopeEntry(`"vendor/**"`, span(34, 45), span(26, 45)),
				),
				ast.DefinitionsSection{},
				ast.TokensSection{},
				ast.RulesSection{},
			),
		},
		"When the scope keyword is missing, a diagnostic is returned.": {
			inSource:        "x",
			wantDocument:    document(span(0, 1), scopeSection(span(0, 1)), ast.DefinitionsSection{}, ast.TokensSection{}, ast.RulesSection{}),
			wantDiagnostics: []parser.Diagnostic{diagnostic("Expected 'scope', found 'identifier'.", 0, 1)},
		},
		"When the opening brace is missing, a diagnostic is returned.": {
			inSource:        "scope }",
			wantDocument:    document(span(0, 5), scopeSection(span(0, 5)), ast.DefinitionsSection{}, ast.TokensSection{}, ast.RulesSection{}),
			wantDiagnostics: []parser.Diagnostic{diagnostic("Expected '{', found '}'.", 6, 7)},
		},
		"When the closing brace is missing, a diagnostic is returned.": {
			inSource:        "scope {",
			wantDocument:    document(span(0, 7), scopeSection(span(0, 7)), ast.DefinitionsSection{}, ast.TokensSection{}, ast.RulesSection{}),
			wantDiagnostics: []parser.Diagnostic{diagnostic("Expected '}', found 'EOF'.", 7, 7)},
		},
		"When an include entry has no string, a diagnostic is returned.": {
			inSource: "scope { include }",
			wantDocument: document(
				span(0, 17),
				scopeSection(span(0, 17), invalidIncludeScopeEntry(token.TokenRightBrace, "}", span(16, 17), span(8, 17))),
				ast.DefinitionsSection{},
				ast.TokensSection{},
				ast.RulesSection{},
			),
			wantDiagnostics: []parser.Diagnostic{diagnostic("Expected 'string', found '}'.", 16, 17)},
		},
		"When an unexpected token appears inside scope, a diagnostic is returned.": {
			inSource:        "scope { x }",
			wantDocument:    document(span(0, 9), scopeSection(span(0, 9)), ast.DefinitionsSection{}, ast.TokensSection{}, ast.RulesSection{}),
			wantDiagnostics: []parser.Diagnostic{diagnostic("Expected 'include', found 'identifier'.", 8, 9)},
		},
		"When a token appears after the scope block, a diagnostic is returned.": {
			inSource:        "scope {} x",
			wantDocument:    document(span(0, 10), scopeSection(span(0, 8)), ast.DefinitionsSection{}, ast.TokensSection{}, ast.RulesSection{}),
			wantDiagnostics: []parser.Diagnostic{diagnostic("Expected 'EOF', found 'identifier'.", 9, 10)},
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
