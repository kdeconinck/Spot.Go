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
		Rules: make([]ir.Rule, 0, len(ruleList)),
	}

	copy(program.Expressions.ChildIDs, reinterpretExpressionChildren(expressions.ChildIDs))
	compileExpressionArena(source, resolution, &program.Expressions)
	copy(program.SyntaxExpressions.ChildIDs, reinterpretSyntaxExpressionChildren(syntaxExpressions.ChildIDs))
	compileSyntaxExpressionArena(source, resolution, &program.SyntaxExpressions)

	for idx := range tokenList {
		tok := tokenList[idx]
		program.Tokens = append(program.Tokens, ir.Token{
			Name:       tok.Name.Value(source),
			Expression: ir.ExpressionID(tok.Expression),
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
		program.Rules = append(program.Rules, compileRule(source, ruleList[idx], resolution))
	}

	return program
}

func compileRule(source string, rule ast.Rule, resolution resolver.Resolution) ir.Rule {
	matchTokenIndex, _ := resolution.TokenIndex(rule.Match.Token.Value(source))

	return ir.Rule{
		Name:       rule.Name.Value(source),
		MatchToken: matchTokenIndex,
		Where:      compileCondition(source, rule.Where),
		Report:     compileReport(source, rule.Report, resolution),
	}
}

func compileCondition(source string, condition ast.RuleCondition) ir.Condition {
	if condition.Property.Value(source) == "" {
		return ir.Condition{
			Property: ir.ConditionPropertyNone,
		}
	}

	compiled := ir.Condition{
		Property: conditionProperty(source, condition.Property),
		Operator: conditionOperator(condition.Operator),
	}

	if compiled.Property == ir.ConditionPropertyText {
		compiled.String = stringValue(source, condition.Value)

		return compiled
	}

	compiled.Integer = integerValue(source, condition.Value)

	return compiled
}

func compileReport(source string, report ast.RuleReport, resolution resolver.Resolution) ir.Report {
	targetTokenIndex, _ := resolution.TokenIndex(report.Target.Value(source))

	return ir.Report{
		Severity:    severityValue(report.Severity),
		TargetToken: targetTokenIndex,
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

func compileSyntaxExpressionArena(source string, resolution resolver.Resolution, arena *ir.SyntaxExpressionArena) {
	expressions := resolution.Document.SyntaxExpressions

	for idx := range expressions.Nodes {
		expression := expressions.Nodes[idx]
		node := ir.SyntaxExpressionNode{
			FirstElementIdx:  expression.FirstElementIdx,
			AmountOfElements: expression.AmountOfElements,
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

	return ir.ConditionPropertyLength
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
