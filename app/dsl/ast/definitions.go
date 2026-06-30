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

// DefinitionsSection is a parsed definitions section.
type DefinitionsSection struct {
	// FirstElementsIdx is the index of the section's first definition in Document.DefinitionList.
	FirstElementsIdx uint32

	// AmountOfElements is the number of definitions in the section.
	AmountOfElements uint32

	// Span is the byte range covered by the definitions section.
	Span location.Span
}

// Definition is a reusable character-level expression declaration.
type Definition struct {
	// Name is the identifier token naming the definition.
	Name token.Token

	// Expression is the root expression node assigned to the definition.
	Expression DefinitionExpressionID

	// Span is the byte range covered by the definition.
	Span location.Span
}

// DefinitionExpressionKind identifies the form of a definition expression.
type DefinitionExpressionKind uint8

const (
	// DefinitionExpressionCharacter is a single character literal expression.
	DefinitionExpressionCharacter DefinitionExpressionKind = iota

	// DefinitionExpressionString is a string literal expression.
	DefinitionExpressionString

	// DefinitionExpressionRange is a character range expression.
	DefinitionExpressionRange

	// DefinitionExpressionReference is a reference to another definition.
	DefinitionExpressionReference

	// DefinitionExpressionConcatenation is a sequence of adjacent expressions.
	DefinitionExpressionConcatenation

	// DefinitionExpressionAlternation is a list of alternative expressions.
	DefinitionExpressionAlternation

	// DefinitionExpressionGroup is a parenthesized expression.
	DefinitionExpressionGroup

	// DefinitionExpressionRepetition is a repeated expression.
	DefinitionExpressionRepetition
)

// DefinitionExpressionID identifies a node in a DefinitionExpressionArena.
type DefinitionExpressionID uint32

// DefinitionExpressionArena stores parsed definition and token expression nodes in flat slices.
//
// Nodes contains the actual expression records. ChildIDs stores the adjacency data for nodes that have children,
// such as alternations, concatenations, groups, and repetitions. A node's FirstChildIndex and ChildCount describe
// which segment of ChildIDs belongs to that node.
type DefinitionExpressionArena struct {
	// Nodes contains every expression node referenced by the parsed document.
	Nodes []DefinitionExpressionNode

	// ChildIDs stores child node identifiers for branch nodes.
	ChildIDs []DefinitionExpressionID
}

// Node returns the expression node identified by id.
func (arena DefinitionExpressionArena) Node(id DefinitionExpressionID) DefinitionExpressionNode {
	return arena.Nodes[id]
}

// Children returns the child expression identifiers for node.
func (arena DefinitionExpressionArena) Children(node DefinitionExpressionNode) []DefinitionExpressionID {
	return arena.ChildIDs[node.FirstElementIdx : node.FirstElementIdx+node.AmountOfElements]
}

// DefinitionExpressionNode is a parsed character-level definition expression.
type DefinitionExpressionNode struct {
	// Kind identifies the form of expression.
	Kind DefinitionExpressionKind

	// Start is the first token in the expression.
	Start token.Token

	// End is the final character literal in a range expression.
	End token.Token

	// Operator is the postfix operator token in a repetition expression.
	Operator token.Token

	// FirstElementIdx is the start offset of this node's children in ChildIDs.
	FirstElementIdx uint32

	// AmountOfElements is the number of children stored for this node.
	AmountOfElements uint32

	// Span is the byte range covered by the expression.
	Span location.Span
}
