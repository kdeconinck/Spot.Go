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
	"github.com/kdeconinck/spot/location"
	"github.com/kdeconinck/spot/qa/claim"
)

func Test_Parse_DSL(t *testing.T) {
	t.Parallel()

	// Arrange.
	source := testDsl()
	wantDocument := document(
		span(0, 458),
		scopeSection(
			span(0, 55),
			includeScopeEntry(`"**/*.go"`, span(20, 29), span(12, 29)),
			excludeScopeEntry(`"vendor/**"`, span(42, 53), span(34, 53)),
		),
		definitionsSection(
			span(56, 171),
			defineAlternation(
				"letter",
				span(74, 80),
				span(74, 102),
				rangeExpr("'a'", span(83, 86), "'z'", span(88, 91)),
				rangeExpr("'A'", span(94, 97), "'Z'", span(99, 102)),
			),
			defineAlternation(
				"identifierStart",
				span(107, 122),
				span(107, 137),
				refExpr("letter", span(125, 131)),
				charExpr("'_'", span(134, 137)),
			),
			defineConcatenation(
				"value",
				span(142, 147),
				span(142, 169),
				refExpr("letter", span(150, 156)),
				oneOrMore(
					groupExpr(
						alternationExpr(
							charExpr("'a'", span(158, 161)),
							charExpr("'b'", span(164, 167)),
						),
						span(157, 168),
					),
					span(168, 169),
				),
			),
		),
		tokensSection(
			span(172, 287),
			defineToken(
				"Identifier",
				span(185, 195),
				span(185, 220),
				concatenationExpr(
					refExpr("identifierStart", span(198, 213)),
					zeroOrMore(refExpr("value", span(214, 219)), span(219, 220)),
				),
			),
			defineToken(
				"KeywordPublic",
				span(225, 238),
				span(225, 249),
				stringExpr(`"public"`, span(241, 249)),
			),
			defineSkippedToken(
				"Whitespace",
				span(254, 264),
				span(281, 285),
				span(254, 285),
				oneOrMore(
					groupExpr(
						alternationExpr(
							charExpr("' '", span(268, 271)),
							charExpr(`'\t'`, span(274, 278)),
						),
						span(267, 279),
					),
					span(279, 280),
				),
			),
		),
		rulesSection(
			span(288, 458),
			defineRuleWithWhere(
				"PublicIdentifier",
				span(305, 321),
				span(300, 456),
				matchRule("Identifier", span(338, 348), span(332, 348)),
				whereCondition(
					identifierToken("Identifier", span(363, 373)),
					identifierToken("text", span(374, 378)),
					operatorToken(token.TokenEqualEqual, "==", span(379, 381)),
					literalToken(token.TokenString, `"public"`, span(382, 390)),
					span(357, 390),
				),
				reportRule(
					token.TokenWarn,
					"warn",
					span(406, 410),
					"Identifier",
					span(414, 424),
					`"Public identifier found"`,
					span(425, 450),
					span(399, 450),
				),
			),
		),
	)

	// Act.
	gotDocument, gotDiagnostics := parser.Parse(source)

	// Assert.
	claim.DeepEqual(t, "When parsing a full DSL file, a document is returned.", wantDocument, gotDocument, "Document")
	claim.Equal(t, "When parsing a full DSL file, a document is returned.", 0, len(gotDiagnostics), "Diagnostic Count")
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

	benchmark_Parse(b, makeDsl(size))
}

func makeDsl(size int) string {
	const scopeBlock = `    include "**/*.go"
    exclude "vendor/**"
`

	const definitionsBlock = `    lower = 'a'..'z'
    upper = 'A'..'Z'
    letter = lower | upper
    underscore = '_'
    identifierStart = letter | underscore
    optionalLetter = letter?
    repeatedLetter = letter*
    value = letter ('a' | 'b')+
`

	const tokensBlock = `    Identifier = identifierStart value*
    KeywordPublic = "public"
    KeywordInternal = "internal"
    Whitespace = (' ' | '\t')+ skip
`

	const rulesBlock = `    rule PublicIdentifier {
        match Identifier
        where Identifier.text == "public"
        report warn at Identifier "Public identifier found"
    }
    rule LongIdentifier {
        match Identifier
        where Identifier.length > 3
        report err at Identifier "Long identifier found"
    }
    rule AnyIdentifier {
        match Identifier
        report info at Identifier "Identifier found"
    }
`

	return "" +
		"scope {\n" +
		strings.Repeat(scopeBlock, size) +
		"}\n" +
		"definitions {\n" +
		strings.Repeat(definitionsBlock, size) +
		"}\n" +
		"tokens {\n" +
		strings.Repeat(tokensBlock, size) +
		"}\n" +
		"rules {\n" +
		strings.Repeat(rulesBlock, size) +
		"}"
}

