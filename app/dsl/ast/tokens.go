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

// TokensSection is a parsed tokens section.
type TokensSection struct {
	// FirstElementIdx is the index of the section's first token definition in Document.TokenList.
	FirstElementIdx uint32

	// AmountOfElements is the number of token definitions in the section.
	AmountOfElements uint32

	// Span is the byte range covered by the tokens section.
	Span location.Span
}

// TokenDefinition is a token expression declaration.
type TokenDefinition struct {
	// Name is the identifier token naming the emitted token.
	Name token.Token

	// Expression is the root expression node assigned to the token.
	Expression DefinitionExpressionID

	// Fallback marks a token declaration that consumes one otherwise-unmatched byte.
	Fallback token.Token

	// Skip marks a token declaration whose matches are not emitted.
	Skip token.Token

	// Span is the byte range covered by the token definition.
	Span location.Span
}
