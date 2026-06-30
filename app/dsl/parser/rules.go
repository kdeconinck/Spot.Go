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

func (p *parser) parseOptionalRulesSection() (ast.RulesSection, error) {
	if !p.isAt(token.TokenRules) {
		return ast.RulesSection{}, nil
	}

	return p.parseRulesSection()
}

func (p *parser) parseRulesSection() (ast.RulesSection, error) {
	start := p.current

	p.advance()

	if err := p.match(token.TokenLeftBrace); err != nil {
		return ast.RulesSection{}, err
	}

	firstRule := uint32(len(p.document.RuleList))

	for p.startsRule() {
		rule, err := p.parseRuleDeclaration()

		if err != nil {
			return ast.RulesSection{}, err
		}

		p.document.RuleList = append(p.document.RuleList, rule)
	}

	end, err := p.expectSectionEnd(token.TokenRule)

	if err != nil {
		return ast.RulesSection{}, err
	}

	return ast.RulesSection{
		FirstElementIdx:  firstRule,
		AmountOfElements: uint32(len(p.document.RuleList)) - firstRule,
		Span:             span(start.Span.Start, end.Span.End),
	}, nil
}

func (p *parser) startsRule() bool {
	return p.isAt(token.TokenRule) || p.atSeverity()
}

func (p *parser) parseRuleDeclaration() (ast.Rule, error) {
	if p.isAt(token.TokenRule) {
		return p.parseRule()
	}

	return p.parseSelectorRule()
}

func (p *parser) parseRule() (ast.Rule, error) {
	start := p.current

	p.advance()

	name, err := p.expect(token.TokenIdentifier)

	if err != nil {
		return ast.Rule{}, err
	}

	if err := p.match(token.TokenLeftBrace); err != nil {
		return ast.Rule{}, err
	}

	match, err := p.parseRuleMatch()

	if err != nil {
		return ast.Rule{}, err
	}

	where, err := p.parseOptionalRuleCondition()

	if err != nil {
		return ast.Rule{}, err
	}

	report, err := p.parseRuleReport()

	if err != nil {
		return ast.Rule{}, err
	}

	end, err := p.expect(token.TokenRightBrace)

	if err != nil {
		return ast.Rule{}, err
	}

	return ast.Rule{
		Name:   name,
		Match:  match,
		Where:  where,
		Report: report,
		Span:   span(start.Span.Start, end.Span.End),
	}, nil
}

func (p *parser) parseSelectorRule() (ast.Rule, error) {
	severity, err := p.expectSeverity()

	if err != nil {
		return ast.Rule{}, err
	}

	message, err := p.expect(token.TokenString)

	if err != nil {
		return ast.Rule{}, err
	}

	if _, err := p.expect(token.TokenColon); err != nil {
		return ast.Rule{}, err
	}

	match, err := p.parseSelectorRuleMatch()

	if err != nil {
		return ast.Rule{}, err
	}

	return ast.Rule{
		Match: match,
		Report: ast.RuleReport{
			Severity: severity,
			Target:   match.Target,
			Message:  message,
			Span:     span(severity.Span.Start, message.Span.End),
		},
		Span: span(severity.Span.Start, match.Span.End),
	}, nil
}

func (p *parser) parseSelectorRuleMatch() (ast.RuleMatch, error) {
	first, err := p.expect(token.TokenIdentifier)

	if err != nil {
		return ast.RuleMatch{}, err
	}

	if p.isAt(token.TokenColon) {
		return p.parseNegatedSelectorMatch(first)
	}

	if p.isAt(token.TokenGreater) || p.isAt(token.TokenIdentifier) {
		return p.parseRelatedSelectorMatch(first)
	}

	return ast.RuleMatch{
		Kind:   ast.RuleMatchNode,
		Target: first,
		Span:   first.Span,
	}, nil
}

func (p *parser) parseRelatedSelectorMatch(scopeTarget token.Token) (ast.RuleMatch, error) {
	scopeKind := ast.RuleMatchScopeInside

	if p.isAt(token.TokenGreater) {
		scopeKind = ast.RuleMatchScopeParent
		p.advance()
	}

	target, err := p.expect(token.TokenIdentifier)

	if err != nil {
		return ast.RuleMatch{}, err
	}

	return ast.RuleMatch{
		Kind:        ast.RuleMatchNode,
		Target:      target,
		ScopeKind:   scopeKind,
		ScopeTarget: scopeTarget,
		Span:        span(scopeTarget.Span.Start, target.Span.End),
	}, nil
}

