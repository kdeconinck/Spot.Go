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
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Act.
			gotDocument, gotDiagnostics := parser.Parse(tc.inSource)

			// Assert.
			claim.Equal(t, tc.name, tc.wantDocument, gotDocument, "Document")
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

func dsl(_ int) string {
	return "scope {}"
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
