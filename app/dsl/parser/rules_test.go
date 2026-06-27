// =====================================================================================================================
// == LICENSE:                 Copyright (c) 2026 Kevin De Coninck.
// == SPDX-License-Identifier: LicenseRef-PolyForm-Noncommercial-1.0.0
// =====================================================================================================================

package parser_test

import (
	"strings"
	"testing"

	"github.com/kdeconinck/spot/dsl/parser"
	"github.com/kdeconinck/spot/dsl/token"
	"github.com/kdeconinck/spot/qa/claim"
)

func Test_Parse_Rules(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name            string
		inSource        string
		wantDocument    token.Document
		wantDiagnostics []parser.Diagnostic
	}{
		{
			name:     "When parsing an empty rules block, a document is returned.",
			inSource: "scope {} rules {}",
			wantDocument: token.Document{
				Scope: token.ScopeSection{
					Span: span(0, 8),
				},
				Rules: token.RulesSection{
					Span: span(9, 17),
				},
				Span: span(0, 17),
			},
		},
		{
			name:     "When parsing a rules block with a rule, a document is returned.",
			inSource: "scope {} rules { rule PublicIdentifier { match Identifier report warn at Identifier \"Public identifier found\" } }",
			wantDocument: token.Document{
				Scope: token.ScopeSection{
					Span: span(0, 8),
				},
				Rules: token.RulesSection{
					Rules: []token.Rule{
						rule("PublicIdentifier", 22, 38, ruleMatch("Identifier", 47, 57, 41, 57), ruleReport(token.TokenWarn, "warn", 65, 69, "Identifier", 73, 83, "\"Public identifier found\"", 84, 109, 58, 109), 17, 111),
					},
					Span: span(9, 113),
				},
				Span: span(0, 113),
			},
		},
		{
			name:     "When parsing a rules block with a string condition, a document is returned.",
			inSource: "scope {} rules { rule PublicIdentifier { match Identifier where Identifier.text == \"public\" report warn at Identifier \"Public identifier found\" } }",
			wantDocument: token.Document{
				Scope: token.ScopeSection{
					Span: span(0, 8),
				},
				Rules: token.RulesSection{
					Rules: []token.Rule{
						ruleWithWhere("PublicIdentifier", 22, 38, ruleMatch("Identifier", 47, 57, 41, 57), ruleCondition("Identifier", 64, 74, "text", 75, 79, token.TokenEqualEqual, "==", 80, 82, token.TokenString, "\"public\"", 83, 91, 58, 91), ruleReport(token.TokenWarn, "warn", 99, 103, "Identifier", 107, 117, "\"Public identifier found\"", 118, 143, 92, 143), 17, 145),
					},
					Span: span(9, 147),
				},
				Span: span(0, 147),
			},
		},
		{
			name:     "When parsing a rules block with an integer condition, a document is returned.",
			inSource: "scope {} rules { rule LongIdentifier { match Identifier where Identifier.length > 3 report warn at Identifier \"Long identifier found\" } }",
			wantDocument: token.Document{
				Scope: token.ScopeSection{
					Span: span(0, 8),
				},
				Rules: token.RulesSection{
					Rules: []token.Rule{
						ruleWithWhere("LongIdentifier", 22, 36, ruleMatch("Identifier", 45, 55, 39, 55), ruleCondition("Identifier", 62, 72, "length", 73, 79, token.TokenGreater, ">", 80, 81, token.TokenInteger, "3", 82, 83, 56, 83), ruleReport(token.TokenWarn, "warn", 91, 95, "Identifier", 99, 109, "\"Long identifier found\"", 110, 133, 84, 133), 17, 135),
					},
					Span: span(9, 137),
				},
				Span: span(0, 137),
			},
		},
		{
			name:     "When the rules opening brace is missing, a diagnostic is returned.",
			inSource: "scope {} rules }",
			wantDocument: token.Document{
				Scope: token.ScopeSection{
					Span: span(0, 8),
				},
				Rules: token.RulesSection{
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
			wantDocument: token.Document{
				Scope: token.ScopeSection{
					Span: span(0, 8),
				},
				Rules: token.RulesSection{
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
			wantDocument: token.Document{
				Scope: token.ScopeSection{
					Span: span(0, 8),
				},
				Rules: token.RulesSection{
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
			wantDocument: token.Document{
				Scope: token.ScopeSection{
					Span: span(0, 8),
				},
				Rules: token.RulesSection{
					Rules: []token.Rule{
						rule("PublicIdentifier", 22, 38, token.RuleMatch{}, token.RuleReport{}, 17, 38),
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
			wantDocument: token.Document{
				Scope: token.ScopeSection{
					Span: span(0, 8),
				},
				Rules: token.RulesSection{
					Rules: []token.Rule{
						rule("PublicIdentifier", 22, 38, ruleMatchWithKind(token.TokenReport, "report", 41, 47, 41, 47), ruleReport(token.TokenWarn, "warn", 48, 52, "Identifier", 56, 66, "\"x\"", 67, 70, 41, 70), 17, 72),
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
			wantDocument: token.Document{
				Scope: token.ScopeSection{
					Span: span(0, 8),
				},
				Rules: token.RulesSection{
					Rules: []token.Rule{
						rule("PublicIdentifier", 22, 38, ruleMatch("Identifier", 47, 57, 41, 57), ruleReport(token.TokenAt, "at", 65, 67, "Identifier", 68, 78, "\"x\"", 79, 82, 58, 82), 17, 84),
					},
					Span: span(9, 86),
				},
				Span: span(0, 86),
			},
			wantDiagnostics: []parser.Diagnostic{
				diagnostic("Expected 'warn', found 'at'.", 65, 67),
			},
		},
		{
			name:     "When a where property is missing, a diagnostic is returned.",
			inSource: "scope {} rules { rule PublicIdentifier { match Identifier where Identifier. == \"public\" report warn at Identifier \"x\" } }",
			wantDocument: token.Document{
				Scope: token.ScopeSection{
					Span: span(0, 8),
				},
				Rules: token.RulesSection{
					Rules: []token.Rule{
						ruleWithWhere("PublicIdentifier", 22, 38, ruleMatch("Identifier", 47, 57, 41, 57), ruleConditionWithKinds("Identifier", token.TokenIdentifier, 64, 74, "==", token.TokenEqualEqual, 76, 78, "==", token.TokenEqualEqual, 76, 78, "\"public\"", token.TokenString, 79, 87, 58, 87), ruleReport(token.TokenWarn, "warn", 95, 99, "Identifier", 103, 113, "\"x\"", 114, 117, 88, 117), 17, 119),
					},
					Span: span(9, 121),
				},
				Span: span(0, 121),
			},
			wantDiagnostics: []parser.Diagnostic{
				diagnostic("Expected 'identifier', found '=='.", 76, 78),
			},
		},
		{
			name:     "When a where operator is missing, a diagnostic is returned.",
			inSource: "scope {} rules { rule PublicIdentifier { match Identifier where Identifier.text \"public\" report warn at Identifier \"x\" } }",
			wantDocument: token.Document{
				Scope: token.ScopeSection{
					Span: span(0, 8),
				},
				Rules: token.RulesSection{
					Rules: []token.Rule{
						ruleWithWhere("PublicIdentifier", 22, 38, ruleMatch("Identifier", 47, 57, 41, 57), ruleCondition("Identifier", 64, 74, "text", 75, 79, token.TokenString, "\"public\"", 80, 88, token.TokenString, "\"public\"", 80, 88, 58, 88), ruleReport(token.TokenWarn, "warn", 96, 100, "Identifier", 104, 114, "\"x\"", 115, 118, 89, 118), 17, 120),
					},
					Span: span(9, 122),
				},
				Span: span(0, 122),
			},
			wantDiagnostics: []parser.Diagnostic{
				diagnostic("Expected '==', found 'string'.", 80, 88),
			},
		},
		{
			name:     "When a where literal is missing, a diagnostic is returned.",
			inSource: "scope {} rules { rule PublicIdentifier { match Identifier where Identifier.text == report warn at Identifier \"x\" } }",
			wantDocument: token.Document{
				Scope: token.ScopeSection{
					Span: span(0, 8),
				},
				Rules: token.RulesSection{
					Rules: []token.Rule{
						ruleWithWhere("PublicIdentifier", 22, 38, ruleMatch("Identifier", 47, 57, 41, 57), ruleCondition("Identifier", 64, 74, "text", 75, 79, token.TokenEqualEqual, "==", 80, 82, token.TokenReport, "report", 83, 89, 58, 89), ruleReport(token.TokenWarn, "warn", 90, 94, "Identifier", 98, 108, "\"x\"", 109, 112, 83, 112), 17, 114),
					},
					Span: span(9, 116),
				},
				Span: span(0, 116),
			},
			wantDiagnostics: []parser.Diagnostic{
				diagnostic("Expected 'string', found 'report'.", 83, 89),
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
		strings.Repeat("    rule PublicIdentifier {\n        match Identifier\n        where Identifier.text == \"public\"\n        report warn at Identifier \"Public identifier found\"\n    }\n", size) +
		"}"
}