func testDsl() string {
	const scopeBlock = `    include "**/*.go"
    exclude "vendor/**"
`

	const definitionsBlock = `    letter = 'a'..'z' | 'A'..'Z'
    identifierStart = letter | '_'
    value = letter ('a' | 'b')+
`

	const tokensBlock = `    Identifier = identifierStart value*
    KeywordPublic = "public"
    Whitespace = (' ' | '\t')+ skip
`

	const rulesBlock = `    rule PublicIdentifier {
        match Identifier
        where Identifier.text == "public"
        report warn at Identifier "Public identifier found"
    }
`

	return "" +
		"scope {\n" +
		scopeBlock +
		"}\n" +
		"definitions {\n" +
		definitionsBlock +
		"}\n" +
		"tokens {\n" +
		tokensBlock +
		"}\n" +
		"rules {\n" +
		rulesBlock +
		"}"
}

func diagnostic(message string, start, end location.Position) parser.Diagnostic {
	return parser.Diagnostic{
		Message: message,
		Span:    span(start, end),
	}
}

func document(span location.Span, scope ast.ScopeSection, definitions ast.DefinitionsSection, tokens ast.TokensSection, rules ast.RulesSection) ast.Document {
	return ast.Document{
		Scope:       scope,
		Definitions: definitions,
		Tokens:      tokens,
		Rules:       rules,
		Span:        span,
	}
}

func scopeSection(sectionSpan location.Span, entries ...ast.ScopeEntry) ast.ScopeSection {
	return ast.ScopeSection{
		Entries: entries,
		Span:    sectionSpan,
	}
}

func definitionsSection(sectionSpan location.Span, definitions ...ast.Definition) ast.DefinitionsSection {
	return ast.DefinitionsSection{
		Definitions: definitions,
		Span:        sectionSpan,
	}
}

func tokensSection(sectionSpan location.Span, tokens ...ast.TokenDefinition) ast.TokensSection {
	return ast.TokensSection{
		Tokens: tokens,
		Span:   sectionSpan,
	}
}

func rulesSection(sectionSpan location.Span, rules ...ast.Rule) ast.RulesSection {
	return ast.RulesSection{
		Rules: rules,
		Span:  sectionSpan,
	}
}

func span(start, end location.Position) location.Span {
	return location.Span{
		Start: start,
		End:   end,
	}
}

func includeScopeEntry(pattern string, patternSpan, entrySpan location.Span) ast.ScopeEntry {
	return scopeEntry(ast.ScopeEntryInclude, token.TokenString, pattern, patternSpan.Start, patternSpan.End, entrySpan.Start, entrySpan.End)
}

func excludeScopeEntry(pattern string, patternSpan, entrySpan location.Span) ast.ScopeEntry {
	return scopeEntry(ast.ScopeEntryExclude, token.TokenString, pattern, patternSpan.Start, patternSpan.End, entrySpan.Start, entrySpan.End)
}

func invalidIncludeScopeEntry(patternKind token.TokenKind, pattern string, patternSpan, entrySpan location.Span) ast.ScopeEntry {
	return scopeEntry(ast.ScopeEntryInclude, patternKind, pattern, patternSpan.Start, patternSpan.End, entrySpan.Start, entrySpan.End)
}

func scopeEntry(kind ast.ScopeEntryKind, patternKind token.TokenKind, pattern string, patternStart, patternEnd, entryStart, entryEnd location.Position) ast.ScopeEntry {
	return ast.ScopeEntry{
		Kind: kind,
		Pattern: token.Token{
			Kind: patternKind,
			Text: pattern,
			Span: span(patternStart, patternEnd),
		},
		Span: span(entryStart, entryEnd),
	}
}

func defineToken(name string, nameSpan, definitionSpan location.Span, expression ast.DefinitionExpression) ast.TokenDefinition {
	return tokenDefinition(name, nameSpan.Start, nameSpan.End, expression, definitionSpan.Start, definitionSpan.End)
}

