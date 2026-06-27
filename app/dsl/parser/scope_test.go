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

	for _, tc := range []struct {
		name            string
		inSource        string
		wantDocument    ast.Document
		wantDiagnostics []parser.Diagnostic
	}{
		{
			name:     "When parsing an empty scope block, a document is returned.",
			inSource: "scope {}",
			wantDocument: ast.Document{
				Scope: ast.ScopeSection{
					Span: span(0, 8),
				},
				Span: span(0, 8),
			},
		},
		{
			name:     "When parsing a scope block with an include entry, a document is returned.",
			inSource: "scope { include \"**/*.go\" }",
			wantDocument: ast.Document{
				Scope: ast.ScopeSection{
					Entries: []ast.ScopeEntry{
						scopeEntry(ast.ScopeEntryInclude, token.TokenString, "\"**/*.go\"", 16, 25, 8, 25),
					},
					Span: span(0, 27),
				},
				Span: span(0, 27),
			},
		},
		{
			name:     "When parsing a scope block with an exclude entry, a document is returned.",
			inSource: "scope { exclude \"vendor/**\" }",
			wantDocument: ast.Document{
				Scope: ast.ScopeSection{
					Entries: []ast.ScopeEntry{
						scopeEntry(ast.ScopeEntryExclude, token.TokenString, "\"vendor/**\"", 16, 27, 8, 27),
					},
					Span: span(0, 29),
				},
				Span: span(0, 29),
			},
		},
		{
			name:     "When parsing a scope block with include and exclude entries, a document is returned.",
			inSource: "scope { include \"**/*.go\" exclude \"vendor/**\" }",
			wantDocument: ast.Document{
				Scope: ast.ScopeSection{
					Entries: []ast.ScopeEntry{
						scopeEntry(ast.ScopeEntryInclude, token.TokenString, "\"**/*.go\"", 16, 25, 8, 25),
						scopeEntry(ast.ScopeEntryExclude, token.TokenString, "\"vendor/**\"", 34, 45, 26, 45),
					},
					Span: span(0, 47),
				},
				Span: span(0, 47),
			},
		},
		{
			name:     "When the scope keyword is missing, a diagnostic is returned.",
			inSource: "x",
			wantDocument: ast.Document{
				Scope: ast.ScopeSection{
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
			wantDocument: ast.Document{
				Scope: ast.ScopeSection{
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
			wantDocument: ast.Document{
				Scope: ast.ScopeSection{
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
			wantDocument: ast.Document{
				Scope: ast.ScopeSection{
					Entries: []ast.ScopeEntry{
						scopeEntry(ast.ScopeEntryInclude, token.TokenRightBrace, "}", 16, 17, 8, 17),
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
			wantDocument: ast.Document{
				Scope: ast.ScopeSection{
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
			wantDocument: ast.Document{
				Scope: ast.ScopeSection{
					Span: span(0, 8),
				},
				Span: span(0, 10),
			},
			wantDiagnostics: []parser.Diagnostic{
				diagnostic("Expected 'EOF', found 'identifier'.", 9, 10),
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
