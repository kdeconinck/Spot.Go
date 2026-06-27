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
func Compile(document ast.Document) ir.Program {
	definitions := map[string]ast.Definition{}
	tokenIndexes := map[string]int{}

	for idx := range document.Definitions.Definitions {
		definition := document.Definitions.Definitions[idx]
		definitions[definition.Name.Text] = definition
	}

	program := ir.Program{
		Tokens: make([]ir.Token, 0, len(document.Tokens.Tokens)),
		Rules:  make([]ir.Rule, 0, len(document.Rules.Rules)),
	}

	for idx := range document.Tokens.Tokens {
		tok := document.Tokens.Tokens[idx]
		tokenIndexes[tok.Name.Text] = idx
		program.Tokens = append(program.Tokens, ir.Token{
			Name:       tok.Name.Text,
			Expression: compileExpression(tok.Expression, definitions),
			Skip:       tok.Skip.Kind == token.TokenSkip,
		})
	}

	for idx := range document.Rules.Rules {
		program.Rules = append(program.Rules, compileRule(document.Rules.Rules[idx], tokenIndexes))
	}

	return program
}

func compileRule(rule ast.Rule, tokenIndexes map[string]int) ir.Rule {
	return ir.Rule{
		Name:       rule.Name.Text,
		MatchToken: tokenIndexes[rule.Match.Token.Text],
		Where:      compileCondition(rule.Where),
		Report:     compileReport(rule.Report, tokenIndexes),
	}
}

func compileCondition(condition ast.RuleCondition) ir.Condition {
	if condition.Property.Text == "" {
		return ir.Condition{
			Property: ir.ConditionPropertyNone,
		}
	}

	compiled := ir.Condition{
		Property: conditionProperty(condition.Property),
		Operator: conditionOperator(condition.Operator),
	}

	if compiled.Property == ir.ConditionPropertyText {
		compiled.String = stringValue(condition.Value)

		return compiled
	}

	compiled.Integer = integerValue(condition.Value)

	return compiled
}

func compileReport(report ast.RuleReport, tokenIndexes map[string]int) ir.Report {
	return ir.Report{
		Severity:    severityValue(report.Severity),
		TargetToken: tokenIndexes[report.Target.Text],
		Message:     stringValue(report.Message),
	}
}

func compileExpression(expression ast.DefinitionExpression, definitions map[string]ast.Definition) ir.Expression {
	switch expression.Kind {
	case ast.DefinitionExpressionCharacter:
		return ir.Expression{
			Kind:      ir.ExpressionCharacter,
			Character: characterValue(expression.Start),
		}

	case ast.DefinitionExpressionString:
		return ir.Expression{
			Kind:   ir.ExpressionString,
			String: stringValue(expression.Start),
		}

	case ast.DefinitionExpressionRange:
		return ir.Expression{
			Kind:       ir.ExpressionRange,
			RangeStart: characterValue(expression.Start),
			RangeEnd:   characterValue(expression.End),
		}

	case ast.DefinitionExpressionReference:
		return compileExpression(definitions[expression.Start.Text].Expression, definitions)

	case ast.DefinitionExpressionConcatenation:
		return ir.Expression{
			Kind:  ir.ExpressionConcatenation,
			Terms: compileTerms(expression.Terms, definitions),
		}

	case ast.DefinitionExpressionAlternation:
		return ir.Expression{
			Kind:  ir.ExpressionAlternation,
			Terms: compileTerms(expression.Terms, definitions),
		}

	case ast.DefinitionExpressionGroup:
		return compileExpression(*expression.Inner, definitions)

	default:
		return ir.Expression{
			Kind:       ir.ExpressionRepetition,
			Inner:      pointer(compileExpression(*expression.Inner, definitions)),
			Repetition: repetitionKind(expression.Operator.Kind),
		}
	}
}

func compileTerms(expressions []ast.DefinitionExpression, definitions map[string]ast.Definition) []ir.Expression {
	terms := make([]ir.Expression, 0, len(expressions))

	for idx := range expressions {
		terms = append(terms, compileExpression(expressions[idx], definitions))
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

func conditionProperty(token token.Token) ir.ConditionProperty {
	if token.Text == "text" {
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

func characterValue(tok token.Token) byte {
	if tok.Text[1] != '\\' {
		return tok.Text[1]
	}

	switch tok.Text[2] {
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

func stringValue(token token.Token) string {
	value := make([]byte, 0, len(token.Text)-2)

	for idx := 1; idx < len(token.Text)-1; idx++ {
		character := token.Text[idx]

		if character != '\\' {
			value = append(value, character)
			continue
		}

		idx++

		switch token.Text[idx] {
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

func integerValue(token token.Token) int {
	value, _ := strconv.Atoi(token.Text)

	return value
}

func pointer(expression ir.Expression) *ir.Expression {
	return &expression
}