func defineSkippedToken(name string, nameSpan, skipSpan, definitionSpan location.Span, expression ast.DefinitionExpression) ast.TokenDefinition {
	return tokenDefinitionWithSkip(name, nameSpan.Start, nameSpan.End, expression, skipSpan.Start, skipSpan.End, definitionSpan.Start, definitionSpan.End)
}

func tokenDefinition(name string, nameStart, nameEnd location.Position, expression ast.DefinitionExpression, definitionStart, definitionEnd location.Position) ast.TokenDefinition {
	return ast.TokenDefinition{
		Name: token.Token{
			Kind: token.TokenIdentifier,
			Text: name,
			Span: span(nameStart, nameEnd),
		},
		Expression: expression,
		Span:       span(definitionStart, definitionEnd),
	}
}

func tokenDefinitionWithSkip(name string, nameStart, nameEnd location.Position, expression ast.DefinitionExpression, skipStart, skipEnd, definitionStart, definitionEnd location.Position) ast.TokenDefinition {
	return ast.TokenDefinition{
		Name: token.Token{
			Kind: token.TokenIdentifier,
			Text: name,
			Span: span(nameStart, nameEnd),
		},
		Expression: expression,
		Skip: token.Token{
			Kind: token.TokenSkip,
			Text: "skip",
			Span: span(skipStart, skipEnd),
		},
		Span: span(definitionStart, definitionEnd),
	}
}

func rule(name string, nameStart, nameEnd location.Position, match ast.RuleMatch, report ast.RuleReport, ruleStart, ruleEnd location.Position) ast.Rule {
	return ruleWithWhere(name, nameStart, nameEnd, match, ast.RuleCondition{}, report, ruleStart, ruleEnd)
}

func defineRuleWithWhere(name string, nameSpan, ruleSpan location.Span, match ast.RuleMatch, where ast.RuleCondition, report ast.RuleReport) ast.Rule {
	return ruleWithWhere(name, nameSpan.Start, nameSpan.End, match, where, report, ruleSpan.Start, ruleSpan.End)
}

func ruleWithWhere(name string, nameStart, nameEnd location.Position, match ast.RuleMatch, where ast.RuleCondition, report ast.RuleReport, ruleStart, ruleEnd location.Position) ast.Rule {
	return ast.Rule{
		Name: token.Token{
			Kind: token.TokenIdentifier,
			Text: name,
			Span: span(nameStart, nameEnd),
		},
		Match:  match,
		Where:  where,
		Report: report,
		Span:   span(ruleStart, ruleEnd),
	}
}

func matchRule(tok string, tokenSpan, matchSpan location.Span) ast.RuleMatch {
	return ruleMatch(tok, tokenSpan.Start, tokenSpan.End, matchSpan.Start, matchSpan.End)
}

func identifierToken(text string, tokenSpan location.Span) token.Token {
	return token.Token{
		Kind: token.TokenIdentifier,
		Text: text,
		Span: tokenSpan,
	}
}

func operatorToken(kind token.TokenKind, text string, tokenSpan location.Span) token.Token {
	return token.Token{
		Kind: kind,
		Text: text,
		Span: tokenSpan,
	}
}

func literalToken(kind token.TokenKind, text string, tokenSpan location.Span) token.Token {
	return token.Token{
		Kind: kind,
		Text: text,
		Span: tokenSpan,
	}
}

func whereCondition(subject, property, operator, value token.Token, conditionSpan location.Span) ast.RuleCondition {
	return ast.RuleCondition{
		Subject:  subject,
		Property: property,
		Operator: operator,
		Value:    value,
		Span:     conditionSpan,
	}
}

func reportRule(severityKind token.TokenKind, severity string, severitySpan location.Span, target string, targetSpan location.Span, message string, messageSpan, reportSpan location.Span) ast.RuleReport {
	return ruleReport(severityKind, severity, severitySpan.Start, severitySpan.End, target, targetSpan.Start, targetSpan.End, message, messageSpan.Start, messageSpan.End, reportSpan.Start, reportSpan.End)
}

func ruleMatch(tok string, tokenStart, tokenEnd, matchStart, matchEnd location.Position) ast.RuleMatch {
	return ruleMatchWithKind(token.TokenIdentifier, tok, tokenStart, tokenEnd, matchStart, matchEnd)
}

