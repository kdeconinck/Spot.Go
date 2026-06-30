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

func Test_Parse_ReturnsDocument(t *testing.T) {
	t.Parallel()

	// Arrange.
	source := fullHappyPathDSL(1)
	wantTree := normalizeMultilineLiteral(`
		Document
		  Scope
		    Include "**/*.go"
		    Exclude "vendor/**"
		  Definitions
		    Definition lower
		      Range 'a' 'z'
		    Definition upper
		      Range 'A' 'Z'
		    Definition digit
		      Range '0' '9'
		    Definition underscore
		      Character '_'
		    Definition letter
		      Alternation
		        Reference lower
		        Reference upper
		    Definition identifierStart
		      Alternation
		        Reference letter
		        Reference underscore
		    Definition identifierPart
		      Alternation
		        Reference letter
		        Reference digit
		        Reference underscore
		    Definition optionalSign
		      Repetition ?
		        Group
		          Alternation
		            Character '+'
		            Character '-'
		    Definition repeatedLetter
		      Repetition *
		        Reference letter
		    Definition value
		      Concatenation
		        Reference letter
		        Repetition +
		          Group
		            Alternation
		              Character 'a'
		              Character 'b'
		  Tokens
		    Token Identifier
		      Concatenation
		        Reference identifierStart
		        Repetition *
		          Reference identifierPart
		    Token KeywordPublic
		      String "public"
		    Token KeywordInternal
		      String "internal"
		    Token SignedInteger
		      Concatenation
		        Reference optionalSign
		        Repetition +
		          Reference digit
		    Token Whitespace
		      Repetition +
		        Group
		          Alternation
		            Character ' '
		            Character '\t'
		            Character '\n'
		            Character '\r'
		      Skip
		  Rules
		    Rule PublicIdentifier
		      Match Identifier
		      Where
		        Subject Identifier
		        Property text
		        Operator ==
		        Value "public"
		      Report
		        Severity warn
		        At Identifier
		        Message "Public identifier found"
		    Rule InternalIdentifier
		      Match Identifier
		      Where
		        Subject Identifier
		        Property text
		        Operator !=
		        Value "internal"
		      Report
		        Severity info
		        At Identifier
		        Message "Internal identifier found"
		    Rule ShortIdentifier
		      Match Identifier
		      Where
		        Subject Identifier
		        Property length
		        Operator <
		        Value 3
		      Report
		        Severity info
		        At Identifier
		        Message "Short identifier found"
		    Rule MediumIdentifier
		      Match Identifier
		      Where
		        Subject Identifier
		        Property length
		        Operator <=
		        Value 4
		      Report
		        Severity warn
		        At Identifier
		        Message "Medium identifier found"
		    Rule LongIdentifier
		      Match Identifier
		      Where
		        Subject Identifier
		        Property length
		        Operator >
		        Value 5
		      Report
		        Severity err
		        At Identifier
		        Message "Long identifier found"
		    Rule VeryLongIdentifier
		      Match Identifier
		      Where
		        Subject Identifier
		        Property length
		        Operator >=
		        Value 6
		      Report
		        Severity err
		        At Identifier
		        Message "Very long identifier found"
		    Rule AnyIdentifier
		      Match Identifier
		      Report
		        Severity info
		        At Identifier
		        Message "Identifier found"
	`)

	// Act.
	gotDocument, gotErr := parser.Parse(source)

	// Assert.
	claim.Equal(t, "When parsing realistic DSL source, no parse error is returned.", "", formatParseError(gotErr), "Parse Error")
	claim.Equal(t, "When parsing realistic DSL source, the full happy-path document structure is returned.", wantTree, renderDocument(source, gotDocument, false), "Document")
}

