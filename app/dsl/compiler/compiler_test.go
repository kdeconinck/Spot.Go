// =====================================================================================================================
// == LICENSE:                 Copyright (c) 2026 Kevin De Coninck.
// == SPDX-License-Identifier: LicenseRef-PolyForm-Noncommercial-1.0.0
// =====================================================================================================================

// Verify the public API of the compiler package.
//
// Tests in this package are written against the exported API only.
// This ensures that behavior is tested through the same surface that external consumers would use.
package compiler_test

import (
	"strconv"
	"strings"
	"testing"

	"github.com/kdeconinck/spot/dsl/compiler"
	"github.com/kdeconinck/spot/dsl/parser"
	"github.com/kdeconinck/spot/dsl/resolver"
	"github.com/kdeconinck/spot/dsl/validator"
	"github.com/kdeconinck/spot/qa/claim"
	"github.com/kdeconinck/spot/runtime/ir"
)

func Test_Compile_DSL(t *testing.T) {
	t.Parallel()

	for tcName, tc := range map[string]struct {
		inSource    string
		wantProgram string
	}{
		"When compiling a full DSL file, a program is returned.": {
			inSource: dsl(0),
			wantProgram: normalizeMultilineLiteral(`
				Program
				  Tokens
				    Token Identifier
				      Concatenation
				        Alternation
				          Range 'a' 'z'
				          Range 'A' 'Z'
				        Repetition *
				          Alternation
				            Alternation
				              Range 'a' 'z'
				              Range 'A' 'Z'
				            Range '0' '9'
				            Character '_'
				    Token KeywordPublic
				      String "public"
				    Token KeywordInternal
				      String "internal"
				    Token Whitespace
				      Repetition +
				        Alternation
				          Character ' '
				          Character '\t'
				      Skip
				  Syntax
				    Node Word
				      Alternation
				        Token Identifier
				        Token KeywordPublic
				    Node WordPair
				      Concatenation
				        Capture left
				          Node Word
				        Capture right
				          Node Word
				    Node OptionalWord
				      Capture value
				        Repetition ?
				          Alternation
				            Node Word
				            Token KeywordInternal
				    Node WordList
				      Capture values
				        Repetition +
				          Node Word
				    Node Root
				      Concatenation
				        Capture optional
				          Node OptionalWord
				        Capture pair
				          Node WordPair
				        Capture list
				          Node WordList
				  Rules
				    Rule PublicIdentifier
				      MatchToken Identifier
				      Where text == "public"
				      Report warn at Identifier "Public identifier found"
				    Rule RootRule
				      MatchNode Root
				      Where length > 0
				      Report info at Root "Root found"
			`),
		},
	} {
		t.Run(tcName, func(t *testing.T) {
			t.Parallel()

			// Arrange.
			document, parseErr := parser.Parse(tc.inSource)
			resolution := resolver.Resolve(tc.inSource, document)
			validationDiagnostics := validator.Validate(tc.inSource, resolution)

			// Act.
			gotProgram := compiler.Compile(tc.inSource, resolution)

			// Assert.
			claim.Equal(t, tcName, error(nil), parseErr, "Parse Error")
			claim.Equal(t, tcName, 0, len(validationDiagnostics), "Validation Diagnostic Count")
			claim.Equal(t, tcName, tc.wantProgram, renderProgram(gotProgram), "Program")
		})
	}
}

func Benchmark_Compile_DSL_0(b *testing.B)    { benchmark_Compile_DSL(b, 0) }
func Benchmark_Compile_DSL_1(b *testing.B)    { benchmark_Compile_DSL(b, 1) }
func Benchmark_Compile_DSL_10(b *testing.B)   { benchmark_Compile_DSL(b, 10) }
func Benchmark_Compile_DSL_100(b *testing.B)  { benchmark_Compile_DSL(b, 100) }
func Benchmark_Compile_DSL_1000(b *testing.B) { benchmark_Compile_DSL(b, 1000) }

