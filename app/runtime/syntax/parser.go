// =====================================================================================================================
// == LICENSE:                 Copyright (c) 2026 Kevin De Coninck.
// == SPDX-License-Identifier: LicenseRef-PolyForm-Noncommercial-1.0.0
// =====================================================================================================================

// Package syntax parses scanned tokens into flat syntax trees using compiled Spot syntax definitions.
package syntax

import (
	"errors"

	"github.com/kdeconinck/spot/runtime/ir"
	"github.com/kdeconinck/spot/runtime/scanner"
)

// Parser matches one compiled syntax root against a token stream.
type Parser struct {
	program         ir.Program
	rootNode        int
	scratchChildIDs []NodeID
	checkpoints     []parseCheckpoint
}

type parseCheckpoint struct {
	end             int
	nodeCount       int
	childCount      int
	scratchChildCnt int
}

// New returns a parser for rootNodeName.
func New(program ir.Program, rootNodeName string) (Parser, error) {
	for idx := range program.SyntaxNodes {
		if program.SyntaxNodes[idx].Name == rootNodeName {
			return Parser{
				program:  program,
				rootNode: idx,
			}, nil
		}
	}

	return Parser{}, errors.New(`syntax node "` + rootNodeName + `" is not declared`)
}

// Parse matches the parser's root node against tokens.
//
// Parse succeeds only when the root node consumes the entire token slice.
func (parser *Parser) Parse(tokens []scanner.Token) (Tree, bool) {
	var tree Tree

	if !parser.ParseInto(tokens, &tree) {
		return Tree{}, false
	}

	return tree, true
}

// ParseInto matches the parser's root node against tokens and stores the result in tree.
//
// ParseInto reuses tree's existing storage. This is the preferred API for repeated parsing in performance-sensitive
// code paths.
func (parser *Parser) ParseInto(tokens []scanner.Token, tree *Tree) bool {
	tree.Reset(tokens)
	parser.scratchChildIDs = parser.scratchChildIDs[:0]
	parser.checkpoints = parser.checkpoints[:0]

	if cap(tree.Nodes) < len(tokens) {
		tree.Nodes = make([]Node, 0, len(tokens))
	}

	if cap(tree.ChildIDs) < len(tokens) {
		tree.ChildIDs = make([]NodeID, 0, len(tokens))
	}

	end, rootNodeID, ok := parser.matchSyntaxNode(tokens, tree, parser.rootNode, 0)

	if !ok || end != len(tokens) {
		tree.Reset(nil)

		return false
	}

	tree.Root = rootNodeID

	return true
}

func (parser *Parser) matchSyntaxNode(tokens []scanner.Token, tree *Tree, syntaxNodeIndex int, start int) (int, NodeID, bool) {
	syntaxNode := parser.program.SyntaxNodes[syntaxNodeIndex]
	savedNodes := len(tree.Nodes)
	savedChildren := len(tree.ChildIDs)
	savedScratchChildren := len(parser.scratchChildIDs)
	end, ok := parser.matchSyntaxExpression(tokens, tree, syntaxNode.Expression, start)

	if !ok {
		tree.Nodes = tree.Nodes[:savedNodes]
		tree.ChildIDs = tree.ChildIDs[:savedChildren]
		parser.scratchChildIDs = parser.scratchChildIDs[:savedScratchChildren]

		return 0, 0, false
	}

	firstChild := uint32(len(tree.ChildIDs))
	directChildren := parser.scratchChildIDs[savedScratchChildren:]
	tree.ChildIDs = append(tree.ChildIDs, directChildren...)
	parser.scratchChildIDs = parser.scratchChildIDs[:savedScratchChildren]
	nodeID := NodeID(len(tree.Nodes))
	tree.Nodes = append(tree.Nodes, Node{
		Kind:             uint32(syntaxNodeIndex),
		FirstTokenIndex:  uint32(start),
		AmountOfTokens:   uint32(end - start),
		FirstElementIdx:  firstChild,
		AmountOfElements: uint32(len(directChildren)),
	})

	return end, nodeID, true
}