func ruleMatchWithKind(kind token.TokenKind, tok string, tokenStart, tokenEnd, matchStart, matchEnd location.Position) ast.RuleMatch {
	return ast.RuleMatch{
		Token: token.Token{
			Kind: kind,
			Text: tok,
			Span: span(tokenStart, tokenEnd),
		},
		Span: span(matchStart, matchEnd),
	}
}

func ruleCondition(subject string, subjectStart, subjectEnd location.Position, property string, propertyStart, propertyEnd location.Position, operatorKind token.TokenKind, operator string, operatorStart, operatorEnd location.Position, valueKind token.TokenKind, value string, valueStart, valueEnd, conditionStart, conditionEnd location.Position) ast.RuleCondition {
	return ruleConditionWithKinds(subject, token.TokenIdentifier, subjectStart, subjectEnd, property, token.TokenIdentifier, propertyStart, propertyEnd, operator, operatorKind, operatorStart, operatorEnd, value, valueKind, valueStart, valueEnd, conditionStart, conditionEnd)
}

func ruleConditionWithKinds(subject string, subjectKind token.TokenKind, subjectStart, subjectEnd location.Position, property string, propertyKind token.TokenKind, propertyStart, propertyEnd location.Position, operator string, operatorKind token.TokenKind, operatorStart, operatorEnd location.Position, value string, valueKind token.TokenKind, valueStart, valueEnd, conditionStart, conditionEnd location.Position) ast.RuleCondition {
	return ast.RuleCondition{
		Subject: token.Token{
			Kind: subjectKind,
			Text: subject,
			Span: span(subjectStart, subjectEnd),
		},
		Property: token.Token{
			Kind: propertyKind,
			Text: property,
			Span: span(propertyStart, propertyEnd),
		},
		Operator: token.Token{
			Kind: operatorKind,
			Text: operator,
			Span: span(operatorStart, operatorEnd),
		},
		Value: token.Token{
			Kind: valueKind,
			Text: value,
			Span: span(valueStart, valueEnd),
		},
		Span: span(conditionStart, conditionEnd),
	}
}

func ruleReport(severityKind token.TokenKind, severity string, severityStart, severityEnd location.Position, target string, targetStart, targetEnd location.Position, message string, messageStart, messageEnd, reportStart, reportEnd location.Position) ast.RuleReport {
	return ast.RuleReport{
		Severity: token.Token{
			Kind: severityKind,
			Text: severity,
			Span: span(severityStart, severityEnd),
		},
		Target: token.Token{
			Kind: token.TokenIdentifier,
			Text: target,
			Span: span(targetStart, targetEnd),
		},
		Message: token.Token{
			Kind: token.TokenString,
			Text: message,
			Span: span(messageStart, messageEnd),
		},
		Span: span(reportStart, reportEnd),
	}
}

func characterDefinition(name string, nameStart, nameEnd location.Position, expressionKind token.TokenKind, expression string, expressionStart, expressionEnd, definitionStart, definitionEnd location.Position) ast.Definition {
	return ast.Definition{
		Name: token.Token{
			Kind: token.TokenIdentifier,
			Text: name,
			Span: span(nameStart, nameEnd),
		},
		Expression: characterExpression(expressionKind, expression, expressionStart, expressionEnd),
		Span:       span(definitionStart, definitionEnd),
	}
}

func rangeDefinition(name string, nameStart, nameEnd location.Position, start string, startStart, startEnd location.Position, end string, endStart, endEnd, definitionStart, definitionEnd location.Position) ast.Definition {
	return rangeDefinitionWithEndKind(name, nameStart, nameEnd, start, startStart, startEnd, token.TokenCharacter, end, endStart, endEnd, definitionStart, definitionEnd)
}

func rangeDefinitionWithEndKind(name string, nameStart, nameEnd location.Position, start string, startStart, startEnd location.Position, endKind token.TokenKind, end string, endStart, endEnd, definitionStart, definitionEnd location.Position) ast.Definition {
	return ast.Definition{
		Name: token.Token{
			Kind: token.TokenIdentifier,
			Text: name,
			Span: span(nameStart, nameEnd),
		},
		Expression: rangeExpression(start, startStart, startEnd, endKind, end, endStart, endEnd),
		Span:       span(definitionStart, definitionEnd),
	}
}

