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
	source := dsl(1)
	wantDocument := ast.Document{
		Scope: ast.ScopeSection{
			Entries: []ast.ScopeEntry{
				scopeEntry(ast.ScopeEntryInclude, token.TokenString, "\"**/*.go\"", 20, 29, 12, 29),
				scopeEntry(ast.ScopeEntryExclude, token.TokenString, "\"vendor/**\"", 42, 53, 34, 53),
			},
			Span: span(0, 55),
		},
		Definitions: ast.DefinitionsSection{
			Definitions: []ast.Definition{
				alternationDefinition("letter", 74, 80, 74, 102, rangeExpression("'a'", 83, 86, token.TokenCharacter, "'z'", 88, 91), rangeExpression("'A'", 94, 97, token.TokenCharacter, "'Z'", 99, 102)),
				alternationDefinition("identifierStart", 107, 122, 107, 137, referenceExpression("letter", 125, 131), characterExpression(token.TokenCharacter, "'_'", 134, 137)),
				concatenationDefinition("value", 142, 147, 142, 169, referenceExpression("letter", 150, 156), repetitionExpression(groupExpression(alternationExpression(characterExpression(token.TokenCharacter, "'a'", 158, 161), characterExpression(token.TokenCharacter, "'b'", 164, 167)), 157, 168), token.TokenPlus, "+", 168, 169)),
			},
			Span: span(56, 171),
		},
		Tokens: ast.TokensSection{
			Tokens: []ast.TokenDefinition{
				tokenDefinition("Identifier", 185, 195, concatenationExpression(referenceExpression("identifierStart", 198, 213), repetitionExpression(referenceExpression("value", 214, 219), token.TokenStar, "*", 219, 220)), 185, 220),
				tokenDefinition("KeywordPublic", 225, 238, stringExpression("\"public\"", 241, 249), 225, 249),
				tokenDefinitionWithSkip("Whitespace", 254, 264, repetitionExpression(groupExpression(alternationExpression(characterExpression(token.TokenCharacter, "' '", 268, 271), characterExpression(token.TokenCharacter, "'\\t'", 274, 278)), 267, 279), token.TokenPlus, "+", 279, 280), 281, 285, 254, 285),
			},
			Span: span(172, 287),
		},
		Rules: ast.RulesSection{
			Rules: []ast.Rule{
				ruleWithWhere("PublicIdentifier", 305, 321, ruleMatch("Identifier", 338, 348, 332, 348), ruleCondition("Identifier", 363, 373, "text", 374, 378, token.TokenEqualEqual, "==", 379, 381, token.TokenString, "\"public\"", 382, 390, 357, 390), ruleReport(token.TokenWarn, "warn", 406, 410, "Identifier", 414, 424, "\"Public identifier found\"", 425, 450, 399, 450), 300, 456),
			},
			Span: span(288, 458),
		},
		Span: span(0, 458),
	}

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

	benchmark_Parse(b, dsl(size))
}

func dsl(size int) string {
	return "scope {\n" +
		strings.Repeat("    include \"**/*.go\"\n    exclude \"vendor/**\"\n", size) +
		"}\n" +
		"definitions {\n" +
		strings.Repeat("    letter = 'a'..'z' | 'A'..'Z'\n    identifierStart = letter | '_'\n    value = letter ('a' | 'b')+\n", size) +
		"}\n" +
		"tokens {\n" +
		strings.Repeat("    Identifier = identifierStart value*\n    KeywordPublic = \"public\"\n    Whitespace = (' ' | '\\t')+ skip\n", size) +
		"}\n" +
		"rules {\n" +
		strings.Repeat("    rule PublicIdentifier {\n        match Identifier\n        where Identifier.text == \"public\"\n        report warn at Identifier \"Public identifier found\"\n    }\n", size) +
		"}"
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
