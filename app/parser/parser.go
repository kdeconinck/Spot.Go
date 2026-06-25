// =====================================================================================================================
// == LICENSE:                 Copyright (c) 2026 Kevin De Coninck.
// == SPDX-License-Identifier: LicenseRef-PolyForm-Noncommercial-1.0.0
// =====================================================================================================================

package parser

import (
	"github.com/kdeconinck/spot/location"
	"github.com/kdeconinck/spot/scanner"
	"github.com/kdeconinck/spot/syntax"
)

// Parse parses DSL source text into a syntax document.
func Parse(src string) (syntax.Document, []Diagnostic) {
	parser := parser{
		scanner: scanner.New(src),
	}

	parser.current = parser.scanner.Next()
	parser.next = parser.scanner.Next()
	document := parser.parseDocument()

	return document, parser.diagnostics
}

type parser struct {
	scanner     scanner.Scanner
	current     syntax.Token
	next        syntax.Token
	diagnostics []Diagnostic
}

func (parser *parser) parseDocument() syntax.Document {
	diagnosticCount := len(parser.diagnostics)
	scope := parser.parseScopeSection()

	if len(parser.diagnostics) != diagnosticCount {
		return syntax.Document{
			Scope: scope,
			Span:  scope.Span,
		}
	}

	diagnosticCount = len(parser.diagnostics)
	definitions := parser.parseOptionalDefinitionsSection()

	if len(parser.diagnostics) != diagnosticCount {
		return syntax.Document{
			Scope:       scope,
			Definitions: definitions,
			Span:        span(scope.Span.Start, definitions.Span.End),
		}
	}

	diagnosticCount = len(parser.diagnostics)
	tokens := parser.parseOptionalTokensSection()

	if len(parser.diagnostics) != diagnosticCount {
		return syntax.Document{
			Scope:       scope,
			Definitions: definitions,
			Tokens:      tokens,
			Span:        span(scope.Span.Start, tokens.Span.End),
		}
	}

	end := parser.expect(syntax.TokenEOF)

	return syntax.Document{
		Scope:       scope,
		Definitions: definitions,
		Tokens:      tokens,
		Span:        span(scope.Span.Start, end.Span.End),
	}
}

func (parser *parser) parseScopeSection() syntax.ScopeSection {
	if !parser.at(syntax.TokenScope) {
		parser.addDiagnostic(syntax.TokenScope)

		return syntax.ScopeSection{
			Span: parser.current.Span,
		}
	}

	start := parser.expect(syntax.TokenScope)

	if !parser.match(syntax.TokenLeftBrace) {
		return syntax.ScopeSection{
			Span: start.Span,
		}
	}

	var entries []syntax.ScopeEntry

	for parser.at(syntax.TokenInclude) || parser.at(syntax.TokenExclude) {
		entries = append(entries, parser.parseScopeEntry())
	}

	end := parser.expectSectionEnd(syntax.TokenInclude)

	return syntax.ScopeSection{
		Entries: entries,
		Span:    span(start.Span.Start, end.Span.End),
	}
}

func (parser *parser) parseScopeEntry() syntax.ScopeEntry {
	start := parser.current
	kind := syntax.ScopeEntryInclude

	if parser.at(syntax.TokenExclude) {
		kind = syntax.ScopeEntryExclude
	}

	parser.advance()
	pattern := parser.expect(syntax.TokenString)

	return syntax.ScopeEntry{
		Kind:    kind,
		Pattern: pattern,
		Span:    span(start.Span.Start, pattern.Span.End),
	}
}

func (parser *parser) parseOptionalDefinitionsSection() syntax.DefinitionsSection {
	if !parser.at(syntax.TokenDefinitions) {
		return syntax.DefinitionsSection{}
	}

	return parser.parseDefinitionsSection()
}

