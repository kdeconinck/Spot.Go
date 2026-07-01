// =====================================================================================================================
// == LICENSE:                 Copyright (c) 2026 Kevin De Coninck.
// == SPDX-License-Identifier: LicenseRef-PolyForm-Noncommercial-1.0.0
// =====================================================================================================================

// Package compiler compiles validated Spot DSL syntax into runtime-oriented data structures.
package compiler

import (
	"strconv"

	"github.com/kdeconinck/spot/dsl/ast"
	"github.com/kdeconinck/spot/dsl/resolver"
	"github.com/kdeconinck/spot/dsl/token"
	"github.com/kdeconinck/spot/runtime/ir"
)

// Compile compiles validated resolved Spot DSL syntax into a runtime program.
func Compile(source string, resolution resolver.Resolution) ir.Program {
	expressions := resolution.Document.Expressions
	syntaxExpressions := resolution.Document.SyntaxExpressions
	tokenList := resolution.Tokens
	syntaxNodes := resolution.SyntaxNodes
	ruleList := resolution.Rules
	program := ir.Program{
		Tokens: make([]ir.Token, 0, len(tokenList)),
		Expressions: ir.ExpressionArena{
			Nodes:    make([]ir.ExpressionNode, len(expressions.Nodes)),
			ChildIDs: make([]ir.ExpressionID, len(expressions.ChildIDs)),
			Strings:  make([]string, 0, countStringExpressions(expressions)),
		},
		SyntaxNodes: make([]ir.SyntaxNode, 0, len(syntaxNodes)),
		SyntaxExpressions: ir.SyntaxExpressionArena{
			Nodes:    make([]ir.SyntaxExpressionNode, len(syntaxExpressions.Nodes)),
			ChildIDs: make([]ir.SyntaxExpressionID, len(syntaxExpressions.ChildIDs)),
		},
		SyntaxRoot: findSyntaxRootIndex(source, resolution),
		Rules:      make([]ir.Rule, 0, len(ruleList)),
	}

	fieldIDs := make(map[string]uint32)
	ensureFieldID := func(name string) uint32 {
		if id, ok := fieldIDs[name]; ok {
			return id
		}

		id := uint32(len(program.SyntaxFields))
		fieldIDs[name] = id
		program.SyntaxFields = append(program.SyntaxFields, name)

		return id
	}

	copy(program.Expressions.ChildIDs, reinterpretExpressionChildren(expressions.ChildIDs))
	compileExpressionArena(source, resolution, &program.Expressions)
	copy(program.SyntaxExpressions.ChildIDs, reinterpretSyntaxExpressionChildren(syntaxExpressions.ChildIDs))
	compileSyntaxExpressionArena(source, resolution, &program.SyntaxExpressions, ensureFieldID)

	for idx := range tokenList {
		tok := tokenList[idx]
		program.Tokens = append(program.Tokens, ir.Token{
			Name:       tok.Name.Value(source),
			Expression: ir.ExpressionID(tok.Expression),
			Fallback:   tok.Fallback.Kind == token.TokenFallback,
			Skip:       tok.Skip.Kind == token.TokenSkip,
		})
	}

	for idx := range syntaxNodes {
		syntaxNode := syntaxNodes[idx]
		program.SyntaxNodes = append(program.SyntaxNodes, ir.SyntaxNode{
			Name:       syntaxNode.Name.Value(source),
			Expression: ir.SyntaxExpressionID(syntaxNode.Expression),
		})
	}

	for idx := range ruleList {
		program.Rules = append(program.Rules, compileRule(source, ruleList[idx], resolution, ensureFieldID))
	}

	return program
}