func (parser *Parser) matchSyntaxExpression(tokens []scanner.Token, tree *Tree, expressionID ir.SyntaxExpressionID, start int) (int, bool) {
	expression := parser.program.SyntaxExpressions.Node(expressionID)

	switch expression.Kind {
	case ir.SyntaxExpressionReference:
		return parser.matchSyntaxReference(tokens, tree, expression, start)

	case ir.SyntaxExpressionAny:
		return parser.matchAnySyntaxToken(tokens, start)

	case ir.SyntaxExpressionConcatenation:
		return parser.matchSyntaxConcatenation(tokens, tree, expression, start)

	case ir.SyntaxExpressionAlternation:
		return parser.matchSyntaxAlternation(tokens, tree, expression, start)

	case ir.SyntaxExpressionGroup:
		return parser.matchSyntaxExpression(tokens, tree, parser.program.SyntaxExpressions.Children(expression)[0], start)

	default:
		return parser.matchSyntaxRepetition(tokens, tree, expression, start)
	}
}

func (parser *Parser) matchAnySyntaxToken(tokens []scanner.Token, start int) (int, bool) {
	if start >= len(tokens) {
		return 0, false
	}

	return start + 1, true
}

func (parser *Parser) matchSyntaxReference(tokens []scanner.Token, tree *Tree, expression ir.SyntaxExpressionNode, start int) (int, bool) {
	if expression.ReferenceKind == ir.SyntaxReferenceToken {
		if start >= len(tokens) {
			return 0, false
		}

		if tokens[start].Name != parser.program.Tokens[expression.Reference].Name {
			return 0, false
		}

		return start + 1, true
	}

	end, childNodeID, ok := parser.matchSyntaxNode(tokens, tree, int(expression.Reference), start)

	if !ok {
		return 0, false
	}

	parser.scratchChildIDs = append(parser.scratchChildIDs, childNodeID)

	return end, true
}

func (parser *Parser) matchSyntaxConcatenation(tokens []scanner.Token, tree *Tree, expression ir.SyntaxExpressionNode, start int) (int, bool) {
	return parser.matchSyntaxConcatenationChildren(tokens, tree, parser.program.SyntaxExpressions.Children(expression), start)
}

func (parser *Parser) matchSyntaxAlternation(tokens []scanner.Token, tree *Tree, expression ir.SyntaxExpressionNode, start int) (int, bool) {
	for _, childID := range parser.program.SyntaxExpressions.Children(expression) {
		savedNodes := len(tree.Nodes)
		savedChildren := len(tree.ChildIDs)
		savedScratchChildren := len(parser.scratchChildIDs)
		end, ok := parser.matchSyntaxExpression(tokens, tree, childID, start)

		if ok {
			return end, true
		}

		tree.Nodes = tree.Nodes[:savedNodes]
		tree.ChildIDs = tree.ChildIDs[:savedChildren]
		parser.scratchChildIDs = parser.scratchChildIDs[:savedScratchChildren]
	}

	return 0, false
}

func (parser *Parser) matchSyntaxRepetition(tokens []scanner.Token, tree *Tree, expression ir.SyntaxExpressionNode, start int) (int, bool) {
	childID := parser.program.SyntaxExpressions.Children(expression)[0]

	switch expression.Repetition {
	case ir.RepetitionZeroOrOne:
		savedNodes := len(tree.Nodes)
		savedChildren := len(tree.ChildIDs)
		savedScratchChildren := len(parser.scratchChildIDs)
		end, ok := parser.matchSyntaxExpression(tokens, tree, childID, start)

		if ok {
			return end, true
		}

		tree.Nodes = tree.Nodes[:savedNodes]
		tree.ChildIDs = tree.ChildIDs[:savedChildren]
		parser.scratchChildIDs = parser.scratchChildIDs[:savedScratchChildren]

		return start, true

	case ir.RepetitionZeroOrMore:
		return parser.matchSyntaxMany(tokens, tree, childID, start, false)

	default:
		return parser.matchSyntaxMany(tokens, tree, childID, start, true)
	}
}