func (p *parser) parseNegatedSelectorMatch(target token.Token) (ast.RuleMatch, error) {
	start := target.Span.Start

	if _, err := p.expect(token.TokenColon); err != nil {
		return ast.RuleMatch{}, err
	}

	if _, err := p.expect(token.TokenNot); err != nil {
		return ast.RuleMatch{}, err
	}

	if _, err := p.expect(token.TokenLeftParen); err != nil {
		return ast.RuleMatch{}, err
	}

	scopeTarget, err := p.expect(token.TokenIdentifier)

	if err != nil {
		return ast.RuleMatch{}, err
	}

	scopeKind := ast.RuleMatchScopeOutside

	if p.isAt(token.TokenGreater) {
		scopeKind = ast.RuleMatchScopeParentOutside
		p.advance()
	}

	end, err := p.expect(token.TokenStar)

	if err != nil {
		return ast.RuleMatch{}, err
	}

	if _, err := p.expect(token.TokenRightParen); err != nil {
		return ast.RuleMatch{}, err
	}

	return ast.RuleMatch{
		Kind:        ast.RuleMatchNode,
		Target:      target,
		ScopeKind:   scopeKind,
		ScopeTarget: scopeTarget,
		Span:        span(start, end.Span.End),
	}, nil
}

func (p *parser) parseRuleMatch() (ast.RuleMatch, error) {
	start, err := p.expect(token.TokenMatch)

	if err != nil {
		return ast.RuleMatch{}, err
	}

	matchKind := ast.RuleMatchToken

	if p.isAt(token.TokenNode) {
		p.advance()
		matchKind = ast.RuleMatchNode
	}

	target, err := p.expect(token.TokenIdentifier)

	if err != nil {
		return ast.RuleMatch{}, err
	}

	scopeKind := ast.RuleMatchScopeNone
	scopeTarget := token.Token{}
	end := target

	if p.isAt(token.TokenInside) || p.isAt(token.TokenOutside) {
		if p.isAt(token.TokenInside) {
			scopeKind = ast.RuleMatchScopeInside
		} else {
			scopeKind = ast.RuleMatchScopeOutside
		}

		p.advance()

		scopeTarget, err = p.expect(token.TokenIdentifier)

		if err != nil {
			return ast.RuleMatch{}, err
		}

		end = scopeTarget
	}

	return ast.RuleMatch{
		Kind:        matchKind,
		Target:      target,
		ScopeKind:   scopeKind,
		ScopeTarget: scopeTarget,
		Span:        span(start.Span.Start, end.Span.End),
	}, nil
}

func (p *parser) parseOptionalRuleCondition() (ast.RuleCondition, error) {
	if !p.isAt(token.TokenWhere) {
		return ast.RuleCondition{}, nil
	}

	return p.parseRuleCondition()
}

func (p *parser) parseRuleCondition() (ast.RuleCondition, error) {
	start := p.current

	p.advance()

	subject, err := p.expect(token.TokenIdentifier)

	if err != nil {
		return ast.RuleCondition{}, err
	}

	if _, err := p.expect(token.TokenDot); err != nil {
		return ast.RuleCondition{}, err
	}

	property, err := p.expect(token.TokenIdentifier)

	if err != nil {
		return ast.RuleCondition{}, err
	}

	operator, err := p.expectComparisonOperator()

	if err != nil {
		return ast.RuleCondition{}, err
	}

	value, err := p.expectConditionLiteral()

	if err != nil {
		return ast.RuleCondition{}, err
	}

	return ast.RuleCondition{
		Subject:  subject,
		Property: property,
		Operator: operator,
		Value:    value,
		Span:     span(start.Span.Start, value.Span.End),
	}, nil
}

func (p *parser) parseRuleReport() (ast.RuleReport, error) {
	start, err := p.expect(token.TokenReport)

	if err != nil {
		return ast.RuleReport{}, err
	}

	severity, err := p.expectSeverity()

	if err != nil {
		return ast.RuleReport{}, err
	}

	if _, err := p.expect(token.TokenAt); err != nil {
		return ast.RuleReport{}, err
	}

	target, err := p.expect(token.TokenIdentifier)

	if err != nil {
		return ast.RuleReport{}, err
	}

	message, err := p.expect(token.TokenString)

	if err != nil {
		return ast.RuleReport{}, err
	}

	return ast.RuleReport{
		Severity: severity,
		Target:   target,
		Message:  message,
		Span:     span(start.Span.Start, message.Span.End),
	}, nil
}

func (p *parser) expectComparisonOperator() (token.Token, error) {
	if p.isAt(token.TokenEqualEqual) ||
		p.isAt(token.TokenBangEqual) ||
		p.isAt(token.TokenLess) ||
		p.isAt(token.TokenLessEqual) ||
		p.isAt(token.TokenGreater) ||
		p.isAt(token.TokenGreaterEqual) {
		tok := p.current
		p.advance()

		return tok, nil
	}

	return token.Token{}, p.unexpected(token.TokenEqualEqual)
}

func (p *parser) expectConditionLiteral() (token.Token, error) {
	if p.isAt(token.TokenString) || p.isAt(token.TokenInteger) {
		token := p.current
		p.advance()

		return token, nil
	}

	return token.Token{}, p.unexpected(token.TokenString)
}

func (p *parser) expectSeverity() (token.Token, error) {
	if p.isAt(token.TokenInfo) || p.isAt(token.TokenWarn) || p.isAt(token.TokenErr) {
		token := p.current
		p.advance()

		return token, nil
	}

	return token.Token{}, p.unexpected(token.TokenWarn)
}

func (p *parser) atSeverity() bool {
	return p.isAt(token.TokenInfo) || p.isAt(token.TokenWarn) || p.isAt(token.TokenErr)
}