func compileRule(source string, rule ast.Rule, resolution resolver.Resolution, ensureFieldID func(string) uint32) ir.Rule {
	matchKind := ir.RuleMatchToken
	matchIndex := 0
	relationKind := ir.RuleMatchRelationNone
	relatedMatchIndex := 0
	matchScopeKind := ir.RuleMatchScopeNone
	matchScopeIndex := 0

	if rule.Match.Kind == ast.RuleMatchNode {
		matchKind = ir.RuleMatchSyntaxNode
		matchIndex, _ = resolution.SyntaxNodeIndex(rule.Match.Target.Value(source))
		if rule.Match.RelationKind == ast.RuleMatchRelationAdjacentSibling {
			relationKind = ir.RuleMatchRelationAdjacentSibling
			relatedMatchIndex, _ = resolution.SyntaxNodeIndex(rule.Match.RelatedTarget.Value(source))
		}

		switch rule.Match.ScopeKind {
		case ast.RuleMatchScopeParent:
			matchScopeKind = ir.RuleMatchScopeParent

		case ast.RuleMatchScopeInside:
			matchScopeKind = ir.RuleMatchScopeInside

		case ast.RuleMatchScopeParentOutside:
			matchScopeKind = ir.RuleMatchScopeParentOutside

		case ast.RuleMatchScopeOutside:
			matchScopeKind = ir.RuleMatchScopeOutside
		}

		if matchScopeKind != ir.RuleMatchScopeNone {
			matchScopeIndex, _ = resolution.SyntaxNodeIndex(rule.Match.ScopeTarget.Value(source))
		}
	} else {
		matchIndex, _ = resolution.TokenIndex(rule.Match.Target.Value(source))
	}

	return ir.Rule{
		Name:              rule.Name.Value(source),
		MatchKind:         matchKind,
		MatchIndex:        matchIndex,
		RelationKind:      relationKind,
		RelatedMatchIndex: relatedMatchIndex,
		MatchScopeKind:    matchScopeKind,
		MatchScopeIndex:   matchScopeIndex,
		Where:             compileCondition(source, rule.Where, ensureFieldID),
		Report:            compileReport(source, rule, resolution),
	}
}

func compileCondition(source string, condition ast.RuleCondition, ensureFieldID func(string) uint32) ir.Condition {
	if condition.Property.Value(source) == "" {
		return ir.Condition{
			LeftSubject:  ir.ConditionSubjectNone,
			LeftProperty: ir.ConditionPropertyNone,
		}
	}

	compiled := ir.Condition{
		LeftSubject:  conditionSubject(source, condition.Subject),
		LeftPath:     compileConditionPath(source, condition.Path, ensureFieldID),
		LeftProperty: conditionProperty(source, condition.Property),
		Operator:     conditionOperator(condition.Operator),
	}

	if condition.OtherProperty.Value(source) != "" {
		compiled.RightSubject = conditionSubject(source, condition.OtherSubject)
		compiled.RightPath = compileConditionPath(source, condition.OtherPath, ensureFieldID)
		compiled.RightProperty = conditionProperty(source, condition.OtherProperty)

		return compiled
	}

	if compiled.LeftProperty == ir.ConditionPropertyText {
		compiled.String = stringValue(source, condition.Value)

		return compiled
	}

	compiled.Integer = integerValue(source, condition.Value)

	return compiled
}

func compileReport(source string, rule ast.Rule, resolution resolver.Resolution) ir.Report {
	report := rule.Report
	targetKind := ir.RuleMatchToken
	targetIndex := 0

	if rule.Match.Kind == ast.RuleMatchNode {
		targetKind = ir.RuleMatchSyntaxNode
		targetIndex, _ = resolution.SyntaxNodeIndex(report.Target.Value(source))
	} else {
		targetIndex, _ = resolution.TokenIndex(report.Target.Value(source))
	}

	return ir.Report{
		Severity:    severityValue(report.Severity),
		TargetKind:  targetKind,
		TargetIndex: targetIndex,
		Message:     stringValue(source, report.Message),
	}
}

