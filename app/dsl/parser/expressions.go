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

func (p *parser) parseExpression(allowString bool) (ast.DefinitionExpressionID, error) {
	first, err := p.parseConcatenation(allowString)

	if err != nil {
		return 0, err
	}

	if !p.isAt(token.TokenPipe) {
		return first, nil
	}

	var buffer [4]ast.DefinitionExpressionID
	terms := buffer[:0]
	terms = append(terms, first)

	for p.consume(token.TokenPipe) {
		term, err := p.parseConcatenation(allowString)

		if err != nil {
			return 0, err
		}

		terms = append(terms, term)
	}

	return p.addExpressionBranch(
		ast.DefinitionExpressionAlternation,
		terms,
		span(p.expressionNode(first).Span.Start, p.expressionNode(terms[len(terms)-1]).Span.End),
	), nil
}

func (p *parser) parseConcatenation(allowString bool) (ast.DefinitionExpressionID, error) {
	first, err := p.parseRepetition(allowString)

	if err != nil {
		return 0, err
	}

	if !p.atExpressionContinuationStart(allowString) {
		return first, nil
	}

	var buffer [4]ast.DefinitionExpressionID
	terms := buffer[:0]
	terms = append(terms, first)

	for p.atExpressionContinuationStart(allowString) {
		term, err := p.parseRepetition(allowString)

		if err != nil {
			return 0, err
		}

		terms = append(terms, term)
	}

	return p.addExpressionBranch(
		ast.DefinitionExpressionConcatenation,
		terms,
		span(p.expressionNode(first).Span.Start, p.expressionNode(terms[len(terms)-1]).Span.End),
	), nil
}

func (p *parser) parseRepetition(allowString bool) (ast.DefinitionExpressionID, error) {
	inner, err := p.parsePrimary(allowString)

	if err != nil {
		return 0, err
	}

	if !p.isAt(token.TokenQuestion) && !p.isAt(token.TokenStar) && !p.isAt(token.TokenPlus) {
		return inner, nil
	}

	operator := p.current

	p.advance()

	return p.addExpressionNode(ast.DefinitionExpressionNode{
		Kind:             ast.DefinitionExpressionRepetition,
		Operator:         operator,
		FirstElementIdx:  p.appendExpressionChildren(inner),
		AmountOfElements: 1,
		Span:             span(p.expressionNode(inner).Span.Start, operator.Span.End),
	}), nil
}

func (p *parser) parsePrimary(allowString bool) (ast.DefinitionExpressionID, error) {
	if p.isAt(token.TokenLeftParen) {
		return p.parseGroupedExpression(allowString)
	}

	if allowString && p.isAt(token.TokenString) {
		start := p.current

		p.advance()

		return p.addExpressionNode(ast.DefinitionExpressionNode{
			Kind:  ast.DefinitionExpressionString,
			Start: start,
			Span:  start.Span,
		}), nil
	}

	if p.isAt(token.TokenIdentifier) {
		reference := p.current

		p.advance()

		return p.addExpressionNode(ast.DefinitionExpressionNode{
			Kind:  ast.DefinitionExpressionReference,
			Start: reference,
			Span:  reference.Span,
		}), nil
	}

	start, err := p.expect(token.TokenCharacter)

	if err != nil {
		return 0, err
	}

	if !p.consume(token.TokenDotDot) {
		return p.addExpressionNode(ast.DefinitionExpressionNode{
			Kind:  ast.DefinitionExpressionCharacter,
			Start: start,
			Span:  start.Span,
		}), nil
	}

	end, err := p.expect(token.TokenCharacter)

	if err != nil {
		return 0, err
	}

	return p.addExpressionNode(ast.DefinitionExpressionNode{
		Kind:  ast.DefinitionExpressionRange,
		Start: start,
		End:   end,
		Span:  span(start.Span.Start, end.Span.End),
	}), nil
}

func (p *parser) atExpressionContinuationStart(allowString bool) bool {
	if p.isAt(token.TokenIdentifier) && p.next.Kind == token.TokenEqual {
		return false
	}

	return p.isAt(token.TokenLeftParen) || p.isAt(token.TokenIdentifier) || p.isAt(token.TokenCharacter) || allowString && p.isAt(token.TokenString)
}

func (p *parser) parseGroupedExpression(allowString bool) (ast.DefinitionExpressionID, error) {
	start := p.current

	p.advance()

	inner, err := p.parseExpression(allowString)

	if err != nil {
		return 0, err
	}

	end, err := p.expect(token.TokenRightParen)

	if err != nil {
		return 0, err
	}

	return p.addExpressionNode(ast.DefinitionExpressionNode{
		Kind:             ast.DefinitionExpressionGroup,
		FirstElementIdx:  p.appendExpressionChildren(inner),
		AmountOfElements: 1,
		Span:             span(start.Span.Start, end.Span.End),
	}), nil
}

func (p *parser) addExpressionBranch(kind ast.DefinitionExpressionKind, children []ast.DefinitionExpressionID, exprSpan location.Span) ast.DefinitionExpressionID {
	firstElementIdx := p.appendExpressionChildren(children...)

	return p.addExpressionNode(ast.DefinitionExpressionNode{
		Kind:             kind,
		FirstElementIdx:  firstElementIdx,
		AmountOfElements: uint32(len(children)),
		Span:             exprSpan,
	})
}

func (p *parser) appendExpressionChildren(children ...ast.DefinitionExpressionID) uint32 {
	firstElementIdx := uint32(len(p.document.Expressions.ChildIDs))
	p.document.Expressions.ChildIDs = append(p.document.Expressions.ChildIDs, children...)

	return firstElementIdx
}

func (p *parser) addExpressionNode(node ast.DefinitionExpressionNode) ast.DefinitionExpressionID {
	id := ast.DefinitionExpressionID(len(p.document.Expressions.Nodes))
	p.document.Expressions.Nodes = append(p.document.Expressions.Nodes, node)

	return id
}

func (p *parser) expressionNode(id ast.DefinitionExpressionID) ast.DefinitionExpressionNode {
	return p.document.Expressions.Node(id)
}
