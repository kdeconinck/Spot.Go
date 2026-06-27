// =====================================================================================================================
// == LICENSE:                 Copyright (c) 2026 Kevin De Coninck.
// == SPDX-License-Identifier: LicenseRef-PolyForm-Noncommercial-1.0.0
// =====================================================================================================================

// Package parser parses Spot DSL source text into syntax data structures.
package parser

import (
	"github.com/kdeconinck/spot/dsl/ast"
	"github.com/kdeconinck/spot/dsl/token"
)

func (p *parser) parseOptionalRulesSection() ast.RulesSection {
	if !p.at(token.TokenRules) {
		return ast.RulesSection{}
	}

	return p.parseRulesSection()
}

func (p *parser) parseRulesSection() ast.RulesSection {
	start := p.expect(token.TokenRules)

	if !p.match(token.TokenLeftBrace) {
		return ast.RulesSection{
			Span: start.Span,
		}
	}

	var rules []ast.Rule

	for p.at(token.TokenRule) {
		rules = append(rules, p.parseRule())
	}

	end := p.expectSectionEnd(token.TokenRule)

	return ast.RulesSection{
		Rules: rules,
		Span:  span(start.Span.Start, end.Span.End),
	}
}

func (p *parser) parseRule() ast.Rule {
	start := p.expect(token.TokenRule)
	name := p.expect(token.TokenIdentifier)

	if !p.match(token.TokenLeftBrace) {
		return ast.Rule{
			Name: name,
			Span: span(start.Span.Start, name.Span.End),
		}
	}

	match := p.parseRuleMatch()
	where := p.parseOptionalRuleCondition()
	report := p.parseRuleReport()
	end := p.expect(token.TokenRightBrace)

	return ast.Rule{
		Name:   name,
		Match:  match,
		Where:  where,
		Report: report,
		Span:   span(start.Span.Start, end.Span.End),
	}
}

func (p *parser) parseRuleMatch() ast.RuleMatch {
	start := p.expect(token.TokenMatch)
	tok := p.expect(token.TokenIdentifier)

	return ast.RuleMatch{
		Token: tok,
		Span:  span(start.Span.Start, tok.Span.End),
	}
}

func (p *parser) parseOptionalRuleCondition() ast.RuleCondition {
	if !p.at(token.TokenWhere) {
		return ast.RuleCondition{}
	}

	return p.parseRuleCondition()
}

func (p *parser) parseRuleCondition() ast.RuleCondition {
	start := p.expect(token.TokenWhere)
	subject := p.expect(token.TokenIdentifier)
	p.expect(token.TokenDot)
	property := p.expect(token.TokenIdentifier)
	operator := p.expectComparisonOperator()
	value := p.expectConditionLiteral()

	return ast.RuleCondition{
		Subject:  subject,
		Property: property,
		Operator: operator,
		Value:    value,
		Span:     span(start.Span.Start, value.Span.End),
	}
}

func (p *parser) parseRuleReport() ast.RuleReport {
	start := p.expect(token.TokenReport)
	severity := p.expectSeverity()
	p.expect(token.TokenAt)
	target := p.expect(token.TokenIdentifier)
	message := p.expect(token.TokenString)

	return ast.RuleReport{
		Severity: severity,
		Target:   target,
		Message:  message,
		Span:     span(start.Span.Start, message.Span.End),
	}
}

func (p *parser) expectComparisonOperator() token.Token {
	if p.at(token.TokenEqualEqual) ||
		p.at(token.TokenBangEqual) ||
		p.at(token.TokenLess) ||
		p.at(token.TokenLessEqual) ||
		p.at(token.TokenGreater) ||
		p.at(token.TokenGreaterEqual) {
		tok := p.current
		p.advance()

		return tok
	}

	tok := p.current
	p.addDiagnostic(token.TokenEqualEqual)

	return tok
}

func (p *parser) expectConditionLiteral() token.Token {
	if p.at(token.TokenString) || p.at(token.TokenInteger) {
		token := p.current
		p.advance()

		return token
	}

	tok := p.current
	p.addDiagnostic(token.TokenString)

	return tok
}

func (p *parser) expectSeverity() token.Token {
	if p.at(token.TokenInfo) || p.at(token.TokenWarn) || p.at(token.TokenErr) {
		token := p.current
		p.advance()

		return token
	}

	tok := p.current
	p.addDiagnostic(token.TokenWarn)

	return tok
}
