// =====================================================================================================================
// == LICENSE:                 Copyright (c) 2026 Kevin De Coninck.
// == SPDX-License-Identifier: LicenseRef-PolyForm-Noncommercial-1.0.0
// =====================================================================================================================

// Package parser parses Spot DSL source text into syntax data structures.
package parser

import (
	"github.com/kdeconinck/spot/dsl/ast"
	"github.com/kdeconinck/spot/dsl/token"
	"github.com/kdeconinck/spot/location"
)

func (p *parser) parseSyntaxExpression() (ast.SyntaxExpressionID, error) {
	first, err := p.parseSyntaxConcatenation()

	if err != nil {
		return 0, err
	}

	if !p.isAt(token.TokenPipe) {
		return first, nil
	}

	var buffer [4]ast.SyntaxExpressionID
	terms := buffer[:0]
	terms = append(terms, first)

	for p.consume(token.TokenPipe) {
		term, err := p.parseSyntaxConcatenation()

		if err != nil {
			return 0, err
		}

		terms = append(terms, term)
	}

	return p.addSyntaxExpressionBranch(
		ast.SyntaxExpressionAlternation,
		terms,
		span(p.syntaxExpressionNode(first).Span.Start, p.syntaxExpressionNode(terms[len(terms)-1]).Span.End),
	), nil
}

func (p *parser) parseSyntaxConcatenation() (ast.SyntaxExpressionID, error) {
	first, err := p.parseSyntaxRepetition()

	if err != nil {
		return 0, err
	}

	if !p.atSyntaxExpressionContinuationStart() {
		return first, nil
	}

	var buffer [4]ast.SyntaxExpressionID
	terms := buffer[:0]
	terms = append(terms, first)

	for p.atSyntaxExpressionContinuationStart() {
		term, err := p.parseSyntaxRepetition()

		if err != nil {
			return 0, err
		}

		terms = append(terms, term)
	}

	return p.addSyntaxExpressionBranch(
		ast.SyntaxExpressionConcatenation,
		terms,
		span(p.syntaxExpressionNode(first).Span.Start, p.syntaxExpressionNode(terms[len(terms)-1]).Span.End),
	), nil
}

func (p *parser) parseSyntaxRepetition() (ast.SyntaxExpressionID, error) {
	inner, err := p.parseSyntaxPrimary()

	if err != nil {
		return 0, err
	}

	if !p.isAt(token.TokenQuestion) && !p.isAt(token.TokenStar) && !p.isAt(token.TokenPlus) {
		return inner, nil
	}

	operator := p.current

	p.advance()

	return p.addSyntaxExpressionNode(ast.SyntaxExpressionNode{
		Kind:             ast.SyntaxExpressionRepetition,
		Operator:         operator,
		FirstElementIdx:  p.appendSyntaxExpressionChildren(inner),
		AmountOfElements: 1,
		Span:             span(p.syntaxExpressionNode(inner).Span.Start, operator.Span.End),
	}), nil
}

func (p *parser) parseSyntaxPrimary() (ast.SyntaxExpressionID, error) {
	if p.isAt(token.TokenLeftParen) {
		return p.parseGroupedSyntaxExpression()
	}

	reference, err := p.expect(token.TokenIdentifier)

	if err != nil {
		return 0, err
	}

	return p.addSyntaxExpressionNode(ast.SyntaxExpressionNode{
		Kind:      ast.SyntaxExpressionReference,
		Reference: reference,
		Span:      reference.Span,
	}), nil
}

func (p *parser) atSyntaxExpressionContinuationStart() bool {
	if p.isAt(token.TokenIdentifier) && p.next.Kind == token.TokenEqual {
		return false
	}

	return p.isAt(token.TokenLeftParen) || p.isAt(token.TokenIdentifier)
}

func (p *parser) parseGroupedSyntaxExpression() (ast.SyntaxExpressionID, error) {
	start := p.current

	p.advance()

	inner, err := p.parseSyntaxExpression()

	if err != nil {
		return 0, err
	}

	end, err := p.expect(token.TokenRightParen)

	if err != nil {
		return 0, err
	}

	return p.addSyntaxExpressionNode(ast.SyntaxExpressionNode{
		Kind:             ast.SyntaxExpressionGroup,
		FirstElementIdx:  p.appendSyntaxExpressionChildren(inner),
		AmountOfElements: 1,
		Span:             span(start.Span.Start, end.Span.End),
	}), nil
}

func (p *parser) addSyntaxExpressionBranch(kind ast.SyntaxExpressionKind, children []ast.SyntaxExpressionID, exprSpan location.Span) ast.SyntaxExpressionID {
	firstElementIdx := p.appendSyntaxExpressionChildren(children...)

	return p.addSyntaxExpressionNode(ast.SyntaxExpressionNode{
		Kind:             kind,
		FirstElementIdx:  firstElementIdx,
		AmountOfElements: uint32(len(children)),
		Span:             exprSpan,
	})
}

func (p *parser) appendSyntaxExpressionChildren(children ...ast.SyntaxExpressionID) uint32 {
	firstElementIdx := uint32(len(p.document.SyntaxExpressions.ChildIDs))
	p.document.SyntaxExpressions.ChildIDs = append(p.document.SyntaxExpressions.ChildIDs, children...)

	return firstElementIdx
}

func (p *parser) addSyntaxExpressionNode(node ast.SyntaxExpressionNode) ast.SyntaxExpressionID {
	id := ast.SyntaxExpressionID(len(p.document.SyntaxExpressions.Nodes))
	p.document.SyntaxExpressions.Nodes = append(p.document.SyntaxExpressions.Nodes, node)

	return id
}

func (p *parser) syntaxExpressionNode(id ast.SyntaxExpressionID) ast.SyntaxExpressionNode {
	return p.document.SyntaxExpressions.Node(id)
}
