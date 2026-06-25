// =====================================================================================================================
// == LICENSE:                 Copyright (c) 2026 Kevin De Coninck.
// == SPDX-License-Identifier: LicenseRef-PolyForm-Noncommercial-1.0.0
// =====================================================================================================================

package parser_test

import (
	"strings"
	"testing"

	"github.com/kdeconinck/spot/parser"
	"github.com/kdeconinck/spot/qa/claim"
	"github.com/kdeconinck/spot/syntax"
)

func Test_Parse_Rules(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name            string
		inSource        string
		wantDocument    syntax.Document
		wantDiagnostics []parser.Diagnostic
	}{
		{
			name:     "When parsing an empty rules block, a document is returned.",
			inSource: "scope {} rules {}",
			wantDocument: syntax.Document{
				Scope: syntax.ScopeSection{
					Span: span(0, 8),
				},
				Rules: syntax.RulesSection{
					Span: span(9, 17),
				},
				Span: span(0, 17),
			},
		},
		{
			name:     "When parsing a rules block with a rule, a document is returned.",
			inSource: "scope {} rules { rule PublicIdentifier { match Identifier report warn at Identifier \"Public identifier found\" } }",
			wantDocument: syntax.Document{
				Scope: syntax.ScopeSection{
					Span: span(0, 8),
				},
				Rules: syntax.RulesSection{
					Rules: []syntax.Rule{
						rule("PublicIdentifier", 22, 38, ruleMatch("Identifier", 47, 57, 41, 57), ruleReport(syntax.TokenWarn, "warn", 65, 69, "Identifier", 73, 83, "\"Public identifier found\"", 84, 109, 58, 109), 17, 111),
					},
					Span: span(9, 113),
				},
				Span: span(0, 113),
			},
		},
		{
			name:     "When the rules opening brace is missing, a diagnostic is returned.",
			inSource: "scope {} rules }",
			wantDocument: syntax.Document{
				Scope: syntax.ScopeSection{
					Span: span(0, 8),
				},
				Rules: syntax.RulesSection{
					Span: span(9, 14),
				},
				Span: span(0, 14),
			},
			wantDiagnostics: []parser.Diagnostic{
				diagnostic("Expected '{', found '}'.", 15, 16),
			},
		},
		{
			name:     "When the rules closing brace is missing, a diagnostic is returned.",
			inSource: "scope {} rules {",
			wantDocument: syntax.Document{
				Scope: syntax.ScopeSection{
					Span: span(0, 8),
				},
				Rules: syntax.RulesSection{
					Span: span(9, 16),
				},
				Span: span(0, 16),
			},
			wantDiagnostics: []parser.Diagnostic{
				diagnostic("Expected '}', found 'EOF'.", 16, 16),
			},
		},
		{
			name:     "When an unexpected token appears inside rules, a diagnostic is returned.",
			inSource: "scope {} rules { x }",
			wantDocument: syntax.Document{
				Scope: syntax.ScopeSection{
					Span: span(0, 8),
				},
				Rules: syntax.RulesSection{
					Span: span(9, 18),
				},
				Span: span(0, 18),
			},
			wantDiagnostics: []parser.Diagnostic{
				diagnostic("Expected 'rule', found 'identifier'.", 17, 18),
			},
		},
		{
			name:     "When a rule opening brace is missing, a diagnostic is returned.",
			inSource: "scope {} rules { rule PublicIdentifier }",
			wantDocument: syntax.Document{
				Scope: syntax.ScopeSection{
					Span: span(0, 8),
				},
				Rules: syntax.RulesSection{
					Rules: []syntax.Rule{
						rule("PublicIdentifier", 22, 38, syntax.RuleMatch{}, syntax.RuleReport{}, 17, 38),
					},
					Span: span(9, 40),
				},
				Span: span(0, 40),
			},
			wantDiagnostics: []parser.Diagnostic{
				diagnostic("Expected '{', found '}'.", 39, 40),
			},
		},
		{
			name:     "When a rule is missing a match statement, diagnostics are returned.",
			inSource: "scope {} rules { rule PublicIdentifier { report warn at Identifier \"x\" } }",
			wantDocument: syntax.Document{
				Scope: syntax.ScopeSection{
					Span: span(0, 8),
				},
				Rules: syntax.RulesSection{
					Rules: []syntax.Rule{
						rule("PublicIdentifier", 22, 38, ruleMatchWithKind(syntax.TokenReport, "report", 41, 47, 41, 47), ruleReport(syntax.TokenWarn, "warn", 48, 52, "Identifier", 56, 66, "\"x\"", 67, 70, 41, 70), 17, 72),
					},
					Span: span(9, 74),
				},
				Span: span(0, 74),
			},
			wantDiagnostics: []parser.Diagnostic{
				diagnostic("Expected 'match', found 'report'.", 41, 47),
				diagnostic("Expected 'identifier', found 'report'.", 41, 47),
			},
		},
		{
			name:     "When a report severity is missing, a diagnostic is returned.",
			inSource: "scope {} rules { rule PublicIdentifier { match Identifier report at Identifier \"x\" } }",
			wantDocument: syntax.Document{
				Scope: syntax.ScopeSection{
					Span: span(0, 8),
				},
				Rules: syntax.RulesSection{
					Rules: []syntax.Rule{
						rule("PublicIdentifier", 22, 38, ruleMatch("Identifier", 47, 57, 41, 57), ruleReport(syntax.TokenAt, "at", 65, 67, "Identifier", 68, 78, "\"x\"", 79, 82, 58, 82), 17, 84),
					},
					Span: span(9, 86),
				},
				Span: span(0, 86),
			},
			wantDiagnostics: []parser.Diagnostic{
				diagnostic("Expected 'warn', found 'at'.", 65, 67),
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
		"    Identifier = \"identifier\"\n" +
		"}\n" +
		"rules {\n" +
		strings.Repeat("    rule PublicIdentifier {\n        match Identifier\n        report warn at Identifier \"Public identifier found\"\n    }\n", size) +
		"}"
}
