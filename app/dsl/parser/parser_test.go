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
	"strconv"
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
	wantTree := snapshot(`
		Document (0..458)
		  Scope (0..55)
		    Include "**/*.go" (12..29)
		    Exclude "vendor/**" (34..53)
		  Definitions (56..171)
		    Definition letter (74..102)
		      Alternation (83..102)
		        Range 'a' 'z' (83..91)
		        Range 'A' 'Z' (94..102)
		    Definition identifierStart (107..137)
		      Alternation (125..137)
		        Reference letter (125..131)
		        Character '_' (134..137)
		    Definition value (142..169)
		      Concatenation (150..169)
		        Reference letter (150..156)
		        Repetition + (157..169)
		          Group (157..168)
		            Alternation (158..167)
		              Character 'a' (158..161)
		              Character 'b' (164..167)
		  Tokens (172..287)
		    Token Identifier (185..220)
		      Concatenation (198..220)
		        Reference identifierStart (198..213)
		        Repetition * (214..220)
		          Reference value (214..219)
		    Token KeywordPublic (225..249)
		      String "public" (241..249)
		    Token Whitespace (254..285)
		      Repetition + (267..280)
		        Group (267..279)
		          Alternation (268..278)
		            Character ' ' (268..271)
		            Character '\t' (274..278)
		      Skip (281..285)
		  Rules (288..458)
		    Rule PublicIdentifier (300..456)
		      Match Identifier (332..348)
		      Where (357..390)
		        Subject Identifier (363..373)
		        Property text (374..378)
		        Operator == (379..381)
		        Value "public" (382..390)
		      Report (399..450)
		        Severity warn (406..410)
		        At Identifier (414..424)
		        Message "Public identifier found" (425..450)
	`)

	// Act.
	gotDocument, gotDiagnostics := parser.Parse(source)

	// Assert.
	claim.Equal(t, "When parsing a full DSL file, no diagnostics are returned.", "", debugDiagnostics(gotDiagnostics), "Diagnostics")
	claim.Equal(t, "When parsing a full DSL file, the document structure and spans are returned.", wantTree, debugDocument(gotDocument, true), "Document")
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

func snapshot(text string) string {
	lines := strings.Split(strings.TrimSpace(text), "\n")

	for idx := range lines {
		lines[idx] = strings.TrimRight(strings.TrimLeft(lines[idx], "\t"), " ")
	}

	return strings.Join(lines, "\n")
}

func debugDocument(document ast.Document, includeSpans bool) string {
	var builder strings.Builder

	writeDocument(&builder, document, 0, includeSpans)

	return strings.TrimSpace(builder.String())
}

func debugDiagnostics(diagnostics []parser.Diagnostic) string {
	var builder strings.Builder

	for idx := range diagnostics {
		builder.WriteString(diagnostics[idx].Message)
		builder.WriteString(" [")
		builder.WriteString(strconv.Itoa(int(diagnostics[idx].Span.Start)))
		builder.WriteString(":")
		builder.WriteString(strconv.Itoa(int(diagnostics[idx].Span.End)))
		builder.WriteString("]\n")
	}

	return strings.TrimSpace(builder.String())
}