func benchmark_Compile_DSL(b *testing.B, size int) {
	b.Helper()

	source := dsl(size)
	document, parseErr := parser.Parse(source)
	resolution := resolver.Resolve(source, document)
	validationDiagnostics := validator.Validate(source, resolution)
	claim.Equal(b, "Compile DSL benchmark parse error.", error(nil), parseErr, "Parse Error")
	claim.Equal(b, "Compile DSL benchmark validation diagnostics.", 0, len(validationDiagnostics), "Validation Diagnostic Count")

	for b.Loop() {
		_ = compiler.Compile(source, resolution)
	}
}

func dsl(size int) string {
	var sb strings.Builder

	sb.WriteString("scope {\n")
	sb.WriteString("    include \"**/*.go\"\n")
	sb.WriteString("    exclude \"vendor/**\"\n")

	for range size {
		sb.WriteString("    include \"**/*.go\"\n")
		sb.WriteString("    exclude \"vendor/**\"\n")
	}

	sb.WriteString("}\n")
	sb.WriteString("definitions {\n")
	sb.WriteString("    letter = 'a'..'z' | 'A'..'Z'\n")
	sb.WriteString("    digit = '0'..'9'\n")
	sb.WriteString("    identifierStart = letter\n")
	sb.WriteString("    identifier = identifierStart (identifierStart | digit | '_')*\n")

	for idx := 1; idx <= size; idx++ {
		sb.WriteString("    letter")
		sb.WriteString(strconv.Itoa(idx))
		sb.WriteString(" = 'a'..'z' | 'A'..'Z'\n")
		sb.WriteString("    digit")
		sb.WriteString(strconv.Itoa(idx))
		sb.WriteString(" = '0'..'9'\n")
		sb.WriteString("    identifierStart")
		sb.WriteString(strconv.Itoa(idx))
		sb.WriteString(" = letter")
		sb.WriteString(strconv.Itoa(idx))
		sb.WriteString("\n")
		sb.WriteString("    identifier")
		sb.WriteString(strconv.Itoa(idx))
		sb.WriteString(" = identifierStart")
		sb.WriteString(strconv.Itoa(idx))
		sb.WriteString(" (identifierStart")
		sb.WriteString(strconv.Itoa(idx))
		sb.WriteString(" | digit")
		sb.WriteString(strconv.Itoa(idx))
		sb.WriteString(" | '_')*\n")
	}

	sb.WriteString("}\n")
	sb.WriteString("tokens {\n")
	sb.WriteString("    Identifier = identifier\n")
	sb.WriteString("    KeywordPublic = \"public\"\n")
	sb.WriteString("    KeywordInternal = \"internal\"\n")
	sb.WriteString("    Whitespace = (' ' | '\\t')+ skip\n")

	for idx := 1; idx <= size; idx++ {
		sb.WriteString("    Identifier")
		sb.WriteString(strconv.Itoa(idx))
		sb.WriteString(" = identifier")
		sb.WriteString(strconv.Itoa(idx))
		sb.WriteString("\n")
		sb.WriteString("    KeywordPublic")
		sb.WriteString(strconv.Itoa(idx))
		sb.WriteString(" = \"public\"\n")
		sb.WriteString("    KeywordInternal")
		sb.WriteString(strconv.Itoa(idx))
		sb.WriteString(" = \"internal\"\n")
		sb.WriteString("    Whitespace")
		sb.WriteString(strconv.Itoa(idx))
		sb.WriteString(" = (' ' | '\\t')+ skip\n")
	}

	sb.WriteString("}\n")
	sb.WriteString("syntax {\n")
	sb.WriteString("    node Word {\n")
	sb.WriteString("        oneOf {\n")
	sb.WriteString("            Identifier\n")
	sb.WriteString("            KeywordPublic\n")
	sb.WriteString("        }\n")
	sb.WriteString("    }\n")
	sb.WriteString("    node WordPair {\n")
	sb.WriteString("        left: Word\n")
	sb.WriteString("        right: Word\n")
	sb.WriteString("    }\n")
	sb.WriteString("    node OptionalWord {\n")
	sb.WriteString("        value?: oneOf {\n")
	sb.WriteString("            Word\n")
	sb.WriteString("            KeywordInternal\n")
	sb.WriteString("        }\n")
	sb.WriteString("    }\n")
	sb.WriteString("    node WordList {\n")
	sb.WriteString("        values: Word+\n")
	sb.WriteString("    }\n")
	sb.WriteString("    node Root {\n")
	sb.WriteString("        optional: OptionalWord\n")
	sb.WriteString("        pair: WordPair\n")
	sb.WriteString("        list: WordList\n")
	sb.WriteString("    }\n")

	for idx := 1; idx <= size; idx++ {
		sb.WriteString("    node Word")
		sb.WriteString(strconv.Itoa(idx))
		sb.WriteString(" {\n")
		sb.WriteString("        oneOf {\n")
		sb.WriteString("            Identifier")
		sb.WriteString(strconv.Itoa(idx))
		sb.WriteString("\n")
		sb.WriteString("            KeywordPublic")
		sb.WriteString(strconv.Itoa(idx))
		sb.WriteString("\n")
		sb.WriteString("        }\n")
		sb.WriteString("    }\n")
		sb.WriteString("    node WordPair")
		sb.WriteString(strconv.Itoa(idx))
		sb.WriteString(" {\n")
		sb.WriteString("        left: Word")
		sb.WriteString(strconv.Itoa(idx))
		sb.WriteString("\n")
		sb.WriteString("        right: Word")
		sb.WriteString(strconv.Itoa(idx))
		sb.WriteString("\n")
		sb.WriteString("    }\n")
		sb.WriteString("    node OptionalWord")
		sb.WriteString(strconv.Itoa(idx))
		sb.WriteString(" {\n")
		sb.WriteString("        value?: oneOf {\n")
		sb.WriteString("            Word")
		sb.WriteString(strconv.Itoa(idx))
		sb.WriteString("\n")
		sb.WriteString("            KeywordInternal")
		sb.WriteString(strconv.Itoa(idx))
		sb.WriteString("\n")
		sb.WriteString("        }\n")
		sb.WriteString("    }\n")
		sb.WriteString("    node WordList")
		sb.WriteString(strconv.Itoa(idx))
		sb.WriteString(" {\n")
		sb.WriteString("        values: Word")
		sb.WriteString(strconv.Itoa(idx))
		sb.WriteString("+\n")
		sb.WriteString("    }\n")
		sb.WriteString("    node Root")
		sb.WriteString(strconv.Itoa(idx))
		sb.WriteString(" {\n")
		sb.WriteString("        optional: OptionalWord")
		sb.WriteString(strconv.Itoa(idx))
		sb.WriteString("\n")
		sb.WriteString("        pair: WordPair")
		sb.WriteString(strconv.Itoa(idx))
		sb.WriteString("\n")
		sb.WriteString("        list: WordList")
		sb.WriteString(strconv.Itoa(idx))
		sb.WriteString("\n")
		sb.WriteString("    }\n")
	}

	sb.WriteString("}\n")
	sb.WriteString("rules {\n")
	sb.WriteString("    rule PublicIdentifier {\n")
	sb.WriteString("        match Identifier\n")
	sb.WriteString("        where Identifier.text == \"public\"\n")
	sb.WriteString("        report warn at Identifier \"Public identifier found\"\n")
	sb.WriteString("    }\n")
	sb.WriteString("    rule RootRule {\n")
	sb.WriteString("        match node Root\n")
	sb.WriteString("        where Root.length > 0\n")
	sb.WriteString("        report info at Root \"Root found\"\n")
	sb.WriteString("    }\n")

	for idx := 1; idx <= size; idx++ {
		sb.WriteString("    rule PublicIdentifier")
		sb.WriteString(strconv.Itoa(idx))
		sb.WriteString(" {\n")
		sb.WriteString("        match Identifier")
		sb.WriteString(strconv.Itoa(idx))
		sb.WriteString("\n")
		sb.WriteString("        where Identifier")
		sb.WriteString(strconv.Itoa(idx))
		sb.WriteString(".text == \"public\"\n")
		sb.WriteString("        report warn at Identifier")
		sb.WriteString(strconv.Itoa(idx))
		sb.WriteString(" \"Public identifier found\"\n")
		sb.WriteString("    }\n")
		sb.WriteString("    rule RootRule")
		sb.WriteString(strconv.Itoa(idx))
		sb.WriteString(" {\n")
		sb.WriteString("        match node Root")
		sb.WriteString(strconv.Itoa(idx))
		sb.WriteString("\n")
		sb.WriteString("        where Root")
		sb.WriteString(strconv.Itoa(idx))
		sb.WriteString(".length > 0\n")
		sb.WriteString("        report info at Root")
		sb.WriteString(strconv.Itoa(idx))
		sb.WriteString(" \"Root found\"\n")
		sb.WriteString("    }\n")
	}

	sb.WriteString("}")

	return sb.String()
}

