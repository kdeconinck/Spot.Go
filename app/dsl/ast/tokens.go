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
	// Tokens are the declarations inside the tokens section.
	Tokens []TokenDefinition

	// Span is the byte range covered by the tokens section.
	Span location.Span
}

// TokenDefinition is a token expression declaration.
type TokenDefinition struct {
	// Name is the identifier token naming the emitted token.
	Name token.Token

	// Expression is the expression assigned to the token.
	Expression DefinitionExpression

	// Skip marks a token declaration whose matches are not emitted.
	Skip token.Token

	// Span is the byte range covered by the token definition.
	Span location.Span
}
