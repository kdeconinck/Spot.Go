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
		  Syntax
		    Node Word
		      Alternation
		        Reference Identifier
		        Reference KeywordPublic
		    Node WordPair
		      Concatenation
		        Capture left
		          Reference Word
		        Capture right
		          Reference Word
		    Node OptionalWord
		      Capture value
		        Repetition ?
		          Alternation
		            Reference Word
		            Reference KeywordInternal
		    Node WordList
		      Capture values
		        Repetition +
		          Reference Word
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
		Document [0:1974]
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
		  Syntax [579:908]
		    Node Word [592:678]
		      Alternation [612:678]
		        Reference Identifier [632:642]
		        Reference KeywordPublic [655:668]
		    Node WordPair [689:749]
		      Concatenation [713:749]
		        Capture left [713:723]
		          Reference Word [719:723]
		        Capture right [732:743]
		          Reference Word [739:743]
		    Node OptionalWord [754:852]
		      Capture value [782:852]
		        Repetition ? [787:852]
		          Alternation [790:852]
		            Reference Word [810:814]
		            Reference KeywordInternal [827:842]
		    Node WordList [863:900]
		      Capture values [887:900]
		        Repetition + [895:900]
		          Reference Word [895:899]
		  Rules [909:1974]
		    Rule PublicIdentifier [921:1077]
		      Match Identifier [953:969]
		      Where [978:1011]
		        Subject Identifier [984:994]
		        Property text [995:999]
		        Operator == [1000:1002]
		        Value "public" [1003:1011]
		      Report [1020:1071]
		        Severity warn [1027:1031]
		        At Identifier [1035:1045]
		        Message "Public identifier found" [1046:1071]
		    Rule InternalIdentifier [1082:1244]
		      Match Identifier [1116:1132]
		      Where [1141:1176]
		        Subject Identifier [1147:1157]
		        Property text [1158:1162]
		        Operator != [1163:1165]
		        Value "internal" [1166:1176]
		      Report [1185:1238]
		        Severity info [1192:1196]
		        At Identifier [1200:1210]
		        Message "Internal identifier found" [1211:1238]
		    Rule ShortIdentifier [1249:1397]
		      Match Identifier [1280:1296]
		      Where [1305:1332]
		        Subject Identifier [1311:1321]
		        Property length [1322:1328]
		        Operator < [1329:1330]
		        Value 3 [1331:1332]
		      Report [1341:1391]
		        Severity info [1348:1352]
		        At Identifier [1356:1366]
		        Message "Short identifier found" [1367:1391]
		    Rule MediumIdentifier [1402:1553]
		      Match Identifier [1434:1450]
		      Where [1459:1487]
		        Subject Identifier [1465:1475]
		        Property length [1476:1482]
		        Operator <= [1483:1485]
		        Value 4 [1486:1487]
		      Report [1496:1547]
		        Severity warn [1503:1507]
		        At Identifier [1511:1521]
		        Message "Medium identifier found" [1522:1547]
		    Rule LongIdentifier [1558:1703]
		      Match Identifier [1588:1604]
		      Where [1613:1640]
		        Subject Identifier [1619:1629]
		        Property length [1630:1636]
		        Operator > [1637:1638]
		        Value 5 [1639:1640]
		      Report [1649:1697]
		        Severity err [1656:1659]
		        At Identifier [1663:1673]
		        Message "Long identifier found" [1674:1697]
		    Rule VeryLongIdentifier [1708:1863]
		      Match Identifier [1742:1758]
		      Where [1767:1795]
		        Subject Identifier [1773:1783]
		        Property length [1784:1790]
		        Operator >= [1791:1793]
		        Value 6 [1794:1795]
		      Report [1804:1857]
		        Severity err [1811:1814]
		        At Identifier [1818:1828]
		        Message "Very long identifier found" [1829:1857]
		    Rule AnyIdentifier [1868:1972]
		      Match Identifier [1897:1913]
		      Report [1922:1966]
		        Severity info [1929:1933]
		        At Identifier [1937:1947]
		        Message "Identifier found" [1948:1966]
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

	const syntaxBlock = `    node Word {
        oneOf {
            Identifier
            KeywordPublic
        }
    }
    node WordPair {
        left: Word
        right: Word
    }
    node OptionalWord {
        value?: oneOf {
            Word
            KeywordInternal
        }
    }
    node WordList {
        values: Word+
    }
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
		"syntax {\n" +
		strings.Repeat(syntaxBlock, size) +
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

	const syntaxBlock = `    node Word {
        oneOf {
            Identifier
            KeywordPublic
        }
    }
    node WordPair {
        left: Word
        right: Word
    }
    node OptionalWord {
        value?: oneOf {
            Word
            KeywordInternal
        }
    }
    node WordList {
        values: Word+
    }
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
		"syntax {\n" +
		strings.Repeat(syntaxBlock, size) +
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

			if definition.Fallback.Kind == token.TokenFallback {
				appendIndentedLine(builder, depth+3, formatLabelWithSpan("Fallback", definition.Fallback.Span, includeSpans))
			} else {
				appendExpression(source, builder, document.Expressions, definition.Expression, depth+3, includeSpans)
			}

			if definition.Skip.Kind == token.TokenSkip {
				appendIndentedLine(builder, depth+3, formatLabelWithSpan("Skip", definition.Skip.Span, includeSpans))
			}
		}
	}

	if document.Syntax.Span != (location.Span{}) {
		appendIndentedLine(builder, depth+1, formatLabelWithSpan("Syntax", document.Syntax.Span, includeSpans))

		for _, syntaxNode := range document.SectionSyntaxNodes(document.Syntax) {
			appendIndentedLine(builder, depth+2, formatLabelWithSpan("Node "+syntaxNode.Name.Value(source), syntaxNode.Span, includeSpans))
			appendSyntaxExpression(source, builder, document.SyntaxExpressions, syntaxNode.Expression, depth+3, includeSpans)
		}
	}

	if document.Rules.Span != (location.Span{}) {
		appendIndentedLine(builder, depth+1, formatLabelWithSpan("Rules", document.Rules.Span, includeSpans))

		for _, rule := range document.SectionRules(document.Rules) {
			ruleLabel := "Rule"

			if rule.Name.Value(source) != "" {
				ruleLabel += " " + rule.Name.Value(source)
			}

			appendIndentedLine(builder, depth+2, formatLabelWithSpan(ruleLabel, rule.Span, includeSpans))
			matchLabel := "Match " + rule.Match.Target.Value(source)

			if rule.Match.Kind == ast.RuleMatchNode {
				matchLabel = "Match node " + rule.Match.Target.Value(source)
			}

			if rule.Match.RelationKind == ast.RuleMatchRelationAdjacentSibling {
				matchLabel = "Match " + rule.Match.RelatedTarget.Value(source) + " + " + rule.Match.Target.Value(source)
			}

			switch rule.Match.ScopeKind {
			case ast.RuleMatchScopeParent:
				matchLabel += " parent " + rule.Match.ScopeTarget.Value(source)

			case ast.RuleMatchScopeInside:
				matchLabel += " inside " + rule.Match.ScopeTarget.Value(source)

			case ast.RuleMatchScopeParentOutside:
				matchLabel += " outside parent " + rule.Match.ScopeTarget.Value(source)

			case ast.RuleMatchScopeOutside:
				matchLabel += " outside " + rule.Match.ScopeTarget.Value(source)
			}

			appendIndentedLine(builder, depth+3, formatLabelWithSpan(matchLabel, rule.Match.Span, includeSpans))

			if rule.Where.Span != (location.Span{}) {
				appendIndentedLine(builder, depth+3, formatLabelWithSpan("Where", rule.Where.Span, includeSpans))
				appendIndentedLine(builder, depth+4, formatLabelWithSpan("Subject "+rule.Where.Subject.Value(source), rule.Where.Subject.Span, includeSpans))
				for _, pathSegment := range rule.Where.Path {
					appendIndentedLine(builder, depth+4, formatLabelWithSpan("Path "+pathSegment.Value(source), pathSegment.Span, includeSpans))
				}
				appendIndentedLine(builder, depth+4, formatLabelWithSpan("Property "+rule.Where.Property.Value(source), rule.Where.Property.Span, includeSpans))
				appendIndentedLine(builder, depth+4, formatLabelWithSpan("Operator "+rule.Where.Operator.Value(source), rule.Where.Operator.Span, includeSpans))

				if rule.Where.OtherProperty.Value(source) != "" {
					appendIndentedLine(builder, depth+4, formatLabelWithSpan("Other Subject "+rule.Where.OtherSubject.Value(source), rule.Where.OtherSubject.Span, includeSpans))
					for _, pathSegment := range rule.Where.OtherPath {
						appendIndentedLine(builder, depth+4, formatLabelWithSpan("Other Path "+pathSegment.Value(source), pathSegment.Span, includeSpans))
					}
					appendIndentedLine(builder, depth+4, formatLabelWithSpan("Other Property "+rule.Where.OtherProperty.Value(source), rule.Where.OtherProperty.Span, includeSpans))
				} else {
					appendIndentedLine(builder, depth+4, formatLabelWithSpan("Value "+rule.Where.Value.Value(source), rule.Where.Value.Span, includeSpans))
				}
			}

			appendIndentedLine(builder, depth+3, formatLabelWithSpan("Report", rule.Report.Span, includeSpans))
			appendIndentedLine(builder, depth+4, formatLabelWithSpan("Severity "+rule.Report.Severity.Value(source), rule.Report.Severity.Span, includeSpans))
			appendIndentedLine(builder, depth+4, formatLabelWithSpan("At "+rule.Report.Target.Value(source), rule.Report.Target.Span, includeSpans))
			appendIndentedLine(builder, depth+4, formatLabelWithSpan("Message "+rule.Report.Message.Value(source), rule.Report.Message.Span, includeSpans))
		}
	}
}