func Test_Parse_PreservesSpans(t *testing.T) {
	t.Parallel()

	// Arrange.
	source := fullHappyPathDSL(1)
	want := normalizeMultilineLiteral(`
		Document [0:1644]
		  Scope [0:55]
		    Include "**/*.go" [12:29]
		    Exclude "vendor/**" [34:53]
		  Definitions [56:366]
		    Definition lower [74:90]
		      Range 'a' 'z' [82:90]
		    Definition upper [95:111]
		      Range 'A' 'Z' [103:111]
		    Definition digit [116:132]
		      Range '0' '9' [124:132]
		    Definition underscore [137:153]
		      Character '_' [150:153]
		    Definition letter [158:180]
		      Alternation [167:180]
		        Reference lower [167:172]
		        Reference upper [175:180]
		    Definition identifierStart [185:222]
		      Alternation [203:222]
		        Reference letter [203:209]
		        Reference underscore [212:222]
		    Definition identifierPart [227:271]
		      Alternation [244:271]
		        Reference letter [244:250]
		        Reference digit [253:258]
		        Reference underscore [261:271]
		    Definition optionalSign [276:303]
		      Repetition ? [291:303]
		        Group [291:302]
		          Alternation [292:301]
		            Character '+' [292:295]
		            Character '-' [298:301]
		    Definition repeatedLetter [308:332]
		      Repetition * [325:332]
		        Reference letter [325:331]
		    Definition value [337:364]
		      Concatenation [345:364]
		        Reference letter [345:351]
		        Repetition + [352:364]
		          Group [352:363]
		            Alternation [353:362]
		              Character 'a' [353:356]
		              Character 'b' [359:362]
		  Tokens [367:578]
		    Token Identifier [380:424]
		      Concatenation [393:424]
		        Reference identifierStart [393:408]
		        Repetition * [409:424]
		          Reference identifierPart [409:423]
		    Token KeywordPublic [429:453]
		      String "public" [445:453]
		    Token KeywordInternal [458:486]
		      String "internal" [476:486]
		    Token SignedInteger [491:526]
		      Concatenation [507:526]
		        Reference optionalSign [507:519]
		        Repetition + [520:526]
		          Reference digit [520:525]
		    Token Whitespace [531:576]
		      Repetition + [544:571]
		        Group [544:570]
		          Alternation [545:569]
		            Character ' ' [545:548]
		            Character '\t' [551:555]
		            Character '\n' [558:562]
		            Character '\r' [565:569]
		      Skip [572:576]
		  Rules [579:1644]
		    Rule PublicIdentifier [591:747]
		      Match Identifier [623:639]
		      Where [648:681]
		        Subject Identifier [654:664]
		        Property text [665:669]
		        Operator == [670:672]
		        Value "public" [673:681]
		      Report [690:741]
		        Severity warn [697:701]
		        At Identifier [705:715]
		        Message "Public identifier found" [716:741]
		    Rule InternalIdentifier [752:914]
		      Match Identifier [786:802]
		      Where [811:846]
		        Subject Identifier [817:827]
		        Property text [828:832]
		        Operator != [833:835]
		        Value "internal" [836:846]
		      Report [855:908]
		        Severity info [862:866]
		        At Identifier [870:880]
		        Message "Internal identifier found" [881:908]
		    Rule ShortIdentifier [919:1067]
		      Match Identifier [950:966]
		      Where [975:1002]
		        Subject Identifier [981:991]
		        Property length [992:998]
		        Operator < [999:1000]
		        Value 3 [1001:1002]
		      Report [1011:1061]
		        Severity info [1018:1022]
		        At Identifier [1026:1036]
		        Message "Short identifier found" [1037:1061]
		    Rule MediumIdentifier [1072:1223]
		      Match Identifier [1104:1120]
		      Where [1129:1157]
		        Subject Identifier [1135:1145]
		        Property length [1146:1152]
		        Operator <= [1153:1155]
		        Value 4 [1156:1157]
		      Report [1166:1217]
		        Severity warn [1173:1177]
		        At Identifier [1181:1191]
		        Message "Medium identifier found" [1192:1217]
		    Rule LongIdentifier [1228:1373]
		      Match Identifier [1258:1274]
		      Where [1283:1310]
		        Subject Identifier [1289:1299]
		        Property length [1300:1306]
		        Operator > [1307:1308]
		        Value 5 [1309:1310]
		      Report [1319:1367]
		        Severity err [1326:1329]
		        At Identifier [1333:1343]
		        Message "Long identifier found" [1344:1367]
		    Rule VeryLongIdentifier [1378:1533]
		      Match Identifier [1412:1428]
		      Where [1437:1465]
		        Subject Identifier [1443:1453]
		        Property length [1454:1460]
		        Operator >= [1461:1463]
		        Value 6 [1464:1465]
		      Report [1474:1527]
		        Severity err [1481:1484]
		        At Identifier [1488:1498]
		        Message "Very long identifier found" [1499:1527]
		    Rule AnyIdentifier [1538:1642]
		      Match Identifier [1567:1583]
		      Report [1592:1636]
		        Severity info [1599:1603]
		        At Identifier [1607:1617]
		        Message "Identifier found" [1618:1636]
	`)

	// Act.
	gotDocument, gotErr := parser.Parse(source)

	// Assert.
	claim.Equal(t, "When parsing realistic DSL source, no parse error is returned.", "", formatParseError(gotErr), "Parse Error")
	claim.Equal(t, "When parsing realistic DSL source, document spans are preserved.", want, renderDocument(source, gotDocument, true), "Document")
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

	benchmark_Parse(b, fullHappyPathBenchmarkDSL(size))
}

