// =====================================================================================================================
// == LICENSE:                 Copyright (c) 2026 Kevin De Coninck.
// == SPDX-License-Identifier: LicenseRef-PolyForm-Noncommercial-1.0.0
// =====================================================================================================================

// Package syntax parses scanned tokens into flat syntax trees using compiled Spot syntax definitions.
package syntax

import "github.com/kdeconinck/spot/runtime/scanner"

const noFieldID = ^uint32(0)

// NodeID identifies a node in a Tree.
type NodeID uint32

// Tree is a flat runtime syntax tree.
//
// Nodes stores matched syntax nodes. ChildEdges stores the parent-to-child relationships between those nodes.
type Tree struct {
	// Tokens is the token slice the tree was parsed from.
	Tokens []scanner.Token

	// Root is the root syntax node of the matched tree.
	Root NodeID

	// Nodes stores every matched syntax node.
	Nodes []Node

	// ChildEdges stores child node identifiers plus optional field labels for parent nodes.
	ChildEdges []ChildEdge
}

// Reset prepares tree to store a syntax tree for tokens while reusing any existing capacity.
func (tree *Tree) Reset(tokens []scanner.Token) {
	tree.Tokens = tokens
	tree.Root = 0
	tree.Nodes = tree.Nodes[:0]
	tree.ChildEdges = tree.ChildEdges[:0]
}

// Node returns the runtime syntax node identified by id.
func (tree Tree) Node(id NodeID) Node {
	return tree.Nodes[id]
}

// Children returns the child node identifiers for node.
func (tree Tree) Children(node Node) []ChildEdge {
	return tree.ChildEdges[node.FirstElementIdx : node.FirstElementIdx+node.AmountOfElements]
}

// ChildByField returns the first child edge of node captured under fieldID.
func (tree Tree) ChildByField(node Node, fieldID uint32) (ChildEdge, bool) {
	for _, edge := range tree.Children(node) {
		if edge.FieldID == fieldID {
			return edge, true
		}
	}

	return ChildEdge{}, false
}

// ChildEdge is one parent-to-child runtime syntax edge.
type ChildEdge struct {
	// ChildID identifies the child syntax node.
	ChildID NodeID

	// FieldID identifies the named capture on this edge. It is max uint32 when unlabeled.
	FieldID uint32
}

// Node is one matched syntax node in a Tree.
type Node struct {
	// Kind is the source-order syntax node index in ir.Program.SyntaxNodes.
	Kind uint32

	// FirstTokenIndex is the start offset of this node's token range in Tree.Tokens.
	FirstTokenIndex uint32

	// AmountOfTokens is the number of tokens covered by this node.
	AmountOfTokens uint32

	// FirstElementIdx is the start offset of this node's children in Tree.ChildEdges.
	FirstElementIdx uint32

	// AmountOfElements is the number of children stored for this node.
	AmountOfElements uint32
}