func normalizeMultilineLiteral(text string) string {
	lines := strings.Split(strings.TrimSpace(text), "\n")

	for idx := range lines {
		lines[idx] = strings.TrimRight(strings.TrimLeft(lines[idx], "\t"), " ")
	}

	return strings.Join(lines, "\n")
}

func renderProgram(program ir.Program) string {
	var builder strings.Builder

	builder.WriteString("Program\n")
	builder.WriteString("  Tokens\n")

	for idx := range program.Tokens {
		tok := program.Tokens[idx]
		builder.WriteString("    Token ")
		builder.WriteString(tok.Name)
		builder.WriteByte('\n')
		if tok.Fallback {
			appendIndentedLine(&builder, 3, "Fallback")
		} else {
			appendExpression(&builder, program, tok.Expression, 3)
		}

		if tok.Skip {
			appendIndentedLine(&builder, 3, "Skip")
		}
	}

	if len(program.SyntaxNodes) > 0 {
		builder.WriteString("  Syntax\n")

		for idx := range program.SyntaxNodes {
			syntaxNode := program.SyntaxNodes[idx]
			appendIndentedLine(&builder, 2, "Node "+syntaxNode.Name)
			appendSyntaxExpression(&builder, program, syntaxNode.Expression, 3)
		}
	}

	builder.WriteString("  Rules\n")

	for idx := range program.Rules {
		rule := program.Rules[idx]
		ruleLabel := "Rule"

		if rule.Name != "" {
			ruleLabel += " " + rule.Name
		}

		appendIndentedLine(&builder, 2, ruleLabel)
		appendIndentedLine(&builder, 3, renderRuleMatch(program, rule))
		appendIndentedLine(&builder, 3, "Where "+renderCondition(program, rule.Where))
		appendIndentedLine(&builder, 3, "Report "+renderSeverity(rule.Report.Severity)+" at "+renderReportTarget(program, rule.Report)+" "+strconv.Quote(rule.Report.Message))
	}

	return strings.TrimSpace(builder.String())
}

