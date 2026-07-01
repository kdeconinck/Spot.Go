// =====================================================================================================================
// == LICENSE:                 Copyright (c) 2026 Kevin De Coninck.
// == SPDX-License-Identifier: LicenseRef-PolyForm-Noncommercial-1.0.0
// =====================================================================================================================

// Package parser parses Spot DSL source text into syntax data structures.
package parser

import (
	"github.com/kdeconinck/spot/dsl/ast"
	"github.com/kdeconinck/spot/dsl/lexer"
	"github.com/kdeconinck/spot/dsl/token"
	"github.com/kdeconinck/spot/location"
)

// Parse parses DSL source text into a syntax document. It stops at the first syntax error and returns it directly.
func Parse(src string) (ast.Document, error) {
	capacity := measureDocumentCapacity(src)
	parser := newParser(src, capacity)
	parser.current = parser.lexer.Next()
	parser.next = parser.lexer.Next()
	document, err := parser.parseDocument()

	if err != nil {
		return ast.Document{}, err
	}

	return document, nil
}

type parser struct {
	lexer    lexer.Lexer
	current  token.Token
	next     token.Token
	document ast.Document
}

func (p *parser) advance() {
	p.current = p.next
	p.next = p.lexer.Next()
}

func (p *parser) expectSectionEnd(entryKind token.TokenKind) (token.Token, error) {
	if p.isAt(token.TokenRightBrace) {
		return p.expect(token.TokenRightBrace)
	}

	if p.isAt(token.TokenEOF) {
		return token.Token{}, p.unexpected(token.TokenRightBrace)
	} else {
		return token.Token{}, p.unexpected(entryKind)
	}
}

func (p *parser) isAt(kind token.TokenKind) bool {
	return p.current.Kind == kind
}

func (p *parser) expect(kind token.TokenKind) (token.Token, error) {
	if p.isAt(kind) {
		token := p.current
		p.advance()

		return token, nil
	}

	return token.Token{}, p.unexpected(kind)
}

func (p *parser) consume(kind token.TokenKind) bool {
	if !p.isAt(kind) {
		return false
	}

	p.advance()

	return true
}

func (p *parser) match(kind token.TokenKind) error {
	if !p.isAt(kind) {
		return p.unexpected(kind)
	}

	p.advance()

	return nil
}

func (p *parser) unexpected(kind token.TokenKind) Diagnostic {
	return Diagnostic{
		Message: "Expected '" + kind.String() + "', found '" + p.current.Kind.String() + "'.",
		Span:    p.current.Span,
	}
}

func span(start, end location.Position) location.Span {
	return location.Span{
		Start: start,
		End:   end,
	}
}

func maxPosition(left, right location.Position) location.Position {
	if left > right {
		return left
	}

	return right
}

func newParser(src string, capacity documentCapacity) parser {
	return parser{
		lexer: lexer.New(src),
		document: ast.Document{
			Expressions: ast.DefinitionExpressionArena{
				Nodes:    make([]ast.DefinitionExpressionNode, 0, capacity.amountOfExpressionNodes),
				ChildIDs: make([]ast.DefinitionExpressionID, 0, capacity.amountOfExpressionChildReferences),
			},
			SyntaxExpressions: ast.SyntaxExpressionArena{
				Nodes:    make([]ast.SyntaxExpressionNode, 0, capacity.amountOfSyntaxExpressionNodes),
				ChildIDs: make([]ast.SyntaxExpressionID, 0, capacity.amountOfSyntaxExpressionChildren),
			},
			ScopeEntries:   make([]ast.ScopeEntry, 0, capacity.amountOfScopeElements),
			DefinitionList: make([]ast.Definition, 0, capacity.amountOfDefinitionElements),
			TokenList:      make([]ast.TokenDefinition, 0, capacity.amountOfTokenElements),
			SyntaxNodeList: make([]ast.SyntaxNode, 0, capacity.amountOfSyntaxElements),
			RuleList:       make([]ast.Rule, 0, capacity.amountOfRuleElements),
		},
	}
}