func (parser *Parser) matchSyntaxConcatenationChildren(tokens []scanner.Token, tree *Tree, children []ir.SyntaxExpressionID, start int) (int, bool) {
	if len(children) == 0 {
		return start, true
	}

	expression := parser.program.SyntaxExpressions.Node(children[0])

	if expression.Kind == ir.SyntaxExpressionRepetition {
		return parser.matchSyntaxRepetitionWithTail(tokens, tree, expression, children[1:], start)
	}

	checkpoint := parser.snapshot(tree, start)
	next, ok := parser.matchSyntaxExpression(tokens, tree, children[0], start)

	if !ok {
		return 0, false
	}

	end, ok := parser.matchSyntaxConcatenationChildren(tokens, tree, children[1:], next)

	if ok {
		return end, true
	}

	parser.restore(tree, checkpoint)

	return 0, false
}

func (parser *Parser) matchSyntaxRepetitionWithTail(tokens []scanner.Token, tree *Tree, expression ir.SyntaxExpressionNode, tail []ir.SyntaxExpressionID, start int) (int, bool) {
	minimum := 0

	if expression.Repetition == ir.RepetitionOneOrMore {
		minimum = 1
	}

	childID := parser.program.SyntaxExpressions.Children(expression)[0]
	base := len(parser.checkpoints)
	parser.checkpoints = append(parser.checkpoints, parser.snapshot(tree, start))
	current := start

	for {
		next, ok := parser.matchSyntaxExpression(tokens, tree, childID, current)

		if !ok || next == current {
			break
		}

		current = next
		parser.checkpoints = append(parser.checkpoints, parser.snapshot(tree, current))
	}

	for idx := len(parser.checkpoints) - 1; idx >= base+minimum; idx-- {
		checkpoint := parser.checkpoints[idx]
		parser.restore(tree, checkpoint)
		end, ok := parser.matchSyntaxConcatenationChildren(tokens, tree, tail, checkpoint.end)

		if ok {
			parser.checkpoints = parser.checkpoints[:base]

			return end, true
		}
	}

	parser.restore(tree, parser.checkpoints[base])
	parser.checkpoints = parser.checkpoints[:base]

	return 0, false
}

func (parser *Parser) matchSyntaxMany(tokens []scanner.Token, tree *Tree, childID ir.SyntaxExpressionID, start int, requireOne bool) (int, bool) {
	current := start
	count := 0

	for {
		savedNodes := len(tree.Nodes)
		savedChildren := len(tree.ChildIDs)
		savedScratchChildren := len(parser.scratchChildIDs)
		next, ok := parser.matchSyntaxExpression(tokens, tree, childID, current)

		if !ok {
			tree.Nodes = tree.Nodes[:savedNodes]
			tree.ChildIDs = tree.ChildIDs[:savedChildren]
			parser.scratchChildIDs = parser.scratchChildIDs[:savedScratchChildren]

			break
		}

		if next == current {
			tree.Nodes = tree.Nodes[:savedNodes]
			tree.ChildIDs = tree.ChildIDs[:savedChildren]
			parser.scratchChildIDs = parser.scratchChildIDs[:savedScratchChildren]

			break
		}

		current = next
		count++
	}

	if requireOne && count == 0 {
		return 0, false
	}

	return current, true
}

func (parser *Parser) snapshot(tree *Tree, end int) parseCheckpoint {
	return parseCheckpoint{
		end:             end,
		nodeCount:       len(tree.Nodes),
		childCount:      len(tree.ChildIDs),
		scratchChildCnt: len(parser.scratchChildIDs),
	}
}

func (parser *Parser) restore(tree *Tree, checkpoint parseCheckpoint) {
	tree.Nodes = tree.Nodes[:checkpoint.nodeCount]
	tree.ChildIDs = tree.ChildIDs[:checkpoint.childCount]
	parser.scratchChildIDs = parser.scratchChildIDs[:checkpoint.scratchChildCnt]
}