func fullHappyPathDSL(size int) string {
	const scopeBlock = `    include "**/*.go"
    exclude "vendor/**"
`

	const definitionsBlock = `    lower = 'a'..'z'
    upper = 'A'..'Z'
    digit = '0'..'9'
    underscore = '_'
    letter = lower | upper
    identifierStart = letter | underscore
    identifierPart = letter | digit | underscore
    optionalSign = ('+' | '-')?
    repeatedLetter = letter*
    value = letter ('a' | 'b')+
`

	const tokensBlock = `    Identifier = identifierStart identifierPart*
    KeywordPublic = "public"
    KeywordInternal = "internal"
    SignedInteger = optionalSign digit+
    Whitespace = (' ' | '\t' | '\n' | '\r')+ skip
`

	const rulesBlock = `    rule PublicIdentifier {
        match Identifier
        where Identifier.text == "public"
        report warn at Identifier "Public identifier found"
    }
    rule InternalIdentifier {
        match Identifier
        where Identifier.text != "internal"
        report info at Identifier "Internal identifier found"
    }
    rule ShortIdentifier {
        match Identifier
        where Identifier.length < 3
        report info at Identifier "Short identifier found"
    }
    rule MediumIdentifier {
        match Identifier
        where Identifier.length <= 4
        report warn at Identifier "Medium identifier found"
    }
    rule LongIdentifier {
        match Identifier
        where Identifier.length > 5
        report err at Identifier "Long identifier found"
    }
    rule VeryLongIdentifier {
        match Identifier
        where Identifier.length >= 6
        report err at Identifier "Very long identifier found"
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

func fullHappyPathBenchmarkDSL(size int) string {
	const scopeBlock = `    include "**/*.go"
    exclude "vendor/**"
`

	const definitionsBlock = `    lower = 'a'..'z'
    upper = 'A'..'Z'
    digit = '0'..'9'
    underscore = '_'
    letter = lower | upper
    identifierStart = letter | underscore
    identifierPart = letter | digit | underscore
    optionalSign = ('+' | '-')?
    repeatedLetter = letter*
    value = letter ('a' | 'b')+
`

	const tokensBlock = `    Plus = '+'
    Lower = 'a'..'z'
    Identifier = identifierStart identifierPart*
    KeywordPublic = "public"
    KeywordInternal = "internal"
    Sign = "+" | "-"
    OptionalSign = ("+" | "-")?
    SignedInteger = optionalSign digit+
    Whitespace = (' ' | '\t' | '\n' | '\r')+ skip
`

	const rulesBlock = `    rule PublicIdentifier {
        match Identifier
        where Identifier.text == "public"
        report warn at Identifier "Public identifier found"
    }
    rule InternalIdentifier {
        match Identifier
        where Identifier.text != "internal"
        report info at Identifier "Internal identifier found"
    }
    rule ShortIdentifier {
        match Identifier
        where Identifier.length < 3
        report info at Identifier "Short identifier found"
    }
    rule MediumIdentifier {
        match Identifier
        where Identifier.length <= 4
        report warn at Identifier "Medium identifier found"
    }
    rule LongIdentifier {
        match Identifier
        where Identifier.length > 5
        report err at Identifier "Long identifier found"
    }
    rule VeryLongIdentifier {
        match Identifier
        where Identifier.length >= 6
        report err at Identifier "Very long identifier found"
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

// normalizeMultilineLiteral trims a raw multiline literal into the exact shape used by string comparisons in tests.
func normalizeMultilineLiteral(text string) string {
	lines := strings.Split(strings.TrimSpace(text), "\n")

	for idx := range lines {
		lines[idx] = strings.TrimRight(strings.TrimLeft(lines[idx], "\t"), " ")
	}

	return strings.Join(lines, "\n")
}

// renderDocument renders a parsed document into a stable tree-shaped string for assertions.
func renderDocument(source string, document ast.Document, includeSpans bool) string {
	var builder strings.Builder

	appendDocument(source, &builder, document, 0, includeSpans)

	return strings.TrimSpace(builder.String())
}

// formatParseError renders parser errors in the same compact span format used by assertions.
func formatParseError(err error) string {
	if err == nil {
		return ""
	}

	diagnostic, ok := err.(parser.Diagnostic)

	if !ok {
		return err.Error()
	}

	return diagnostic.Message +
		" [" +
		strconv.Itoa(int(diagnostic.Span.Start)) +
		":" +
		strconv.Itoa(int(diagnostic.Span.End)) +
		"]"
}

// appendDocument writes the full document tree into builder.
func appendDocument(source string, builder *strings.Builder, document ast.Document, depth int, includeSpans bool) {
	appendIndentedLine(builder, depth, formatLabelWithSpan("Document", document.Span, includeSpans))

	if document.Scope.Span != (location.Span{}) {
		appendIndentedLine(builder, depth+1, formatLabelWithSpan("Scope", document.Scope.Span, includeSpans))

		for _, entry := range document.ScopeSectionEntries(document.Scope) {
			label := "Include "
			if entry.Kind == ast.ScopeEntryExclude {
				label = "Exclude "
			}

			appendIndentedLine(builder, depth+2, formatLabelWithSpan(label+entry.Pattern.Value(source), entry.Span, includeSpans))
		}
	}

	if document.Definitions.Span != (location.Span{}) {
		appendIndentedLine(builder, depth+1, formatLabelWithSpan("Definitions", document.Definitions.Span, includeSpans))

		for _, definition := range document.SectionDefinitions(document.Definitions) {
			appendIndentedLine(builder, depth+2, formatLabelWithSpan("Definition "+definition.Name.Value(source), definition.Span, includeSpans))
			appendExpression(source, builder, document.Expressions, definition.Expression, depth+3, includeSpans)
		}
	}

	if document.Tokens.Span != (location.Span{}) {
		appendIndentedLine(builder, depth+1, formatLabelWithSpan("Tokens", document.Tokens.Span, includeSpans))

		for _, definition := range document.SectionTokens(document.Tokens) {
			appendIndentedLine(builder, depth+2, formatLabelWithSpan("Token "+definition.Name.Value(source), definition.Span, includeSpans))
			appendExpression(source, builder, document.Expressions, definition.Expression, depth+3, includeSpans)

			if definition.Skip.Kind == token.TokenSkip {
				appendIndentedLine(builder, depth+3, formatLabelWithSpan("Skip", definition.Skip.Span, includeSpans))
			}
		}
	}

	if document.Rules.Span != (location.Span{}) {
		appendIndentedLine(builder, depth+1, formatLabelWithSpan("Rules", document.Rules.Span, includeSpans))

		for _, rule := range document.SectionRules(document.Rules) {
			appendIndentedLine(builder, depth+2, formatLabelWithSpan("Rule "+rule.Name.Value(source), rule.Span, includeSpans))
			appendIndentedLine(builder, depth+3, formatLabelWithSpan("Match "+rule.Match.Token.Value(source), rule.Match.Span, includeSpans))

			if rule.Where.Span != (location.Span{}) {
				appendIndentedLine(builder, depth+3, formatLabelWithSpan("Where", rule.Where.Span, includeSpans))
				appendIndentedLine(builder, depth+4, formatLabelWithSpan("Subject "+rule.Where.Subject.Value(source), rule.Where.Subject.Span, includeSpans))
				appendIndentedLine(builder, depth+4, formatLabelWithSpan("Property "+rule.Where.Property.Value(source), rule.Where.Property.Span, includeSpans))
				appendIndentedLine(builder, depth+4, formatLabelWithSpan("Operator "+rule.Where.Operator.Value(source), rule.Where.Operator.Span, includeSpans))
				appendIndentedLine(builder, depth+4, formatLabelWithSpan("Value "+rule.Where.Value.Value(source), rule.Where.Value.Span, includeSpans))
			}

			appendIndentedLine(builder, depth+3, formatLabelWithSpan("Report", rule.Report.Span, includeSpans))
			appendIndentedLine(builder, depth+4, formatLabelWithSpan("Severity "+rule.Report.Severity.Value(source), rule.Report.Severity.Span, includeSpans))
			appendIndentedLine(builder, depth+4, formatLabelWithSpan("At "+rule.Report.Target.Value(source), rule.Report.Target.Span, includeSpans))
			appendIndentedLine(builder, depth+4, formatLabelWithSpan("Message "+rule.Report.Message.Value(source), rule.Report.Message.Span, includeSpans))
		}
	}
}

// appendExpression writes one expression subtree into builder.
func appendExpression(source string, builder *strings.Builder, expressions ast.DefinitionExpressionArena, expressionID ast.DefinitionExpressionID, depth int, includeSpans bool) {
	expression := expressions.Node(expressionID)

	switch expression.Kind {
	case ast.DefinitionExpressionReference:
		appendIndentedLine(builder, depth, formatLabelWithSpan("Reference "+expression.Start.Value(source), expression.Span, includeSpans))

	case ast.DefinitionExpressionCharacter:
		appendIndentedLine(builder, depth, formatLabelWithSpan("Character "+expression.Start.Value(source), expression.Span, includeSpans))

	case ast.DefinitionExpressionString:
		appendIndentedLine(builder, depth, formatLabelWithSpan("String "+expression.Start.Value(source), expression.Span, includeSpans))

	case ast.DefinitionExpressionRange:
		appendIndentedLine(builder, depth, formatLabelWithSpan("Range "+expression.Start.Value(source)+" "+expression.End.Value(source), expression.Span, includeSpans))

	case ast.DefinitionExpressionAlternation:
		appendIndentedLine(builder, depth, formatLabelWithSpan("Alternation", expression.Span, includeSpans))

		for _, childID := range expressions.Children(expression) {
			appendExpression(source, builder, expressions, childID, depth+1, includeSpans)
		}

	case ast.DefinitionExpressionConcatenation:
		appendIndentedLine(builder, depth, formatLabelWithSpan("Concatenation", expression.Span, includeSpans))

		for _, childID := range expressions.Children(expression) {
			appendExpression(source, builder, expressions, childID, depth+1, includeSpans)
		}

	case ast.DefinitionExpressionRepetition:
		appendIndentedLine(builder, depth, formatLabelWithSpan("Repetition "+expression.Operator.Value(source), expression.Span, includeSpans))
		appendExpression(source, builder, expressions, expressions.Children(expression)[0], depth+1, includeSpans)

	case ast.DefinitionExpressionGroup:
		appendIndentedLine(builder, depth, formatLabelWithSpan("Group", expression.Span, includeSpans))
		appendExpression(source, builder, expressions, expressions.Children(expression)[0], depth+1, includeSpans)
	}
}

// formatLabelWithSpan optionally appends a source span to a rendered tree label.
func formatLabelWithSpan(text string, span location.Span, includeSpan bool) string {
	if !includeSpan {
		return text
	}

	return text + " [" + strconv.Itoa(int(span.Start)) + ":" + strconv.Itoa(int(span.End)) + "]"
}

// appendIndentedLine writes one indented line into builder.
func appendIndentedLine(builder *strings.Builder, depth int, text string) {
	builder.WriteString(strings.Repeat("  ", depth))
	builder.WriteString(text)
	builder.WriteByte('\n')
}