func referenceDefinition(name string, nameStart, nameEnd location.Position, reference string, referenceStart, referenceEnd, definitionStart, definitionEnd location.Position) ast.Definition {
	return ast.Definition{
		Name: token.Token{
			Kind: token.TokenIdentifier,
			Text: name,
			Span: span(nameStart, nameEnd),
		},
		Expression: referenceExpression(reference, referenceStart, referenceEnd),
		Span:       span(definitionStart, definitionEnd),
	}
}

func defineAlternation(name string, nameSpan, definitionSpan location.Span, terms ...ast.DefinitionExpression) ast.Definition {
	return alternationDefinition(name, nameSpan.Start, nameSpan.End, definitionSpan.Start, definitionSpan.End, terms...)
}

func defineConcatenation(name string, nameSpan, definitionSpan location.Span, terms ...ast.DefinitionExpression) ast.Definition {
	return concatenationDefinition(name, nameSpan.Start, nameSpan.End, definitionSpan.Start, definitionSpan.End, terms...)
}

func defineCharacter(name string, nameSpan, definitionSpan location.Span, expression ast.DefinitionExpression) ast.Definition {
	return groupDefinition(name, nameSpan.Start, nameSpan.End, expression, definitionSpan.Start, definitionSpan.End)
}

func defineReference(name string, nameSpan, definitionSpan location.Span, expression ast.DefinitionExpression) ast.Definition {
	return groupDefinition(name, nameSpan.Start, nameSpan.End, expression, definitionSpan.Start, definitionSpan.End)
}

func defineGroup(name string, nameSpan, definitionSpan location.Span, expression ast.DefinitionExpression) ast.Definition {
	return groupDefinition(name, nameSpan.Start, nameSpan.End, expression, definitionSpan.Start, definitionSpan.End)
}

func defineRepetition(name string, nameSpan, definitionSpan location.Span, expression ast.DefinitionExpression) ast.Definition {
	return repetitionDefinition(name, nameSpan.Start, nameSpan.End, expression, definitionSpan.Start, definitionSpan.End)
}

func groupDefinition(name string, nameStart, nameEnd location.Position, expression ast.DefinitionExpression, definitionStart, definitionEnd location.Position) ast.Definition {
	return ast.Definition{
		Name: token.Token{
			Kind: token.TokenIdentifier,
			Text: name,
			Span: span(nameStart, nameEnd),
		},
		Expression: expression,
		Span:       span(definitionStart, definitionEnd),
	}
}

func repetitionDefinition(name string, nameStart, nameEnd location.Position, expression ast.DefinitionExpression, definitionStart, definitionEnd location.Position) ast.Definition {
	return ast.Definition{
		Name: token.Token{
			Kind: token.TokenIdentifier,
			Text: name,
			Span: span(nameStart, nameEnd),
		},
		Expression: expression,
		Span:       span(definitionStart, definitionEnd),
	}
}

func concatenationDefinition(name string, nameStart, nameEnd, definitionStart, definitionEnd location.Position, terms ...ast.DefinitionExpression) ast.Definition {
	return ast.Definition{
		Name: token.Token{
			Kind: token.TokenIdentifier,
			Text: name,
			Span: span(nameStart, nameEnd),
		},
		Expression: concatenationExpression(terms...),
		Span:       span(definitionStart, definitionEnd),
	}
}

func alternationDefinition(name string, nameStart, nameEnd, definitionStart, definitionEnd location.Position, terms ...ast.DefinitionExpression) ast.Definition {
	return ast.Definition{
		Name: token.Token{
			Kind: token.TokenIdentifier,
			Text: name,
			Span: span(nameStart, nameEnd),
		},
		Expression: alternationExpression(terms...),
		Span:       span(definitionStart, definitionEnd),
	}
}

func concatenationExpr(terms ...ast.DefinitionExpression) ast.DefinitionExpression {
	return concatenationExpression(terms...)
}

func alternationExpr(terms ...ast.DefinitionExpression) ast.DefinitionExpression {
	return alternationExpression(terms...)
}

func concatenationExpression(terms ...ast.DefinitionExpression) ast.DefinitionExpression {
	return ast.DefinitionExpression{
		Kind:  ast.DefinitionExpressionConcatenation,
		Terms: terms,
		Span:  span(terms[0].Span.Start, terms[len(terms)-1].Span.End),
	}
}