func (parser *parser) parseDefinitionsSection() syntax.DefinitionsSection {
	start := parser.expect(syntax.TokenDefinitions)

	if !parser.match(syntax.TokenLeftBrace) {
		return syntax.DefinitionsSection{
			Span: start.Span,
		}
	}

	var definitions []syntax.Definition

	for parser.at(syntax.TokenIdentifier) {
		definitions = append(definitions, parser.parseDefinition())
	}

	end := parser.expectSectionEnd(syntax.TokenIdentifier)

	return syntax.DefinitionsSection{
		Definitions: definitions,
		Span:        span(start.Span.Start, end.Span.End),
	}
}

func (parser *parser) parseDefinition() syntax.Definition {
	name := parser.expect(syntax.TokenIdentifier)
	parser.expect(syntax.TokenEqual)
	expression := parser.parseExpression(false)

	return syntax.Definition{
		Name:       name,
		Expression: expression,
		Span:       span(name.Span.Start, expression.Span.End),
	}
}

func (parser *parser) parseOptionalTokensSection() syntax.TokensSection {
	if !parser.at(syntax.TokenTokens) {
		return syntax.TokensSection{}
	}

	return parser.parseTokensSection()
}

func (parser *parser) parseTokensSection() syntax.TokensSection {
	start := parser.expect(syntax.TokenTokens)

	if !parser.match(syntax.TokenLeftBrace) {
		return syntax.TokensSection{
			Span: start.Span,
		}
	}

	var tokens []syntax.TokenDefinition

	for parser.at(syntax.TokenIdentifier) {
		tokens = append(tokens, parser.parseTokenDefinition())
	}

	end := parser.expectSectionEnd(syntax.TokenIdentifier)

	return syntax.TokensSection{
		Tokens: tokens,
		Span:   span(start.Span.Start, end.Span.End),
	}
}

func (parser *parser) parseTokenDefinition() syntax.TokenDefinition {
	name := parser.expect(syntax.TokenIdentifier)
	parser.expect(syntax.TokenEqual)
	expression := parser.parseExpression(true)
	end := expression.Span.End
	var skip syntax.Token

	if parser.at(syntax.TokenSkip) {
		skip = parser.expect(syntax.TokenSkip)
		end = skip.Span.End
	}

	return syntax.TokenDefinition{
		Name:       name,
		Expression: expression,
		Skip:       skip,
		Span:       span(name.Span.Start, end),
	}
}

func (parser *parser) parseExpression(allowString bool) syntax.DefinitionExpression {
	first := parser.parseConcatenation(allowString)

	if !parser.at(syntax.TokenPipe) {
		return first
	}

	terms := []syntax.DefinitionExpression{
		first,
	}

	for parser.consume(syntax.TokenPipe) {
		terms = append(terms, parser.parseConcatenation(allowString))
	}

	return syntax.DefinitionExpression{
		Kind:  syntax.DefinitionExpressionAlternation,
		Terms: terms,
		Span:  span(first.Span.Start, terms[len(terms)-1].Span.End),
	}
}

func (parser *parser) parseConcatenation(allowString bool) syntax.DefinitionExpression {
	first := parser.parseRepetition(allowString)

	if !parser.atExpressionContinuationStart(allowString) {
		return first
	}

	terms := []syntax.DefinitionExpression{
		first,
	}

	for parser.atExpressionContinuationStart(allowString) {
		terms = append(terms, parser.parseRepetition(allowString))
	}

	return syntax.DefinitionExpression{
		Kind:  syntax.DefinitionExpressionConcatenation,
		Terms: terms,
		Span:  span(first.Span.Start, terms[len(terms)-1].Span.End),
	}
}

func (parser *parser) parseRepetition(allowString bool) syntax.DefinitionExpression {
	inner := parser.parsePrimary(allowString)

	if !parser.at(syntax.TokenQuestion) && !parser.at(syntax.TokenStar) && !parser.at(syntax.TokenPlus) {
		return inner
	}

	operator := parser.current
	parser.advance()

	return syntax.DefinitionExpression{
		Kind:     syntax.DefinitionExpressionRepetition,
		Operator: operator,
		Inner:    &inner,
		Span:     span(inner.Span.Start, operator.Span.End),
	}
}

