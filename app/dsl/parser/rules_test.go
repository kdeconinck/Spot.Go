// =====================================================================================================================
// == LICENSE:                 Copyright (c) 2026 Kevin De Coninck.
// == SPDX-License-Identifier: LicenseRef-PolyForm-Noncommercial-1.0.0
// =====================================================================================================================

package parser_test

import (
	"strings"
	"testing"

	"github.com/kdeconinck/spot/dsl/ast"
	"github.com/kdeconinck/spot/dsl/parser"
	"github.com/kdeconinck/spot/dsl/token"
	"github.com/kdeconinck/spot/qa/claim"
)

func Test_Parse_Rules(t *testing.T) {
	t.Parallel()

	for tcName, tc := range map[string]struct {
		inSource        string
		wantDocument    ast.Document
		wantDiagnostics []parser.Diagnostic
	}{
		"When parsing an empty rules block, a document is returned.": {
			inSource: "scope {} rules {}",
			wantDocument: document(
				span(0, 17),
				scopeSection(span(0, 8)),
				ast.DefinitionsSection{},
				ast.TokensSection{},
				rulesSection(span(9, 17)),
			),
		},
		"When parsing a rules block with a rule, a document is returned.": {
			inSource: "scope {} rules { rule PublicIdentifier { match Identifier report warn at Identifier \"Public identifier found\" } }",
			wantDocument: document(
				span(0, 113),
				scopeSection(span(0, 8)),
				ast.DefinitionsSection{},
				ast.TokensSection{},
				rulesSection(
					span(9, 113),
					rule(
						"PublicIdentifier",
						22, 38,
						matchRule("Identifier", span(47, 57), span(41, 57)),
						reportRule(token.TokenWarn, "warn", span(65, 69), "Identifier", span(73, 83), `"Public identifier found"`, span(84, 109), span(58, 109)),
						17, 111,
					),
				),
			),
		},
		"When parsing a rules block with a string condition, a document is returned.": {
			inSource: "scope {} rules { rule PublicIdentifier { match Identifier where Identifier.text == \"public\" report warn at Identifier \"Public identifier found\" } }",
			wantDocument: document(
				span(0, 147),
				scopeSection(span(0, 8)),
				ast.DefinitionsSection{},
				ast.TokensSection{},
				rulesSection(
					span(9, 147),
					defineRuleWithWhere(
						"PublicIdentifier",
						span(22, 38),
						span(17, 145),
						matchRule("Identifier", span(47, 57), span(41, 57)),
						whereCondition(
							identifierToken("Identifier", span(64, 74)),
							identifierToken("text", span(75, 79)),
							operatorToken(token.TokenEqualEqual, "==", span(80, 82)),
							literalToken(token.TokenString, `"public"`, span(83, 91)),
							span(58, 91),
						),
						reportRule(token.TokenWarn, "warn", span(99, 103), "Identifier", span(107, 117), `"Public identifier found"`, span(118, 143), span(92, 143)),
					),
				),
			),
		},
		"When parsing a rules block with an integer condition, a document is returned.": {
			inSource: "scope {} rules { rule LongIdentifier { match Identifier where Identifier.length > 3 report warn at Identifier \"Long identifier found\" } }",
			wantDocument: document(
				span(0, 137),
				scopeSection(span(0, 8)),
				ast.DefinitionsSection{},
				ast.TokensSection{},
				rulesSection(
					span(9, 137),
					defineRuleWithWhere(
						"LongIdentifier",
						span(22, 36),
						span(17, 135),
						matchRule("Identifier", span(45, 55), span(39, 55)),
						whereCondition(
							identifierToken("Identifier", span(62, 72)),
							identifierToken("length", span(73, 79)),
							operatorToken(token.TokenGreater, ">", span(80, 81)),
							literalToken(token.TokenInteger, "3", span(82, 83)),
							span(56, 83),
						),
						reportRule(token.TokenWarn, "warn", span(91, 95), "Identifier", span(99, 109), `"Long identifier found"`, span(110, 133), span(84, 133)),
					),
				),
			),
		},
		"When the rules opening brace is missing, a diagnostic is returned.": {
			inSource:        "scope {} rules }",
			wantDocument:    document(span(0, 14), scopeSection(span(0, 8)), ast.DefinitionsSection{}, ast.TokensSection{}, rulesSection(span(9, 14))),
			wantDiagnostics: []parser.Diagnostic{diagnostic("Expected '{', found '}'.", 15, 16)},
		},
		"When the rules closing brace is missing, a diagnostic is returned.": {
			inSource:        "scope {} rules {",
			wantDocument:    document(span(0, 16), scopeSection(span(0, 8)), ast.DefinitionsSection{}, ast.TokensSection{}, rulesSection(span(9, 16))),
			wantDiagnostics: []parser.Diagnostic{diagnostic("Expected '}', found 'EOF'.", 16, 16)},
		},
		"When an unexpected token appears inside rules, a diagnostic is returned.": {
			inSource:        "scope {} rules { x }",
			wantDocument:    document(span(0, 18), scopeSection(span(0, 8)), ast.DefinitionsSection{}, ast.TokensSection{}, rulesSection(span(9, 18))),
			wantDiagnostics: []parser.Diagnostic{diagnostic("Expected 'rule', found 'identifier'.", 17, 18)},
		},
		"When a rule opening brace is missing, a diagnostic is returned.": {
			inSource: "scope {} rules { rule PublicIdentifier }",
			wantDocument: document(
				span(0, 40),
				scopeSection(span(0, 8)),
				ast.DefinitionsSection{},
				ast.TokensSection{},
				rulesSection(
					span(9, 40),
					rule("PublicIdentifier", 22, 38, ast.RuleMatch{}, ast.RuleReport{}, 17, 38),
				),
			),
			wantDiagnostics: []parser.Diagnostic{diagnostic("Expected '{', found '}'.", 39, 40)},
		},
		"When a rule is missing a match statement, diagnostics are returned.": {
			inSource: "scope {} rules { rule PublicIdentifier { report warn at Identifier \"x\" } }",
			wantDocument: document(
				span(0, 74),
				scopeSection(span(0, 8)),
				ast.DefinitionsSection{},
				ast.TokensSection{},
				rulesSection(
					span(9, 74),
					rule(
						"PublicIdentifier",
						22, 38,
						ruleMatchWithKind(token.TokenReport, "report", 41, 47, 41, 47),
						reportRule(token.TokenWarn, "warn", span(48, 52), "Identifier", span(56, 66), `"x"`, span(67, 70), span(41, 70)),
						17, 72,
					),
				),
			),
			wantDiagnostics: []parser.Diagnostic{
				diagnostic("Expected 'match', found 'report'.", 41, 47),
				diagnostic("Expected 'identifier', found 'report'.", 41, 47),
			},
		},
		"When a report severity is missing, a diagnostic is returned.": {
			inSource: "scope {} rules { rule PublicIdentifier { match Identifier report at Identifier \"x\" } }",
			wantDocument: document(
				span(0, 86),
				scopeSection(span(0, 8)),
				ast.DefinitionsSection{},
				ast.TokensSection{},
				rulesSection(
					span(9, 86),
					rule(
						"PublicIdentifier",
						22, 38,
						matchRule("Identifier", span(47, 57), span(41, 57)),
						reportRule(token.TokenAt, "at", span(65, 67), "Identifier", span(68, 78), `"x"`, span(79, 82), span(58, 82)),
						17, 84,
					),
				),
			),
			wantDiagnostics: []parser.Diagnostic{diagnostic("Expected 'warn', found 'at'.", 65, 67)},
		},
		"When a where property is missing, a diagnostic is returned.": {
			inSource: "scope {} rules { rule PublicIdentifier { match Identifier where Identifier. == \"public\" report warn at Identifier \"x\" } }",
			wantDocument: document(
				span(0, 121),
				scopeSection(span(0, 8)),
				ast.DefinitionsSection{},
				ast.TokensSection{},
				rulesSection(
					span(9, 121),
					defineRuleWithWhere(
						"PublicIdentifier",
						span(22, 38),
						span(17, 119),
						matchRule("Identifier", span(47, 57), span(41, 57)),
						ruleConditionWithKinds("Identifier", token.TokenIdentifier, 64, 74, "==", token.TokenEqualEqual, 76, 78, "==", token.TokenEqualEqual, 76, 78, "\"public\"", token.TokenString, 79, 87, 58, 87),
						reportRule(token.TokenWarn, "warn", span(95, 99), "Identifier", span(103, 113), `"x"`, span(114, 117), span(88, 117)),
					),
				),
			),
			wantDiagnostics: []parser.Diagnostic{diagnostic("Expected 'identifier', found '=='.", 76, 78)},
		},
		"When a where operator is missing, a diagnostic is returned.": {
			inSource: "scope {} rules { rule PublicIdentifier { match Identifier where Identifier.text \"public\" report warn at Identifier \"x\" } }",
			wantDocument: document(
				span(0, 122),
				scopeSection(span(0, 8)),
				ast.DefinitionsSection{},
				ast.TokensSection{},
				rulesSection(
					span(9, 122),
					defineRuleWithWhere(
						"PublicIdentifier",
						span(22, 38),
						span(17, 120),
						matchRule("Identifier", span(47, 57), span(41, 57)),
						ruleCondition("Identifier", 64, 74, "text", 75, 79, token.TokenString, "\"public\"", 80, 88, token.TokenString, "\"public\"", 80, 88, 58, 88),
						reportRule(token.TokenWarn, "warn", span(96, 100), "Identifier", span(104, 114), `"x"`, span(115, 118), span(89, 118)),
					),
				),
			),
			wantDiagnostics: []parser.Diagnostic{diagnostic("Expected '==', found 'string'.", 80, 88)},
		},
		"When a where literal is missing, a diagnostic is returned.": {
			inSource: "scope {} rules { rule PublicIdentifier { match Identifier where Identifier.text == report warn at Identifier \"x\" } }",
			wantDocument: document(
				span(0, 116),
				scopeSection(span(0, 8)),
				ast.DefinitionsSection{},
				ast.TokensSection{},
				rulesSection(
					span(9, 116),
					defineRuleWithWhere(
						"PublicIdentifier",
						span(22, 38),
						span(17, 114),
						matchRule("Identifier", span(47, 57), span(41, 57)),
						whereCondition(
							identifierToken("Identifier", span(64, 74)),
							identifierToken("text", span(75, 79)),
							operatorToken(token.TokenEqualEqual, "==", span(80, 82)),
							literalToken(token.TokenReport, "report", span(83, 89)),
							span(58, 89),
						),
						reportRule(token.TokenWarn, "warn", span(90, 94), "Identifier", span(98, 108), `"x"`, span(109, 112), span(83, 112)),
					),
				),
			),
			wantDiagnostics: []parser.Diagnostic{diagnostic("Expected 'string', found 'report'.", 83, 89)},
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
