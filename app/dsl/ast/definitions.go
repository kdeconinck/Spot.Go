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
	// Definitions are the declarations inside the definitions section.
	Definitions []Definition

	// Span is the byte range covered by the definitions section.
	Span location.Span
}

// Definition is a reusable character-level expression declaration.
type Definition struct {
	// Name is the identifier token naming the definition.
	Name token.Token

	// Expression is the character-level expression assigned to the definition.
	Expression DefinitionExpression

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

// DefinitionExpression is a parsed character-level definition expression.
type DefinitionExpression struct {
	// Kind identifies the form of expression.
	Kind DefinitionExpressionKind

	// Start is the first token in the expression.
	Start token.Token

	// End is the final character literal in a range expression.
	End token.Token

	// Operator is the postfix operator token in a repetition expression.
	Operator token.Token

	// Terms are the child expressions in an alternation or concatenation expression.
	Terms []DefinitionExpression

	// Inner is the expression contained in a grouped or repetition expression.
	Inner *DefinitionExpression

	// Span is the byte range covered by the expression.
	Span location.Span
}