func (parser *parser) parsePrimary(allowString bool) syntax.DefinitionExpression {
	if parser.at(syntax.TokenLeftParen) {
		return parser.parseGroupedExpression(allowString)
	}

	if allowString && parser.at(syntax.TokenString) {
		start := parser.expect(syntax.TokenString)

		return syntax.DefinitionExpression{
			Kind:  syntax.DefinitionExpressionString,
			Start: start,
			Span:  start.Span,
		}
	}

	if parser.at(syntax.TokenIdentifier) {
		reference := parser.expect(syntax.TokenIdentifier)

		return syntax.DefinitionExpression{
			Kind:  syntax.DefinitionExpressionReference,
			Start: reference,
			Span:  reference.Span,
		}
	}

	start := parser.expect(syntax.TokenCharacter)

	if !parser.consume(syntax.TokenDotDot) {
		return syntax.DefinitionExpression{
			Kind:  syntax.DefinitionExpressionCharacter,
			Start: start,
			Span:  start.Span,
		}
	}

	end := parser.expect(syntax.TokenCharacter)

	return syntax.DefinitionExpression{
		Kind:  syntax.DefinitionExpressionRange,
		Start: start,
		End:   end,
		Span:  span(start.Span.Start, end.Span.End),
	}
}

func (parser *parser) atExpressionContinuationStart(allowString bool) bool {
	if parser.at(syntax.TokenIdentifier) && parser.next.Kind == syntax.TokenEqual {
		return false
	}

	return parser.at(syntax.TokenLeftParen) ||
		parser.at(syntax.TokenIdentifier) ||
		parser.at(syntax.TokenCharacter) ||
		allowString && parser.at(syntax.TokenString)
}

func (parser *parser) parseGroupedExpression(allowString bool) syntax.DefinitionExpression {
	start := parser.expect(syntax.TokenLeftParen)
	inner := parser.parseExpression(allowString)
	end := parser.expect(syntax.TokenRightParen)

	return syntax.DefinitionExpression{
		Kind:  syntax.DefinitionExpressionGroup,
		Inner: &inner,
		Span:  span(start.Span.Start, end.Span.End),
	}
}

func (parser *parser) expectSectionEnd(entryKind syntax.TokenKind) syntax.Token {
	if parser.at(syntax.TokenRightBrace) {
		return parser.expect(syntax.TokenRightBrace)
	}

	if parser.at(syntax.TokenEOF) {
		parser.addDiagnostic(syntax.TokenRightBrace)
	} else {
		parser.addDiagnostic(entryKind)
	}

	return parser.current
}

func (parser *parser) expect(kind syntax.TokenKind) syntax.Token {
	if parser.at(kind) {
		token := parser.current
		parser.advance()

		return token
	}

	token := parser.current
	parser.addDiagnostic(kind)

	return token
}

func (parser *parser) match(kind syntax.TokenKind) bool {
	if !parser.at(kind) {
		parser.addDiagnostic(kind)

		return false
	}

	parser.advance()

	return true
}

func (parser *parser) consume(kind syntax.TokenKind) bool {
	if !parser.at(kind) {
		return false
	}

	parser.advance()

	return true
}

func (parser *parser) at(kind syntax.TokenKind) bool {
	return parser.current.Kind == kind
}

func (parser *parser) addDiagnostic(kind syntax.TokenKind) {
	parser.diagnostics = append(parser.diagnostics, Diagnostic{
		Message: "Expected '" + kind.String() + "', found '" + parser.current.Kind.String() + "'.",
		Span:    parser.current.Span,
	})
}

func (parser *parser) advance() {
	parser.current = parser.next
	parser.next = parser.scanner.Next()
}

func span(start, end location.Position) location.Span {
	return location.Span{
		Start: start,
		End:   end,
	}
}