func renderRuleMatch(program ir.Program, rule ir.Rule) string {
	label := ""

	if rule.RelationKind == ir.RuleMatchRelationAdjacentSibling {
		return "Match " + program.SyntaxNodes[rule.RelatedMatchIndex].Name + " + " + program.SyntaxNodes[rule.MatchIndex].Name
	}

	if rule.MatchKind == ir.RuleMatchSyntaxNode {
		label = "MatchNode " + program.SyntaxNodes[rule.MatchIndex].Name
	} else {
		label = "MatchToken " + program.Tokens[rule.MatchIndex].Name
	}

	switch rule.MatchScopeKind {
	case ir.RuleMatchScopeParent:
		label += " parent " + program.SyntaxNodes[rule.MatchScopeIndex].Name

	case ir.RuleMatchScopeInside:
		label += " inside " + program.SyntaxNodes[rule.MatchScopeIndex].Name

	case ir.RuleMatchScopeParentOutside:
		label += " outside parent " + program.SyntaxNodes[rule.MatchScopeIndex].Name

	case ir.RuleMatchScopeOutside:
		label += " outside " + program.SyntaxNodes[rule.MatchScopeIndex].Name
	}

	return label
}

func appendExpression(builder *strings.Builder, program ir.Program, expressionID ir.ExpressionID, depth int) {
	expression := program.Expressions.Node(expressionID)

	switch expression.Kind {
	case ir.ExpressionCharacter:
		appendIndentedLine(builder, depth, "Character "+strconv.QuoteRuneToASCII(rune(expression.Character)))

	case ir.ExpressionString:
		appendIndentedLine(builder, depth, "String "+strconv.Quote(program.Expressions.String(expression.StringID)))

	case ir.ExpressionRange:
		appendIndentedLine(
			builder,
			depth,
			"Range "+strconv.QuoteRuneToASCII(rune(expression.RangeStart))+" "+strconv.QuoteRuneToASCII(rune(expression.RangeEnd)),
		)

	case ir.ExpressionReference, ir.ExpressionGroup:
		appendExpression(builder, program, firstExpressionChild(program.Expressions, expression), depth)

	case ir.ExpressionConcatenation:
		appendIndentedLine(builder, depth, "Concatenation")

		for _, childID := range program.Expressions.Children(expression) {
			appendExpression(builder, program, childID, depth+1)
		}

	case ir.ExpressionAlternation:
		appendIndentedLine(builder, depth, "Alternation")

		for _, childID := range program.Expressions.Children(expression) {
			appendExpression(builder, program, childID, depth+1)
		}

	default:
		appendIndentedLine(builder, depth, "Repetition "+renderRepetition(expression.Repetition))
		appendExpression(builder, program, firstExpressionChild(program.Expressions, expression), depth+1)
	}
}