func compileExpressionArena(source string, resolution resolver.Resolution, arena *ir.ExpressionArena) {
	expressions := resolution.Document.Expressions

	for idx := range expressions.Nodes {
		expression := expressions.Nodes[idx]
		node := ir.ExpressionNode{
			FirstElementIdx:  expression.FirstElementIdx,
			AmountOfElements: expression.AmountOfElements,
		}

		switch expression.Kind {
		case ast.DefinitionExpressionCharacter:
			node.Kind = ir.ExpressionCharacter
			node.Character = characterValue(source, expression.Start)

		case ast.DefinitionExpressionString:
			node.Kind = ir.ExpressionString
			node.StringID = uint32(len(arena.Strings))
			arena.Strings = append(arena.Strings, stringValue(source, expression.Start))

		case ast.DefinitionExpressionRange:
			node.Kind = ir.ExpressionRange
			node.RangeStart = characterValue(source, expression.Start)
			node.RangeEnd = characterValue(source, expression.End)

		case ast.DefinitionExpressionReference:
			definitionIndex, _ := resolution.DefinitionIndex(expression.Start.Value(source))
			node.Kind = ir.ExpressionReference
			node.Reference = ir.ExpressionID(resolution.Definitions[definitionIndex].Expression)

		case ast.DefinitionExpressionConcatenation:
			node.Kind = ir.ExpressionConcatenation

		case ast.DefinitionExpressionAlternation:
			node.Kind = ir.ExpressionAlternation

		case ast.DefinitionExpressionGroup:
			node.Kind = ir.ExpressionGroup

		default:
			node.Kind = ir.ExpressionRepetition
			node.Repetition = repetitionKind(expression.Operator.Kind)
		}

		arena.Nodes[idx] = node
	}
}

func compileSyntaxExpressionArena(source string, resolution resolver.Resolution, arena *ir.SyntaxExpressionArena, ensureFieldID func(string) uint32) {
	expressions := resolution.Document.SyntaxExpressions

	for idx := range expressions.Nodes {
		expression := expressions.Nodes[idx]
		node := ir.SyntaxExpressionNode{
			FirstElementIdx:  expression.FirstElementIdx,
			AmountOfElements: expression.AmountOfElements,
			FieldID:          ^uint32(0),
		}

		switch expression.Kind {
		case ast.SyntaxExpressionReference:
			node.Kind = ir.SyntaxExpressionReference

			if tokenIndex, ok := resolution.TokenIndex(expression.Reference.Value(source)); ok {
				node.ReferenceKind = ir.SyntaxReferenceToken
				node.Reference = uint32(tokenIndex)
			} else {
				syntaxNodeIndex, _ := resolution.SyntaxNodeIndex(expression.Reference.Value(source))
				node.ReferenceKind = ir.SyntaxReferenceNode
				node.Reference = uint32(syntaxNodeIndex)
			}

		case ast.SyntaxExpressionAny:
			node.Kind = ir.SyntaxExpressionAny

		case ast.SyntaxExpressionCapture:
			node.Kind = ir.SyntaxExpressionCapture
			node.FieldID = ensureFieldID(expression.Field.Value(source))

		case ast.SyntaxExpressionConcatenation:
			node.Kind = ir.SyntaxExpressionConcatenation

		case ast.SyntaxExpressionAlternation:
			node.Kind = ir.SyntaxExpressionAlternation

		case ast.SyntaxExpressionGroup:
			node.Kind = ir.SyntaxExpressionGroup

		default:
			node.Kind = ir.SyntaxExpressionRepetition
			node.Repetition = repetitionKind(expression.Operator.Kind)
		}

		arena.Nodes[idx] = node
	}
}

func compileConditionPath(source string, path []token.Token, ensureFieldID func(string) uint32) []uint32 {
	if len(path) == 0 {
		return nil
	}

	compiled := make([]uint32, 0, len(path))

	for idx := range path {
		compiled = append(compiled, ensureFieldID(path[idx].Value(source)))
	}

	return compiled
}

func reinterpretExpressionChildren(children []ast.DefinitionExpressionID) []ir.ExpressionID {
	compiled := make([]ir.ExpressionID, len(children))

	for idx := range children {
		compiled[idx] = ir.ExpressionID(children[idx])
	}

	return compiled
}

func reinterpretSyntaxExpressionChildren(children []ast.SyntaxExpressionID) []ir.SyntaxExpressionID {
	compiled := make([]ir.SyntaxExpressionID, len(children))

	for idx := range children {
		compiled[idx] = ir.SyntaxExpressionID(children[idx])
	}

	return compiled
}

func countStringExpressions(expressions ast.DefinitionExpressionArena) int {
	count := 0

	for idx := range expressions.Nodes {
		if expressions.Nodes[idx].Kind == ast.DefinitionExpressionString {
			count++
		}
	}

	return count
}

