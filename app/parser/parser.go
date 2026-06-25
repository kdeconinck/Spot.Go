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

	end := parser.expect(syntax.TokenEOF)

	return syntax.Document{
		Scope: scope,
		Span: location.Span{
			Start: scope.Span.Start,
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

	if parser.current.Kind != syntax.TokenRightBrace {
		parser.addDiagnostic(syntax.TokenRightBrace)

		return syntax.ScopeSection{
			Span: location.Span{
				Start: start.Span.Start,
				End:   parser.current.Span.End,
			},
		}, false
	}

	end := parser.current
	parser.advance()

	return syntax.ScopeSection{
		Span: location.Span{
			Start: start.Span.Start,
			End:   end.Span.End,
		},
	}, true
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
