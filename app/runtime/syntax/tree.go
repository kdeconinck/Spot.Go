// =====================================================================================================================
// == LICENSE:                 Copyright (c) 2026 Kevin De Coninck.
// == SPDX-License-Identifier: LicenseRef-PolyForm-Noncommercial-1.0.0
// =====================================================================================================================

// Package syntax parses scanned tokens into flat syntax trees using compiled Spot syntax definitions.
package syntax

import "github.com/kdeconinck/spot/runtime/scanner"

// NodeID identifies a node in a Tree.
type NodeID uint32

// Tree is a flat runtime syntax tree.
//
// Nodes stores matched syntax nodes. ChildIDs stores the parent-to-child relationships between those nodes.
type Tree struct {
	// Tokens is the token slice the tree was parsed from.
	Tokens []scanner.Token

	// Root is the root syntax node of the matched tree.
	Root NodeID

	// Nodes stores every matched syntax node.
	Nodes []Node

	// ChildIDs stores child node identifiers for parent nodes.
	ChildIDs []NodeID
}

// Reset prepares tree to store a syntax tree for tokens while reusing any existing capacity.
func (tree *Tree) Reset(tokens []scanner.Token) {
	tree.Tokens = tokens
	tree.Root = 0
	tree.Nodes = tree.Nodes[:0]
	tree.ChildIDs = tree.ChildIDs[:0]
}

// Node returns the runtime syntax node identified by id.
func (tree Tree) Node(id NodeID) Node {
	return tree.Nodes[id]
}

// Children returns the child node identifiers for node.
func (tree Tree) Children(node Node) []NodeID {
	return tree.ChildIDs[node.FirstElementIdx : node.FirstElementIdx+node.AmountOfElements]
}

// Node is one matched syntax node in a Tree.
type Node struct {
	// Kind is the source-order syntax node index in ir.Program.SyntaxNodes.
	Kind uint32

	// FirstTokenIndex is the start offset of this node's token range in Tree.Tokens.
	FirstTokenIndex uint32

	// AmountOfTokens is the number of tokens covered by this node.
	AmountOfTokens uint32

	// FirstElementIdx is the start offset of this node's children in Tree.ChildIDs.
	FirstElementIdx uint32

	// AmountOfElements is the number of children stored for this node.
	AmountOfElements uint32
}
