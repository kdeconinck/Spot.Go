// =====================================================================================================================
// == LICENSE:                 Copyright (c) 2026 Kevin De Coninck.
// == SPDX-License-Identifier: LicenseRef-PolyForm-Noncommercial-1.0.0
// =====================================================================================================================

// Package parser parses Spot DSL source text into syntax data structures.
package parser

import "github.com/kdeconinck/spot/dsl/token"

func (parser *parser) parseOptionalRulesSection() token.RulesSection {
	if !parser.at(token.TokenRules) {
		return token.RulesSection{}
	}

	return parser.parseRulesSection()
}

func (parser *parser) parseRulesSection() token.RulesSection {
	start := parser.expect(token.TokenRules)

	if !parser.match(token.TokenLeftBrace) {
		return token.RulesSection{
			Span: start.Span,
		}
	}

	var rules []token.Rule

	for parser.at(token.TokenRule) {
		rules = append(rules, parser.parseRule())
	}

	end := parser.expectSectionEnd(token.TokenRule)

	return token.RulesSection{
		Rules: rules,
		Span:  span(start.Span.Start, end.Span.End),
	}
}

func (parser *parser) parseRule() token.Rule {
	start := parser.expect(token.TokenRule)
	name := parser.expect(token.TokenIdentifier)

	if !parser.match(token.TokenLeftBrace) {
		return token.Rule{
			Name: name,
			Span: span(start.Span.Start, name.Span.End),
		}
	}

	match := parser.parseRuleMatch()
	where := parser.parseOptionalRuleCondition()
	report := parser.parseRuleReport()
	end := parser.expect(token.TokenRightBrace)

	return token.Rule{
		Name:   name,
		Match:  match,
		Where:  where,
		Report: report,
		Span:   span(start.Span.Start, end.Span.End),
	}
}

func (parser *parser) parseRuleMatch() token.RuleMatch {
	start := parser.expect(token.TokenMatch)
	tok := parser.expect(token.TokenIdentifier)

	return token.RuleMatch{
		Token: tok,
		Span:  span(start.Span.Start, tok.Span.End),
	}
}

func (parser *parser) parseOptionalRuleCondition() token.RuleCondition {
	if !parser.at(token.TokenWhere) {
		return token.RuleCondition{}
	}

	return parser.parseRuleCondition()
}

func (parser *parser) parseRuleCondition() token.RuleCondition {
	start := parser.expect(token.TokenWhere)
	subject := parser.expect(token.TokenIdentifier)
	parser.expect(token.TokenDot)
	property := parser.expect(token.TokenIdentifier)
	operator := parser.expectComparisonOperator()
	value := parser.expectConditionLiteral()

	return token.RuleCondition{
		Subject:  subject,
		Property: property,
		Operator: operator,
		Value:    value,
		Span:     span(start.Span.Start, value.Span.End),
	}
}

func (parser *parser) parseRuleReport() token.RuleReport {
	start := parser.expect(token.TokenReport)
	severity := parser.expectSeverity()
	parser.expect(token.TokenAt)
	target := parser.expect(token.TokenIdentifier)
	message := parser.expect(token.TokenString)

	return token.RuleReport{
		Severity: severity,
		Target:   target,
		Message:  message,
		Span:     span(start.Span.Start, message.Span.End),
	}
}

func (parser *parser) expectComparisonOperator() token.Token {
	if parser.at(token.TokenEqualEqual) ||
		parser.at(token.TokenBangEqual) ||
		parser.at(token.TokenLess) ||
		parser.at(token.TokenLessEqual) ||
		parser.at(token.TokenGreater) ||
		parser.at(token.TokenGreaterEqual) {
		tok := parser.current
		parser.advance()

		return tok
	}

	tok := parser.current
	parser.addDiagnostic(token.TokenEqualEqual)

	return tok
}

func (parser *parser) expectConditionLiteral() token.Token {
	if parser.at(token.TokenString) || parser.at(token.TokenInteger) {
		token := parser.current
		parser.advance()

		return token
	}

	tok := parser.current
	parser.addDiagnostic(token.TokenString)

	return tok
}

func (parser *parser) expectSeverity() token.Token {
	if parser.at(token.TokenInfo) || parser.at(token.TokenWarn) || parser.at(token.TokenErr) {
		token := parser.current
		parser.advance()

		return token
	}

	tok := parser.current
	parser.addDiagnostic(token.TokenWarn)

	return tok
}
