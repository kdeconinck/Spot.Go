// =====================================================================================================================
// == LICENSE:                 Copyright (c) 2026 Kevin De Coninck.
// == SPDX-License-Identifier: LicenseRef-PolyForm-Noncommercial-1.0.0
// =====================================================================================================================

// Package ast defines data structures that represent the AST (Abstract Syntax Tree) for Spot DSL syntax.
package ast

import (
	"github.com/kdeconinck/spot/dsl/token"
	"github.com/kdeconinck/spot/location"
)

// SyntaxSection is a parsed syntax section.
type SyntaxSection struct {
	// FirstElementIdx is the index of the section's first node declaration in Document.SyntaxNodeList.
	FirstElementIdx uint32

	// AmountOfElements is the number of node declarations in the section.
	AmountOfElements uint32

	// Span is the byte range covered by the syntax section.
	Span location.Span
}

// SyntaxNode is a declared syntax node kind.
type SyntaxNode struct {
	// Name is the identifier token naming the syntax node kind.
	Name token.Token

	// Expression is the root syntax expression assigned to the node.
	Expression SyntaxExpressionID

	// Span is the byte range covered by the node declaration.
	Span location.Span
}

// SyntaxExpressionKind identifies the form of a syntax expression.
type SyntaxExpressionKind uint8

const (
	// SyntaxExpressionReference is a reference to a token declaration or another syntax node.
	SyntaxExpressionReference SyntaxExpressionKind = iota

	// SyntaxExpressionConcatenation is a sequence of adjacent syntax expressions.
	SyntaxExpressionConcatenation

	// SyntaxExpressionAlternation is a list of alternative syntax expressions.
	SyntaxExpressionAlternation

	// SyntaxExpressionGroup is a parenthesized syntax expression.
	SyntaxExpressionGroup

	// SyntaxExpressionRepetition is a repeated syntax expression.
	SyntaxExpressionRepetition
)

// SyntaxExpressionID identifies a node in a SyntaxExpressionArena.
type SyntaxExpressionID uint32

// SyntaxExpressionArena stores parsed syntax-node expressions in flat slices.
//
// Nodes contains the actual expression records. ChildIDs stores the adjacency data for nodes that have children,
// such as alternations, concatenations, groups, and repetitions. A node's FirstElementIdx and AmountOfElements
// describe which segment of ChildIDs belongs to that node.
type SyntaxExpressionArena struct {
	// Nodes contains every syntax expression node referenced by the parsed document.
	Nodes []SyntaxExpressionNode

	// ChildIDs stores child node identifiers for branch nodes.
	ChildIDs []SyntaxExpressionID
}

// Node returns the syntax expression node identified by id.
func (arena SyntaxExpressionArena) Node(id SyntaxExpressionID) SyntaxExpressionNode {
	return arena.Nodes[id]
}

// Children returns the child syntax expression identifiers for node.
func (arena SyntaxExpressionArena) Children(node SyntaxExpressionNode) []SyntaxExpressionID {
	return arena.ChildIDs[node.FirstElementIdx : node.FirstElementIdx+node.AmountOfElements]
}

// SyntaxExpressionNode is a parsed syntax-node expression.
type SyntaxExpressionNode struct {
	// Kind identifies the form of expression.
	Kind SyntaxExpressionKind

	// Reference is the referenced token or syntax node name in a reference expression.
	Reference token.Token

	// Operator is the postfix operator token in a repetition expression.
	Operator token.Token

	// FirstElementIdx is the start offset of this node's children in ChildIDs.
	FirstElementIdx uint32

	// AmountOfElements is the number of children stored for this node.
	AmountOfElements uint32

	// Span is the byte range covered by the expression.
	Span location.Span
}