func alternationExpression(terms ...ast.DefinitionExpression) ast.DefinitionExpression {
	return ast.DefinitionExpression{
		Kind:  ast.DefinitionExpressionAlternation,
		Terms: terms,
		Span:  span(terms[0].Span.Start, terms[len(terms)-1].Span.End),
	}
}

func charExpr(text string, tokenSpan location.Span) ast.DefinitionExpression {
	return characterExpression(token.TokenCharacter, text, tokenSpan.Start, tokenSpan.End)
}

func stringExpr(text string, tokenSpan location.Span) ast.DefinitionExpression {
	return stringExpression(text, tokenSpan.Start, tokenSpan.End)
}

func groupExpr(inner ast.DefinitionExpression, groupSpan location.Span) ast.DefinitionExpression {
	return groupExpression(inner, groupSpan.Start, groupSpan.End)
}

func zeroOrMore(inner ast.DefinitionExpression, operatorSpan location.Span) ast.DefinitionExpression {
	return repetitionExpression(inner, token.TokenStar, "*", operatorSpan.Start, operatorSpan.End)
}

func oneOrMore(inner ast.DefinitionExpression, operatorSpan location.Span) ast.DefinitionExpression {
	return repetitionExpression(inner, token.TokenPlus, "+", operatorSpan.Start, operatorSpan.End)
}

func zeroOrOne(inner ast.DefinitionExpression, operatorSpan location.Span) ast.DefinitionExpression {
	return repetitionExpression(inner, token.TokenQuestion, "?", operatorSpan.Start, operatorSpan.End)
}

func characterExpression(kind token.TokenKind, text string, start, end location.Position) ast.DefinitionExpression {
	return ast.DefinitionExpression{
		Kind: ast.DefinitionExpressionCharacter,
		Start: token.Token{
			Kind: kind,
			Text: text,
			Span: span(start, end),
		},
		Span: span(start, end),
	}
}

func refExpr(text string, tokenSpan location.Span) ast.DefinitionExpression {
	return referenceExpression(text, tokenSpan.Start, tokenSpan.End)
}

func rangeExpr(start string, startSpan location.Span, end string, endSpan location.Span) ast.DefinitionExpression {
	return rangeExpression(start, startSpan.Start, startSpan.End, token.TokenCharacter, end, endSpan.Start, endSpan.End)
}

func stringExpression(text string, start, end location.Position) ast.DefinitionExpression {
	return ast.DefinitionExpression{
		Kind: ast.DefinitionExpressionString,
		Start: token.Token{
			Kind: token.TokenString,
			Text: text,
			Span: span(start, end),
		},
		Span: span(start, end),
	}
}

func groupExpression(inner ast.DefinitionExpression, start, end location.Position) ast.DefinitionExpression {
	return ast.DefinitionExpression{
		Kind:  ast.DefinitionExpressionGroup,
		Inner: &inner,
		Span:  span(start, end),
	}
}

func repetitionExpression(inner ast.DefinitionExpression, operatorKind token.TokenKind, operator string, operatorStart, operatorEnd location.Position) ast.DefinitionExpression {
	return ast.DefinitionExpression{
		Kind: ast.DefinitionExpressionRepetition,
		Operator: token.Token{
			Kind: operatorKind,
			Text: operator,
			Span: span(operatorStart, operatorEnd),
		},
		Inner: &inner,
		Span:  span(inner.Span.Start, operatorEnd),
	}
}

func referenceExpression(text string, start, end location.Position) ast.DefinitionExpression {
	return ast.DefinitionExpression{
		Kind: ast.DefinitionExpressionReference,
		Start: token.Token{
			Kind: token.TokenIdentifier,
			Text: text,
			Span: span(start, end),
		},
		Span: span(start, end),
	}
}

func rangeExpression(start string, startStart, startEnd location.Position, endKind token.TokenKind, end string, endStart, endEnd location.Position) ast.DefinitionExpression {
	return ast.DefinitionExpression{
		Kind: ast.DefinitionExpressionRange,
		Start: token.Token{
			Kind: token.TokenCharacter,
			Text: start,
			Span: span(startStart, startEnd),
		},
		End: token.Token{
			Kind: endKind,
			Text: end,
			Span: span(endStart, endEnd),
		},
		Span: span(startStart, endEnd),
	}
}
