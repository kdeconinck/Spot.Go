// =====================================================================================================================
// == LICENSE:                 Copyright (c) 2026 Kevin De Coninck.
// == SPDX-License-Identifier: LicenseRef-PolyForm-Noncommercial-1.0.0
// =====================================================================================================================

// Package compiler compiles validated Spot DSL syntax into runtime-oriented data structures.
package compiler

import (
	"strconv"

	"github.com/kdeconinck/spot/dsl/ast"
	"github.com/kdeconinck/spot/dsl/token"
	"github.com/kdeconinck/spot/runtime/ir"
)

// Compile compiles a validated Spot DSL document into a runtime program.
func Compile(source string, document ast.Document) ir.Program {
	definitions := map[string]ast.Definition{}
	tokenIndexes := map[string]int{}
	definitionList := document.SectionDefinitions(document.Definitions)
	tokenList := document.SectionTokens(document.Tokens)
	ruleList := document.SectionRules(document.Rules)

	for idx := range definitionList {
		definition := definitionList[idx]
		definitions[definition.Name.Value(source)] = definition
	}

	program := ir.Program{
		Tokens: make([]ir.Token, 0, len(tokenList)),
		Rules:  make([]ir.Rule, 0, len(ruleList)),
	}

	for idx := range tokenList {
		tok := tokenList[idx]
		tokenIndexes[tok.Name.Value(source)] = idx
		program.Tokens = append(program.Tokens, ir.Token{
			Name:       tok.Name.Value(source),
			Expression: compileExpression(source, tok.Expression, document.Expressions, definitions),
			Skip:       tok.Skip.Kind == token.TokenSkip,
		})
	}

	for idx := range ruleList {
		program.Rules = append(program.Rules, compileRule(source, ruleList[idx], tokenIndexes))
	}

	return program
}

func compileRule(source string, rule ast.Rule, tokenIndexes map[string]int) ir.Rule {
	return ir.Rule{
		Name:       rule.Name.Value(source),
		MatchToken: tokenIndexes[rule.Match.Token.Value(source)],
		Where:      compileCondition(source, rule.Where),
		Report:     compileReport(source, rule.Report, tokenIndexes),
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

func compileReport(source string, report ast.RuleReport, tokenIndexes map[string]int) ir.Report {
	return ir.Report{
		Severity:    severityValue(report.Severity),
		TargetToken: tokenIndexes[report.Target.Value(source)],
		Message:     stringValue(source, report.Message),
	}
}

func compileExpression(source string, expressionID ast.DefinitionExpressionID, expressions ast.DefinitionExpressionArena, definitions map[string]ast.Definition) ir.Expression {
	expression := expressions.Node(expressionID)

	switch expression.Kind {
	case ast.DefinitionExpressionCharacter:
		return ir.Expression{
			Kind:      ir.ExpressionCharacter,
			Character: characterValue(source, expression.Start),
		}

	case ast.DefinitionExpressionString:
		return ir.Expression{
			Kind:   ir.ExpressionString,
			String: stringValue(source, expression.Start),
		}

	case ast.DefinitionExpressionRange:
		return ir.Expression{
			Kind:       ir.ExpressionRange,
			RangeStart: characterValue(source, expression.Start),
			RangeEnd:   characterValue(source, expression.End),
		}

	case ast.DefinitionExpressionReference:
		return compileExpression(source, definitions[expression.Start.Value(source)].Expression, expressions, definitions)

	case ast.DefinitionExpressionConcatenation:
		return ir.Expression{
			Kind:  ir.ExpressionConcatenation,
			Terms: compileTerms(source, expressions.Children(expression), expressions, definitions),
		}

	case ast.DefinitionExpressionAlternation:
		return ir.Expression{
			Kind:  ir.ExpressionAlternation,
			Terms: compileTerms(source, expressions.Children(expression), expressions, definitions),
		}

	case ast.DefinitionExpressionGroup:
		return compileExpression(source, expressions.Children(expression)[0], expressions, definitions)

	default:
		return ir.Expression{
			Kind:       ir.ExpressionRepetition,
			Inner:      pointer(compileExpression(source, expressions.Children(expression)[0], expressions, definitions)),
			Repetition: repetitionKind(expression.Operator.Kind),
		}
	}
}

func compileTerms(source string, expressionIDs []ast.DefinitionExpressionID, expressions ast.DefinitionExpressionArena, definitions map[string]ast.Definition) []ir.Expression {
	terms := make([]ir.Expression, 0, len(expressionIDs))

	for idx := range expressionIDs {
		terms = append(terms, compileExpression(source, expressionIDs[idx], expressions, definitions))
	}

	return terms
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

func pointer(expression ir.Expression) *ir.Expression {
	return &expression
}
