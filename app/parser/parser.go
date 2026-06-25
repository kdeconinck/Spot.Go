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

	parser.advance()
	document := parser.parseDocument()

	return document, parser.diagnostics
}

type parser struct {
	scanner     scanner.Scanner
	current     syntax.Token
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

	end := parser.expect(syntax.TokenEOF)

	return syntax.Document{
		Scope:       scope,
		Definitions: definitions,
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
	expression := parser.parseDefinitionExpression()

	return syntax.Definition{
		Name:       name,
		Expression: expression,
		Span:       span(name.Span.Start, expression.Span.End),
	}
}

func (parser *parser) parseDefinitionExpression() syntax.DefinitionExpression {
	first := parser.parseDefinitionPrimary()

	if !parser.at(syntax.TokenPipe) {
		return first
	}

	terms := []syntax.DefinitionExpression{
		first,
	}

	for parser.consume(syntax.TokenPipe) {
		terms = append(terms, parser.parseDefinitionPrimary())
	}

	return syntax.DefinitionExpression{
		Kind:  syntax.DefinitionExpressionAlternation,
		Terms: terms,
		Span:  span(first.Span.Start, terms[len(terms)-1].Span.End),
	}
}

func (parser *parser) parseDefinitionPrimary() syntax.DefinitionExpression {
	if parser.at(syntax.TokenLeftParen) {
		return parser.parseGroupedDefinitionExpression()
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

func (parser *parser) parseGroupedDefinitionExpression() syntax.DefinitionExpression {
	start := parser.expect(syntax.TokenLeftParen)
	inner := parser.parseDefinitionExpression()
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
	parser.current = parser.scanner.Next()
}

func span(start, end location.Position) location.Span {
	return location.Span{
		Start: start,
		End:   end,
	}
}