func appendSyntaxExpression(builder *strings.Builder, program ir.Program, expressionID ir.SyntaxExpressionID, depth int) {
	expression := program.SyntaxExpressions.Node(expressionID)

	switch expression.Kind {
	case ir.SyntaxExpressionReference:
		label := "Token "

		if expression.ReferenceKind == ir.SyntaxReferenceNode {
			label = "Node "
		}

		appendIndentedLine(builder, depth, label+syntaxReferenceName(program, expression))

	case ir.SyntaxExpressionAny:
		appendIndentedLine(builder, depth, "Any")

	case ir.SyntaxExpressionCapture:
		appendIndentedLine(builder, depth, "Capture "+program.SyntaxFields[expression.FieldID])
		appendSyntaxExpression(builder, program, firstSyntaxExpressionChild(program.SyntaxExpressions, expression), depth+1)

	case ir.SyntaxExpressionConcatenation:
		appendIndentedLine(builder, depth, "Concatenation")

		for _, childID := range program.SyntaxExpressions.Children(expression) {
			appendSyntaxExpression(builder, program, childID, depth+1)
		}

	case ir.SyntaxExpressionAlternation:
		appendIndentedLine(builder, depth, "Alternation")

		for _, childID := range program.SyntaxExpressions.Children(expression) {
			appendSyntaxExpression(builder, program, childID, depth+1)
		}

	case ir.SyntaxExpressionGroup:
		appendIndentedLine(builder, depth, "Group")
		appendSyntaxExpression(builder, program, firstSyntaxExpressionChild(program.SyntaxExpressions, expression), depth+1)

	default:
		appendIndentedLine(builder, depth, "Repetition "+renderRepetition(expression.Repetition))
		appendSyntaxExpression(builder, program, firstSyntaxExpressionChild(program.SyntaxExpressions, expression), depth+1)
	}
}

func firstExpressionChild(arena ir.ExpressionArena, expression ir.ExpressionNode) ir.ExpressionID {
	if expression.Kind == ir.ExpressionReference {
		return expression.Reference
	}

	return arena.Children(expression)[0]
}

func firstSyntaxExpressionChild(arena ir.SyntaxExpressionArena, expression ir.SyntaxExpressionNode) ir.SyntaxExpressionID {
	return arena.Children(expression)[0]
}

