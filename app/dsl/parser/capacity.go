// =====================================================================================================================
// == LICENSE:                 Copyright (c) 2026 Kevin De Coninck.
// == SPDX-License-Identifier: LicenseRef-PolyForm-Noncommercial-1.0.0
// =====================================================================================================================

// Package parser parses Spot DSL source text into syntax data structures.
package parser

import (
	"github.com/kdeconinck/spot/dsl/lexer"
	"github.com/kdeconinck/spot/dsl/token"
)

// Describes the exact flat storage required for one parsed document.
//
// The parser stores syntax in flat slices inside ast.Document instead of building a pointer-linked tree.
// Each field in documentCapacity maps directly to one of those flat slices.
//
// The sizing pass counts:
//   - Section entries stored in Document.ScopeEntries.
//   - Definition declarations stored in Document.DefinitionList.
//   - Token declarations stored in Document.TokenList.
//   - Rule declarations stored in Document.RuleList.
//   - Expression nodes stored in DefinitionExpressionArena.Nodes.
//   - Expression child references stored in DefinitionExpressionArena.ChildIDs.
//
// The real parser then uses these counts to allocate each slice once with exact capacity before parsing
// the document for real. This avoids growth reallocations while keeping the final AST layout compact and flat.
type documentCapacity struct {
	amountOfScopeElements             int // The number of scope entries stored in Document.ScopeEntries.
	amountOfDefinitionElements        int // The number of definitions stored in Document.DefinitionList.
	amountOfTokenElements             int // The number of token declarations stored in Document.TokenList.
	amountOfRuleElements              int // The number of rules stored in Document.RuleList.
	amountOfExpressionNodes           int // The number of expression nodes stored in DefinitionExpressionArena.Nodes.
	amountOfExpressionChildReferences int // The number of child references stored in DefinitionExpressionArena.ChildIDs.
}

// Walks src once and counts the exact size of every flat AST slice.
//
// The real parser does not build a pointer-linked tree. It stores the document in flat arrays, so it first needs to
// know how many entries each array will hold:
//   - Scope entries in Document.ScopeEntries.
//   - Definitions in Document.DefinitionList.
//   - Tokens in Document.TokenList.
//   - Rules in Document.RuleList.
//   - Expression nodes in DefinitionExpressionArena.Nodes.
//   - Expression child links in DefinitionExpressionArena.ChildIDs.
//
// "Expression node" means one syntax element inside a definition or token expression. Leaves such as a character,
// string, range, or reference each use one node. Composite forms such as alternation, concatenation, grouping, and
// repetition also use one node each.
//
// "Child link" means one parent -> child relation between two expression nodes. The flat arena stores nodes and
// links separately, so every composite node counts both itself and the references to its children.
//
// Examples:
//
//   - `letter = 'a'..'z'`
//     (*) 1 definition.
//     (*) 1 expression node for the range.
//     (*) 0 child links, because the range is a leaf.
//
//   - `value = ('a' | 'b')+`
//     (*) 1 definition.
//     (*) 5 expression nodes: repetition, group, alternation, `'a'`, `'b'`.
//     (*) 4 child links: repetition -> group, group -> alternation, alternation -> `'a'`, alternation -> `'b'`.
//
// The sizing pass follows the same grammar shape as the real parser, but it only counts storage. It does not build
// nodes, report diagnostics, or decide whether the source is valid.
func measureDocumentCapacity(src string) documentCapacity {
	parser := sizingParser{
		lexer: lexer.New(src),
	}

	parser.current = parser.lexer.Next()
	parser.next = parser.lexer.Next()

	parser.measureDocument()

	return parser.capacity
}

// Walks the DSL once to measure the exact flat storage required by the real parser.
// It mirrors the parser's grammar flow closely enough to produce exact capacities, but it never materializes syntax
// nodes or parse errors. The real parser remains the single authority for syntax validity.
type sizingParser struct {
	lexer    lexer.Lexer
	current  token.Token
	next     token.Token
	capacity documentCapacity
}

func (p *sizingParser) advance() {
	p.current = p.next
	p.next = p.lexer.Next()
}

func (p *sizingParser) advanceUnexpected() {
	if p.isAt(token.TokenEOF) {
		return
	}

	p.advance()
}

func (p *sizingParser) expectSectionEnd() {
	if p.isAt(token.TokenRightBrace) {
		p.advance()

		return
	}

	if p.isAt(token.TokenEOF) {
		return
	}

	p.advanceUnexpected()
}

func (p *sizingParser) isAt(kind token.TokenKind) bool {
	return p.current.Kind == kind
}

func (p *sizingParser) expect(kind token.TokenKind) {
	if p.isAt(kind) {
		p.advance()

		return
	}

	p.advanceUnexpected()
}

func (p *sizingParser) consume(kind token.TokenKind) bool {
	if !p.isAt(kind) {
		return false
	}

	p.advance()

	return true
}

func (p *sizingParser) match(kind token.TokenKind) bool {
	if !p.isAt(kind) {
		p.advanceUnexpected()

		return false
	}

	p.advance()

	return true
}