// appendSyntaxExpression writes one syntax expression subtree into builder.
func appendSyntaxExpression(source string, builder *strings.Builder, expressions ast.SyntaxExpressionArena, expressionID ast.SyntaxExpressionID, depth int, includeSpans bool) {
	expression := expressions.Node(expressionID)

	switch expression.Kind {
	case ast.SyntaxExpressionReference:
		appendIndentedLine(builder, depth, formatLabelWithSpan("Reference "+expression.Reference.Value(source), expression.Span, includeSpans))

	case ast.SyntaxExpressionAny:
		appendIndentedLine(builder, depth, formatLabelWithSpan("Any", expression.Span, includeSpans))

	case ast.SyntaxExpressionCapture:
		appendIndentedLine(builder, depth, formatLabelWithSpan("Capture "+expression.Field.Value(source), expression.Span, includeSpans))
		appendSyntaxExpression(source, builder, expressions, expressions.Children(expression)[0], depth+1, includeSpans)

	case ast.SyntaxExpressionAlternation:
		appendIndentedLine(builder, depth, formatLabelWithSpan("Alternation", expression.Span, includeSpans))

		for _, childID := range expressions.Children(expression) {
			appendSyntaxExpression(source, builder, expressions, childID, depth+1, includeSpans)
		}

	case ast.SyntaxExpressionConcatenation:
		appendIndentedLine(builder, depth, formatLabelWithSpan("Concatenation", expression.Span, includeSpans))

		for _, childID := range expressions.Children(expression) {
			appendSyntaxExpression(source, builder, expressions, childID, depth+1, includeSpans)
		}

	case ast.SyntaxExpressionRepetition:
		appendIndentedLine(builder, depth, formatLabelWithSpan("Repetition "+expression.Operator.Value(source), expression.Span, includeSpans))
		appendSyntaxExpression(source, builder, expressions, expressions.Children(expression)[0], depth+1, includeSpans)

	case ast.SyntaxExpressionGroup:
		appendIndentedLine(builder, depth, formatLabelWithSpan("Group", expression.Span, includeSpans))
		appendSyntaxExpression(source, builder, expressions, expressions.Children(expression)[0], depth+1, includeSpans)
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
