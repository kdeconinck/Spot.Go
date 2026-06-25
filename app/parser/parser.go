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
	scope, ok := parser.parseScopeSection()

	if !ok {
		return syntax.Document{
			Scope: scope,
			Span:  scope.Span,
		}
	}

	var definitions syntax.DefinitionsSection
	endPosition := scope.Span.End

	if parser.current.Kind == syntax.TokenDefinitions {
		definitions, ok = parser.parseDefinitionsSection()
		endPosition = definitions.Span.End

		if !ok {
			return syntax.Document{
				Scope:       scope,
				Definitions: definitions,
				Span: location.Span{
					Start: scope.Span.Start,
					End:   endPosition,
				},
			}
		}
	}

	end := parser.expect(syntax.TokenEOF)

	return syntax.Document{
		Scope:       scope,
		Definitions: definitions,
		Span: location.Span{
			Start: scope.Span.Start,
			End:   end.Span.End,
		},
	}
}

func (parser *parser) parseDefinitionsSection() (syntax.DefinitionsSection, bool) {
	start := parser.current
	parser.advance()

	if parser.current.Kind != syntax.TokenLeftBrace {
		parser.addDiagnostic(syntax.TokenLeftBrace)

		return syntax.DefinitionsSection{
			Span: start.Span,
		}, false
	}

	parser.advance()

	var definitions []syntax.Definition

	for parser.current.Kind == syntax.TokenIdentifier {
		definitions = append(definitions, parser.parseDefinition())
	}

	if parser.current.Kind != syntax.TokenRightBrace {
		if parser.current.Kind == syntax.TokenEOF {
			parser.addDiagnostic(syntax.TokenRightBrace)
		} else {
			parser.addDiagnostic(syntax.TokenIdentifier)
		}

		return syntax.DefinitionsSection{
			Definitions: definitions,
			Span: location.Span{
				Start: start.Span.Start,
				End:   parser.current.Span.End,
			},
		}, false
	}

	end := parser.current
	parser.advance()

	return syntax.DefinitionsSection{
		Definitions: definitions,
		Span: location.Span{
			Start: start.Span.Start,
			End:   end.Span.End,
		},
	}, true
}

func (parser *parser) parseDefinition() syntax.Definition {
	name := parser.current
	parser.advance()
	parser.expect(syntax.TokenEqual)
	expression := parser.parseDefinitionExpression()

	return syntax.Definition{
		Name:       name,
		Expression: expression,
		Span: location.Span{
			Start: name.Span.Start,
			End:   expression.Span.End,
		},
	}
}

func (parser *parser) parseDefinitionExpression() syntax.DefinitionExpression {
	start := parser.expect(syntax.TokenCharacter)

	if parser.current.Kind != syntax.TokenDotDot {
		return syntax.DefinitionExpression{
			Kind:  syntax.DefinitionExpressionCharacter,
			Start: start,
			Span:  start.Span,
		}
	}

	parser.advance()
	end := parser.expect(syntax.TokenCharacter)

	return syntax.DefinitionExpression{
		Kind:  syntax.DefinitionExpressionRange,
		Start: start,
		End:   end,
		Span: location.Span{
			Start: start.Span.Start,
			End:   end.Span.End,
		},
	}
}

func (parser *parser) parseScopeSection() (syntax.ScopeSection, bool) {
	if parser.current.Kind != syntax.TokenScope {
		parser.addDiagnostic(syntax.TokenScope)

		return syntax.ScopeSection{
			Span: parser.current.Span,
		}, false
	}

	start := parser.current
	parser.advance()

	if parser.current.Kind != syntax.TokenLeftBrace {
		parser.addDiagnostic(syntax.TokenLeftBrace)

		return syntax.ScopeSection{
			Span: start.Span,
		}, false
	}

	parser.advance()

	var entries []syntax.ScopeEntry

	for parser.current.Kind == syntax.TokenInclude || parser.current.Kind == syntax.TokenExclude {
		entries = append(entries, parser.parseScopeEntry())
	}

	if parser.current.Kind != syntax.TokenRightBrace {
		if parser.current.Kind == syntax.TokenEOF {
			parser.addDiagnostic(syntax.TokenRightBrace)
		} else {
			parser.addDiagnostic(syntax.TokenInclude)
		}

		return syntax.ScopeSection{
			Entries: entries,
			Span: location.Span{
				Start: start.Span.Start,
				End:   parser.current.Span.End,
			},
		}, false
	}

	end := parser.current
	parser.advance()

	return syntax.ScopeSection{
		Entries: entries,
		Span: location.Span{
			Start: start.Span.Start,
			End:   end.Span.End,
		},
	}, true
}

func (parser *parser) parseScopeEntry() syntax.ScopeEntry {
	start := parser.current
	kind := syntax.ScopeEntryInclude

	if start.Kind == syntax.TokenExclude {
		kind = syntax.ScopeEntryExclude
	}

	parser.advance()
	pattern := parser.expect(syntax.TokenString)

	return syntax.ScopeEntry{
		Kind:    kind,
		Pattern: pattern,
		Span: location.Span{
			Start: start.Span.Start,
			End:   pattern.Span.End,
		},
	}
}

func (parser *parser) expect(kind syntax.TokenKind) syntax.Token {
	if parser.current.Kind == kind {
		token := parser.current
		parser.advance()

		return token
	}

	token := parser.current
	parser.addDiagnostic(kind)

	return token
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
