// =====================================================================================================================
// == LICENSE:                 Copyright (c) 2026 Kevin De Coninck.
// == SPDX-License-Identifier: LicenseRef-PolyForm-Noncommercial-1.0.0
// =====================================================================================================================

// Package parser parses Spot DSL source text into syntax data structures.
package parser

import "github.com/kdeconinck/spot/syntax"

func (parser *parser) parseOptionalRulesSection() syntax.RulesSection {
	if !parser.at(syntax.TokenRules) {
		return syntax.RulesSection{}
	}

	return parser.parseRulesSection()
}

func (parser *parser) parseRulesSection() syntax.RulesSection {
	start := parser.expect(syntax.TokenRules)

	if !parser.match(syntax.TokenLeftBrace) {
		return syntax.RulesSection{
			Span: start.Span,
		}
	}

	var rules []syntax.Rule

	for parser.at(syntax.TokenRule) {
		rules = append(rules, parser.parseRule())
	}

	end := parser.expectSectionEnd(syntax.TokenRule)

	return syntax.RulesSection{
		Rules: rules,
		Span:  span(start.Span.Start, end.Span.End),
	}
}

func (parser *parser) parseRule() syntax.Rule {
	start := parser.expect(syntax.TokenRule)
	name := parser.expect(syntax.TokenIdentifier)

	if !parser.match(syntax.TokenLeftBrace) {
		return syntax.Rule{
			Name: name,
			Span: span(start.Span.Start, name.Span.End),
		}
	}

	match := parser.parseRuleMatch()
	report := parser.parseRuleReport()
	end := parser.expect(syntax.TokenRightBrace)

	return syntax.Rule{
		Name:   name,
		Match:  match,
		Report: report,
		Span:   span(start.Span.Start, end.Span.End),
	}
}

func (parser *parser) parseRuleMatch() syntax.RuleMatch {
	start := parser.expect(syntax.TokenMatch)
	token := parser.expect(syntax.TokenIdentifier)

	return syntax.RuleMatch{
		Token: token,
		Span:  span(start.Span.Start, token.Span.End),
	}
}

func (parser *parser) parseRuleReport() syntax.RuleReport {
	start := parser.expect(syntax.TokenReport)
	severity := parser.expectSeverity()
	parser.expect(syntax.TokenAt)
	target := parser.expect(syntax.TokenIdentifier)
	message := parser.expect(syntax.TokenString)

	return syntax.RuleReport{
		Severity: severity,
		Target:   target,
		Message:  message,
		Span:     span(start.Span.Start, message.Span.End),
	}
}

func (parser *parser) expectSeverity() syntax.Token {
	if parser.at(syntax.TokenInfo) || parser.at(syntax.TokenWarn) || parser.at(syntax.TokenErr) {
		token := parser.current
		parser.advance()

		return token
	}

	token := parser.current
	parser.addDiagnostic(syntax.TokenWarn)

	return token
}