func writeDocument(builder *strings.Builder, document ast.Document, depth int, includeSpans bool) {
	writeLine(builder, depth, labelWithSpan("Document", document.Span, includeSpans))

	if document.Scope.Span != (location.Span{}) {
		writeLine(builder, depth+1, labelWithSpan("Scope", document.Scope.Span, includeSpans))

		for idx := range document.Scope.Entries {
			entry := document.Scope.Entries[idx]
			label := "Include "
			if entry.Kind == ast.ScopeEntryExclude {
				label = "Exclude "
			}

			writeLine(builder, depth+2, labelWithSpan(label+entry.Pattern.Text, entry.Span, includeSpans))
		}
	}

	if document.Definitions.Span != (location.Span{}) {
		writeLine(builder, depth+1, labelWithSpan("Definitions", document.Definitions.Span, includeSpans))

		for idx := range document.Definitions.Definitions {
			definition := document.Definitions.Definitions[idx]
			writeLine(builder, depth+2, labelWithSpan("Definition "+definition.Name.Text, definition.Span, includeSpans))
			writeExpression(builder, definition.Expression, depth+3, includeSpans)
		}
	}

	if document.Tokens.Span != (location.Span{}) {
		writeLine(builder, depth+1, labelWithSpan("Tokens", document.Tokens.Span, includeSpans))

		for idx := range document.Tokens.Tokens {
			definition := document.Tokens.Tokens[idx]
			writeLine(builder, depth+2, labelWithSpan("Token "+definition.Name.Text, definition.Span, includeSpans))
			writeExpression(builder, definition.Expression, depth+3, includeSpans)

			if definition.Skip.Kind == token.TokenSkip {
				writeLine(builder, depth+3, labelWithSpan("Skip", definition.Skip.Span, includeSpans))
			}
		}
	}

	if document.Rules.Span != (location.Span{}) {
		writeLine(builder, depth+1, labelWithSpan("Rules", document.Rules.Span, includeSpans))

		for idx := range document.Rules.Rules {
			rule := document.Rules.Rules[idx]
			writeLine(builder, depth+2, labelWithSpan("Rule "+rule.Name.Text, rule.Span, includeSpans))
			writeLine(builder, depth+3, labelWithSpan("Match "+rule.Match.Token.Text, rule.Match.Span, includeSpans))

			if rule.Where.Span != (location.Span{}) {
				writeLine(builder, depth+3, labelWithSpan("Where", rule.Where.Span, includeSpans))
				writeLine(builder, depth+4, labelWithSpan("Subject "+rule.Where.Subject.Text, rule.Where.Subject.Span, includeSpans))
				writeLine(builder, depth+4, labelWithSpan("Property "+rule.Where.Property.Text, rule.Where.Property.Span, includeSpans))
				writeLine(builder, depth+4, labelWithSpan("Operator "+rule.Where.Operator.Text, rule.Where.Operator.Span, includeSpans))
				writeLine(builder, depth+4, labelWithSpan("Value "+rule.Where.Value.Text, rule.Where.Value.Span, includeSpans))
			}

			writeLine(builder, depth+3, labelWithSpan("Report", rule.Report.Span, includeSpans))
			writeLine(builder, depth+4, labelWithSpan("Severity "+rule.Report.Severity.Text, rule.Report.Severity.Span, includeSpans))
			writeLine(builder, depth+4, labelWithSpan("At "+rule.Report.Target.Text, rule.Report.Target.Span, includeSpans))
			writeLine(builder, depth+4, labelWithSpan("Message "+rule.Report.Message.Text, rule.Report.Message.Span, includeSpans))
		}
	}
}

func writeExpression(builder *strings.Builder, expression ast.DefinitionExpression, depth int, includeSpans bool) {
	switch expression.Kind {
	case ast.DefinitionExpressionReference:
		writeLine(builder, depth, labelWithSpan("Reference "+expression.Start.Text, expression.Span, includeSpans))

	case ast.DefinitionExpressionCharacter:
		writeLine(builder, depth, labelWithSpan("Character "+expression.Start.Text, expression.Span, includeSpans))

	case ast.DefinitionExpressionString:
		writeLine(builder, depth, labelWithSpan("String "+expression.Start.Text, expression.Span, includeSpans))

	case ast.DefinitionExpressionRange:
		writeLine(builder, depth, labelWithSpan("Range "+expression.Start.Text+" "+expression.End.Text, expression.Span, includeSpans))

	case ast.DefinitionExpressionAlternation:
		writeLine(builder, depth, labelWithSpan("Alternation", expression.Span, includeSpans))

		for idx := range expression.Terms {
			writeExpression(builder, expression.Terms[idx], depth+1, includeSpans)
		}

	case ast.DefinitionExpressionConcatenation:
		writeLine(builder, depth, labelWithSpan("Concatenation", expression.Span, includeSpans))

		for idx := range expression.Terms {
			writeExpression(builder, expression.Terms[idx], depth+1, includeSpans)
		}

	case ast.DefinitionExpressionRepetition:
		writeLine(builder, depth, labelWithSpan("Repetition "+expression.Operator.Text, expression.Span, includeSpans))
		writeExpression(builder, *expression.Inner, depth+1, includeSpans)

	case ast.DefinitionExpressionGroup:
		writeLine(builder, depth, labelWithSpan("Group", expression.Span, includeSpans))
		writeExpression(builder, *expression.Inner, depth+1, includeSpans)
	}
}

func labelWithSpan(text string, span location.Span, includeSpan bool) string {
	if !includeSpan {
		return text
	}

	return text + " (" + strconv.Itoa(int(span.Start)) + ".." + strconv.Itoa(int(span.End)) + ")"
}

func writeLine(builder *strings.Builder, depth int, text string) {
	builder.WriteString(strings.Repeat("  ", depth))
	builder.WriteString(text)
	builder.WriteByte('\n')
}