func findSyntaxRootIndex(source string, resolution resolver.Resolution) int {
	referenced := make([]bool, len(resolution.SyntaxNodes))
	expressions := resolution.Document.SyntaxExpressions

	for idx := range resolution.SyntaxNodes {
		markReferencedSyntaxNodes(source, resolution, expressions, resolution.SyntaxNodes[idx].Expression, referenced)
	}

	rootIndex := -1

	for idx := range referenced {
		if referenced[idx] {
			continue
		}

		if rootIndex != -1 {
			return -1
		}

		rootIndex = idx
	}

	return rootIndex
}

func markReferencedSyntaxNodes(source string, resolution resolver.Resolution, expressions ast.SyntaxExpressionArena, expressionID ast.SyntaxExpressionID, referenced []bool) {
	expression := expressions.Node(expressionID)

	switch expression.Kind {
	case ast.SyntaxExpressionReference:
		referencedIndex, ok := resolution.SyntaxNodeIndex(expression.Reference.Value(source))

		if ok {
			referenced[referencedIndex] = true
		}

	case ast.SyntaxExpressionAlternation, ast.SyntaxExpressionConcatenation:
		for _, childID := range expressions.Children(expression) {
			markReferencedSyntaxNodes(source, resolution, expressions, childID, referenced)
		}

	case ast.SyntaxExpressionCapture, ast.SyntaxExpressionGroup, ast.SyntaxExpressionRepetition:
		markReferencedSyntaxNodes(source, resolution, expressions, expressions.Children(expression)[0], referenced)
	}
}

func repetitionKind(kind token.TokenKind) ir.RepetitionKind {
	switch kind {
	case token.TokenQuestion:
		return ir.RepetitionZeroOrOne

	case token.TokenStar:
		return ir.RepetitionZeroOrMore

	default:
		return ir.RepetitionOneOrMore
	}
}

func conditionProperty(source string, token token.Token) ir.ConditionProperty {
	if token.Value(source) == "text" {
		return ir.ConditionPropertyText
	}

	if token.Value(source) == "blankLines" {
		return ir.ConditionPropertyBlankLines
	}

	return ir.ConditionPropertyLength
}

func conditionSubject(source string, tok token.Token) ir.ConditionSubjectKind {
	switch tok.Value(source) {
	case "left":
		return ir.ConditionSubjectRelatedMatch

	case "right":
		return ir.ConditionSubjectMatch

	case "gap":
		return ir.ConditionSubjectGap

	default:
		return ir.ConditionSubjectMatch
	}
}

func conditionOperator(tok token.Token) ir.ConditionOperator {
	switch tok.Kind {
	case token.TokenEqualEqual:
		return ir.ConditionOperatorEqual

	case token.TokenBangEqual:
		return ir.ConditionOperatorNotEqual

	case token.TokenLess:
		return ir.ConditionOperatorLess

	case token.TokenLessEqual:
		return ir.ConditionOperatorLessEqual

	case token.TokenGreater:
		return ir.ConditionOperatorGreater

	case token.TokenStartsWith:
		return ir.ConditionOperatorStartsWith

	default:
		return ir.ConditionOperatorGreaterEqual
	}
}

func severityValue(tok token.Token) ir.Severity {
	switch tok.Kind {
	case token.TokenInfo:
		return ir.SeverityInfo

	case token.TokenWarn:
		return ir.SeverityWarn

	default:
		return ir.SeverityErr
	}
}

func characterValue(source string, tok token.Token) byte {
	if tok.Value(source)[1] != '\\' {
		return tok.Value(source)[1]
	}

	switch tok.Value(source)[2] {
	case '\\':
		return '\\'

	case '\'':
		return '\''

	case 'n':
		return '\n'

	case 'r':
		return '\r'

	default:
		return '\t'
	}
}

func stringValue(source string, token token.Token) string {
	value := make([]byte, 0, len(token.Value(source))-2)

	for idx := 1; idx < len(token.Value(source))-1; idx++ {
		character := token.Value(source)[idx]

		if character != '\\' {
			value = append(value, character)
			continue
		}

		idx++

		switch token.Value(source)[idx] {
		case '\\':
			value = append(value, '\\')

		case '"':
			value = append(value, '"')

		case 'n':
			value = append(value, '\n')

		case 'r':
			value = append(value, '\r')

		default:
			value = append(value, '\t')
		}
	}

	return string(value)
}

func integerValue(source string, token token.Token) int {
	value, _ := strconv.Atoi(token.Value(source))

	return value
}