func syntaxReferenceName(program ir.Program, expression ir.SyntaxExpressionNode) string {
	if expression.ReferenceKind == ir.SyntaxReferenceToken {
		return program.Tokens[expression.Reference].Name
	}

	return program.SyntaxNodes[expression.Reference].Name
}

func renderRepetition(repetition ir.RepetitionKind) string {
	switch repetition {
	case ir.RepetitionZeroOrOne:
		return "?"

	case ir.RepetitionZeroOrMore:
		return "*"

	default:
		return "+"
	}
}

func renderCondition(program ir.Program, condition ir.Condition) string {
	switch condition.LeftProperty {
	case ir.ConditionPropertyNone:
		return "none"

	case ir.ConditionPropertyText:
		return renderConditionValue(program, condition.LeftSubject, condition.LeftPath, condition.LeftProperty) + " " + renderOperator(condition.Operator) + " " + renderRightConditionValue(program, condition)

	default:
		return renderConditionValue(program, condition.LeftSubject, condition.LeftPath, condition.LeftProperty) + " " + renderOperator(condition.Operator) + " " + renderRightConditionValue(program, condition)
	}
}

func renderConditionValue(program ir.Program, subject ir.ConditionSubjectKind, path []uint32, property ir.ConditionProperty) string {
	if len(path) == 0 {
		if subject == ir.ConditionSubjectMatch {
			return renderConditionProperty(property)
		}

		return renderConditionSubject(subject) + "." + renderConditionProperty(property)
	}

	return renderConditionSubject(subject) + "." + strings.Join(renderFieldPath(program, path), ".") + "." + renderConditionProperty(property)
}

func renderRightConditionValue(program ir.Program, condition ir.Condition) string {
	if condition.RightSubject != ir.ConditionSubjectNone {
		return renderConditionValue(program, condition.RightSubject, condition.RightPath, condition.RightProperty)
	}

	if condition.LeftProperty == ir.ConditionPropertyText {
		return strconv.Quote(condition.String)
	}

	return strconv.Itoa(condition.Integer)
}

func renderFieldPath(program ir.Program, path []uint32) []string {
	rendered := make([]string, 0, len(path))

	for idx := range path {
		rendered = append(rendered, program.SyntaxFields[path[idx]])
	}

	return rendered
}

func renderConditionSubject(subject ir.ConditionSubjectKind) string {
	switch subject {
	case ir.ConditionSubjectRelatedMatch:
		return "left"

	case ir.ConditionSubjectGap:
		return "gap"

	default:
		return "right"
	}
}

func renderConditionProperty(property ir.ConditionProperty) string {
	switch property {
	case ir.ConditionPropertyText:
		return "text"

	case ir.ConditionPropertyBlankLines:
		return "blankLines"

	default:
		return "length"
	}
}

func renderOperator(operator ir.ConditionOperator) string {
	switch operator {
	case ir.ConditionOperatorEqual:
		return "=="

	case ir.ConditionOperatorNotEqual:
		return "!="

	case ir.ConditionOperatorLess:
		return "<"

	case ir.ConditionOperatorLessEqual:
		return "<="

	case ir.ConditionOperatorGreater:
		return ">"

	case ir.ConditionOperatorStartsWith:
		return "startsWith"

	default:
		return ">="
	}
}

func renderSeverity(severity ir.Severity) string {
	switch severity {
	case ir.SeverityInfo:
		return "info"

	case ir.SeverityWarn:
		return "warn"

	default:
		return "err"
	}
}

func renderReportTarget(program ir.Program, report ir.Report) string {
	if report.TargetKind == ir.RuleMatchSyntaxNode {
		return program.SyntaxNodes[report.TargetIndex].Name
	}

	return program.Tokens[report.TargetIndex].Name
}

func appendIndentedLine(builder *strings.Builder, depth int, text string) {
	builder.WriteString(strings.Repeat("  ", depth))
	builder.WriteString(text)
	builder.WriteByte('\n')
}
