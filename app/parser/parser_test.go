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

	"github.com/kdeconinck/spot/location"
	"github.com/kdeconinck/spot/parser"
	"github.com/kdeconinck/spot/qa/claim"
	"github.com/kdeconinck/spot/syntax"
)

func Test_Parse_DSL(t *testing.T) {
	t.Parallel()

	// Arrange.
	source := dsl(1)
	wantDocument := syntax.Document{
		Scope: syntax.ScopeSection{
			Entries: []syntax.ScopeEntry{
				scopeEntry(syntax.ScopeEntryInclude, syntax.TokenString, "\"**/*.go\"", 20, 29, 12, 29),
				scopeEntry(syntax.ScopeEntryExclude, syntax.TokenString, "\"vendor/**\"", 42, 53, 34, 53),
			},
			Span: span(0, 55),
		},
		Definitions: syntax.DefinitionsSection{
			Definitions: []syntax.Definition{
				alternationDefinition("letter", 74, 80, 74, 102, rangeExpression("'a'", 83, 86, syntax.TokenCharacter, "'z'", 88, 91), rangeExpression("'A'", 94, 97, syntax.TokenCharacter, "'Z'", 99, 102)),
				alternationDefinition("identifierStart", 107, 122, 107, 137, referenceExpression("letter", 125, 131), characterExpression(syntax.TokenCharacter, "'_'", 134, 137)),
				concatenationDefinition("value", 142, 147, 142, 169, referenceExpression("letter", 150, 156), repetitionExpression(groupExpression(alternationExpression(characterExpression(syntax.TokenCharacter, "'a'", 158, 161), characterExpression(syntax.TokenCharacter, "'b'", 164, 167)), 157, 168), syntax.TokenPlus, "+", 168, 169)),
			},
			Span: span(56, 171),
		},
		Tokens: syntax.TokensSection{
			Tokens: []syntax.TokenDefinition{
				tokenDefinition("Identifier", 185, 195, concatenationExpression(referenceExpression("identifierStart", 198, 213), repetitionExpression(referenceExpression("value", 214, 219), syntax.TokenStar, "*", 219, 220)), 185, 220),
				tokenDefinition("KeywordPublic", 225, 238, stringExpression("\"public\"", 241, 249), 225, 249),
				tokenDefinitionWithSkip("Whitespace", 254, 264, repetitionExpression(groupExpression(alternationExpression(characterExpression(syntax.TokenCharacter, "' '", 268, 271), characterExpression(syntax.TokenCharacter, "'\\t'", 274, 278)), 267, 279), syntax.TokenPlus, "+", 279, 280), 281, 285, 254, 285),
			},
			Span: span(172, 287),
		},
		Rules: syntax.RulesSection{
			Rules: []syntax.Rule{
				ruleWithWhere("PublicIdentifier", 305, 321, ruleMatch("Identifier", 338, 348, 332, 348), ruleCondition("Identifier", 363, 373, "text", 374, 378, syntax.TokenEqualEqual, "==", 379, 381, syntax.TokenString, "\"public\"", 382, 390, 357, 390), ruleReport(syntax.TokenWarn, "warn", 406, 410, "Identifier", 414, 424, "\"Public identifier found\"", 425, 450, 399, 450), 300, 456),
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

func scopeEntry(kind syntax.ScopeEntryKind, patternKind syntax.TokenKind, pattern string, patternStart, patternEnd, entryStart, entryEnd location.Position) syntax.ScopeEntry {
	return syntax.ScopeEntry{
		Kind: kind,
		Pattern: syntax.Token{
			Kind: patternKind,
			Text: pattern,
			Span: span(patternStart, patternEnd),
		},
		Span: span(entryStart, entryEnd),
	}
}

func tokenDefinition(name string, nameStart, nameEnd location.Position, expression syntax.DefinitionExpression, definitionStart, definitionEnd location.Position) syntax.TokenDefinition {
	return syntax.TokenDefinition{
		Name: syntax.Token{
			Kind: syntax.TokenIdentifier,
			Text: name,
			Span: span(nameStart, nameEnd),
		},
		Expression: expression,
		Span:       span(definitionStart, definitionEnd),
	}
}

func tokenDefinitionWithSkip(name string, nameStart, nameEnd location.Position, expression syntax.DefinitionExpression, skipStart, skipEnd, definitionStart, definitionEnd location.Position) syntax.TokenDefinition {
	return syntax.TokenDefinition{
		Name: syntax.Token{
			Kind: syntax.TokenIdentifier,
			Text: name,
			Span: span(nameStart, nameEnd),
		},
		Expression: expression,
		Skip: syntax.Token{
			Kind: syntax.TokenSkip,
			Text: "skip",
			Span: span(skipStart, skipEnd),
		},
		Span: span(definitionStart, definitionEnd),
	}
}

func rule(name string, nameStart, nameEnd location.Position, match syntax.RuleMatch, report syntax.RuleReport, ruleStart, ruleEnd location.Position) syntax.Rule {
	return ruleWithWhere(name, nameStart, nameEnd, match, syntax.RuleCondition{}, report, ruleStart, ruleEnd)
}

func ruleWithWhere(name string, nameStart, nameEnd location.Position, match syntax.RuleMatch, where syntax.RuleCondition, report syntax.RuleReport, ruleStart, ruleEnd location.Position) syntax.Rule {
	return syntax.Rule{
		Name: syntax.Token{
			Kind: syntax.TokenIdentifier,
			Text: name,
			Span: span(nameStart, nameEnd),
		},
		Match:  match,
		Where:  where,
		Report: report,
		Span:   span(ruleStart, ruleEnd),
	}
}

func ruleMatch(token string, tokenStart, tokenEnd, matchStart, matchEnd location.Position) syntax.RuleMatch {
	return ruleMatchWithKind(syntax.TokenIdentifier, token, tokenStart, tokenEnd, matchStart, matchEnd)
}

func ruleMatchWithKind(kind syntax.TokenKind, token string, tokenStart, tokenEnd, matchStart, matchEnd location.Position) syntax.RuleMatch {
	return syntax.RuleMatch{
		Token: syntax.Token{
			Kind: kind,
			Text: token,
			Span: span(tokenStart, tokenEnd),
		},
		Span: span(matchStart, matchEnd),
	}
}

func ruleCondition(subject string, subjectStart, subjectEnd location.Position, property string, propertyStart, propertyEnd location.Position, operatorKind syntax.TokenKind, operator string, operatorStart, operatorEnd location.Position, valueKind syntax.TokenKind, value string, valueStart, valueEnd, conditionStart, conditionEnd location.Position) syntax.RuleCondition {
	return ruleConditionWithKinds(subject, syntax.TokenIdentifier, subjectStart, subjectEnd, property, syntax.TokenIdentifier, propertyStart, propertyEnd, operator, operatorKind, operatorStart, operatorEnd, value, valueKind, valueStart, valueEnd, conditionStart, conditionEnd)
}

func ruleConditionWithKinds(subject string, subjectKind syntax.TokenKind, subjectStart, subjectEnd location.Position, property string, propertyKind syntax.TokenKind, propertyStart, propertyEnd location.Position, operator string, operatorKind syntax.TokenKind, operatorStart, operatorEnd location.Position, value string, valueKind syntax.TokenKind, valueStart, valueEnd, conditionStart, conditionEnd location.Position) syntax.RuleCondition {
	return syntax.RuleCondition{
		Subject: syntax.Token{
			Kind: subjectKind,
			Text: subject,
			Span: span(subjectStart, subjectEnd),
		},
		Property: syntax.Token{
			Kind: propertyKind,
			Text: property,
			Span: span(propertyStart, propertyEnd),
		},
		Operator: syntax.Token{
			Kind: operatorKind,
			Text: operator,
			Span: span(operatorStart, operatorEnd),
		},
		Value: syntax.Token{
			Kind: valueKind,
			Text: value,
			Span: span(valueStart, valueEnd),
		},
		Span: span(conditionStart, conditionEnd),
	}
}

func ruleReport(severityKind syntax.TokenKind, severity string, severityStart, severityEnd location.Position, target string, targetStart, targetEnd location.Position, message string, messageStart, messageEnd, reportStart, reportEnd location.Position) syntax.RuleReport {
	return syntax.RuleReport{
		Severity: syntax.Token{
			Kind: severityKind,
			Text: severity,
			Span: span(severityStart, severityEnd),
		},
		Target: syntax.Token{
			Kind: syntax.TokenIdentifier,
			Text: target,
			Span: span(targetStart, targetEnd),
		},
		Message: syntax.Token{
			Kind: syntax.TokenString,
			Text: message,
			Span: span(messageStart, messageEnd),
		},
		Span: span(reportStart, reportEnd),
	}
}

func characterDefinition(name string, nameStart, nameEnd location.Position, expressionKind syntax.TokenKind, expression string, expressionStart, expressionEnd, definitionStart, definitionEnd location.Position) syntax.Definition {
	return syntax.Definition{
		Name: syntax.Token{
			Kind: syntax.TokenIdentifier,
			Text: name,
			Span: span(nameStart, nameEnd),
		},
		Expression: characterExpression(expressionKind, expression, expressionStart, expressionEnd),
		Span:       span(definitionStart, definitionEnd),
	}
}

func rangeDefinition(name string, nameStart, nameEnd location.Position, start string, startStart, startEnd location.Position, end string, endStart, endEnd, definitionStart, definitionEnd location.Position) syntax.Definition {
	return rangeDefinitionWithEndKind(name, nameStart, nameEnd, start, startStart, startEnd, syntax.TokenCharacter, end, endStart, endEnd, definitionStart, definitionEnd)
}

func rangeDefinitionWithEndKind(name string, nameStart, nameEnd location.Position, start string, startStart, startEnd location.Position, endKind syntax.TokenKind, end string, endStart, endEnd, definitionStart, definitionEnd location.Position) syntax.Definition {
	return syntax.Definition{
		Name: syntax.Token{
			Kind: syntax.TokenIdentifier,
			Text: name,
			Span: span(nameStart, nameEnd),
		},
		Expression: rangeExpression(start, startStart, startEnd, endKind, end, endStart, endEnd),
		Span:       span(definitionStart, definitionEnd),
	}
}

func referenceDefinition(name string, nameStart, nameEnd location.Position, reference string, referenceStart, referenceEnd, definitionStart, definitionEnd location.Position) syntax.Definition {
	return syntax.Definition{
		Name: syntax.Token{
			Kind: syntax.TokenIdentifier,
			Text: name,
			Span: span(nameStart, nameEnd),
		},
		Expression: referenceExpression(reference, referenceStart, referenceEnd),
		Span:       span(definitionStart, definitionEnd),
	}
}

func groupDefinition(name string, nameStart, nameEnd location.Position, expression syntax.DefinitionExpression, definitionStart, definitionEnd location.Position) syntax.Definition {
	return syntax.Definition{
		Name: syntax.Token{
			Kind: syntax.TokenIdentifier,
			Text: name,
			Span: span(nameStart, nameEnd),
		},
		Expression: expression,
		Span:       span(definitionStart, definitionEnd),
	}
}

func repetitionDefinition(name string, nameStart, nameEnd location.Position, expression syntax.DefinitionExpression, definitionStart, definitionEnd location.Position) syntax.Definition {
	return syntax.Definition{
		Name: syntax.Token{
			Kind: syntax.TokenIdentifier,
			Text: name,
			Span: span(nameStart, nameEnd),
		},
		Expression: expression,
		Span:       span(definitionStart, definitionEnd),
	}
}

func concatenationDefinition(name string, nameStart, nameEnd, definitionStart, definitionEnd location.Position, terms ...syntax.DefinitionExpression) syntax.Definition {
	return syntax.Definition{
		Name: syntax.Token{
			Kind: syntax.TokenIdentifier,
			Text: name,
			Span: span(nameStart, nameEnd),
		},
		Expression: concatenationExpression(terms...),
		Span:       span(definitionStart, definitionEnd),
	}
}

func alternationDefinition(name string, nameStart, nameEnd, definitionStart, definitionEnd location.Position, terms ...syntax.DefinitionExpression) syntax.Definition {
	return syntax.Definition{
		Name: syntax.Token{
			Kind: syntax.TokenIdentifier,
			Text: name,
			Span: span(nameStart, nameEnd),
		},
		Expression: alternationExpression(terms...),
		Span:       span(definitionStart, definitionEnd),
	}
}

func concatenationExpression(terms ...syntax.DefinitionExpression) syntax.DefinitionExpression {
	return syntax.DefinitionExpression{
		Kind:  syntax.DefinitionExpressionConcatenation,
		Terms: terms,
		Span:  span(terms[0].Span.Start, terms[len(terms)-1].Span.End),
	}
}

func alternationExpression(terms ...syntax.DefinitionExpression) syntax.DefinitionExpression {
	return syntax.DefinitionExpression{
		Kind:  syntax.DefinitionExpressionAlternation,
		Terms: terms,
		Span:  span(terms[0].Span.Start, terms[len(terms)-1].Span.End),
	}
}

func characterExpression(kind syntax.TokenKind, text string, start, end location.Position) syntax.DefinitionExpression {
	return syntax.DefinitionExpression{
		Kind: syntax.DefinitionExpressionCharacter,
		Start: syntax.Token{
			Kind: kind,
			Text: text,
			Span: span(start, end),
		},
		Span: span(start, end),
	}
}

func stringExpression(text string, start, end location.Position) syntax.DefinitionExpression {
	return syntax.DefinitionExpression{
		Kind: syntax.DefinitionExpressionString,
		Start: syntax.Token{
			Kind: syntax.TokenString,
			Text: text,
			Span: span(start, end),
		},
		Span: span(start, end),
	}
}

func groupExpression(inner syntax.DefinitionExpression, start, end location.Position) syntax.DefinitionExpression {
	return syntax.DefinitionExpression{
		Kind:  syntax.DefinitionExpressionGroup,
		Inner: &inner,
		Span:  span(start, end),
	}
}

func repetitionExpression(inner syntax.DefinitionExpression, operatorKind syntax.TokenKind, operator string, operatorStart, operatorEnd location.Position) syntax.DefinitionExpression {
	return syntax.DefinitionExpression{
		Kind: syntax.DefinitionExpressionRepetition,
		Operator: syntax.Token{
			Kind: operatorKind,
			Text: operator,
			Span: span(operatorStart, operatorEnd),
		},
		Inner: &inner,
		Span:  span(inner.Span.Start, operatorEnd),
	}
}

func referenceExpression(text string, start, end location.Position) syntax.DefinitionExpression {
	return syntax.DefinitionExpression{
		Kind: syntax.DefinitionExpressionReference,
		Start: syntax.Token{
			Kind: syntax.TokenIdentifier,
			Text: text,
			Span: span(start, end),
		},
		Span: span(start, end),
	}
}

func rangeExpression(start string, startStart, startEnd location.Position, endKind syntax.TokenKind, end string, endStart, endEnd location.Position) syntax.DefinitionExpression {
	return syntax.DefinitionExpression{
		Kind: syntax.DefinitionExpressionRange,
		Start: syntax.Token{
			Kind: syntax.TokenCharacter,
			Text: start,
			Span: span(startStart, startEnd),
		},
		End: syntax.Token{
			Kind: endKind,
			Text: end,
			Span: span(endStart, endEnd),
		},
		Span: span(startStart, endEnd),
	}
}
